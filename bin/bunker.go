package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

const shadowRealmRoleName = "Shadow Realm"
const shootMisfireChance = 0.2
const timeoutDurationWhenShot = 2 * time.Minute
const timeoutDurationWhenMisfire = 10 * time.Minute

// FIXME: Stop getting the Shadow Realm role by name, or implement a cache

func sendAuthorToShadowRealm(ds *discordgo.Session, mc *discordgo.MessageCreate) error {
	err := sendToShadowRealm(ds, mc.GuildID, mc.Author.ID)
	if err != nil {
		return err
	}
	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("To the Shadow Realm you go %s", mc.Author.Mention()))
	return nil
}

func sendToShadowRealm(ds *discordgo.Session, guildID, userID string) error {
	role, err := guildRoleByName(ds, guildID, shadowRealmRoleName)
	if err != nil {
		return err
	}
	err = ds.GuildMemberRoleAdd(guildID, userID, role.ID)
	if err != nil {
		return err
	}
	return nil
}

func removeShadowRealmRoleAfterDuration(ds *discordgo.Session, guildID, userID string, duration time.Duration) {
	role, err := guildRoleByName(ds, guildID, shadowRealmRoleName)
	notifyIfErr("removeShadowRealmRoleAfterDuration, user: "+userID, err, ds)
	if err != nil {
		return
	}
	go func() {
		time.Sleep(duration)
		ds.GuildMemberRoleRemove(guildID, userID, role.ID)
	}()
}

func shoot(ds *discordgo.Session, mc *discordgo.MessageCreate, userID string) error {
	authorIsShadowRealmed, err := isUserInRole(ds, mc.Author.ID, mc.GuildID, shadowRealmRoleName)
	if err != nil {
		return err
	}
	if authorIsShadowRealmed {
		ds.ChannelMessageSend(mc.ChannelID, "Shadow Realmed people can't shoot dummy")
		return nil
	}

	targetIsShadowRealmed, err := isUserInRole(ds, userID, mc.GuildID, shadowRealmRoleName)
	if err != nil {
		return err
	}
	if targetIsShadowRealmed {
		ds.ChannelMessageSend(mc.ChannelID, "https://giphy.com/gifs/the-simpsons-stop-hes-already-dead-JCAZQKoMefkoX6TyTb")
		return nil
	}

	if rand.Float32() <= shootMisfireChance || userID == ds.State.User.ID {
		ds.ChannelMessageSend(mc.ChannelID, "OOPS! You missed :3c")
		sendToShadowRealm(ds, mc.GuildID, mc.Author.ID)
		removeShadowRealmRoleAfterDuration(ds, mc.GuildID, mc.Author.ID, timeoutDurationWhenMisfire)
		return nil
	}

	ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("<@%s> got shot!", userID))
	sendToShadowRealm(ds, mc.GuildID, userID)
	removeShadowRealmRoleAfterDuration(ds, mc.GuildID, userID, timeoutDurationWhenShot)
	return nil
}
