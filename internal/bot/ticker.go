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

func tickerContent(game string, willRefresh bool, refreshTime time.Time) string {
	ret := (
		"## " + game + "\n"+
		"**Active Codes**\n")

	// non-recent codes
	codes := db.GetCodes(game, db.UnrecentCodes, false)
	if len(codes) > 0 {
		ret += util.CodeListing(codes) + "\n"
	}	

	// recent codes
	codes = db.GetCodes(game, db.RecentCodes, false)
	if len(codes) > 0 {
		ret += "\n**Added Last Update**\n"
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
		footer += fmt.Sprintf("-# Refreshing in <t:%v:R>.\n", refreshTime.Unix())
	} else {
		footer += "-# This ticker will not auto-refresh.\n"
	}

	ret += "\n" + footer

	return ret
}

func UpdateTickersGame(s *discordgo.Session, game string, refreshTime time.Time) {
	tickers, err := db.GetTickers(game)
	if err != nil {
		log.Fatalf("Error getting tickers: %v", err)
	}

	slog.Debug("", "game", game)
	content := tickerContent(game, true, refreshTime)
	slog.Debug("", "len(content)", len(content))
	// slog.Debug(fmt.Sprintf("Content is as follows:\n%v", content))

	for _, emb := range tickers {
		channelID, messageID := emb[0], emb[1]
		
		// update ticker
		edit := discordgo.MessageEdit{
			Channel: channelID,
			ID: messageID,
			Content: &content,
			Embeds: &[]*discordgo.MessageEmbed{}, // delete old embed if still there
		}
		// if _, err := s.ChannelMessageEdit(channelID, messageID, content, ); err != nil {
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