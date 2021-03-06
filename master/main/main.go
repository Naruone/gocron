package main

import (
    "flag"
    "gocron/master"
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
    if err = master.InitConfig(configFile); err != nil {
        log.Fatal("初始化配置错误", err)
    }

    if err = master.InitJobMgr(); err != nil {
        log.Fatal("初始Etcd错误", err)
    }

    if err = master.InitLogMgr(); err != nil {
        log.Fatal("初始日志管理器错误", err)
    }

    //初始化http
    if err = master.InitApiServer(); err != nil {
        log.Fatal("初始化Web服务错误", err)
    }

    for {
        time.Sleep(2 * time.Second)
    }
}
