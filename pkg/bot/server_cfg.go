package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
)

func HandleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CMDArgsMap) {
	channel, _ := strconv.ParseUint(i.ChannelID, 10, 64)
	notifyAdd := true
	notifyRem := false

	if val, exists := opts["channel"]; exists {
		channel, _ = strconv.ParseUint(val.ChannelValue(nil).ID, 10, 64)
	}
	if val, exists := opts["announce_code_additions"]; exists {
		notifyAdd = val.BoolValue()
	}
	if val, exists := opts["announce_code_removals"]; exists {
		notifyRem = val.BoolValue()
	}

	err := db.CreateSubscription(channel, notifyAdd, notifyRem)
	if err == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully subscribed <#%v>!", channel),
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// duplicate? update instead
	if strings.Contains(err.Error(), "Error 1062 (23000): Duplicate entry") {
		err = db.UpdateSubscription(channel, notifyAdd, notifyRem)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Error updating existing subscription for <#%v>: %v", channel, err),
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			return
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Updated existing subscription for <#%v> with provided settings (default otherwise).", channel),
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
	}

	// unknown error
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Error trying to create subscription for <#%v>: %v", channel, err),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func HandleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CMDArgsMap) {
	channel, _ := strconv.ParseUint(i.ChannelID, 10, 64)
	if val, exists := opts["channel"]; exists {
		channel, _ = strconv.ParseUint(val.StringValue(), 10, 64)
	}
	err := db.RemoveSubscription(channel)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error trying to unsubscribe: <#%v>", err),
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Successfully unsubscribed channel!",
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func HandleFilterGames(channelID uint64, game []string) {

}