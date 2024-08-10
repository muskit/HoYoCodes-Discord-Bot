package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

// Given a Project Tactics article containing MiHoYo game codes,
// return a map of codes and their description, as well as
// the datetime which the data was updated.
func ScrapePJT(url string, listIntroText string) (map[string]string, string) {
	// scraped data
	activeCodes := make(map[string]string)
	datetime := ""

	c := colly.NewCollector(colly.AllowedDomains("www.pockettactics.com"))

	// --- callback setup ---
	// TODO: error handling
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s...\n", url)
	})

	// populate codes
	c.OnHTML("strong", func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, listIntroText) {
			fmt.Println("Found intro text, getting code list...")

			list := h.DOM.Parent().Next().Children()
			for i, elem := range list.Nodes {
				entry := elem.FirstChild
				key := entry.FirstChild.Data
				desc := string([]rune(entry.NextSibling.Data)[3:])

				activeCodes[key] = desc
				fmt.Printf("%d: [%s] (%s)\n", i, key, desc)
			}
		}
	})

	// populate datetime
	c.OnHTML("time", func(h *colly.HTMLElement) {
		if h.DOM.HasClass("updated") {
			datetime = h.Attr("datetime")
			fmt.Printf("Update datetime: %s\n", datetime)
		}
	})
	
	// begin scrape
	c.Visit(url)

	// scrape done
	fmt.Println("done")
	return activeCodes, datetime
}

func RunScraper() {
	fmt.Println("--- [HONKAI IMPACT] ---")
	ScrapePJT(
		"https://www.pockettactics.com/honkai-impact/codes",
		"Here are all the new Honkai Impact codes",
	)

	fmt.Println("--- [GENSHIN IMPACT] ---")
	ScrapePJT(
		"https://www.pockettactics.com/genshin-impact/codes",
		"Here are all of the new Genshin Impact codes",
	)

	fmt.Println("--- [HONKAI STAR RAIL] ---")
	ScrapePJT(
		"https://www.pockettactics.com/honkai-star-rail/codes",
		"Here are all of the new Honkai Star Rail codes",
	)

	fmt.Println("---[ZENLESS ZONE ZERO] ---")
	ScrapePJT(
		"https://www.pockettactics.com/zenless-zone-zero/codes",
		"Here are all of the new Zenless Zone Zero codes",
	)
}