package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

const timeoutRoleName = "Shadow Realm"
const shootMisfireChance = 0.2
const nuclearCatastropheChance = 0.006
const timeoutDurationWhenShot = 1 * time.Minute
const timeoutDurationWhenMisfire = 10 * time.Minute
const timeoutDurationWhenNuclearCatastrophe = 30 * time.Second

func removeRoleAfterDuration(ds *discordgo.Session, guildID string, memberID string, roleID string, duration time.Duration) {
	go func() {
		time.Sleep(duration)
		ds.GuildMemberRoleRemove(guildID, memberID, roleID)
	}()
}

func shoot(ds *discordgo.Session, channelID string, guildID string, shooter *discordgo.Member, target *discordgo.Member, timeoutRoleID string) error {
	shooterHasRoleAlready, err := isMemberInRole(shooter, timeoutRoleID)
	if err != nil {
		return err
	}
	if shooterHasRoleAlready {
		ds.ChannelMessageSend(channelID, "Shadow Realmed people can't shoot dummy")
		return nil
	}

	targetHasRoleAlready, err := isMemberInRole(target, timeoutRoleID)
	if err != nil {
		return err
	}
	if targetHasRoleAlready {
		ds.ChannelMessageSend(channelID, "https://giphy.com/gifs/the-simpsons-stop-hes-already-dead-JCAZQKoMefkoX6TyTb")
		return nil
	}

	if rand.Float32() <= nuclearCatastropheChance {
		ds.ChannelMessageSend(channelID, "https://c.tenor.com/fxSZIUDpQIMAAAAC/explosion-nichijou.gif")
		/*members, err := ds.GuildMembers(guildID, "0", 1000)
		if err != nil {
			return fmt.Errorf("guild members: %w", err)
		}
		for _, member := range members {
			if member.User.ID == ds.State.User.ID {
				continue
			}
			ds.GuildMemberRoleAdd(guildID, member.User.ID, timeoutRoleID)
			removeRoleAfterDuration(ds, guildID, member.User.ID, timeoutRoleID, timeoutDurationWhenNuclearCatastrophe)
		}*/
		return nil
	}

	if rand.Float32() <= shootMisfireChance || target.User.ID == ds.State.User.ID {
		ds.ChannelMessageSend(channelID, "OOPS! You missed :3c")
		/*ds.GuildMemberRoleAdd(guildID, shooter.User.ID, timeoutRoleID)
		removeRoleAfterDuration(ds, guildID, shooter.User.ID, timeoutRoleID, timeoutDurationWhenMisfire)*/
		return nil
	}

	ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got shot!", target.User.Mention()))
	/*ds.GuildMemberRoleAdd(guildID, target.User.ID, timeoutRoleID)
	removeRoleAfterDuration(ds, guildID, target.User.ID, timeoutRoleID, timeoutDurationWhenShot)*/
	return nil
}
