package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const userMustBeAdminMessage = "Only the bot's admin can do that"
const commandReceivedMessage = "Gotcha!"
const dmNotSentError = "Failed to send you a DM. Did you disable DMs in your privacy settings?"

var commandPrefixRegex = regexp.MustCompile(`^!\w+\s*`)

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
	"!randomAbyssLineup":         answerRandomAbyssLineup,
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
- **!randomAbyssLineup**: The bot will give you two random teams and some replacements. Have fun ¯\_(ツ)_/¯. Optional: Write 8+ character names separated by commas and the bot will only choose from those
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
	startParametricReminder(ds, mc, ctx)
	ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the Parametric Transformer in 7 days!")
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ok := stopParametricReminder(ds, mc, ctx)
	if ok {
		ds.ChannelMessageSend(mc.ChannelID, "I'll stop reminding you "+mc.Author.Mention())
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "You weren't being reminded already "+mc.Author.Mention()+" but ok")
	}
}

func answerRandomAbyssLineup(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	var firstTeam, secondTeam [4]string
	var replacements []string

	// Process Input and generate the teams
	inputString := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	if inputString != "" {
		log.Println("inputstr", inputString)
		inputchars := strings.Split(inputString, ",")
		if len(inputchars) < genshinTeamSize*2 {
			ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf(`Not enough characters! Please enter at least %d`, genshinTeamSize*2))
			return
		}
		firstTeam, secondTeam, replacements = randomAbyssLineup(inputchars...)
	} else {
		log.Println("inputstr", inputString)
		firstTeam, secondTeam, replacements = randomAbyssLineup()
	}

	// Format the teams into readable text
	formattedFirstTeam, formattedSecondTeam, formattedReplacements := "```\n", "```\n", "```\n"
	for _, r := range replacements {
		formattedReplacements += r + "\n"
	}
	for i := 0; i < genshinTeamSize; i++ {
		formattedFirstTeam += firstTeam[i] + "\n"
		formattedSecondTeam += secondTeam[i] + "\n"
	}
	formattedFirstTeam += "```"
	formattedSecondTeam += "```"
	formattedReplacements += "```"

	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf(`
**First half:**
%s
**Second half:**
%s
**Replacements:**
%s
`, formattedFirstTeam, formattedSecondTeam, formattedReplacements))
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
	userMessageSend(mc.Author.ID, reminderBody, ds)
}

func answerReboot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := reboot()
	if err != nil {
		errorMessageSend("Error: "+err.Error(), ds, mc)
	}
}

func answerShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, _ := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will shutdown in %v", timeToWait))
	err := shutdown(timeToWait)
	if err != nil {
		errorMessageSend("Error: "+err.Error(), ds, mc)
	}
}

func answerAbortShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	err := abortShutdown()
	if err != nil {
		errorMessageSend("Error: "+err.Error(), ds, mc)
	}
}
