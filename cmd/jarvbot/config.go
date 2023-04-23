package main

import "time"

// TODO make this a config file

const timeoutRoleName = "Shadow Realm"
const shootMisfireChance = 0.2
const nuclearCatastropheChance = 0.006
const timeoutDurationWhenShot = 2 * time.Minute
const timeoutDurationWhenMisfire = 5 * time.Minute
const timeoutDurationWhenNuclearCatastrophe = 30 * time.Second

const dailyCheckInReminderCRON = "CRON_TZ=Asia/Shanghai 0 0 * * *"
const dailyCheckInReminderMessage = "Remember to do the Daily Check-In! https://webstatic-sea.mihoyo.com/ys/event/signin-sea/index.html?act_id=e202102251931481"
const parametricReminderCRON = "0 * * * *"
const parametricReminderMessage = "Remember to use the Parametric Transformer!\nI will remind you again in 7 days."
const playStoreReminderCRON = "0 * * * *"
const playStoreReminderMessage = "Remember to get the weekly Play Store prize!\nI will remind you again in 7 days."

//const react4RolesCRON = "0 * * * 0"
const react4RolesCRON = "* * * * *"
