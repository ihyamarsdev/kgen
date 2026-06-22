package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SelectorModel struct {
	Files        []string
	Cursor       int
	SelectedFile string
	Quitted      bool
}

func InitialSelectorModel(files []string) SelectorModel {
	return SelectorModel{
		Files:  files,
		Cursor: 0,
	}
}

func (m *SelectorModel) Init() tea.Cmd {
	return nil
}

func (m *SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Quitted = true
			return m, tea.Quit

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			} else {
				m.Cursor = len(m.Files)
			}

		case "down", "j":
			if m.Cursor < len(m.Files) {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "enter":
			if m.Cursor == len(m.Files) {
				m.Quitted = true
				return m, tea.Quit
			}
			m.SelectedFile = m.Files[m.Cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *SelectorModel) View() string {
	if m.Quitted && m.SelectedFile == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" Edit Generated Files "))
	sb.WriteString("\n\n")
	sb.WriteString(HeaderStyle.Render("Select a file to view or edit in your terminal editor:"))
	sb.WriteString("\n\n")

	for i, f := range m.Files {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(White)
		if m.Cursor == i {
			cursor = ActiveInputStyle.Render("> ")
			style = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(f)))
	}

	cursor := "  "
	btnStyle := NormalButton
	if m.Cursor == len(m.Files) {
		cursor = ActiveInputStyle.Render("> ")
		btnStyle = FocusedButton
	}
	sb.WriteString("\n" + cursor + btnStyle.Render("Finish & Exit") + "\n")

	sb.WriteString(HelpStyle.Render("\nUse up/down to navigate. Enter to edit. Esc/q to exit."))

	return sb.String()
}
