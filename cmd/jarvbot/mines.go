package main

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/rngx"
)

var mineTriggerMessages = []string{
	"Ooops, <user> stepped on a mine! :3c",
	"Boom boom boom boom!~ <user> blew out of the room!~",
	"<user> hit a mine... Skill issue.",
	"<user> detonated a perfectly placed mine!",
	"<user> triggered a mine.\nPress F to pay respects.",
	"<user> stepped on a mine and didn't have enough Explosion Resistance.",
	"<user> found a hidden mine! Sadly, it blew up when they picked it up.",
	"<user> just won the Big Mine Lottery! Enjoy your <role> prize.",
	"This channel was NOT safe. <user> blew up, goodbye.",
	"R.I.P. <user>\n\nSpoke at the wrong time and place.\n\n<joinyear> - <curryear>",
	"Mine detected. Oh wait, too late for <user>.",
	"!mineexplode <user>",
	// Game references
	"Rocketboo missed the Ethereal and hit <user> instead!",
	"<user> just pulled: Kaboom the Cannon!",
	"Klee's Jumpy Dumpty landed on <user>'s head!",
}

type PlaceMinesQueryInput struct {
	Where            string  `short:"w" long:"where" default:""`          // Channel id, or empty for whole guild
	Guild            string  `short:"g" long:"guild" default:""`          // Guild id, for admin user
	ChancePercentage float64 `short:"c" long:"chance" default:"1.0"`      // 5.0 = 5%
	Amount           int     `short:"a" long:"amount" default:"10"`       // Amount of mines
	Duration         string  `short:"d" long:"duration" default:"2m"`     // 99h99m99s
	CustomMessage    string  `short:"m" long:"custom_message" default:""` // Custom message that replaces the default ones
}

func answerPlaceMines(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	existing, err := serverDS.getMinesByGuild(mc.GuildID)
	if err != nil {
		adminNotifyIfErr("answerPlaceMines", err, ds)
		return false
	}

	if len(existing) >= minesMaxSetsPerGuild && mc.Author.ID != adminID {
		ds.ChannelMessageSend(mc.ChannelID, "You have too many mine sets in this server!")
		return false
	}

	var input PlaceMinesQueryInput
	err = parseCommandArgs(&input, mc.Content)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Wrong format error: "+err.Error())
		return false
	}

	if !channelBelongsToGuild(ds, input.Where, mc.GuildID) && mc.Author.ID != adminID {
		ds.ChannelMessageSend(mc.ChannelID, "Good try, but that channel doesn't belong to this Server. The Discord Police is on its way.")
		return false
	}

	if len(input.CustomMessage) > minesMaxCustomMessageLength && mc.Author.ID != adminID {
		ds.ChannelMessageSend(mc.ChannelID, "That custom message is too long.")
		return false
	}

	var guildID string
	if input.Guild != "" && mc.Author.ID == adminID {
		guildID = input.Guild
	} else {
		guildID = mc.GuildID
	}

	amount := minInt(input.Amount, minesMaxAmount)
	duration := minInt(int(stringToDuration(input.Duration).Seconds()), minesMaxDurationSeconds)
	chance := math.Max(minesMinChance, math.Min(minesMaxChance, input.ChancePercentage/100))
	err = serverDS.addMines(guildID, input.Where, amount, chance, duration, input.CustomMessage)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not add mine set: "+err.Error())
		return false
	}

	ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	return true
}

type CheckMinesQueryInput struct {
	Guild string `short:"g" long:"guild" default:""` // Guild id, for admin user
}

