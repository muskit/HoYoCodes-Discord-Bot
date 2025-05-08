package scraper

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gocolly/colly"
	"github.com/muskit/hoyocodes-discord-bot/pkg/util"
)

type ScrapeConfig struct {
	Game string
	URL string
	Heading string
}

var HI3_ScrCfg ScrapeConfig = ScrapeConfig{
	Game: "Honkai Impact 3rd",
	URL: "https://www.pockettactics.com/honkai-impact/codes",
	Heading: "Here are all the new Honkai Impact codes",
}
var GI_ScrCfg ScrapeConfig = ScrapeConfig{
	Game: "Genshin Impact",
	URL: "https://www.pockettactics.com/genshin-impact/codes",
	Heading: "Here are all of the new Genshin Impact codes",
}
var HSR_ScrCfg ScrapeConfig = ScrapeConfig{
	Game: "Honkai Star Rail",
	URL: "https://www.pockettactics.com/honkai-star-rail/codes",
	Heading: "Here are all of the new Honkai Star Rail codes",
}
var ZZZ_ScrCfg ScrapeConfig = ScrapeConfig{
	Game: "Zenless Zone Zero",
	URL: "https://www.pockettactics.com/zenless-zone-zero/codes",
	Heading: "Here are all of the new ZZZ codes",
}

var Configs []ScrapeConfig = []ScrapeConfig{
	HI3_ScrCfg,
	GI_ScrCfg,
	HSR_ScrCfg,
	ZZZ_ScrCfg,
}

// Given a Project Tactics article containing MiHoYo game codes,
// return a map of codes and their description, as well as
// the datetime which the data was updated.
func ScrapePJT(cfg ScrapeConfig) (map[string]string, string) {
	slog.Info(fmt.Sprintf("[%s] - %s\n", cfg.Game, cfg.Heading))

	// scraped data
	activeCodes := make(map[string]string)
	datetime := ""

	c := colly.NewCollector(colly.AllowedDomains("www.pockettactics.com"))

	// --- callback setup ---
	c.OnRequest(func(r *colly.Request) {
		slog.Info(fmt.Sprintf("Visiting %s", cfg.URL))
	})

	// populate codes
	c.OnHTML("strong, b", func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, cfg.Heading) {
			slog.Info("", "header", h.Text)
			if strings.Contains(h.Text, "expire") {
				slog.Info("Appears to have expired according to header; stopping...")
				return
			}

			slog.Info("Gathering codes...")

			listContainer := h.DOM.Parent().Next()
			list := listContainer.Children()

			for i, e := range list.Nodes {
				key := ""
				entry := e.FirstChild
				if entry.FirstChild != nil {
					key = entry.FirstChild.Data
				} else {
					key = entry.Data
				}
				key = util.AlphaNumStrip(key)
				slog.Debug(fmt.Sprintf("key: %v", key))

				desc := ""
				entryNext := entry.NextSibling
				if entryNext != nil {
					slog.Debug(fmt.Sprintf("entryNext: %v", entryNext.Data))
					desc = string([]rune(entryNext.Data))
					desc = util.AlphaNumStrip(desc)
				} else {
					slog.Warn(fmt.Sprintf("%v has no description element", key))
				}
				activeCodes[key] = desc
				slog.Info(fmt.Sprintf("%d: [%s] (%s)\n", i, key, desc))
			}
		}
	})

	// populate datetime
	c.OnHTML("time", func(h *colly.HTMLElement) {
		if h.DOM.HasClass("updated") {
			datetime = h.Attr("datetime")
			slog.Info(fmt.Sprintf("Update datetime: %s", datetime))
		}
	})

	// begin scrape
	c.Visit(cfg.URL)

	// TODO: check that data to return is good

	if len(activeCodes) == 0 {
		slog.Warn("Returning 0 codes!")
	} else {
		slog.Info(fmt.Sprintf("Found %d codes", len(activeCodes)))
	}

	slog.Debug("Finished scraping.")
	return activeCodes, datetime
}