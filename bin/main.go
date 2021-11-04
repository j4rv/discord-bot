package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var token string
var adminID string

func main() {
	initFlags()
	initDB()
	ds := initDiscordSession()
	initGenshinServices(ds)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChan
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
		fmt.Println("Warning: Admin user ID not set")
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

		msgCommand := strings.TrimSpace(commandPrefixRegex.FindString(message))
		for key, command := range commands {
			if key == msgCommand {
				command(ds, mc, ctx)
				break
			}
		}
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
			fmt.Println("error creating channel:", err)
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

// for single line errors only!
func errorMessageSend(body string, ds *discordgo.Session, mc *discordgo.MessageCreate) {
	ds.ChannelMessageSend(mc.ChannelID, "```diff\n- "+body+"\n```")
}
