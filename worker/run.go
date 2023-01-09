package worker

import (
	_ "crontab/worker/config"
	"crontab/worker/service"
	"log"
)

func RunWorker() {
	var err error

	// 初始化任务调度器，计算收到的任务数据执行时间，到期调度执行
	service.InitJobScheduler()

	// 初始化任务管理服务，该实例会监听etcd中的任务数据，并传到调度器中进行调度
	if err = service.InitJobMgr(); err != nil {
		log.Fatal(err.Error())
	}

	select {}
}
