package bot

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-set/v3"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

func HandleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	notifyAdd := true
	notifyRem := false

	if val, exists := opts["announce_code_additions"]; exists {
		notifyAdd = val.BoolValue()
	}
	if val, exists := opts["announce_code_removals"]; exists {
		notifyRem = val.BoolValue()
	}

	err := db.CreateSubscription(i.ChannelID, i.GuildID, notifyAdd, notifyRem)
	if err != nil {
		// duplicate? update instead
		if db.IsDuplicateErr(err) {
			err = db.UpdateSubscription(i.ChannelID, notifyAdd, notifyRem)
			if err != nil {
				RespondPrivate(s, i, fmt.Sprintf("Error updating existing subscription for <#%v>: %v", i.ChannelID, err))
				return
			} else {
				RespondPrivate(s, i, fmt.Sprintf("Resubscribed <#%v> with provided settings (default otherwise)!", i.ChannelID))
				return
			}
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error trying to create subscription for <#%v>: %v", i.ChannelID, err))
		return
	}

	RespondPrivate(s, i, fmt.Sprintf("Successfully subscribed <#%v>!", i.ChannelID))
}

func HandleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	// check if channel is subscribed
	if _, err := db.GetSubscription(i.ChannelID); err != nil {
		if err == sql.ErrNoRows {
			// RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", i.ChannelID))
			RespondPrivate(s, i, fmt.Sprintf("No subscription exists for <#%v>.", i.ChannelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", i.ChannelID, err))
		return
	} 

	// err := db.DeactivateSubscription(i.ChannelID)
	err := db.DeleteSubscription(i.ChannelID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error trying to unsubscribe: <#%v>", err))
	} else {
		RespondPrivate(s, i, fmt.Sprintf("Successfully unsubscribed <#%v>!", i.ChannelID))
	}
}

func HandleFilterGames(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	// check if channel is subscribed
	if _, err := db.GetSubscription(i.ChannelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", i.ChannelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", i.ChannelID, err))
		return
	} 

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

	err := db.SetGameFilters(i.ChannelID, games)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error setting game filters for <#%v>: %v", i.ChannelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully set game filters for <#%v>!", i.ChannelID))
}

func HandleAddPingRole(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	// check if channel is subscribed
	if _, err := db.GetSubscription(i.ChannelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", i.ChannelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", i.ChannelID, err))
		return
	} 

	roleID := opts["role"].RoleValue(nil, "").ID
	err := db.AddPingRole(i.ChannelID, roleID)
	if err != nil && !db.IsDuplicateErr(err) {
		RespondPrivate(s, i, fmt.Sprintf("Error adding ping role for <@&%v> in <#%v>: %v", roleID, i.ChannelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully added ping role for <@&%v> in <#%v>!", roleID, i.ChannelID))
}

func HandleRemovePingRole(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	// check if channel is subscribed
	if _, err := db.GetSubscription(i.ChannelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", i.ChannelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", i.ChannelID, err))
		return
	} 

	roleID := opts["role"].RoleValue(nil, "").ID
	err := db.RemovePingRole(i.ChannelID, roleID)
	if err != nil  {
		RespondPrivate(s, i, fmt.Sprintf("Error removing ping role <@&%v> from <#%v>: %v", roleID, i.ChannelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully removed ping role <@&%v> from <#%v>!", roleID, i.ChannelID))
}

func HandleCheckSubscription(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	if i.GuildID != "" { // don't run in DM environment
		if allChan := opts["all_channels"]; allChan != nil && allChan.BoolValue() {
			// get channels of server
			channels, err := db.GetGuildSubscriptions(i.GuildID)
			if err != nil {
				RespondPrivate(s, i, fmt.Sprintf("Error trying to get server channels: %v", err))
				return
			}
			
			result := fmt.Sprintf("**Subscriptions in server ID %v**\n", i.GuildID)
			for _, ch := range channels {
				result += getSubsPrint(&ch) + "\n\n"
			}
			RespondPrivate(s, i, strings.TrimRight(result, " \n\t"))
			return
		}
	}

	info, err := db.GetSubscription(i.ChannelID)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("No data available for <#%v>!", i.ChannelID))
			return
		}
		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", i.ChannelID, err))
		return 
	} 

	RespondPrivate(s, i, getSubsPrint(info))
}