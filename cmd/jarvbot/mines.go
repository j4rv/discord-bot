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

type placeMinesQueryInput struct {
	Where            string  `short:"w" long:"where" default:""`          // Channel id, or empty for whole guild
	Guild            string  `short:"g" long:"guild" default:""`          // Guild id, for admin user
	ChancePercentage float64 `short:"c" long:"chance" default:"1.0"`      // 5.0 = 5%
	Amount           int     `short:"a" long:"amount" default:"10"`       // Amount of mines
	Duration         string  `short:"d" long:"duration" default:"2m"`     // 99h99m99s
	CustomMessage    string  `short:"m" long:"custom_message" default:""` // Custom message that replaces the default ones
	TriggerText      string  `short:"t" long:"trigger_text" default:""`   // If set, the mines will only explode when the message contains that text
}

type validatedMineInput struct {
	GuildID       string
	ChannelID     string
	Amount        int
	Chance        float64
	Duration      int
	CustomMessage string
	TriggerText   string
}

func parseAndValidatePlaceMinesInput(ds *discordgo.Session, mc *discordgo.MessageCreate) (*validatedMineInput, string) {
	var input placeMinesQueryInput
	if err := parseCommandArgs(&input, mc.Content); err != nil {
		return nil, "Wrong format error: " + err.Error()
	}

	existing, err := serverDS.getMinesByGuild(mc.GuildID)
	if err != nil {
		adminNotifyIfErr("parseAndValidatePlaceMinesInput", err, ds)
		return nil, "Internal server error."
	}
	if len(existing) >= minesMaxSetsPerGuild && mc.Author.ID != adminID {
		return nil, "You have too many mine sets in this server!"
	}

	if !channelBelongsToGuild(ds, input.Where, mc.GuildID) && mc.Author.ID != adminID {
		return nil, "Good try, but that channel doesn't belong to this Server. The Discord Police is on its way."
	}

	if len(input.CustomMessage) > minesMaxCustomMessageLength && mc.Author.ID != adminID {
		return nil, "That custom message is too long."
	}

	if len(input.TriggerText) > minesMaxTriggerTextLength && mc.Author.ID != adminID {
		return nil, "That trigger text is too long."
	}

	if input.Amount <= 0 {
		return nil, "Mine amount must be greater than 0."
	}
	amount := minInt(input.Amount, minesMaxAmount)

	duration := int(stringToDuration(input.Duration).Seconds())
	if duration < 0 {
		return nil, "Duration must be positive."
	}
	duration = minInt(duration, minesMaxDurationSeconds)

	guildID := mc.GuildID
	if input.Guild != "" && mc.Author.ID == adminID {
		guildID = input.Guild
	}

	var chance float64
	if input.TriggerText != "" {
		chance = 0
	} else {
		chance = math.Max(minesMinChance, math.Min(minesMaxChance, input.ChancePercentage/100))
	}

	return &validatedMineInput{
		GuildID:       guildID,
		ChannelID:     input.Where,
		Amount:        amount,
		Chance:        chance,
		Duration:      duration,
		CustomMessage: input.CustomMessage,
		TriggerText:   input.TriggerText,
	}, ""
}

func answerPlaceMines(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	validInput, errMsg := parseAndValidatePlaceMinesInput(ds, mc)
	if errMsg != "" {
		ds.ChannelMessageSend(mc.ChannelID, errMsg)
		return false
	}

	err := serverDS.addMines(validInput.GuildID, validInput.ChannelID,
		validInput.Amount, validInput.Chance, validInput.Duration,
		validInput.CustomMessage, validInput.TriggerText,
	)
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
	if err := parseCommandArgs(&input, mc.Content); err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Wrong format error: "+err.Error())
		return false
	}

	var guildID string
	if input.Guild != "" && mc.Author.ID == adminID {
		guildID = input.Guild
	} else {
		guildID = mc.GuildID
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

	// Headers
	items := []string{"ID", "Channel", "Amount", "Chance", "Msg", "Trigger"}
	columnsAmount := len(items)

	for _, m := range mines {
		channel := m.ChannelID
		if channel == "" {
			channel = "global"
		} else {
			ch, _ := ds.Channel(channel)
			if ch != nil {
				channel = ch.Name
			}
		}

		// Tidying
		chance := fmt.Sprintf("%.4f%%", m.Chance*100)
		msg := truncateString(m.CustomMessage, 30)
		trigger := truncateString(m.TriggerText, 20)

		items = append(items, fmt.Sprint(m.ID), channel, fmt.Sprint(m.Amount), chance, msg, trigger)
	}

	table := formatInColumns(items, columnsAmount, true)
	ds.ChannelMessageSend(mc.ChannelID, "```\n"+table+"```")
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

	// Shuffle the sets, to make multiple mine set with similar chances have a similar chance to explode
	rand.Shuffle(len(minesets), func(i, j int) { minesets[i], minesets[j] = minesets[j], minesets[i] })

	// Trigger text explosions
	lowerContent := strings.ToLower(mc.Content)
	for _, mineset := range minesets {
		if mineset.TriggerText != "" && strings.Contains(lowerContent, mineset.TriggerText) {
			err = serverDS.decrementMines(mineset.ID, 1)
			adminNotifyIfErr("decrementMines", err, ds)
			processMinesetTrigger(ds, mc, &mineset)
			return
		}
	}

	// Random explosions
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
		serverNotifyIfErr("fetch timeout role", err, mc.GuildID, ds)
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
	if mineset.DurationSeconds == 0 {
		return
	}
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
