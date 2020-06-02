package worker

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"gocron/common"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	G_JobMgr *JobMgr
)

func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,
		DialTimeout: time.Duration(G_config.DialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	G_JobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

//保存任务
func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		putResp   *clientv3.PutResponse
		jobPath   string
		jobStr    []byte
		oldJobObj *common.Job
	)
	if jobStr, err = json.Marshal(job); err != nil {
		return
	}

	jobPath = common.JOB_SAVE_DIR + job.Name
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobPath, string(jobStr), clientv3.WithPrevKV()); err != nil {
		return
	}
	if putResp.PrevKv != nil {
		if oldJobObj, err = common.UnpackJob(putResp.PrevKv.Value); err != nil {
			err = nil
			return
		}
		oldJob = oldJobObj
	}
	return
}

//任务列表
func (jobMgr *JobMgr) JobList() (jobList []*common.Job, err error) {
	var (
		getResp *clientv3.GetResponse
		job     *common.Job
	)
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	jobList = make([]*common.Job, 0)

	for _, v := range getResp.Kvs {
		if job, err = common.UnpackJob(v.Value); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}

	return
}

//删除任务
func (jobMgr *JobMgr) JobDelete(jobName string) (oldJob *common.Job, err error) {
	var (
		delResp *clientv3.DeleteResponse
		dJob    *common.Job
	)
	if delResp, err = jobMgr.kv.Delete(context.TODO(), common.JOB_SAVE_DIR+jobName, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(delResp.PrevKvs) != 0 {
		if dJob, err = common.UnpackJob(delResp.PrevKvs[0].Value); err != nil {
			err = nil
			return
		}
		oldJob = dJob
	}
	return
}

//杀死任务
func (jobMgr *JobMgr) KillJob(name string) (err error) {
	var (
		jobPath        string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}
	leaseId = leaseGrantResp.ID

	jobPath = common.JOB_KILLER_DIR + name
	if _, err = jobMgr.kv.Put(context.TODO(), jobPath, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
