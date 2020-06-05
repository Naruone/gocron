package worker

import (
    "context"
    "go.etcd.io/etcd/clientv3"
    "gocron/common"
    "net"
    "time"
)

var (
    G_register *Register
)

type Register struct {
    client  *clientv3.Client
    kv      clientv3.KV
    lease   clientv3.Lease
    localIp string
}

func InitRegister() (err error) {
    var (
        ipv4   string
        config clientv3.Config
        client *clientv3.Client
        kv     clientv3.KV
        lease  clientv3.Lease
    )
    if ipv4, err = getLocalIP(); err != nil {
        return
    }

    config = clientv3.Config{
        Endpoints:   G_config.EtcdEndPoints,
        DialTimeout: time.Duration(G_config.DialTimeout) * time.Millisecond,
    }

    if client, err = clientv3.New(config); err != nil {
        return
    }

    kv = clientv3.NewKV(client)
    lease = clientv3.NewLease(client)

    G_register = &Register{
        client:  client,
        kv:      kv,
        lease:   lease,
        localIp: ipv4,
    }
    go G_register.keepOnAlive()
    return
}

func (register *Register) keepOnAlive() {
    var (
        regKey        = common.JOB_WORKER_DIR + register.localIp
        leaseResp     *clientv3.LeaseGrantResponse
        ctx           context.Context
        cancelFunc    context.CancelFunc
        keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
        keepAlive     *clientv3.LeaseKeepAliveResponse
        err           error
    )
    for {
        cancelFunc = nil
        if leaseResp, err = register.lease.Grant(context.TODO(), 5); err != nil {
            goto RETRY
        }
        if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseResp.ID); err != nil {
            goto RETRY
        }
        ctx, cancelFunc = context.WithCancel(context.TODO())

        if _, err = register.kv.Put(ctx, regKey, "", clientv3.WithLease(leaseResp.ID)); err != nil {
            goto RETRY
        }
        for {
            select {
            case keepAlive = <-keepAliveChan:
                if keepAlive == nil {
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

// 获取本机网卡IP
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
