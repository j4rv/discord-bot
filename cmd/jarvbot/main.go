package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
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

func initDB() {
	db := sqlx.MustOpen("sqlite3", dbFilename)
	if err := db.Ping(); err != nil {
		panic("DB did not answer ping: " + err.Error())
	}
	createTables(db)
	genshinDS = genshinDataStore{db}
	commandDS = commandDataStore{db}
	moddingDS = moddingDataStore{db}
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

	r4rsCRON := cron.New()
	_, err = r4rsCRON.AddFunc(react4RolesCRON, react4RolesCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to react4RolesCRON", err, ds)
	} else {
		r4rsCRON.Start()
	}
}

// initSlashCommands returns a function to remove the registered slash commands for graceful shutdowns
func initSlashCommands(ds *discordgo.Session) func() {
	ds.AddHandler(func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
		if h, ok := slashHandlers[ic.ApplicationCommandData().Name]; ok {
			h(ds, ic)
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

// for single line strings only!
func errorMessage(body string) string {
	return "```diff\n- " + body + "\n```"
}

func notifyIfErr(context string, err error, ds *discordgo.Session) {
	if err != nil {
		msg := "ERROR [" + context + "]: " + err.Error()
		log.Println(msg)
		sendDirectMessage(adminID, errorMessage(msg), ds)
	}
}
