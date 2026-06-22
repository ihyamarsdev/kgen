package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Indigo  = lipgloss.Color("#5A4FCF")
	Magenta = lipgloss.Color("#EE49B6")
	Cyan    = lipgloss.Color("#00F2FE")
	Slate   = lipgloss.Color("#2D3748")
	Gray    = lipgloss.Color("#718096")
	Green   = lipgloss.Color("#00E676")
	Red     = lipgloss.Color("#FF1744")
	White   = lipgloss.Color("#FFFFFF")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Indigo).
			Padding(0, 2).
			Bold(true).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			MarginBottom(1)

	StepStyle = lipgloss.NewStyle().
			Foreground(Magenta).
			Bold(true)

	ActiveInputStyle = lipgloss.NewStyle().
				Foreground(Cyan).
				Bold(true)

	InactiveInputStyle = lipgloss.NewStyle().
				Foreground(Gray)

	GrayStyle = lipgloss.NewStyle().Foreground(Gray)

	FocusedButton = lipgloss.NewStyle().
			Foreground(White).
			Background(Magenta).
			Padding(0, 3).
			Bold(true)

	NormalButton = lipgloss.NewStyle().
			Foreground(White).
			Background(Slate).
			Padding(0, 3)

	CheckboxChecked = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	CheckboxUnchecked = lipgloss.NewStyle().
				Foreground(Gray)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true).
			MarginTop(1)

	SummaryBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Indigo).
			Padding(1, 2).
			MarginTop(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)
)
