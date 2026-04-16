package tui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rickysurya/news-tui/pkg/config"
	"github.com/rickysurya/news-tui/pkg/scraper"
)

const pageSize = 10

type scrapeDoneMsg struct{}
type progressMsg float64

type model struct {
	filtered    []config.Article
	cursor      int
	db          *sql.DB
	page        int
	loading     bool
	progress    progress.Model
	urls        []string
	selectors   []config.Selector
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
		articles, err := scraper.GetArticles(m.db, 0)
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
			case "esc":
				m.searching = false
				m.searchInput.SetValue("")
				articles, _ := scraper.GetArticles(m.db, m.page)
				m.filtered = articles
				m.cursor = 0
				return m, nil
			case "j", "down":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
				return m, nil
			case "k", "up":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				query := m.searchInput.Value()
				if query == "" {
					m.filtered, _ = scraper.GetArticles(m.db, 0)
				} else {
					m.filtered, _ = scraper.SearchArticles(m.db, query)
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
			articles, err := scraper.GetArticles(m.db, m.page+1)
			if err != nil || len(articles) == 0 {
				return m, nil
			}
			m.page++
			m.filtered = articles
			m.cursor = 0
			return m, nil
		case "p":
			if m.page == 0 {
				return m, nil
			}
			articles, err := scraper.GetArticles(m.db, m.page-1)
			if err != nil {
				return m, nil
			}
			m.page--
			m.filtered = articles
			m.cursor = 0
			return m, nil
		}
	}
	return m, nil
}

func Start(db *sql.DB, urls []string, selectors []config.Selector) error {
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
		c := scraper.NewCollector(db, selectors)
		total := float64(len(urls))
		for i, url := range urls {
			c.Visit(url)
			p.Send(progressMsg(float64(i+1) / total))
		}
		p.Send(scrapeDoneMsg{})
	}()

	_, err := p.Run()
	return err
}
