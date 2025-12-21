package bot

import (
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/muskit/hoyocodes-discord-bot/internal/db"
	"github.com/muskit/hoyocodes-discord-bot/pkg/consts"
	"github.com/muskit/hoyocodes-discord-bot/pkg/util"
)

func appendCodeFields(fields []*discordgo.MessageEmbedField, codes [][]string, game string) []*discordgo.MessageEmbedField {
	for _, code := range codes {
		var val string
		if codeURL := util.CodeRedeemURL(code[0], game); codeURL != nil {
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
	fieldLists := [][]*discordgo.MessageEmbedField{}

	unrecentCodes := db.GetCodes(game, db.Unrecent, false)
	recentCodes := db.GetCodes(game, db.Recent, false)
	livestreamCodes := db.GetCodes(game, db.All, true)
	numCodes := len(unrecentCodes)+len(recentCodes)+len(livestreamCodes)

	// code embeds
	if len(unrecentCodes) > 0 {
		fields := appendCodeFields([]*discordgo.MessageEmbedField{}, unrecentCodes, game)
		fieldLists = append(fieldLists, util.DownstackIntoSlices(fields, 25)...)
	}
	if len(recentCodes) > 0 {
		fields := []*discordgo.MessageEmbedField{
			{
				Name: "--- Recently Added ---",
			},
		}
		fields = appendCodeFields(fields, recentCodes, game)
		fieldLists = append(fieldLists, util.DownstackIntoSlices(fields, 25)...)
	}
	if len(livestreamCodes) > 0 {
		fields := []*discordgo.MessageEmbedField{
			{
				Name: "--- Livestream Codes (use ASAP; may expire soon!) ---",
			},
		}
		fields = appendCodeFields(fields, livestreamCodes, game)
		fieldLists = append(fieldLists, util.DownstackIntoSlices(fields, 25)...)
	}

	// footer embed
	footerFields := []*discordgo.MessageEmbedField{}

	redeem, exists := consts.RedeemURL[game]
	redeemField := &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("%d codes reported active", numCodes),
	}
	if exists {
		redeemField.Value = fmt.Sprintf("**[Redemption page](%v)**", redeem)
	}
	footerFields = append(footerFields, redeemField)

	checkTime, updateTime, err := db.GetScrapeTimes(game)
	if err != nil {
		log.Fatalf("Error getting update time for %v: %v", game, err)
	}
	timeField := fmt.Sprintf("-# Checked <t:%v:R>; [source](%v) updated <t:%v:R>.", checkTime.Unix(), consts.ArticleURL[game], updateTime.Unix())
	if willRefresh {
		refreshTime := checkTime.Add(consts.UpdateInterval)
		timeField += fmt.Sprintf("\n-# Refreshing <t:%v:R>.", refreshTime.Unix())
	}

	footerFields = append(footerFields,
		&discordgo.MessageEmbedField{
			Value: timeField,
		},
	)

	// assemble downstacked
	downstacked := []*discordgo.MessageEmbed{}
	for i, curFields := range fieldLists {
		curEmbed := discordgo.MessageEmbed{ Color: color[game], }
		if i == 0 {
			curEmbed = discordgo.MessageEmbed{
				Color: color[game],
				Title: game,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: image[game],
				},
			}
		}
		curEmbed.Fields = curFields
		downstacked = append(downstacked, &curEmbed)
	}

	footerEmbed := discordgo.MessageEmbed{Color: color[game], Fields: footerFields}
	return append(downstacked, &footerEmbed)
}

func UpdateEmbedTickersGame(s *discordgo.Session, game string) {
	tickers, err := db.GetGameTickers(game)
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


		var err error = nil
		for attempts := 0; attempts <= 5; attempts++ {
			if attempts == 5 {
				log.Fatalf("Error updating ticker after 5 attempts: %v", err)
			}

			_, err = s.ChannelMessageEditComplex(&edit);
			if err != nil {
				slog.Debug(fmt.Sprintf("Encountered problem updating ticker msg #%v!\n\tError() = %s", edit.ID, err.Error()))
				if strings.Contains(err.Error(), "HTTP 404") {
					// message no longer exists -- delete from db
					err := db.RemoveTicker(messageID)
					if err != nil {
						slog.Error(fmt.Sprintf("Error removing 404'd ticker #%v from db during update: %v", edit.ID, err))
						break
					} else {
						slog.Warn("Successfully removed 404'd ticker.")
						break
					}
				} else if strings.Contains(err.Error(), "HTTP 403") {
					slog.Warn(fmt.Sprintf("HTTP Forbidden 403 while editing ticker %v: %v", messageID, err))
					break
				} else {
					slog.Warn(fmt.Sprintf("Unknown error updating ticker #%v!!\n\tError() = %s", edit.ID, err.Error()))
				}
			} else {
				break
			}

			// we errored; attempt to resolve by delaying repeat:
			// HTTP 503 Service Unavailable, upstream connect error or disconnect/reset before headers. reset reason: overflow
			slog.Warn(fmt.Sprintf("Error updating ticker: %v", err))
			time.Sleep(5*time.Second)
		}
	}
}
