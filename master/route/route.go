package route

import "net/http"

// InitRouter 创建路由
func InitRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave)     // 保存任务
	mux.HandleFunc("/job/delete", handleJobDelete) //删除任务

	return mux
}
