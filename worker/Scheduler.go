package worker

import (
    "fmt"
    "gocron/common"
    "time"
)

type Scheduler struct {
    jobPlanTable map[string]*common.JobSchedulePlan
    jobEventChan chan *common.JobEvent
}

var (
    G_scheduler *Scheduler
)

func (scheduler *Scheduler) scheduleLoop() {
    var (
        scheduleAfter time.Duration
        scheduleTimer *time.Timer
        jobEvent      *common.JobEvent
    )
    scheduleAfter = scheduler.TrySchedule()
    scheduleTimer = time.NewTimer(scheduleAfter)

    for {
        select {
        case <-scheduleTimer.C:
        case jobEvent = <-scheduler.jobEventChan:
            scheduler.jobEventHandle(jobEvent)
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
        jobName  string
    )

    if len(scheduler.jobPlanTable) == 0 {
        scheduleAfter = 1 * time.Second
        return
    }
    now = time.Now()
    for jobName, jobPlan = range scheduler.jobPlanTable {
        if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
            // todo 执行任务
            fmt.Println("执行了任务", jobName)
            jobPlan.NextTime = jobPlan.Expr.Next(now)
        }

        if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
            nearTime = &jobPlan.NextTime
        }
        scheduleAfter = (*nearTime).Sub(now)
    }
    return
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

func (scheduler *Scheduler) PushEvent(jobEvent *common.JobEvent) {
    scheduler.jobEventChan <- jobEvent
}

func InitScheduler() (err error) {
    G_scheduler = &Scheduler{
        jobPlanTable: make(map[string]*common.JobSchedulePlan),
        jobEventChan: make(chan *common.JobEvent),
    }

    go G_scheduler.scheduleLoop()
    return
}
