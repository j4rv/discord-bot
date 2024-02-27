package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Command Answers

func answerLiquid(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	_, err := ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("%06d, you know what to do with this. ", rand.Intn(450000)))
	return err == nil
}

func answerDon(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	notifyIfErr("answerDon, couldn't get timeoutRole", err, ds)
	if err != nil {
		return false
	}

	if isMemberInRole(mc.Member, timeoutRole.ID) {
		ds.ChannelMessageSend(mc.ChannelID, "Stay Realmed scum")
		return false
	}

	err = ds.GuildMemberRoleAdd(mc.GuildID, mc.Author.ID, timeoutRole.ID)
	notifyIfErr("answerDon, couldn't add timeoutRole", err, ds)
	if err != nil {
		return false
	}
	removeRoleAfterDuration(ds, mc.GuildID, mc.Author.ID, timeoutRole.ID, 10*time.Minute)
	_, err = ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("To the Shadow Realm you go %s", mc.Author.Mention()))
	return err == nil
}

func answerShoot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	match := commandWithMention.FindStringSubmatch(mc.Content)
	if match == nil || len(match) != 2 {
		ds.ChannelMessageSend(mc.ChannelID, commandWithMentionError)
		return false
	}

	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	notifyIfErr("answerShoot: get timeout role", err, ds)
	if err != nil {
		return false
	}

	shooter, err := ds.GuildMember(mc.GuildID, mc.Author.ID)
	notifyIfErr("answerShoot: get shooter member", err, ds)
	if err != nil {
		return false
	}

	target, err := ds.GuildMember(mc.GuildID, match[1])
	notifyIfErr("answerShoot: get target member", err, ds)
	if err != nil {
		return false
	}

	err = shoot(ds, mc.ChannelID, mc.GuildID, shooter, target, timeoutRole.ID)
	notifyIfErr("answerShoot: shoot", err, ds)
	return err == nil
}

func getTimeoutRole(ds *discordgo.Session, guildID string) (*discordgo.Role, error) {
	customRoleName, err := serverDS.getServerProperty(guildID, customTimeoutRoleNameKey)
	if err != nil {
		customRoleName = defaultTimeoutRoleName
	}
	return guildRoleByName(ds, guildID, customRoleName)
}

func setCustomTimeoutRole(ds *discordgo.Session, guildID string, roleName string) error {
	return serverDS.setServerProperty(guildID, customTimeoutRoleNameKey, roleName)
}

// Internal functions

func shoot(ds *discordgo.Session, channelID string, guildID string, shooter *discordgo.Member, target *discordgo.Member, timeoutRoleID string) error {
	shooterHasRoleAlready := isMemberInRole(shooter, timeoutRoleID)
	if shooterHasRoleAlready {
		ds.ChannelMessageSend(channelID, "Shadow Realmed people can't shoot dummy")
		return nil
	}

	targetHasRoleAlready := isMemberInRole(target, timeoutRoleID)
	if targetHasRoleAlready {
		ds.ChannelMessageSend(channelID, "https://giphy.com/gifs/the-simpsons-stop-hes-already-dead-JCAZQKoMefkoX6TyTb")
		return nil
	}

	if rand.Float32() <= nuclearCatastropheChance {
		ds.ChannelMessageSend(channelID, "https://c.tenor.com/fxSZIUDpQIMAAAAC/explosion-nichijou.gif")
		members, err := ds.GuildMembers(guildID, "0", 1000)
		if err != nil {
			return fmt.Errorf("guild members: %w", err)
		}
		for _, member := range members {
			if member.User.ID == ds.State.User.ID {
				continue
			}
			ds.GuildMemberRoleAdd(guildID, member.User.ID, timeoutRoleID)
			removeRoleAfterDuration(ds, guildID, member.User.ID, timeoutRoleID, timeoutDurationWhenNuclearCatastrophe)
		}
		return nil
	}

	if rand.Float32() <= shootMisfireChance || target.User.ID == ds.State.User.ID {
		ds.ChannelMessageSend(channelID, "OOPS! You missed :3c")
		ds.GuildMemberRoleAdd(guildID, shooter.User.ID, timeoutRoleID)
		removeRoleAfterDuration(ds, guildID, shooter.User.ID, timeoutRoleID, timeoutDurationWhenMisfire)
		return nil
	}

	ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got shot!", target.User.Mention()))
	ds.GuildMemberRoleAdd(guildID, target.User.ID, timeoutRoleID)
	removeRoleAfterDuration(ds, guildID, target.User.ID, timeoutRoleID, timeoutDurationWhenShot)
	return nil
}
