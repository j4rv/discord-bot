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
)

var token string
var adminID string

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	initFlags()
	initDB()
	ds := initDiscordSession()
	initGenshinCRONs(ds)
	removeSlashCommands := initSlashCommands(ds)

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL-C to exit.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChan

	removeSlashCommands()
	ds.Close()
}

func initFlags() {
	flag.StringVar(&token, "token", "", "Bot Token")
	flag.StringVar(&adminID, "adminID", "", "The ID of the bot's admin")
	flag.Parse()
	if token == "" {
		panic("Provide a token flag!")
	}
	if adminID == "" {
		log.Println("Warning: Admin user ID not set")
	}
}

func initDiscordSession() *discordgo.Session {
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		panic("error creating Discord session: " + err.Error())
	}

	backgroundCtx := context.Background()

	// Register the messageCreate func as a callback for MessageCreate events.
	ds.AddHandler(onMessageCreated(backgroundCtx))
	ds.Identify.Intents |= discordgo.IntentsGuildMessages
	ds.Identify.Intents |= discordgo.IntentsDirectMessages

	// Open a websocket connection to Discord and begin listening.
	err = ds.Open()
	if err != nil {
		panic("error opening connection: " + err.Error())
	}

	return ds
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

func isMemberInRole(member *discordgo.Member, roleID string) (bool, error) {
	for _, r := range member.Roles {
		if r == roleID {
			return true, nil
		}
	}
	return false, nil
}

// for single line strings only!
func errorMessage(body string) string {
	return "```diff\n- " + body + "\n```"
}

func notifyIfErr(context string, err error, ds *discordgo.Session) {
	if err != nil {
		msg := "[" + context + "] an error happened: " + err.Error()
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
