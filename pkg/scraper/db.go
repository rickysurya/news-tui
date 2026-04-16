package scraper

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rickysurya/news-tui/pkg/config"
)

func InitDB(path string) (*sql.DB, error) {
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

func SaveArticle(db *sql.DB, a config.Article) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO articles(title, link) VALUES(?, ?)`,
		a.Title, a.Link,
	)
	return err
}

func GetArticles(db *sql.DB, page int) ([]config.Article, error) {
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

	var articles []config.Article
	for rows.Next() {
		var a config.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Link, &a.ScrapedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func SearchArticles(db *sql.DB, query string) ([]config.Article, error) {
	rows, err := db.Query(
		`SELECT id, title, link, scraped_at FROM articles WHERE title LIKE ? ORDER BY scraped_at DESC LIMIT 10`,
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("failed to close rows:", err)
		}
	}()

	var articles []config.Article
	for rows.Next() {
		var a config.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Link, &a.ScrapedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}
