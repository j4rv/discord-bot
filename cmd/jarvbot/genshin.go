package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/genshinchargen"
	artis "github.com/j4rv/genshinartis"
	"github.com/j4rv/rollssim"
)

const gachaPullChanceIterations = 2000
const averagePullsNote = "**Note:** The average rolls spent on each banner include successful attempts and failed attempts. This includes the best case scenarios of not needing to spend all your pulls to get your desired characters/weapons, and the worst case scenarios of spending all your pulls and not getting your desired characters/weapons."

// Command Answers

func answerParametricTransformer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := genshinDS.addOrUpdateParametricReminder(mc.Author.ID)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the Parametric Transformer in 7 days!")
	}
	return err == nil
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := genshinDS.removeParametricReminder(mc.Author.ID)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
	return err == nil
}

func answerPlayStore(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := genshinDS.addOrUpdatePlayStoreReminder(mc.Author.ID)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "I will remind you about the PlayStore in 7 days!")
	}
	return err == nil
}

func answerPlayStoreStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := genshinDS.removePlayStoreReminder(mc.Author.ID)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
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
	err := genshinDS.addDailyCheckInReminder(mc.Author.ID)
	if err == nil {
		_, err = ds.ChannelMessageSend(mc.ChannelID, commandReceivedMessage)
	}
	return err == nil
}

func answerGenshinDailyCheckInStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := genshinDS.removeDailyCheckInReminder(mc.Author.ID)
	if err == nil {
		ds.ChannelMessageSend(mc.ChannelID, "Ok, I'll stop reminding you")
	}
	return err == nil
}

// Slash Command answers

func answerGenshinChance(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	defer func() {
		if r := recover(); r != nil {
			adminNotifyIfErr("answerGenshinChance panic", fmt.Errorf("%v", r), ds)
		}
	}()
	options := optionMap(ic.ApplicationCommandData().Options)
	rollCount := int(options["roll_count"].IntValue())
	charCount := int(options["char_count"].IntValue())
	weaponCount := int(options["weapon_count"].IntValue())
	charPity := optionIntValueOrZero(options["char_pity"])
	charGuaranteed := optionBoolValueOrFalse(options["guaranteed_sr_char"])
	charRarePity := optionIntValueOrZero(options["char_rare_pity"])
	charRareGuaranteed := optionBoolValueOrFalse(options["guaranteed_rare_char"])
	weaponPity := optionIntValueOrZero(options["weapon_pity"])
	weaponGuaranteed := optionBoolValueOrFalse(options["guaranteed_sr_weapon"])
	weaponRarePity := optionIntValueOrZero(options["weapon_rare_pity"])
	weaponRareGuaranteed := optionBoolValueOrFalse(options["guaranteed_rare_weapon"])

	cumResult := rollssim.WantedRollsResult{}
	successCount := 0

	for i := 0; i < gachaPullChanceIterations; i++ {
		charBanner := rollssim.GenshinCharRoller{
			MihoyoRoller: rollssim.MihoyoRoller{
				CurrSRPity:           charPity,
				GuaranteedRateUpSR:   charGuaranteed,
				CurrRarePity:         charRarePity,
				GuaranteedRateUpRare: charRareGuaranteed,
			},
		}
		weaponBanner := rollssim.GenshinWeaponRoller{
			MihoyoRoller: rollssim.MihoyoRoller{
				CurrSRPity:           weaponPity,
				GuaranteedRateUpSR:   weaponGuaranteed,
				CurrRarePity:         weaponRarePity,
				GuaranteedRateUpRare: weaponRareGuaranteed,
			},
			FatePoints: 0,
		}
		result := rollssim.CalcGenshinWantedRolls(rollCount, charCount, weaponCount, &charBanner, &weaponBanner)
		cumResult.Add(result)

		if result.CharacterBannerRateUpSRCount >= charCount && result.WeaponBannerChosenRateUpCount >= weaponCount {
			successCount++
		}
	}

	_, err := textRespond(ds, ic, formatGenshinChanceResult(cumResult, successCount))
	serverNotifyIfErr("answerGenshinChance", err, ic.GuildID, ds)
}

func formatGenshinChanceResult(result rollssim.WantedRollsResult, successCount int) string {
	formatted := fmt.Sprintf("## Chance of Success: %.2f%%\n",
		divideToFloat(successCount, gachaPullChanceIterations)*100,
	)

	if result.CharacterBannerRollCount > 0 {
		formatted += fmt.Sprintf(`### Character banner:
	Average Rate-Up 5\* Characters: %.2f
	Average Standard 5\* Characters: %.2f
	Average Rate-Up 4\* Characters: %.2f
	Average Standard 4\*s: %.2f
	Average Wishes spent on Character banner: %.2f
	
`,
			divideToFloat(result.CharacterBannerRateUpSRCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerStdSRCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerRateUpRareCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerStdRareCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerRollCount, gachaPullChanceIterations),
		)
	}

	if result.WeaponBannerRollCount > 0 {
		formatted += fmt.Sprintf(`### Weapon banner:
	Average Chosen Rate-Up 5\* Weapons: %.2f
	Average Non-Chosen Rate-Up 5\* Weapons: %.2f
	Average Standard 5\* Weapons: %.2f
	Average Rate-Up 4\* Weapons: %.2f
	Average Standard 4\*s: %.2f
	Average Wishes spent on Weapon banner: %.2f
	
`,
			divideToFloat(result.WeaponBannerChosenRateUpCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerNotChosenRateUpCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerStdSRCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerRateUpRareCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerStdRareCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerRollCount, gachaPullChanceIterations),
		)
	}

	formatted += averagePullsNote

	return formatted
}

