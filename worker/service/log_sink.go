package service

import (
	"context"
	"crontab/common"
	"crontab/worker/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var (
	// G_logSink 单例
	G_logSink *LogSink
)

// LogSink mongodb存储日志
type LogSink struct {
	client         *mongo.Client // mongodb客户端
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

// saveLogs 批量写入日志
func (logSink *LogSink) saveLogs(batch *common.LogBatch) {
	logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
}

// writeLoop 日志存储协程
func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch // 当前的批次
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch // 超时批次
	)

	for {
		select {
		case log = <-logSink.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				// 让这个批次超时自动提交(给1秒的时间）
				commitTimer = time.AfterFunc(
					time.Duration(config.G_settings.Log.AutoCommit)*time.Millisecond,
					func(batch *common.LogBatch) func() { // 避免闭包，这里值拷贝logBatch
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}

			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			// 如果批次满了, 就立即发送
			if len(logBatch.Logs) >= config.G_settings.Log.BatchSize {
				// 发送日志
				logSink.saveLogs(logBatch)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan: // 过期的批次（定时任务触发）
			// 判断过期批次是否仍旧是当前的批次（因为定时任务是time包底层协程负责触发时，有可能批次刚好满了已被提交，这里我们重复判断下避免重复提交）
			if timeoutBatch != logBatch {
				continue // 跳过已经被提交的批次
			}
			// 把批次写入到mongo中
			logSink.saveLogs(timeoutBatch)
			// 清空logBatch
			logBatch = nil
		}
	}
}

func InitLogSink() (err error) {
	G_logSink = &LogSink{
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}
	timeout := time.Duration(config.G_settings.Mongodb.TimeOut) * time.Millisecond

	// 建立mongodb连接
	if G_logSink.client, err = mongo.Connect(context.TODO(), &options.ClientOptions{
		Hosts:   []string{config.G_settings.Mongodb.Host},
		Timeout: &timeout,
	}); err != nil {
		return err
	}

	G_logSink.logCollection = G_logSink.client.Database("cron").Collection("log")

	// 启动一个mongodb处理协程
	go G_logSink.writeLoop()
	return
}

// Append 发送日志
func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
		// 队列满了就丢弃
	}
}
