package db

import (
	"fmt"

	"github.com/qqliaoxin/jsonsql/conf"
	"github.com/qqliaoxin/jsonsql/logger"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var (
	db  *sqlx.DB
	err error
)

// 初始化数据库连接
func init() {
	config := conf.ReadDBConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.UserName, config.PassWord, config.Host, config.Port, config.DataBase)
	db, err = sqlx.Open("mysql", dsn)
	if err != nil {
		logger.Debugf("mysql connect server failed, err:%v\n", err)
		return
	}
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(10)
}

// 执行查询
func Query(sql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	data := make([]map[string]interface{}, 0, 8)
	values := make([]interface{}, count)
	valuePointers := make([]interface{}, count)
	for rows.Next() {
		for i := 0; i < count; i++ {
			valuePointers[i] = &values[i]
		}
		rows.Scan(valuePointers...)
		entry := make(map[string]interface{})
		for i, column := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[column] = v
		}
		data = append(data, entry)
	}
	return data, nil
}

func Inster(sql string, args ...interface{}) (int64, error) {
	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("begin trans failed, err:%v\n", err)
		return 0, err
	}

	defer func() {
		// 捕获panic
		if p := recover(); p != nil {
			// 回滚
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			fmt.Println("rollback")
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			fmt.Println("commit")
		}
	}()

	result, err := tx.Exec(sql, args...)
	if err != nil {
		fmt.Printf("exec failed, err:%v\n", err)
		return 0, err
	}
	insertID, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("get insert id failed, err:%v\n", err)
		return 0, err
	}
	fmt.Printf("insert data success, id:%d\n", insertID)
	return insertID, nil
}

func Update(sql string, args ...interface{}) (int64, error) {
	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("begin trans failed, err:%v\n", err)
		return 0, err
	}

	defer func() {
		// 捕获panic
		if p := recover(); p != nil {
			// 回滚
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			fmt.Println("rollback")
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			fmt.Println("commit")
		}
	}()

	result, err := tx.Exec(sql, args...)
	if err != nil {
		fmt.Printf("exec failed, err:%v\n", err)
		return 0, err
	}
	// 操作影响的行数
	updateID, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("get update failed, err:%v\n", err)
		return 0, err
	}
	fmt.Printf("update data success, id:%d\n", updateID)
	return updateID, nil
}

func Delete(sql string, args ...interface{}) (int64, error) {
	// 开启事务
	tx, err := db.Beginx()
	if err != nil {
		fmt.Printf("begin trans failed, err:%v\n", err)
		return 0, err
	}

	defer func() {
		// 捕获panic
		if p := recover(); p != nil {
			// 回滚
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			fmt.Println("rollback")
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			fmt.Println("commit")
		}
	}()

	result, err := tx.Exec(sql, args...)
	if err != nil {
		fmt.Printf("exec failed, err:%v\n", err)
		return 0, err
	}
	// 操作影响的行数
	updateID, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("get delete failed, err:%v\n", err)
		return 0, err
	}
	fmt.Printf("delete success, id:%d\n", updateID)
	return updateID, nil
}
