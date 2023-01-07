package service

import (
	"context"
	"crontab/common"
	"crontab/master/config"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// G_jobServ 任务服务，将任务信息维护在etcd中
var G_jobServ *JobServ

type JobServ struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

// InitJobServ 初始化任务服务实例
func InitJobServ() (err error) {
	var client *clientv3.Client

	// 连接etcd
	if client, err = clientv3.New(clientv3.Config{
		Endpoints:   config.G_settings.Etcd.EndPoints,
		DialTimeout: time.Millisecond * time.Duration(config.G_settings.Etcd.DialTimeout),
	}); err != nil {
		return err
	}

	G_jobServ = &JobServ{
		client: client,
		kv:     clientv3.NewKV(client),
		lease:  clientv3.NewLease(client),
	}

	return nil
}

// SaveJob 保存任务
func (jm *JobServ) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobkey   string
		jobValue []byte
		putResp  *clientv3.PutResponse
	)

	jobkey = fmt.Sprintf("%s%s", common.JOB_SAVE_DIR, job.Name)
	if jobValue, err = json.Marshal(job); err != nil {
		return nil, err
	}

	// 保存至任务数据到etcd中，同时返回该key到旧值(上个版本)
	if putResp, err = jm.kv.Put(context.TODO(), jobkey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return nil, err
	}
	if putResp.PrevKv != nil {
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJob); err != nil {
			err = nil //旧值序列化失败也无所谓，put成功即返回nil
			return
		}
	}
	return oldJob, nil
}

// DeleteJob 删除任务
func (jm *JobServ) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	// etcd中保存任务的key
	jobKey = fmt.Sprintf("%s%s", common.JOB_SAVE_DIR, name)

	// 从etcd中删除它
	if delResp, err = jm.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	// 返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		// 解析一下旧值, 返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return oldJob, nil
}

// ListJobs 列举任务
func (jobMgr *JobServ) ListJobs() (jobList []*common.Job, err error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)

	// 任务保存的目录
	dirKey = common.JOB_SAVE_DIR

	// 获取目录下所有任务信息
	if getResp, err = jobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}

	// 初始化数组空间
	jobList = make([]*common.Job, 0)
	// len(jobList) == 0

	// 遍历所有任务, 进行反序列化
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return jobList, nil
}

// KillJob 杀死任务
func (jobMgr *JobServ) KillJob(name string) (err error) {
	// 更新一下key=/cron/killer/任务名
	var (
		killerKey      string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	// worker会监听到该key会杀死对应任务
	killerKey = common.JOB_KILLER_DIR + name

	// 让worker监听到一次put操作, 创建一个租约让其稍后自动过期即可
	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	// 租约ID
	leaseId = leaseGrantResp.ID

	// 设置killer标记
	if _, err = jobMgr.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
