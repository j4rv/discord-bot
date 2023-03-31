package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/lib/eightball"
	"github.com/j4rv/discord-bot/lib/genshinchargen"
	artis "github.com/j4rv/genshinartis"
)

var strongboxMinAmount = 1.0
var strongboxMaxAmount = 64.0
var warnMessageMinLength = 1
var warnMessageMaxLength = 320

const avatarTargetSize = "1024"

var moderatorMemberPermissions int64 = discordgo.PermissionBanMembers

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
				Description: "The artifact set name in GOOD format (Like 'GladiatorsFinale')",
				Required:    true,
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
	{
		Name:        "character",
		Description: "Generate a Genshin Impact character",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name",
				Description: "The character's name",
				Required:    true,
				MaxLength:   20,
			},
		},
	},
	{
		Name:        "abyss_challenge",
		Description: "Try yo beat Abyss with the result",
	},
	{
		Name:                     "warn",
		DefaultMemberPermissions: &moderatorMemberPermissions,
		Description:              "Warn a user (mods)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user you want to warn",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "The reason for the warning",
				Required:    true,
				MinLength:   &(warnMessageMinLength),
				MaxLength:   warnMessageMaxLength,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "ping",
				Description: "If true, the bot will DM the warned user.",
				Required:    true,
			},
		},
	},
	{
		Name:                     "warnings",
		DefaultMemberPermissions: &moderatorMemberPermissions,
		Description:              "Check the warnings for a user (mods)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user",
				Required:    true,
			},
		},
	},
}

var slashHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"help":            answerHelp,
	"8ball":           answer8ball,
	"avatar":          answerAvatar,
	"strongbox":       answerStrongbox,
	"character":       answerCharacter,
	"abyss_challenge": answerAbyssChallenge,
	"warn":            answerWarn,
	"warnings":        answerWarnings,
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

func textRespond(ds *discordgo.Session, ic *discordgo.InteractionCreate, textResponse string) {
	err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: textResponse,
		},
	})
	notifyIfErr("textRespond", err, ds)
}

func fileRespond(ds *discordgo.Session, ic *discordgo.InteractionCreate, messageContent, fileName, fileData string) {
	err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: messageContent,
			Files: []*discordgo.File{
				{
					ContentType: "text/plain",
					Name:        fileName,
					Reader:      strings.NewReader(fileData),
				},
			},
		},
	})
	notifyIfErr("fileRespond", err, ds)
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

func answerWarn(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	g, err := ds.State.Guild(ic.GuildID)
	if err != nil {
		textRespond(ds, ic, "Couldn't get the Guild's name :(")
		return
	}

	user := ic.ApplicationCommandData().Options[0].UserValue(ds)
	message := ic.ApplicationCommandData().Options[1].StringValue()
	ping := ic.ApplicationCommandData().Options[2].BoolValue()
	err = moddingDS.WarnUser(user.ID, interactionUser(ic).ID, ic.GuildID, message)
	if err != nil {
		textRespond(ds, ic, "There was an error storing the warning: "+err.Error())
		return
	}

	if ping {
		formattedWarningMessage := fmt.Sprintf("**You have been warned in %s server** for the following reason:\n*%s*", g.Name, message)
		_, err = userMessageSend(user.ID, formattedWarningMessage, ds)
		if err != nil {
			textRespond(ds, ic, "Warning recorded, but couldn't send the warning to the user: "+err.Error())
			return
		}
	}

	textRespond(ds, ic, fmt.Sprintf("The user %s#%s has been warned. Reason: '%s'", user.Username, user.Discriminator, message))
}

func answerWarnings(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	user := ic.ApplicationCommandData().Options[0].UserValue(ds)
	warnings, err := moddingDS.UserWarnings(user.ID, ic.GuildID)
	if err != nil {
		textRespond(ds, ic, "Couldn't get the user warnings: "+err.Error())
		return
	}

	responseMsg := fmt.Sprintf("%s has been warned %d times:\n", user.Mention(), len(warnings))
	for _, warning := range warnings {
		responseMsg += warning.ShortString() + "\n"
	}

	if len(responseMsg) < discordMaxMessageLength {
		textRespond(ds, ic, responseMsg)
	} else {
		fileRespond(ds, ic, "Damn that user has been warned a lot", fmt.Sprintf("%s_warnings.txt", user.Username), responseMsg)
	}
}

func answerStrongbox(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	set := ic.ApplicationCommandData().Options[0].StringValue()
	amount := int(ic.ApplicationCommandData().Options[1].IntValue())

	message := fmt.Sprintf("%s is Strongboxing %d %s artifacts:\n", interactionUser(ic).Mention(), amount, set)
	var arts []*artis.Artifact
	for i := 0; i < amount; i++ {
		art := artis.RandomArtifactOfSet(set, artis.StrongboxBase4Chance)
		arts = append(arts, art)
	}

	good, err := json.Marshal(artis.ExportToGOOD(arts))
	if err != nil {
		notifyIfErr("answerStrongbox_jsonMarshal", err, ds)
		textRespond(ds, ic, "Oops, error")
		return
	}

	fileRespond(ds, ic, message, "StrongboxResult.json", string(good))
}

func answerAbyssChallenge(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	textRespond(ds, ic, newAbyssChallenge())
}

func answerCharacter(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	name := ic.ApplicationCommandData().Options[0].StringValue()
	textRespond(ds, ic, genshinchargen.NewChar(name, unixDay()).PrettyString())
}

const helpResponse = `Available commands:
- **!source**: Links to the bot's source code
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
