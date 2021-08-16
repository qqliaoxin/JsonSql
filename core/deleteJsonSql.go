package core

import (
	"errors"
	"fmt"
	"strings"

	conf "github.com/qqliaoxin/jsonsql/conf"
	"github.com/qqliaoxin/jsonsql/db"
	"github.com/qqliaoxin/jsonsql/logger"
)

type DeleteMysqlExecutor struct {
	Table  string        `json:"table"`
	Where  string        `json:"where"`
	Sql    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

type JsonDeleteSqlExecutor struct {
	TableDeleteSql map[string]*DeleteMysqlExecutor `json:"tableDeleteSql"` //查询内容集
	Data           map[string]interface{}          `json:"data"`           //查询返回数据集
	config         *conf.Config                    //读取配置
	Err            error                           `json:"err"`
}

func NewDeleteJsonSQL(ctx map[string]interface{}) *JsonSqlExecutor {
	njsql := NewDeleteJsonSqlExecutor()
	if njsql.Err == nil {
		njsql.Err = njsql.doDeleteSortedMap(ctx)
	}
	return njsql
}

// 创建Delete构造体
func NewDeleteJsonSqlExecutor() *JsonSqlExecutor {
	conf := conf.ReadJsonConfig()
	jsonDeleteSqlExecutor := &JsonSqlExecutor{
		TableDeleteSql: make(map[string]*DeleteMysqlExecutor),
		TableSql:       make(map[string]*GetMysqlExecutor),
		Data:           make(map[string]interface{}),
	}
	if len(conf.Table) == 0 {
		jsonDeleteSqlExecutor.Err = errors.New("Config no init Table list!")
		return jsonDeleteSqlExecutor
	}
	jsonDeleteSqlExecutor.config = conf
	return jsonDeleteSqlExecutor
}

// json map 数据分组
func (e *JsonSqlExecutor) doDeleteSortedMap(ctx map[string]interface{}) error {
	// 正常 Delete table 生成处理
	if len(ctx) > 0 {
		err = e.createDeleteSql(ctx)
		if err != nil {
			return err
		}
	}

	// 执行查询
	err = e.DeleteSql()
	if err != nil {
		return err
	}

	return nil
}

// 构造 SQL 开始 ctx 构造的对象, exeSql 是否是要执行sql 语句. sql@ 并不需要执行，只生成sql语句。
func (e *JsonSqlExecutor) createDeleteSql(ctx map[string]interface{}) error {
	for t, v := range ctx {
		// 构造 table
		err = e.toDeleteTable(t, v)
		if err != nil {
			return err
		}
		// 构造 where 查询条件
		err = e.toMakeDeleteWhere(t, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Table 表转化
func (e *JsonSqlExecutor) toDeleteTable(t string, val interface{}) error {
	m := &DeleteMysqlExecutor{}
	//判断元素是否在配置数组当中
	if _, ok := e.config.Table[t]; ok {
		m.Table += fmt.Sprintf(" %s", e.config.Table[t].TableName)
	} else {
		return errors.New(fmt.Sprintf("Table [%s] is no exit;", t))
	}
	e.TableDeleteSql[t] = m
	return nil
}

//核心 Delete 业务逻辑处理
func (e *JsonSqlExecutor) toMakeDeleteWhere(t string, val interface{}) error {
	if vals, ok := val.(map[string]interface{}); ok {
		for k, v := range vals {
			if strings.HasPrefix(k, "@") {
				switch k {
				case "@w":
					e.toDeleteWhere(t, v)
				case "@where":
					e.toDeleteWhere(t, v)
				default:
					return errors.New(fmt.Sprintf("The [%s] is no exit;", k))
				}
			}
		}
		// 生成sql语句
		e.toDeleteSql(t)
	}
	return nil
}

func (e *JsonSqlExecutor) toDeleteWhere(t string, wVal interface{}) error {
	w := e.TableDeleteSql[t]
	andFirst := true
	where := make(map[string]interface{})
	if mw, ok := wVal.(map[string]interface{}); ok {
		for k, v := range mw {
			// 优先处理没有 $ | 的条件,解决map 无序不好控制问题。
			if !strings.HasPrefix(k, "&") && !strings.HasPrefix(k, "|") {
				if andFirst {
					w.Where += fmt.Sprintf(" %s", e.makeDeleteWhere(t, k, v))
					andFirst = false
				} else {
					w.Where += fmt.Sprintf(" AND %s", e.makeDeleteWhere(t, k, v))
				}
			} else {
				where[k] = v
			}
		}
	}

	// 其它的条件处理
	for k, v := range where {
		if strings.HasPrefix(k, "&") {
			w.Where += fmt.Sprintf(" AND %s", e.makeDeleteWhere(t, k[1:], v))
		} else if strings.HasPrefix(k, "|") {
			w.Where += fmt.Sprintf(" OR %s", e.makeDeleteWhere(t, k[1:], v))
		}
	}

	return nil
}

// where column 逻辑处理 Key
func (e *JsonSqlExecutor) makeDeleteWhere(t string, k string, wcVal interface{}) string {
	m := e.TableDeleteSql[t]
	if strings.HasSuffix(k, "@") {
		// key 处理
		a := strings.Index(k[:len(k)-1], "/")
		if a > 0 {
			return fmt.Sprintf(" %s.%s = %s", strings.ToLower(k[:a]), k[a+1:len(k)-1], e.makeWhereToTableORColumnt(t, wcVal))
		} else if _sql_t, ok := e.SqlJon[wcVal.(string)]; ok {
			m.Params = append(m.Params, _sql_t.Params...)
			return fmt.Sprintf(" %s IN (%s)", k[:len(k)-1], _sql_t.Sql)
		} else {
			return fmt.Sprintf(" %s = %s", k[:len(k)-1], e.makeWhereToTableORColumnt(t, wcVal))
		}
	} else if strings.HasSuffix(k, "[]") {
		if wcV, ok := wcVal.(string); ok {
			if strings.HasSuffix(wcV, "@") {
				if _sql_t, ok := e.SqlJon[wcV]; ok {
					m.Params = append(m.Params, _sql_t.Params...)
					return fmt.Sprintf(" %s IN (%s)", k[:len(k)-2], _sql_t.Sql)
				}
			}
		} else {
			inStr := strings.Join(parseIntListString(wcVal), ",")
			return fmt.Sprintf(" %s IN (%s)", k[:len(k)-2], inStr)
		}
	} else if strings.HasSuffix(k, "%") {
		m.Params = append(m.Params, wcVal)
		return fmt.Sprintf(" %s LIKE '%s?%s'", k[:len(k)-1], "%", "%")
	} else if strings.HasSuffix(k, ">") {
		m.Params = append(m.Params, wcVal)
		return fmt.Sprintf(" %s > ?", k[:len(k)-1])
	} else if strings.HasSuffix(k, ">=") {
		m.Params = append(m.Params, wcVal)
		return fmt.Sprintf(" %s >= ?", k[:len(k)-2])
	} else if strings.HasSuffix(k, "<") {
		m.Params = append(m.Params, wcVal)
		return fmt.Sprintf(" %s < ?", k[:len(k)-1])
	} else if strings.HasSuffix(k, "<=") {
		m.Params = append(m.Params, wcVal)
		return fmt.Sprintf(" %s <= ?", k[:len(k)-1])
	} else {
		m.Params = append(m.Params, wcVal)
		return fmt.Sprintf(" %s = ?", k)
	}
	return ""
}

func (e *JsonSqlExecutor) toDeleteSql(t string) error {
	m := e.TableDeleteSql[t]
	m.Sql = "DELETE FROM "
	if m.Table != "" {
		m.Sql += m.Table
	}

	if len(m.Where) > 0 {
		m.Sql += " WHERE"
		m.Sql += m.Where
	}
	return nil
}

// 执行查询
func (e *JsonSqlExecutor) DeleteSql() error {
	for table, _ := range e.TableDeleteSql {
		m := e.TableDeleteSql[table]
		if e.config.Debug {
			logger.Info("*************************************************************************")
			logger.Debugf("params: %v", m.Params...)
			logger.Debugf("%s", m.Sql)
			logger.Info("*************************************************************************")
		}
		data, err := db.Delete(m.Sql, m.Params...)
		if err != nil {
			return errors.New(fmt.Sprintf("Delete err: %s", err.Error()))
		}
		e.Data[table] = data
	}
	return nil
}
