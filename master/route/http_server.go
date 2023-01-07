package route

import (
	"crontab/master/config"
	"fmt"
	"net"
	"net/http"
	"time"
)

var (
	// G_apiServer 单例对象
	G_apiServer *ApiServer
)

// ApiServer 任务到http接口
type ApiServer struct {
	httpServer *http.Server
}

// InitApiServer 启动http服务
func InitApiServer() (err error) {
	var (
		listener   net.Listener
		httpServer *http.Server
	)

	// 监听端口
	if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", config.G_settings.Server.Port)); err != nil {
		return err
	}

	// 启动服务
	httpServer = &http.Server{
		ReadTimeout:  time.Millisecond * time.Duration(config.G_settings.Server.ReadTimeout),
		WriteTimeout: time.Millisecond * time.Duration(config.G_settings.Server.WriteTimeout),
		Handler:      InitRouter(),
	}
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listener)
	fmt.Println("【server】listen port to " + fmt.Sprintf("%d", config.G_settings.Server.Port))
	return nil
}
