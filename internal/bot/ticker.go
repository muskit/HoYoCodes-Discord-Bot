package bot

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cdfmlr/ellipsis"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
)

var color map[string]int = map[string]int{
	"Honkai Impact 3rd": 0x2ECFE2,
	"Genshin Impact": 0xdbc06f,
	"Honkai Star Rail": 0x5475d8,
	"Zenless Zone Zero": 0xcc7b30,
}

var articleURL map[string]string = map[string]string{
	"Honkai Impact 3rd": "https://www.pockettactics.com/honkai-impact/codes",
	"Genshin Impact": "https://www.pockettactics.com/genshin-impact/codes",
	"Honkai Star Rail": "https://www.pockettactics.com/honkai-star-rail/codes",
	"Zenless Zone Zero": "https://www.pockettactics.com/zenless-zone-zero/codes",
}

var image map[string]string = map[string]string{
	"Honkai Impact 3rd": "https://cdn2.steamgriddb.com/icon/ba95d78a7c942571185308775a97a3a0.png",
	"Genshin Impact": "https://static.wikia.nocookie.net/gensin-impact/images/8/80/Genshin_Impact.png",
	"Honkai Star Rail": "https://static.wikia.nocookie.net/houkai-star-rail/images/8/84/Honkai_Star_Rail_App.png",
	"Zenless Zone Zero": "https://fastcdn.hoyoverse.com/static-resource-v2/2023/11/02/bf82c4f8573eb6292f338a3ec41c1615_6171503094506184079.png",
}

var redeemURL map[string]string = map[string]string{
	"Genshin Impact": "https://genshin.hoyoverse.com/en/gift",
	"Honkai Star Rail": "https://hsr.hoyoverse.com/gift",
	"Zenless Zone Zero": "https://zenless.hoyoverse.com/redemption",
}

func CodeListing(codes [][]string) string {
	ret := ""
	for _, elem := range codes {
		var line string
		code, description := elem[0], ellipsis.Ending(elem[1], 50)
		line = fmt.Sprintf("- `%v` - %v", code, description)
		ret += line + "\n"
	}
	return strings.Trim(ret, " \n	")
}

func createCodePrint(game string, willRefresh bool) string {
	ret := (
		"## " + game + "\n"+
		"**Active Codes**\n")

	// non-recent codes
	codes := db.GetCodes(game, db.UnrecentCodes, false)
	if len(codes) > 0 {
		ret += CodeListing(codes) + "\n"
	}	

	// recent codes
	codes = db.GetCodes(game, db.RecentCodes, false)
	if len(codes) > 0 {
		ret += "\n**Added Last Update**\n"
		ret += CodeListing(codes) + "\n"
	}	

	// livestream codes
	codes = db.GetCodes(game, db.AllCodes, true)
	if len(codes) > 0 {
		ret += "\n**Livestream (use ASAP; may expire sooner!)**\n"
		ret += CodeListing(codes) + "\n"
	}	

	// redemption shortcut
	redeem, exists := redeemURL[game]
	if exists {
		ret += fmt.Sprintf("\n**[Redemption Shortcut](<%v>)**", redeem)
	}

	// footer (stats & refresh time)
	checkTime, updateTime, err := db.GetScrapeTimes(game)
	if err != nil {
		log.Fatalf("Error getting update time for %v: %v", game, err)
	}
	footer := fmt.Sprintf("-# Checked <t:%v:R>; [source](<%v>) updated <t:%v:R>.\n", checkTime.Unix(), articleURL[game], updateTime.Unix())
	if willRefresh {
		refreshTime := checkTime.Add(2*time.Hour) // TODO: set update interval in config
		footer += fmt.Sprintf("-# Refreshing in <t:%v:R>.\n", refreshTime.Unix())
	} else {
		footer += "-# This ticker will not auto-refresh.\n"
	}

	ret += "\n" + footer

	return ret
}

func HandleCreateTicker(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)
	game := opts["game"].StringValue()

	content := createCodePrint(game, true)
	message, err := s.ChannelMessageSend(strconv.FormatUint(channelID, 10), content)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error creating ticker: %v", err))
		return
	}

	messageID, _ := strconv.ParseUint(message.ID, 10, 64)
	err = db.AddTicker(messageID, game, channelID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf(
			"Created ticker but can't save for updating: %v\n" +
			"This ticker will not update; please delete it and try again.", err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully created ticker in <#%v> for %v!", channelID, game))
}

func HandleDeleteTicker(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	messageURL := opts["message_link"].StringValue()

	url, err := url.Parse(messageURL)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error parsing message link: %v", err))
		return
	}

	pTrim := strings.Trim(url.Path, "/")
	path := strings.Split(pTrim, "/")
	if len(path) != 4 {
		RespondPrivate(s, i, fmt.Sprintf("Bad URL: path length is %v, expected 4.", len(path)))
		return
	}

	channelID := path[2]
	messageID := path[3]

	// message deletion abuse prevention
	msg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Couldn't fetch message: %v", err))
		return
	}
	if msg.Author.String() != s.State.User.String() {
		RespondPrivate(s, i, "Can't delete message as I didn't make it!")
		return
	}

	if err = s.ChannelMessageDelete(channelID, messageID); err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error deleting message %v: %v", messageID, err))
		return
	}

	// remove message from DB
	id, _ := strconv.ParseUint(messageID, 10, 64)
	err = db.RemoveTicker(id)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error removing ticker from tracking: %v", err))
		return
	}
	RespondPrivate(s, i, "Ticker successfully removed!")
}

func HandleActiveCodes(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	game := opts["game"].StringValue()
	content := createCodePrint(game, false)
	resp := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}
	s.InteractionRespond(i.Interaction, &resp)
}