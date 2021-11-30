package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/lib/genshinartis"
)

const userMustBeAdminMessage = "Only the bot's admin can do that"
const commandReceivedMessage = "Gotcha!"
const commandSuccessMessage = "Successfully donette!"
const commandErrorHappened = "I could not do that :( sorry"
const commandWithTwoArgumentsError = "Something went wrong, please make sure that the command has two arguments with the following format: '!command (...) (...)'"
const commandWithMentionError = "Something went wrong, please make sure that the command has an user mention"

const roleEveryone = "@everyone"

var commandPrefixRegex = regexp.MustCompile(`^!\w+\s*`)
var commandWithTwoArguments = regexp.MustCompile(`^!\w+\s*(\(.{1,36}\))\s*(\(.{1,36}\))`)
var commandWithMention = regexp.MustCompile(`^!\w+\s*<@!?(\d{18})>`)

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

// the command key must be lowercased
var commands = map[string]command{
	// public
	"!help":                      answerHelp,
	"!source":                    simpleTextResponse("Source code: https://github.com/j4rv/discord-bot"),
	"!genshindailycheckin":       answerGenshinDailyCheckIn,
	"!genshindailycheckinstop":   answerGenshinDailyCheckInStop,
	"!parametrictransformer":     answerParametricTransformer,
	"!parametrictransformerstop": answerParametricTransformerStop,
	"!randomabysslineup":         answerRandomAbyssLineup,
	"!randomartifact":            answerRandomArtifact,
	"!randomartifactset":         answerRandomArtifactSet,
	"!randomdomainrun":           answerRandomDomainRun,
	"!ayayaify":                  answerAyayaify,
	"!remindme":                  answerRemindme,
	"!roll":                      answerRoll,
	// hidden or easter eggs
	"!hello":  answerHello,
	"!liquid": answerLiquid,
	"!don":    answerDon,
	//"!shoot":  answerShoot,
	// only available for the bot owner
	"!roleids":         adminOnly(answerRoleIDs),
	"!addcommand":      adminOnly(answerAddCommand),
	"!removecommand":   adminOnly(answerRemoveCommand),
	"!listcommands":    adminOnly(answerListCommands),
	"!allowspamming":   adminOnly(answerAllowSpamming),
	"!preventspamming": adminOnly(answerPreventSpamming),
	"!reboot":          adminOnly(answerReboot),
	"!shutdown":        adminOnly(answerShutdown),
	"!abortshutdown":   adminOnly(answerAbortShutdown),
}

func processCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, fullCommand string, ctx context.Context) {
	channelIsSpammable, err := commandDS.isChannelSpammable(mc.ChannelID)
	notifyIfErr("processCommand: isChannelSpammable", err, ds)
	if !channelIsSpammable && isUserOnCooldown(mc.Author.ID) {
		userMessageSend(mc.Author.ID, "Pls don't spam the bot commands uwu", ds)
		ds.MessageReactionAdd(mc.ChannelID, mc.Message.ID, "❌")
		return
	}

	commandKey := strings.TrimSpace(commandPrefixRegex.FindString(fullCommand))
	command, ok := commands[strings.ToLower(commandKey)]
	if ok {
		onSuccessCommandCall(mc, channelIsSpammable)
		command(ds, mc, ctx)
		return
	}

	response, err := commandDS.simpleCommandResponse(commandKey)
	notifyIfErr("simpleCommandResponse", err, ds)
	if err == nil {
		onSuccessCommandCall(mc, channelIsSpammable)
		ds.ChannelMessageSend(mc.ChannelID, response)
	}
}

func onSuccessCommandCall(mc *discordgo.MessageCreate, channelIsSpammable bool) {
	log.Printf("[%s] [%s] %s", mc.ChannelID, mc.Author.Username, mc.Content)
	if !channelIsSpammable {
		resetUserCooldown(mc.Author.ID)
	}
}

const helpResponse = `Available commands:
- **!source**: Links to the bot's source code
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed
- **!roll [99]**: Rolls a dice with the specified sides amount
- **!genshinDailyCheckIn**: Will remind you to do the Genshin Daily Check-In
- **!genshinDailyCheckInStop**: The bot will stop reminding you to do the Genshin Daily Check-In
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days. Use it again to reset the reminder
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!randomAbyssLineup**: The bot will give you two random teams and some replacements. Have fun ¯\_(ツ)_/¯. Optional: Write 8+ character names separated by commas and the bot will only choose from those
- **!randomArtifact**: Generates a random Lv20 Genshin Impact artifact
- **!randomArtifactSet**: Generates five random Lv20 Genshin Impact artifacts
- **!randomDomainRun (set A) (set B)**: Generates two random Lv20 Genshin Impact artifacts from the input sets
`

