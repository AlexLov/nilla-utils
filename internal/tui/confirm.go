package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	message string
	done    bool
	answer  bool
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.done = true
			m.answer = true
		default:
			m.done = true
			m.answer = false
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}

	return fmt.Sprintf("\n%s [y/n]", m.message)
}

func RunConfirm(msg string) (bool, error) {
	// Initialize model
	init := confirmModel{
		message: msg,
	}

	// Create a bubbletea program
	p := tea.NewProgram(
		init,
		tea.WithOutput(os.Stderr),
	)

	// Run program
	m, err := p.Run()
	if err != nil {
		return false, err
	}

	model := m.(confirmModel)

	return model.answer, nil
}
