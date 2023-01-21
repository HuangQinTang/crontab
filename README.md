### 分布式定时shell任务
- master服务提供管理后台，可在后台设置任务（cron表达式以及对应要执行的shell命令），任务数据会落盘在etcd中。
- 各个worker节点竞争任务执行。执行过后将日志记录到mongodb中

### master
**配置文件**，/master/config/config.toml

**启动**，项目根目录执行 `go run /cmd/master/main.go`

### worker
**配置文件**，/worker/config/onfig.toml

**启动**，项目根目录执行 `go run /cmd/worker/main.go`