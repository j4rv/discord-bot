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

func isMod(ds *discordgo.Session, userID, channelID string) bool {
	perms, err := ds.UserChannelPermissions(userID, channelID)
	if err != nil {
		notifyIfErr(fmt.Sprintf("ERROR isMod failed for user %s in channel %s\n", userID, channelID), err, ds)
		return false
	}
	log.Println(perms)
	log.Println(perms & discordgo.PermissionAdministrator)
	return perms&discordgo.PermissionAdministrator != 0
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
