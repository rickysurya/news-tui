package scraper

import (
	"database/sql"
	"log"

	"github.com/gocolly/colly/v2"
	"github.com/rickysurya/news-tui/pkg/config"
)

func NewCollector(db *sql.DB, selectors []config.Selector) *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	for _, s := range selectors {
		registerHandler(c, db, s)
	}
	return c
}

func registerHandler(c *colly.Collector, db *sql.DB, s config.Selector) {
	c.OnHTML(s.Container, func(e *colly.HTMLElement) {
		article := config.Article{
			Title: e.ChildText(s.Title),
			Link:  e.Request.AbsoluteURL(e.ChildAttr(s.Link, "href")),
		}
		if article.Title == "" || article.Link == "" {
			return
		}
		if err := SaveArticle(db, article); err != nil {
			log.Println("failed to save article to db:", err)
		}
	})
}

func Visit(c *colly.Collector, urls []string) {
	for _, url := range urls {
		if err := c.Visit(url); err != nil {
			log.Printf("failed to visit %s: %v\n", url, err)
		}
	}
}
