package handler

import (
	"io/ioutil"
	"net/http"

	"github.com/qqliaoxin/jsonsql/core"
	"github.com/qqliaoxin/jsonsql/logger"
)

func InsterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleCors(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	if data, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error("请求参数有问题: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		handleRequestJson(2, data, w)
	}
}

//执行 JsonSql 核心处理逻辑
func (c *QueryContext) doInster() {
	m := core.NewInsJsonSQL(c.req)
	c.err = m.Err
	c.data = m.Data
}
