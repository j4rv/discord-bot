package main

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	// driver for sqlite3
	_ "github.com/mattn/go-sqlite3"
)

const dbFilename = "db.sqlite"

var genshinDS genshinDataStore

func initDB() {
	db := sqlx.MustOpen("sqlite3", dbFilename)
	if db.Ping() != nil {
		panic("DB did not answer ping")
	}
	CreateTables(db)
	genshinDS = genshinDataStore{db}
}

func CreateTables(db *sqlx.DB) {
	createTableDailyCheckInReminder(db)
}

func createTableDailyCheckInReminder(db *sqlx.DB) {
	createTable("DailyCheckInReminder", []string{
		"DiscordUserID VARCHAR(18) UNIQUE",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("DailyCheckInReminder", "DiscordUserID", db)
}

// genshin

type genshinDataStore struct {
	db *sqlx.DB
}

func (s genshinDataStore) AddDailyCheckInReminder(userID string) error {
	_, err := s.db.Exec(`INSERT INTO DailyCheckInReminder (DiscordUserID) VALUES (?)`,
		userID)
	return err
}

func (s genshinDataStore) RemoveDailyCheckInReminder(userID string) error {
	_, err := s.db.Exec(`DELETE FROM DailyCheckInReminder WHERE DiscordUserID = ?`,
		userID)
	return err
}

func (s genshinDataStore) AllDailyCheckInReminderUserIDs() []string {
	var userIDs []string
	s.db.Select(&userIDs, `SELECT DiscordUserID FROM DailyCheckInReminder`)
	return userIDs
}

// methods for repetitive stuff

func createTable(table string, columns []string, db *sqlx.DB) {
	if len(columns) == 0 {
		panic("createTable method is for tables with at least one column")
	}
	// Using Sprintf since this internal method does not use user inputs
	statement := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s INTEGER PRIMARY KEY AUTOINCREMENT,%s);", table, table, strings.Join(columns, ","))
	db.MustExec(statement)
}

func createIndex(table, column string, db *sqlx.DB) {
	indexName := fmt.Sprintf("%s_%s", table, column)
	// Using Sprintf since this internal method does not use user inputs
	createIndexStatement := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s);", indexName, table, column)
	db.MustExec(createIndexStatement)
}
