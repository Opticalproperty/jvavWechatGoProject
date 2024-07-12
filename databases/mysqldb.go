package databases

import (
	"database/sql"
	"log"
	"time"
)

// 配置数据库连接信息
const (
	dsn             = "root:Wall8023.@tcp(117.72.37.95:3306)/wechat"
	maxOpenConns    = 10
	maxIdleConns    = 5
	connMaxLifetime = 30 * time.Minute
)

func getMysqlConnect() (db *sql.DB) {

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 配置数据库连接池
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	return db
}
