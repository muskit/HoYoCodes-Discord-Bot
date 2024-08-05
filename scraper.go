package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

func scrapeHonkaiStarRail() map[string]string {
	const URL = "https://www.pockettactics.com/honkai-star-rail/codes"
	const INTRO_TEXT = "Here are all of the new Honkai Star Rail codes"

	activeCodes := make(map[string]string)

	c := colly.NewCollector(colly.AllowedDomains("www.pockettactics.com"))

	// callback setup
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s...\n", URL)
	})

	c.OnHTML("strong", func(h *colly.HTMLElement) {
		if strings.Contains(h.Text, INTRO_TEXT) {
			fmt.Println("Found intro text, getting code list...")

			list := h.DOM.Parent().Next().Children()
			for i, elem := range list.Nodes {
				entry := elem.FirstChild
				key, desc := entry.FirstChild.Data, strings.TrimLeft(entry.NextSibling.Data[5:], " ")

				activeCodes[key] = desc
				fmt.Printf("%d: [%s] (%s)\n", i, key, desc)
			}
		}
	})
	
	// begin scrape
	c.Visit(URL)
	fmt.Println("done")

	return activeCodes
}

func RunScraper() {
	scrapeHonkaiStarRail()
}