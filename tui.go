package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const pageSize = 10

type scrapeDoneMsg struct{}
type progressMsg float64

type model struct {
	articles  []Article
	cursor    int
	db        *sql.DB
	page      int
	loading   bool
	progress  progress.Model
	urls      []string
	selectors []selector
	width     int
	height    int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = min(msg.Width-4, 60)

	case progressMsg:
		cmd := m.progress.SetPercent(float64(msg))
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case scrapeDoneMsg:
		articles, err := getArticles(m.db, 0)
		if err != nil {
			return m, nil
		}
		m.articles = articles
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
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
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFD0"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFD0")).
			PaddingLeft(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A5C5C")).
			PaddingLeft(3)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#004444")).
			MarginTop(1).
			PaddingLeft(1)
)

func (m model) View() string {
	if m.loading {
		content := fmt.Sprintf("\n  fetching latest news...\n\n  %s\n", m.progress.View())
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(content)
	}

	var s strings.Builder

	s.WriteString(headerStyle.Render("▲ MARKET NEWS") + "\n\n")

	for i, a := range m.articles {
		title := hyperlink(a.Link, a.Title)
		if i == m.cursor {
			s.WriteString(selectedStyle.Render("⬡ "+title) + "\n\n")
		} else {
			s.WriteString(normalStyle.Render("  "+title) + "\n\n")
		}
	}

	s.WriteString(footerStyle.Render("j/k navigate · n/p page · q quit"))
	return s.String()
}

func hyperlink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

func startTUI(db *sql.DB, urls []string, selectors []selector) error {
	prog := progress.New(
		progress.WithScaledGradient("#9B59B6", "#00FFD0"),
		progress.WithoutPercentage(),
	)

	m := model{
		db:        db,
		loading:   true,
		progress:  prog,
		urls:      urls,
		selectors: selectors,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	go func() {
		c := newCollector(db, selectors)
		total := float64(len(urls))
		for i, url := range urls {
			if err := c.Visit(url); err != nil {
				log.Printf("failed visiting %s\n", url)
			}
			p.Send(progressMsg(float64(i+1) / total))
		}
		p.Send(scrapeDoneMsg{})
	}()

	_, err := p.Run()
	return err
}
