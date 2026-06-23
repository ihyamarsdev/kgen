package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListModel is a reusable Bubble Tea model for presenting a vertical
// cursor-based list with optional footer button.
//
// Both SelectorModel (file picker) and ChartListModel (chart folder picker)
// share ~90% identical logic. This type eliminates the duplication while
// keeping the existing SelectorModel and ChartListModel APIs intact as thin
// wrappers.
//
// All methods use pointer receivers per AGENTS.md guidelines.
type ListModel struct {
	Items         []string
	Cursor        int
	SelectedIndex int // -1 when no item selected, index of selected item otherwise
	SelectedItem  string
	Quitted       bool
	ItemCount     int // len(Items) — virtual items don't count

	// Optional footer button.
	FooterText     string
	ActiveFooter   lipgloss.Style
	InactiveFooter lipgloss.Style

	// Display options.
	Title         string
	Header        string
	Help          string
	ActiveStyle   lipgloss.Style
	InactiveStyle lipgloss.Style
	CursorStyle   lipgloss.Style
}

// NewListModel creates a ListModel pre-populated with the given items.
func NewListModel(items []string) *ListModel {
	return &ListModel{
		Items:          items,
		ItemCount:      len(items),
		Cursor:         0,
		SelectedIndex:  -1,
		ActiveFooter:   FocusedButton,
		InactiveFooter: NormalButton,
		ActiveStyle:    lipgloss.NewStyle().Foreground(Cyan).Bold(true),
		InactiveStyle:  lipgloss.NewStyle().Foreground(White),
		CursorStyle:    ActiveInputStyle,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) Update(msg tea.Msg) (*ListModel, tea.Cmd) {
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
				m.Cursor = m.ItemCount
			}

		case "down", "j":
			if m.Cursor < m.ItemCount {
				m.Cursor++
			} else {
				m.Cursor = 0
			}

		case "enter":
			if m.Cursor == m.ItemCount {
				// Virtual footer button
				m.Quitted = true
				return m, tea.Quit
			}
			m.SelectedIndex = m.Cursor
			m.SelectedItem = m.Items[m.Cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ListModel) View() string {
	if m.Quitted && m.SelectedItem == "" {
		return ""
	}

	var sb strings.Builder

	if m.Title != "" {
		sb.WriteString(TitleStyle.Render(" " + m.Title + " "))
		sb.WriteString("\n\n")
	}
	if m.Header != "" {
		sb.WriteString(HeaderStyle.Render(m.Header))
		sb.WriteString("\n\n")
	}

	for i, item := range m.Items {
		cursor := "  "
		style := m.InactiveStyle
		if m.Cursor == i {
			cursor = m.CursorStyle.Render("> ")
			style = m.ActiveStyle
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item)))
	}

	// Footer button (optional).
	if m.FooterText != "" {
		cursor := "  "
		btnStyle := m.InactiveFooter
		if m.Cursor == m.ItemCount {
			cursor = m.CursorStyle.Render("> ")
			btnStyle = m.ActiveFooter
		}
		sb.WriteString("\n" + cursor + btnStyle.Render(m.FooterText) + "\n")
	}

	if m.Help != "" {
		sb.WriteString(HelpStyle.Render("\n" + m.Help))
	}

	return sb.String()
}
