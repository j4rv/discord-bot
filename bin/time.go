package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var stringHoursRegex = regexp.MustCompile(`^(\d{1,2})h\s*`)
var stringMinsRegex = regexp.MustCompile(`^(\d{1,2})m\s*`)
var stringSecsRegex = regexp.MustCompile(`^(\d{1,2})s\s*`)

// Format: "!remindme 99h 99m 99s <body>"
// Returns: The duration and the body
func processTimedCommand(commandKey string, commandBody string) (time.Duration, string) {
	if len(commandBody) < len(commandKey) {
		return 0, ""
	}

	var result time.Duration

	// remove the prefix
	commandBody = commandBody[len(commandKey):]
	commandBody = strings.TrimLeft(commandBody, " ")

	n, commandBody := extractTimeUnit(commandBody, stringHoursRegex)
	result += time.Duration(n) * time.Hour

	n, commandBody = extractTimeUnit(commandBody, stringMinsRegex)
	result += time.Duration(n) * time.Minute

	n, commandBody = extractTimeUnit(commandBody, stringSecsRegex)
	result += time.Duration(n) * time.Second

	return result, commandBody
}

// From a string like "5m blablabla", uses the regexp provided (for example, stringMinsRegex)
// to remove the duration part ("5m ")
// and return the correct number of duration units (5)
func extractTimeUnit(s string, re *regexp.Regexp) (int, string) {
	found := re.FindStringSubmatch(s)
	if len(found) == 0 {
		return 0, s
	}
	s = re.ReplaceAllString(s, "")
	s = strings.TrimLeft(s, " ")
	foundInt, _ := strconv.Atoi(found[1])
	return foundInt, s
}
