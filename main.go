package main

import (
	"log"

	"github.com/rickysurya/news-tui/pkg/config"
	"github.com/rickysurya/news-tui/pkg/scraper"
	"github.com/rickysurya/news-tui/pkg/tui"
)

func main() {
	urls, selectors := config.LoadConfig()

	db, err := scraper.InitDB("news.db")
	if err != nil {
		log.Fatal("failed to initialize db:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("failed to close db:", err)
		}
	}()

	if err := tui.Start(db, urls, selectors); err != nil {
		log.Fatal(err)
	}
}
