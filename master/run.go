package master

import (
	_ "crontab/master/config"
	"crontab/master/route"
	"crontab/master/service"
	"log"
)

func RunMasterServer() {
	var err error

	// 初始化任务管理器
	if err = service.InitJobServ(); err != nil {
		log.Fatal(err.Error())
	}

	// 启动Api http服务
	if err = route.InitApiServer(); err != nil {
		log.Fatal(err.Error())
	}

	select {}
}
