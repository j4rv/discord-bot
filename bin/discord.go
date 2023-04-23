package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// util functions for discordgo stuff

func removeRoleAfterDuration(ds *discordgo.Session, guildID string, memberID string, roleID string, duration time.Duration) {
	go func() {
		time.Sleep(duration)
		ds.GuildMemberRoleRemove(guildID, memberID, roleID)
	}()
}

func findRole(roleID string, roles []*discordgo.Role) *discordgo.Role {
	for _, r := range roles {
		if r.ID == roleID {
			return r
		}
	}
	return nil
}
