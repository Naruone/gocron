package worker

import (
    "context"
    "github.com/coreos/etcd/mvcc/mvccpb"
    "go.etcd.io/etcd/clientv3"
    "gocron/common"
    "time"
)

type JobMgr struct {
    client  *clientv3.Client
    kv      clientv3.KV
    lease   clientv3.Lease
    watcher clientv3.Watcher
}

var (
    G_JobMgr *JobMgr
)

func InitJobMgr() (err error) {
    var (
        config  clientv3.Config
        client  *clientv3.Client
        kv      clientv3.KV
        lease   clientv3.Lease
        watcher clientv3.Watcher
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
    watcher = clientv3.NewWatcher(client)
    G_JobMgr = &JobMgr{
        client:  client,
        kv:      kv,
        lease:   lease,
        watcher: watcher,
    }

    G_JobMgr.watchJobs()
    return
}

//任务列表
func (jobMgr *JobMgr) watchJobs() (err error) {
    var (
        getResp            *clientv3.GetResponse
        job                *common.Job
        watchStartRevision int64
        watchChan          clientv3.WatchChan
        watchResp          clientv3.WatchResponse
        watchEvent         *clientv3.Event
        jobEvent           *common.JobEvent
    )
    if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
        return
    }

    for _, v := range getResp.Kvs {
        if job, err = common.UnpackJob(v.Value); err == nil {
            //todo 传给scheduler
            jobEvent = &common.JobEvent{
                EventType: common.JOB_EVENT_SAVE,
                Job:       job,
            }
            G_scheduler.PushEvent(jobEvent)
        }
    }

    go func() {
        watchStartRevision = getResp.Header.Revision + 1
        watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())

        for watchResp = range watchChan {
            for _, watchEvent = range watchResp.Events {
                switch watchEvent.Type {
                case mvccpb.PUT:
                    if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
                        continue
                    }
                    jobEvent = &common.JobEvent{
                        EventType: common.JOB_EVENT_SAVE,
                        Job:       job,
                    }
                case mvccpb.DELETE:
                    jobEvent = &common.JobEvent{
                        EventType: common.JOB_EVENT_DELETE,
                        Job: &common.Job{
                            Name: string(watchEvent.Kv.Key),
                        },
                    }
                }
                G_scheduler.PushEvent(jobEvent)
            }
        }
    }()
    return
}

func (jobMgr *JobMgr) CreateJobLock(jobName string) *JobLock {
    return InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
}
