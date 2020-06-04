package worker

import (
    "gocron/common"
    "os/exec"
    "time"
)

type Executor struct {
}

var (
    G_executor *Executor
)

func InitExecutor() (err error) {
    G_executor = &Executor{}
    return
}

func (executor *Executor) ExecuteJob(jobInfo *common.JobExecuteInfo) {
    go func() {
        var (
            cmd     *exec.Cmd
            result  *common.JobExecuteResult
            jobLock *JobLock
            err     error
        )

        result = &common.JobExecuteResult{
            ExecuteInfo: jobInfo,
            Output:      make([]byte, 0),
        }

        jobLock = G_JobMgr.CreateJobLock(jobInfo.Job.Name)
        err = jobLock.TryLock()
        defer jobLock.UnLock()

        if err == nil { //上锁成功
            result.StartTime = time.Now()
            cmd = exec.CommandContext(jobInfo.CancelCtx, "/bin/bash", "-c", jobInfo.Job.Command)
            result.Output, result.Err = cmd.CombinedOutput()
            result.EndTime = time.Now()
        } else { // 上锁失败
            result.Err = err
            result.EndTime = time.Now()
        }
        G_scheduler.PushJobResult(result)
    }()
}
