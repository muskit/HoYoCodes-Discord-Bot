package bot

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/go-set/v3"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

func HandleSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)
	guildID, _ := strconv.ParseUint(i.GuildID, 10, 64)

	notifyAdd := true
	notifyRem := false

	if val, exists := opts["announce_code_additions"]; exists {
		notifyAdd = val.BoolValue()
	}
	if val, exists := opts["announce_code_removals"]; exists {
		notifyRem = val.BoolValue()
	}

	err := db.CreateSubscription(channelID, guildID, notifyAdd, notifyRem)
	if err != nil {
		// duplicate? update instead
		if db.IsDuplicateErr(err) {
			err = db.UpdateSubscription(channelID, notifyAdd, notifyRem)
			if err != nil {
				RespondPrivate(s, i, fmt.Sprintf("Error updating existing subscription for <#%v>: %v", channelID, err))
				return
			} else {
				RespondPrivate(s, i, fmt.Sprintf("Resubscribed <#%v> with provided settings (default otherwise)!", channelID))
				return
			}
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error trying to create subscription for <#%v>: %v", channelID, err))
		return
	}

	RespondPrivate(s, i, fmt.Sprintf("Successfully subscribed <#%v>!", channelID))
}

func HandleUnsubscribe(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)

	// check if channel is subscribed
	if _, err := db.GetSubscription(channelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", channelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", channelID, err))
		return
	} 

	err := db.DeactivateSubscription(channelID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error trying to unsubscribe: <#%v>", err))
	} else {
		RespondPrivate(s, i, fmt.Sprintf("Successfully unsubscribed <#%v>!", channelID))
	}
}

func HandleFilterGames(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)

	// check if channel is subscribed
	if _, err := db.GetSubscription(channelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", channelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", channelID, err))
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

	err := db.SetGameFilters(channelID, games)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error setting game filters for <#%v>: %v", channelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully set game filters for <#%v>!", channelID))
}

func HandleAddPingRole(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)

	// check if channel is subscribed
	if _, err := db.GetSubscription(channelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", channelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", channelID, err))
		return
	} 

	roleID, _ := strconv.ParseUint(opts["role"].RoleValue(nil, "").ID, 10, 64)
	err := db.AddPingRole(channelID, roleID)
	if err != nil && !db.IsDuplicateErr(err) {
		RespondPrivate(s, i, fmt.Sprintf("Error adding ping role for <@&%v> in <#%v>: %v", roleID, channelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully added ping role for <@&%v> in <#%v>!", roleID, channelID))
}

func HandleRemovePingRole(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)

	// check if channel is subscribed
	if _, err := db.GetSubscription(channelID); err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("Please subscribe <#%v> first before running this command.", channelID))
			return
		}

		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", channelID, err))
		return
	} 

	roleID, _ := strconv.ParseUint(opts["role"].RoleValue(nil, "").ID, 10, 64)
	err := db.RemovePingRole(channelID, roleID)
	if err != nil  {
		RespondPrivate(s, i, fmt.Sprintf("Error removing ping role <@&%v> from <#%v>: %v", roleID, channelID, err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully removed ping role <@&%v> from <#%v>!", roleID, channelID))
}

func getSubsPrint(sub *db.Subscription) string {
	const TEMPLATE string = (
		"# <#%v>\n"+
		"**Active:** %v\n"+
		"**Announce additions:** %v\n"+
		"**Announce removals:** %v\n"+
		"**Game filter:**\n"+
		"%v" + 
		"**Roles to ping:**\n"+
		"%v")

	// get games
	gameList := ""
	games, err := db.GetSubscriptionGames(sub.ChannelID)
	if err != nil {
		return fmt.Sprintf("Error getting games for <#%v>: %v", sub.ChannelID, err)
	}
	for _, g := range games {
		gameList += fmt.Sprintf("- %v\n", g)
	}
	gameList = strings.TrimLeft(gameList, " \n")

	// get ping roles
	roleList := ""
	roles, err := db.GetPingRoles(sub.ChannelID)
	if err != nil {
		return fmt.Sprintf("Error getting ping roles for <#%v>: %v", sub.ChannelID, err)
	}
	for _, r := range roles {
		roleList += fmt.Sprintf("- <@&%v>\n", r)
	}
	roleList = strings.Trim(roleList, " \n")

	return fmt.Sprintf(TEMPLATE, sub.ChannelID, sub.Active, sub.AnnounceAdds, sub.AnnounceRems, gameList, roleList)
}

func HandleShowSubscription(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	if i.GuildID != "" { // don't run in DM environment
		if allChan := opts["all_channels"]; allChan != nil && allChan.BoolValue() {
			guildID, _ := strconv.ParseUint(i.GuildID, 10, 64)
			
			// get channels of server
			channels, err := db.GetSubscriptionsFromGuild(guildID)
			if err != nil {
				RespondPrivate(s, i, fmt.Sprintf("Error trying to get server channels: %v", err))
				return
			}
			
			result := fmt.Sprintf("**Subscriptions in server ID %v**\n", i.GuildID)
			for _, ch := range channels {
				result += getSubsPrint(&ch) + "\n"
			}
			RespondPrivate(s, i, result)
			return
		}
	}

	channelID := GetChannelID(i, opts)
	info, err := db.GetSubscription(channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondPrivate(s, i, fmt.Sprintf("No data available for <#%v>! This channel was never subscribed.", channelID))
			return
		}
		// unknown error
		RespondPrivate(s, i, fmt.Sprintf("Error checking subscription for <#%v>: %v", channelID, err))
		return 
	} 

	RespondPrivate(s, i, getSubsPrint(info))
}