package worker

import (
	_ "crontab/worker/config"
	"crontab/worker/service"
	"log"
)

func RunWorker() {
	var err error

	// 4.日志落盘服务，将执行器处理任务的结果记录到mongodb中
	if err = service.InitLogSink(); err != nil {
		log.Fatal(err.Error())
	}

	// 3.任务执行器，启动一个协程，执行调度器传过来的任务
	service.InitExecutor()

	// 2.任务调度器，计算收到的任务执行时间，到期调度到执行器上执行
	service.InitJobScheduler()

	// 1.任务管理服务，该实例会监听etcd中的任务数据，并传到调度器中进行调度
	if err = service.InitJobMgr(); err != nil {
		log.Fatal(err.Error())
	}

	select {}
}
