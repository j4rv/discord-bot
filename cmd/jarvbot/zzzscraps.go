//go:build zzzscraps

// Optional commands that require a private dependency
// Note to self: Use `export GOPRIVATE=github.com/j4rv/zenless-scrapper` first

package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	zzzscraps "github.com/j4rv/zenless-scrapper"
	"golang.org/x/text/message"
)

var enPrinter = message.NewPrinter(message.MatchLanguage("en"))

func init() {
	zzzscraps.InitDb()
	zzzscraps.InitLevelCurves()
	commands["!zzzcredits"] = simpleTextResponse("Thank you to Leifa, Hawichii (and indirectly Dimbreath)")
	commands["!zzzdbupdate"] = answerZzzDbUpdate
	commands["!zzzdb"] = answerZzzDbEndgame
	commands["!zzzendgame"] = answerZzzDbEndgame
	commands["!zzzenemy"] = answerZzzDbEnemySearch

	buttonReducerMap["zzzzonelist"] = handleZzzZoneListBtn
	buttonReducerMap["zzzzone"] = handleZzzZoneBtn
	buttonReducerMap["zzzlayer"] = handleZzzLayerDescriptionBtn
	buttonReducerMap["zzzroom"] = handleZzzRoomBtn
	buttonReducerMap["zzzenemy"] = handleZzzRoomEnemyBtn
	buttonReducerMap["zzzenemycard"] = handleZzzEnemySearchCardBtn
}

// Command answers

func answerZzzDbEndgame(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	buttons := make([]*discordgo.Button, 0, 3)
	buttons = append(buttons,
		newButton("Deadly Assault", discordgo.PrimaryButton, zzzDbCustomId("zzzzonelist", mc.Author.ID, "DA", "0", "", "", "", "")),
		newButton("Shiyu Defense", discordgo.PrimaryButton, zzzDbCustomId("zzzzonelist", mc.Author.ID, "SD", "0", "", "", "", "")),
	)

	_, err := ds.ChannelMessageSendComplex(mc.ChannelID, &discordgo.MessageSend{
		Content:    "Select the Game Mode",
		Components: *buildButtonComponents(buttons),
		Reference:  &discordgo.MessageReference{MessageID: mc.ID, ChannelID: mc.ChannelID, GuildID: mc.GuildID},
	})

	serverNotifyIfErr("answerZenlessZone couldn't respond", err, mc.GuildID, ds)
	return err != nil
}

func answerZzzDbEnemySearch(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	textSearch := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	if len(textSearch) > 100 {
		ds.ChannelMessageSend(mc.ChannelID, "Error: Text too large")
		return false
	}

	buttons, err := buildEnemySearchButtons(textSearch, mc.Author.ID)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
		return true
	}

	_, err = ds.ChannelMessageSendComplex(mc.ChannelID, &discordgo.MessageSend{
		Content:    fmt.Sprintf("Found %d enemies", len(buttons)),
		Components: *buildButtonComponents(buttons),
		Reference:  &discordgo.MessageReference{MessageID: mc.ID, ChannelID: mc.ChannelID, GuildID: mc.GuildID},
	})

	adminNotifyIfErr("answerZenlessZone couldn't respond", err, ds)
	return err != nil
}

func answerZzzDbUpdate(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	cmd := exec.Command("git", "pull")
	cmd.Dir = zzzscraps.BaseDataFolder
	out, err := cmd.CombinedOutput()

	adminNotifyIfErr(fmt.Sprintf("git pull failed:\n%s", string(out)), err, ds)
	if err == nil {
		zzzscraps.RebuildDb()
		if len(out) != 0 {
			ds.ChannelMessageSend(mc.ChannelID, string(out))
		} else {
			ds.ChannelMessageSend(mc.ChannelID, commandSuccessMessage)
		}
	}

	return err != nil
}

// Button handlers

