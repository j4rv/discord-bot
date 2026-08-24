package main

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/j4rv/discord-bot/pkg/uwu"
	"github.com/patrickmn/go-cache"
)

const bunkerServerID = "807055417120129085"
const bunkerGeneralChannelID = "828303425414365214"

var commandShootRegex = regexp.MustCompile(
	`^!shoot\s*(?:<@!?(\d+)>|(everyone|@everyone|here|@here))$`,
)

var shootEveryoneDodgeMessages = []string{
	"<user> dodged the bullets!",
	"<user> hid in time.",
	"<user> somehow survived the chaos.",
	"<user> got lucky.",
}

var shootEveryoneResponses = []string{
	"https://gif.fxtwitter.com/tweet_video/HMRq6rma4AAniOp.webp",
	"https://klipy.com/gifs/nichijou-misato-2",
}

// Command Answers

func answerLiquid(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	_, err := ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("%06d, you know what to do with this. ", rand.Intn(650000)))
	return err == nil
}

func answerDon(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	serverNotifyIfErr("answerDon, couldn't get timeoutRole", err, mc.GuildID, ds)
	if err != nil {
		return false
	}

	if isMemberInRole(mc.Member, timeoutRole.ID) {
		ds.ChannelMessageSend(mc.ChannelID, "Stay Realmed scum")
		return false
	}

	err = ds.GuildMemberRoleAdd(mc.GuildID, mc.Author.ID, timeoutRole.ID)
	serverNotifyIfErr("answerDon, couldn't add timeoutRole", err, mc.GuildID, ds)
	if err != nil {
		return false
	}
	removeShadowRealmRoleAfterDuration(mc.GuildID, mc.Author.ID, timeoutRole.ID, 10*time.Minute)
	_, err = ds.ChannelMessageSend(mc.ChannelID, fmt.Sprintf("To the Shadow Realm you go %s", mc.Author.Mention()))
	return err == nil
}

func answerShoot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	match := commandShootRegex.FindStringSubmatch(mc.Content)
	if match == nil {
		ds.ChannelMessageSend(mc.ChannelID, commandWithMentionError)
		return false
	}

	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	serverNotifyIfErr("answerShoot: get timeout role", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find the Timeout Role, maybe I'm missing permissions or it does not exist :(")
		return false
	}

	shooter, err := ds.GuildMember(mc.GuildID, mc.Author.ID)
	serverNotifyIfErr("answerShoot: get shooter member", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find you in this server, maybe I'm missing permissions u_u")
		return false
	}

	if isMemberInRole(shooter, timeoutRole.ID) {
		ds.ChannelMessageSend(mc.ChannelID, "Shadow Realmed people can't shoot dummy")
		return false
	}

	// Special case for @everyone/@here.
	if match[2] != "" {
		err = shootEveryone(ds, mc.ChannelID, mc.GuildID, shooter, timeoutRole.ID)
		serverNotifyIfErr("answerShoot: shoot everyone", err, mc.GuildID, ds)
		return err == nil
	}

	target, err := ds.GuildMember(mc.GuildID, match[1])
	serverNotifyIfErr("answerShoot: get target member", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Couldn't find member with user ID: "+match[1]+", maybe I'm missing permissions u_u")
		return false
	}

	err = shoot(ds, mc.ChannelID, mc.GuildID, shooter, target, timeoutRole.ID)
	serverNotifyIfErr("answerShoot: shoot", err, mc.GuildID, ds)
	return err == nil
}

func answerForceNuke(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	timeoutRole, err := getTimeoutRole(ds, mc.GuildID)
	serverNotifyIfErr("answerForceNuke: get timeout role", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Could not find the Timeout Role, maybe I'm missing permissions or it does not exist :(")
		return false
	}
	return handleNuke(ds, mc.ChannelID, mc.GuildID, timeoutRole.ID, nuclearCatastropheResponse) == nil
}

func answerSniperShoot(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	if mc.Author == nil {
		return false
	}

	targetID := strings.TrimSpace(commandPrefixRegex.ReplaceAllString(mc.Content, ""))
	if targetID == ds.State.User.ID {
		ds.ChannelMessageSend(mc.ChannelID, "https://tenor.com/view/anya-forger-anya-spy-x-family-gif-17200077238442027522")
		return false
	}

	target, err := ds.GuildMember(bunkerServerID, targetID)
	serverNotifyIfErr("answerShoot: get target member", err, mc.GuildID, ds)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "Couldn't find Bunker member with user ID: "+targetID)
		return false
	}

	timeoutRole, err := getTimeoutRole(ds, bunkerServerID)
	serverNotifyIfErr("answerSniperShoot: get timeout role", err, mc.GuildID, ds)
	if err != nil {
		return false
	}

	ds.ChannelMessageSend(bunkerGeneralChannelID, fmt.Sprintf("%s got sniped by %s!", target.User.Mention(), mc.Author.Mention()))
	ds.GuildMemberRoleAdd(bunkerServerID, target.User.ID, timeoutRole.ID)
	removeShadowRealmRoleAfterDuration(bunkerServerID, target.User.ID, timeoutRole.ID, timeoutDurationWhenShot)
	ds.ChannelMessageSend(mc.ChannelID, "https://tenor.com/view/gun-anime-sniper-scope-scoping-gif-17545837")
	return true
}

