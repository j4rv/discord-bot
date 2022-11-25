package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/lib/eightball"
	artis "github.com/j4rv/genshinartis"
)

var strongboxMinAmount = 1.0
var strongboxMaxAmount = 10.0

const avatarTargetSize = "1024"

var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "help",
		Description: "Help",
	},
	{
		Name:        "8ball",
		Description: "Ask the all-knowing 8 Ball",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "Your question to the 8 Ball",
				Required:    true,
			},
		},
	},
	{
		Name:        "avatar",
		Description: "Check the full-sized avatar of a user",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The discord user",
				Required:    true,
			},
		},
	},
	{
		Name:        "strongbox",
		Description: "Do Strongbox rolls of your set of choice",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "set",
				Description: "The artifact set name",
				Required:    true,
				Choices:     genshinSetChoices(),
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "amount",
				Description: "The amount of artifacts that will be generated",
				Required:    true,
				MinValue:    &strongboxMinAmount,
				MaxValue:    strongboxMaxAmount,
			},
		},
	},
}

var slashHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"help":      answerHelp,
	"8ball":     answer8ball,
	"avatar":    answerAvatar,
	"strongbox": answerStrongbox,
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
			log.Panicf("Cannot create '%v' command: %v", slashCommand.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return func() {
		log.Println("Removing registered slash commands...")
		for _, v := range registeredCommands {
			err := ds.ApplicationCommandDelete(ds.State.User.ID, "", v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}

func textRespond(ds *discordgo.Session, ic *discordgo.InteractionCreate, textResponse string) {
	err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: textResponse,
		},
	})
	notifyIfErr("textRespond", err, ds)
}

// -----------------------------------------------------

func answer8ball(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	question := ic.ApplicationCommandData().Options[0].StringValue()
	response := fmt.Sprintf("%s asked: %s\nThe 8 Ball says...\n'%s'",
		interactionUser(ic).Mention(), question, eightball.Response())
	textRespond(ds, ic, response)
}

func answerAvatar(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	user := ic.ApplicationCommandData().Options[0].UserValue(ds)
	textRespond(ds, ic, user.AvatarURL(avatarTargetSize))
}

func answerHelp(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	if interactionUser(ic).ID != adminID {
		textRespond(ds, ic, helpResponse)
	} else {
		textRespond(ds, ic, helpResponseAdmin)
	}
}

const helpResponse = `Available commands:
- **!source**: Links to the bot's source code
- **!ayayaify [message]**: Ayayaifies your message
- **!remindme [99h 99m 99s] [message]**: Reminds you of the message after the specified time has passed (beta)
- **!roll [99]**: Rolls a dice with the specified sides amount
- **!genshinDailyCheckIn**: Will remind you to do the Genshin Daily Check-In
- **!genshinDailyCheckInStop**: The bot will stop reminding you to do the Genshin Daily Check-In
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days. Use it again to reset the reminder
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!randomAbyssLineup**: The bot will give you two random teams and some replacements. Have fun ¯\_(ツ)_/¯. Optional: Write 8+ character names separated by commas and the bot will only choose from those
- **!randomArtifact**: Generates a random Lv20 Genshin Impact artifact
- **!randomArtifactSet**: Generates five random Lv20 Genshin Impact artifacts
- **!randomDomainRun (set A) (set B)**: Generates two random Lv20 Genshin Impact artifacts from the input sets
- **!randomStrongbox (set)**: Generates three random artifacts from the input set
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

func genshinSetChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(artis.AllArtifactSets))
	for i, set := range artis.AllArtifactSets {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  string(set),
			Value: set,
		}
	}
	return choices
}

func answerStrongbox(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	set := ic.ApplicationCommandData().Options[0].StringValue()
	amount := int(ic.ApplicationCommandData().Options[1].IntValue())

	response := fmt.Sprintf("%s is Strongboxing %d %s artifacts:\n", interactionUser(ic).Mention(), amount, set)

	for i := 0; i < amount; i++ {
		art := artis.RandomArtifactOfSet(set, artis.StrongboxBase4Chance)
		response += formatGenshinArtifact(art)
	}

	textRespond(ds, ic, response)
}
