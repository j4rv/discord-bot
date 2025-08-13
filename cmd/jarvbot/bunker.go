package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const bunkerServerID = "807055417120129085"
const bunkerGeneralChannelID = "828303425414365214"

// Command Answers

func answerLiquid(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	_, err := ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("%06d, you know what to do with this. ", rand.Intn(450000)))
	return err == nil
}

func answerDon(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	serverNotifyIfErr("answerDon, couldn't get timeoutRole", err, mc.GuildID, ds)
	if err != nil {
		return false
	}

	if isMemberInRole(mc.Member, timeoutRole.ID) {
		ds.ChannelMessageSend(mc.ChannelID, "Stay Realmed scum")
		return false
	}

	err = ds.GuildMemberRoleAdd(mc.GuildID, mc.Author.ID, timeoutRole.ID)
	serverNotifyIfErr("answerDon, couldn't add timeoutRole", err, mc.GuildID, ds)
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
	serverNotifyIfErr("answerShoot: get timeout role", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find the Timeout Role, maybe I'm missing permissions or it does not exist :(")
		return false
	}

	shooter, err := ds.GuildMember(mc.GuildID, mc.Author.ID)
	serverNotifyIfErr("answerShoot: get shooter member", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find you in this server, maybe I'm missing permissions u_u")
		return false
	}

	target, err := ds.GuildMember(mc.GuildID, match[1])
	serverNotifyIfErr("answerShoot: get target member", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Couldn't find member with user ID: "+match[1]+", maybe I'm missing permissions u_u")
		return false
	}

	err = shoot(ds, mc.ChannelID, mc.GuildID, shooter, target, timeoutRole.ID)
	serverNotifyIfErr("answerShoot: shoot", err, mc.GuildID, ds)
	return err == nil
}

func answerNukeTest(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	serverNotifyIfErr("answerNukeTest: get timeout role", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find the Timeout Role, maybe I'm missing permissions or it does not exist :(")
		return false
	}
	return handleNuke(ds, mc.ChannelID, mc.GuildID, timeoutRole.ID) == nil
}

func answerSniperShoot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	if mc.Author == nil {
		return false
	}

	targetID := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	if targetID == ds.State.User.ID {
		ds.ChannelMessageSend(mc.ChannelID, "https://tenor.com/view/anya-forger-anya-spy-x-family-gif-17200077238442027522")
		return false
	}

	target, err := ds.GuildMember(bunkerServerID, targetID)
	serverNotifyIfErr("answerShoot: get target member", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Couldn't find Bunker member with user ID: "+targetID)
		return false
	}

	timeoutRole, err := getTimeoutRole(ds, bunkerServerID)
	serverNotifyIfErr("answerSniperShoot: get timeout role", err, mc.GuildID, ds)
	if err != nil {
		return false
	}

	ds.ChannelMessageSend(bunkerGeneralChannelID, fmt.Sprintf("%s got sniped by %s!", target.User.Mention(), mc.Author.Mention()))
	ds.GuildMemberRoleAdd(bunkerServerID, target.User.ID, timeoutRole.ID)
	removeRoleAfterDuration(ds, bunkerServerID, target.User.ID, timeoutRole.ID, timeoutDurationWhenShot)
	ds.ChannelMessageSend(mc.ChannelID, "https://tenor.com/view/gun-anime-sniper-scope-scoping-gif-17545837")
	return true
}

func getTimeoutRole(ds *discordgo.Session, guildID string) (*discordgo.Role, error) {
	customRoleName, err := serverDS.getServerProperty(guildID, serverPropCustomTimeoutRoleName)
	if err != nil {
		customRoleName = defaultTimeoutRoleName
	}
	return guildRoleByName(ds, guildID, customRoleName)
}

func setCustomTimeoutRole(ds *discordgo.Session, guildID string, roleName string) error {
	return serverDS.setServerProperty(guildID, serverPropCustomTimeoutRoleName, roleName)
}

// Internal functions

func shoot(ds *discordgo.Session, channelID string, guildID string, shooter *discordgo.Member, target *discordgo.Member, timeoutRoleID string) error {
	if isMemberInRole(shooter, timeoutRoleID) {
		ds.ChannelMessageSend(channelID, "Shadow Realmed people can't shoot dummy")
		return nil
	}

	if isMemberInRole(target, timeoutRoleID) {
		ds.ChannelMessageSend(channelID, "https://giphy.com/gifs/the-simpsons-stop-hes-already-dead-JCAZQKoMefkoX6TyTb")
		return nil
	}

	// Nuke logic
	if rand.Float32() <= nuclearCatastropheChance {
		return handleNuke(ds, channelID, guildID, timeoutRoleID)
	}

	// Miss logic
	if rand.Float32() <= shootMisfireChance || target.User.Bot {
		ds.ChannelMessageSend(channelID, "OOPS! You missed :3c")
		err := ds.GuildMemberRoleAdd(guildID, shooter.User.ID, timeoutRoleID)
		if err == nil {
			removeRoleAfterDuration(ds, guildID, shooter.User.ID, timeoutRoleID, timeoutDurationWhenMisfire)
		}
		return nil
	}

	// Normal shot
	ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got shot!", target.User.Mention()))
	err := ds.GuildMemberRoleAdd(guildID, target.User.ID, timeoutRoleID)
	if err == nil {
		removeRoleAfterDuration(ds, guildID, target.User.ID, timeoutRoleID, timeoutDurationWhenShot)
	}
	return nil
}

func handleNuke(ds *discordgo.Session, channelID, guildID, timeoutRoleID string) error {
	memberAfter := ""
	var survivors []*discordgo.Member
	var dead []*discordgo.Member

	for {
		members, err := ds.GuildMembers(guildID, memberAfter, 1000)
		if err != nil {
			return fmt.Errorf("could not get guild members: %w", err)
		}

		for _, member := range members {
			if member.User.Bot {
				continue
			}

			if rand.Float32() <= nuclearCatastropheDeathRatio {
				dead = append(dead, member)
			} else {
				survivors = append(survivors, member)
			}
		}

		if len(members) < 1000 {
			break
		}
		memberAfter = members[len(members)-1].User.ID
	}

	// Minimum deaths (for tiny servers)
	if len(dead) < nuclearCatastropheMinDeaths && len(survivors) > 0 {
		needed := nuclearCatastropheMinDeaths - len(dead)
		if needed > len(survivors) {
			needed = len(survivors)
		}

		rand.Shuffle(len(survivors), func(i, j int) {
			survivors[i], survivors[j] = survivors[j], survivors[i]
		})
		dead = append(dead, survivors[:needed]...)
		survivors = survivors[needed:]
	}

	// Maximum deaths (for huge servers)
	if len(dead) > nuclearCatastropheMaxDeaths {
		rand.Shuffle(len(dead), func(i, j int) {
			dead[i], dead[j] = dead[j], dead[i]
		})
		dead = dead[:nuclearCatastropheMaxDeaths]
	}

	ds.ChannelMessageSend(channelID, "https://c.tenor.com/fxSZIUDpQIMAAAAC/explosion-nichijou.gif")
	for _, member := range dead {
		ds.ChannelMessageSend(channelID, fmt.Sprintf("%s died in the explosion!", member.User.Mention()))
		err := ds.GuildMemberRoleAdd(guildID, member.User.ID, timeoutRoleID)
		if err == nil {
			removeRoleAfterDuration(ds, guildID, member.User.ID, timeoutRoleID, timeoutDurationWhenNuclearCatastrophe)
		}
	}

	return nil
}
