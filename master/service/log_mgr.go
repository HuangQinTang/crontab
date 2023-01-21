package service

import (
	"context"
	"crontab/common"
	"crontab/master/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	G_logMgr *LogMgr
)

// LogMgr mongodb日志管理
type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

// InitLogMgr 初始化日志服务
func InitLogMgr() (err error) {
	G_logMgr = &LogMgr{}
	timeout := time.Duration(config.G_settings.Mongodb.TimeOut) * time.Millisecond

	// 建立mongodb连接
	if G_logMgr.client, err = mongo.Connect(context.TODO(), &options.ClientOptions{
		Hosts:   []string{config.G_settings.Mongodb.Host},
		Timeout: &timeout,
	}); err != nil {
		return err
	}

	G_logMgr.logCollection = G_logMgr.client.Database("cron").Collection("log")

	return nil
}

// ListLog 查看任务日志
func (logMgr *LogMgr) ListLog(name string, skip int64, limit int64) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
	)

	// len(logArr)
	logArr = make([]*common.JobLog, 0)

	// 过滤条件
	filter = &common.JobLogFilter{JobName: name}

	// 按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	// 查询

	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, &options.FindOptions{
		Sort: logSort, Skip: &skip, Limit: &limit,
	}); err != nil {
		return nil, err
	}

	// 延迟释放游标
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		// 反序列化BSON
		if err = cursor.Decode(jobLog); err != nil {
			continue // 有日志不合法
		}

		logArr = append(logArr, jobLog)
	}
	return
}
