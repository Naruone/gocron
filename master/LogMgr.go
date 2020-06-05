package master

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "gocron/common"
    "time"
)

var (
    G_logMgr *LogMgr
)

type LogMgr struct {
    collection *mongo.Collection
}

func InitLogMgr() (err error) {
    var (
        client        *mongo.Client
        clientOptions *options.ClientOptions
    )
    clientOptions = options.Client().ApplyURI(G_config.MongodbUri).
        SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout) * time.Second)
    if client, err = mongo.Connect(context.TODO(), clientOptions); err != nil {
        return
    }

    G_logMgr = &LogMgr{
        collection: client.Database("cron").Collection("log"),
    }
    return
}

func (logMgr *LogMgr) ListLog(jobName string, skip int64, limit int64) (logList []*common.JobLog) {
    var (
        findopt *options.FindOptions
        cursor  *mongo.Cursor
        err     error
        jobLog  *common.JobLog
    )
    logList = make([]*common.JobLog, 0)

    findopt = options.Find()
    findopt.SetLimit(limit)
    findopt.SetSkip(skip)
    findopt.SetSort(bson.D{{"start_time", -1}})

    if cursor, err = logMgr.collection.Find(context.TODO(), bson.D{{"job_name", jobName}}, findopt); err != nil {
        return
    }
    defer cursor.Close(context.TODO())
    for cursor.Next(context.TODO()) {
        jobLog = &common.JobLog{}
        if err = cursor.Decode(jobLog); err == nil {
            logList = append(logList, jobLog)
        }
    }
    return
}
