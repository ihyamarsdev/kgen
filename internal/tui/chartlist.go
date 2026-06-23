package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ChartListModel is a thin wrapper around ListModel for selecting Helm chart
// directories. Kept for backwards compatibility — all logic is delegated to
// ListModel.
type ChartListModel struct {
	Charts        []string
	Cursor        int
	SelectedChart string
	Quitted       bool

	list *ListModel
}

// InitialChartListModel creates a ChartListModel pre-populated with the
// given chart names.
func InitialChartListModel(charts []string) ChartListModel {
	lm := NewListModel(charts)
	lm.Title = "Select a Helm Chart"
	lm.Header = "Choose a chart to view or edit its files:"
	lm.FooterText = "Cancel"
	lm.Help = "Use up/down to navigate. Enter to select. Esc/q to cancel."
	return ChartListModel{
		Charts: charts,
		list:   lm,
	}
}

func (m *ChartListModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ChartListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	lm, cmd := m.list.Update(msg)
	m.list = lm
	m.Cursor = m.list.Cursor
	m.Quitted = m.list.Quitted
	if m.list.SelectedItem != "" {
		m.SelectedChart = m.list.SelectedItem
	}
	return m, cmd
}

func (m *ChartListModel) View() string {
	return m.list.View()
}
