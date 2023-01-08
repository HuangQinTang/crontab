package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

// Job 定时任务
type Job struct {
	Name     string `json:"name"`      //任务名
	Command  string `json:"command"`   //shell命令
	CronExpr string `json:"cron_expr"` //cron表达式
}

// UnpackJob 反序列化Job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)

	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// JobEvent Job 变化事件
type JobEvent struct {
	EventType JobEventType //  SAVE, DELETE
	Job       *Job
}

// BuildJobEvent 任务变化事件有2种：1）更新任务 2）删除任务
func BuildJobEvent(eventType JobEventType, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// ExtractJobName 从etcd的key中提取任务名
// /cron/jobs/job10抹掉/cron/jobs/ return job10
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

// JobSchedulePlan 任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cronexpr表达式
	NextTime time.Time            // 下次调度时间
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	// 解析JOB的cron表达式
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	// 生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}
