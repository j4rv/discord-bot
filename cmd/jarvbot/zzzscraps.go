//go:build zzzscraps

// Optional commands that require a private dependency
// Note to self: Use `export GOPRIVATE=github.com/j4rv/*` first

package main

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	zzzscraps "github.com/j4rv/zenless-scrapper"
	"golang.org/x/text/message"
)

func init() {
	zzzscraps.RepopulateDb = true
	zzzscraps.InitDb()
	zzzscraps.InitLevelCurves()
	commands["!zzzcredits"] = simpleTextResponse("Thank you to Leifa, Hawichii (and indirectly Dimbreath)")
	commands["!zzzdbupdate"] = answerZzzDbUpdate
	commands["!zzzdb"] = answerZzzDb

	buttonReducerMap["zzzzonelist"] = handleZzzZoneListBtn
	buttonReducerMap["zzzzone"] = handleZzzZoneBtn
	buttonReducerMap["zzzlayer"] = handleZzzLayerDescriptionBtn
	buttonReducerMap["zzzroom"] = handleZzzRoomBtn
}

// Command answers

func answerZzzDb(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	buttons := make([]*discordgo.Button, 0, 3)
	buttons = append(buttons,
		newButton("Deadly Assault", discordgo.PrimaryButton, zzzDbCustomId("zzzzonelist", "DA", "0", "", "", "")),
		newButton("Shiyu Defense", discordgo.PrimaryButton, zzzDbCustomId("zzzzonelist", "SD", "0", "", "", "")),
	)

	_, err := sendMessageWithButtons(ds, mc.ChannelID, "Select the Game Mode", buttons)
	if err != nil {
		serverNotifyIfErr("answerZenlessZone couldn't respond", err, mc.GuildID, ds)
		return false
	}
	return true
}

func answerZzzDbUpdate(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	cmd := exec.Command("git", "pull")
	cmd.Dir = zzzscraps.BaseDataFolder
	out, err := cmd.CombinedOutput()
	adminNotifyIfErr(fmt.Sprintf("git pull failed:\n%s", string(out)), err, ds)
	return err != nil
}

// Button handlers

func handleZzzZoneListBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	ackInteraction(ds, ic)
	gameMode := data[1]
	pageRaw := data[2]
	page, err := strconv.Atoi(pageRaw)
	if err != nil {
		return err
	}
	var rangeStart, rangeEnd int
	content := ""

	switch gameMode {
	case "DA":
		rangeStart, rangeEnd = zzzscraps.DeadlyAssaultStartId, zzzscraps.DeadlyAssaultEndId
		content = "**Showing Deadly Assault Zone IDs**"
	case "SD":
		rangeStart, rangeEnd = zzzscraps.ShiyuDefenseStartId, zzzscraps.ShiyuDefenseEndId
		content = "**Showing Shiyu Defense Zone IDs**"
	case "TS":
		rangeStart, rangeEnd = zzzscraps.ThreshholdSimulationStartId, zzzscraps.ThreshholdSimulationEndId
		content = "**Showing Threshold Simulation Zone IDs**"
	}

	zones, err := zzzscraps.GetLatestZoneIDsInRange(rangeStart, rangeEnd, 10, page)

	buttons := make([]*discordgo.Button, 0, len(zones))
	for _, id := range zones {
		idStr := strconv.Itoa(id)
		buttons = append(buttons, newButton(fmt.Sprintf("Zone %s", idStr), discordgo.PrimaryButton, zzzDbCustomId("zzzzone", gameMode, pageRaw, idStr, "", "")))
	}
	buttons = append(buttons, newButtonWithEnabled("Older Zones", discordgo.SecondaryButton,
		zzzDbCustomId("zzzzonelist", gameMode, strconv.Itoa(page+1), "", "", ""),
		len(zones) == 10))
	buttons = append(buttons, newButtonWithEnabled("Newer Zones", discordgo.SecondaryButton,
		zzzDbCustomId("zzzzonelist", gameMode, strconv.Itoa(page-1), "", "", ""),
		page != 0))

	buttons = append(buttons,
		newButtonWithEnabled("Shiyu Defense", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", "SD", "0", "", "", ""), gameMode != "SD"),
		newButtonWithEnabled("Deadly Assault", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", "DA", "0", "", "", ""), gameMode != "DA"),
		newButtonWithEnabled("Threshold Simulation", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", "TS", "0", "", "", ""), false),
	)

	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzZoneBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	ackInteraction(ds, ic)
	gameMode := data[1]
	pageRaw := data[2]
	zoneIdRaw := data[3]
	zoneId, err := strconv.Atoi(zoneIdRaw)
	if err != nil {
		return err
	}

	zones, err := zzzscraps.GetZonesById(zoneId)
	if err != nil {
		return err
	} else if len(zones) == 0 {
		return fmt.Errorf("No zones found")
	}

	buttons := make([]*discordgo.Button, 0, len(zones))
	for _, z := range zones {
		if z.Name == "" {
			continue
		}
		layerID := strconv.Itoa(z.LayerInfoId)
		buttons = append(buttons, newButton(z.Name, discordgo.PrimaryButton, zzzDbCustomId("zzzlayer", gameMode, pageRaw, zoneIdRaw, layerID, "")))
	}
	buttons = append(buttons, newButton("Back", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", gameMode, pageRaw, "", "", "")))

	content := fmt.Sprintf("**Showing Zone %d**\n", zoneId)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzLayerDescriptionBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	ackInteraction(ds, ic)
	gameMode := data[1]
	pageRaw := data[2]
	zoneIdRaw := data[3]
	layerIdRaw := data[4]
	layerId, err := strconv.Atoi(layerIdRaw)
	if err != nil {
		return err
	}
	layer, err := zzzscraps.GetLayerById(layerId)
	if err != nil {
		return err
	}

	buttons := buildLayerButtons(layer, gameMode, pageRaw, zoneIdRaw, layerIdRaw)

	content := layerResponse(layer)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzRoomBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	ackInteraction(ds, ic)
	gameMode := data[1]
	pageRaw := data[2]
	zoneIdRaw := data[3]
	layerIdRaw := data[4]
	roomIndexRaw := data[5]
	layerId, err := strconv.Atoi(layerIdRaw)
	if err != nil {
		return err
	}
	roomIndex, err := strconv.Atoi(roomIndexRaw)
	if err != nil {
		return err
	}
	layer, err := zzzscraps.GetLayerById(layerId)
	if err != nil {
		return err
	}
	room := layer.Rooms[roomIndex]

	buttons := buildLayerButtons(layer, gameMode, pageRaw, zoneIdRaw, layerIdRaw)

	content := layerRoomResponse(layer, room)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

// Text response formatters

func layerRoomResponse(layer *zzzscraps.LayerInfo, room *zzzscraps.RoomInfo) string {
	var response strings.Builder
	lvlAdjustMap := zzzscraps.GetLevelAdjustMap(layer)
	fmt.Fprintf(&response, "**Showing Layer Room:** %d \n", room.Id)
	response.WriteString(roomResponse(room, layer.EnemyLevel, lvlAdjustMap))
	return response.String()
}

func roomResponse(r *zzzscraps.RoomInfo, enemyLvl int, lvlAdjust map[int]zzzscraps.EnemyLevelAdjust) string {
	var response strings.Builder

	weaknesses := zzzscraps.TranslateWeaknesses(r.EnemyWeaknesses)
	if weaknesses != "" {
		fmt.Fprintf(&response, "```Weaknesses: %s```", weaknesses)
	}
	response.WriteString(roomStageEffectsResponse(r))
	response.WriteRune('\n')
	response.WriteString(enemiesResponse(r.Enemies, enemyLvl, lvlAdjust, r.IsDeadlyAssault() || r.IsThresholdSimulation()))
	return response.String()
}

func roomStageEffectsResponse(r *zzzscraps.RoomInfo) string {
	var response strings.Builder
	for i, e := range r.StageEffects {
		if i != 0 {
			response.WriteRune('\n')
		}
		if e.BuffName != "" {
			fmt.Fprintf(&response, "%s:\n%s\n", e.BuffName, e.BuffDesc)
		} else {
			fmt.Fprintf(&response, "Unnamed:\n%s\n", e.BuffDesc)
		}
	}
	result := response.String()
	if result == "" {
		return result
	}
	return "```" + result + "```"
}

func enemiesResponse(enemies []*zzzscraps.Enemy, lvl int, lvlAdjust map[int]zzzscraps.EnemyLevelAdjust, isMultiHpBars bool) string {
	var response strings.Builder
	p := message.NewPrinter(message.MatchLanguage("en"))

	for _, e := range enemies {
		fmt.Fprintf(&response, "**%s**", e.CardConfig.BriefName)
		response.WriteRune('\n')

		if isMultiHpBars {
			hp, _ := zzzscraps.CalcEnemyHp(e, lvl, lvlAdjust, *zzzscraps.EndgameHpLevelCurve)
			response.WriteString("**HP for 60k DMG Score:** ")
			p.Fprintf(&response, "%.1f\n", hp*zzzscraps.DeadlyAssault65kDmgScoreHpMult)
			response.WriteString("**HP for 20k DMG Score:** ")
			p.Fprintf(&response, "%.1f\n", hp*zzzscraps.DeadlyAssault20kDmgScoreHpMult)
			response.WriteString("**HP for 15k DMG Score:** ")
			p.Fprintf(&response, "%.1f\n", hp*zzzscraps.DeadlyAssault15kDmgScoreHpMult)

		} else {
			response.WriteString("**HP:** ")
			hp, _ := zzzscraps.CalcEnemyHp(e, lvl, lvlAdjust, *zzzscraps.EndgameHpLevelCurve)
			p.Fprintf(&response, "%.1f\n", hp)
		}

		response.WriteString("**DEF:** ")
		def, _ := zzzscraps.CalcEnemyDef(e, lvl, lvlAdjust, *zzzscraps.EndgameDefLevelCurve)
		p.Fprintf(&response, "%.1f\n", def)

		response.WriteString("**Daze:** ")
		daze, _ := zzzscraps.CalcEnemyDazeBar(e, lvl, lvlAdjust, *zzzscraps.EndgameDazeLevelCurve)
		p.Fprintf(&response, "%.1f\n", daze)

		//response.WriteString("**Buildup:** ")
		//buildup, _ := zzzscraps.CalcEnemyBuildupBar(e, lvl, lvlAdjust, *zzzscraps.EndgameBuildupLevelCurve)
		//fmt.Fprintf(&response, "%.1f\n", buildup)

		response.WriteRune('\n')
	}
	return response.String()
}

func layerResponse(l *zzzscraps.LayerInfo) string {
	var response strings.Builder
	fmt.Fprintf(&response, "**Showing Layer:** %d \n", l.Id)
	response.WriteString(levelAbilitiesResponse(l.LevelAbilities))
	return response.String()
}

func levelAbilitiesResponse(l []*zzzscraps.LevelAbility) string {
	var response strings.Builder
	for _, buff := range l {
		if buff.BuffName != "" {
			response.WriteString(buff.BuffName)
			response.WriteRune('\n')
		}
		response.WriteString(buff.BuffDesc)
		response.WriteRune('\n')
	}
	result := response.String()
	if result == "" {
		return result
	}
	return "```" + result + "```"
}

// Utils

func buildLayerButtons(layer *zzzscraps.LayerInfo, gameMode, pageRaw, zoneIdRaw, layerIdRaw string) []*discordgo.Button {
	buttons := make([]*discordgo.Button, len(layer.Rooms)+2)
	buttons[0] = newButton(fmt.Sprintf("Layer %d", layer.Id), discordgo.PrimaryButton, zzzDbCustomId("zzzlayer", gameMode, pageRaw, zoneIdRaw, layerIdRaw, ""))
	for i, room := range layer.Rooms {
		roomIndex := strconv.Itoa(i)
		buttons[i+1] = newButton(fmt.Sprintf("Room %d", room.Id), discordgo.PrimaryButton, zzzDbCustomId("zzzroom", gameMode, pageRaw, zoneIdRaw, layerIdRaw, roomIndex))
	}
	buttons[len(buttons)-1] = newButton("Back", discordgo.SecondaryButton, zzzDbCustomId("zzzzone", gameMode, pageRaw, zoneIdRaw, "", ""))
	return buttons
}

func zzzDbCustomId(action, mode, page, zone, layer, room string) string {
	return strings.Join([]string{action, mode, page, zone, layer, room}, buttonCustomIdSeparator)
}
