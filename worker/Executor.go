package worker

import (
    "context"
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
            cmd    *exec.Cmd
            output []byte
            err    error
            result *common.JobExecuteResult
        )
        result = &common.JobExecuteResult{
            ExecuteInfo: jobInfo,
            Output:      make([]byte, 0),
            StartTime:   time.Now(),
        }
        cmd = exec.CommandContext(context.TODO(), "/bin/bash", "-c", jobInfo.Job.Command)
        output, err = cmd.CombinedOutput()
        result.Output = output
        result.Err = err
        result.EndTime = time.Now()
        G_scheduler.PushJobResult(result)
    }()
}
