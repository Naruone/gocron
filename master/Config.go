package master

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "os"
    "strings"
)

type Config struct {
    ApiPort               int      `json:"apiPort"`
    ApiReadTimeOut        int      `json:"apiReadTimeout"`
    ApiWriteTimeOut       int      `json:"apiWriteTimeout"`
    EtcdEndPoints         []string `json:"etcdEndPoints"`
    DialTimeout           int      `json:"dialTimeout"`
    WebRoot               string   `json:"webRoot"`
    MongodbUri            string   `json:"mongodbUri"`
    MongodbConnectTimeout int      `json:"mongodbConnectTimeout"`
}

var (
    G_config *Config
)

func InitConfig(filename string) (err error) {
    var (
        content   []byte
        config    Config
        envConfig = make([]string, 0)
        mainPath  string
    )
    if content, err = ioutil.ReadFile(filename); err != nil {
        return
    }

    if err = json.Unmarshal(content, &config); err != nil {
        return
    }

    for _, k := range []string{"ETCD_N1", "ETCD_N2", "ETCD_N3"} {
        etcdStr := os.Getenv(k)
        if etcdStr != "" {
            envConfig = append(envConfig, etcdStr)
        }
    }
    if len(envConfig) == 0 {
        err = errors.New("etcd 配置错误, 请至少配置ETCD_N1, ETCD_N2, ETCD_N3 其中一个")
        return
    }
    if mongoConf := os.Getenv("MONGODB"); mongoConf != "" {
        config.MongodbUri = "mongodb://" + mongoConf
    } else {
        err = errors.New("mongodb 配置错误, 请配置 MONGODB")
        return
    }
    if mainPath, err = os.Getwd(); err != nil {
        return
    }
    config.WebRoot = mainPath + strings.TrimLeft(config.WebRoot, ".")
    config.EtcdEndPoints = envConfig
    G_config = &config
    return
}
