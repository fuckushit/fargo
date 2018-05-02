package util

import (
	"bdlib/db"
	"strconv"
	"strings"

	"bdlib/config"
	"bdlib/crypt"
)

// DBStore DB 连接池 struct.
type DBStore struct {
	DBPool chan *db.DB
}

var gDbSection config.Section

// NewDBStore 初始化数据库连接池
// 垮库操作 section 配置文件中多库配置
func NewDBStore(cfg config.Configer, section string, dbCryptKey string) (store *DBStore, er error) {
	store = new(DBStore)

	// 获取 db 配置信息
	dbSection, err := cfg.GetSection(section)
	if err != nil {
		er = err
		return
	}
	gDbSection = dbSection
	if _, ok := dbSection["pass"]; ok {
		if dbSection["pass"] != "" {
			dbSection["pass"], err = crypt.Decrypt(dbSection["pass"], dbCryptKey)
			if err != nil {
				er = err
				return
			}
		}

	}
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

	store.DBPool = make(chan *db.DB, numDB)
	for i := 0; i < numDB; i++ {
		db, err := db.NewConnection(dbSection)
		if err != nil {
			er = err
			return
		}

		store.DBPool <- db
	}
	return
}

// NewDBStoreByDBName 初始化数据库连接池
// 垮库操作 section 配置文件中多库配置
func NewDBStoreByDBName(cfg config.Configer, section string, dbName, dbCryptKey string) (store *DBStore, er error) {
	store = new(DBStore)

	// 获取 db 配置信息
	dbSection, err := cfg.GetSection(section)
	if err != nil {
		er = err
		return
	}
	gDbSection = dbSection
	if _, ok := dbSection["pass"]; ok {
		dbSection["pass"], err = crypt.Decrypt(dbSection["pass"], dbCryptKey)
		if err != nil {
			er = err
			return
		}
	}

	dbSection["dbname"] = dbName

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

	store.DBPool = make(chan *db.DB, numDB)
	for i := 0; i < numDB; i++ {
		db, err := db.NewConnection(dbSection)
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
func (d *DBStore) GetConn() (db *db.DB) {
	if len(d.DBPool) == 0 {
		return nil
	}

	db = <-d.DBPool

	return
}

// ReturnConn 归还 db 连接
func (d *DBStore) ReturnConn(db *db.DB) {
	d.DBPool <- db
	return
}

// Addslashes mysql 防注入 转义.
func Addslashes(in string) (out string) {
	r := strings.NewReplacer("'", "\\'", "\"", "\\\"", "\\", "\\\\", "NULL", "\\NULL")
	out = r.Replace(in)

	return
}
