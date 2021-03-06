package main

import (
    "flag"
    "gocron/worker"
    "log"
    "runtime"
    "time"
)

func initEnv() {
    runtime.GOMAXPROCS(runtime.NumCPU())
}

var (
    configFile string
)

func initFlag() {
    flag.StringVar(&configFile, "config", "./config.json", "配置文件路径")
    flag.Parse()
}

func main() {
    var (
        err error
    )
    //初始化允许时
    initEnv()
    initFlag()

    //初始化配置
    if err = worker.InitConfig(configFile); err != nil {
        log.Fatal("初始化配置错误", err)
    }

    if err = worker.InitRegister(); err != nil {
        log.Fatal("注册worker节点失败", err)
    }

    if err = worker.InitLogSink(); err != nil {
        log.Fatal("启动日志记录服务", err)
    }
    if err = worker.InitExecutor(); err != nil {
        log.Fatal("初始化执行器错误", err)
    }

    if err = worker.InitScheduler(); err != nil {
        log.Fatal("初始化调度器错误", err)
    }

    if err = worker.InitJobMgr(); err != nil {
        log.Fatal("初始化Etcd错误", err)
    }

    for {
        time.Sleep(1 * time.Second)
    }
}
