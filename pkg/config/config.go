package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type Article struct {
	ID        int
	Title     string
	Link      string
	ScrapedAt string
}

type Selector struct {
	Container string
	Title     string
	Link      string
}

func LoadConfig() ([]string, []Selector) {
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

	var selectors []Selector
	for i := 1; ; i++ {
		container := os.Getenv(fmt.Sprintf("SELECTOR_%d_CONTAINER", i))
		if container == "" {
			break
		}
		selectors = append(selectors, Selector{
			Container: container,
			Title:     os.Getenv(fmt.Sprintf("SELECTOR_%d_TITLE", i)),
			Link:      os.Getenv(fmt.Sprintf("SELECTOR_%d_LINK", i)),
		})
	}

	return urls, selectors
}
