package main

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type buttonReducer func(*discordgo.Session, *discordgo.InteractionCreate, []string) error

var buttonReducerMap = make(map[string]buttonReducer)

func buttonCustomIdReducer(ds *discordgo.Session, ic *discordgo.InteractionCreate, id string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in processCommand: %s\n%s", r, string(debug.Stack()))
		}
	}()

	parts := strings.Split(id, ";")
	reducerId := parts[0]
	reducer, ok := buttonReducerMap[reducerId]
	if !ok {
		err := ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown or expired button interaction with ID: " + reducerId,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return err
		}
		return fmt.Errorf("could not find a button reducer for %s", reducerId)
	}
	log.Println("Executing button reducer", id)
	return reducer(ds, ic, parts)
}

func onInteractionCreate(ctx context.Context) func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	return func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
		if ic.Type != discordgo.InteractionMessageComponent {
			return
		}

		customID := ic.MessageComponentData().CustomID
		err := buttonCustomIdReducer(ds, ic, customID)
		if err != nil {
			log.Println("Interaction failed for customID", customID, err)
		}
	}
}

// Utils

func editInteractionMessage(ds *discordgo.Session, ic *discordgo.InteractionCreate, content string, buttons []*discordgo.Button) {
	_, err := ds.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         ic.Message.ID,
		Channel:    ic.ChannelID,
		Content:    &content,
		Components: buildButtonComponents(buttons),
	})
	if err != nil {
		log.Println("ChannelMessageEditComplex error:", err)
	}
}

func ackInteraction(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate})
}

func newButtonWithEnabled(label string, style discordgo.ButtonStyle, customID string, enabled bool) *discordgo.Button {
	return &discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: customID,
		Disabled: !enabled,
	}
}

func newButton(label string, style discordgo.ButtonStyle, customID string) *discordgo.Button {
	return &discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: customID,
	}
}

func buildButtonComponents(buttons []*discordgo.Button) *[]discordgo.MessageComponent {
	var components []discordgo.MessageComponent
	buttonsPerRow := 5
	maxRows := 5

	for row := range maxRows {
		start := row * buttonsPerRow
		if start >= len(buttons) {
			break
		}

		end := start + buttonsPerRow
		if end > len(buttons) {
			end = len(buttons)
		}

		rowButtons := make([]discordgo.MessageComponent, end-start)
		for i := start; i < end; i++ {
			rowButtons[i-start] = buttons[i]
		}

		components = append(components, discordgo.ActionsRow{
			Components: rowButtons,
		})
	}

	return &components
}
