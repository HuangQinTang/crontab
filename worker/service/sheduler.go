package service

import (
	"crontab/common"
	"fmt"
	"log"
	"time"
)

// G_scheduler 调度服务，将任务数据维护到内存中并计算何时执行，调度执行到期的任务
var G_scheduler *Scheduler

// Scheduler 任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              // etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan // 任务调度计划表，key是任务名，value是任务的执行计划(任务数据、执行时间)
	jobExecutingTable map[string]*common.JobExecuteInfo  // 任务执行表，key正在执行的任务名，value任务执行信息
	jobResultChan     chan *common.JobExecuteResult      // 任务结果队列
}

// InitJobScheduler 初始化调度器
func InitJobScheduler() {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult),
	}
	// 启动调度协程
	go G_scheduler.scheduleLoop()
	return
}

// scheduleLoop 调度协程
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration            //下次调度的时间
		scheduleTimer *time.Timer              //定时器
		jobResult     *common.JobExecuteResult //任务执行结果
	)

	// 计算下次调度的时间(任务为空时，默认1秒)
	scheduleAfter = scheduler.trySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	// 定时任务common.Job
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: //监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
		case jobResult = <-scheduler.jobResultChan: // 监听任务执行结果
			scheduler.handleJobResult(jobResult)
		}

		// 尝试调度一次任务，返回下个需要执行的任务时间戳
		scheduleAfter = scheduler.trySchedule()
		// 根据新的最近要执行任务的时间重置定时器间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

// TrySchedule 调度任务，尝试执行到期任务，重新计算最近要调度的任务到期时间戳并返回
func (scheduler *Scheduler) trySchedule() (scheduleAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)

	// 如果任务表为空话，睡眠1秒
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return scheduleAfter
	}

	// 当前时间
	now = time.Now()

	// 遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) { //现在就要执行的任务
			scheduler.startJob(jobPlan)
			//fmt.Println(time.Now().Format(time.Stamp), "任务名称:", jobPlan.Job.Name, "任务内容", jobPlan.Job.Command)
			jobPlan.NextTime = jobPlan.Expr.Next(now) // 更新下次执行时间
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	// 下次调度间隔（最近要执行的任务调度时间 - 当前时间）
	scheduleAfter = (*nearTime).Sub(now)
	return scheduleAfter
}

// startJob 执行任务
func (scheduler *Scheduler) startJob(jobPlan *common.JobSchedulePlan) {
	// 调度 和 执行 是2件事情
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)

	// 执行的任务可能运行很久, 1分钟会调度60次，但是只能执行1次, 防止并发！
	// 如果任务正在执行，跳过本次调度
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		//fmt.Printf("%s,存在尚未退出的携程,跳过执行:", jobPlan.Job.Name)
		return
	}

	// 构建执行状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)

	// 保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	//fmt.Println("执行任务:", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

// handleJobEvent 处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExisted      bool
		err             error
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
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
	case common.JOB_EVENT_KILL: // 强杀任务事件
		// 取消掉Command执行, 判断任务是否在执行中
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // 触发command杀死shell子进程, 任务得到退出
		}
	}
}

// PushJobEvent 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// PushJobResult 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}

// handleJobResult 处理任务结果
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	// todo 后续存入mongodb
	fmt.Printf("执行结果，%#v, 错误%v\n", result.ExecuteInfo.Job.Name, result.Err)
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)
}
