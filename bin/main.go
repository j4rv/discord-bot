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

func main() {
	initFlags()
	startBot()
}

func initFlags() {
	flag.StringVar(&token, "token", "", "Bot Token")
	flag.Parse()
}

func startBot() {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	backgroundCtx := context.Background()

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(onMessageCreated(backgroundCtx))

	dg.Identify.Intents |= discordgo.IntentsGuildMessages
	dg.Identify.Intents |= discordgo.IntentsDirectMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signalChan
	dg.Close()
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

		for key, command := range commands {
			if strings.HasPrefix(message, key) {
				command(ds, mc, ctx)
				break
			}
		}
	}
}

var userChannels = map[string]*discordgo.Channel{}

func getUserChannel(userID string, ds *discordgo.Session, mc *discordgo.MessageCreate) (*discordgo.Channel, error) {
	userChannel, ok := userChannels[userID]
	if !ok {
		createdChannel, err := ds.UserChannelCreate(mc.Author.ID)
		if err != nil {
			// If an error occurred, we failed to create the channel.
			//
			// Some common causes are:
			// 1. We don't share a server with the user (not possible here).
			// 2. We opened enough DM channels quickly enough for Discord to
			//    label us as abusing the endpoint, blocking us from opening
			//    new ones.
			fmt.Println("error creating channel:", err)
			ds.ChannelMessageSend(mc.ChannelID, "Something went wrong while sending the DM!")
			return nil, err
		}
		userChannels[userID] = createdChannel
		return createdChannel, nil
	}
	return userChannel, nil
}

func userMessageSend(userID string, body string, ds *discordgo.Session, mc *discordgo.MessageCreate) (*discordgo.Message, error) {
	userChannel, err := getUserChannel(mc.Author.ID, ds, mc)
	if err != nil {
		return nil, err
	}

	message, err := ds.ChannelMessageSend(userChannel.ID, body)
	if err != nil {
		fmt.Println("error sending DM message:", err)
		ds.ChannelMessageSend(mc.ChannelID, "Failed to send you a DM. Did you disable DM in your privacy settings? "+mc.Author.Mention())
	}

	return message, err
}
