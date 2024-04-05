package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/ppgen"
)

const roleEveryone = "@everyone"
const globalGuildID = ""

var commandPrefixRegex = regexp.MustCompile(`^!\w+\s*`)
var commandWithTwoArguments = regexp.MustCompile(`^!\w+\s*(\(.{1,36}\))\s*(\(.{1,36}\))`)
var commandWithMention = regexp.MustCompile(`^!\w+\s*<@!?(\d+)>`)

type command func(*discordgo.Session, *discordgo.MessageCreate, context.Context) bool

func onMessageCreated(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if mc.Author.ID == ds.State.User.ID {
			return
		}

		// Ignore all messages that don't start with '!'
		if len(mc.Content) == 0 || mc.Content[0] != '!' {
			return
		}

		trimmedMsg := strings.TrimSpace(mc.Content)
		processCommand(ds, mc, trimmedMsg, ctx)
	}
}

// the command key must be lowercased
var commands = map[string]command{
	// public
	"!version":                   simpleTextResponse("v3.0.3"),
	"!source":                    simpleTextResponse("Source code: https://github.com/j4rv/discord-bot"),
	"!genshindailycheckin":       answerGenshinDailyCheckIn,
	"!genshindailycheckinstop":   answerGenshinDailyCheckInStop,
	"!parametrictransformer":     answerParametricTransformer,
	"!parametrictransformerstop": answerParametricTransformerStop,
	"!playstore":                 answerPlayStore,
	"!playstorestop":             answerPlayStoreStop,
	"!randomabysslineup":         notSpammable(answerRandomAbyssLineup),
	"!randomartifact":            notSpammable(answerRandomArtifact),
	"!randomartifactset":         notSpammable(answerRandomArtifactSet),
	"!randomdomainrun":           notSpammable(answerRandomDomainRun),
	"!remindme":                  notSpammable(answerRemindme),
	"!roll":                      notSpammable(answerRoll),
	// hidden or easter eggs
	"!hello":  notSpammable(answerHello),
	"!liquid": notSpammable(answerLiquid),
	"!don":    notSpammable(answerDon),
	"!shoot":  notSpammable(answerShoot),
	"!pp":     notSpammable(answerPP),
	// only available for discord mods
	"!roleids":              modOnly(answerRoleIDs),
	"!react4roles":          modOnly(answerMakeReact4RolesMsg),
	"!addcommand":           modOnly(answerAddCommand),
	"!removecommand":        modOnly(answerRemoveCommand),
	"!allowspamming":        modOnly(answerAllowSpamming),
	"!preventspamming":      modOnly(answerPreventSpamming),
	"!setcustomtimeoutrole": modOnly(answerSetCustomTimeoutRole),
	// only available for the bot owner
	"!addglobalcommand":    adminOnly(answerAddGlobalCommand),
	"!removeglobalcommand": adminOnly(answerRemoveGlobalCommand),
	"!listcommands":        adminOnly(answerListCommands),
	"!reboot":              adminOnly(answerReboot),
	"!shutdown":            adminOnly(answerShutdown),
	"!abortshutdown":       adminOnly(answerAbortShutdown),
}

func processCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, fullCommand string, ctx context.Context) {
	commandKey := strings.TrimSpace(commandPrefixRegex.FindString(fullCommand))
	command, ok := commands[strings.ToLower(commandKey)]
	if ok {
		if command(ds, mc, ctx) {
			onSuccessCommandCall(mc)
		}
		return
	}

	response, err := commandDS.simpleCommandResponse(commandKey, mc.GuildID)
	notifyIfErr("simpleCommandResponse", err, ds)
	if err == nil {
		if notSpammable(simpleTextResponse(response))(ds, mc, ctx) {
			onSuccessCommandCall(mc)
		}
	}
}

func onSuccessCommandCall(mc *discordgo.MessageCreate) {
	log.Printf("[%s] [%s] %s", mc.ChannelID, mc.Author.Username, mc.Content)
	channelIsSpammable, _ := commandDS.isChannelSpammable(mc.ChannelID)
	if !channelIsSpammable {
		resetUserCooldown(mc.Author.ID)
	}
}

// Command Answers

func simpleTextResponse(body string) func(*discordgo.Session, *discordgo.MessageCreate, context.Context) bool {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
		_, err := ds.ChannelMessageSend(mc.ChannelID, body)
		return err == nil
	}
}

func answerHello(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	var err error
	if mc.Author.ID == adminID {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Hewwo master uwu")
	} else {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Hello!")
	}
	return err == nil
}

func answerPP(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	seed, err := strconv.ParseInt(mc.Author.ID, 10, 64)
	seed *= unixDay()
	notifyIfErr("answerPP: parsing user id: "+mc.Author.ID, err, ds)
	if err != nil {
		return false
	}
	pp := ppgen.NewPenisWithSeed(seed)
	_, err = ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("%s's penis: %s", mc.Author.Mention(), pp))
	return err == nil
}

