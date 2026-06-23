package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SelectorModel is a thin wrapper around ListModel for selecting files to edit.
// Kept for backwards compatibility — all logic is delegated to ListModel.
type SelectorModel struct {
	Files        []string
	Cursor       int
	SelectedFile string
	Quitted      bool

	list *ListModel
}

// InitialSelectorModel creates a SelectorModel pre-populated with the given file paths.
func InitialSelectorModel(files []string) SelectorModel {
	lm := NewListModel(files)
	lm.Title = "Edit Generated Files"
	lm.Header = "Select a file to view or edit in your terminal editor:"
	lm.FooterText = "Finish & Exit"
	lm.Help = "Use up/down to navigate. Enter to edit. Esc/q to exit."
	return SelectorModel{
		Files: files,
		list:  lm,
	}
}

func (m *SelectorModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	lm, cmd := m.list.Update(msg)
	m.list = lm
	m.Cursor = m.list.Cursor
	m.Quitted = m.list.Quitted
	if m.list.SelectedItem != "" {
		m.SelectedFile = m.list.SelectedItem
	}
	return m, cmd
}

func (m *SelectorModel) View() string {
	return m.list.View()
}
