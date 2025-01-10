package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-set/v3"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
)

func HandleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CMDArgsMap) {
	channel := GetChannel(i, opts)
	notifyAdd := true
	notifyRem := false

	if val, exists := opts["announce_code_additions"]; exists {
		notifyAdd = val.BoolValue()
	}
	if val, exists := opts["announce_code_removals"]; exists {
		notifyRem = val.BoolValue()
	}

	err := db.CreateSubscription(channel, notifyAdd, notifyRem)
	if err == nil {
		RespondPrivate(s, i, fmt.Sprintf("Successfully subscribed <#%v>!", channel))
		return
	}

	// duplicate? update instead
	if db.IsDuplicateKey(err) {
		err = db.UpdateSubscription(channel, notifyAdd, notifyRem)
		if err != nil {
			RespondPrivate(s, i, fmt.Sprintf("Error updating existing subscription for <#%v>: %v", channel, err))
			return
		} else {
			RespondPrivate(s, i, fmt.Sprintf("Updated existing subscription for <#%v> with provided settings (default otherwise)", channel))
			return
		}
	}

	// unknown error
	RespondPrivate(s, i, fmt.Sprintf("Error trying to create subscription for <#%v>: %v", channel, err))
}

func HandleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CMDArgsMap) {
	channel := GetChannel(i, opts)

	err := db.RemoveSubscription(channel)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error trying to unsubscribe: <#%v>", err))
	} else {
		RespondPrivate(s, i, "Successfully unsubscribed channel!")
	}
}

func HandleFilterGames(s *discordgo.Session, i *discordgo.InteractionCreate, opts CMDArgsMap) {
	channelID := GetChannel(i, opts)

	games := set.New[string](4)
	if val, exists := opts["game_1"]; exists {
		games.Insert(val.StringValue())
	}
	if val, exists := opts["game_2"]; exists {
		games.Insert(val.StringValue())
	}
	if val, exists := opts["game_3"]; exists {
		games.Insert(val.StringValue())
	}
	if val, exists := opts["game_4"]; exists {
		games.Insert(val.StringValue())
	}

	err := db.SetGameFilters(channelID, games)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error setting game filters for <#%v>: %v", channelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully set game filters for <#%v>!", channelID))
}