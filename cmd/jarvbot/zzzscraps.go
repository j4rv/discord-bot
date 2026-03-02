//go:build zzzscraps

// Optional commands that require a private dependency
// Note to self: Use `export GOPRIVATE=github.com/j4rv/*` first

package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	zzzscraps "github.com/j4rv/zenless-scrapper"
)

func init() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in zzzscraps init", r)
		}
	}()
	zzzscraps.RepopulateDb = true
	zzzscraps.InitDb()
	zzzscraps.InitLevelCurves()
	commands["!zzzcredits"] = simpleTextResponse("Thank you to Leifa, Hawichii (and indirectly Dimbreath)")
	commands["!zzzlayer"] = answerZenlessLayer
}

func answerZenlessLayer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	commandBody := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	layerId, err := strconv.Atoi(commandBody)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
		return false
	}

	zzzLayer, err := zzzscraps.GetLayerById(layerId)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
		return false
	} else if zzzLayer == nil {
		ds.ChannelMessageSend(mc.ChannelID, "Nil layer data")
		return false
	}

	// Buttons
	buttons := make([]discordgo.Button, len(zzzLayer.Rooms)+1)
	{
		buttonID := uuid.New().String()
		data := cachedFuncData{
			interactionDataZzzScrapsObj: zzzLayer,
		}
		buttons[0] = discordgo.Button{
			Label:    "Buffs",
			Style:    discordgo.PrimaryButton,
			CustomID: buttonID,
		}
		AddButtonHandler(buttonID, mc.Author.ID, data, layerBuffsDataResponse, buttons)
	}

	for i, room := range zzzLayer.Rooms {
		roomID := room.Id
		buttonID := uuid.New().String()

		data := cachedFuncData{
			interactionDataZzzScrapsObj: zzzLayer,
			interactionDataZzzRoomIndex: i,
		}

		buttons[i+1] = discordgo.Button{
			Label:    fmt.Sprintf("Room %d", roomID),
			Style:    discordgo.PrimaryButton,
			CustomID: buttonID,
		}
		AddButtonHandler(buttonID, mc.Author.ID, data, layerRoomDataResponse, buttons)
	}

	// Send the initial message
	content := layerBuffsResponse(*zzzLayer)
	msgResponse, err := sendMessageWithButtons(ds, mc.ChannelID, content, buttons)
	if err != nil {
		serverNotifyIfErr("ZZZ Layer couldn't respond", err, mc.GuildID, ds)
		return false
	}

	// Update the data with the message ID
	for i := range buttons {
		if x, ok := interactionCache.Get(buttons[i].CustomID); ok {
			entry := x.(cachedFunc)
			entry.data[interactionDataOriginalMessageId] = msgResponse.ID
			interactionCache.Add(buttons[i].CustomID, entry)
		}
	}

	return true
}

func layerBuffsDataResponse(d cachedFuncData) string {
	layer := d[interactionDataZzzScrapsObj]
	return layerBuffsResponse(*layer.(*zzzscraps.LayerInfo))
}

func layerRoomDataResponse(d cachedFuncData) string {
	layer := d[interactionDataZzzScrapsObj].(*zzzscraps.LayerInfo)
	roomIndex := d[interactionDataZzzRoomIndex].(int)
	room := layer.Rooms[roomIndex]
	lvlAdjustMap := zzzscraps.GetLevelAdjustMap(layer)
	return roomResponse(room, layer.EnemyLevel, lvlAdjustMap)
}

func roomResponse(r zzzscraps.RoomInfo, enemyLvl int, lvlAdjust map[int]zzzscraps.EnemyLevelAdjust) string {
	var response strings.Builder
	response.WriteString("**Weaknesses:** ")
	response.WriteString(zzzscraps.TranslateWeaknesses(r.EnemyWeaknesses))
	response.WriteRune('\n')
	response.WriteRune('\n')
	response.WriteString(enemiesResponse(r.Enemies, enemyLvl, lvlAdjust))
	return response.String()
}

func enemiesResponse(enemies []zzzscraps.Enemy, lvl int, lvlAdjust map[int]zzzscraps.EnemyLevelAdjust) string {
	var response strings.Builder
	for _, e := range enemies {
		fmt.Fprintf(&response, "**%s**", e.CardConfig.BriefName)
		response.WriteRune('\n')

		response.WriteString("**HP:** ")
		hp, _ := zzzscraps.CalcEnemyHp(&e, lvl, lvlAdjust, *zzzscraps.EndgameHpLevelCurve)
		fmt.Fprintf(&response, "%.1f", hp)
		response.WriteRune('\n')

		response.WriteString("**DEF:** ")
		def, _ := zzzscraps.CalcEnemyDef(&e, lvl, lvlAdjust, *zzzscraps.EndgameDefLevelCurve)
		fmt.Fprintf(&response, "%.1f", def)
		response.WriteRune('\n')

		response.WriteString("**Daze:** ")
		daze, _ := zzzscraps.CalcEnemyDazeBar(&e, lvl, lvlAdjust, *zzzscraps.EndgameDazeLevelCurve)
		fmt.Fprintf(&response, "%.1f", daze)
		response.WriteRune('\n')

		//response.WriteString("**Buildup:** ")
		//buildup, _ := zzzscraps.CalcEnemyBuildupBar(&e, lvl, lvlAdjust, *zzzscraps.EndgameBuildupLevelCurve)
		//fmt.Fprintf(&response, "%.1f", buildup)
		//response.WriteRune('\n')

		response.WriteRune('\n')
	}
	return response.String()
}

func layerBuffsResponse(l zzzscraps.LayerInfo) string {
	return levelAbilitiesResponse(l.LevelAbilities)
}

func levelAbilitiesResponse(l []zzzscraps.LevelAbility) string {
	var response strings.Builder
	for _, buff := range l {
		if buff.BuffName != "" {
			response.WriteString(buff.BuffName)
			response.WriteRune('\n')
		}
		response.WriteString(buff.BuffDesc)
		response.WriteRune('\n')
	}
	return response.String()
}
