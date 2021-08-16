package core

import (
	"errors"
	"fmt"
	"strings"

	conf "github.com/qqliaoxin/jsonsql/conf"
	"github.com/qqliaoxin/jsonsql/db"
	"github.com/qqliaoxin/jsonsql/logger"
)

type InsterMysqlExecutor struct {
	Table   string        `json:"table"`
	Columns string        `json:"columns"`
	Value   []interface{} `json:"value"`
	Sql     string        `json:"sql"`
	Select  string        `json:"select"`
}

type JsonInsterSqlExecutor struct {
	SqlJon      map[string]*SqlJon              `json:"sqlJon"`         //内关联查询Table结构
	TableInsSql map[string]*InsterMysqlExecutor `json:"tableInsterSql"` //插入内容集
	Data        map[string]interface{}          `json:"data"`           //查询返回数据集
	config      *conf.Config                    //读取配置
	Err         error                           `json:"err"`
}

func NewInsJsonSQL(ctx map[string]interface{}) *JsonSqlExecutor {
	njsql := NewInsterJsonSqlExecutor()
	if njsql.Err == nil {
		njsql.Err = njsql.doInsterSortedMap(ctx)
	}
	return njsql
}

// 创建Inster构造体
func NewInsterJsonSqlExecutor() *JsonSqlExecutor {
	conf := conf.ReadJsonConfig()
	jsonInsSqlExecutor := &JsonSqlExecutor{
		SqlJon:      make(map[string]*SqlJon),
		TableInsSql: make(map[string]*InsterMysqlExecutor),
		TableSql:    make(map[string]*GetMysqlExecutor),
		Data:        make(map[string]interface{}),
	}
	if len(conf.Table) == 0 {
		jsonInsSqlExecutor.Err = errors.New("Config no init Table list!")
		return jsonInsSqlExecutor
	}
	jsonInsSqlExecutor.config = conf
	return jsonInsSqlExecutor
}

// json map 数据分组
func (e *JsonSqlExecutor) doInsterSortedMap(ctx map[string]interface{}) error {
	sqlJon := make(map[string]interface{})
	table := make(map[string]interface{})
	// 数据分组
	for key, value := range ctx {
		if strings.HasSuffix(key, "@") && strings.Index(key, "|") < 0 {
			sqlJon[key] = value
		} else {
			table[key] = value
		}
	}

	// 先生成 @ 引用关联
	if len(sqlJon) > 0 {
		sortedMap(sqlJon, func(k string, v interface{}) {
			// fmt.Printf("jon::Key:%+v\n", k)
			if jon, ok := v.(map[string]interface{}); ok {
				jm := &SqlJon{}
				e.createSql(jon)
				jm.Table = jon
				for k, _ := range jon {
					jm.Sql = e.TableSql[k].Sql
					jm.Params = e.TableSql[k].Params
					delete(e.TableSql, k)
				}
				e.SqlJon[k] = jm //放到内关联组
			}
		})
	}
	// 正常 inster table 生成处理
	if len(table) > 0 {
		err = e.createInsSql(table)
		if err != nil {
			return err
		}
	}

	// 执行查询
	err = e.insterSql()
	if err != nil {
		return err
	}

	return nil
}

// 构造 SQL 开始 ctx 构造的对象, exeSql 是否是要执行sql 语句. sql@ 并不需要执行，只生成sql语句。
func (e *JsonSqlExecutor) createInsSql(ctx map[string]interface{}) error {
	for t, v := range ctx {
		// 构造 table
		err = e.toInsTable(t, v)
		if err != nil {
			return err
		}
		// 构造 where 查询条件
		err = e.toMakeInsWhere(t, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Table 表转化
func (e *JsonSqlExecutor) toInsTable(t string, val interface{}) error {
	m := &InsterMysqlExecutor{}
	//判断元素是否在配置数组当中
	if _, ok := e.config.Table[t]; ok {
		m.Table += fmt.Sprintf(" %s", e.config.Table[t].TableName)
	} else {
		return errors.New(fmt.Sprintf("Table [%s] is no exit;", t))
	}
	e.TableInsSql[t] = m
	return nil
}

//核心 Inster 业务逻辑处理
func (e *JsonSqlExecutor) toMakeInsWhere(t string, val interface{}) error {
	if vals, ok := val.(map[string]interface{}); ok {
		for k, v := range vals {
			if strings.HasPrefix(k, "@") {
				switch k {
				case "@c":
					e.toInsColumn(t, v)
				case "@column":
					e.toInsColumn(t, v)
				case "@v":
					e.toInsValues(t, v)
				case "@values":
					e.toInsValues(t, v)
				default:
					return errors.New(fmt.Sprintf("The [%s] is no exit;", k))
				}
			} else {
				if !strings.HasPrefix(k, "@") {

					if len(e.TableInsSql[t].Value) == 0 {
						mk := make(map[string]interface{})
						mk[k] = v

						e.TableInsSql[t].Select = k
						err = e.createSql(mk)
						if err != nil {
							return err
						}
					}
				}
			}
		}
		// 生成sql语句
		e.toInsSql(t)
	}
	return nil
}

// column 字段处理
func (e *JsonSqlExecutor) toInsColumn(t string, cVal interface{}) error {
	m := e.TableInsSql[t]
	if columnsStr, ok := cVal.(string); ok {
		m.Columns = columnsStr
	}
	return nil
}

func (e *JsonSqlExecutor) toInsValues(t string, val interface{}) error {
	m := e.TableInsSql[t]
	if value, ok := val.([]interface{}); ok {
		m.Value = value
	}
	return nil

}

func (e *JsonSqlExecutor) toInsSql(t string) error {
	m := e.TableInsSql[t]
	m.Sql = "INSERT INTO "
	if m.Table != "" {
		m.Sql += m.Table
	}

	if len(m.Columns) > 0 {
		m.Sql += fmt.Sprintf(" (%s)", m.Columns)
	}

	if len(m.Value) > 0 {
		m.Sql += " VALUES("
		for i := 0; i < len(m.Value); i++ {
			if i == len(m.Value)-1 {
				m.Sql += "?"
			} else {
				m.Sql += "?,"
			}

		}
		m.Sql += ")"
	} else {
		m.Value = e.TableSql[e.TableInsSql[t].Select].Params
		m.Sql += " " + e.TableSql[e.TableInsSql[t].Select].Sql
	}
	return nil
}

// 执行查询
func (e *JsonSqlExecutor) insterSql() error {
	for table, _ := range e.TableInsSql {
		m := e.TableInsSql[table]
		if e.config.Debug {
			logger.Info("*************************************************************************")
			logger.Debugf("values: %v", m.Value)
			logger.Debugf("%s", m.Sql)
			logger.Info("*************************************************************************")
		}
		data, err := db.Inster(m.Sql, m.Value...)
		if err != nil {
			return errors.New(fmt.Sprintf("insert err: %s", err.Error()))
		}
		e.Data[table] = data
	}
	return nil
}
