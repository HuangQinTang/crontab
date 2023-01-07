package route

import "net/http"

// InitRouter 创建路由
func InitRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave) // 保存任务

	return mux
}