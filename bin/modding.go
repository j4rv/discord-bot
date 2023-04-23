package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type UserWarning struct {
	ID         int       `db:"UserWarning"`
	UserID     string    `db:"DiscordUserID"`
	WarnedByID string    `db:"WarnedByID"`
	GuildID    string    `db:"GuildID"`
	Reason     string    `db:"Reason"`
	CreatedAt  time.Time `db:"CreatedAt"`
}

func (u UserWarning) ShortString() string {
	return fmt.Sprintf("By <@%s> at <t:%d>, reason: '%s'", u.WarnedByID, u.CreatedAt.Unix(), u.Reason)
}

// Command Answers

func answerRoleIDs(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	roles, err := ds.GuildRoles(mc.GuildID)
	notifyIfErr("answerRoleIDs", err, ds)
	if err != nil {
		return false
	}
	var response string
	for _, role := range roles {
		if role.Name == roleEveryone {
			continue
		}
		response += fmt.Sprintf("%s: %s\n", role.Name, role.ID)
	}
	userMessageSend(adminID, response, ds)
	return true
}

// Slash commands

func answerWarn(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	g, err := ds.State.Guild(ic.GuildID)
	if err != nil {
		textRespond(ds, ic, "Couldn't get the Guild's name :(")
		return
	}

	user := ic.ApplicationCommandData().Options[0].UserValue(ds)
	message := ic.ApplicationCommandData().Options[1].StringValue()
	ping := ic.ApplicationCommandData().Options[2].BoolValue()
	err = moddingDS.warnUser(user.ID, interactionUser(ic).ID, ic.GuildID, message)
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
	warnings, err := moddingDS.userWarnings(user.ID, ic.GuildID)
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
