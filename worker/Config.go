package worker

import (
    "encoding/json"
    "io/ioutil"
)

type Config struct {
    EtcdEndPoints         []string `json:"etcdEndPoints"`
    DialTimeout           int      `json:"dialTimeout"`
    MongodbUri            string   `json:"mongodbUri"`
    MongodbConnectTimeout int      `json:"mongodbConnectTimeout"`
}

var (
    G_config *Config
)

func InitConfig(filename string) (err error) {
    var (
        bytes  []byte
        config Config
    )
    if bytes, err = ioutil.ReadFile(filename); err != nil {
        return
    }

    if err = json.Unmarshal(bytes, &config); err != nil {
        return
    }
    G_config = &config
    return
}
