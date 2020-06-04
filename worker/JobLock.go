package worker

import (
    "context"
    "go.etcd.io/etcd/clientv3"
    "gocron/common"
)

type JobLock struct {
    kv         clientv3.KV
    lease      clientv3.Lease
    jobName    string
    cancelFunc context.CancelFunc //取消锁需要
    leaseId    clientv3.LeaseID   //取消锁需要
    isLocked   bool
}

func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
    jobLock = &JobLock{
        kv:      kv,
        lease:   lease,
        jobName: jobName,
    }
    return
}

func (jobLock *JobLock) TryLock() (err error) {
    var (
        leaseGrant   *clientv3.LeaseGrantResponse
        leaseId      clientv3.LeaseID
        txn          clientv3.Txn
        txnResp      *clientv3.TxnResponse
        ctx          context.Context
        cancelFunc   context.CancelFunc
        jobLockKey   string
        keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
    )
    ctx, cancelFunc = context.WithCancel(context.TODO())

    if leaseGrant, err = jobLock.lease.Grant(ctx, 5); err != nil {
        return
    }
    leaseId = leaseGrant.ID

    if keepRespChan, err = jobLock.lease.KeepAlive(ctx, leaseId); err != nil {
        cancelFunc()                                  // 取消自动续租
        jobLock.lease.Revoke(context.TODO(), leaseId) //  释放租约
        return
    }

    go func() {
        var (
            keepAliveResp *clientv3.LeaseKeepAliveResponse
        )
        for {
            select {
            case keepAliveResp = <-keepRespChan:
                if keepAliveResp == nil {
                    goto END
                }
            }
        }
    END:
    }()

    jobLockKey = common.JOB_LOCK_DIR + jobLock.jobName
    txn = jobLock.kv.Txn(context.TODO())
    txn.If(clientv3.Compare(clientv3.CreateRevision(jobLockKey), "=", 0)).
        Then(clientv3.OpPut(jobLockKey, "", clientv3.WithLease(leaseId))).
        Else(clientv3.OpGet(jobLockKey))

    if txnResp, err = txn.Commit(); err != nil {
        cancelFunc()                                  // 取消自动续租
        jobLock.lease.Revoke(context.TODO(), leaseId) //  释放租约
        return
    }

    if !txnResp.Succeeded {
        err = common.ERR_LOCK_ALREADY_REQUIRED
        cancelFunc()                                  // 取消自动续租
        jobLock.lease.Revoke(context.TODO(), leaseId) //  释放租约
        return
    }

    jobLock.isLocked = true
    jobLock.cancelFunc = cancelFunc
    jobLock.leaseId = leaseId
    return
}

func (jobLock *JobLock) UnLock() {
    if jobLock.isLocked {
        jobLock.cancelFunc()
        jobLock.lease.Revoke(context.TODO(), jobLock.leaseId)
    }
}
