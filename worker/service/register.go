package service

import (
	"context"
	"crontab/common"
	"crontab/worker/config"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net"
	"time"
)

// Register 注册节点到etcd： /cron/workers/IP地址
type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string // 本机IP
}

var (
	G_register *Register
)

// getLocalIP 获取本机网卡IP
func getLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet // IP地址
		isIpNet bool
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	// 取第一个非lo的网卡IP
	for _, addr = range addrs {
		// 这个网络地址是IP地址: ipv4, ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			// 跳过IPV6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String() // 192.168.1.1
				return
			}
		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

// keepOnline 注册到/cron/workers/IP, 并自动续租
func (register *Register) keepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		// 注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP

		cancelFunc = nil

		// 创建租约
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		// 自动续租
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到etcd
		if _, err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		// 处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

// InitRegister 服务注册
func InitRegister() (err error) {
	var (
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIp string
	)

	// 建立连接
	if client, err = clientv3.New(clientv3.Config{
		Endpoints:   config.G_settings.Etcd.EndPoints,                                     // 集群地址
		DialTimeout: time.Duration(config.G_settings.Etcd.DialTimeout) * time.Millisecond, // 连接超时
	}); err != nil {
		return err
	}

	// 本机IP
	if localIp, err = getLocalIP(); err != nil {
		return err
	}
	//fmt.Println(localIp)

	// 得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIp,
	}

	// 服务注册
	go G_register.keepOnline()

	return nil
}
