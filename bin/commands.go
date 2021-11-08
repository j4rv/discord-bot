package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/lib/genshinartis"
)

const userMustBeAdminMessage = "Only the bot's admin can do that"
const commandReceivedMessage = "Gotcha!"
const commandErrorHappened = "I could not do that :( sorry"
const dmNotSentError = "Could not send you a DM. Did you disable DMs in your privacy settings?"

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
	"!source":                    simpleTextResponse("Source code: https://github.com/j4rv/discord-bot"),
	"!genshinDailyCheckIn":       answerGenshinDailyCheckIn,
	"!genshinDailyCheckInStop":   answerGenshinDailyCheckInStop,
	"!parametricTransformer":     answerParametricTransformer,
	"!parametricTransformerStop": answerParametricTransformerStop,
	"!randomAbyssLineup":         answerRandomAbyssLineup,
	"!randomArtifact":            answerRandomArtifact,
	"!randomArtifactSet":         answerRandomArtifactSet,
	"!ayayaify":                  answerAyayaify,
	"!remindme":                  answerRemindme,
	// hidden or easter eggs
	"!hello":   answerHello,
	"!jarv":    simpleTextResponse("Hipster metaslave"),
	"!naz":     simpleTextResponse("Retired whale™️"),
	"!jam":     simpleTextResponse(":3c"),
	"!spiwar":  simpleTextResponse("C1 bedo"),
	"!kono":    simpleTextResponse("Adorable"),
	"!rosb":    simpleTextResponse(":BASED:"),
	"!twns":    simpleTextResponse("Do your own math"),
	"!vins":    simpleTextResponse("Gold albedolympics medalist"),
	"!moona":   simpleTextResponse("Bandori master!"),
	"!ossify":  simpleTextResponse("Based husbando puller"),
	"!teatime": simpleTextResponse("Waifu dominoes master"),
	"!bmaster": simpleTextResponse("Professional shitposter"),
	"!ruben":   simpleTextResponse("Carbo"),
	"!pablo":   simpleTextResponse("Gafas"),
	"!gura":    simpleTextResponse("A"),
	// only available for the bot owner
	"!reboot":        adminOnly(answerReboot),
	"!shutdown":      adminOnly(answerShutdown),
	"!abortShutdown": adminOnly(answerAbortShutdown),
}

const helpResponse = `Available commands:
- **!source**: Links to the bot's source code
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed
- **!genshinDailyCheckIn**: Will remind you to do the Genshin Daily Check-In
- **!genshinDailyCheckInStop**: The bot will stop reminding you to do the Genshin Daily Check-In
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days. Use it again to reset the reminder
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!randomAbyssLineup**: The bot will give you two random teams and some replacements. Have fun ¯\_(ツ)_/¯. Optional: Write 8+ character names separated by commas and the bot will only choose from those
- **!randomArtifact**: Generates a random Lv20 Genshin Impact artifact
- **!randomArtifactSet**: Generates five random Lv20 Genshin Impact artifacts
`

const helpResponseAdmin = helpResponse + `
Admin only:
- **!reboot**: Reboot the bot's system
- **!shutdown** [99h 99m 99s]: Shuts down the bot's system
- **!abortShutdown**: Aborts the bot's system shutdown
`

func answerHelp(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID != adminID {
		ds.ChannelMessageSend(mc.ChannelID, helpResponse)
	} else {
		ds.ChannelMessageSend(mc.ChannelID, helpResponseAdmin)
	}
}

func answerHello(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	if mc.Author.ID == adminID {
		ds.ChannelMessageSend(mc.ChannelID, "Hewwo master uwu")
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "Hello!")
	}
}

func simpleTextResponse(body string) func(*discordgo.Session, *discordgo.MessageCreate, context.Context) {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
		ds.ChannelMessageSend(mc.ChannelID, body)
	}
}

func answerAyayaify(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	bodyToAyayaify := strings.Replace(mc.Content, "!ayayaify ", "", 1)
	bodyToAyayaify = strings.ReplaceAll(bodyToAyayaify, "A", "AYAYA")
	bodyToAyayaify = strings.ReplaceAll(bodyToAyayaify, "a", "ayaya")
	ds.ChannelMessageSend(mc.ChannelID, bodyToAyayaify)
}

func answerParametricTransformer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := startParametricReminder(ds, mc, ctx)
	checkErr("answerParametricTransformer", err, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, errorMessage(commandErrorHappened))
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the Parametric Transformer in 7 days!")
	}
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := stopParametricReminder(ds, mc, ctx)
	checkErr("answerParametricTransformerStop", err, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, errorMessage(commandErrorHappened))
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
}

func answerRandomAbyssLineup(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	var firstTeam, secondTeam [4]string
	var replacements []string

	// Process Input and generate the teams
	inputString := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	inputChars := strings.Split(inputString, ",")
	if inputChars[0] != "" && len(inputChars) < genshinTeamSize*2 {
		ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf(`Not enough characters! Please enter at least %d`, genshinTeamSize*2))
		return
	}
	for i := range inputChars {
		inputChars[i] = strings.TrimSpace(inputChars[i])
	}
	firstTeam, secondTeam, replacements = randomAbyssLineup(inputChars...)

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

func answerRandomArtifact(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	artifact := genshinartis.RandomArtifact()
	ds.ChannelMessageSend(mc.ChannelID, formatGenshinArtifact(artifact))
}

func answerRandomArtifactSet(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	flower := genshinartis.RandomArtifactOfSlot(genshinartis.SlotFlower)
	plume := genshinartis.RandomArtifactOfSlot(genshinartis.SlotPlume)
	sands := genshinartis.RandomArtifactOfSlot(genshinartis.SlotSands)
	goblet := genshinartis.RandomArtifactOfSlot(genshinartis.SlotGoblet)
	circlet := genshinartis.RandomArtifactOfSlot(genshinartis.SlotCirclet)
	msg := formatGenshinArtifact(flower)
	msg += formatGenshinArtifact(plume)
	msg += formatGenshinArtifact(sands)
	msg += formatGenshinArtifact(goblet)
	msg += formatGenshinArtifact(circlet)
	ds.ChannelMessageSend(mc.ChannelID, msg)
}

func answerGenshinDailyCheckIn(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	startDailyCheckInReminder(ds, mc, ctx)
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
}

func answerGenshinDailyCheckInStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	stopDailyCheckInReminder(ds, mc, ctx)
	ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
}

// FIXME: Limit its usage by user (max 3 active reminders?)
func answerRemindme(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, reminderBody := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will remind you in %v with the message '%s'", timeToWait, reminderBody))
	time.Sleep(timeToWait)
	userMessageSend(mc.Author.ID, reminderBody, ds)
}

func answerReboot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := reboot()
	checkErr("reboot", err, ds)
}

func answerShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, _ := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will shutdown in %v", timeToWait))
	err := shutdown(timeToWait)
	checkErr("shutdown", err, ds)
}

func answerAbortShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	err := abortShutdown()
	checkErr("abortShutdown", err, ds)
}
