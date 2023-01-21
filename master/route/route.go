package route

import "net/http"

// InitRouter 创建路由
func InitRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/job/save", handleJobSave)       // 保存任务
	mux.HandleFunc("/job/delete", handleJobDelete)   // 删除任务
	mux.HandleFunc("/job/list", handleJobList)       // 任务列表
	mux.HandleFunc("/job/kill", handleJobKill)       // 强制杀死任务
	mux.HandleFunc("/job/log", handleJobLog)         //查看任务执行日志
	mux.HandleFunc("/worker/list", handleWorkerList) //查看worker节点信息

	// 静态文件服务
	staticHandler := http.FileServer(http.Dir("./static"))
	mux.Handle("/", http.StripPrefix("/", staticHandler))
	return mux
}
