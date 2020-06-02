package master

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeOut  int      `json:"apiReadTimeout"`
	ApiWriteTimeOut int      `json:"apiWriteTimeout"`
	EtcdEndPoints   []string `json:"etcdEndPoints"`
	DialTimeout     int      `json:"dialTimeout"`
	WebRoot         string   `json:"webRoot"`
}

var (
	cLock    sync.Mutex
	G_config *Config
)

func InitConfig(filename string) (err error) {
	cLock.Lock()
	defer cLock.Unlock()
	if G_config != nil {
		return
	}
	var (
		content []byte
		config  Config
	)
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	if err = json.Unmarshal(content, &config); err != nil {
		return
	}
	G_config = &config
	return
}
