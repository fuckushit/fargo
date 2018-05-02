package mysql

import (
	"bdlib/config"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	_ "github.com/go-sql-driver/mysql.v1" // to ues package mysql init function.
)

// DB db handler.
type DB struct {

	// Host address -> ip:port
	host string

	// User
	user string

	// Password
	pass string

	// Database name
	dbname string

	// Charset
	charset string

	// Debug model
	debug bool

	// .db file path for sqlite
	dsn string

	db *sql.DB

	// Database transaction
	tx *sql.Tx

	result sql.Result
	rows   *sql.Rows
}

// Config The config of DB.
type Config map[string]string

// Row redefine the slice result.
type Row []string

// RowMap redefine the map result.
type RowMap map[string]string

// sqlExecuter the sql interface for the handle function.
type sqlExecuter interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// sqlExecuters executers that explements the 'sqlExecuter' interface container.
var sqlExecuters = make(map[string]sqlExecuter)

const (
	// DEFAULTCHARSET db connect default charset to use.
	DEFAULTCHARSET = "UTF8"
	// DBDRIVER db connect driver to use there.
	DBDRIVER = "mysql"
)

// Register register an db executer that elements the 'sqlExecuter' interface,
// panic when there is an executer that name already.
func Register(name string, sqlExecuter sqlExecuter) {
	if sqlExecuter == nil {
		panic("message: sqlExecuter is nil")
	}
	if _, ok := sqlExecuters[name]; ok {
		panic("message: sqlExecuter called twice for sqlExecuter " + name)
	}
	sqlExecuters[name] = sqlExecuter
}

// NewConnectionPool Create database connection pool.
func NewConnectionPool(dbSec config.Section, poolSize int) (dbPool chan *DB, err error) {
	var db *DB
	dbPool = make(chan *DB, poolSize)
	for i := 0; i < poolSize; i++ {
		db, err = NewConnection(dbSec)
		if err != nil {
			return
		}
		dbPool <- db
	}
	return
}

