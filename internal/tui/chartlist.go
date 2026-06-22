package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChartListModel is a Bubble Tea model that presents an interactive vertical
// list of Helm chart directories for the user to choose from.
//
// It follows the same visual pattern as SelectorModel: a titled box, cursor
// navigation with up/k and down/j, enter to confirm, and a virtual "Cancel"
// button at the bottom of the list.
//
// All methods use pointer receivers per AGENTS.md guidelines.
type ChartListModel struct {
	Charts        []string // chart folder names (not full paths)
	Cursor        int
	SelectedChart string // set when the user presses enter on a chart
	Quitted       bool
}

// InitialChartListModel creates a ChartListModel pre-populated with the
// given chart names.
func InitialChartListModel(charts []string) ChartListModel {
	return ChartListModel{
		Charts:  charts,
		Cursor:  0,
	}
}

func (m *ChartListModel) Init() tea.Cmd {
	return nil
}

func (m *ChartListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.Cursor = len(m.Charts)
			}

		case "down", "j":
			if m.Cursor < len(m.Charts) {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "enter":
			if m.Cursor == len(m.Charts) {
				// Virtual "Cancel" button
				m.Quitted = true
				return m, tea.Quit
			}
			m.SelectedChart = m.Charts[m.Cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ChartListModel) View() string {
	if m.Quitted && m.SelectedChart == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(TitleStyle.Render(" Select a Helm Chart "))
	sb.WriteString("\n\n")
	sb.WriteString(HeaderStyle.Render("Choose a chart to view or edit its files:"))
	sb.WriteString("\n\n")

	for i, chart := range m.Charts {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(White)
		if m.Cursor == i {
			cursor = ActiveInputStyle.Render("> ")
			style = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(chart)))
	}

	cursor := "  "
	btnStyle := NormalButton
	if m.Cursor == len(m.Charts) {
		cursor = ActiveInputStyle.Render("> ")
		btnStyle = FocusedButton
	}
	sb.WriteString("\n" + cursor + btnStyle.Render("Cancel") + "\n")

	sb.WriteString(HelpStyle.Render("\nUse up/down to navigate. Enter to select. Esc/q to cancel."))

	return sb.String()
}
