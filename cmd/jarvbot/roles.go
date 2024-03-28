package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func findRoleInSlice(roleID string, roles []*discordgo.Role) *discordgo.Role {
	for _, r := range roles {
		if r.ID == roleID {
			return r
		}
	}
	return nil
}

func guildRoleByName(ds *discordgo.Session, guildID string, roleName string) (*discordgo.Role, error) {
	roles, err := ds.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.Name == roleName {
			return r, nil
		}
	}
	return nil, fmt.Errorf("role with name %s not found in guild with id %s", roleName, guildID)
}

func isMemberInRole(member *discordgo.Member, roleID string) bool {
	for _, r := range member.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}
