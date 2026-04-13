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
	ID        int
	Title     string
	Link      string
	ScrapedAt string
}

type selector struct {
	container string
	title     string
	link      string
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

func getArticles(db *sql.DB, page int) ([]Article, error) {
	const limit = 10
	rows, err := db.Query(
		`SELECT id, title, link, scraped_at FROM articles ORDER BY scraped_at DESC LIMIT ? OFFSET ?`,
		limit, page*limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Link, &a.ScrapedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
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

func scrape(db *sql.DB, urls []string, selectors []selector) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	for _, s := range selectors {
		registerHandler(c, db, s)
	}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	for _, url := range urls {
		if err := c.Visit(url); err != nil {
			log.Printf("Failed to visit %s: %v\n", url, err)
		}
	}
}

func newCollector(db *sql.DB, selectors []selector) *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	for _, s := range selectors {
		registerHandler(c, db, s)
	}
	return c
}

func main() {
	urls, selectors := loadConfig()

	db, err := initDB("news.db")
	if err != nil {
		log.Fatal("Failed to initialize db:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("failed to close db:", err)
		}
	}()

	if err := startTUI(db, urls, selectors); err != nil {
		log.Fatal(err)
	}
}