func getTimeoutRole(ds *discordgo.Session, guildID string) (*discordgo.Role, error) {
	customRoleName, err := serverDS.getServerProperty(guildID, serverPropCustomTimeoutRoleName)
	if err != nil {
		customRoleName = defaultTimeoutRoleName
	}
	return guildRoleByName(ds, guildID, customRoleName)
}

func getTimeoutRoleName(ds *discordgo.Session, guildID string) string {
	timeoutRole, err := getTimeoutRole(ds, guildID)
	if err != nil {
		return defaultTimeoutRoleName
	}
	return timeoutRole.Name
}

func setCustomTimeoutRole(ds *discordgo.Session, guildID string, roleName string) error {
	return serverDS.setServerProperty(guildID, serverPropCustomTimeoutRoleName, roleName)
}

// Internal functions

func shoot(ds *discordgo.Session, channelID string, guildID string, shooter *discordgo.Member, target *discordgo.Member, timeoutRoleID string) error {
	if isMemberInRole(target, timeoutRoleID) {
		ds.ChannelMessageSend(channelID, "https://giphy.com/gifs/the-simpsons-stop-hes-already-dead-JCAZQKoMefkoX6TyTb")
		return nil
	}

	isAF := isAprilFools()
	var nukeAFMultiplier float32 = 1
	var shootAFMultiplier float32 = 1
	if isAF {
		nukeAFMultiplier = 10
		shootAFMultiplier = 3
	}

	// Nuke logic
	if rand.Float32() <= nuclearCatastropheChance*nukeAFMultiplier {
		return handleNuke(ds, channelID, guildID, timeoutRoleID, nuclearCatastropheResponse)
	}

	// Crit shot
	if rand.Float32() <= shootCritChance*shootAFMultiplier {
		ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got shot!! Critical Hit!!", target.User.Mention()))
		err := ds.GuildMemberRoleAdd(guildID, target.User.ID, timeoutRoleID)
		if err == nil {
			removeShadowRealmRoleAfterDuration(guildID, target.User.ID, timeoutRoleID, timeoutDurationWhenCritShot)
		}
		return nil
	}

	// Miss logic
	if rand.Float32() <= shootMisfireChance*shootAFMultiplier || target.User.Bot {
		ds.ChannelMessageSend(channelID, "OOPS! You missed :3c")
		err := ds.GuildMemberRoleAdd(guildID, shooter.User.ID, timeoutRoleID)
		if err == nil {
			removeShadowRealmRoleAfterDuration(guildID, shooter.User.ID, timeoutRoleID, timeoutDurationWhenMisfire)
		}
		return nil
	}

	// Normal shot
	ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got shot!", target.User.Mention()))
	err := ds.GuildMemberRoleAdd(guildID, target.User.ID, timeoutRoleID)
	if err == nil {
		removeShadowRealmRoleAfterDuration(guildID, target.User.ID, timeoutRoleID, timeoutDurationWhenShot)
	}
	return nil
}

func shootEveryone(ds *discordgo.Session, channelID, guildID string, shooter *discordgo.Member, timeoutRoleID string) error {
	if value, _ := serverDS.getServerProperty(guildID, serverPropShadowFeatureMassShootings); value != serverPropYes {
		ds.ChannelMessageSend(channelID, "Mass shootings are not allowed in this server! :<")
		return nil
	}

	ds.ChannelMessageSend(channelID, shootEveryoneResponses[rand.Intn(len(shootEveryoneResponses))])

	activeUsers, err := activeChannelMembers(ds, channelID, false)
	if err != nil {
		return fmt.Errorf("could not fetch active users: %w", err)
	}

	if len(activeUsers) == 0 {
		return fmt.Errorf("no active users found in the channel")
	}

	for _, user := range activeUsers {
		member, err := ds.GuildMember(guildID, user.ID)
		if err != nil {
			continue
		}
		if member.User.ID == shooter.User.ID || isMemberInRole(member, timeoutRoleID) {
			continue
		}

		// dodge chance, same misfire chance
		if rand.Float32() <= shootMisfireChance {
			msg := shootEveryoneDodgeMessages[rand.Intn(len(shootEveryoneDodgeMessages))]
			ds.ChannelMessageSend(channelID, strings.ReplaceAll(msg, "<user>", user.Mention()))
			continue
		}

		ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got shot!", user.Mention()))
		if err := ds.GuildMemberRoleAdd(guildID, user.ID, timeoutRoleID); err == nil {
			removeShadowRealmRoleAfterDuration(guildID, user.ID, timeoutRoleID, timeoutDurationWhenShotEveryone)
		}
	}

	if err := ds.GuildMemberRoleAdd(guildID, shooter.User.ID, timeoutRoleID); err == nil {
		ds.ChannelMessageSend(channelID, fmt.Sprintf("%s got captured by the mod police!", shooter.Mention()))
		removeShadowRealmRoleAfterDuration(guildID, shooter.User.ID, timeoutRoleID, timeoutDurationWhenEveryoneShooter)
	}

	return nil
}

