package scraper

import (
	"log"
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
func scrapePJT(url string, heading string) (map[string]string, string) {
	// scraped data
	activeCodes := make(map[string]string)
	datetime := ""

	c := colly.NewCollector(colly.AllowedDomains("www.pockettactics.com"))

	// --- callback setup ---
	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting %s...\n", url)
	})
	log.Printf("Searching for \"%s\"", heading)

	// populate codes
	c.OnHTML("strong, b", func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, heading) {
			log.Printf("FOUND HEADER: %s\n", h.Text)
			if strings.Contains(h.Text, "expire") {
				log.Println("Appears to have expired according to header; stopping...")
				return
			}

			log.Println("Gathering codes...")

			listContainer := h.DOM.Parent().Next()
			list := listContainer.Children()

			for _, elem := range list.Nodes {
				entry := elem.FirstChild
				key := entry.FirstChild.Data
				desc := string([]rune(entry.NextSibling.Data)[3:])

				activeCodes[key] = desc
				// log.Printf("%d: [%s] (%s)\n", i, key, desc)
			}
		} else {
			// log.Printf("Didn't find \"%s\"\n", identifierText)
		}
	})

	// populate datetime
	c.OnHTML("time", func(h *colly.HTMLElement) {
		if h.DOM.HasClass("updated") {
			datetime = h.Attr("datetime")
			log.Printf("Update datetime: %s\n", datetime)
		}
	})

	// begin scrape
	c.Visit(url)

	// TODO: check that data to return is good

	// log.Printf("%d codes", len(activeCodes))
	if len(activeCodes) == 0 {
		log.Println("WARNING: returning 0 codes!")
	}

	log.Println("done")
	return activeCodes, datetime
}

func ScrapeGame(config ScrapeConfig) (map[string]string, string) {
	log.Printf("--- [%s] ---\n", config.Game)
	return scrapePJT(
		config.URL,
		config.Heading,
	)
}