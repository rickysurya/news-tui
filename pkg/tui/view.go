package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

	for i, a := range m.filtered {
		title := hyperlink(a.Link, a.Title)
		if i == m.cursor {
			s.WriteString(selectedStyle.Render("⬡ "+title) + "\n\n")
		} else {
			s.WriteString(normalStyle.Render("  "+title) + "\n\n")
		}
	}

	if m.searching {
		s.WriteString(searchStyle.Render("/ " + m.searchInput.View()))
		s.WriteString(footerStyle.Render("\u2191/\u2193 navigate · esc quit"))
	} else {
		s.WriteString(footerStyle.Render("j/k navigate · n/p page · / search · q quit"))
	}

	return s.String()
}

func hyperlink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// satisfy the tea.Model interface — Update is in model.go
var _ tea.Model = model{}
