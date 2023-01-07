package route

import (
	"crontab/common"
	"crontab/master/service"
	"encoding/json"
	"errors"
	"net/http"
)

// handleJobSave 保存任务
// POST /job/save application/x-www-form-urlencoded job={"name":"job1", "command":"echo hello", "cronExpr":"* * * * * *"}
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
	if postJob == "" {
		err = errors.New("参数不能为空")
		goto ERR
	}
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	// 3.保存到etcd中
	if oldJob, err = service.G_jobServ.SaveJob(&job); err != nil {
		goto ERR
	}

	// 4.响应
	common.ReturnOkJson(resp, oldJob)
	return
ERR:
	common.ReturnFailJson(resp, err)
	return
}

// handleJobDelete 删除任务接口
// POST /job/delete application/x-www-form-urlencoded name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error // interface{}
		name   string
		oldJob *common.Job
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 删除的任务名
	name = req.PostForm.Get("name")

	// etcd中删除任务
	if oldJob, err = service.G_jobServ.DeleteJob(name); err != nil {
		goto ERR
	}

	common.ReturnOkJson(resp, oldJob)
	return
ERR:
	common.ReturnFailJson(resp, err)
	return
}
