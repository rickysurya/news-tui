package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	_ "github.com/mattn/go-sqlite3"
)

type Article struct {
	Title string
	Link  string
}

type selector struct {
	container string
	title     string
	link      string
}

func (a Article) Print() {
	fmt.Println("Headline:", a.Title)
	fmt.Println("Link:", a.Link)
	fmt.Println("--------------------------")
}

func loadConfig() ([]string, []selector) {
	f, err := os.Open(".env")
	if err != nil {
		log.Fatal("no .env file found")
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Println("failed to close .env", err)
		}
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			if err := os.Setenv(parts[0], parts[1]); err != nil {
				log.Fatalf("failed to set env %s: %v", parts[0], err)
			}
		}
	}

	var urls []string
	for i := 1; ; i++ {
		url := os.Getenv(fmt.Sprintf("URL_%d", i))
		if url == "" {
			break
		}
		urls = append(urls, url)
	}

	var selectors []selector
	for i := 1; ; i++ {
		container := os.Getenv(fmt.Sprintf("SELECTOR_%d_CONTAINER", i))
		if container == "" {
			break
		}
		selectors = append(selectors, selector{
			container: container,
			title:     os.Getenv(fmt.Sprintf("SELECTOR_%d_TITLE", i)),
			link:      os.Getenv(fmt.Sprintf("SELECTOR_%d_LINK", i)),
		})
	}

	return urls, selectors
}

func initDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			title      TEXT NOT NULL,
			link       TEXT NOT NULL UNIQUE,
			scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return db, err
}

func saveArticle(db *sql.DB, a Article) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO articles(title, link) VALUES(?, ?)`,
		a.Title, a.Link,
	)
	return err
}

func showLatestArticles(db *sql.DB, limit int) {
	rows, err := db.Query(
		`SELECT title, link, scraped_at FROM articles ORDER BY scraped_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		log.Println("failed to fetch articles from db:", err)
		return
	}

	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}()

	fmt.Printf("===== Latest %d Articles =====\n", limit)
	for rows.Next() {
		var title, link, scrapedAt string
		if err := rows.Scan(&title, &link, &scrapedAt); err != nil {
			log.Println("row scan error:", err)
			continue
		}
		fmt.Printf("[%s] %s\n%s\n--------------------------\n", scrapedAt, title, link)
	}
}

func registerHandler(c *colly.Collector, db *sql.DB, s selector) {
	c.OnHTML(s.container, func(e *colly.HTMLElement) {
		article := Article{
			Title: e.ChildText(s.title),
			Link:  e.Request.AbsoluteURL(e.ChildAttr(s.link, "href")),
		}
		if article.Title == "" || article.Link == "" {
			return
		}
		if err := saveArticle(db, article); err != nil {
			log.Println("failed to save article to db:", err)
		}
	})
}

func main() {
	urls, selectors := loadConfig()

	db, err := initDB("news.db")
	if err != nil {
		log.Fatal("failed to initialize db:", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Println("failed to close db:", err)
		}
	}()

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	for _, s := range selectors {
		registerHandler(c, db, s)
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	for _, url := range urls {
		if err := c.Visit(url); err != nil {
			log.Printf("failed to visit %s: %v\n", url, err)
		}
	}

	showLatestArticles(db, 20)
}
