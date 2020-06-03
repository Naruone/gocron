package common

import (
    "encoding/json"
    "github.com/gorhill/cronexpr"
    "time"
)

type Job struct {
    Name     string `json:"name"`
    Command  string `json:"command"`
    CronExpr string `json:"cronExpr"`
}

type Response struct {
    Errno int         `json:"errno"`
    Msg   string      `json:"msg"`
    Data  interface{} `json:"data"`
}

type JobEvent struct {
    EventType int
    Job       *Job
}

//任务调度计划
type JobSchedulePlan struct {
    Job      *Job
    Expr     *cronexpr.Expression
    NextTime time.Time
}

func BuildResponse(code int, msg string, data interface{}) (resp []byte, err error) {
    var response Response

    response = Response{
        Errno: code,
        Msg:   msg,
        Data:  data,
    }

    resp, err = json.Marshal(response)
    return
}

//json 字符串转 Job
func UnpackJob(value []byte) (ret *Job, err error) {
    job := &Job{}
    if err = json.Unmarshal(value, job); err != nil {
        return
    }
    ret = job
    return
}

func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
    var (
        expr *cronexpr.Expression
    )

    // 解析JOB的cron表达式
    if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
        return
    }

    jobSchedulePlan = &JobSchedulePlan{
        Job:      job,
        Expr:     expr,
        NextTime: expr.Next(time.Now()),
    }
    return
}
