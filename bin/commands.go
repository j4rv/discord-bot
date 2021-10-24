package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type command func(*discordgo.Session, *discordgo.MessageCreate, context.Context)

var commands = map[string]command{
	// public
	"!help":                      answerHelp,
	"!parametricTransformerStop": answerParametricTransformerStop,
	"!parametricTransformer":     answerParametricTransformer,
	"!ayayaify":                  answerAyayaify,
	"!remindme":                  answerRemindme,
	// hidden or easter eggs
	"!hello": answerHello,
	"!ruben": answerRuben,
	"!pablo": answerPablo,
	"!drive": answerDrive,
	// only available for the bot owner
	"!reboot":        doReboot,
	"!shutdown":      doShutdown,
	"!abortShutdown": doAbortShutdown,
}

const helpResponse = `Available commands:
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed
`

func answerHelp(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, helpResponse)
}

func answerHello(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID == adminID {
		ds.ChannelMessageSend(mc.ChannelID, "Hewwo master uwu")
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "Hello!")
	}
}

func answerDrive(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "J4RV's shared drive folder: https://drive.google.com/drive/folders/1JHlnWqoevIpZCHG4EdjQZN9vqJC0O8wA")
}

func answerRuben(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "carbo")
}

func answerPablo(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "gafas")
}

func answerAyayaify(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	bodyToAyayaify := strings.Replace(mc.Content, "!ayayaify ", "", 1)
	bodyToAyayaify = strings.Replace(bodyToAyayaify, "A", "AYAYA", 1)
	bodyToAyayaify = strings.Replace(bodyToAyayaify, "a", "ayaya", 1)
	ds.ChannelMessageSend(mc.ChannelID, bodyToAyayaify)
}

func answerParametricTransformer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	_, err := userMessageSend(mc.Author.ID, "I will remind you about the Parametric Transformer in 7 days!", ds, mc)
	if err != nil {
		return
	}

	ds.ChannelMessageSend(mc.ChannelID, "Gotcha!")
	startParametricReminder(mc.Author.ID, ds, mc, ctx)
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ok := stopParametricReminder(mc.Author.ID, ds, mc, ctx)
	if ok {
		ds.ChannelMessageSend(mc.ChannelID, "I'll stop reminding you "+mc.Author.Mention())
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "You weren't being reminded already "+mc.Author.Mention()+" but ok")
	}
}

func answerRemindme(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, reminderBody := processTimedCommand("!remindme", mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will remind you in %v with the message '%s'", timeToWait, reminderBody))
	time.Sleep(timeToWait)
	userMessageSend(mc.Author.ID, reminderBody, ds, mc)
}

func doReboot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID != adminID {
		mc.Author.Mention()
		ds.ChannelMessageSend(mc.ChannelID, userMustBeAdminMessage)
		return
	}
	reboot()
}

func doShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID != adminID {
		mc.Author.Mention()
		ds.ChannelMessageSend(mc.ChannelID, userMustBeAdminMessage)
		return
	}
	timeToWait, _ := processTimedCommand("!shutdown", mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will shutdown in %v", timeToWait))
	err := shutdown(timeToWait)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
	}
}

func doAbortShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID != adminID {
		mc.Author.Mention()
		ds.ChannelMessageSend(mc.ChannelID, userMustBeAdminMessage)
		return
	}
	ds.ChannelMessageSend(mc.ChannelID, "Gotcha!")
	err := abortShutdown()
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
	}
}
