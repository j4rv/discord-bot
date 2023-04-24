package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/genshinchargen"
	"github.com/j4rv/discord-bot/pkg/rngx"
	artis "github.com/j4rv/genshinartis"
)

const genshinTeamSize = 4

// Command Answers

func answerParametricTransformer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := startParametricReminder(ds, mc, ctx)
	notifyIfErr("answerParametricTransformer", err, ds)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the Parametric Transformer in 7 days!")
	}
	return err == nil
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := stopParametricReminder(ds, mc, ctx)
	notifyIfErr("answerParametricTransformerStop", err, ds)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
	return err == nil
}

func answerPlayStore(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := startPlayStoreReminder(ds, mc, ctx)
	notifyIfErr("answerPlayStore", err, ds)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the PlayStore in 7 days!")
	}
	return err == nil
}

func answerPlayStoreStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := stopPlayStoreReminder(ds, mc, ctx)
	notifyIfErr("answerPlayStoreStop", err, ds)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
	return err == nil
}

func answerRandomAbyssLineup(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	var firstTeam, secondTeam [4]string
	var replacements []string

	// Process Input and generate the teams
	inputString := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	inputChars := strings.Split(inputString, ",")
	if inputChars[0] != "" && len(inputChars) < genshinTeamSize*2 {
		ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf(`Not enough characters! Please enter at least %d`, genshinTeamSize*2))
		return false
	}
	for i := range inputChars {
		inputChars[i] = strings.TrimSpace(inputChars[i])
	}
	firstTeam, secondTeam, replacements = randomAbyssLineup(inputChars...)

	// Format the teams into readable text
	formattedFirstTeam, formattedSecondTeam, formattedReplacements := "```\n", "```\n", "```\n"
	for _, r := range replacements {
		formattedReplacements += r + "\n"
	}
	for i := 0; i < genshinTeamSize; i++ {
		formattedFirstTeam += firstTeam[i] + "\n"
		formattedSecondTeam += secondTeam[i] + "\n"
	}
	formattedFirstTeam += "```"
	formattedSecondTeam += "```"
	formattedReplacements += "```"

	_, err := ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf(`
You can only replace one character on each team with one of the replacements.

**Team 1:**
%s
**Team 2:**
%s
**Replacements:**
%s
`, formattedFirstTeam, formattedSecondTeam, formattedReplacements))
	return err == nil
}

func answerRandomArtifact(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	artifact := artis.RandomArtifact(artis.DomainBase4Chance)
	_, err := ds.ChannelMessageSend(mc.ChannelID, formatGenshinArtifact(artifact))
	return err == nil
}

func answerRandomArtifactSet(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	flower := artis.RandomArtifactOfSlot(artis.SlotFlower, artis.DomainBase4Chance)
	plume := artis.RandomArtifactOfSlot(artis.SlotPlume, artis.DomainBase4Chance)
	sands := artis.RandomArtifactOfSlot(artis.SlotSands, artis.DomainBase4Chance)
	goblet := artis.RandomArtifactOfSlot(artis.SlotGoblet, artis.DomainBase4Chance)
	circlet := artis.RandomArtifactOfSlot(artis.SlotCirclet, artis.DomainBase4Chance)
	msg := formatGenshinArtifact(flower)
	msg += formatGenshinArtifact(plume)
	msg += formatGenshinArtifact(sands)
	msg += formatGenshinArtifact(goblet)
	msg += formatGenshinArtifact(circlet)
	_, err := ds.ChannelMessageSend(mc.ChannelID, msg)
	return err == nil
}

