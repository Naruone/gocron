package main

import (
    "flag"
    "gocron/worker"
    "log"
    "runtime"
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

    if err = worker.InitJobMgr(); err != nil {
        log.Fatal("初始Etcd错误", err)
    }
}
