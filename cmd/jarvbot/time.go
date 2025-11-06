package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var stringHoursRegex = regexp.MustCompile(`^(\d{1,2})h`)
var stringMinsRegex = regexp.MustCompile(`^(\d{1,2})m`)
var stringSecsRegex = regexp.MustCompile(`^(\d{1,2})s`)

const secondsInADay = 60 * 60 * 24

// only for Shadow Realm role, does not allow duplicates
func removeShadowRealmRoleAfterDuration(guildID, memberID, roleID string, duration time.Duration) {
	data := guildID + ";" + roleID
	currentCooldowns, _ := schedulerDS.getScheduledActionsByTargetIDAndActionTypeAndActionData(
		memberID, actionTypeRemoveRole, data,
	)
	newTime := time.Now().UTC().Add(duration)

	// If there exists a longer cooldown, just return
	for _, o := range currentCooldowns {
		if o.ScheduledFor.After(newTime) {
			return
		}
	}

	if err := schedulerDS.addScheduledAction(newTime, memberID, targetTypeUser, actionTypeRemoveRole, data); err != nil {
		log.Printf("addScheduledAction failed: %v", err)
		return
	}
	for _, o := range currentCooldowns {
		_ = schedulerDS.removeScheduledAction(o.ID)
	}
}

var usersOnExpensiveOperationCooldown = make(map[string]struct{})

func userExpensiveOperationOnCooldown(userID string) bool {
	if userID == adminID {
		return false
	}
	_, inCooldown := usersOnExpensiveOperationCooldown[userID]
	return inCooldown
}

func userExecutedExpensiveOperation(userID string) {
	usersOnExpensiveOperationCooldown[userID] = struct{}{}

	go func() {
		time.Sleep(expensiveOperationCooldown)
		delete(usersOnExpensiveOperationCooldown, userID)
	}()
}

// Format: "!<command> 99h 99m 99s <body>"
// Returns: The duration and the body
// TODO: Extract to an utilities library?
func processTimedCommand(commandBody string) (time.Duration, string) {
	var result time.Duration
	commandBody = commandPrefixRegex.ReplaceAllString(commandBody, "")

	n, commandBody := extractTimeUnit(commandBody, stringHoursRegex)
	result += time.Duration(n) * time.Hour

	n, commandBody = extractTimeUnit(commandBody, stringMinsRegex)
	result += time.Duration(n) * time.Minute

	n, commandBody = extractTimeUnit(commandBody, stringSecsRegex)
	result += time.Duration(n) * time.Second

	return result, commandBody
}

func stringToDuration(s string) time.Duration {
	var result time.Duration
	n, s := extractTimeUnit(s, stringHoursRegex)
	result += time.Duration(n) * time.Hour
	n, s = extractTimeUnit(s, stringMinsRegex)
	result += time.Duration(n) * time.Minute
	n, _ = extractTimeUnit(s, stringSecsRegex)
	result += time.Duration(n) * time.Second
	return result
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

func unixDay() int64 {
	return time.Now().Unix() / secondsInADay
}