func isValidInteractionUser(ds *discordgo.Session, ic *discordgo.InteractionCreate) bool {
	data := strings.Split(ic.MessageComponentData().CustomID, buttonCustomIdSeparator)
	if len(data) < 2 {
		return false
	}

	if interactionUser(ic).ID != data[1] {
		ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You can't do that! :<",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return false
	}
	return true
}

func handleZzzZoneListBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	if !isValidInteractionUser(ds, ic) {
		return nil
	}
	ackInteraction(ds, ic)

	ownerId := data[1]
	gameMode := data[2]
	pageRaw := data[3]
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
		buttons = append(buttons, newButton(fmt.Sprintf("Zone %s", idStr), discordgo.PrimaryButton, zzzDbCustomId("zzzzone", ownerId, gameMode, pageRaw, idStr, "", "", "")))
	}
	buttons = append(buttons, newButtonWithEnabled("Older Zones", discordgo.SecondaryButton,
		zzzDbCustomId("zzzzonelist", ownerId, gameMode, strconv.Itoa(page+1), "", "", "", ""),
		len(zones) == 10))
	buttons = append(buttons, newButtonWithEnabled("Newer Zones", discordgo.SecondaryButton,
		zzzDbCustomId("zzzzonelist", ownerId, gameMode, strconv.Itoa(page-1), "", "", "", ""),
		page != 0))

	if gameMode != "SD" {
		buttons = append(buttons, newButton("Shiyu Defense", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", ownerId, "SD", "0", "", "", "", "")))
	}
	if gameMode != "DA" {
		buttons = append(buttons, newButton("Deadly Assault", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", ownerId, "DA", "0", "", "", "", "")))
	}
	/*if gameMode != "TS" {
		buttons = append(buttons, newButton("Threshold Simulation", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", ownerId, "TS", "0", "", "", "")))
	}*/

	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzZoneBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	if !isValidInteractionUser(ds, ic) {
		return nil
	}
	ackInteraction(ds, ic)

	ownerId := data[1]
	gameMode := data[2]
	pageRaw := data[3]
	zoneIdRaw := data[4]
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
		buttons = append(buttons, newButton(z.Name, discordgo.PrimaryButton, zzzDbCustomId("zzzlayer", ownerId, gameMode, pageRaw, zoneIdRaw, layerID, "", "")))
	}
	buttons = append(buttons, newButton("Back", discordgo.SecondaryButton, zzzDbCustomId("zzzzonelist", ownerId, gameMode, pageRaw, "", "", "", "")))

	content := fmt.Sprintf("**Showing Zone %d**\n", zoneId)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzLayerDescriptionBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	if !isValidInteractionUser(ds, ic) {
		return nil
	}
	ackInteraction(ds, ic)

	ownerId := data[1]
	gameMode := data[2]
	pageRaw := data[3]
	zoneIdRaw := data[4]
	layerIdRaw := data[5]
	layerId, err := strconv.Atoi(layerIdRaw)
	if err != nil {
		return err
	}
	layer, err := zzzscraps.GetLayerById(layerId)
	if err != nil {
		return err
	}

	buttons := buildLayerButtons(layer, ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw)

	content := layerResponse(layer)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzRoomBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	if !isValidInteractionUser(ds, ic) {
		return nil
	}
	ackInteraction(ds, ic)

	ownerId := data[1]
	gameMode := data[2]
	pageRaw := data[3]
	zoneIdRaw := data[4]
	layerIdRaw := data[5]
	roomIndexRaw := data[6]
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

	buttons := buildRoomButtons(layer, ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, roomIndex)

	content := layerRoomResponse(layer, room)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzRoomEnemyBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	if !isValidInteractionUser(ds, ic) {
		return nil
	}
	ackInteraction(ds, ic)

	ownerId := data[1]
	gameMode := data[2]
	pageRaw := data[3]
	zoneIdRaw := data[4]
	layerIdRaw := data[5]
	roomIndexRaw := data[6]
	enemyIndexRaw := data[7]
	layerId, err := strconv.Atoi(layerIdRaw)
	if err != nil {
		return err
	}
	roomIndex, err := strconv.Atoi(roomIndexRaw)
	if err != nil {
		return err
	}
	enemyIndex, err := strconv.Atoi(enemyIndexRaw)
	if err != nil {
		return err
	}
	layer, err := zzzscraps.GetLayerById(layerId)
	if err != nil {
		return err
	}
	room := layer.Rooms[roomIndex]
	enemy := room.Enemies[enemyIndex]

	lvlAdjustMap := zzzscraps.GetLevelAdjustMap(layer)
	content := detailedEnemyResponse(enemy, layer.EnemyLevel, lvlAdjustMap, room.IsDeadlyAssault() || room.IsThresholdSimulation())
	buttons := buildRoomButtons(layer, ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, roomIndex)
	editInteractionMessage(ds, ic, content, buttons)
	return nil
}

func handleZzzEnemySearchCardBtn(ds *discordgo.Session, ic *discordgo.InteractionCreate, data []string) error {
	if !isValidInteractionUser(ds, ic) {
		return nil
	}
	ackInteraction(ds, ic)

	ownerId := data[1]
	textSearch := data[2]
	cardIdStr := data[3]
	enemyIdStr := data[4]

	cardId, err := strconv.Atoi(cardIdStr)
	if err != nil {
		return err
	}

	var enemy *zzzscraps.Enemy
	if enemyIdStr == "" {
		enemies, err := zzzscraps.GetEnemiesByCardId(cardId)
		if err != nil {
			return err
		}
		enemy = enemies[0]
	} else {
		enemyId, err := strconv.Atoi(enemyIdStr)
		if err != nil {
			return err
		}
		enemy, err = zzzscraps.GetEnemyById(enemyId)
		if err != nil {
			return err
		}
	}

	buttons, err := buildEnemySearchEnemyCardButtons(textSearch, ownerId, cardId)
	if err != nil {
		return err
	}

	content := detailedEnemyResponse(enemy, 70, nil, false)
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

	for _, enemy := range enemies {
		fmt.Fprintf(&response, "**%s**\n", enemy.CardConfig.BriefName)
		writeEnemyStats(&response, enemy, lvl, lvlAdjust, isMultiHpBars)
		response.WriteRune('\n')
	}

	return response.String()
}

func enemyResponse(enemy *zzzscraps.Enemy, lvl int, lvlAdjust map[int]zzzscraps.EnemyLevelAdjust, isMultiHpBars bool) string {
	var response strings.Builder

	fmt.Fprintf(&response, "**%s**\n", enemy.CardConfig.BriefName)
	writeEnemyStats(&response, enemy, lvl, lvlAdjust, isMultiHpBars)
	response.WriteRune('\n')

	return response.String()
}

func detailedEnemyResponse(enemy *zzzscraps.Enemy, lvl int, lvlAdjust map[int]zzzscraps.EnemyLevelAdjust, isMultiHpBars bool) string {
	var response strings.Builder
	p := enPrinter

	fmt.Fprintf(&response, "# **%s**\n", enemy.CardConfig.BriefName)
	fmt.Fprintf(&response, "-# Id: %d\n", enemy.Id)
	fmt.Fprintf(&response, "-# %s\n", enemy.CardConfig.SkillDesc)
	fmt.Fprintf(&response, "-# Group: %s\n", enemy.CardConfig.GroupDesc)
	fmt.Fprintf(&response, "-# Tags: %s\n\n", enemy.Tags)

	writeEnemyStats(&response, enemy, lvl, lvlAdjust, isMultiHpBars)

	response.WriteString("**ATK:** ")
	atk, _ := zzzscraps.CalcEnemyAtk(enemy, lvl, lvlAdjust, *zzzscraps.EndgameDefLevelCurve)
	p.Fprintf(&response, "%.1f\n\n", atk)

	response.WriteString("**Stun multiplier:** ")
	p.Fprintf(&response, "%.0f%%\n", zzzscraps.CalcEnemyStunMultAsPct(enemy))

	response.WriteString("**Stun duration:** ")
	p.Fprintf(&response, "%.1fs\n\n", zzzscraps.CalcEnemyStunDurationInSecs(enemy))

	response.WriteRune('\n')

	dmgRes := zzzscraps.CalcEnemyResAsPct(enemy)
	dazeRes := zzzscraps.CalcEnemyDazeResAsPct(enemy)
	buildupRes := zzzscraps.CalcEnemyBuildupResAsPct(enemy)

	for _, elemName := range zzzscraps.ElementList {
		elemDmgRes := dmgRes[elemName]
		elemDazeRes := dazeRes[elemName]
		elemBuildupRes := buildupRes[elemName]
		if elemDmgRes == 0 && elemDazeRes == 0 && elemBuildupRes == 0 {
			continue
		}
		p.Fprintf(&response, "**%s DMG Res:** %.0f%%\n", elemName, elemDmgRes)
		p.Fprintf(&response, "**%s Daze Res:** %.0f%%\n", elemName, elemDazeRes)
		p.Fprintf(&response, "**%s Buildup Res:** %.0f%%\n", elemName, elemBuildupRes)
		response.WriteRune('\n')
	}

	//response.WriteString("**Buildup:** ")
	//buildup, _ := zzzscraps.CalcEnemyBuildupBar(e, lvl, lvlAdjust, *zzzscraps.EndgameBuildupLevelCurve)
	//fmt.Fprintf(&response, "%.1f\n", buildup)

	response.WriteRune('\n')
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

func buildLayerButtons(layer *zzzscraps.LayerInfo, ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw string) []*discordgo.Button {
	buttons := make([]*discordgo.Button, len(layer.Rooms)+2)
	buttons[0] = newButton("Layer info", discordgo.PrimaryButton, zzzDbCustomId("zzzlayer", ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, "", ""))
	for i := range layer.Rooms {
		roomIndex := strconv.Itoa(i)
		buttons[i+1] = newButton(fmt.Sprintf("Side %d", i+1), discordgo.PrimaryButton, zzzDbCustomId("zzzroom", ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, roomIndex, ""))
	}
	buttons[len(buttons)-1] = newButton("Back", discordgo.SecondaryButton, zzzDbCustomId("zzzzone", ownerId, gameMode, pageRaw, zoneIdRaw, "", "", ""))
	return buttons
}

func buildRoomButtons(layer *zzzscraps.LayerInfo, ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw string, roomIndex int) []*discordgo.Button {
	room := layer.Rooms[roomIndex]
	roomIndexStr := strconv.Itoa(roomIndex)
	buttons := make([]*discordgo.Button, len(room.Enemies)+3)
	buttons[0] = newButton("Layer info", discordgo.PrimaryButton, zzzDbCustomId("zzzlayer", ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, "", ""))
	buttons[1] = newButton(fmt.Sprintf("Side %d", roomIndex+1), discordgo.PrimaryButton, zzzDbCustomId("zzzroom", ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, roomIndexStr, ""))
	for i, enemy := range layer.Rooms[roomIndex].Enemies {
		enemyIndexStr := strconv.Itoa(i)
		buttons[i+2] = newButton(fmt.Sprint(enemy.CardConfig.BriefName), discordgo.PrimaryButton, zzzDbCustomId("zzzenemy", ownerId, gameMode, pageRaw, zoneIdRaw, layerIdRaw, roomIndexStr, enemyIndexStr))
	}
	buttons[len(buttons)-1] = newButton("Back", discordgo.SecondaryButton, zzzDbCustomId("zzzzone", ownerId, gameMode, pageRaw, zoneIdRaw, "", "", ""))
	return buttons
}

func buildEnemySearchButtons(textSearch, ownerId string) ([]*discordgo.Button, error) {
	cards, err := zzzscraps.SearchEnemyCardsByName(textSearch)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, errors.New("No enemies found")
	}

	buttons := make([]*discordgo.Button, 0, 3)
	for i, enemyCard := range cards {
		if i >= 5 {
			break
		}
		cardIdStr := strconv.Itoa(enemyCard.Id)
		buttons = append(buttons,
			newButton(enemyCard.BriefName, discordgo.PrimaryButton, zzzDbCustomId("zzzenemycard", ownerId, textSearch, cardIdStr, "", "", "", "")),
		)
	}

	return buttons, nil
}

func buildEnemySearchEnemyCardButtons(textSearch, ownerId string, enemyCardId int) ([]*discordgo.Button, error) {
	cards, err := zzzscraps.SearchEnemyCardsByName(textSearch)
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, errors.New("No enemies found")
	}
	var selectedCard *zzzscraps.EnemyCard

	enemies, err := zzzscraps.GetEnemiesByCardId(enemyCardId)
	if err != nil {
		return nil, err
	}

	buttons := make([]*discordgo.Button, 0, 3)
	for i, enemyCard := range cards {
		if i >= 5 {
			break
		}
		if enemyCard.Id == enemyCardId {
			selectedCard = enemyCard
		}
		cardIdStr := strconv.Itoa(enemyCard.Id)
		buttons = append(buttons,
			newButton(enemyCard.BriefName, discordgo.PrimaryButton, zzzDbCustomId("zzzenemycard", ownerId, textSearch, cardIdStr, "", "", "", "")),
		)
	}
	for _, enemy := range enemies {
		cardIdStr := strconv.Itoa(selectedCard.Id)
		enemyIdStr := strconv.Itoa(enemy.Id)
		buttons = append(buttons,
			newButton(fmt.Sprintf("%s (%d)", selectedCard.BriefName, enemy.Id),
				discordgo.PrimaryButton, zzzDbCustomId("zzzenemycard", ownerId, textSearch, cardIdStr, enemyIdStr, "", "", ""),
			),
		)
	}

	return buttons, nil
}

func zzzDbCustomId(action, ownerId, mode, page, zone, layer, room, enemy string) string {
	return strings.Join([]string{action, ownerId, mode, page, zone, layer, room, enemy}, buttonCustomIdSeparator)
}

func writeEnemyStats(
	w *strings.Builder,
	enemy *zzzscraps.Enemy,
	lvl int,
	lvlAdjust map[int]zzzscraps.EnemyLevelAdjust,
	isMultiHpBars bool,
) {
	p := enPrinter

	w.WriteString("**Level:** ")
	p.Fprintf(w, "%d\n", lvl)

	hp, _ := zzzscraps.CalcEnemyHp(enemy, lvl, lvlAdjust, *zzzscraps.EndgameHpLevelCurve)

	if isMultiHpBars {
		w.WriteString("**HP for 60k DMG Score:** ")
		p.Fprintf(w, "%.1f\n", hp*zzzscraps.DeadlyAssault60kDmgScoreHpMult)
		w.WriteString("**HP for 20k DMG Score:** ")
		p.Fprintf(w, "%.1f\n", hp*zzzscraps.DeadlyAssault20kDmgScoreHpMult)
		w.WriteString("**HP for 15k DMG Score:** ")
		p.Fprintf(w, "%.1f\n", hp*zzzscraps.DeadlyAssault15kDmgScoreHpMult)
	} else {
		w.WriteString("**HP:** ")
		p.Fprintf(w, "%.1f\n", hp)
	}

	w.WriteString("**DEF:** ")
	def, _ := zzzscraps.CalcEnemyDef(enemy, lvl, lvlAdjust, *zzzscraps.EndgameDefLevelCurve)
	p.Fprintf(w, "%.1f\n", def)

	w.WriteString("**Daze:** ")
	daze, _ := zzzscraps.CalcEnemyDazeBar(enemy, lvl, lvlAdjust, *zzzscraps.EndgameDazeLevelCurve)
	p.Fprintf(w, "%.1f\n", daze)
}
