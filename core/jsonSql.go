package core

import (
	"errors"
	"fmt"
	"strings"

	conf "github.com/qqliaoxin/jsonsql/conf"
	"github.com/qqliaoxin/jsonsql/db"
	"github.com/qqliaoxin/jsonsql/logger"
)

var (
	err error
)

type MysqlExecutor struct {
	Table   string        `json:"table"`
	Columns []string      `json:"columns"`
	Where   string        `json:"where"`
	Params  []interface{} `json:"params"`
	Order   string        `json:"order"`
	Group   string        `json:"group"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	Sql     string        `json:"sql"`
}

type SqlJon struct {
	Table  map[string]interface{} `json:"table"`
	Sql    string                 `json:"sql"`
	Params []interface{}          `json:"params"`
}

type JsonSqlExecutor struct {
	SqlJon   map[string]*SqlJon        `json:"sqlJon"`   //内关联查询Table结构
	TableSql map[string]*MysqlExecutor `json:"tableSql"` //查询内容集
	Data     map[string]interface{}    `json:"data"`     //查询返回数据集
	config   *conf.Config              //读取配置
	Err      error                     `json:"err"`
}

func NewJsonSQL(ctx map[string]interface{}) *JsonSqlExecutor {
	njsql := NewJsonSqlExecutor()
	if njsql.Err == nil {
		njsql.Err = njsql.doSortedMap(ctx)
		// njsql.Err = njsql.createSql(ctx, true)
	}
	return njsql
}

// 创建构造体
func NewJsonSqlExecutor() *JsonSqlExecutor {
	conf := conf.ReadJsonConfig()
	jsonSqlExecutor := &JsonSqlExecutor{
		SqlJon:   make(map[string]*SqlJon),
		TableSql: make(map[string]*MysqlExecutor),
		Data:     make(map[string]interface{}),
	}
	if len(conf.Table) == 0 {
		jsonSqlExecutor.Err = errors.New("Config no init Table list!")
		return jsonSqlExecutor
	}
	jsonSqlExecutor.config = conf
	return jsonSqlExecutor
}

// json map 数据分组
func (e *JsonSqlExecutor) doSortedMap(ctx map[string]interface{}) error {
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
	// 正常table 生成处理
	if len(table) > 0 {
		err = e.createSql(table)
		if err != nil {
			return err
		}
	}

	// 执行查询
	err = e.querySql()
	if err != nil {
		return err
	}

	return nil
}

// 构造 SQL 开始 ctx 构造的对象, exeSql 是否是要执行sql 语句. sql@ 并不需要执行，只生成sql语句。
func (e *JsonSqlExecutor) createSql(ctx map[string]interface{}) error {
	for t, v := range ctx {
		// 构造 table
		err = e.toTable(t, v)
		if err != nil {
			return err
		}
		// 构造 where 查询条件
		err = e.toMakeWhere(t, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Table 表转化
func (e *JsonSqlExecutor) toTable(t string, val interface{}) error {
	tables := strings.Split(t, "|")
	//HasPrefix 判断字符串 s 是否以 prefix 开头：
	//HasSuffix 判断字符串 s 是否以 suffix 结尾：
	// 表数组 table
	m := &MysqlExecutor{}
	for i := 0; i < len(tables); i++ {
		//判断元素是否在数组内
		if _, ok := e.config.Table[tables[i]]; ok {
			m.Table += fmt.Sprintf(" %s AS %s", e.config.Table[tables[i]].TableName, strings.ToLower(tables[i]))
			if i == len(tables)-1 {
				m.Table += ""
			} else {
				m.Table += ","
			}
		} else {
			// @ 引用表处理
			if strings.HasSuffix(tables[i], "@") {
				if _sql_t, ok := e.SqlJon[tables[i]]; ok {
					m.Table += fmt.Sprintf(" (%s) AS %s", _sql_t.Sql, strings.ToLower(tables[i][:len(tables[i])-1]))
					m.Params = append(m.Params, _sql_t.Params...) //多引用 @ 条件 params 值问题解决
				}
			} else {
				return errors.New(fmt.Sprintf("Table [%s] is no exit;", tables[i]))
			}
		}
	}
	e.TableSql[t] = m
	return nil
}

//核心where业务逻辑处理
func (e *JsonSqlExecutor) toMakeWhere(t string, val interface{}) error {
	if vals, ok := val.(map[string]interface{}); ok {
		for k, v := range vals {
			if strings.HasPrefix(k, "@") {
				switch k {
				case "@c":
					e.column(t, v)
				case "@column":
					e.column(t, v)
				case "@w":
					e.where(t, v)
				case "@where":
					e.where(t, v)
				case "@g":
					e.group(t, v)
				case "@group":
					e.group(t, v)
				case "@o":
					e.order(t, v)
				case "@order":
					e.order(t, v)
				case "@l":
					e.limit(t, v)
				case "@limit":
					e.limit(t, v)
				case "@os":
					e.offset(t, v)
				case "@offset":
					e.offset(t, v)
				default:
					return errors.New(fmt.Sprintf("The [%s] is no exit;", k))
				}
			}
		}
		// 生成sql语句
		e.toSql(t)
	}
	return nil
}

// column 字段处理
func (e *JsonSqlExecutor) column(t string, cVal interface{}) error {
	m := e.TableSql[t]

	if columnsStr, ok := cVal.(string); ok {
		columns := strings.Split(columnsStr, ",")
		//数组 元素搜索
		// sort.Strings(str_array)
		// index := sort.SearchStrings(str_array, target)

		if len(e.SqlJon) > 0 {
			for i, column := range columns {
				a := strings.Index(column, "@") //引用逻辑处理
				if a > 0 {
					_sql := column[:a+1] //取出 sql@
					if _sql_t, ok := e.SqlJon[_sql]; ok {
						//column 子查询
						b := strings.HasPrefix(column[a+1:], "/")
						if b {
							m.Columns = append(m.Columns, fmt.Sprintf("(%s) AS %s", _sql_t.Sql, column[a+2:]))
						} else {
							m.Columns = append(m.Columns, fmt.Sprintf("(%s) AS column[%d]", _sql_t.Sql, i))
						}
						m.Params = append(m.Params, _sql_t.Params...) //多引用 @ 条件 params 值问题解决
					} else {
						m.Columns = append(m.Columns, column)
						// return errors.New(fmt.Sprintf("[%s] column [%s] is no exit;", t, _sql))
					}
				} else {
					m.Columns = append(m.Columns, column)
				}
			}
		} else {
			m.Columns = columns
		}
	}
	return nil
}

// 生成 where 条件
func (e *JsonSqlExecutor) where(t string, wVal interface{}) error {
	w := e.TableSql[t]
	andFirst := true

	where := make(map[string]interface{})
	if mw, ok := wVal.(map[string]interface{}); ok {
		for k, v := range mw {
			// 优先处理没有 $ | 的条件,解决map 无序不好控制问题。
			if !strings.HasPrefix(k, "&") && !strings.HasPrefix(k, "|") {
				if andFirst {
					w.Where += fmt.Sprintf(" %s", e.makeWhere(t, k, v))
					andFirst = false
				} else {
					w.Where += fmt.Sprintf(" AND %s", e.makeWhere(t, k, v))
				}
			} else {
				where[k] = v
			}
		}
	}

	// 其它的条件处理
	for k, v := range where {
		if strings.HasPrefix(k, "&") {
			w.Where += fmt.Sprintf(" AND %s", e.makeWhere(t, k[1:], v))
		} else if strings.HasPrefix(k, "|") {
			w.Where += fmt.Sprintf(" OR %s", e.makeWhere(t, k[1:], v))
		}
	}

	return nil
}

// where column 逻辑处理 Key
func (e *JsonSqlExecutor) makeWhere(t string, k string, wcVal interface{}) string {
	m := e.TableSql[t]
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

// where 条件 @引用 value 转换  table.column
func (e *JsonSqlExecutor) makeWhereToTableORColumnt(t string, wcVal interface{}) string {
	m := e.TableSql[t]
	// @引用 value 处理
	if wcV, ok := wcVal.(string); ok {
		b := strings.Index(wcV, "/")
		if b > 0 {
			return fmt.Sprintf("%s.%s", strings.ToLower(wcV[:b]), wcV[b+1:len(wcV)])
		} else {
			m.Params = append(m.Params, wcVal)
			return "?"
		}
	} else {
		m.Params = append(m.Params, parseNum(wcVal))
		return "?"
	}
}

func (e *JsonSqlExecutor) group(t string, gVal interface{}) error {
	m := e.TableSql[t]
	if gV, ok := gVal.(string); ok {
		m.Group = fmt.Sprintf(" GROUP BY %s", gV[:len(gV)-1])
	}
	return nil
}

// order by 条件转换
func (e *JsonSqlExecutor) order(t string, oVal interface{}) error {
	m := e.TableSql[t]
	if oV, ok := oVal.(string); ok {
		if strings.HasSuffix(oV, "-") || strings.HasSuffix(oV, "+") {
			if strings.HasSuffix(oV, "-") {
				m.Order = fmt.Sprintf(" ORDER BY %s DESC", oV[:len(oV)-1])
			} else {
				m.Order = fmt.Sprintf(" ORDER BY %s ASC", oV[:len(oV)-1])
			}
		} else {
			m.Order = fmt.Sprintf(" %s", oV)
		}
	}
	return nil
}

func (e *JsonSqlExecutor) limit(t string, lVal interface{}) error {
	m := e.TableSql[t]
	m.Limit = parseNum(lVal)
	return nil
}

func (e *JsonSqlExecutor) offset(t string, osVal interface{}) error {
	m := e.TableSql[t]
	m.Offset = parseNum(osVal)
	return nil
}

// 查询语句组装
func (e *JsonSqlExecutor) toSql(t string) error {
	m := e.TableSql[t]
	m.Sql = "SELECT "
	if len(m.Columns) > 0 {
		m.Sql += strings.Join(m.Columns, ",")
	} else {
		m.Sql += "*"
	}
	m.Sql += " FROM"
	if m.Table != "" {
		m.Sql += m.Table
	}
	if m.Where != "" {
		m.Sql += " WHERE"
		m.Sql += m.Where
	}
	if m.Group != "" {
		m.Sql += m.Group
	}
	if m.Order != "" {
		m.Sql += m.Order
	}
	if m.Limit > 0 {
		m.Sql += fmt.Sprintf(" LIMIT %d", m.Limit)
		if m.Offset > 0 {
			m.Sql += fmt.Sprintf(" OFFSET  %d", m.Offset)
		}
	}
	return nil
}

// 执行查询
func (e *JsonSqlExecutor) querySql() error {
	for table, _ := range e.TableSql {
		m := e.TableSql[table]
		if e.config.Debug {
			logger.Info("*************************************************************************")
			logger.Debugf("params: %v", m.Params)
			logger.Debugf("%s", m.Sql)
			logger.Info("*************************************************************************")
		}
		data, err := db.Query(m.Sql, m.Params...)
		if err != nil {
			return errors.New(fmt.Sprintf("query err: %s", err.Error()))
		}
		e.Data[table] = data
	}
	return nil
}
