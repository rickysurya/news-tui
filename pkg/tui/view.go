package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	boxWidth   = 80
	boxHeight  = 22
	titleWidth = 74
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#1EDF6F")).
			Padding(0, 1).
			Width(boxWidth).
			Height(boxHeight)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#39FF6F")).
			PaddingLeft(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#39FF6F")).
			PaddingLeft(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F7A3C")).
			PaddingLeft(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C5955")).
			MarginTop(1).
			PaddingLeft(1)

	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#1EDF6F")).
			Padding(0, 1).
			Width(boxWidth).
			MarginBottom(1)

	searchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1EDF6F"))

	noResultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C5955")).
			PaddingLeft(1).
			Italic(true)
)

func truncate(s string, max int) string {
	if len([]rune(s)) <= max {
		return s
	}
	return string([]rune(s)[:max-3]) + "..."
}
func (m model) View() string {
	if m.loading {
		content := fmt.Sprintf("fetching latest news...\n\n%s", m.progress.View())
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(content)
	}

	var articles strings.Builder
	if len(m.filtered) == 0 {
		articles.WriteString(noResultStyle.Render("no results") + "\n")
	} else {
		for i, a := range m.filtered {
			title := hyperlink(a.Link, truncate(a.Title, titleWidth))
			if i == m.cursor {
				articles.WriteString(selectedStyle.Render(title) + "\n\n")
			} else {
				articles.WriteString(normalStyle.Render(title) + "\n\n")
			}
			if i == len(m.filtered)-1 {
				remaining := 10 - len(m.filtered)
				for range make([]struct{}, remaining) {
					articles.WriteString("\n\n")
				}
			}
		}
	}

	var s strings.Builder
	s.WriteString(headerStyle.Render("▲ MARKET NEWS") + "\n\n")
	s.WriteString(boxStyle.Render(articles.String()) + "\n")

	if m.searching {
		s.WriteString(searchBoxStyle.Render(
			searchLabelStyle.Render("  ")+m.searchInput.View(),
		) + "\n")
		s.WriteString(footerStyle.Render("\u2191/\u2193 up/down · esc quit"))
	} else {
		s.WriteString(footerStyle.Render("j/\u2193 down · k/\u2191 up · n/\u2192 next · p/\u2190 previous · / search · q quit"))
	}
	return s.String()
}

func hyperlink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// satisfy the tea.Model interface; Update is in model.go
var _ tea.Model = model{}
