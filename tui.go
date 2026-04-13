package main

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	articles []Article
	cursor   int
	db       *sql.DB
	page     int
}

func (m model) Init() tea.Cmd {
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.cursor = (m.cursor + 1) % len(m.articles)
		case "k", "up":
			m.cursor = (m.cursor - 1 + len(m.articles)) % len(m.articles)
		case "n":
			articles, err := getArticles(m.db, m.page+1)
			if err != nil || len(articles) == 0 {
				return m, nil
			}
			m.page++
			m.articles = articles
			m.cursor = 0
			return m, nil
		case "p":
			if m.page == 0 {
				return m, nil
			}
			articles, err := getArticles(m.db, m.page-1)
			if err != nil {
				return m, nil
			}
			m.page--
			m.articles = articles
			m.cursor = 0
			return m, nil
		}
	}
	return m, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFD0")).
			MarginBottom(1).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFD0")).
			PaddingLeft(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D6B6B")).
			PaddingLeft(3)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#004444")).
			MarginTop(1).
			PaddingLeft(1)
)

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("▲ MARKET NEWS") + "\n\n")

	for i, a := range m.articles {
		title := hyperlink(a.Link, a.Title)
		if i == m.cursor {
			s.WriteString(selectedStyle.Render("⬡ "+title) + "\n\n")
		} else {
			s.WriteString(normalStyle.Render("  "+title) + "\n\n")
		}
	}
	s.WriteString(helpStyle.Render("  j/k navigate · n/p page · q quit"))
	return s.String()
}

func hyperlink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}
func startTUI(db *sql.DB) error {
	articles, err := getArticles(db, 0)
	if err != nil {
		return err
	}
	p := tea.NewProgram(
		model{articles: articles, db: db, page: 0},
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	return err
}
