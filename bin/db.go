package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	// driver for sqlite3
	_ "github.com/mattn/go-sqlite3"
)

var moddingDS moddingDataStore
var genshinDS genshinDataStore
var commandDS commandDataStore

var errZeroRowsAffected = errors.New("zero rows were affected")

func createTables(db *sqlx.DB) {
	createTableDailyCheckInReminder(db)
	createTableParametricReminder(db)
	createTablePlayStoreReminder(db)
	createTableSimpleCommand(db)
	createTableSpammableChannel(db)
	createTableUserWarning(db)
	createTableReact4RoleMessage(db)
}

func createTableDailyCheckInReminder(db *sqlx.DB) {
	createTable("DailyCheckInReminder", []string{
		"DiscordUserID VARCHAR(20) UNIQUE NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
}

func createTableParametricReminder(db *sqlx.DB) {
	createTable("ParametricReminder", []string{
		"DiscordUserID VARCHAR(20) UNIQUE NOT NULL",
		"LastReminder TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("ParametricReminder", "LastReminder", db)
}

func createTablePlayStoreReminder(db *sqlx.DB) {
	createTable("PlayStoreReminder", []string{
		"DiscordUserID VARCHAR(20) UNIQUE NOT NULL",
		"LastReminder TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("PlayStoreReminder", "LastReminder", db)
}

func createTableSimpleCommand(db *sqlx.DB) {
	createTable("SimpleCommand", []string{
		"Key VARCHAR(36) NOT NULL COLLATE NOCASE",
		"Response TEXT NOT NULL",
		"GuildID VARCHAR(20) NOT NULL DEFAULT ''",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"UNIQUE(Key, GuildID)",
	}, db)
	createIndex("SimpleCommand", "Key", db)
}

func createTableSpammableChannel(db *sqlx.DB) {
	createTable("SpammableChannel", []string{
		"ChannelID VARCHAR(20) UNIQUE NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("SpammableChannel", "ChannelID", db)
}

func createTableUserWarning(db *sqlx.DB) {
	createTable("UserWarning", []string{
		"DiscordUserID VARCHAR(20) NOT NULL",
		"WarnedByID VARCHAR(20) NOT NULL",
		"GuildID VARCHAR(20) NOT NULL",
		"Reason VARCHAR(320) NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("UserWarning", "DiscordUserID", db)
}

func createTableReact4RoleMessage(db *sqlx.DB) {
	createTable("React4RoleMessage", []string{
		"MessageID VARCHAR(20) NOT NULL",
		"EmojiID VARCHAR(20) NOT NULL",
		"RoleID VARCHAR(20) NOT NULL",
		"RequiredRoleID VARCHAR(20)",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("React4RoleMessage", "MessageID", db)
}

// commands

type commandDataStore struct {
	db *sqlx.DB
}

func (c commandDataStore) addSimpleCommand(key, response, guildID string) error {
	_, err := c.db.Exec(`INSERT INTO SimpleCommand (Key, Response, GuildID) VALUES (?, ?, ?)`,
		key, response, guildID)
	return err
}

func (c commandDataStore) removeSimpleCommand(key, guildID string) error {
	res, err := c.db.Exec(`DELETE FROM SimpleCommand WHERE Key = ? AND GuildID = ?`,
		key, guildID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return errZeroRowsAffected
	}
	return nil
}

func (c commandDataStore) simpleCommandResponse(key, guildID string) (string, error) {
	var response []string
	err := c.db.Select(&response, `SELECT Response FROM SimpleCommand WHERE Key = ? AND (GuildID = ? OR GuildID = '') COLLATE NOCASE`,
		key, guildID)
	if len(response) == 0 {
		return "", err
	}
	return response[0], err
}

func (c commandDataStore) allSimpleCommandKeys(guildID string) ([]string, error) {
	var keys []string
	err := c.db.Select(&keys, `SELECT Key FROM SimpleCommand WHERE GuildID = ? OR GuildID = ''`, guildID)
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

// modding

type moddingDataStore struct {
	db *sqlx.DB
}

func (s moddingDataStore) warnUser(userID, modID, guildID, reason string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO UserWarning (DiscordUserID, WarnedByID, GuildID, Reason, CreatedAt) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		userID, modID, guildID, reason)
	return err
}

func (s moddingDataStore) userWarnings(userID, guildID string) ([]UserWarning, error) {
	warnings := []UserWarning{}
	err := s.db.Select(&warnings, `SELECT * FROM UserWarning WHERE DiscordUserID = ? AND GuildID = ?`,
		userID, guildID)
	return warnings, err
}

func (s moddingDataStore) addReact4Roles(r4rs []React4RoleMessage) error {
	for _, r4r := range r4rs {
		_, err := s.db.Exec(`INSERT OR REPLACE INTO React4RoleMessage (MessageID, EmojiID, RoleID, RequiredRoleID) VALUES (?, ?, ?, ?)`,
			r4r.MessageID, r4r.EmojiID, r4r.RoleID, r4r.RequiredRoleID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s moddingDataStore) react4Roles(messageID string) ([]React4RoleMessage, error) {
	var r4rs []React4RoleMessage
	err := s.db.Select(&r4rs, `SELECT * FROM React4RoleMessage WHERE MessageID = ?`,
		messageID)
	return r4rs, err
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
