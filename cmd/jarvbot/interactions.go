package main

import (
	"context"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/golang-lru/simplelru"
)

var interactionCache, _ = simplelru.NewLRU(10000, nil)

type cachedFuncData map[int]any

type cachedFunc struct {
	fn     func(*discordgo.Session, *discordgo.InteractionCreate)
	author string
	data   cachedFuncData
}

func onInteractionCreate(ctx context.Context) func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
	return func(ds *discordgo.Session, ic *discordgo.InteractionCreate) {
		if ic.Type != discordgo.InteractionMessageComponent {
			return
		}

		customID := ic.MessageComponentData().CustomID
		userID := interactionUser(ic).ID
		if x, ok := interactionCache.Get(customID); ok {
			entry := x.(cachedFunc)
			if entry.author != userID {
				ds.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "You can't do that!",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}
			entry.fn(ds, ic)
		}
	}
}

// Utils

func sendMessageWithButtons(ds *discordgo.Session, channelId, content string, buttons []discordgo.Button) (*discordgo.Message, error) {
	return ds.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Content: content,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: func() []discordgo.MessageComponent {
					comps := make([]discordgo.MessageComponent, len(buttons))
					for i, b := range buttons {
						comps[i] = b
					}
					return comps
				}(),
			},
		},
	})
}

func AddButtonHandler(
	customID string,
	authorID string,
	data cachedFuncData, // pass the full data map
	newContentFn func(data cachedFuncData) string,
	buttons []discordgo.Button,
) {
	interactionCache.Add(customID, cachedFunc{
		fn: func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
			log.Printf("Button %s clicked\n", customID)

			// Respond to interaction to avoid "interaction failed"
			err := s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			if err != nil {
				log.Println("InteractionRespond error:", err)
				return
			}

			// Compute the new content dynamically from the data
			content := newContentFn(data)

			// Build components once
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: make([]discordgo.MessageComponent, len(buttons)),
				},
			}
			for i, b := range buttons {
				components[0].(discordgo.ActionsRow).Components[i] = b
			}

			// Edit the original message
			originalMessageID := data[interactionDataOriginalMessageId].(string)
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				ID:         originalMessageID,
				Channel:    ic.ChannelID,
				Content:    &content,
				Components: &components,
			})
			if err != nil {
				log.Println("ChannelMessageEditComplex error:", err)
			}
		},
		author: authorID,
		data:   data,
	})
}