func handleNuke(ds *discordgo.Session, channelID, guildID, timeoutRoleID, firstResponse string) error {
	ds.ChannelMessageSend(channelID, firstResponse)

	activeUsers, err := activeChannelMembers(ds, channelID, false)
	if err != nil {
		return fmt.Errorf("could not fetch active users: %w", err)
	}

	if len(activeUsers) == 0 {
		return fmt.Errorf("no active users found in the channel")
	}

	deathCount := nuclearCatastropheMinDeaths + int(float64(len(activeUsers))*0.25)
	if deathCount > len(activeUsers) {
		deathCount = len(activeUsers)
	}
	if deathCount > nuclearCatastropheMaxDeaths {
		deathCount = nuclearCatastropheMaxDeaths
	}

	rand.Shuffle(len(activeUsers), func(i, j int) {
		activeUsers[i], activeUsers[j] = activeUsers[j], activeUsers[i]
	})
	dead := activeUsers[:deathCount]

	for _, user := range dead {
		ds.ChannelMessageSend(channelID, fmt.Sprintf("%s died in the explosion!", user.Mention()))
		if err := ds.GuildMemberRoleAdd(guildID, user.ID, timeoutRoleID); err == nil {
			removeShadowRealmRoleAfterDuration(guildID, user.ID, timeoutRoleID, timeoutDurationWhenNuclearCatastrophe)
		}
	}

	return nil
}

// uwu

var uwuifier = uwu.NewUwuifier()
var uwuJailedUsers = cache.New(cache.NoExpiration, 0)

func answerUwuJail(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	match := commandWithMention.FindStringSubmatch(mc.Content)
	if match == nil || len(match) != 2 {
		ds.ChannelMessageSend(mc.ChannelID, commandWithMentionError)
		return false
	}

	targetID := match[1]

	err := serverDS.setServerProperty(mc.GuildID, serverPropUwuJailedUser, targetID)
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "I could not uwu jail them :( sowwy")
		return false
	}

	uwuJailedUsers.Set(mc.GuildID, targetID, cache.NoExpiration)
	ds.ChannelMessageSend(mc.ChannelID, "UwU Jailed!! >:3c")
	return true
}

func answerUwuRelease(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	err := serverDS.setServerProperty(mc.GuildID, serverPropUwuJailedUser, "")
	if err != nil {
		ds.ChannelMessageSend(mc.ChannelID, "I could not uwu release them :( sowwy")
		return false
	}

	uwuJailedUsers.Delete(mc.GuildID)
	ds.ChannelMessageSend(mc.ChannelID, "The uwu prisoner has been released~")
	return true
}

func newMessageUwuCheck(ds *discordgo.Session, mc *discordgo.MessageCreate) bool {
	value, ok := uwuJailedUsers.Get(mc.GuildID)
	if !ok {
		uwufiedUserId, err := serverDS.getServerProperty(mc.GuildID, serverPropUwuJailedUser)
		if err != nil {
			return false
		}

		uwuJailedUsers.Set(mc.GuildID, uwufiedUserId, cache.NoExpiration)
		value = uwufiedUserId
	}

	uwufiedUserId := value.(string)
	if uwufiedUserId == "" || mc.Author.ID != uwufiedUserId {
		return false
	}

	_, err := sendAsUser(ds, mc.Author, mc.ChannelID, uwuifier.UwUify(mc.Content), mc.ReferencedMessage)
	if err != nil {
		serverNotifyIfErr("newMessageUwuCheck::sendAsUser", err, mc.GuildID, ds)
		return false
	}

	ds.State.MessageRemove(mc.Message)
	err = ds.ChannelMessageDelete(mc.ChannelID, mc.ID)
	if err != nil {
		serverNotifyIfErr("newMessageUwuCheck::ds.ChannelMessageDelete", err, mc.GuildID, ds)
		return false
	}

	return true
}
