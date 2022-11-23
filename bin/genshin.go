package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/genshinartis"
	"github.com/robfig/cron/v3"
)

const dailyCheckInReminderCRON = "CRON_TZ=Asia/Shanghai 0 0 * * *"
const dailyCheckInReminderMessage = "Remember to do the Daily Check-In! https://webstatic-sea.mihoyo.com/ys/event/signin-sea/index.html?act_id=e202102251931481"
const parametricReminderCRON = "0 * * * *"
const parametricReminderMessage = "Remember to use the Parametric Transformer!\nI will remind you again in 7 days."
const playStoreReminderCRON = "0 * * * *"
const playStoreReminderMessage = "Remember to get the weekly Play Store prize!\nI will remind you again in 7 days."
const genshinTeamSize = 4

func initGenshinCRONs(ds *discordgo.Session) {
	dailyCheckInCRON := cron.New()
	_, err := dailyCheckInCRON.AddFunc(dailyCheckInReminderCRON, dailyCheckInCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to dailyCheckInCRON", err, ds)
	} else {
		dailyCheckInCRON.Start()
	}

	parametricCRON := cron.New()
	_, err = parametricCRON.AddFunc(parametricReminderCRON, parametricCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to parametricCRON", err, ds)
	} else {
		parametricCRON.Start()
	}

	playStoreCRON := cron.New()
	_, err = playStoreCRON.AddFunc(playStoreReminderCRON, playStoreCRONFunc(ds))
	if err != nil {
		notifyIfErr("AddFunc to playStoreCRON", err, ds)
	} else {
		playStoreCRON.Start()
	}
}

func dailyCheckInCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allDailyCheckInReminderUserIDs()
		notifyIfErr("allDailyCheckInReminderUserIDs", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to do the Daily CheckIn", len(userIDs))
			for _, userID := range userIDs {
				userMessageSend(userID, dailyCheckInReminderMessage, ds)
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
				userMessageSend(userID, parametricReminderMessage, ds)
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
				userMessageSend(userID, playStoreReminderMessage, ds)
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
		firstTeam[i] = extractRandomStringFromSlice(&chars)
		secondTeam[i] = extractRandomStringFromSlice(&chars)
	}

	if len(chars) < genshinTeamSize {
		replacements = chars
	} else {
		for i := 0; i < genshinTeamSize; i++ {
			replacements = append(replacements, extractRandomStringFromSlice(&chars))
		}
	}

	return firstTeam, secondTeam, replacements
}

func formatGenshinArtifact(artifact *genshinartis.Artifact) string {
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

func allGenshinChars() []string {
	return []string{
		"Albedo",
		"Aloy",
		"Amber",
		"Arataki Itto",
		"Barbara",
		"Beidou",
		"Bennett",
		"Candace",
		"Chongyun",
		"Collei",
		"Cyno",
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
		"Keqing",
		"Klee",
		"Kujou Sara",
		"Kuki Shinobu",
		"Layla",
		"Lisa",
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
		"Yelan",
		"Yoimiya",
		"Yun Jin",
		"Zhongli",
	}
}
