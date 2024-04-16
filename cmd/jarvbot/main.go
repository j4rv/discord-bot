package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

func main() {
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
			log.Println("ERROR couldnt add handler for slash command:", ic.ApplicationCommandData().Name)
		}
	})

	registeredCommands := make([]*discordgo.ApplicationCommand, len(slashCommands))
	for i, slashCommand := range slashCommands {
		log.Println("Registering command:", slashCommand.Name)
		cmd, err := ds.ApplicationCommandCreate(ds.State.User.ID, "", slashCommand)
		if err != nil {
			notifyIfErr("Creating command: "+slashCommand.Name, err, ds)
			log.Printf("Cannot create '%v' command: %v", slashCommand.Name, err)
		}
		registeredCommands[i] = cmd
	}

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
	}
}

func cleanStateMessagesInChannel(ds *discordgo.Session, channel *discordgo.Channel) {
	for _, msg := range channel.Messages {
		if msg.Timestamp.Before(time.Now().Add(-stateMessageMaxLifetime)) {
			ds.State.MessageRemove(msg)
		}
	}
}

func sendAsUserWebhook(ds *discordgo.Session, channelID string) (*discordgo.Webhook, error) {
	hooks, err := ds.ChannelWebhooks(channelID)
	if err != nil {
		return nil, err
	}

	if len(hooks) == 0 {
		return ds.WebhookCreate(channelID, "SendAsUser", ds.State.User.AvatarURL(""))
	}

	return hooks[0], nil
}

func sendAsUser(ds *discordgo.Session, user *discordgo.User, channelID string, content string) (*discordgo.Message, error) {
	if user == nil || ds == nil || channelID == "" || content == "" {
		return nil, nil
	}

	webhook, err := sendAsUserWebhook(ds, channelID)
	if err != nil {
		return nil, err
	}

	return ds.WebhookExecute(webhook.ID, webhook.Token, false, &discordgo.WebhookParams{
		Content:   content,
		Username:  user.GlobalName,
		AvatarURL: user.AvatarURL(""),
	})
}

func diff(body, prefix string) string {
	lines := strings.Split(body, "\n")
	var formattedBody string
	for _, line := range lines {
		formattedBody += prefix + line + "\n"
	}
	return "```diff\n" + formattedBody + "```"
}

func notifyIfErr(context string, err error, ds *discordgo.Session) {
	if err != nil {
		msg := "ERROR [" + context + "]: " + err.Error()
		log.Println(msg)
		sendDirectMessage(adminID, diff(msg, "- "), ds)
	}
}
