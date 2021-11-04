package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

const dailyCheckInReminderCRON = "0 18 * * *"
const dailyCheckInReminderMessage = "Remember to do the Daily Check-In! https://webstatic-sea.mihoyo.com/ys/event/signin-sea/index.html?act_id=e202102251931481"
const genshinTeamSize = 4

func initGenshinServices(ds *discordgo.Session) {
	dailyCheckInCRON := cron.New()
	_, err := dailyCheckInCRON.AddFunc(dailyCheckInReminderCRON, func() {
		for _, userID := range genshinDS.AllDailyCheckInReminderUserIDs() {
			userMessageSend(userID, dailyCheckInReminderMessage, ds)
		}
	})
	if err != nil {
		fmt.Println("Error while configuring Genshin's daily check-in CRON:", err)
	} else {
		dailyCheckInCRON.Start()
	}
}

var usersWithParametricReminder = map[string]context.CancelFunc{}

func startParametricReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	userID := mc.Author.ID
	stopParametricReminder(ds, mc, ctx)
	cancellableCtx, cancel := context.WithCancel(ctx)
	usersWithParametricReminder[userID] = cancel
	runParametricReminder(ds, mc, cancellableCtx)
}

func stopParametricReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	userID := mc.Author.ID
	cancelExistingReminder, ok := usersWithParametricReminder[userID]
	if ok {
		cancelExistingReminder()
		delete(usersWithParametricReminder, userID)
	}
	return ok
}

// could be better with a CRON library? ... ¯\_(ツ)_/¯
func runParametricReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	userID := mc.Author.ID
	select {
	case <-time.After(7 * 24 * time.Hour):
		_, err := userMessageSend(userID, "Remember to use the Parametric Transformer!", ds)
		if err != nil {
			return
		}
		runParametricReminder(ds, mc, ctx)
	case <-ctx.Done():
		fmt.Println("stopped ParametricReminder for user", userID)
	}
}

func startDailyCheckInReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	genshinDS.AddDailyCheckInReminder(mc.Author.ID)
}

func stopDailyCheckInReminder(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	genshinDS.RemoveDailyCheckInReminder(mc.Author.ID)
}

func randomAbyssLineup(chars ...string) (firstTeam, secondTeam [genshinTeamSize]string, replacements []string) {
	if len(chars) == 0 {
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
