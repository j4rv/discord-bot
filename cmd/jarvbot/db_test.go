package main

import (
	"os"
	"testing"
)

func initTestDB() {
	dbFilename = "test.sqlite"
	if _, err := os.Stat(dbFilename); err == nil {
		err = os.Remove(dbFilename)
		if err != nil {
			panic("Failed to delete the database file: " + err.Error())
		}
	}
	initDB()
}

func TestServerProperties(t *testing.T) {
	initTestDB()

	_, err := serverDS.getServerProperty("0000", "key")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	serverDS.setServerProperty("0000", "key", "value")
	val, err := serverDS.getServerProperty("0000", "key")
	if err != nil {
		t.Error(err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got '%s'", val)
	}

	err = serverDS.setServerProperty("0000", "key", "value2")
	if err != nil {
		t.Error(err)
	}
	val, err = serverDS.getServerProperty("0000", "key")
	if err != nil {
		t.Error(err)
	}
	if val != "value2" {
		t.Errorf("Expected 'value2', got '%s'", val)
	}
}
