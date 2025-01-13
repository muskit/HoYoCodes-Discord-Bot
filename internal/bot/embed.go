package bot

import (
	"fmt"
	"log"
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

var fieldSpacer discordgo.MessageEmbedField = discordgo.MessageEmbedField{
	Name: "\u200B",
}

func appendCodeParam(redeemURL string, code string) string {
	return redeemURL + "?code=" + code
}

func appendCodeFields(fields []*discordgo.MessageEmbedField, codes [][]string, game string) []*discordgo.MessageEmbedField {
	for _, code := range codes {
		var val string
		if url, exists := redeemURL[game]; exists {
			val = fmt.Sprintf("[%v](%v)", code[1], appendCodeParam(url, code[0]))
		} else {
			val = code [1]
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name: code[0],
			Value: val,
			Inline: true,
		})
	}
	return fields
}

func createEmbed(game string) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{}

	// non-recent codes
	codes := db.GetCodes(game, false, false)
	if len(codes) > 0 {
		fields = append(fields,
			&discordgo.MessageEmbedField{
				Name: "--- Active Codes ---",
			},
		)
		fields = appendCodeFields(fields, codes, game)
	}

	codes = db.GetCodes(game, true, false)
	if len(codes) > 0 {
		fields = append(fields,
			&fieldSpacer,
			&discordgo.MessageEmbedField{
				Name: "--- Recently-Added Codes ---",
			},
		)
		fields = appendCodeFields(fields, codes, game)
	}

	// livestream codes
	codes = db.GetCodes(game, false, true)
	if len(codes) > 0 {
		fields = append(fields,
			&fieldSpacer,
			&discordgo.MessageEmbedField{
				Name: "--- Livestream Codes ---",
			},
		)
		fields = appendCodeFields(fields, codes, game)
	}

	redeem, exists := redeemURL[game]
	if exists {
		fields = append(fields,
			// &fieldSpacer,
			&discordgo.MessageEmbedField{
				Value: fmt.Sprintf("**[Redemption page](%v)**", redeem),
			},
		)
	}

	checkTime, updateTime, err := db.GetScrapeStats(game)
	if err != nil {
		log.Fatalf("Error getting update time for %v: %v", game, err)
	}
	fields = append(fields,
		&discordgo.MessageEmbedField{
			Value: fmt.Sprintf("-# Checked <t:%v:R>; source updated <t:%v:R>.", checkTime.Unix(), updateTime.Unix()),
		},
	)

	embed := &discordgo.MessageEmbed{
		Color: color[game],
		Title: game,
		URL: articleURL[game],
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: image[game],
		},
		Fields: fields,
	}

	return embed
}

// returns message ID of embed
func HandleCreateEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
	channelID := GetChannelID(i, opts)
	game := opts["game"].StringValue()

	embed := createEmbed(game)
	message, err := s.ChannelMessageSendEmbed(strconv.FormatUint(channelID, 10), embed)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error creating embed: %v", err))
		return
	}

	messageID, _ := strconv.ParseUint(message.ID, 10, 64)
	err = db.AddEmbed(messageID, game, channelID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf(
			"Created embed but can't save for updating: %v\n" +
			"This embed will not update; please delete it and try again.", err))
		return
	}
	RespondPrivate(s, i, fmt.Sprintf("Successfully created embed in <#%v> for %v!", channelID, game))
}

func HandleDeleteEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, opts CmdOptMap) {
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
	if err := s.ChannelMessageDelete(channelID, messageID); err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error deleting message %v: %v", messageID, err))
		return
	}

	// remove message from DB
	id, _ := strconv.ParseUint(messageID, 10, 64)
	err = db.RemoveEmbed(id)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error removing embed from tracking: %v", err))
		return
	}
	RespondPrivate(s, i, "Embed successfully removed!")
}

