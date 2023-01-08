package common

const (
	// JOB_SAVE_DIR 任务保存目录
	JOB_SAVE_DIR = "/cron/jobs/"

	// JOB_KILLER_DIR 任务强杀目录
	JOB_KILLER_DIR = "/cron/killer/"

	// JOB_LOCK_DIR 任务锁目录
	JOB_LOCK_DIR = "/cron/lock/"

	// JOB_WORKER_DIR 服务注册目录
	JOB_WORKER_DIR = "/cron/workers/"

	// JOB_EVENT_SAVE 保存任务事件
	JOB_EVENT_SAVE JobEventType = 1

	// JOB_EVENT_DELETE 删除任务事件
	JOB_EVENT_DELETE JobEventType = 2

	// JOB_EVENT_KILL 强杀任务事件
	JOB_EVENT_KILL = 3
)

// JobEventType 任务事件类型，对应 mvccpb.Event_EventType
type JobEventType int32