func answerStarRailChance(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	defer func() {
		if r := recover(); r != nil {
			adminNotifyIfErr("answerStarRailChance panic", fmt.Errorf("%v", r), ds)
		}
	}()
	options := optionMap(ic.ApplicationCommandData().Options)
	rollCount := int(options["roll_count"].IntValue())
	charCount := int(options["char_count"].IntValue())
	weaponCount := int(options["weapon_count"].IntValue())
	charPity := optionIntValueOrZero(options["char_pity"])
	charGuaranteed := optionBoolValueOrFalse(options["guaranteed_sr_char"])
	charRarePity := optionIntValueOrZero(options["char_rare_pity"])
	charRareGuaranteed := optionBoolValueOrFalse(options["guaranteed_rare_char"])
	weaponPity := optionIntValueOrZero(options["weapon_pity"])
	weaponGuaranteed := optionBoolValueOrFalse(options["guaranteed_sr_weapon"])
	weaponRarePity := optionIntValueOrZero(options["weapon_rare_pity"])
	weaponRareGuaranteed := optionBoolValueOrFalse(options["guaranteed_rare_weapon"])

	cumResult := rollssim.WantedRollsResult{}
	successCount := 0

	for i := 0; i < gachaPullChanceIterations; i++ {
		charBanner := rollssim.StarRailCharRoller{
			MihoyoRoller: rollssim.MihoyoRoller{
				CurrSRPity:           charPity,
				GuaranteedRateUpSR:   charGuaranteed,
				CurrRarePity:         charRarePity,
				GuaranteedRateUpRare: charRareGuaranteed,
			},
		}
		weaponBanner := rollssim.StarRailLCRoller{
			MihoyoRoller: rollssim.MihoyoRoller{
				CurrSRPity:           weaponPity,
				GuaranteedRateUpSR:   weaponGuaranteed,
				CurrRarePity:         weaponRarePity,
				GuaranteedRateUpRare: weaponRareGuaranteed,
			},
		}
		result := rollssim.CalcStarRailWantedRolls(rollCount, charCount, weaponCount, &charBanner, &weaponBanner)
		cumResult.Add(result)

		if result.CharacterBannerRateUpSRCount >= charCount && result.WeaponBannerRateUpSRCount >= weaponCount {
			successCount++
		}
	}

	_, err := textRespond(ds, ic, formatStarRailChanceResult(cumResult, successCount))
	serverNotifyIfErr("answerStarRailChance", err, ic.GuildID, ds)
}

func formatStarRailChanceResult(result rollssim.WantedRollsResult, successCount int) string {
	formatted := fmt.Sprintf("## Chance of Success: %.2f%%\n",
		divideToFloat(successCount, gachaPullChanceIterations)*100,
	)

	if result.CharacterBannerRollCount > 0 {
		formatted += fmt.Sprintf(`### Character banner:
	Average Rate-Up 5\* Characters: %.2f
	Average Standard 5\* Characters: %.2f
	Average Rate-Up 4\* Characters: %.2f
	Average Standard 4\*s: %.2f
	Average Warps spent on Character banner: %.2f
	
`,
			divideToFloat(result.CharacterBannerRateUpSRCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerStdSRCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerRateUpRareCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerStdRareCount, gachaPullChanceIterations),
			divideToFloat(result.CharacterBannerRollCount, gachaPullChanceIterations),
		)
	}

	if result.WeaponBannerRollCount > 0 {
		formatted += fmt.Sprintf(`### Light Cone banner:
	Average Rate-Up 5\* Light Cones: %.2f
	Average Standard 5\* Light Cones: %.2f
	Average Rate-Up 4\* Light Cones: %.2f
	Average Standard 4\*s: %.2f
	Average Warps spent on Light Cone banner: %.2f
	
`,
			divideToFloat(result.WeaponBannerRateUpSRCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerStdSRCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerRateUpRareCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerStdRareCount, gachaPullChanceIterations),
			divideToFloat(result.WeaponBannerRollCount, gachaPullChanceIterations),
		)
	}

	formatted += averagePullsNote

	return formatted
}

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
		serverNotifyIfErr("answerStrongbox_jsonMarshal", err, ic.GuildID, ds)
		textRespond(ds, ic, "Oops, error")
		return
	}

	interactionFileRespond(ds, ic, message, "StrongboxResult.json", string(good))
}

func answerCharacter(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	name := ic.ApplicationCommandData().Options[0].StringValue()
	textRespond(ds, ic, genshinchargen.NewChar(name, unixDay()).PrettyString())
}

// CRONs

func dailyCheckInCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allDailyCheckInReminderUserIDs()
		adminNotifyIfErr("allDailyCheckInReminderUserIDs", err, ds)
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
		adminNotifyIfErr("allParametricReminderUserIDsToBeReminded", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to use the Parametric Transformer", len(userIDs))
			for _, userID := range userIDs {
				sendDirectMessage(userID, parametricReminderMessage, ds)
				err := genshinDS.addOrUpdateParametricReminder(userID)
				adminNotifyIfErr("addOrUpdateParametricReminder", err, ds)
			}
		}
	}
}

func playStoreCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allPlayStoreReminderUserIDsToBeReminded()
		adminNotifyIfErr("allPlayStoreReminderUserIDsToBeReminded", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to get the Play Store prize", len(userIDs))
			for _, userID := range userIDs {
				sendDirectMessage(userID, playStoreReminderMessage, ds)
				err := genshinDS.addOrUpdatePlayStoreReminder(userID)
				adminNotifyIfErr("addOrUpdatePlayStoreReminder", err, ds)
			}
		}
	}
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