const helpResponseAdmin = helpResponse + `
Admin only:
- **!addCommand [!key] [response]**: Adds a simple command
- **!removeCommand [!key]**: Removes a simple command
- **!listCommands**: Lists all current simple commands
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

func answerLiquid(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("%06d, you know what to do with this. ", rand.Intn(1000000)))
}

func answerDon(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeoutRole, err := guildRoleByName(ds, mc.GuildID, timeoutRoleName)
	notifyIfErr("answerDon", err, ds)
	if err != nil {
		return
	}
	err = ds.GuildMemberRoleAdd(mc.GuildID, mc.Author.ID, timeoutRole.ID)
	notifyIfErr("answerDon", err, ds)
	if err != nil {
		return
	}
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("To the Shadow Realm you go %s", mc.Author.Mention()))
}

func answerShoot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	match := commandWithMention.FindStringSubmatch(mc.Content)
	if match == nil || len(match) != 2 {
		ds.ChannelMessageSend(mc.ChannelID, commandWithMentionError)
		return
	}

	timeoutRole, err := guildRoleByName(ds, mc.GuildID, timeoutRoleName)
	notifyIfErr("answerShoot: get timeout role", err, ds)
	if err != nil {
		return
	}

	shooter, err := ds.GuildMember(mc.GuildID, mc.Author.ID)
	notifyIfErr("answerShoot: get shooter member", err, ds)
	if err != nil {
		return
	}

	target, err := ds.GuildMember(mc.GuildID, match[1])
	notifyIfErr("answerShoot: get target member", err, ds)
	if err != nil {
		return
	}

	err = shoot(ds, mc.ChannelID, mc.GuildID, shooter, target, timeoutRole.ID)
	notifyIfErr("answerShoot: shoot", err, ds)
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
	notifyIfErr("answerParametricTransformer", err, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, errorMessage(commandErrorHappened))
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the Parametric Transformer in 7 days!")
	}
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := stopParametricReminder(ds, mc, ctx)
	notifyIfErr("answerParametricTransformerStop", err, ds)
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
You can only replace one character on each team with one of the replacements.

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

func answerRandomDomainRun(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	match := commandWithTwoArguments.FindStringSubmatch(mc.Content)
	if match == nil || len(match) != 3 {
		ds.ChannelMessageSend(mc.ChannelID, commandWithTwoArgumentsError)
		return
	}

	// we also remove the "(" and ")" chars
	set1 := match[1][1 : len(match[1])-1]
	set2 := match[2][1 : len(match[2])-1]
	art1 := genshinartis.RandomArtifactFromDomain(set1, set2)
	art2 := genshinartis.RandomArtifactFromDomain(set1, set2)
	msg := formatGenshinArtifact(art1)
	msg += formatGenshinArtifact(art2)
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

func answerRoll(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	diceSides, err := strconv.Atoi(commandBody)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "This command needs a numeric argument")
		return
	}
	if diceSides <= 0 {
		ds.ChannelMessageSend(mc.ChannelID, "Dice sides amount must be positive!")
		return
	}
	result := rand.Intn(diceSides) + 1
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("You rolled a %d!", result))
}

func answerRoleIDs(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	roles, err := ds.GuildRoles(mc.GuildID)
	notifyIfErr("answerRoleIDs", err, ds)
	if err != nil {
		return
	}
	var response string
	for _, role := range roles {
		if role.Name == roleEveryone {
			continue
		}
		response += fmt.Sprintf("%s: %s\n", role.Name, role.ID)
	}
	userMessageSend(adminID, response, ds)
}

// ---------- Simple command stuff ----------

func answerAddCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	commandBody := commandPrefixRegex.ReplaceAllString(mc.Content, "")
	key := strings.TrimSpace(commandPrefixRegex.FindString(commandBody))
	if key == "" {
		ds.ChannelMessageSend(mc.ChannelID, errorMessage("Could not get the key from the command body"))
		return
	}
	response := commandPrefixRegex.ReplaceAllString(commandBody, "")
	err := commandDS.addSimpleCommand(key, response)
	notifyIfErr("addSimpleCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
}

func answerRemoveCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	err := commandDS.removeSimpleCommand(commandBody)
	notifyIfErr("removeSimpleCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
}

func answerListCommands(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	keys, err := commandDS.allSimpleCommandKeys()
	notifyIfErr("removeSimpleCommand", err, ds)
	if len(keys) != 0 {
		ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Current commands: %v", keys))
	}
}

func answerAllowSpamming(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := commandDS.addSpammableChannel(mc.ChannelID)
	notifyIfErr("addSpammableChannel", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	}
}

func answerPreventSpamming(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := commandDS.removeSpammableChannel(mc.ChannelID)
	notifyIfErr("removeSpammableChannel", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	}
	notifyIfErr("MessageReactionAdd", err, ds)
}

// ---------- Server commands ----------

func answerReboot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	err := reboot()
	notifyIfErr("reboot", err, ds)
}

func answerShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	timeToWait, _ := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will shutdown in %v", timeToWait))
	err := shutdown(timeToWait)
	notifyIfErr("shutdown", err, ds)
}

func answerAbortShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	err := abortShutdown()
	notifyIfErr("abortShutdown", err, ds)
}

// ---------- Cooldowns ----------

var lastUserCommandTime = map[string]time.Time{}

const commandCooldown = time.Minute * 15

func resetUserCooldown(userID string) {
	lastUserCommandTime[userID] = time.Now()
}

func isUserOnCooldown(userID string) bool {
	lastTime, ok := lastUserCommandTime[userID]
	if !ok {
		return false
	}
	return time.Now().Before(lastTime.Add(commandCooldown))
}
