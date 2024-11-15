package main

import (
	"errors"
	"time"
)

// TODO make this a config file

var dbFilename = "db.sqlite"

var errNotFound = errors.New("not found")

var zero = 0.0
var strongboxMinAmount = 1.0
var strongboxMaxAmount = 1000.0
var warnMessageMinLength = 1
var warnMessageMaxLength = 320

const avatarTargetSize = "1024"

const cleanStateMessagesCRON = "0 * * * *"
const stateMessageMaxLifetime = 24 * time.Hour
const maxMessageCount = 100
const expensiveOperationCooldown = 15 * time.Second
const commandCooldown = time.Minute * 15

// https://discord.com/branding
const colorBlue = 0x5865F2
const colorYellow = 0xFEE75C
const colorRed = 0xED4245

const serverPropCustomTimeoutRoleName = "custom_timeout_role_name"
const serverPropAnnounceHere = "announce_here"
const serverPropMessageLogs = "message_logs"
const serverPropFixBadEmbedLinks = "fix_twitter_links"
const serverPropYes = "Y"
const serverPropNo = "N"

const defaultTimeoutRoleName = "Shadow Realm"
const shootMisfireChance = 0.2
const nuclearCatastropheChance = 0.006
const nuclearCatastropheDeathRatio = 0.015
const timeoutDurationWhenShot = 4 * time.Minute
const timeoutDurationWhenMisfire = 8 * time.Minute
const timeoutDurationWhenNuclearCatastrophe = 2 * time.Minute

const backupCRON = "0 0 * * 1"
const dailyCheckInReminderCRON = "CRON_TZ=Asia/Shanghai 0 0 * * *"
const dailyCheckInReminderMessage = `Remember to do the Daily Check-In!
- [Genshin Impact](https://webstatic-sea.mihoyo.com/ys/event/signin-sea/index.html?act_id=e202102251931481)
- [Honkai: Star Rail](https://act.hoyolab.com/bbs/event/signin/hkrpg/index.html?act_id=e202303301540311)
- [Zenless Zone Zero](https://act.hoyolab.com/bbs/event/signin/zzz/e202406031448091.html?act_id=e202406031448091)
Use !genshindailycheckinstop if you want to stop these reminders.`
const parametricReminderCRON = "0 * * * *"
const parametricReminderMessage = "Remember to use the Parametric Transformer!\nI will remind you again in 7 days.\nUse !parametrictransformerstop if you want to stop these reminders."
const playStoreReminderCRON = "0 * * * *"
const playStoreReminderMessage = "Remember to get the weekly Play Store prize!\nI will remind you again in 7 days.\nUse !playstorestop if you want to stop these reminders."
const react4RolesCRON = "0 0 * * 6"

// Messages
const userMustBeAdminMessage = "Only the bot's admin can do that"
const userMustBeModMessage = "Only a mod can do that"
const notAGuildMessage = "This command can only be used on a server"
const commandReceivedMessage = "Gotcha!"
const commandSuccessMessage = "Successfully donette!"
const commandWithTwoArgumentsError = "Something went wrong, please make sure to use the command with the following format: '!command (...) (...)'"
const commandWithMentionError = "Something went wrong, please make sure that the command has a user mention"
const expensiveOperationErrorMsg = "You just executed an expensive operation, please wait a bit u_u"
