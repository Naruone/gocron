package master

import (
    "gocron/common"
    "net"
    "net/http"
    "strconv"
    "sync"
    "time"
)

type ApiServer struct {
    httpServer *http.Server
}

var (
    G_apiServer *ApiServer
    sLock       sync.Mutex
)

func InitApiServer() (err error) {
    sLock.Lock()
    defer sLock.Unlock()
    if G_apiServer != nil {
        return
    }
    var (
        httpServer   *http.Server
        mux          *http.ServeMux
        listener     net.Listener
        staticDir    http.Dir
        staticHandle http.Handler
    )

    staticDir = http.Dir(G_config.WebRoot)
    staticHandle = http.FileServer(staticDir)
    mux = http.NewServeMux()

    // 配置路由
    mux.Handle("/", http.StripPrefix("/", staticHandle))
    mux.HandleFunc("/job/save", saveJobHandle)
    mux.HandleFunc("/job/list", jobListHandle)
    mux.HandleFunc("/job/delete", jobDeleteHandle)
    mux.HandleFunc("/job/kill", jobKillHandle)
    mux.HandleFunc("/job/log", jobLogHandle)
    mux.HandleFunc("/worker/list", workerListHandle)

    httpServer = &http.Server{
        ReadTimeout:  time.Duration(G_config.ApiReadTimeOut) * time.Millisecond,
        WriteTimeout: time.Duration(G_config.ApiWriteTimeOut) * time.Millisecond,
        Handler:      mux,
    }
    if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
        return
    }

    G_apiServer = &ApiServer{httpServer: httpServer}

    go httpServer.Serve(listener)
    return
}

// 保存任务 job={"jobName":"xxx","command":"xxx","conExpr":"* * * * *"}
func saveJobHandle(w http.ResponseWriter, r *http.Request) {
    var (
        jobStr string
        job    *common.Job
        oldJob *common.Job
        bytes  []byte
        err    error
    )
    if err = r.ParseForm(); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }
    jobStr = r.PostForm.Get("job")

    if job, err = common.UnpackJob([]byte(jobStr)); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }

    if oldJob, err = G_JobMgr.SaveJob(job); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }
    if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
        w.Write(bytes)
    }
}

func jobListHandle(w http.ResponseWriter, r *http.Request) {
    var (
        bytes   []byte
        jobList []*common.Job
        err     error
    )

    if jobList, err = G_JobMgr.JobList(); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }

    if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
        w.Write(bytes)
    }
}

func jobDeleteHandle(w http.ResponseWriter, r *http.Request) {
    var (
        jobName string
        bytes   []byte
        oldJob  *common.Job
        err     error
    )
    if err = r.ParseForm(); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }
    jobName = r.PostForm.Get("name")
    if oldJob, err = G_JobMgr.JobDelete(jobName); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }
    if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
        w.Write(bytes)
    }
}

func jobKillHandle(w http.ResponseWriter, r *http.Request) {
    var (
        err   error
        name  string
        bytes []byte
    )
    if err = r.ParseForm(); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }
    name = r.PostForm.Get("name")
    if err = G_JobMgr.KillJob(name); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }
    if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
        w.Write(bytes)
    }
}

func jobLogHandle(w http.ResponseWriter, r *http.Request) {
}

func workerListHandle(w http.ResponseWriter, r *http.Request) {
    var (
        bytes      []byte
        err        error
        wokerLists []string
    )
    if err = r.ParseForm(); err != nil {
        if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
            w.Write(bytes)
        }
        return
    }

    wokerLists = G_JobMgr.ListWorkers()
    if bytes, err = common.BuildResponse(0, "success", wokerLists); err == nil {
        w.Write(bytes)
    }
}
