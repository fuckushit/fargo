package util

import (
	"bdlib/mysql"
	"strconv"

	"bdlib/config"
)

// DBManager ...
type DBManager struct {
	Pool *DBStore // dbchan TODO 业务层是否需要读写分离...
}

// DBStore DB 连接池 struct.
type DBStore struct {
	DBPool chan *mysql.DB
}

var gDbSection config.Section

// NewDBStore 初始化数据库连接池
// 垮库操作 section 配置文件中多库配置
func NewDBStore(dbSection config.Section) (store *DBStore, er error) {
	store = new(DBStore)
	numDBStr, err := dbSection.GetValue("poolsize")
	if err != nil {
		er = err
		return
	}
	numDB, err := strconv.Atoi(numDBStr)
	if err != nil {
		er = err
		return
	}

	store.DBPool = make(chan *mysql.DB, numDB)
	for i := 0; i < numDB; i++ {
		db, err := mysql.NewConnection(dbSection)
		if err != nil {
			er = err
			return
		}
		store.DBPool <- db
	}

	return
}

// GetConn 获取 db 连接
// 获取 db 连接 然后 ping 一下 如果 ping 不通的话重连
func (d *DBStore) GetConn() *mysql.DB {
	if len(d.DBPool) == 0 {
		return nil
	}
	return <-d.DBPool
}

// ReturnConn 归还 db 连接
func (d *DBStore) ReturnConn(db *mysql.DB) {
	d.DBPool <- db
	return
}

// GetDB _
func (m *DBManager) GetDB() (db *mysql.DB) {
	db = m.Pool.GetConn()
	return
}

// PutDB _
func (m *DBManager) PutDB(db *mysql.DB) {
	m.Pool.ReturnConn(db)
	return
}
