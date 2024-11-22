package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/ppgen"
)

const roleEveryone = "@everyone"
const globalGuildID = ""

var ogNonRootTwitterLinkRegex = regexp.MustCompile(`\b(?:https?://)?(?:www\.)?(?:twitter|x)\.com/\S+\b`)
var ogNonRootPixivLinkRegex = regexp.MustCompile(`\b(?:https?://)?(?:www\.)?pixiv\.net/\S+\b`)
var commandPrefixRegex = regexp.MustCompile(`^!\w+\s*`)
var commandWithTwoArguments = regexp.MustCompile(`^!\w+\s*(\(.{1,36}\))\s*(\(.{1,36}\))`)
var commandWithMention = regexp.MustCompile(`^!\w+\s*<@!?(\d+)>`)

var badEmbedLinkReplacements = map[*regexp.Regexp]string{
	regexp.MustCompile(`\b(?:https?://)?(?:www\.)?(?:twitter|x)\.com\b`): "https://fxtwitter.com",
	regexp.MustCompile(`\b(?:https?://)?(?:www\.)?pixiv\.net\b`):         "https://phixiv.net",
}

var messageLinkFixToOgAuthorId = map[*discordgo.Message]string{}

type command func(*discordgo.Session, *discordgo.MessageCreate, context.Context) bool

func onMessageCreated(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if mc.Author.ID == ds.State.User.ID {
			return
		}

		if len(mc.Content) == 0 {
			return
		}

		// Process commands
		if mc.Content[0] == '!' {
			trimmedMsg := strings.TrimSpace(mc.Content)
			processCommand(ds, mc, trimmedMsg, ctx)
			return
		}

		// Twitter links replacement
		if ogNonRootTwitterLinkRegex.MatchString(mc.Content) ||
			ogNonRootPixivLinkRegex.MatchString(mc.Content) {
			processMessageWithBadEmbedLinks(ds, mc, ctx)
			return
		}
	}
}

// the command key must be lowercased
var commands = map[string]command{
	// public
	"!version":                   simpleTextResponse("v3.7.3"),
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
	"!shoot":                     notSpammable(answerShoot),
	"!pp":                        notSpammable(answerPP),
	// hidden or easter eggs
	"!hello":        notSpammable(answerHello),
	"!liquid":       notSpammable(answerLiquid),
	"!don":          notSpammable(answerDon),
	"!sniper_shoot": notSpammable(answerSniperShoot),
	// only available for discord mods
	"!roleids":              guildOnly(modOnly(answerRoleIDs)),
	"!react4roles":          guildOnly(modOnly(answerMakeReact4RolesMsg)),
	"!addcommand":           guildOnly(modOnly(answerAddCommand)),
	"!removecommand":        guildOnly(modOnly(answerRemoveCommand)),
	"!deletecommand":        guildOnly(modOnly(answerRemoveCommand)),
	"!commandcreator":       guildOnly(modOnly(answerCommandCreator)),
	"!listcommands":         modOnly(answerListCommands),
	"!listservercommands":   guildOnly(modOnly(answerListGuildCommands)),
	"!listglobalcommands":   guildOnly(modOnly(answerListGlobalCommands)),
	"!allowspamming":        guildOnly(modOnly(answerAllowSpamming)),
	"!preventspamming":      guildOnly(modOnly(answerPreventSpamming)),
	"!setcustomtimeoutrole": guildOnly(modOnly(answerSetCustomTimeoutRole)),
	"!announcehere":         guildOnly(modOnly(answerAnnounceHere)),
	"!fixbadembedlinks":     guildOnly(modOnly(answerFixBadEmbedLinks)),
	"!messagelogs":          guildOnly(modOnly(answerMessageLogs)),
	"!commandstats":         guildOnly(modOnly(answerCommandStats)),
	// only available for the bot owner
	"!abort":               adminOnly(answerAbort),
	"!guildlist":           adminOnly(answerGuildList),
	"!addglobalcommand":    adminOnly(answerAddGlobalCommand),
	"!removeglobalcommand": adminOnly(answerRemoveGlobalCommand),
	"!deleteglobalcommand": adminOnly(answerRemoveGlobalCommand),
	"!announce":            adminOnly(answerAnnounce),
	"!dbbackup":            adminOnly(answerDbBackup),
	"!runtimestats":        adminOnly(answerRuntimeStats),
	"!reboot":              adminOnly(answerReboot),
	"!shutdown":            adminOnly(answerShutdown),
	"!abortshutdown":       adminOnly(answerAbortShutdown),
}

func processCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, fullCommand string, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			notifyIfErr("processCommand", fmt.Errorf("panic in command %s: %s\n%s", fullCommand, r, string(debug.Stack())), ds)
		}
	}()

	commandKey := strings.TrimSpace(commandPrefixRegex.FindString(fullCommand))
	lowercaseCommandKey := strings.ToLower(commandKey)
	command, ok := commands[lowercaseCommandKey]
	if ok {
		if command(ds, mc, ctx) {
			onSuccessCommandCall(mc, lowercaseCommandKey)
		}
		return
	}

	response, err := commandDS.simpleCommandResponse(commandKey, mc.GuildID)
	notifyIfErr("simpleCommandResponse", err, ds)
	if err == nil {
		if notSpammable(simpleTextResponse(response))(ds, mc, ctx) {
			onSuccessCommandCall(mc, commandKey)
		}
	}
}

func onSuccessCommandCall(mc *discordgo.MessageCreate, commandKey string) {
	log.Printf("[%s] [%s] %s", mc.ChannelID, mc.Author.Username, commandKey)
	if mc.GuildID != globalGuildID {
		commandDS.increaseCommandCountStat(mc.GuildID, commandKey)
	}
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
	notifyIfErr("answerPP: parsing user id: "+mc.Author.ID, err, ds)
	if err != nil {
		return false
	}
	seed *= unixDay()
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

func answerAnnounceHere(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := serverDS.setServerProperty(mc.GuildID, serverPropAnnounceHere, mc.ChannelID)
	notifyIfErr("answerAnnounceHere", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, "Okay! Will send announcements in this channel")
	}
	return err == nil
}

func answerFixBadEmbedLinks(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	currSetting, _ := serverDS.getServerProperty(mc.GuildID, serverPropFixBadEmbedLinks)
	newSetting := serverPropYes
	if currSetting == serverPropYes {
		newSetting = serverPropNo
	}
	err := serverDS.setServerProperty(mc.GuildID, serverPropFixBadEmbedLinks, newSetting)
	if err == nil && newSetting == serverPropYes {
		ds.ChannelMessageSend(mc.ChannelID, "Okay! Will fix bad embed links")
	} else if err == nil && newSetting == serverPropNo {
		ds.ChannelMessageSend(mc.ChannelID, "Okay! Will not fix bad embed links")
	}
	return err == nil
}

func processMessageWithBadEmbedLinks(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	currSetting, _ := serverDS.getServerProperty(mc.GuildID, serverPropFixBadEmbedLinks)
	if currSetting != serverPropYes {
		return
	}

	messageContent := mc.Content
	for rgx, rpl := range badEmbedLinkReplacements {
		messageContent = rgx.ReplaceAllString(messageContent, rpl)
	}
	fixedMsg, err := sendAsUser(ds, mc.Author, mc.ChannelID, messageContent, mc.ReferencedMessage)
	if err != nil {
		notifyIfErr("processMessageWithBadEmbedLinks::sendAsUser", err, ds)
		return
	}
	messageLinkFixToOgAuthorId[fixedMsg] = mc.Author.ID

	ds.State.MessageRemove(mc.Message)
	err = ds.ChannelMessageDelete(mc.ChannelID, mc.ID)
	if err != nil {
		notifyIfErr("processMessageWithBadEmbedLinks::ds.ChannelMessageDelete", err, ds)
		return
	}
}

