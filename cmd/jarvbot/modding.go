package main

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kylelemons/godebug/diff"
)

func onGuildJoin(ctx context.Context) func(ds *discordgo.Session, gc *discordgo.GuildCreate) {
	return func(ds *discordgo.Session, gc *discordgo.GuildCreate) {
		sendDirectMessage(adminID, fmt.Sprintf("Joined guild: %s with id: %s", gc.Name, gc.ID), ds)
	}
}

func onMessageDeleted(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageDelete) {
	return func(ds *discordgo.Session, mc *discordgo.MessageDelete) {
		defer func() {
			if r := recover(); r != nil {
				notifyIfErr("onMessageReacted", fmt.Errorf("panic in onMessageDeleted: %s\n%s", r, string(debug.Stack())), ds)
			}
		}()

		if mc.BeforeDelete != nil && mc.BeforeDelete.Author != nil {
			// dont mind if bot messages get deleted
			if mc.BeforeDelete.Author.Bot {
				return
			}

			logsChannelID, err := serverDS.getServerProperty(mc.GuildID, serverPropMessageLogs)
			if err != nil {
				return
			}

			ds.ChannelMessageSendEmbed(
				logsChannelID,
				&discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    mc.BeforeDelete.Author.Username,
						IconURL: mc.BeforeDelete.Author.AvatarURL(""),
					},
					Color:       colorRed,
					Title:       "Message deleted",
					Description: messageToString(mc.BeforeDelete),
				},
			)
		}
	}
}

func onMessageUpdated(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageUpdate) {
	return func(ds *discordgo.Session, mc *discordgo.MessageUpdate) {
		defer func() {
			if r := recover(); r != nil {
				notifyIfErr("onMessageReacted", fmt.Errorf("panic in onMessageUpdated: %s\n%s", r, string(debug.Stack())), ds)
			}
		}()

		if mc != nil && mc.BeforeUpdate != nil && mc.Author != nil {
			// dont mind if bot messages get updated
			if mc.Author.Bot {
				return
			}

			logsChannelID, err := serverDS.getServerProperty(mc.GuildID, serverPropMessageLogs)
			if err != nil {
				return
			}

			if mc.BeforeUpdate.Content == mc.Message.Content {
				return
			}

			ds.ChannelMessageSendEmbed(
				logsChannelID,
				&discordgo.MessageEmbed{
					Author: &discordgo.MessageEmbedAuthor{
						Name:    mc.Author.Username,
						IconURL: mc.Author.AvatarURL(""),
					},
					Color:       colorYellow,
					Title:       "Message edited",
					Description: messageUpdatedToString(mc.BeforeUpdate, mc.Message),
				},
			)
		}
	}
}

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
	sendDirectMessage(mc.Author.ID, response, ds)
	return true
}

// Slash Command answers

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
		_, err = sendDirectMessage(user.ID, formattedWarningMessage, ds)
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
		interactionFileRespond(ds, ic, "Damn that user has been warned a lot", fmt.Sprintf("%s_warnings.txt", user.Username), responseMsg)
	}
}

func messageToString(m *discordgo.Message) string {
	str := "In channel: <#" + m.ChannelID + ">"
	if m.Author != nil {
		str += "\nAuthor: " + m.Author.Mention()
	}
	if m.Content != "" {
		str += "```" + m.Content + "```"
	}
	if m.Attachments != nil && len(m.Attachments) > 0 {
		str += "\nAttachments:"
		for _, a := range m.Attachments {
			str += "\n" + a.URL
		}
	}
	return str
}

func messageUpdatedToString(from, to *discordgo.Message) string {
	str := "In channel: <#" + from.ChannelID + ">"
	if from.Author != nil {
		str += "\nAuthor: " + from.Author.Mention()
	}
	str += markdownDiffBlock(diff.Diff(from.Content, to.Content), "")
	str += fmt.Sprintf("\n[Link to message](https://discord.com/channels/%s/%s/%s)",
		from.GuildID, from.ChannelID, from.ID)
	return str
}
