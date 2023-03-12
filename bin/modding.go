package main

import (
	"fmt"
	"time"
)

type UserWarning struct {
	ID         int       `db:"UserWarning"`
	UserID     string    `db:"DiscordUserID"`
	WarnedByID string    `db:"WarnedByID"`
	GuildID    string    `db:"GuildID"`
	Reason     string    `db:"Reason"`
	CreatedAt  time.Time `db:"CreatedAt"`
}

func (u UserWarning) ShortString() string {
	return fmt.Sprintf("By <@%s> at <t:%d>, reason: '%s'", u.WarnedByID, u.CreatedAt.Unix(), u.Reason)
}
