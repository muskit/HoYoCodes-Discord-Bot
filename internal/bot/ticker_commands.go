package bot

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
)

var color map[string]int = map[string]int{
	"Honkai Impact 3rd": 0x2ECFE2,
	"Genshin Impact": 0xdbc06f,
	"Honkai Star Rail": 0x5475d8,
	"Zenless Zone Zero": 0xcc7b30,
}

var image map[string]string = map[string]string{
	"Honkai Impact 3rd": "https://cdn2.steamgriddb.com/icon/ba95d78a7c942571185308775a97a3a0.png",
	"Genshin Impact": "https://static.wikia.nocookie.net/gensin-impact/images/8/80/Genshin_Impact.png",
	"Honkai Star Rail": "https://static.wikia.nocookie.net/houkai-star-rail/images/8/84/Honkai_Star_Rail_App.png",
	"Zenless Zone Zero": "https://fastcdn.hoyoverse.com/static-resource-v2/2023/11/02/bf82c4f8573eb6292f338a3ec41c1615_6171503094506184079.png",
}

func HandleCreateTicker(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	guildID := i.GuildID
	game := opts["game"].StringValue()
	embeds := tickerEmbeds(game, true)

	message, err := s.ChannelMessageSendEmbeds(i.ChannelID, embeds)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error creating ticker: %v", err))
		return
	}

	messageID := message.ID
	err = db.AddTicker(messageID, game, i.ChannelID, guildID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf(
			"Created ticker but can't save for updating: %v\n" +
			"This ticker will not update; please delete it and try again.", err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully created ticker in <#%v> for %v!", i.ChannelID, game))
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
	err = db.RemoveTicker(messageID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error removing ticker from tracking: %v", err))
		return
	}
	RespondPrivate(s, i, "Ticker successfully removed!")
}

func HandleGetTickers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	tickers, err := db.GetGuildTickers(i.GuildID)
	if err != nil {
		log.Fatalf("Error getting tickers from guild %v: %v", i.GuildID, err)
	}

	out := fmt.Sprintf("**Tickers in server ID %v**\n", i.GuildID)
	for _, t := range tickers {
		url := fmt.Sprintf(consts.MessageLinkTemplate, i.GuildID, t.ChannelID, t.MessageID)
		out += fmt.Sprintf("- %v (%v)\n", url, t.Game)
	}
	RespondPrivate(s, i, strings.Trim(out, " \t\n"))
}

func HandleActiveCodes(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	game := opts["game"].StringValue()
	embeds := tickerEmbeds(game, false)
	resp := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
		},
	}
	if notifyPvt, exists := opts["private"]; exists && notifyPvt.BoolValue() {
		resp.Data.Flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &resp)
}