func answerDeleteLinkFixMessage(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	interactionUserId := interactionUser(ic).ID
	ogAuthorId, ok := messageLinkFixToOgAuthorId[ic.Message]
	if !ok && !isAdmin(interactionUserId) {
		ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Could not find original author, only a mod can delete that message",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Check if the user can delete the message
	if interactionUserId != ogAuthorId || !isAdmin(interactionUserId) {
		ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You did not send that message",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Try to delete and respond if it was successful
	fixedMsgID := ic.ApplicationCommandData().TargetID
	err := ds.ChannelMessageDelete(ic.ChannelID, fixedMsgID)
	notifyIfErr("answerDeleteLinkFixMessage::ChannelMessageDelete", err, ds)
	if err != nil {
		ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Sorry, I could not delete the message u_u",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Message deleted ^w^",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func answerMessageLogs(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := serverDS.setServerProperty(mc.GuildID, serverPropMessageLogs, mc.ChannelID)
	notifyIfErr("answerMessageLogs", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, "Okay! Will send message logs in this channel")
	}
	return err == nil
}

func answerCommandStats(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	stats, err := commandDS.guildCommandStats(mc.GuildID)
	if err != nil {
		notifyIfErr("answerCommandStats: get command stats", err, ds)
		return false
	}
	statsMsg := ""
	for _, s := range stats {
		statsMsg += fmt.Sprintf("%s: %d\n", s.Command, s.Count)
	}
	_, err = ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
		Title:       "Command stats",
		Description: statsMsg,
	})
	return err == nil
}

func answerGuildList(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	guilds, err := ds.UserGuilds(100, "", "", true)
	if err != nil {
		notifyIfErr("answerGuildList", err, ds)
		return false
	}
	guildsMsg := ""
	for _, g := range guilds {
		guildsMsg += fmt.Sprintf("%s [%s] - Member count %d - Presence count %d\n", g.Name, g.ID, g.ApproximateMemberCount, g.ApproximatePresenceCount)
	}
	_, err = ds.ChannelMessageSend(mc.ChannelID, guildsMsg)
	return err == nil
}

// ---------- Simple command stuff ----------

func answerAddCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := commandPrefixRegex.ReplaceAllString(mc.Content, "")
	key := strings.TrimSpace(commandPrefixRegex.FindString(commandBody))
	if key == "" {
		ds.ChannelMessageSend(mc.ChannelID, markdownDiffBlock("Could not get the key from the command body", "- "))
		return false
	}
	response := commandPrefixRegex.ReplaceAllString(commandBody, "")
	if response == "" {
		ds.ChannelMessageSend(mc.ChannelID, markdownDiffBlock("Could not get the response from the command body", "- "))
		return false
	}
	err := commandDS.addSimpleCommand(key, response, mc.GuildID, mc.Author.ID)
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
		ds.ChannelMessageSend(mc.ChannelID, markdownDiffBlock("Could not get the key from the command body", "- "))
		return false
	}
	response := commandPrefixRegex.ReplaceAllString(commandBody, "")
	if response == "" {
		ds.ChannelMessageSend(mc.ChannelID, markdownDiffBlock("Could not get the response from the command body", "- "))
		return false
	}
	err := commandDS.addSimpleCommand(key, response, globalGuildID, mc.Author.ID)
	notifyIfErr("addGlobalCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerRemoveCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	err := commandDS.removeSimpleCommand(commandBody, mc.GuildID)
	if err == errZeroRowsAffected {
		ds.ChannelMessageSend(mc.ChannelID, "I could not find that command! sowwy u_u")
	}
	notifyIfErr("removeSimpleCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerCommandCreator(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	if commandBody == "" {
		return false
	}
	if commandBody[0] != '!' {
		commandBody = "!" + commandBody
	}

	creator, err := commandDS.getCommandCreator(commandBody, mc.GuildID)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find command creator. I'm sowwy u_u")
		return false
	}
	ds.ChannelMessageSendComplex(mc.ChannelID, &discordgo.MessageSend{
		Content:         fmt.Sprintf("Command creator: <@%s>", creator),
		AllowedMentions: &discordgo.MessageAllowedMentions{},
	})
	return true
}

func answerRemoveGlobalCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	err := commandDS.removeSimpleCommand(commandBody, globalGuildID)
	if err == errZeroRowsAffected {
		ds.ChannelMessageSend(mc.ChannelID, "I could not find that command! sowwy u_u")
	}
	notifyIfErr("removeGlobalCommand", err, ds)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return err == nil
}

func answerAnnounce(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	properties, err := serverDS.getServerProperties(serverPropAnnounceHere)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not get server properties: "+err.Error())
		return false
	}
	log.Println("Server properties", properties)

	errors := ""
	for _, prop := range properties {
		_, err = ds.ChannelMessageSend(prop.PropertyValue, commandBody)
		log.Println("Sending message to channel", prop.PropertyValue, "in server", prop.ServerID, "with content", commandBody)
		if err != nil {
			errors += fmt.Sprintf("Could not send message to channel %s in server %s: %s\n", prop.PropertyValue, prop.ServerID, err)
		}
	}

	if errors != "" {
		ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
			Title:       "Errors while announcing",
			Description: errors,
		})
	} else {
		ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	}
	return errors == ""
}

func answerRuntimeStats(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	msg := fmt.Sprintf("Number of CPUs: %d\n", runtime.NumCPU())
	msg += fmt.Sprintf("Number of goroutines: %d\n", runtime.NumGoroutine())
	msg += fmt.Sprintf("Current allocated memory: %.2f MBs\n", float64(mem.Alloc)/1_000_000)
	msg += fmt.Sprintf("Total allocated memory: %.2f MBs\n", float64(mem.TotalAlloc)/1_000_000)
	msg += fmt.Sprintf("System memory reserved: %.2f MBs\n", float64(mem.Sys)/1_000_000)
	msg += fmt.Sprintf("Number of memory allocations: %d\n", mem.Mallocs)
	msg += fmt.Sprintf("GC Pause Time (ms): %.2f\n", float64(mem.PauseTotalNs)/1_000_000)
	msg += fmt.Sprintf("GC Pause Count: %d\n", mem.NumGC)
	ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
		Title:       "Runtime stats",
		Description: msg,
	})
	return true
}

func answerListCommands(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	keys, err := commandDS.allSimpleCommandKeys(mc.GuildID, true)
	notifyIfErr("answerListCommands::allSimpleCommandKeys", err, ds)
	if len(keys) != 0 {
		sort.Strings(keys)
		msg := ""
		for _, k := range keys {
			msg += k + "\n"
		}
		ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
			Title:       "All simple commands available",
			Description: msg,
		})
	}
	return err == nil
}

func answerListGuildCommands(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	keys, err := commandDS.allSimpleCommandKeys(mc.GuildID, false)
	notifyIfErr("answerListCommands::allSimpleCommandKeys", err, ds)
	if len(keys) != 0 {
		sort.Strings(keys)
		msg := ""
		for _, k := range keys {
			msg += k + "\n"
		}
		ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
			Title:       "Simple commands created in this server",
			Description: msg,
		})
	}
	return err == nil
}

func answerListGlobalCommands(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	keys, err := commandDS.allSimpleCommandKeys("", false)
	notifyIfErr("answerListCommands::allSimpleCommandKeys", err, ds)
	if len(keys) != 0 {
		sort.Strings(keys)
		msg := ""
		for _, k := range keys {
			msg += k + "\n"
		}
		ds.ChannelMessageSendEmbed(mc.ChannelID, &discordgo.MessageEmbed{
			Title:       "All global commands available",
			Description: msg,
		})
	}
	return err == nil
}

// ---------- Server commands ----------

func answerAbort(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := ds.Close()
	notifyIfErr("abort", err, ds)
	abortChannel <- os.Interrupt
	return err == nil
}

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

func guildOnly(wrapped command) command {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
		if mc.GuildID == globalGuildID {
			ds.ChannelMessageSend(mc.ChannelID, notAGuildMessage)
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
