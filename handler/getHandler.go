package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"jsonSql/core"
	"jsonSql/logger"
)

func GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		//logger.Infof("%v", r.Header)
		handleCors(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	if data, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error("[Body]请求参数有问题: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		handleRequestJson(data, w)
	}
}

func handleCors(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "http://jsonsql.cn")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Headers", "content-type")
	w.Header().Add("Access-Control-Request-Method", "POST")
}

type QueryContext struct {
	req  map[string]interface{}
	code int
	data map[string]interface{}
	err  error
}

func handleRequestJson(data []byte, w http.ResponseWriter) {
	handleCors(w)
	logger.Infof("request: %s", string(data))
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(data, &bodyMap); err != nil {
		logger.Error("请求 JSON 格式有问题: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	NewQueryContext(bodyMap).response(w)
}

func NewQueryContext(bodyMap map[string]interface{}) *QueryContext {
	return &QueryContext{
		code: http.StatusOK,
		req:  bodyMap,
	}
}

func (c *QueryContext) response(w http.ResponseWriter) {
	if c.err == nil {
		c.doQuery()
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

//执行 JsonSql 核心处理逻辑
func (c *QueryContext) doQuery() {
	m := core.NewJsonSQL(c.req)
	c.err = m.Err
	c.data = m.Data
}
