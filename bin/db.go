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
var commandDS commandDataStore

func initDB() {
	db := sqlx.MustOpen("sqlite3", dbFilename)
	if db.Ping() != nil {
		panic("DB did not answer ping")
	}
	createTables(db)
	genshinDS = genshinDataStore{db}
	commandDS = commandDataStore{db}
}

func createTables(db *sqlx.DB) {
	createTableDailyCheckInReminder(db)
	createTableParametricReminder(db)
	createTablePlayStoreReminder(db)
	createTableSimpleCommand(db)
	createTableSpammableChannel(db)
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

func createTablePlayStoreReminder(db *sqlx.DB) {
	createTable("PlayStoreReminder", []string{
		"DiscordUserID VARCHAR(18) UNIQUE NOT NULL",
		"LastReminder TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("PlayStoreReminder", "LastReminder", db)
}

func createTableSimpleCommand(db *sqlx.DB) {
	createTable("SimpleCommand", []string{
		"Key VARCHAR(36) UNIQUE NOT NULL COLLATE NOCASE",
		"Response TEXT NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("SimpleCommand", "Key", db)
}

func createTableSpammableChannel(db *sqlx.DB) {
	createTable("SpammableChannel", []string{
		"ChannelID VARCHAR(18) UNIQUE NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("SpammableChannel", "ChannelID", db)
}

// commands

type commandDataStore struct {
	db *sqlx.DB
}

func (c commandDataStore) addSimpleCommand(key, response string) error {
	_, err := c.db.Exec(`INSERT INTO SimpleCommand (Key, Response) VALUES (?, ?)`,
		key, response)
	return err
}

func (c commandDataStore) removeSimpleCommand(key string) error {
	_, err := c.db.Exec(`DELETE FROM SimpleCommand WHERE Key = ?`,
		key)
	return err
}

func (c commandDataStore) simpleCommandResponse(key string) (string, error) {
	var response []string
	err := c.db.Select(&response, `SELECT Response FROM SimpleCommand WHERE Key = ? COLLATE NOCASE`, key)
	if len(response) == 0 {
		return "", err
	}
	return response[0], err
}

func (c commandDataStore) allSimpleCommandKeys() ([]string, error) {
	var keys []string
	err := c.db.Select(&keys, `SELECT Key FROM SimpleCommand`)
	return keys, err
}

func (c commandDataStore) addSpammableChannel(channelID string) error {
	_, err := c.db.Exec(`INSERT INTO SpammableChannel (ChannelID) VALUES (?)`,
		channelID)
	return err
}

func (c commandDataStore) removeSpammableChannel(channelID string) error {
	_, err := c.db.Exec(`DELETE FROM SpammableChannel WHERE ChannelID = ?`,
		channelID)
	return err
}

func (c commandDataStore) isChannelSpammable(channelID string) (bool, error) {
	var isSpammable []uint8
	err := c.db.Select(&isSpammable, `SELECT 1 FROM SpammableChannel WHERE ChannelID = ?`, channelID)
	if len(isSpammable) == 0 {
		return false, err
	}
	return true, err
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

func (s genshinDataStore) addOrUpdatePlayStoreReminder(userID string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO PlayStoreReminder (DiscordUserID, LastReminder) VALUES (?, CURRENT_TIMESTAMP)`,
		userID)
	return err
}

func (s genshinDataStore) removePlayStoreReminder(userID string) error {
	_, err := s.db.Exec(`DELETE FROM PlayStoreReminder WHERE DiscordUserID = ?`,
		userID)
	return err
}

func (s genshinDataStore) allPlayStoreReminderUserIDsToBeReminded() ([]string, error) {
	var userIDs []string
	err := s.db.Select(&userIDs, `SELECT DiscordUserID FROM PlayStoreReminder WHERE LastReminder <= datetime('now', '-7 days')`)
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
