package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yeka/zip"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"

	// driver for sqlite3
	"github.com/mattn/go-sqlite3"
)

var moddingDS moddingDataStore
var genshinDS genshinDataStore
var commandDS commandDataStore
var serverDS serverDataStore
var dbMaintenance dbMaintenanceService

var errZeroRowsAffected = errors.New("zero rows were affected")
var errDuplicateCommand = errors.New("a command with the same name already exists in this server")

func createTables(db *sqlx.DB) {
	createTableDailyCheckInReminder(db)
	createTableParametricReminder(db)
	createTablePlayStoreReminder(db)
	createTableSimpleCommand(db)
	createTableCommandStats(db)
	createTableSpammableChannel(db)
	createTableUserWarning(db)
	createTableReact4RoleMessage(db)
	createTableServerProperties(db)
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
		"CreatedBy VARCHAR(20)",
		"UNIQUE(Key, GuildID)",
	}, db)
	createIndex("SimpleCommand", "Key", db)
}

func createTableCommandStats(db *sqlx.DB) {
	createTable("CommandStats", []string{
		"GuildID VARCHAR(20) NOT NULL DEFAULT ''",
		"Command VARCHAR(36) NOT NULL COLLATE NOCASE",
		"Count INTEGER NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"UNIQUE(GuildID, Command)",
	}, db)
	createIndex("CommandStats", "GuildID", db)
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
		"ChannelID VARCHAR(20) NOT NULL",
		"MessageID VARCHAR(20) NOT NULL",
		"EmojiID VARCHAR(20) NOT NULL",
		"EmojiName VARCHAR(32) NOT NULL",
		"RoleID VARCHAR(20) NOT NULL",
		"RequiredRoleID VARCHAR(20)",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
	}, db)
	createIndex("React4RoleMessage", "MessageID", db)
}

func createTableServerProperties(db *sqlx.DB) {
	createTable("ServerProperties", []string{
		"ServerID VARCHAR(20) NOT NULL",
		"PropertyName VARCHAR(32) NOT NULL",
		"PropertyValue TEXT NOT NULL",
		"CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
		"UNIQUE(ServerID, PropertyName)",
	}, db)
	createIndex("ServerProperties", "ServerID", db)
}

// commands

type commandDataStore struct {
	db *sqlx.DB
}

type CommandStat struct {
	GuildID string `db:"GuildID"`
	Command string `db:"Command"`
	Count   int    `db:"Count"`
}

func (c commandDataStore) addSimpleCommand(key, response, guildID, creatorUserID string) error {
	_, err := c.db.Exec(`INSERT INTO SimpleCommand (Key, Response, GuildID, CreatedBy) VALUES (?, ?, ?, ?)`,
		key, response, guildID, creatorUserID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return errDuplicateCommand
		}
	}
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

func (c commandDataStore) getCommandCreator(key, guildID string) (string, error) {
	var creator string
	err := c.db.Get(&creator, `SELECT CreatedBy FROM SimpleCommand WHERE Key = ? AND (GuildID = ?) COLLATE NOCASE`,
		key, guildID)
	return creator, err
}

func (c commandDataStore) simpleCommandResponse(key, guildID string) (string, error) {
	var response []string
	err := c.db.Select(&response, `
		SELECT Response FROM SimpleCommand
		WHERE Key = ? AND (GuildID = ? OR GuildID = '') COLLATE NOCASE
		ORDER BY CASE WHEN GuildID = '' THEN 0 ELSE 1 END
		LIMIT 1`,
		key, guildID)
	if len(response) == 0 {
		return "", err
	}
	return response[0], err
}

func (c commandDataStore) allSimpleCommandKeys(guildID string, includeGlobal bool) ([]string, error) {
	var keys []string
	if !includeGlobal {
		err := c.db.Select(&keys, `SELECT Key FROM SimpleCommand WHERE GuildID = ?`, guildID)
		return keys, err
	} else {
		err := c.db.Select(&keys, `SELECT Key FROM SimpleCommand WHERE GuildID = ? OR GuildID = ''`, guildID)
		return keys, err
	}
}

func (c commandDataStore) paginatedSimpleCommandKeys(guildID string, includeGlobal bool, page int, pageSize int) ([]string, error) {
	var keys []string
	if !includeGlobal {
		err := c.db.Select(&keys, `SELECT Key FROM SimpleCommand WHERE GuildID = ? ORDER BY Key LIMIT ? OFFSET ?`, guildID, pageSize, (page-1)*pageSize)
		return keys, err
	} else {
		err := c.db.Select(&keys, `SELECT Key FROM SimpleCommand WHERE GuildID = ? OR GuildID = '' ORDER BY Key LIMIT ? OFFSET ?`, guildID, pageSize, (page-1)*pageSize)
		return keys, err
	}
}

func (c commandDataStore) increaseCommandCountStat(guildID, commandKey string) error {
	_, err := c.db.Exec(`INSERT OR REPLACE INTO CommandStats (GuildID, Command, Count)
	                     VALUES (?, ?,
	                       COALESCE((SELECT Count FROM CommandStats WHERE GuildID=? AND Command=?), 0) + 1)`,
		guildID, commandKey, guildID, commandKey)
	return err
}

func (c commandDataStore) guildCommandStats(guildID string) ([]CommandStat, error) {
	var stats []CommandStat
	err := c.db.Select(&stats, `SELECT GuildID, Command, Count FROM CommandStats WHERE GuildID = ? ORDER BY Count DESC, Command ASC`, guildID)
	return stats, err
}

