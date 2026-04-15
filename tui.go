package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type scrapeDoneMsg struct{}
type progressMsg float64

type model struct {
	filtered    []Article
	cursor      int
	db          *sql.DB
	page        int
	loading     bool
	progress    progress.Model
	urls        []string
	selectors   []selector
	width       int
	height      int
	searching   bool
	searchInput textinput.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		m.filtered = articles
		m.loading = false
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = min(m.width-4, 60)

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.searching {
			switch msg.String() {
			case "esc", "ctrl+c":
				m.searching = false
				m.searchInput.SetValue("")
				articles, err := getArticles(m.db, m.page)
				if err != nil {
					log.Println("failed to fetch articles :", err)
				}
				m.filtered = articles
				m.cursor = 0
				return m, nil
			case "down":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
				return m, nil
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				query := strings.ToLower(m.searchInput.Value())
				if query == "" {
					articles, err := getArticles(m.db, 0)
					if err != nil {
						log.Println("failed to fetch articles :", err)
					}
					m.filtered = articles
				} else {
					articles, err := searchArticles(m.db, query)
					if err != nil {
						log.Println("failed to fetch articles :", err)
					}
					m.filtered = articles
				}
				m.cursor = 0
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink
		case "j", "down":
			m.cursor = (m.cursor + 1) % len(m.filtered)
		case "k", "up":
			m.cursor = (m.cursor - 1 + len(m.filtered)) % len(m.filtered)
		case "n":
			if !m.searching {
				nextPage := m.page + 1
				articles, err := getArticles(m.db, nextPage)
				if err != nil || len(articles) == 0 {
					return m, nil
				}
				m.filtered = articles
				m.cursor = 0
				m.page = nextPage
				return m, nil
			}
		case "p":
			if !m.searching && m.page > 0 {
				prevPage := m.page - 1
				articles, err := getArticles(m.db, prevPage)
				if err != nil {
					return m, nil
				}
				m.filtered = articles
				m.cursor = 0
				m.page = prevPage
				return m, nil
			}
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

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFD0")).
			PaddingLeft(1)
)

func (m model) View() string {
	if m.loading {
		content := fmt.Sprintf("fetching latest news...\n\n%s", m.progress.View())
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(content)
	}

	var s strings.Builder

	s.WriteString(headerStyle.Render("▲ MARKET NEWS") + "\n\n")

	list := m.filtered
	for i, a := range list {
		title := hyperlink(a.Link, a.Title)
		if i == m.cursor {
			s.WriteString(selectedStyle.Render("\u25c9 "+title) + "\n\n")
		} else {
			s.WriteString(normalStyle.Render("  "+title) + "\n\n")
		}
	}

	if m.searching {
		s.WriteString(searchStyle.Render("/ " + m.searchInput.View()))
		s.WriteString(footerStyle.Render("\nesc quit ·\u2191/\u2193 navigate"))
	} else {
		s.WriteString(footerStyle.Render("j/k navigate · n/p page · / search · q quit"))
	}

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

	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 100
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFD0"))
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFD0"))

	m := model{
		db:          db,
		loading:     true,
		progress:    prog,
		urls:        urls,
		selectors:   selectors,
		searchInput: ti,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	go func() {
		c := newCollector(db, selectors)
		total := float64(len(urls))
		for i, url := range urls {
			if err := c.Visit(url); err != nil {
				log.Printf("error upon visiting %s: %s", url, err)
			}
			p.Send(progressMsg(float64(i+1) / total))
		}
		p.Send(scrapeDoneMsg{})
	}()

	_, err := p.Run()
	return err
}
