package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

// FLAGS
var token string
var adminID string
var noSlashCommands bool

const discordMaxMessageLength = 2000

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	initFlags()
	initDB()
	ds := initDiscordSession()
	initCRONs(ds)

	if !noSlashCommands {
		removeSlashCommands := initSlashCommands(ds)
		defer removeSlashCommands()
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL-C to exit.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChan

	ds.Close()
}

func initFlags() {
	flag.StringVar(&token, "token", "", "Bot Token")
	flag.StringVar(&adminID, "adminID", "", "The ID of the bot's admin")
	flag.BoolVar(&noSlashCommands, "noSlashCommands", false, "The bot will not init slash commands, boots faster.")
	flag.Parse()
	if token == "" {
		panic("Provide a token flag!")
	}
	if adminID == "" {
		log.Println("Warning: Admin user ID not set")
	}
}

func initDiscordSession() *discordgo.Session {
	log.Println("Initiating Discord Session")
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		panic("error creating Discord session: " + err.Error())
	}

	backgroundCtx := context.Background()

	ds.AddHandler(onMessageCreated(backgroundCtx))
	ds.AddHandler(onMessageReacted(backgroundCtx))
	ds.AddHandler(onMessageUnreacted(backgroundCtx))

	ds.Identify.Intents |= discordgo.IntentGuilds
	ds.Identify.Intents |= discordgo.IntentGuildMembers
	ds.Identify.Intents |= discordgo.IntentGuildMessages
	ds.Identify.Intents |= discordgo.IntentGuildMessageReactions
	ds.Identify.Intents |= discordgo.IntentDirectMessages

	// Open a websocket connection to Discord and begin listening.
	err = ds.Open()
	if err != nil {
		panic("error opening connection: " + err.Error())
	}

	return ds
}

func initCRONs(ds *discordgo.Session) {
	// TODO: CRON that checks if a React4Role message still exists, if it doesnt, remove it from DB (once a week for example)
	log.Println("Initiating CRONs")
	dailyCheckInCRON := cron.New()
	_, err := dailyCheckInCRON.AddFunc(dailyCheckInReminderCRON, dailyCheckInCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to dailyCheckInCRON", err, ds)
	} else {
		dailyCheckInCRON.Start()
	}

	parametricCRON := cron.New()
	_, err = parametricCRON.AddFunc(parametricReminderCRON, parametricCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to parametricCRON", err, ds)
	} else {
		parametricCRON.Start()
	}

	playStoreCRON := cron.New()
	_, err = playStoreCRON.AddFunc(playStoreReminderCRON, playStoreCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to playStoreCRON", err, ds)
	} else {
		playStoreCRON.Start()
	}
}

// React to every new message
func onMessageCreated(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	return func(ds *discordgo.Session, mc *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if mc.Author.ID == ds.State.User.ID {
			return
		}

		message := strings.TrimSpace(mc.Content)

		// Ignore all messages that don't start with '!'
		if len(message) == 0 || message[0] != '!' {
			return
		}

		processCommand(ds, mc, message, ctx)
	}
}

var userChannels = map[string]*discordgo.Channel{}

func getUserChannel(userID string, ds *discordgo.Session) (*discordgo.Channel, error) {
	userChannel, ok := userChannels[userID]
	if !ok {
		createdChannel, err := ds.UserChannelCreate(userID)
		if err != nil {
			// If an error occurred, we failed to create the channel.
			//
			// Some common causes are:
			// 1. We don't share a server with the user (not possible here).
			// 2. We opened enough DM channels quickly enough for Discord to
			//    label us as abusing the endpoint, blocking us from opening
			//    new ones.
			log.Println("error creating user channel:", err)
			return nil, err
		}
		userChannels[userID] = createdChannel
		return createdChannel, nil
	}
	return userChannel, nil
}

func userMessageSend(userID string, body string, ds *discordgo.Session) (*discordgo.Message, error) {
	userChannel, err := getUserChannel(userID, ds)
	if err != nil {
		return nil, err
	}
	return ds.ChannelMessageSend(userChannel.ID, body)
}

func guildRoleByName(ds *discordgo.Session, guildID string, roleName string) (*discordgo.Role, error) {
	roles, err := ds.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.Name == roleName {
			return r, nil
		}
	}
	return nil, fmt.Errorf("role with name %s not found in guild with id %s", roleName, guildID)
}

func isMemberInRole(member *discordgo.Member, roleID string) bool {
	for _, r := range member.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}

// for single line strings only!
func errorMessage(body string) string {
	return "```diff\n- " + body + "\n```"
}

func notifyIfErr(context string, err error, ds *discordgo.Session) {
	if err != nil {
		msg := "ERROR [" + context + "]: " + err.Error()
		log.Println(msg)
		userMessageSend(adminID, errorMessage(msg), ds)
	}
}

func interactionUser(ic *discordgo.InteractionCreate) *discordgo.User {
	if ic.User != nil {
		return ic.User
	}
	if ic.Member != nil {
		return ic.Member.User
	}
	return nil
}
