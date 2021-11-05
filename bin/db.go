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
	createTables(db)
	genshinDS = genshinDataStore{db}
}

func createTables(db *sqlx.DB) {
	createTableDailyCheckInReminder(db)
	createTableParametricReminder(db)
}

func createTableDailyCheckInReminder(db *sqlx.DB) {
	createTable("DailyCheckInReminder", []string{
		"DiscordUserID VARCHAR(18) UNIQUE NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
}

func createTableParametricReminder(db *sqlx.DB) {
	createTable("ParametricReminder", []string{
		"DiscordUserID VARCHAR(18) UNIQUE NOT NULL",
		"LastReminder TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("ParametricReminder", "LastReminder", db)
}

// genshin

type genshinDataStore struct {
	db *sqlx.DB
}

func (s genshinDataStore) addDailyCheckInReminder(userID string) error {
	_, err := s.db.Exec(`INSERT INTO DailyCheckInReminder (DiscordUserID) VALUES (?)`,
		userID)
	return err
}

func (s genshinDataStore) removeDailyCheckInReminder(userID string) error {
	_, err := s.db.Exec(`DELETE FROM DailyCheckInReminder WHERE DiscordUserID = ?`,
		userID)
	return err
}

func (s genshinDataStore) allDailyCheckInReminderUserIDs() ([]string, error) {
	var userIDs []string
	err := s.db.Select(&userIDs, `SELECT DiscordUserID FROM DailyCheckInReminder`)
	return userIDs, err
}

func (s genshinDataStore) addOrUpdateParametricReminder(userID string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO ParametricReminder (DiscordUserID, LastReminder) VALUES (?, CURRENT_TIMESTAMP)`,
		userID)
	return err
}

func (s genshinDataStore) removeParametricReminder(userID string) error {
	_, err := s.db.Exec(`DELETE FROM ParametricReminder WHERE DiscordUserID = ?`,
		userID)
	return err
}

func (s genshinDataStore) allParametricReminderUserIDsToBeReminded() ([]string, error) {
	var userIDs []string
	err := s.db.Select(&userIDs, `SELECT DiscordUserID FROM ParametricReminder WHERE LastReminder <= datetime('now', '-7 days')`)
	return userIDs, err
}

// methods for repetitive stuff

func createTable(table string, columns []string, db *sqlx.DB) {
	if len(columns) == 0 {
		panic("createTable method is for tables with at least one column")
	}
	statement := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s INTEGER PRIMARY KEY AUTOINCREMENT,%s);",
		table, table, strings.Join(columns, ","))
	db.MustExec(statement)
}

func createIndex(table, column string, db *sqlx.DB) {
	indexName := fmt.Sprintf("%s_%s", table, column)
	statement := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s);", indexName, table, column)
	db.MustExec(statement)
}