func (c commandDataStore) paginatedGuildCommandStats(guildID string, page int, pageSize int) ([]CommandStat, error) {
	var stats []CommandStat
	err := c.db.Select(&stats, `SELECT GuildID, Command, Count FROM CommandStats WHERE GuildID = ? ORDER BY Count DESC, Command ASC LIMIT ? OFFSET ?`, guildID, pageSize, (page-1)*pageSize)
	return stats, err
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
		_, err := s.db.Exec(`INSERT OR REPLACE INTO React4RoleMessage (ChannelID, MessageID, EmojiID, EmojiName, RoleID, RequiredRoleID) VALUES (?, ?, ?, ?, ?, ?)`,
			r4r.ChannelID, r4r.MessageID, r4r.EmojiID, r4r.EmojiName, r4r.RoleID, r4r.RequiredRoleID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s moddingDataStore) react4Roles(channelID, messageID string) ([]React4RoleMessage, error) {
	var r4rs []React4RoleMessage
	err := s.db.Select(&r4rs, `SELECT * FROM React4RoleMessage WHERE ChannelID = ? AND MessageID = ?`,
		channelID, messageID)
	return r4rs, err
}

func (s moddingDataStore) allReact4Roles() ([]React4RoleMessage, error) {
	var r4rs []React4RoleMessage
	err := s.db.Select(&r4rs, `SELECT * FROM React4RoleMessage`)
	return r4rs, err
}

func (s moddingDataStore) deleteReact4Roles(channelID, messageID string) error {
	_, err := s.db.Exec(`DELETE * FROM React4RoleMessage WHERE ChannelID = ? AND MessageID = ?`,
		channelID, messageID)
	return err
}

// server

type serverDataStore struct {
	db *sqlx.DB
}

type ServerProperty struct {
	ServerID      string `db:"ServerID"`
	PropertyName  string `db:"PropertyName"`
	PropertyValue string `db:"PropertyValue"`
}

func (s *serverDataStore) setServerProperty(serverID, propertyName, propertyValue string) error {
	_, err := s.db.Exec(`
		INSERT INTO ServerProperties (ServerID, PropertyName, PropertyValue) 
		VALUES (?, ?, ?) 
		ON CONFLICT(ServerID, PropertyName) 
		DO UPDATE SET PropertyValue = excluded.PropertyValue`,
		serverID, propertyName, propertyValue)
	return err
}

func (s *serverDataStore) getServerProperty(serverID, propertyName string) (string, error) {
	var propertyValue string
	err := s.db.Get(&propertyValue, `SELECT PropertyValue FROM ServerProperties WHERE ServerID = ? AND PropertyName = ?`,
		serverID, propertyName)
	return propertyValue, err
}

func (s *serverDataStore) getServerProperties(propertyName string) ([]ServerProperty, error) {
	var properties []ServerProperty
	err := s.db.Select(&properties, `SELECT ServerID, PropertyName, PropertyValue FROM ServerProperties WHERE PropertyName = ?`,
		propertyName)
	return properties, err
}

// maintenance

type dbMaintenanceService struct {
	db *sqlx.DB
}

func (s dbMaintenanceService) vacuum() error {
	_, err := s.db.Exec(`VACUUM`)
	return err
}

func (s dbMaintenanceService) analyze() error {
	_, err := s.db.Exec(`ANALYZE`)
	return err
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

// backups

func doDbBackup(ds *discordgo.Session) error {
	adminChannel, err := getUserChannel(adminID, ds)
	if err != nil {
		return errors.New("Could not get admin channel: " + err.Error())
	}

	reader, err := os.Open(dbFilename)
	if err != nil {
		return errors.New("Could not open database: " + err.Error())
	}
	defer reader.Close()

	zippedBackup, err := createEncryptedZipReader(reader, backupPassword)
	if err != nil {
		return errors.New("Could not create encrypted zip reader: " + err.Error())
	}

	todayStr := time.Now().Format("2006-01-02")
	_, err = ds.ChannelFileSend(adminChannel.ID, "jarvbot_db_"+todayStr+".zip", zippedBackup)
	if err != nil {
		return errors.New("Could not send backup: " + err.Error())
	}

	return nil
}

func backupCRONFunc(ds *discordgo.Session) func() {
	return func() {
		dbMaintenance.vacuum()
		log.Println("Database vacuum done")

		dbMaintenance.analyze()
		log.Println("Database analyze done")

		err := doDbBackup(ds)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Periodic backup done")
		}
	}
}

// createEncryptedZipReader returns a reader for a zip file with the given file encrypted inside using the given password and AES256Encryption
func createEncryptedZipReader(fileToZip *os.File, password string) (io.Reader, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	fileToZipBase := filepath.Base(fileToZip.Name())
	fileInfo, err := fileToZip.Stat()
	if err != nil {
		return nil, err
	}

	header := &zip.FileHeader{
		Name:   fileToZipBase,
		Method: zip.Deflate,
	}
	header.SetModTime(fileInfo.ModTime())
	header.SetMode(fileInfo.Mode())

	encryptedWriter, err := zipWriter.Encrypt(header.Name, password, zip.AES256Encryption)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(encryptedWriter, fileToZip)
	if err != nil {
		return nil, err
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return &buf, nil
}
