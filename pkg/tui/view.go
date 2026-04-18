package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	boxWidth   = 70
	boxHeight  = 22
	titleWidth = 66
	listLines  = 20
)

var (
	boxStyle = lipgloss.NewStyle().
		// top, right, bottom, left
		Border(lipgloss.NormalBorder(), true, true, true, true).
		BorderForeground(lipgloss.Color("#1EDF6F")).
		Width(boxWidth)
		// Height(boxHeight)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1EDF6F"))

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

	searchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#39FF6F"))

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
		articles.WriteString(noResultStyle.Render("no results"))
		articles.WriteString(strings.Repeat("\n", listLines))
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

	divider := dividerStyle.Render(strings.Repeat("\u2500", boxWidth))
	searchArea := searchLabelStyle.Render("\uf002  ") + m.searchInput.View()

	leftPane := lipgloss.JoinVertical(lipgloss.Left, searchArea, divider, articles.String())
	s.WriteString(boxStyle.Render(leftPane) + "\n")

	if m.searching {
		s.WriteString(footerStyle.Render("enter open browser · \u25b2/\u25bc navigate · esc cancel"))
	} else {
		s.WriteString(footerStyle.Render("j/\u25bc down · k/\u25b2 up · n/\u25b6 next · p/\u25c0 previous · / search · q quit"))
	}
	return s.String()
}

func hyperlink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// satisfy the tea.Model interface; Update is in model.go
var _ tea.Model = model{}
