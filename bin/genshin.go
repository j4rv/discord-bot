package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var usersWithParametricReminder = map[string]context.CancelFunc{}

func startParametricReminder(userID string, ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	stopParametricReminder(userID, ds, mc, ctx)
	cancellableCtx, cancel := context.WithCancel(ctx)
	usersWithParametricReminder[userID] = cancel
	runParametricReminder(userID, ds, mc, cancellableCtx)
}

func stopParametricReminder(userID string, ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	cancelExistingReminder, ok := usersWithParametricReminder[userID]
	if ok {
		cancelExistingReminder()
		delete(usersWithParametricReminder, userID)
	}
	return ok
}

// could be better with a CRON library? ... ¯\_(ツ)_/¯
func runParametricReminder(userID string, ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	select {
	case <-time.After(7 * 24 * time.Hour):
		_, err := userMessageSend(userID, "Remember to use the Parametric Transformer!", ds, mc)
		if err != nil {
			return
		}
		runParametricReminder(userID, ds, mc, ctx)
	case <-ctx.Done():
		fmt.Println("stopped ParametricReminder for user", userID)
	}
}
