package handler

import (
	"encoding/json"
	"net/http"

	"github.com/qqliaoxin/jsonsql/logger"
)

func handleCors(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "http://jsonsql.cn")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "content-type")
	w.Header().Add("Access-Control-Request-Method", "POST")
}

type QueryContext struct {
	even int
	req  map[string]interface{}
	code int
	data map[string]interface{}
	err  error
}

func handleRequestJson(even int, data []byte, w http.ResponseWriter) {
	handleCors(w)
	logger.Infof("request: %s", string(data))
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(data, &bodyMap); err != nil {
		logger.Error("请求 JSON 格式有问题: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	NewQueryContext(even, bodyMap).response(w)
}

func NewQueryContext(even int, bodyMap map[string]interface{}) *QueryContext {
	return &QueryContext{
		code: http.StatusOK,
		req:  bodyMap,
		even: even,
	}
}

//json返回值
func (c *QueryContext) response(w http.ResponseWriter) {
	if c.err == nil {
		switch c.even {
		case 1:
			c.doQuery()
		case 2:
			c.doInster()
		case 3:
			c.doUpdate()
		case 4:
			c.doDelete()
		}
	}
	w.WriteHeader(http.StatusOK)
	dataMap := make(map[string]interface{})
	dataMap["code"] = c.code
	if c.err != nil {
		dataMap["message"] = c.err.Error()
		dataMap["code"] = 500
	} else {
		for k, v := range c.data {
			dataMap[k] = v
		}
	}
	if respBody, err := json.Marshal(dataMap); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		//logger.Debugf("返回数据 %s", string(respBody))
		if _, err = w.Write(respBody); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
