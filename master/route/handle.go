package route

import (
	"crontab/common"
	"crontab/master/service"
	"encoding/json"
	"errors"
	"github.com/gorhill/cronexpr"
	"net/http"
	"strconv"
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
	if _, err = cronexpr.Parse(job.CronExpr); err != nil {
		err = errors.New("cron表达式解析失败，请检查格式")
		goto ERR
	}

	// 3.保存到etcd中
	if oldJob, err = service.G_jobMgr.SaveJob(&job); err != nil {
		goto ERR
	}

	// 4.响应
	common.ReturnOkJson(resp, oldJob)
	return
ERR:
	common.ReturnFailJson(resp, err.Error())
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
	if oldJob, err = service.G_jobMgr.DeleteJob(name); err != nil {
		goto ERR
	}

	common.ReturnOkJson(resp, oldJob)
	return
ERR:
	common.ReturnFailJson(resp, err.Error())
	return
}

// handleJobList 列举所有crontab任务
// GET /job/list
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList []*common.Job
		err     error
	)

	// 获取任务列表
	if jobList, err = service.G_jobMgr.ListJobs(); err != nil {
		goto ERR
	}

	common.ReturnOkJson(resp, jobList)
	return

ERR:
	common.ReturnFailJson(resp, err.Error())
	return
}

// handleJobKill 强制杀死某个任务
// POST /job/kill application/x-www-form-urlencoded name=job1
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err  error
		name string
	)

	// 解析POST表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 要杀死的任务名
	name = req.PostForm.Get("name")

	// 杀死任务
	if err = service.G_jobMgr.KillJob(name); err != nil {
		goto ERR
	}
	common.ReturnOkJson(resp, nil)
	return
ERR:
	common.ReturnFailJson(resp, err.Error())
	return
}

// handleJobLog 查询任务日志
func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		name       string // 任务名字
		skipParam  string // 从第几条开始
		limitParam string // 返回多少条
		skip       int
		limit      int
		logArr     []*common.JobLog
	)

	// 解析GET参数
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 获取请求参数 /job/log?name=job10&skip=0&limit=10
	name = req.Form.Get("name")
	skipParam = req.Form.Get("skip")
	limitParam = req.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = service.G_logMgr.ListLog(name, int64(skip), int64(limit)); err != nil {
		goto ERR
	}

	// 正常应答
	common.ReturnOkJson(resp, logArr)
	return

ERR:
	common.ReturnFailJson(resp, err.Error())
	return
}

// handleWorkerList 获取健康worker节点列表
func handleWorkerList(resp http.ResponseWriter, req *http.Request) {
	var (
		workerArr []string
		err       error
	)

	if workerArr, err = service.G_jobMgr.ListWorkers(); err != nil {
		goto ERR
	}

	// 正常应答
	common.ReturnOkJson(resp, workerArr)
	return

ERR:
	common.ReturnFailJson(resp, err.Error())
	return
}
