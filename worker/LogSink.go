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
    logChan    chan *common.JobLog
    collection *mongo.Collection
}

func (logSink *LogSink) Append(jobLog *common.JobLog) {
    logSink.logChan <- jobLog
}

func (logSink *LogSink) saveLog(jobLog *common.JobLog) {
    logSink.collection.InsertOne(context.TODO(), jobLog)
}

func (logSink *LogSink) writeLoop() {
    var (
        jobLog *common.JobLog
    )
    for {
        select {
        case jobLog = <-logSink.logChan:
            logSink.saveLog(jobLog)
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
        logChan:    make(chan *common.JobLog, 1000),
        collection: collection,
    }

    go G_logSink.writeLoop()
    return
}
