package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
)

// FLAGS
var token string
var adminID string
var noSlashCommands bool

const discordMaxMessageLength = 2000

var abortChannel chan os.Signal

func main() {
	initFlags()
	initDB()
	ds := initDiscordSession()
	initCRONs(ds)

	var removeSlashCommands func()
	if !noSlashCommands {
		removeSlashCommands = initSlashCommands(ds)
	}

	// Wait here until CTRL-C or other term signal is received.
	abortChannel = make(chan os.Signal, 1)
	signal.Notify(abortChannel, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-abortChannel

	if !noSlashCommands {
		removeSlashCommands()
	}
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

func initDB() {
	db := sqlx.MustOpen("sqlite3", dbFilename)
	if err := db.Ping(); err != nil {
		panic("DB did not answer ping: " + err.Error())
	}
	createTables(db)
	genshinDS = genshinDataStore{db}
	commandDS = commandDataStore{db}
	moddingDS = moddingDataStore{db}
	serverDS = serverDataStore{db}
}

func initDiscordSession() *discordgo.Session {
	log.Println("Initiating Discord Session")
	ds, err := discordgo.New("Bot " + token)
	if err != nil {
		panic("error creating Discord session: " + err.Error())
	}

	backgroundCtx := context.Background()

	//ds.AddHandler(onGuildJoin(backgroundCtx))
	ds.AddHandler(onMessageCreated(backgroundCtx))
	ds.AddHandler(onMessageUpdated(backgroundCtx))
	ds.AddHandler(onMessageDeleted(backgroundCtx))
	ds.AddHandler(onMessageReacted(backgroundCtx))
	ds.AddHandler(onMessageUnreacted(backgroundCtx))

	ds.Identify.Intents |= discordgo.IntentGuilds
	ds.Identify.Intents |= discordgo.IntentGuildMembers
	ds.Identify.Intents |= discordgo.IntentGuildMessages
	ds.Identify.Intents |= discordgo.IntentGuildMessageReactions
	ds.Identify.Intents |= discordgo.IntentDirectMessages
	ds.Identify.Intents |= discordgo.IntentGuildWebhooks
	ds.State.MaxMessageCount = maxMessageCount

	// Open a websocket connection to Discord and begin listening.
	err = ds.Open()
	if err != nil {
		panic("error opening connection: " + err.Error())
	}

	return ds
}

func initCRONs(ds *discordgo.Session) {
	log.Println("Initiating CRONs")

	initCron := func(name string, cronSpec string, f func()) {
		cronJob := cron.New()
		_, err := cronJob.AddFunc(cronSpec, f)
		if err != nil {
			notifyIfErr("AddFunc to "+name, err, ds)
		} else {
			cronJob.Start()
		}
	}

	initCron("backupCRON", backupCRON, backupCRONFunc(ds))
	initCron("dailyCheckInCRON", dailyCheckInReminderCRON, dailyCheckInCRONFunc(ds))
	initCron("cleanStateMessagesCRON", cleanStateMessagesCRON, cleanStateMessagesCRONFunc(ds))
	initCron("parametricCRON", parametricReminderCRON, parametricCRONFunc(ds))
	initCron("playStoreCRON", playStoreReminderCRON, playStoreCRONFunc(ds))
	initCron("react4RolesCRON", react4RolesCRON, react4RolesCRONFunc(ds))
}

// initSlashCommands returns a function to remove the registered slash commands for graceful shutdowns
func initSlashCommands(ds *discordgo.Session) func() {
	ds.AddHandler(func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
		if h, ok := slashHandlers[ic.ApplicationCommandData().Name]; ok {
			h(ds, ic)
		} else {
			notifyIfErr("Slash command not found:"+ic.ApplicationCommandData().Name, nil, ds)
		}
	})

	registeredCommands := make([]*discordgo.ApplicationCommand, len(slashCommands))
	for i, slashCommand := range slashCommands {
		log.Println("Registering command:", slashCommand.Name)
		cmd, err := ds.ApplicationCommandCreate(ds.State.User.ID, "", slashCommand)
		if err != nil {
			notifyIfErr("Creating command: "+slashCommand.Name, err, ds)
			log.Printf("Cannot create '%v' command: %v", slashCommand.Name, err)
			continue
		}
		registeredCommands[i] = cmd
	}
	log.Println("Finished registering slash commands")

	return func() {
		log.Println("Removing registered slash commands...")
		for _, v := range registeredCommands {
			err := ds.ApplicationCommandDelete(ds.State.User.ID, "", v.ID)
			if err != nil {
				notifyIfErr("Deleting command: "+v.Name, err, ds)
				log.Printf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}

func cleanStateMessagesCRONFunc(ds *discordgo.Session) func() {
	return func() {
		log.Println("Doing state message cleanup")
		for _, pc := range ds.State.PrivateChannels {
			cleanStateMessagesInChannel(ds, pc)
		}
		for _, gc := range ds.State.Guilds {
			for _, gc := range gc.Channels {
				cleanStateMessagesInChannel(ds, gc)
			}
		}
		for m := range messageLinkFixToOgAuthorId {
			if messagePastLifetime(m) {
				delete(messageLinkFixToOgAuthorId, m)
			}
		}
	}
}

func cleanStateMessagesInChannel(ds *discordgo.Session, channel *discordgo.Channel) {
	for _, msg := range channel.Messages {
		if messagePastLifetime(msg) {
			ds.State.MessageRemove(msg)
		}
	}
}

func messagePastLifetime(msg *discordgo.Message) bool {
	if msg == nil {
		return true
	}
	return msg.Timestamp.Before(time.Now().Add(-stateMessageMaxLifetime))
}

func sendAsUserWebhook(ds *discordgo.Session, channelID string) (*discordgo.Webhook, error) {
	hooks, _ := ds.ChannelWebhooks(channelID)

	for _, hook := range hooks {
		if hook.Name == "SendAsUser" {
			return hook, nil
		}
	}

	return ds.WebhookCreate(channelID, "SendAsUser", ds.State.User.AvatarURL(""))
}

func fileMessageSend(ds *discordgo.Session, channelId, messageContent, fileName, fileData string) (*discordgo.Message, error) {
	return ds.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Content: messageContent,
		Files: []*discordgo.File{
			{
				ContentType: "text/plain",
				Name:        fileName,
				Reader:      strings.NewReader(fileData),
			},
		},
	})
}

// sendAsUser sends a message as the given user
// It will either use a Webhook when possible (to keep the sender's username and avatar)
// Or it will send a normal message with a mention of the original user
func sendAsUser(ds *discordgo.Session, user *discordgo.User, channelID string, content string, referencedMessage *discordgo.Message) (*discordgo.Message, error) {
	if user == nil || ds == nil || channelID == "" || content == "" {
		return nil, fmt.Errorf("user, ds, channelID, or content is nil")
	}

	// Webhook doesn't allow "replies" :(
	if referencedMessage != nil {
		return ds.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content:   fmt.Sprintf("%s:\n%s", user.Mention(), content),
			Reference: referencedMessage.Reference(),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				RepliedUser: true,
			},
		})
	}

	webhook, err := sendAsUserWebhook(ds, channelID)
	if err != nil {
		// webhook didn't work, let's try sending a normal silent message with a mention
		return ds.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content:         fmt.Sprintf("%s:\n%s", user.Mention(), content),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		})
	}

	return ds.WebhookExecute(webhook.ID, webhook.Token, true, &discordgo.WebhookParams{
		Content:   content,
		Username:  user.GlobalName,
		AvatarURL: user.AvatarURL(""),
	})
}

func markdownDiffBlock(body, prefix string) string {
	lines := strings.Split(body, "\n")
	var formattedBody string
	if prefix == "" {
		formattedBody = body
	} else {
		for _, line := range lines {
			formattedBody += prefix + line + "\n"
		}
	}
	return "```diff\n" + formattedBody + "```"
}

func notifyIfErr(context string, err error, ds *discordgo.Session) {
	if err != nil {
		msg := "ERROR [" + context + "]: " + err.Error()
		log.Println(msg)
		sendDirectMessage(adminID, markdownDiffBlock(msg, "- "), ds)
	}
}
