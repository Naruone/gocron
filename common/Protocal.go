package common

import "encoding/json"

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func BuildResponse(code int, msg string, data interface{}) (resp []byte, err error) {
	var response Response

	response = Response{
		Errno: code,
		Msg:   msg,
		Data:  data,
	}

	resp, err = json.Marshal(response)
	return
}

//json 字符串转 Job
func UnpackJob(value []byte) (ret *Job, err error) {
	job := &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}
