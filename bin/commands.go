package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const userMustBeAdminMessage = "Only the bot's admin can do that"
const commandReceivedMessage = "Gotcha!"

type command func(*discordgo.Session, *discordgo.MessageCreate, context.Context)

func adminOnly(wrapped command) command {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
		if mc.Author.ID != adminID {
			mc.Author.Mention()
			ds.ChannelMessageSend(mc.ChannelID, userMustBeAdminMessage)
			return
		}
		wrapped(ds, mc, ctx)
	}
}

var commands = map[string]command{
	// public
	"!help":                      answerHelp,
	"!genshinDailyCheckInStop":   answerGenshinDailyCheckInStop,
	"!genshinDailyCheckIn":       answerGenshinDailyCheckIn,
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
	"!reboot":        adminOnly(answerReboot),
	"!shutdown":      adminOnly(answerShutdown),
	"!abortShutdown": adminOnly(answerAbortShutdown),
}

const helpResponse = `Available commands:
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed
- **!genshinDailyCheckIn**: Will remind you to do the Genshin Daily Check-In
- **!genshinDailyCheckInStop**: The bot will stop reminding you to do the Genshin Daily Check-In
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
`

const adminOnlyCommands = `
Admin only:
- **!reboot**: Reboot the bot's system
- **!shutdown** [99h 99m 99s]: Shuts down the bot's system
- **!abortShutdown**: Aborts the bot's system shutdown
`

func answerHelp(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID == adminID {
		ds.ChannelMessageSend(mc.ChannelID, helpResponse+adminOnlyCommands)
	} else {
		ds.ChannelMessageSend(mc.ChannelID, helpResponse)
	}
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
	bodyToAyayaify = strings.ReplaceAll(bodyToAyayaify, "A", "AYAYA")
	bodyToAyayaify = strings.ReplaceAll(bodyToAyayaify, "a", "ayaya")
	ds.ChannelMessageSend(mc.ChannelID, bodyToAyayaify)
}

func answerParametricTransformer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	_, err := userMessageSend(mc.Author.ID, "I will remind you about the Parametric Transformer in 7 days!", ds, mc)
	if err != nil {
		return
	}

	startParametricReminder(ds, mc, ctx)
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ok := stopParametricReminder(ds, mc, ctx)
	if ok {
		ds.ChannelMessageSend(mc.ChannelID, "I'll stop reminding you "+mc.Author.Mention())
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "You weren't being reminded already "+mc.Author.Mention()+" but ok")
	}
}

func answerGenshinDailyCheckIn(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	startDailyCheckInReminder(ds, mc, ctx)
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
}

func answerGenshinDailyCheckInStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	stopDailyCheckInReminder(ds, mc, ctx)
	ds.ChannelMessageSend(mc.ChannelID, "I'll stop reminding you "+mc.Author.Mention())
}

func answerRemindme(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, reminderBody := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will remind you in %v with the message '%s'", timeToWait, reminderBody))
	time.Sleep(timeToWait)
	userMessageSend(mc.Author.ID, reminderBody, ds, mc)
}

func answerReboot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	reboot()
}

func answerShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, _ := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will shutdown in %v", timeToWait))
	err := shutdown(timeToWait)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
	}
}

func answerAbortShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	err := abortShutdown()
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
	}
}
