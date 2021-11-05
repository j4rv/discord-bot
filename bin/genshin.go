package main

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

const dailyCheckInReminderCRON = "0 18 * * *"
const dailyCheckInReminderMessage = "Remember to do the Daily Check-In! https://webstatic-sea.mihoyo.com/ys/event/signin-sea/index.html?act_id=e202102251931481"
const parametricReminderCRON = "0 * * * *"
const parametricReminderMessage = "Remember to use the Parametric Transformer!"
const genshinTeamSize = 4

func initGenshinServices(ds *discordgo.Session) {
	dailyCheckInCRON := cron.New()
	_, err := dailyCheckInCRON.AddFunc(dailyCheckInReminderCRON, dailyCheckInCRONFunc(ds))
	if err != nil {
		checkErr("AddFunc to dailyCheckInCRON", err, ds)
	} else {
		dailyCheckInCRON.Start()
	}

	parametricCRON := cron.New()
	_, err = parametricCRON.AddFunc(parametricReminderCRON, parametricCRONFunc(ds))
	if err != nil {
		checkErr("AddFunc to parametricCRON", err, ds)
	} else {
		parametricCRON.Start()
	}
}

func dailyCheckInCRONFunc(ds *discordgo.Session) func() {
	return func() {
		userIDs, err := genshinDS.allDailyCheckInReminderUserIDs()
		checkErr("allDailyCheckInReminderUserIDs", err, ds)
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
		checkErr("allParametricReminderUserIDsToBeReminded", err, ds)
		if len(userIDs) > 0 {
			log.Printf("Reminding %d users to use the Parametric Transformer", len(userIDs))
			for _, userID := range userIDs {
				userMessageSend(userID, parametricReminderMessage, ds)
				err := genshinDS.addOrUpdateParametricReminder(userID)
				checkErr("addOrUpdateParametricReminder", err, ds)
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

func allGenshinChars() []string {
	return []string{
		"Albedo",
		"Aloy",
		"Amber",
		"Barbara",
		"Beidou",
		"Bennett",
		"Chongyun",
		"Diluc",
		"Diona",
		"Eula",
		"Fischl",
		"Ganyu",
		"Hu Tao",
		"Jean",
		"Kaeya",
		"Kaedehara Kazuha",
		"Kamisato Ayaka",
		"Keqing",
		"Klee",
		"Kujou Sara",
		"Lisa",
		"Mona",
		"Ningguang",
		"Noelle",
		"Qiqi",
		"Raiden Shogun",
		"Razor",
		"Rosaria",
		"Sangonomiya Kokomi",
		"Sayu",
		"Sucrose",
		"Tartaglia",
		"Thoma",
		"Traveler",
		"Venti",
		"Xiangling",
		"Xiao",
		"Xingqiu",
		"Xinyan",
		"Yanfei",
		"Yoimiya",
		"Zhongli",
	}
}
