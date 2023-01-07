package route

import (
	"crontab/common"
	"crontab/master/service"
	"encoding/json"
	"net/http"
)

// handleJobSave 保存任务
// POST application/x-www-form-urlencoded job={"name":"job1", "command":"echo hello", "cronExpr":"* * * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
	)

	// 1.解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	// 2.获取表单中到job字段，并反序列化
	postJob = req.PostForm.Get("job")
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	// 3.保存到etcd中
	if oldJob, err = service.G_jobServ.SaveJob(&job); err != nil {
		goto ERR
	}

	// 4.响应
	common.NewResponse(resp, common.SetResp(common.Success, "请求成功", oldJob)).ReturnJson()
	return
ERR:
	common.NewResponse(resp, common.SetResp(common.Fail, "请求失败,"+err.Error(), nil)).ReturnJson()
	return
}
