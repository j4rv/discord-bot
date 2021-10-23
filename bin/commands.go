package main

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type command func(*discordgo.Session, *discordgo.MessageCreate, context.Context)

var commands = map[string]command{
	"!help":                      answerHelp,
	"!hello":                     answerHello,
	"!ruben":                     answerRuben,
	"!pablo":                     answerPablo,
	"!drive":                     answerDrive,
	"!parametricTransformer":     answerParametricTransformer,
	"!parametricTransformerStop": answerParametricTransformerStop,
}

const helpResponse = `Available commands:
- **!parametricTransformer**: Will remind you to use the Parametric Transformer every 7 days
- **!parametricTransformerStop**: The bot will stop reminding you to use the Parametric Transformer
- **!pablo**: jijiji
- **!ruben**: jijiji
`

func answerHelp(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, helpResponse)
}

func answerHello(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "Hello!!")
}

func answerDrive(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "[J4RV's Shared](https://drive.google.com/drive/folders/1JHlnWqoevIpZCHG4EdjQZN9vqJC0O8wA)")
}

func answerRuben(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "carbo")
}

func answerPablo(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ds.ChannelMessageSend(mc.ChannelID, "gafas")
}

func answerParametricTransformer(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	_, err := userMessageSend(mc.Author.ID, "I will remind you about the Parametric Transformer in 7 days!", ds, mc)
	if err != nil {
		return
	}

	ds.ChannelMessageSend(mc.ChannelID, "Gotcha!")
	startParametricReminder(mc.Author.ID, ds, mc, ctx)
}

func answerParametricTransformerStop(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) {
	ok := stopParametricReminder(mc.Author.ID, ds, mc, ctx)
	if ok {
		ds.ChannelMessageSend(mc.ChannelID, "I'll stop reminding you "+mc.Author.Mention())
	} else {
		ds.ChannelMessageSend(mc.ChannelID, "You weren't being reminded already "+mc.Author.Mention()+" but ok")
	}
}
