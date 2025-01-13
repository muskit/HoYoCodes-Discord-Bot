package scraper

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gocolly/colly"
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
	slog.Debug(fmt.Sprintf("[%s]\n", cfg.Game))

	// scraped data
	activeCodes := make(map[string]string)
	datetime := ""

	c := colly.NewCollector(colly.AllowedDomains("www.pockettactics.com"))

	// --- callback setup ---
	c.OnRequest(func(r *colly.Request) {
		slog.Debug(fmt.Sprintf("Visiting %s", cfg.URL))
	})
	slog.Debug(fmt.Sprintf("Searching for \"%s\"", cfg.Heading))

	// populate codes
	c.OnHTML("strong, b", func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, cfg.Heading) {
			slog.Debug("", "header", h.Text)
			if strings.Contains(h.Text, "expire") {
				slog.Debug("Appears to have expired according to header; stopping...")
				return
			}

			slog.Debug("Gathering codes...")

			listContainer := h.DOM.Parent().Next()
			list := listContainer.Children()

			for i, elem := range list.Nodes {
				entry := elem.FirstChild
				key := entry.FirstChild.Data
				desc := string([]rune(entry.NextSibling.Data)[3:])

				activeCodes[key] = desc
				slog.Debug(fmt.Sprintf("%d: [%s] (%s)\n", i, key, desc))
			}
		}
	})

	// populate datetime
	c.OnHTML("time", func(h *colly.HTMLElement) {
		if h.DOM.HasClass("updated") {
			datetime = h.Attr("datetime")
			slog.Debug(fmt.Sprintf("Update datetime: %s", datetime))
		}
	})

	// begin scrape
	c.Visit(cfg.URL)

	// TODO: check that data to return is good

	slog.Debug(fmt.Sprintf("%d codes", len(activeCodes)))
	if len(activeCodes) == 0 {
		slog.Warn("Returning 0 codes!", "game", cfg.Game, "heading", cfg.Heading)
	}

	slog.Debug("Finished scraping.")
	return activeCodes, datetime
}