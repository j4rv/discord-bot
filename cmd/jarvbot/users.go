package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// because ic.User will be nil in servers -_-
func interactionUser(ic *discordgo.InteractionCreate) *discordgo.User {
	if ic.User != nil {
		return ic.User
	}
	if ic.Member != nil {
		return ic.Member.User
	}
	return nil
}

func isAdmin(userID string) bool {
	return userID == adminID
}

// isMod first checks if the user has the "Administrator" permission,
// if not, it checks if they are a Mod for the server in the bot's DB
func isMod(ds *discordgo.Session, userID, channelID string) bool {
	if userID == adminID {
		return true
	}

	perms, err := ds.UserChannelPermissions(userID, channelID)
	if err != nil {
		adminNotifyIfErr(fmt.Sprintf("ERROR isMod failed for user %s in channel %s\n", userID, channelID), err, ds)
	} else if perms&discordgo.PermissionAdministrator != 0 {
		return true
	}

	channel, err := ds.Channel(channelID)
	if channel == nil || err != nil {
		adminNotifyIfErr(fmt.Sprintf("ERROR isMod failed when retrieving channel %s\n", channelID), err, ds)
		return false
	}

	isMod, err := serverDS.ListPropertyContains(channel.GuildID, serverPropMods, userID, serverPropListSeparator)
	if err != nil {
		serverNotifyIfErr("Could not check if that user is a mod.", err, channel.GuildID, ds)
		return false
	}
	return isMod
}

var userChannels = map[string]*discordgo.Channel{}

func getUserChannel(userID string, ds *discordgo.Session) (*discordgo.Channel, error) {
	userChannel, ok := userChannels[userID]
	if !ok {
		createdChannel, err := ds.UserChannelCreate(userID)
		if err != nil {
			// If an error occurred, we failed to create the channel.
			//
			// Some common causes are:
			// 1. We don't share a server with the user (not possible here).
			// 2. We opened enough DM channels quickly enough for Discord to
			//    label us as abusing the endpoint, blocking us from opening
			//    new ones.
			log.Println("error creating user channel:", err)
			return nil, err
		}
		userChannels[userID] = createdChannel
		return createdChannel, nil
	}
	return userChannel, nil
}

func sendDirectMessage(userID string, body string, ds *discordgo.Session) (*discordgo.Message, error) {
	userChannel, err := getUserChannel(userID, ds)
	if err != nil {
		return nil, err
	}
	return ds.ChannelMessageSend(userChannel.ID, body)
}

func activeChannelMembers(ds *discordgo.Session, channelID string, keepBots bool) ([]*discordgo.User, error) {
	messagesToCheck := 100
	messages, err := ds.ChannelMessages(channelID, messagesToCheck, "", "", "")
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*discordgo.User)
	for _, msg := range messages {
		if msg.Author.Bot && !keepBots {
			continue
		}
		userMap[msg.Author.ID] = msg.Author
	}

	users := make([]*discordgo.User, 0, len(userMap))
	for _, u := range userMap {
		users = append(users, u)
	}
	return users, nil
}
