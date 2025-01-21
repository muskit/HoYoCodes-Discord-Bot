package bot

import (
	"fmt"
	"strings"

	"github.com/muskit/hoyocodes-discord-bot/internal/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
)

func getSubsPrint(sub *db.Subscription) string {
	const TEMPLATE string = (
		"__**<#%v>**__\n"+
		"**Active:** %v\n"+
		"**Announce additions:** %v\n"+
		"**Announce removals:** %v\n"+
		"**Tracked games:**\n"+
		"%v" + 
		"**Roles to ping:**\n"+
		"%v")

	// get games
	gameList := ""
	games, err := db.GetSubscriptionGames(sub.ChannelID)
	if err != nil {
		return fmt.Sprintf("Error getting games for <#%v>: %v", sub.ChannelID, err)
	}

	if len(games) == 0 {
		// list all games
		games = consts.Games
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
	roleList = strings.TrimLeft(roleList, " \t\n")

	return strings.Trim(fmt.Sprintf(TEMPLATE, sub.ChannelID, sub.Active, sub.AnnounceAdds, sub.AnnounceRems, gameList, roleList), " \t\n")
}