// FIXME: Limit its usage by user (max 3 active reminders?)
func answerRemindme(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeToWait, reminderBody := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will remind you in %v with the message '%s'", timeToWait, reminderBody))
	time.Sleep(timeToWait)
	sendDirectMessage(mc.Author.ID, reminderBody, ds)
	return true
}

func answerRoll(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	diceSides, err := strconv.Atoi(commandBody)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "This command needs a numeric argument")
		return false
	}
	if diceSides <= 0 {
		ds.ChannelMessageSend(mc.ChannelID, "Dice sides amount must be positive!")
		return false
	}
	result := rand.Intn(diceSides) + 1
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("You rolled a %d!", result))
	return true
}

func answerAllowSpamming(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := commandDS.addSpammableChannel(mc.ChannelID)
	notifyIfErr("addSpammableChannel", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	}
	return err == nil
}

func answerPreventSpamming(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := commandDS.removeSpammableChannel(mc.ChannelID)
	notifyIfErr("removeSpammableChannel", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	}
	notifyIfErr("MessageReactionAdd", err, ds)
	return err == nil
}

func answerSetCustomTimeoutRole(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	guildID := mc.GuildID

	timeoutRoleName := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	_, err := guildRoleByName(ds, guildID, timeoutRoleName)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Could not find role '%s'", timeoutRoleName))
		return false
	}

	err = setCustomTimeoutRole(ds, guildID, timeoutRoleName)
	notifyIfErr("setCustomTimeoutRole", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Custom timeout role set to '%s'", timeoutRoleName))
	}
	return err == nil
}

// ---------- Simple command stuff ----------

func answerAddCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := commandPrefixRegex.ReplaceAllString(mc.Content, "")
	key := strings.TrimSpace(commandPrefixRegex.FindString(commandBody))
	if key == "" {
		ds.ChannelMessageSend(mc.ChannelID, errorMessage("Could not get the key from the command body"))
		return false
	}
	response := commandPrefixRegex.ReplaceAllString(commandBody, "")
	err := commandDS.addSimpleCommand(key, response, mc.GuildID)
	notifyIfErr("addSimpleCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerAddGlobalCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := commandPrefixRegex.ReplaceAllString(mc.Content, "")
	key := strings.TrimSpace(commandPrefixRegex.FindString(commandBody))
	if key == "" {
		ds.ChannelMessageSend(mc.ChannelID, errorMessage("Could not get the key from the command body"))
		return false
	}
	response := commandPrefixRegex.ReplaceAllString(commandBody, "")
	err := commandDS.addSimpleCommand(key, response, globalGuildID)
	notifyIfErr("addGlobalCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerRemoveCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	err := commandDS.removeSimpleCommand(commandBody, mc.GuildID)
	notifyIfErr("removeSimpleCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerRemoveGlobalCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	err := commandDS.removeSimpleCommand(commandBody, globalGuildID)
	notifyIfErr("removeGlobalCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerListCommands(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	keys, err := commandDS.allSimpleCommandKeys(mc.GuildID)
	notifyIfErr("answerListCommands::allSimpleCommandKeys", err, ds)
	if len(keys) != 0 {
		sort.Strings(keys)
		msg := ""
		for _, k := range keys {
			msg += k + "\n"
		}
		ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
			Title:       "Simple commands",
			Description: msg,
		})
	}
	return err == nil
}

// ---------- Server commands ----------

func answerReboot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := reboot()
	notifyIfErr("reboot", err, ds)
	return err == nil
}

func answerShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeToWait, _ := processTimedCommand(mc.Content)
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("Gotcha! will shutdown in %v", timeToWait))
	err := shutdown(timeToWait)
	notifyIfErr("shutdown", err, ds)
	return err == nil
}

func answerAbortShutdown(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	err := abortShutdown()
	notifyIfErr("abortShutdown", err, ds)
	return err == nil
}

// Command wrappers

func adminOnly(wrapped command) command {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
		if !isAdmin(mc.Author.ID) {
			ds.ChannelMessageSend(mc.ChannelID, userMustBeAdminMessage)
			return false
		}
		return wrapped(ds, mc, ctx)
	}
}

func modOnly(wrapped command) command {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
		if !(isAdmin(mc.Author.ID) || isMod(ds, mc.Author.ID, mc.ChannelID)) {
			ds.ChannelMessageSend(mc.ChannelID, userMustBeModMessage)
			return false
		}
		return wrapped(ds, mc, ctx)
	}
}

func notSpammable(wrapped command) command {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
		if !isAdmin(mc.Author.ID) {
			channelIsSpammable, err := commandDS.isChannelSpammable(mc.ChannelID)
			notifyIfErr("notSpammable::isChannelSpammable", err, ds)
			if !channelIsSpammable && isUserOnCooldown(mc.Author.ID) {
				sendDirectMessage(mc.Author.ID, "Pls don't spam the bot commands uwu", ds)
				ds.MessageReactionAdd(mc.ChannelID, mc.Message.ID, "❌")
				return false
			}
		}
		return wrapped(ds, mc, ctx)
	}
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
