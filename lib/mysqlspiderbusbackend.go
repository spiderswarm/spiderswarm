package spsw

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLSpiderBusBackend struct {
	SpiderBusBackend

	dbConn *sql.DB
	dsn    string
}

func NewMySQLSpiderBusBackend(dsn string) *MySQLSpiderBusBackend {
	dbConn, err := sql.Open("mysql", dsn+"?autocommit=true")
	if err != nil {
		panic(err)
	}

	dbConn.SetConnMaxLifetime(time.Minute * 3)
	dbConn.SetMaxOpenConns(10)
	dbConn.SetMaxIdleConns(10)

	dbConn.Exec("CREATE TABLE IF NOT EXISTS scheduledTasks (id INT PRIMARY KEY AUTO_INCREMENT, raw LONGTEXT)")
	dbConn.Exec("CREATE TABLE IF NOT EXISTS taskPromises (id INT PRIMARY KEY AUTO_INCREMENT, raw LONGTEXT)")
	dbConn.Exec("CREATE TABLE IF NOT EXISTS items (id INT PRIMARY KEY AUTO_INCREMENT, raw LONGTEXT)")

	return &MySQLSpiderBusBackend{
		dbConn: dbConn,
		dsn:    dsn,
	}
}

func encodeEntry(entry interface{}) []byte {
	buffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buffer)

	encoder.Encode(entry)

	bytes, _ := ioutil.ReadAll(buffer)

	return bytes
}

func decodeEntry(raw []byte, entry interface{}) interface{} {
	buffer := bytes.NewBuffer(raw)
	decoder := json.NewDecoder(buffer)

	err := decoder.Decode(entry)
	if err != nil {
		spew.Dump(err)
	}

	return &entry
}

func (msbb *MySQLSpiderBusBackend) maybePrintError(err error) {
	if err != nil && err != sql.ErrNoRows {
		spew.Dump(err)
	}
}

func (msbb *MySQLSpiderBusBackend) SendScheduledTask(scheduledTask *ScheduledTask) error {
	raw := encodeEntry(scheduledTask)

	_, err := msbb.dbConn.Exec("INSERT INTO scheduledTasks (raw) VALUES (?)", raw)
	if err != nil {
		spew.Dump(err)
	}

	return nil
}

func (msbb *MySQLSpiderBusBackend) ReceiveScheduledTask() *ScheduledTask {
	var row_id int
	var raw []byte

	row := msbb.dbConn.QueryRow("SELECT * FROM scheduledTasks ORDER BY id ASC LIMIT 1")

	err := row.Scan(&row_id, &raw)
	if err != nil {
		return nil
	}

	scheduledTask := &ScheduledTask{}

	decodeEntry(raw, scheduledTask)

	msbb.dbConn.Exec(fmt.Sprintf("DELETE FROM scheduledTasks WHERE id=%d", row_id))

	return scheduledTask
}

func (msbb *MySQLSpiderBusBackend) SendTaskPromise(taskPromise *TaskPromise) error {
	raw := encodeEntry(taskPromise)

	_, err := msbb.dbConn.Exec("INSERT INTO taskPromises (raw) VALUES (?)", raw)
	if err != nil {
		spew.Dump(err)
	}

	return nil
}

func (msbb *MySQLSpiderBusBackend) ReceiveTaskPromise() *TaskPromise {
	var row_id int
	var raw []byte

	row := msbb.dbConn.QueryRow("SELECT * FROM taskPromises ORDER BY id ASC LIMIT 1")

	err := row.Scan(&row_id, &raw)
	if err != nil {
		msbb.maybePrintError(err)
		return nil
	}

	taskPromise := &TaskPromise{}

	decodeEntry(raw, taskPromise)

	msbb.dbConn.Exec(fmt.Sprintf("DELETE FROM taskPromises WHERE id=%d", row_id))

	return taskPromise
}

func (msbb *MySQLSpiderBusBackend) SendItem(item *Item) error {
	raw := encodeEntry(item)

	msbb.dbConn.Exec("INSERT INTO items (raw) VALUES (?)", raw)

	return nil
}

func (msbb *MySQLSpiderBusBackend) ReceiveItem() *Item {
	var row_id int
	var raw []byte

	row := msbb.dbConn.QueryRow("SELECT * FROM items ORDER BY id ASC LIMIT 1")

	err := row.Scan(&row_id, &raw)
	if err != nil {
		msbb.maybePrintError(err)
		return nil
	}

	item := &Item{}

	decodeEntry(raw, item)

	msbb.dbConn.Exec(fmt.Sprintf("DELETE FROM items WHERE id=%d", row_id))

	return item
}

func (msbb *MySQLSpiderBusBackend) getCountForTable(tableName string) int {
	var output string

	query, _ := msbb.dbConn.Prepare(fmt.Sprintf("SELECT COUNT(*) FROM %s;", tableName))

	defer query.Close()

	query.QueryRow().Scan(&output)

	count, _ := strconv.Atoi(output)

	return count
}

func (msbb *MySQLSpiderBusBackend) IsEmpty() bool {
	nScheduledTasks := msbb.getCountForTable("scheduledTasks")
	nPromises := msbb.getCountForTable("taskPromises")
	nItems := msbb.getCountForTable("items")

	return nScheduledTasks == 0 && nPromises == 0 && nItems == 0
}

func (msbb *MySQLSpiderBusBackend) Close() {
	msbb.dbConn.Close()
}
