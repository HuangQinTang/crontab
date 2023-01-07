package route

import "net/http"

// InitRouter 创建路由
func InitRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave)     // 保存任务
	mux.HandleFunc("/job/delete", handleJobDelete) // 删除任务
	mux.HandleFunc("/job/list", handleJobList)     // 任务列表
	mux.HandleFunc("/job/kill", handleJobKill)     // 强制杀死任务

	return mux
}
