package bot

import (
	"fmt"
	"strings"

	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

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
