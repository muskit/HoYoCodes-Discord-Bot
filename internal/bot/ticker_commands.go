package bot

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
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
	channelID := GetChannelID(i, opts)
	game := opts["game"].StringValue()
	embeds := tickerEmbeds(game, true)

	message, err := s.ChannelMessageSendEmbeds(strconv.FormatUint(channelID, 10), embeds)
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
	embeds := tickerEmbeds(game, false)
	resp := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}
	s.InteractionRespond(i.Interaction, &resp)
}