func answerCheckMines(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	var input CheckMinesQueryInput
	err := parseCommandArgs(&input, mc.Content)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Wrong format error: "+err.Error())
		return false
	}
	var guildID string
	var guildName string

	if input.Guild != "" && mc.Author.ID == adminID {
		guildID = input.Guild
		guild, err := ds.Guild(guildID)
		if err != nil {
			ds.ChannelMessageSend(mc.ChannelID, "Could not find Guild: "+err.Error())
			return false
		}
		guildName = guild.Name
	} else {
		guildID = mc.GuildID
		guildName = "this server"
	}

	mines, err := serverDS.getMinesByGuild(guildID)
	if err != nil {
		adminNotifyIfErr("answerCheckMines", err, ds)
		return false
	}

	if len(mines) == 0 {
		ds.ChannelMessageSend(mc.ChannelID, "No mines, wanna place some? :3")
		return true
	}

	var b strings.Builder
	for _, m := range mines {
		percent := m.Chance * 100

		if m.ChannelID == "" {
			fmt.Fprintf(&b,
				"ID: %d, %d mines in %s with %.4f%% chance of exploding.",
				m.ID, m.Amount, guildName, percent)
		} else {
			fmt.Fprintf(&b,
				"ID: %d, %d mines in channel <#%s> with %.4f%% chance of exploding.",
				m.ID, m.Amount, m.ChannelID, percent)
		}

		if m.CustomMessage != "" {
			fmt.Fprintf(&b, " Custom msg: '%s'", m.CustomMessage)
		}

		fmt.Fprintf(&b, "\n")
	}
	ds.ChannelMessageSend(mc.ChannelID, b.String())
	return true
}

func answerRemoveMines(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	fields := strings.Fields(mc.Content)
	if len(fields) == 1 {
		return false
	}

	minesetID, err := strconv.Atoi(fields[1])
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Mines ID was not a number! :<")
		return false
	}

	err = serverDS.removeMines(minesetID, mc.GuildID)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Error: "+err.Error())
		return false
	}

	ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	return true
}

func answerRemoveGuildMines(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := serverDS.removeGuildMines(mc.GuildID)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Error: "+err.Error())
		return false
	}
	ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
	return true
}

func newMessageMineCheck(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	minesets, err := serverDS.getMinesByGuildAndChannel(mc.GuildID, mc.ChannelID)
	if len(minesets) == 0 || err != nil {
		return
	}

	luck := rand.Float64()
	for _, mineset := range minesets {
		if luck <= mineset.Chance {
			err = serverDS.decrementMines(mineset.ID, 1)
			adminNotifyIfErr("decrementMines", err, ds)
			processMinesetTrigger(ds, mc, &mineset)
			return
		}
	}
}

func processMinesetTrigger(ds *discordgo.Session, mc *discordgo.MessageCreate, mineset *MineSet) {
	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	if err != nil {
		return
	}

	// Mine nuke logic
	nukeLuck := rand.Float64()
	if nukeLuck <= minesNukeChance {
		handleNuke(ds, mc.ChannelID, mc.GuildID, timeoutRole.ID, minesNukeResponse)
		err = serverDS.decrementMines(mineset.ID, 4)
		adminNotifyIfErr("decrementMines", err, ds)
		return
	}

	// Normal mine logic
	message := buildMineMessage(ds, mc, mineset)
	_, err = ds.ChannelMessageSend(mc.ChannelID, message)
	serverNotifyIfErr("Mine message could not be sent", err, mc.GuildID, ds)
	if err := ds.GuildMemberRoleAdd(mc.GuildID, mc.Author.ID, timeoutRole.ID); err == nil {
		duration := time.Duration(mineset.DurationSeconds) * time.Second
		removeShadowRealmRoleAfterDuration(mc.GuildID, mc.Author.ID, timeoutRole.ID, duration)
	}
}

func buildMineMessage(ds *discordgo.Session, mc *discordgo.MessageCreate, mineset *MineSet) string {
	var message string
	if strings.TrimSpace(mineset.CustomMessage) != "" {
		message = mineset.CustomMessage
	} else {
		message = rngx.Pick(mineTriggerMessages)
	}
	// cheap replacements
	message = strings.Replace(message, "<user>", mc.Author.Mention(), 1)
	message = strings.Replace(message, "<joinyear>", mc.Member.JoinedAt.Format("2006"), 1)
	message = strings.Replace(message, "<curryear>", time.Now().Format("2006"), 1)
	// more expensive replacements
	if strings.Contains(message, "<role>") {
		rolename := getTimeoutRoleName(ds, mc.GuildID)
		message = strings.Replace(message, "<role>", rolename, 1)
	}
	return message
}
