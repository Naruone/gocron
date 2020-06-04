package worker

import (
    "fmt"
    "gocron/common"
    "time"
)

type Scheduler struct {
    jobPlanTable      map[string]*common.JobSchedulePlan
    jobEventChan      chan *common.JobEvent
    jobExecutingTable map[string]*common.JobExecuteInfo
    executor          *Executor
    jobResultChan     chan *common.JobExecuteResult
}

var (
    G_scheduler *Scheduler
)

func (scheduler *Scheduler) scheduleLoop() {
    var (
        scheduleAfter time.Duration
        scheduleTimer *time.Timer
        jobEvent      *common.JobEvent
        executeResult *common.JobExecuteResult
    )
    scheduleAfter = scheduler.TrySchedule()
    scheduleTimer = time.NewTimer(scheduleAfter)

    for {
        select {
        case <-scheduleTimer.C:
        case jobEvent = <-scheduler.jobEventChan:
            scheduler.jobEventHandle(jobEvent)
        case executeResult = <-scheduler.jobResultChan:
            scheduler.jobResultHandle(executeResult)
        }
        scheduleAfter = scheduler.TrySchedule()
        scheduleTimer.Reset(scheduleAfter)
    }
}

// 尝试运行任务 并且 重新计算任务执行时间
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
    var (
        jobPlan  *common.JobSchedulePlan
        now      time.Time
        nearTime *time.Time
    )

    if len(scheduler.jobPlanTable) == 0 {
        scheduleAfter = 1 * time.Second
        return
    }
    now = time.Now()
    for _, jobPlan = range scheduler.jobPlanTable {
        if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
            // todo 执行任务
            scheduler.TryStartJob(jobPlan)
            jobPlan.NextTime = jobPlan.Expr.Next(now)
        }

        if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
            nearTime = &jobPlan.NextTime
        }
        scheduleAfter = (*nearTime).Sub(now)
    }
    return
}

func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
    var (
        executeInfo *common.JobExecuteInfo
        jobExisted  bool
    )

    if executeInfo, jobExisted = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExisted {
        return
    }
    executeInfo = common.BuildJobExecuteInfo(jobPlan)
    scheduler.jobExecutingTable[jobPlan.Job.Name] = executeInfo
    G_executor.ExecuteJob(executeInfo)
}

func (scheduler *Scheduler) jobEventHandle(jobEvent *common.JobEvent) {
    var (
        jobSchedulePlan *common.JobSchedulePlan
        jobExisted      bool
        err             error
    )
    switch jobEvent.EventType {
    case common.JOB_EVENT_SAVE:
        if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
            return
        }
        scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
    case common.JOB_EVENT_DELETE:
        if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
            delete(scheduler.jobPlanTable, jobEvent.Job.Name)
        }
    }
}

func (scheduler *Scheduler) jobResultHandle(result *common.JobExecuteResult) {
    delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)
    fmt.Println("任务日志", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
    //todo 记录执行日志
}

func (scheduler *Scheduler) PushEvent(jobEvent *common.JobEvent) {
    scheduler.jobEventChan <- jobEvent
}

func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
    scheduler.jobResultChan <- jobResult
}

func InitScheduler() (err error) {
    G_scheduler = &Scheduler{
        jobPlanTable:      make(map[string]*common.JobSchedulePlan),
        jobEventChan:      make(chan *common.JobEvent),
        jobExecutingTable: make(map[string]*common.JobExecuteInfo),
        jobResultChan:     make(chan *common.JobExecuteResult),
    }

    go G_scheduler.scheduleLoop()
    return
}