// Connect connect to db using config.
// Config MUST include "host", "user", "pass" and "dbname" parameters.
// "charset" is optional, which is UTF-8 by default.
// Param "host" should be in the form <ip_address>:<port>
func Connect(c Config) (newdb *DB, err error) {
	newdb = new(DB)
	for _, param := range []string{"host", "user", "pass", "dbname"} {
		if _, ok := c[param]; !ok {
			return nil, fmt.Errorf("Database config param %s not found", param)
		}
	}
	newdb.host = c["host"]
	newdb.user = c["user"]
	newdb.pass = c["pass"]
	newdb.dbname = c["dbname"]

	if _, ok := c["charset"]; ok {
		newdb.charset = c["charset"]
	} else {
		newdb.charset = DEFAULTCHARSET
	}
	if _, ok := c["debug"]; ok {
		newdb.debug = c["debug"] == "true"
	} else {
		newdb.debug = false
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&loc=Local", newdb.user, newdb.pass, newdb.host, newdb.dbname, newdb.charset)
	newdb.db, err = sql.Open(DBDRIVER, dsn)
	if err != nil {
		return nil, err
	}
	if err = newdb.db.Ping(); err != nil {
		return nil, err
	}

	return newdb, nil
}

// NewConnection connect to db using config.Section
// For convenience after reading a config file
//
func NewConnection(cfg config.Section) (*DB, error) {
	dbConfig := make(Config)
	for k, v := range cfg {
		dbConfig[k] = v
	}
	return Connect(dbConfig)
}

// Execute execute a sql query. sqlPattern should use placeholder "?".
//
// i.e insert into person(name, age) values(?,?).
// Because this function uses Stmt.
func (x *DB) Execute(sqlPattern string, arguments ...interface{}) (err error) {
	var stmt *sql.Stmt
	var executer sqlExecuter
	if x.tx != nil {
		// In transaction
		executer = x.tx
	} else {
		executer = x.db
	}
	if x.rows != nil {
		x.rows.Close()
		x.rows = nil
	}

	stmt, err = executer.Prepare(sqlPattern)
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	if err != nil {
		return
	}

	x.result, err = stmt.Exec(arguments...)
	if err != nil {
		return
	}
	return
}

// Query query and store results. sqlPattern should be the same format as Execute.
// After Query, use FetchRow / FetchRowMap / FetchAll to fetch the result.
func (x *DB) Query(sqlPattern string, arguments ...interface{}) (err error) {
	var executer sqlExecuter
	if x.tx != nil {
		// In transaction
		executer = x.tx
	} else {
		executer = x.db
	}
	if x.rows != nil {
		x.rows.Close()
		x.rows = nil
	}
	x.rows, err = executer.Query(sqlPattern, arguments...)
	if err != nil {
		return
	}
	return
}

// Next prepares the next result row for reading with the Scan method. It returns true on success, or false if there is no next result row or an error happened while preparing it. Err should be consulted to distinguish between the two cases.
// Every call to Scan, even the first one, must be preceded by a call to Next.
func (x *DB) next() bool {
	return x.rows.Next()
}

// GetRowsField for fetch columns
func (x *DB) GetRowsField() ([]string, error) {
	return x.rows.Columns()
}

// FetchRow fetch a row of query result.
func (x *DB) FetchRow() (rows Row, err error) {
	if x.rows == nil {
		return nil, io.EOF
	}
	if x.rows.Next() {
		var cols []string
		cols, err = x.rows.Columns()
		if err != nil {
			return nil, err
		}
		scanArgs := make([]interface{}, len(cols))
		values := make([]sql.RawBytes, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = x.rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for _, row := range values {
			rows = append(rows, string(row))
		}
		return
	}
	x.rows.Close()
	return nil, io.EOF
}

// FetchOneRow fetch one row from result, and release result
func (x *DB) FetchOneRow() (row Row, err error) {
	row, err = x.FetchRow()
	if x.rows != nil {
		x.rows.Close()
		x.rows = nil
	}
	return
}

// FetchOne fetch first column from first row result.
func (x *DB) FetchOne() (one string, err error) {
	row, err := x.FetchOneRow()
	if err == nil && len(row) > 0 {
		one = row[0]
	}
	return
}

// FetchRowMap fetch a row of query result, return map form.
func (x *DB) FetchRowMap() (retMap RowMap, err error) {
	retMap = make(map[string]string)
	if x.rows == nil {
		return nil, io.EOF
	}
	if x.rows.Next() {
		var cols []string
		cols, err = x.rows.Columns()
		if err != nil {
			return nil, err
		}
		scanArgs := make([]interface{}, len(cols))
		values := make([]sql.RawBytes, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = x.rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(cols); i++ {
			retMap[cols[i]] = string(values[i])
		}
		return
	}
	x.rows = nil
	return nil, io.EOF
}

// FetchRowMapInterface returns the map[string]interface{} for the Query
func (x *DB) FetchRowMapInterface() (rowMap map[string]interface{}, err error) {
	rowMap = make(map[string]interface{})
	if x.rows == nil {
		return nil, io.EOF
	}
	if x.rows.Next() {
		var cols []string
		cols, err = x.rows.Columns()
		if err != nil {
			return nil, err
		}
		scanArgs := make([]interface{}, len(cols))
		values := make([]sql.RawBytes, len(cols))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = x.rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(cols); i++ {
			rowMap[cols[i]] = values[i]
		}
		return
	}
	x.rows = nil
	return nil, io.EOF
}

// AffectRows return the number of affected rows after Execute,
// Return -1 represent an error occurs.
func (x *DB) AffectRows() (affect int64) {
	var err error
	if x.result == nil {
		return -1
	}
	affect, err = x.result.RowsAffected()
	if err != nil {
		return -1
	}
	return affect
}

// LastInsertID return the number of last inserted id after inserting a record.
// Return -1 represent an error occurs.2
func (x *DB) LastInsertID() (id int64) {
	var err error
	if x.result == nil {
		return -1
	}
	id, err = x.result.LastInsertId()
	if err != nil {
		return -1
	}
	return id
}

// FetchAll fetch all of the results after Query.
func (x *DB) FetchAll() (allrows []Row, err error) {
	var row []string
	for {
		row, err = x.FetchRow()
		if err == io.EOF {
			err = nil
			return
		} else if err != nil {
			return
		}
		allrows = append(allrows, row)
	}
	return
}

// FetchAllMap fetch all of the results with map format after Query.
func (x *DB) FetchAllMap() (allrows []RowMap, err error) {
	var row RowMap
	for {
		row, err = x.FetchRowMap()
		if err == io.EOF {
			err = nil
			return
		} else if err != nil {
			return
		}
		allrows = append(allrows, row)
	}
	return
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
func (x *DB) Ping() (err error) {
	err = x.db.Ping()
	return
}

// Insert insert data(as a map), return last inserted id.
func (x *DB) Insert(tableName string, data map[string]interface{}) (lastID int64, err error) {
	// INSERT INTO tableName(keyStr) VALUES(?,?,...)
	if len(data) == 0 {
		return -1, errors.New("Insert data is empty!")
	}
	dataKeys := []string{}
	dataValues := []string{}
	var values = make([]interface{}, 0)
	for key, val := range data {
		dataKeys = append(dataKeys, key)
		dataValues = append(dataValues, "?")
		values = append(values, val)
	}
	tableName = x.wrapTableName(tableName)
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"
	valStr := strings.Join(dataValues, ",")
	sqlPattern := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tableName, keyStr, valStr)
	err = x.Execute(sqlPattern, values...)
	if err == nil {
		lastID = x.LastInsertID()
	}
	return
}

// Replace execute the replace sql, to replace a field simply with that function,
// Replace INTO tableName(keyStr) VALUES(?,?,...).
func (x *DB) Replace(tableName string, data map[string]interface{}) (affectRows int64, err error) {
	affectRows = 0
	// Replace INTO tableName(keyStr) VALUES(?,?,...)
	if len(data) == 0 {
		return -1, errors.New("Replace data is empty!")
	}
	dataKeys := []string{}
	dataValues := []string{}
	var values = make([]interface{}, 0)
	for key, val := range data {
		dataKeys = append(dataKeys, key)
		dataValues = append(dataValues, "?")
		values = append(values, val)
	}
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"
	valStr := strings.Join(dataValues, ",")
	sqlPattern := fmt.Sprintf("REPLACE INTO `%s`(%s) VALUES(%s)", tableName, keyStr, valStr)
	err = x.Execute(sqlPattern, values...)
	if err == nil {
		affectRows = x.AffectRows()
	}
	return
}

// MultiReplace insert multiple data at once. Return all of the insert ids according to the database.
func (x *DB) MultiReplace(tableName string, data []map[string]interface{}) (lastID int64, err error) {
	if len(data) == 0 {
		err = errors.New("MultiReplace data is empty!")
		return
	}
	dataKeys := []string{}
	placeHolder := []string{}
	for key := range data[0] {
		dataKeys = append(dataKeys, key)
		placeHolder = append(placeHolder, "?")
	}
	placeHolderStr := strings.Join(placeHolder, ",")
	placeHolderStr = fmt.Sprintf("(%s)", placeHolderStr)
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"

	// Check if each of data has equal number of keys
	for _, eachData := range data {
		if len(eachData) != len(dataKeys) {
			err = errors.New("MultiReplace: All insert data should have equal number of fields.")
			return
		}
	}

	tableName = x.wrapTableName(tableName)
	sqlPattern := bytes.NewBufferString(fmt.Sprintf("Replace INTO %s(%s) VALUES ", tableName, keyStr))
	for idx := range data {
		sqlPattern.WriteString(placeHolderStr)
		if idx < len(data)-1 {
			sqlPattern.WriteString(",")
		}
	}
	// use stmt
	var stmt *sql.Stmt
	var executer sqlExecuter
	if x.tx != nil {
		executer = x.tx
	} else {
		executer = x.db
	}
	stmt, err = executer.Prepare(sqlPattern.String())
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	if err != nil {
		return
	}
	values := make([]interface{}, len(dataKeys)*len(data))
	var idx = 0
	for _, eachData := range data {
		for _, key := range dataKeys {
			values[idx] = eachData[key]
			idx++
		}
	}
	x.result, err = stmt.Exec(values...)
	if err == nil {
		lastID = x.LastInsertID()
	}
	return
}

// MultiInsert insert multiple data at once. Return all of the insert ids according to the database.
func (x *DB) MultiInsert(tableName string, data []map[string]interface{}) (lastID int64, err error) {
	if len(data) == 0 {
		err = errors.New("MultiInsert data is empty!")
		return
	}
	dataKeys := []string{}
	placeHolder := []string{}
	for key := range data[0] {
		dataKeys = append(dataKeys, key)
		placeHolder = append(placeHolder, "?")
	}
	placeHolderStr := strings.Join(placeHolder, ",")
	placeHolderStr = fmt.Sprintf("(%s)", placeHolderStr)
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"

	// Check if each of data has equal number of keys
	for _, eachData := range data {
		if len(eachData) != len(dataKeys) {
			err = errors.New("MultiInsert: All insert data should have equal number of fields.")
			return
		}
	}

	tableName = x.wrapTableName(tableName)
	sqlPattern := bytes.NewBufferString(fmt.Sprintf("INSERT INTO %s(%s) VALUES ", tableName, keyStr))
	for idx := range data {
		sqlPattern.WriteString(placeHolderStr)
		if idx < len(data)-1 {
			sqlPattern.WriteString(",")
		}
	}
	// use stmt
	var stmt *sql.Stmt
	var executer sqlExecuter
	if x.tx != nil {
		executer = x.tx
	} else {
		executer = x.db
	}
	stmt, err = executer.Prepare(sqlPattern.String())
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	if err != nil {
		return
	}
	values := make([]interface{}, len(dataKeys)*len(data))
	var idx = 0
	for _, eachData := range data {
		for _, key := range dataKeys {
			values[idx] = eachData[key]
			idx++
		}
	}
	x.result, err = stmt.Exec(values...)
	if err == nil {
		lastID = x.LastInsertID()
	}
	return
}

// MultiInsertIgnore insert multiple data at once and ignore unqiue key. Return all of the insert ids according to the database.
func (x *DB) MultiInsertIgnore(tableName string, data []map[string]interface{}) (lastID int64, err error) {
	if len(data) == 0 {
		err = errors.New("MultiInsert data is empty!")
		return
	}
	dataKeys := []string{}
	placeHolder := []string{}
	for key := range data[0] {
		dataKeys = append(dataKeys, key)
		placeHolder = append(placeHolder, "?")
	}
	placeHolderStr := strings.Join(placeHolder, ",")
	placeHolderStr = fmt.Sprintf("(%s)", placeHolderStr)
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"

	// Check if each of data has equal number of keys
	for _, eachData := range data {
		if len(eachData) != len(dataKeys) {
			err = errors.New("MultiInsert: All insert data should have equal number of fields.")
			return
		}
	}

	tableName = x.wrapTableName(tableName)
	sqlPattern := bytes.NewBufferString(fmt.Sprintf("INSERT IGNORE INTO %s(%s) VALUES ", tableName, keyStr))
	for idx := range data {
		sqlPattern.WriteString(placeHolderStr)
		if idx < len(data)-1 {
			sqlPattern.WriteString(",")
		}
	}
	// use stmt
	var stmt *sql.Stmt
	var executer sqlExecuter
	if x.tx != nil {
		executer = x.tx
	} else {
		executer = x.db
	}
	stmt, err = executer.Prepare(sqlPattern.String())
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	if err != nil {
		return
	}
	values := make([]interface{}, len(dataKeys)*len(data))
	var idx = 0
	for _, eachData := range data {
		for _, key := range dataKeys {
			values[idx] = eachData[key]
			idx++
		}
	}
	x.result, err = stmt.Exec(values...)
	if err == nil {
		lastID = x.LastInsertID()
	}
	return
}

// MultiInsertUpdate insert multiple data at once. Return all of the insert ids according to the database.
func (x *DB) MultiInsertUpdate(tableName string, data []map[string]interface{}, updateKeys []string) (lastID int64, err error) {
	if len(data) == 0 {
		err = errors.New("MultiInsert data is empty!")
		return
	}
	dataKeys := []string{}
	placeHolder := []string{}
	for key := range data[0] {
		dataKeys = append(dataKeys, key)
		placeHolder = append(placeHolder, "?")
	}
	placeHolderStr := strings.Join(placeHolder, ",")
	placeHolderStr = fmt.Sprintf("(%s)", placeHolderStr)
	keyStr := "`" + strings.Join(dataKeys, "`,`") + "`"

	// Check if each of data has equal number of keys
	for _, eachData := range data {
		if len(eachData) != len(dataKeys) {
			err = errors.New("MultiInsert: All insert data should have equal number of fields.")
			return
		}
	}

	tableName = x.wrapTableName(tableName)
	sqlPattern := bytes.NewBufferString(fmt.Sprintf("INSERT INTO %s(%s) VALUES ", tableName, keyStr))
	for idx := range data {
		sqlPattern.WriteString(placeHolderStr)
		if idx < len(data)-1 {
			sqlPattern.WriteString(",")
		}
	}

	// on duplicate key update k = values(k)
	if len(updateKeys) > 0 {
		fmt.Fprintf(sqlPattern, " ON DUPLICATE KEY UPDATE ")
		for idx, k := range updateKeys {
			fmt.Fprintf(sqlPattern, "`%s` = VALUES(`%s`)", k, k)
			if idx < len(updateKeys)-1 {
				fmt.Fprintf(sqlPattern, ",")
			}
		}
	}

	// use stmt
	var stmt *sql.Stmt
	var executer sqlExecuter
	if x.tx != nil {
		executer = x.tx
	} else {
		executer = x.db
	}
	stmt, err = executer.Prepare(sqlPattern.String())
	defer func() {
		if stmt != nil {
			stmt.Close()
		}
	}()

	if err != nil {
		return
	}
	values := make([]interface{}, len(dataKeys)*len(data))
	var idx = 0
	for _, eachData := range data {
		for _, key := range dataKeys {
			values[idx] = eachData[key]
			idx++
		}
	}
	x.result, err = stmt.Exec(values...)
	if err == nil {
		lastID = x.LastInsertID()
	}
	return
}

// Update update data according to condititon.
// Example: Update("userinfo", userInfo, "WHERE uid=?", 3)
func (x *DB) Update(tableName string, data map[string]interface{}, condPattern string, condArgs ...interface{}) (affect int64, err error) {
	if len(data) == 0 {
		return 0, nil
	}
	updatePairs := []string{}
	var values = make([]interface{}, 0)
	for key, val := range data {
		values = append(values, val)
		updatePairs = append(updatePairs, "`"+key+"`=?")
	}
	tableName = x.wrapTableName(tableName)
	sqlPattern := fmt.Sprintf("UPDATE %s SET %s ", tableName, strings.Join(updatePairs, ","))
	sqlPattern += condPattern

	for _, arg := range condArgs {
		values = append(values, arg)
	}
	err = x.Execute(sqlPattern, values...)
	if err == nil {
		affect = x.AffectRows()
	}
	return
}

// Begin Start a transaction.
func (x *DB) Begin() (err error) {
	x.tx, err = x.db.Begin()
	if err != nil {
		return
	}
	if x.rows != nil {
		x.rows.Close()
		x.rows = nil
	}
	return
}

// RollBack when there an error occour in a transaction.
func (x *DB) RollBack() (err error) {
	if x.tx == nil {
		err = errors.New("RollBack error: Not in transaction. Did you use Begin()?")
		return
	}
	if x.rows != nil {
		x.rows.Close()
		x.rows = nil
	}
	err = x.tx.Rollback()
	x.tx = nil
	return
}

// Commit after succussfull transaction then commit.
func (x *DB) Commit() (err error) {
	if x.tx == nil {
		err = errors.New("Commit error: Not in transaction. Did you use Begin()?")
		return
	}
	if x.rows != nil {
		x.rows.Close()
		x.rows = nil
	}
	err = x.tx.Commit()
	x.tx = nil
	return
}

// Close after the execute or query to close the db connection.
func (x *DB) Close() {
	if x.rows != nil {
		x.rows.Close()
	}
	x.rows = nil
	x.tx = nil
	x.result = nil
	x.db.Close()
}

// wrapTableName ...
func (x *DB) wrapTableName(tableName string) (nTableName string) {
	tNames := strings.Split(tableName, ".")
	for idx, t := range tNames {
		t = strings.Trim(t, "`")
		tNames[idx] = "`" + t + "`"
	}
	return strings.Join(tNames, ".")
}