func answerRandomDomainRun(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	match := commandWithTwoArguments.FindStringSubmatch(mc.Content)
	if match == nil || len(match) != 3 {
		ds.ChannelMessageSend(mc.ChannelID, commandWithTwoArgumentsError)
		return false
	}

	// we also remove the "(" and ")" chars
	set1 := match[1][1 : len(match[1])-1]
	set2 := match[2][1 : len(match[2])-1]
	art1 := artis.RandomArtifactFromDomain(set1, set2)
	art2 := artis.RandomArtifactFromDomain(set1, set2)
	msg := formatGenshinArtifact(art1)
	msg += formatGenshinArtifact(art2)
	_, err := ds.ChannelMessageSend(mc.ChannelID, msg)
	return err == nil
}

func answerGenshinDailyCheckIn(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := startDailyCheckInReminder(ds, mc, ctx)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	}
	notifyIfErr("answerGenshinDailyCheckIn", err, ds)
	return err == nil
}

func answerGenshinDailyCheckInStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := stopDailyCheckInReminder(ds, mc, ctx)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
	notifyIfErr("answerGenshinDailyCheckInStop", err, ds)
	return err == nil
}

// Slash Command answers

func answerStrongbox(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	set := ic.ApplicationCommandData().Options[0].StringValue()
	amount := int(ic.ApplicationCommandData().Options[1].IntValue())

	message := fmt.Sprintf("%s is Strongboxing %d %s artifacts:\n", interactionUser(ic).Mention(), amount, set)
	var arts []*artis.Artifact
	for i := 0; i < amount; i++ {
		art := artis.RandomArtifactOfSet(set, artis.StrongboxBase4Chance)
		arts = append(arts, art)
	}

	good, err := json.Marshal(artis.ExportToGOOD(arts))
	if err != nil {
		notifyIfErr("answerStrongbox_jsonMarshal", err, ds)
		textRespond(ds, ic, "Oops, error")
		return
	}

	fileRespond(ds, ic, message, "StrongboxResult.json", string(good))
}

func answerAbyssChallenge(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	textRespond(ds, ic, newAbyssChallenge())
}

func answerCharacter(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	name := ic.ApplicationCommandData().Options[0].StringValue()
	textRespond(ds, ic, genshinchargen.NewChar(name, unixDay()).PrettyString())
}

// CRONs

func dailyCheckInCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allDailyCheckInReminderUserIDs()
		notifyIfErr("allDailyCheckInReminderUserIDs", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to do the Daily CheckIn", len(userIDs))
			for _, userID := range userIDs {
				sendDirectMessage(userID, dailyCheckInReminderMessage, ds)
			}
		}
	}
}

func parametricCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allParametricReminderUserIDsToBeReminded()
		notifyIfErr("allParametricReminderUserIDsToBeReminded", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to use the Parametric Transformer", len(userIDs))
			for _, userID := range userIDs {
				sendDirectMessage(userID, parametricReminderMessage, ds)
				err := genshinDS.addOrUpdateParametricReminder(userID)
				notifyIfErr("addOrUpdateParametricReminder", err, ds)
			}
		}
	}
}

func playStoreCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allPlayStoreReminderUserIDsToBeReminded()
		notifyIfErr("allPlayStoreReminderUserIDsToBeReminded", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to get the Play Store prize", len(userIDs))
			for _, userID := range userIDs {
				sendDirectMessage(userID, playStoreReminderMessage, ds)
				err := genshinDS.addOrUpdatePlayStoreReminder(userID)
				notifyIfErr("addOrUpdatePlayStoreReminder", err, ds)
			}
		}
	}
}

func startDailyCheckInReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) error {
	return genshinDS.addDailyCheckInReminder(mc.Author.ID)
}

func stopDailyCheckInReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) error {
	return genshinDS.removeDailyCheckInReminder(mc.Author.ID)
}

func startParametricReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) error {
	return genshinDS.addOrUpdateParametricReminder(mc.Author.ID)
}

func stopParametricReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) error {
	return genshinDS.removeParametricReminder(mc.Author.ID)
}

func startPlayStoreReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) error {
	return genshinDS.addOrUpdatePlayStoreReminder(mc.Author.ID)
}

func stopPlayStoreReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) error {
	return genshinDS.removePlayStoreReminder(mc.Author.ID)
}

// chars must have either 0 or 8+ elements
func randomAbyssLineup(chars ...string) (firstTeam, secondTeam [genshinTeamSize]string, replacements []string) {
	if len(chars) == 0 || chars[0] == "" {
		chars = allGenshinChars()
	}

	for i := 0; i < genshinTeamSize; i++ {
		firstTeam[i] = rngx.PickAndRemove(&chars)
		secondTeam[i] = rngx.PickAndRemove(&chars)
	}

	if len(chars) < genshinTeamSize {
		replacements = chars
	} else {
		for i := 0; i < genshinTeamSize; i++ {
			replacements = append(replacements, rngx.PickAndRemove(&chars))
		}
	}

	return firstTeam, secondTeam, replacements
}

func formatGenshinArtifact(artifact *artis.Artifact) string {
	return fmt.Sprintf(`
**%s**
**%s (%s)**
 • %s: %.1f
 • %s: %.1f
 • %s: %.1f
 • %s: %.1f
		`, artifact.Set, artifact.Slot, artifact.MainStat,
		artifact.SubStats[0].Stat, artifact.SubStats[0].Value,
		artifact.SubStats[1].Stat, artifact.SubStats[1].Value,
		artifact.SubStats[2].Stat, artifact.SubStats[2].Value,
		artifact.SubStats[3].Stat, artifact.SubStats[3].Value,
	)
}

func newAbyssChallenge() string {
	return fmt.Sprintf("%s %s but %s", rngx.Pick(allGenshinChars()), rngx.Pick(teamTypes), rngx.Pick(handicaps))
}

func allGenshinChars() []string {
	return []string{
		"Albedo",
		"Alhaitham",
		"Aloy",
		"Amber",
		"Arataki Itto",
		"Baizhu",
		"Barbara",
		"Beidou",
		"Bennett",
		"Candace",
		"Chongyun",
		"Collei",
		"Cyno",
		"Dehya",
		"Diluc",
		"Diona",
		"Dori",
		"Eula",
		"Faruzan",
		"Fischl",
		"Ganyu",
		"Gorou",
		"Hu Tao",
		"Jean",
		"Kaedehara Kazuha",
		"Kaeya",
		"Kamisato Ayaka",
		"Kamisato Ayato",
		"Kaveh",
		"Keqing",
		"Klee",
		"Kujou Sara",
		"Kuki Shinobu",
		"Layla",
		"Lisa",
		"Mika",
		"Mona",
		"Nahida",
		"Nilou",
		"Ningguang",
		"Noelle",
		"Qiqi",
		"Raiden Shogun",
		"Razor",
		"Rosaria",
		"Sangonomiya Kokomi",
		"Sayu",
		"Shenhe",
		"Shikanoin Heizou",
		"Sucrose",
		"Tartaglia",
		"Thoma",
		"Tighnari",
		"Traveler",
		"Venti",
		"Wanderer",
		"Xiangling",
		"Xiao",
		"Xingqiu",
		"Xinyan",
		"Yae Miko",
		"Yanfei",
		"Yaoyao",
		"Yelan",
		"Yoimiya",
		"Yun Jin",
		"Zhongli",
	}
}

var teamTypes = []string{
	"Hypercarry",
	"Vape",
	"Melt",
	"Freeze",
	"Taser driver",
	"National",
	"Hyperbloom",
	"Burgeon",
	"Quicken",
	"Physical",
	"Monoelement",
}

var handicaps = []string{
	"no healer/shielders (the other 3 chars)",
	"no 5* or BP weapons (whole team)",
	"no ER (<110%) (whole team)",
	"no resets",
	"only 3 characters",
	"only 4* characters (the other 3 chars)",
	"only 5* characters (the other 3 chars)",
	"only males (the other 3 chars)",
	"only females (the other 3 chars)",
	"only 2 artifacts (whole team)",
}
