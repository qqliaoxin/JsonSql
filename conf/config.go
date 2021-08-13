package conf

import (
	"encoding/json"
	"io/ioutil"
)

//定义配置文件解析后的结构
type TableConfig struct {
	TableName string
	Access    string
}

type Config struct {
	Table   map[string]TableConfig
	Debug   bool
	Explain bool
}

type DbConfig struct {
	DataBase string
	UserName string
	PassWord string
	Host     string
	Port     int
}

type JsonStruct struct {
}

func NewJsonStruct() *JsonStruct {
	return &JsonStruct{}
}

// 读取数据库json配置文件
func ReadDBConfig() *DbConfig {
	jsonParse := NewJsonStruct()
	v := DbConfig{}
	//下面使用的是相对路径，config.json文件和main.go文件处于同一目录下
	jsonParse.Load("./mysql.json", &v)
	return &v
}

// 读取表配置映射json配置文件
func ReadJsonConfig() *Config {
	jsonParse := NewJsonStruct()
	v := Config{}
	//下面使用的是相对路径，config.json文件和main.go文件处于同一目录下
	jsonParse.Load("./config.json", &v)
	return &v
}

func (jst *JsonStruct) Load(filename string, v interface{}) {
	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, v)
	if err != nil {
		return
	}
}
