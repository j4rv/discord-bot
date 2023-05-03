package main

import "time"

// TODO make this a config file

const dbFilename = "db.sqlite"

var strongboxMinAmount = 1.0
var strongboxMaxAmount = 64.0
var warnMessageMinLength = 1
var warnMessageMaxLength = 320

const avatarTargetSize = "1024"

const timeoutRoleName = "Shadow Realm"
const shootMisfireChance = 0.2
const nuclearCatastropheChance = 0.006
const timeoutDurationWhenShot = 4 * time.Minute
const timeoutDurationWhenMisfire = 8 * time.Minute
const timeoutDurationWhenNuclearCatastrophe = 2 * time.Minute

const dailyCheckInReminderCRON = "CRON_TZ=Asia/Shanghai 0 0 * * *"
const dailyCheckInReminderMessage = `Remember to do the Daily Check-In!

Genshin: https://webstatic-sea.mihoyo.com/ys/event/signin-sea/index.html?act_id=e202102251931481

Star Rail: https://act.hoyolab.com/bbs/event/signin/hkrpg/index.html?act_id=e202303301540311`
const parametricReminderCRON = "0 * * * *"
const parametricReminderMessage = "Remember to use the Parametric Transformer!\nI will remind you again in 7 days."
const playStoreReminderCRON = "0 * * * *"
const playStoreReminderMessage = "Remember to get the weekly Play Store prize!\nI will remind you again in 7 days."
const react4RolesCRON = "0 0 * * 6"

// Messages
const userMustBeAdminMessage = "Only the bot's admin can do that"
const userMustBeModMessage = "Only a mod can do that"
const commandReceivedMessage = "Gotcha!"
const commandSuccessMessage = "Successfully donette!"
const commandWithTwoArgumentsError = "Something went wrong, please make sure to use the command with the following format: '!command (...) (...)'"
const commandWithMentionError = "Something went wrong, please make sure that the command has an user mention"
