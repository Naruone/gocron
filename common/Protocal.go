package common

import (
    "context"
    "encoding/json"
    "github.com/gorhill/cronexpr"
    "strings"
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

//任务执行状态
type JobExecuteInfo struct {
    Job        *Job
    PlanTime   time.Time
    RealTime   time.Time
    CancelCtx  context.Context
    CancelFunc context.CancelFunc
}

// 任务执行结果
type JobExecuteResult struct {
    ExecuteInfo *JobExecuteInfo // 执行状态
    Output      []byte          // 脚本输出
    Err         error           // 脚本错误原因
    StartTime   time.Time       // 启动时间
    EndTime     time.Time       // 结束时间
}

type JobLog struct {
    JobName      string `json:"job_name" bson:"job_name"`
    Command      string `json:"command" bson:"command"`
    PlanTime     int64  `json:"plan_time" bson:"plan_time"`
    ScheduleTime int64  `json:"schedule_time" bson:"schedule_time"`
    StartTime    int64  `json:"start_time" bson:"start_time"`
    EndTime      int64  `json:"end_time" bson:"end_time"`
    Output       string `json:"output" bson:"output"`
    Err          string `json:"err" bson:"err"`
}

type LogBatch struct {
    Logs []interface{}
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

func BuildJobExecuteInfo(jobPlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
    jobExecuteInfo = &JobExecuteInfo{
        Job:      jobPlan.Job,
        PlanTime: jobPlan.NextTime,
        RealTime: time.Now(),
    }
    jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
    return
}

//去除etcd job key 中的 /cron/job/前缀
func ExtractJobName(jobKey string) string {
    return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

func ExtractKillJobName(killKey string) string {
    return strings.TrimPrefix(killKey, JOB_KILLER_DIR)
}

func ExtractWorkerIp(workKey string) string {
    return strings.TrimPrefix(workKey, JOB_WORKER_DIR)
}
