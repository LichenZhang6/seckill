package common

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// 创建mysql连接
func NewMysqlConn() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", "root:lichenzhang8220@tcp(127.0.0.1:3306)/test?charset-utf8")
	return
}

// 获取一条数据
func GetResultRow(rows *sql.Rows) map[string]string {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	record := make(map[string]string)
	for rows.Next() {
		rows.Scan(scanArgs...)
		for i, v := range values {
			if v != nil {
				record[columns[i]] = string(v.([]byte))
			}
		}
	}
	return record
}

// 获取所有数据
func GetResultRows(rows *sql.Rows) map[int]map[string]string {

	// 返回所有列
	columns, _ := rows.Columns()

	// 表示一行所有列的值，用[][]byte表示
	vals := make([][]byte, len(columns))

	// 表示一行填充数据
	scans := make([]interface{}, len(columns))

	// 这里scans引用vals，把数据填充到[]byte里
	for k, _ := range vals {
		scans[k] = &vals[k]
	}
	i := 0
	result := make(map[int]map[string]string)
	for rows.Next() {

		// 填充数据
		rows.Scan(scans...)

		// 每行数据
		row := make(map[string]string)

		// 把val中的数据复制到rows中
		for k, v := range vals {
			key := columns[k]

			// 把[]byte数据转成string
			row[key] = string(v)
		}

		// 放入结果集
		result[i] = row
		i++
	}
	return result
}
