package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/enescakir/emoji"
)

type React4RoleMessage struct {
	ID             int       `db:"React4RoleMessage"`
	MessageID      string    `db:"MessageID"`
	EmojiID        string    `db:"EmojiID"`
	EmojiName      string    `db:"-"`
	RoleID         string    `db:"RoleID"`
	RequiredRoleID string    `db:"RequiredRoleID"`
	CreatedAt      time.Time `db:"CreatedAt"`
}

func (r React4RoleMessage) String(roles []*discordgo.Role) string {
	role := findRoleInSlice(r.RoleID, roles)
	if role == nil {
		return "COULDN'T FIND ROLE WITH ID: " + r.RoleID
	}

	if r.RequiredRoleID != "" {
		reqRole := findRoleInSlice(r.RequiredRoleID, roles)
		return fmt.Sprintf("React to this message with %s to get the role %s (requires role %s)",
			r.FormattedEmojiString(), role.Name, reqRole.Name)
	} else {
		return fmt.Sprintf("React to this message with %s to get the role %s",
			r.FormattedEmojiString(), role.Name)
	}
}

func (r React4RoleMessage) FormattedEmojiString() string {
	if r.EmojiID != "" {
		return fmt.Sprintf("<:%s:%s>", r.EmojiName, r.EmojiID)
	} else {
		return ":" + r.EmojiName + ":"
	}
}

func onMessageReacted(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageReactionAdd) {
	return func(ds *discordgo.Session, mc *discordgo.MessageReactionAdd) {
		r4rs, err := moddingDS.react4Roles(mc.MessageID)
		if err != nil {
			notifyIfErr("onMessageReacted::react4Roles", err, ds)
			return
		}

		for _, r4r := range r4rs {
			if mc.Emoji.ID == r4r.EmojiID {

				if r4r.RequiredRoleID != "" && !isMemberInRole(mc.Member, r4r.RequiredRoleID) {
					sendDirectMessage(mc.UserID, "You can't have that role! :<", ds)
					return
				}

				action := fmt.Sprintf("Added role %s to user %s in %s", r4r.RoleID, mc.UserID, mc.GuildID)
				err := ds.GuildMemberRoleAdd(mc.GuildID, mc.UserID, r4r.RoleID)
				notifyIfErr(action, err, ds)
				if err != nil {
					log.Println(action)
				}
			}
		}
	}
}

func onMessageUnreacted(ctx context.Context) func(ds *discordgo.Session, mc *discordgo.MessageReactionRemove) {
	return func(ds *discordgo.Session, mc *discordgo.MessageReactionRemove) {
		r4rs, err := moddingDS.react4Roles(mc.MessageID)
		if err != nil {
			notifyIfErr("onMessageUnreacted::react4Roles", err, ds)
			return
		}
		for _, r4r := range r4rs {
			if mc.Emoji.ID == r4r.EmojiID {
				action := fmt.Sprintf("Removed role %s from user %s in %s", r4r.RoleID, mc.UserID, mc.GuildID)
				err := ds.GuildMemberRoleRemove(mc.GuildID, mc.UserID, r4r.RoleID)
				notifyIfErr(action, err, ds)
				if err != nil {
					log.Println(action)
				}
			}
		}
	}
}

// Command Answers

func answerMakeReact4RolesMsg(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	r4rs := extractReact4Roles(mc.Content)
	if len(r4rs) == 0 {
		log.Println(mc.Content)
		ds.ChannelMessageSend(mc.ChannelID, "Sowwy, I couldn't find any React4Role rules u_u")
		return false
	}

	roles, err := ds.GuildRoles(mc.GuildID)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "I don't have role management perms! >:(")
		return false
	}

	response, err := ds.ChannelMessageSend(mc.ChannelID, buildReact4RolesMessage(r4rs, roles))
	if err != nil {
		return false
	}
	for i, r4r := range r4rs {
		r4rs[i].MessageID = response.ID
		if r4r.EmojiID == "" {
			err = ds.MessageReactionAdd(response.ChannelID, response.ID, emoji.Parse(r4r.FormattedEmojiString()))
		} else {
			err = ds.MessageReactionAdd(response.ChannelID, response.ID, r4r.EmojiName+":"+r4r.EmojiID)
		}
		notifyIfErr("answerMakeReact4RolesMsg::MessageReactionAdd", err, ds)
	}

	err = moddingDS.addReact4Roles(r4rs)
	if err != nil {
		notifyIfErr("answerReact4Roles::addReact4Roles", err, ds)
		ds.ChannelMessageSend(mc.ChannelID, "Something went wrong, blame Jarv :3c")
		return false
	}

	return err == nil
}

// Internal functions

var react4RoleEmoteRgx = `\(\s*<:([^:]+):(\d+)>\s+(\d+)\s*(\d+)?\s*\)`
var react4RoleEmojiRgx = `\(\s*([a-z0-9_]+)\s+(\d+)\s*(\d+)?\s*\)`

func extractReact4Roles(message string) []React4RoleMessage {
	r4rs := []React4RoleMessage{}

	// Emotes
	matcher := regexp.MustCompile(react4RoleEmoteRgx)
	matches := matcher.FindAllStringSubmatch(message, 20)
	for _, m := range matches {
		r4r := React4RoleMessage{}
		r4r.EmojiName = m[1]
		r4r.EmojiID = m[2]
		r4r.RoleID = m[3]
		if len(m) == 5 {
			r4r.RequiredRoleID = m[4]
		}
		r4rs = append(r4rs, r4r)
	}

	// Emojis
	matcher = regexp.MustCompile(react4RoleEmojiRgx)
	matches = matcher.FindAllStringSubmatch(message, 20)
	for _, m := range matches {
		r4r := React4RoleMessage{}
		r4r.EmojiName = m[1]
		r4r.RoleID = m[2]
		if len(m) == 4 {
			r4r.RequiredRoleID = m[3]
		}
		r4rs = append(r4rs, r4r)
	}

	return r4rs
}

func buildReact4RolesMessage(r4rs []React4RoleMessage, roles []*discordgo.Role) string {
	msg := ""
	for _, r4r := range r4rs {
		msg += r4r.String(roles) + "\n"
	}
	return msg
}
