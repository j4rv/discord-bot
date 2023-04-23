package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/lib/eightball"
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

func textRespond(ds *discordgo.Session, ic *discordgo.InteractionCreate, textResponse string) (*discordgo.InteractionResponse, error) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: textResponse,
		},
	}
	err := ds.InteractionRespond(ic.Interaction, response)
	notifyIfErr("textRespond", err, ds)
	return response, err
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

// Slash Command answers

func answerHelp(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	textRespond(ds, ic, "https://github.com/j4rv/discord-bot/wiki/Help")
}

func answerAvatar(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	user := ic.ApplicationCommandData().Options[0].UserValue(ds)
	textRespond(ds, ic, user.AvatarURL(avatarTargetSize))
}

func answer8ball(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	question := ic.ApplicationCommandData().Options[0].StringValue()
	response := fmt.Sprintf("%s asked: %s\nThe 8 Ball says...\n'%s'",
		interactionUser(ic).Mention(), question, eightball.Response())
	textRespond(ds, ic, response)
}
