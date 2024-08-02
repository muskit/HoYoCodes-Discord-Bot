package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func scrapeHonkaiStarRail() {
	const url = "https://www.pockettactics.com/honkai-star-rail/codes"
	c := colly.NewCollector(colly.AllowedDomains("www.pockettactics.com"))

	// callback setup
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s...\n", url)
	})

	c.Visit(url)
}

func RunScraper() {
	scrapeHonkaiStarRail()
}