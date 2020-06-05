package worker

import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "gocron/common"
    "time"
)

var (
    G_logSink *LogSink
)

type LogSink struct {
    logChan        chan *common.JobLog
    autoCommitChan chan *common.LogBatch
    collection     *mongo.Collection
}

func (logSink *LogSink) Append(jobLog *common.JobLog) {
    select {
    case logSink.logChan <- jobLog: //chan 没满就投入
    default: //否则丢弃日志
    }
}

func (logSink *LogSink) saveLogs(logBatch *common.LogBatch) {
    logSink.collection.InsertMany(context.TODO(), logBatch.Logs)
}

func (logSink *LogSink) writeLoop() {
    var (
        jobLog       *common.JobLog
        logBatch     *common.LogBatch
        timeoutBatch *common.LogBatch
        commitTimer  *time.Timer
    )
    for {
        select {
        case jobLog = <-logSink.logChan:
            if logBatch == nil {
                logBatch = &common.LogBatch{}
                commitTimer = time.AfterFunc(time.Duration(G_config.JobLogCommitTimeout)*time.Millisecond,
                    func(batch *common.LogBatch) func() { //防止在使用logBatch时, 该指针被修改, 所以用参数传入
                        return func() {
                            logSink.autoCommitChan <- batch
                        }
                    }(logBatch))
            }
            logBatch.Logs = append(logBatch.Logs, jobLog)
            if len(logBatch.Logs) >= G_config.JobLogBatchSize {
                logSink.saveLogs(logBatch)
                logBatch = nil
                commitTimer.Stop()
            }
        case timeoutBatch = <-logSink.autoCommitChan:
            if logBatch != timeoutBatch { //不等说明在timeout 提交前被提交了
                continue
            }
            logSink.saveLogs(timeoutBatch)
            logBatch = nil
        }
    }
}

func InitLogSink() (err error) {
    var (
        client        *mongo.Client
        collection    *mongo.Collection
        clientOptions *options.ClientOptions
    )
    clientOptions = options.Client().ApplyURI(G_config.MongodbUri).
        SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout) * time.Second)
    if client, err = mongo.Connect(context.TODO(), clientOptions); err != nil {
        return
    }

    collection = client.Database("cron").Collection("log")
    G_logSink = &LogSink{
        logChan:        make(chan *common.JobLog, 1000),
        autoCommitChan: make(chan *common.LogBatch, 1000),
        collection:     collection,
    }

    go G_logSink.writeLoop()
    return
}
