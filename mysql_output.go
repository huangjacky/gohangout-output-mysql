package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

// MySQLOutput 使用的Kafka-go的output插件
type MySQLOutput struct {
	config       map[interface{}]interface{}
	DefaultTable string
	TableKey     string
	User         string
	Password     string
	Host         string
	Port         int
	Database     string

	Conn        *sql.DB
	MaxConns    int
	MaxIdles    int
	MaxLifeTime int
}

/*
New 插件模式的初始化
*/
func New(config map[interface{}]interface{}) interface{} {
	var err error
	p := &MySQLOutput{
		config: config,
	}
	if v, ok := config["TableKey"]; ok {
		p.TableKey = v.(string)
	} else {
		p.TableKey = "table"
	}
	if v, ok := config["DefaultTable"]; ok {
		p.DefaultTable = v.(string)
	} else {
		p.DefaultTable = "test"
	}
	if v, ok := config["User"]; ok {
		p.User = v.(string)
	} else {
		glog.Fatal("User 配置项 必须设置")
	}
	if v, ok := config["Password"]; ok {
		p.Password = v.(string)
	} else {
		glog.Fatal("Password 配置项 必须设置")
	}
	if v, ok := config["Host"]; ok {
		p.Host = v.(string)
	} else {
		glog.Fatal("Host 配置项 必须设置")
	}
	if v, ok := config["Database"]; ok {
		p.Database = v.(string)
	} else {
		glog.Fatal("Database 配置项 必须设置")
	}
	if v, ok := config["Port"]; ok {
		p.Port = v.(int)
	} else {
		p.Port = 3306
	}
	if v, ok := config["MaxConns"]; ok {
		p.MaxConns = v.(int)
	} else {
		p.MaxConns = 256
	}
	if v, ok := config["MaxIdles"]; ok {
		p.MaxIdles = v.(int)
	} else {
		p.MaxIdles = 64
	}
	if v, ok := config["MaxLifeTime"]; ok {
		p.MaxLifeTime = v.(int)
	} else {
		p.MaxLifeTime = 60
	}
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", p.User, p.Password, p.Host, p.Port, p.Database)
	glog.Info("ConnectionString: ", connectionString)
	p.Conn, err = sql.Open("mysql", connectionString)
	if err != nil {
		glog.Fatal("sql.Open: ", err)
	}
	p.Conn.SetMaxOpenConns(p.MaxConns)
	p.Conn.SetMaxIdleConns(p.MaxIdles)
	p.Conn.SetConnMaxLifetime(time.Minute * time.Duration(p.MaxLifeTime))
	p.Conn.Ping()
	return p
}

//Emit 单次事件的处理函数
func (p *MySQLOutput) Emit(event map[string]interface{}) {
	var tableName string
	var keyLen int
	// 减去TableKey， (再减去 @timestamp 这个key, 如果还有@meta 这些呢, 先通过遍历，然后切片的方式来解决）
	if v, ok := event[p.TableKey]; ok {
		tableName = v.(string)
		keyLen = len(event) - 1
	} else {
		tableName = p.DefaultTable
		keyLen = len(event) - 0
	}
	if keyLen == 0 {
		return
	}
	placeholders := make([]string, keyLen)
	fields := make([]string, keyLen)
	values := make([]interface{}, keyLen)
	i := 0
	for k, v := range event {
		if k == p.TableKey || strings.HasPrefix(k, "@") {
			continue
		}
		placeholders[i] = "?"
		fields[i] = fmt.Sprintf("`%s`", k)
		values[i] = v
		i++
	}
	for i = 0; i < keyLen; i++ {
		if placeholders[i] != "?" {
			break
		}
	}
	if i < keyLen {
		placeholders = placeholders[0:i]
		fields = fields[0:i]
		values = values[0:i]
	}
	sql := fmt.Sprintf("INSERT INTO %s(%s)VALUES(%s)", tableName, strings.Join(fields, ","), strings.Join(placeholders, ","))
	stmt, err := p.Conn.Prepare(sql)
	if err != nil {
		glog.Error("SQL: ", sql)
		glog.Error("p.Conn.Prepare: ", err)
		return
	}
	_, err = stmt.Exec(values...)
	if err != nil {
		glog.Error("stmt.Exec: ", err)
		return
	}
}

//Shutdown 关闭需要做的事情
func (p *MySQLOutput) Shutdown() {
	if err := p.Conn.Close(); err != nil {
		glog.Fatal("p.Conn.Close:", err)
	}
}
