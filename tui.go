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
		}
	}
	return m, nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFD0")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#004444")).
			MarginBottom(1).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFD0")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color("#00FFD0")).
			PaddingLeft(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#008B8B")).
			PaddingLeft(2)

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#004444")).
			PaddingLeft(3).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#004444")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("#004444")).
			MarginTop(1).
			PaddingLeft(1)
)

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("▲ MARKET NEWS") + "\n")
	s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#004444")).Render("=======================================") + "\n\n")
	for i, a := range m.articles {
		title := hyperlink(a.Link, a.Title)
		if i == m.cursor {
			s.WriteString(selectedStyle.Render("> "+title) + "\n")
		} else {
			s.WriteString(normalStyle.Render("  "+title) + "\n")
		}
		s.WriteString(metaStyle.Render(a.ScrapedAt) + "\n\n")
	}

	s.WriteString(helpStyle.Render("> j/k navigate · q quit"))
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
		model{articles: articles, db: db},
		tea.WithAltScreen(),
	)
	_, err = p.Run()
	return err
}
