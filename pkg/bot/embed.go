package bot

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/pkg/db"
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

func AppendCodeParam(redeemURL string, code string) string {
	return redeemURL + "?code=" + code
}

func createEmbed(game string) *discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{}

	// non-recent codes
	codes := db.GetCodes(game, false, false)
	for _, code := range codes {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: code[0],
			Value: code[1],
			Inline: true,
		})
	}

	// spacer
	fields = append(fields, 
		&discordgo.MessageEmbedField{
			Name: "\u200B",
		},
		&discordgo.MessageEmbedField{
			Name: "--- Recently-Added Codes ---",
		},
	)

	// TODO: recent codes

	// spacer
	fields = append(fields, 
		&discordgo.MessageEmbedField{
			Name: "\u200B",
		},
		&discordgo.MessageEmbedField{
			Name: "--- Livestream Codes ---",
		},
	)

	// livestream codes
	codes = db.GetCodes(game, false, true)
	for _, code := range codes {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: code[0],
			Value: code[1],
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			// Name: "Current Active Codes",
		},
		Color: color[game],
		Title: game,
		URL: articleURL[game],
		// Image: &discordgo.MessageEmbedImage{
		// 	URL: image[game],
		// },
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: image[game],
		},
		Fields: fields,
		// TODO: use article edit datetime
		Footer: &discordgo.MessageEmbedFooter{Text: "Refreshed"},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
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
	_, err = db.DBCfg.Exec("DELETE FROM Embeds WHERE message_id = ?", messageID)
	if err != nil {
		RespondPrivate(s, i, fmt.Sprintf("Error removing embed from tracking: %v", err))
		return
	}
	RespondPrivate(s, i, "Embed successfully removed!")
}

func UpdateGameEmbeds(s *discordgo.Session, game string) {
	embeds, err := db.GetEmbeds(game)
	if err != nil {
		log.Fatalf("Error getting embeds to update: %v", err)
	}

	embedContent := createEmbed(game)

	for _, emb := range embeds {
		_, err = s.ChannelMessageEditEmbed(emb[0], emb[1], embedContent)
		if err != nil {
			if strings.Contains(err.Error(), "HTTP 404 Not Found") {
				// embed message no longer exists
				msgNum, _ := strconv.ParseUint(emb[1], 10, 64)
				err := db.RemoveEmbed(msgNum, game)
				if err != nil {
					log.Printf("WARNING: error removing 404'd embed from db: %v", err)
				}
			}
		}
	}

	// Contains(err.Error(), "HTTP 404 Not Found") for message/channel not found
	// delete embed from DB if not found according to Discord
}
