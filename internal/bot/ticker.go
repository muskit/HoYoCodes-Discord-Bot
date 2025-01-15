package bot

import (
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
	"github.com/muskit/hoyocodes-discord-bot/pkg/util"
)

var embedSpacer discordgo.MessageEmbedField = discordgo.MessageEmbedField{
	Name: "\u200B",
}

func tickerText(game string, willRefresh bool) string {
	ret := (
		"## " + game + "\n")
		// "**Active Codes**\n")

	// non-recent codes
	codes := db.GetCodes(game, db.UnrecentCodes, false)
	if len(codes) > 0 {
		ret += util.CodeListing(codes) + "\n"
	}	

	// recent codes
	codes = db.GetCodes(game, db.RecentCodes, false)
	if len(codes) > 0 {
		ret += "\n**Recently Added**\n"
		ret += util.CodeListing(codes) + "\n"
	}	

	// livestream codes
	codes = db.GetCodes(game, db.AllCodes, true)
	if len(codes) > 0 {
		ret += "\n**Livestream (use ASAP; may expire sooner!)**\n"
		ret += util.CodeListing(codes) + "\n"
	}	

	// redemption shortcut
	redeem, exists := consts.RedeemURL[game]
	if exists {
		ret += fmt.Sprintf("\n**[Redemption Shortcut](<%v>)**", redeem)
	}

	// footer (stats & refresh time)
	checkTime, updateTime, err := db.GetScrapeTimes(game)
	if err != nil {
		log.Fatalf("Error getting update time for %v: %v", game, err)
	}
	footer := fmt.Sprintf("-# Checked <t:%v:R>; [source](<%v>) updated <t:%v:R>.\n", checkTime.Unix(), consts.ArticleURL[game], updateTime.Unix())
	if willRefresh {
		refreshTime := time.Now().Add(consts.UpdateInterval)
		footer += fmt.Sprintf("-# Refreshing in <t:%v:R>.\n", refreshTime.Unix())
	}

	ret += "\n" + footer

	return ret
}

func appendCodeFields(fields []*discordgo.MessageEmbedField, codes [][]string, game string) []*discordgo.MessageEmbedField {
	for _, code := range codes {
		var val string
		if codeURL := util.CodeRedeemURL(game, code[0]); codeURL != nil {
			val = fmt.Sprintf("[%v](%v)", code[1], *codeURL)
		} else {
			val = code [1]
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "`" + code[0] + "`",
			Value: val,
			Inline: true,
		})
	}
	return fields
}


func tickerEmbeds(game string, willRefresh bool) []*discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{ }

	unrecentCodes := db.GetCodes(game, db.UnrecentCodes, false)
	recentCodes := db.GetCodes(game, db.RecentCodes, false)
	livestreamCodes := db.GetCodes(game, db.AllCodes, true)

	if len(unrecentCodes) > 0 {
		fields = appendCodeFields(fields, unrecentCodes, game)
	}
	if len(recentCodes) > 0 {
		fields = append(fields,
			&embedSpacer,
			&discordgo.MessageEmbedField{
				Name: "--- Recently Added ---",
			},
		)
		fields = appendCodeFields(fields, recentCodes, game)
	}
	if len(livestreamCodes) > 0 {
		fields = append(fields,
			&embedSpacer,
			&discordgo.MessageEmbedField{
				Name: "--- Livestream Codes ---",
			},
		)
		fields = appendCodeFields(fields, livestreamCodes, game)
	}

	redeem, exists := consts.RedeemURL[game]
	if exists {
		fields = append(fields,
			// &fieldSpacer,
			&discordgo.MessageEmbedField{
				Value: fmt.Sprintf("**[Redemption page](%v)**", redeem),
			},
		)
	}

	checkTime, updateTime, err := db.GetScrapeTimes(game)
	if err != nil {
		log.Fatalf("Error getting update time for %v: %v", game, err)
	}
	footer := fmt.Sprintf("-# Checked <t:%v:R>; source updated <t:%v:R>.", checkTime.Unix(), updateTime.Unix())
	if willRefresh {
		refreshTime := checkTime.Add(consts.UpdateInterval)
		footer += fmt.Sprintf("\n-# Refreshing in <t:%v:R>.", refreshTime.Unix())
	}

	fields = append(fields,
		&discordgo.MessageEmbedField{
			Value: footer,
		},
	)

	// assemble embeds
	embeds := []*discordgo.MessageEmbed{}
	fieldLists := util.DownstackIntoSlices(fields, 25)
	for i, curFields := range fieldLists {
		curEmbed := discordgo.MessageEmbed{ Color: color[game], }
		if i == 0 {
			curEmbed = discordgo.MessageEmbed{
				Color: color[game],
				Title: game,
				URL: consts.ArticleURL[game],
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: image[game],
				},
			}
		}
		curEmbed.Fields = curFields
		embeds = append(embeds, &curEmbed)
	}

	return embeds
}

func UpdateEmbedTickersGame(s *discordgo.Session, game string) {
	tickers, err := db.GetTickers(game)
	if err != nil {
		log.Fatalf("Error getting embeds to update: %v", err)
	}

	embeds := tickerEmbeds(game, true)	

	for _, msg := range tickers {
		channelID, messageID := msg[0], msg[1]
		edit := discordgo.MessageEdit{
			Channel: channelID,
			ID: messageID,
			Content: new(string),
			Embeds: &embeds,
		}
		if _, err = s.ChannelMessageEditComplex(&edit); err != nil {
			if strings.Contains(err.Error(), "HTTP 404 Not Found") {
				// message no longer exists -- delete from db
				msgNum, _ := strconv.ParseUint(messageID, 10, 64)
				err := db.RemoveTicker(msgNum)
				if err != nil {
					slog.Error(fmt.Sprintf("404'd removing ticker from db during update: %s", err))
				}
			} else {
				log.Fatalf("Error updating ticker: %v", err)
			}
		}
	}
}

func UpdateTextTickersGame(s *discordgo.Session, game string) {
	tickers, err := db.GetTickers(game)
	if err != nil {
		log.Fatalf("Error getting tickers: %v", err)
	}

	slog.Debug("", "game", game)
	content := tickerText(game, true)
	slog.Debug("", "len(content)", len(content))

	for _, emb := range tickers {
		channelID, messageID := emb[0], emb[1]
		
		// update ticker
		edit := discordgo.MessageEdit{
			Channel: channelID,
			ID: messageID,
			Content: &content,
			Embeds: &[]*discordgo.MessageEmbed{}, // delete old embed if still there
		}
		if _, err := s.ChannelMessageEditComplex(&edit); err != nil {
			if strings.Contains(err.Error(), "HTTP 404 Not Found") {
				// message no longer exists -- delete from db
				msgNum, _ := strconv.ParseUint(messageID, 10, 64)
				err := db.RemoveTicker(msgNum)
				if err != nil {
					slog.Error(fmt.Sprintf("404'd removing ticker from db during update: %s", err))
				}
			} else {
				log.Fatalf("Error updating ticker: %v", err)
			}
		}
	}
}