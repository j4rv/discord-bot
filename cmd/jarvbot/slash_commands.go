package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/eightball"
)

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
		Name:        "genshin_chances",
		Description: "Calculate your chance to get a Genshin Impact character with specific constellations and refinements",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "roll_count",
				Description: "The amount of wishes you have.",
				Required:    true,
				MinValue:    &zero,
				MaxValue:    1500,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "char_count",
				Description: "The amount of characters you want to pull. 0 for none, 1 for C0, 7 for C6.",
				Required:    true,
				MinValue:    &zero,
				MaxValue:    7,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "weapon_count",
				Description: "The amount of weapons you want to pull. 0 for none, 1 for R1, 5 for R5.",
				Required:    true,
				MinValue:    &zero,
				MaxValue:    5,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "char_pity",
				Description: "Your limited character banner 5* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    89,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_sr_char",
				Description: "If the next 5* character is guaranteed to be the rate up.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "char_rare_pity",
				Description: "Your limited character banner 4* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    9,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_rare_char",
				Description: "If the next 4* character is guaranteed to be a rate up.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "weapon_pity",
				Description: "Your limited weapon banner 5* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    89,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_sr_weapon",
				Description: "If the next 5* weapon is guaranteed to be a rate up.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "weapon_rare_pity",
				Description: "Your limited weapon banner 4* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    9,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_rare_weapon",
				Description: "If the next 4* weapon is guaranteed to be a rate up.",
				Required:    false,
			},
		},
	},
	{
		Name:        "star_rail_chances",
		Description: "Your chance to get a Honkai: Star Rail character with specific constellations and refinements",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "roll_count",
				Description: "The amount of warps you have.",
				Required:    true,
				MinValue:    &zero,
				MaxValue:    1400,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "char_count",
				Description: "The amount of characters you want to pull. 0 for none, 1 for E0, 7 for E6.",
				Required:    true,
				MinValue:    &zero,
				MaxValue:    7,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "weapon_count",
				Description: "The amount of light cones you want to pull. 0 for none, 1 for S1, 5 for S5.",
				Required:    true,
				MinValue:    &zero,
				MaxValue:    5,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "char_pity",
				Description: "Your limited character banner 5* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    89,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_sr_char",
				Description: "If the next 5* character is guaranteed to be the rate up.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "char_rare_pity",
				Description: "Your limited character banner 4* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    9,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_rare_char",
				Description: "If the next 4* character is guaranteed to be a rate up.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "weapon_pity",
				Description: "Your limited light cone banner 5* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    89,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_sr_weapon",
				Description: "If the next 5* light cone is guaranteed to be the rate up.",
				Required:    false,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "weapon_rare_pity",
				Description: "Your limited light cone banner 4* pity count.",
				Required:    false,
				MinValue:    &zero,
				MaxValue:    9,
			},
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "guaranteed_rare_weapon",
				Description: "If the next 4* light cone is guaranteed to be a rate up.",
				Required:    false,
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
	{
		Name: "Delete LinkFix Message",
		Type: discordgo.MessageApplicationCommand,
	},
}

var slashHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"help":                   answerHelp,
	"8ball":                  answer8ball,
	"avatar":                 answerAvatar,
	"genshin_chances":        expensiveSlashCommand(answerGenshinChance),
	"star_rail_chances":      expensiveSlashCommand(answerStarRailChance),
	"strongbox":              expensiveSlashCommand(answerStrongbox),
	"character":              answerCharacter,
	"warn":                   answerWarn,
	"warnings":               answerWarnings,
	"Delete LinkFix Message": answerDeleteLinkFixMessage,
}

func expensiveSlashCommand(expensiveOp func(ds *discordgo.Session, ic *discordgo.InteractionCreate)) func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	return func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
		if userExpensiveOperationOnCooldown(interactionUser(ic).ID) {
			sendDirectMessage(interactionUser(ic).ID, expensiveOperationErrorMsg, ds)
			return
		}
		userExecutedExpensiveOperation(interactionUser(ic).ID)
		expensiveOp(ds, ic)
	}
}

func optionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range opts {
		m[opt.Name] = opt
	}
	return m
}

func optionIntValueOrZero(opt *discordgo.ApplicationCommandInteractionDataOption) int {
	if opt == nil || opt.Value == nil {
		return 0
	}
	return int(opt.IntValue())
}

func optionBoolValueOrFalse(opt *discordgo.ApplicationCommandInteractionDataOption) bool {
	if opt == nil || opt.Value == nil {
		return false
	}
	return opt.BoolValue()
}

func textRespond(ds *discordgo.Session, ic *discordgo.InteractionCreate, textResponse string) (*discordgo.InteractionResponse, error) {
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: textResponse,
		},
	}
	err := ds.InteractionRespond(ic.Interaction, response)
	serverNotifyIfErr("textRespond", err, ic.GuildID, ds)
	return response, err
}

func interactionFileRespond(ds *discordgo.Session, ic *discordgo.InteractionCreate, messageContent, fileName, fileData string) {
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
	serverNotifyIfErr("fileRespond", err, ic.GuildID, ds)
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
