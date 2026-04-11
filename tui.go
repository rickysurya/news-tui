package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	articles []Article
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
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Latest Articles\n\n"
	for _, a := range m.articles {
		s += fmt.Sprintf("%s\n%s\n\n", a.Title, a.Link)
	}
	s += "\npress q to quit"
	return s
}

// func main() {
// 	//mock articles
// 	mock_articles := []Article{
// 		{Title: "article 1", Link: "https://article1.com"},
// 		{Title: "article 2", Link: "https://article2.com"},
// 	}
//
// 	p := tea.NewProgram(model{articles: mock_articles})
// 	if _, err := p.Run(); err != nil {
// 		fmt.Println("error:", err)
// 		os.Exit(1)
// 	}
// }
