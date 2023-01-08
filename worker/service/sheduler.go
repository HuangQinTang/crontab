package service

import (
	"crontab/common"
	"log"
)

// G_scheduler 调度服务，将任务数据维护到内存中并计算何时执行，调度执行到期的任务
var G_scheduler *Scheduler

// Scheduler 任务调度
type Scheduler struct {
	jobEventChan chan *common.JobEvent              //  etcd任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan // 任务调度计划表，key是任务名，value是任务的执行计划(任务数据、执行时间)
}

// InitJobScheduler 初始化调度器
func InitJobScheduler() {
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulePlan),
	}
	// 启动调度协程
	go G_scheduler.scheduleLoop()
	return
}

// scheduleLoop 调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent *common.JobEvent
	)

	// 定时任务common.Job
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: //监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		}
	}
}

// 处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExisted      bool
		err             error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: // 保存任务事件
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			log.Println(err.Error())
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE: // 删除任务事件
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

// PushJobEvent 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}
