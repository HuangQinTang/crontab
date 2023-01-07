package service

import (
	"context"
	"crontab/common"
	"crontab/master/config"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// G_jobServ 任务服务，将任务信息维护在etcd中
var G_jobServ *JobServ

const (
	jobKeyPrefix = "/cron/jobs/"
)

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

	jobkey = fmt.Sprintf("%s%s", jobKeyPrefix, job.Name)
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
