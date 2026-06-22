package tui

import (
	"fmt"
	"strconv"
	"strings"

	"kgen/internal/generator"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Step int

const (
	StepAppInfo Step = iota
	StepMode
	StepCustomResources
	StepConfirm
	StepDone
)

type Mode int

const (
	ModeSimple Mode = iota
	ModeAdvanced
	ModeCustom
)

type WizardModel struct {
	Step        Step
	Profile     string
	OutputDir   string

	// Step 1: Inputs
	Inputs      []textinput.Model
	ActiveInput int

	// Step 2: Mode selection
	SelectedMode Mode

	// Step 3: Custom checkboxes
	Resources   []string
	SelectedRes map[string]bool
	ResCursor   int

	// Confirmation button focus (0: Generate, 1: Cancel)
	ConfirmCursor int

	// Result parameters to return to generator
	Confirmed bool
	Quitted   bool
	ShowHelp  bool
}

func InitialModel(profile string, defaultAppName string) WizardModel {
	m := WizardModel{
		Step:          StepAppInfo,
		Profile:       profile,
		SelectedMode:  ModeSimple,
		Resources:     []string{"Deployment", "Service", "Ingress", "HPA"},
		SelectedRes:   map[string]bool{"Deployment": true, "Service": true},
		ResCursor:     0,
		ConfirmCursor: 0,
	}

	// Adjust initial selected mode and resources if profile is prod
	if profile == "prod" {
		m.SelectedMode = ModeAdvanced
		m.SelectedRes["Ingress"] = true
		m.SelectedRes["HPA"] = true
	}

	// Initialize inputs
	m.Inputs = make([]textinput.Model, 4)

	// App Name
	m.Inputs[0] = textinput.New()
	m.Inputs[0].Placeholder = "my-app"
	m.Inputs[0].SetValue(defaultAppName)
	m.Inputs[0].Focus()
	m.Inputs[0].PromptStyle = ActiveInputStyle

	// Namespace
	m.Inputs[1] = textinput.New()
	m.Inputs[1].Placeholder = "default"
	m.Inputs[1].SetValue("default")
	m.Inputs[1].PromptStyle = InactiveInputStyle

	// Container Image
	m.Inputs[2] = textinput.New()
	m.Inputs[2].Placeholder = "nginx:latest"
	m.Inputs[2].SetValue("nginx:latest")
	m.Inputs[2].PromptStyle = InactiveInputStyle

	// Container Port
	m.Inputs[3] = textinput.New()
	m.Inputs[3].Placeholder = "80"
	m.Inputs[3].SetValue("80")
	m.Inputs[3].PromptStyle = InactiveInputStyle

	return m
}

func (m *WizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.ShowHelp {
			m.ShowHelp = false
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitted = true
			return m, tea.Quit

		case "esc":
			if m.Step > StepAppInfo {
				// Go back a step
				if m.Step == StepConfirm && m.SelectedMode != ModeCustom {
					m.Step = StepMode
				} else {
					m.Step--
				}
				return m, nil
			}
			m.Quitted = true
			return m, tea.Quit

		case "?", "h":
			if m.Step != StepAppInfo {
				m.ShowHelp = true
				return m, nil
			}
		}
	}

	// Step specific updates
	switch m.Step {
	case StepAppInfo:
		return m.updateAppInfo(msg)
	case StepMode:
		return m.updateMode(msg)
	case StepCustomResources:
		return m.updateCustomResources(msg)
	case StepConfirm:
		return m.updateConfirm(msg)
	}

	return m, nil
}

func (m *WizardModel) updateAppInfo(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.ActiveInput == len(m.Inputs)-1 {
				// Validate and go to next step
				appName := strings.TrimSpace(m.Inputs[0].Value())
				if appName == "" {
					m.Inputs[0].SetValue("my-app")
				}
				m.Step = StepMode
				return m, nil
			}
			m.nextInput()

		case "up", "shift+tab":
			m.prevInput()

		case "down", "tab":
			m.nextInput()
		}
	}

	// Update active textinput
	var cmd tea.Cmd
	m.Inputs[m.ActiveInput], cmd = m.Inputs[m.ActiveInput].Update(msg)
	return m, cmd
}

func (m *WizardModel) nextInput() {
	m.Inputs[m.ActiveInput].Blur()
	m.Inputs[m.ActiveInput].PromptStyle = InactiveInputStyle
	m.ActiveInput = (m.ActiveInput + 1) % len(m.Inputs)
	m.Inputs[m.ActiveInput].Focus()
	m.Inputs[m.ActiveInput].PromptStyle = ActiveInputStyle
}

func (m *WizardModel) prevInput() {
	m.Inputs[m.ActiveInput].Blur()
	m.Inputs[m.ActiveInput].PromptStyle = InactiveInputStyle
	m.ActiveInput = (m.ActiveInput - 1 + len(m.Inputs)) % len(m.Inputs)
	m.Inputs[m.ActiveInput].Focus()
	m.Inputs[m.ActiveInput].PromptStyle = ActiveInputStyle
}

func (m *WizardModel) updateMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedMode > ModeSimple {
				m.SelectedMode--
			} else {
				m.SelectedMode = ModeCustom
			}

		case "down", "j":
			if m.SelectedMode < ModeCustom {
				m.SelectedMode++
			} else {
				m.SelectedMode = ModeSimple
			}

		case "enter":
			// Set defaults based on mode selection
			switch m.SelectedMode {
			case ModeSimple:
				m.SelectedRes = map[string]bool{
					"Deployment": true,
					"Service":    true,
					"Ingress":    false,
					"HPA":        false,
				}
				m.Step = StepConfirm
			case ModeAdvanced:
				m.SelectedRes = map[string]bool{
					"Deployment": true,
					"Service":    true,
					"Ingress":    true,
					"HPA":        true,
				}
				m.Step = StepConfirm
			case ModeCustom:
				m.Step = StepCustomResources
			}
		}
	}
	return m, nil
}

func (m *WizardModel) updateCustomResources(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.ResCursor > 0 {
				m.ResCursor--
			} else {
				m.ResCursor = len(m.Resources) - 1
			}

		case "down", "j":
			if m.ResCursor < len(m.Resources)-1 {
				m.ResCursor++
			} else {
				m.ResCursor = 0
			}

		case " ":
			res := m.Resources[m.ResCursor]
			m.SelectedRes[res] = !m.SelectedRes[res]

		case "enter":
			m.Step = StepConfirm
		}
	}
	return m, nil
}

func (m *WizardModel) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "right", "l":
			m.ConfirmCursor = 1 - m.ConfirmCursor

		case "enter":
			if m.ConfirmCursor == 0 {
				m.Confirmed = true
			} else {
				m.Quitted = true
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *WizardModel) View() string {
	if m.Quitted {
		return "\n  Operation cancelled.\n\n"
	}

	if m.ShowHelp {
		var hsb strings.Builder
		hsb.WriteString(TitleStyle.Render(" KGen TUI Wizard Help "))
		hsb.WriteString("\n\n")
		hsb.WriteString(HeaderStyle.Render("Keyboard Shortcuts:"))
		hsb.WriteString("\n")

		shortcuts := []struct {
			keys, desc string
		}{
			{"Tab / Shift+Tab", "Navigate text inputs (Step 1)"},
			{"Up / Down (k / j)", "Navigate through list options"},
			{"Space", "Toggle checklist selection (Step 3)"},
			{"Enter", "Advance to next step or confirm actions"},
			{"Esc", "Go back to the previous step (or exit on Step 1)"},
			{"?", "Toggle this help menu"},
			{"q / Ctrl+C", "Cancel generation and quit"},
		}

		for _, s := range shortcuts {
			hsb.WriteString(fmt.Sprintf("  %-18s : %s\n", ActiveInputStyle.Render(s.keys), s.desc))
		}

		hsb.WriteString("\n")
		hsb.WriteString(HelpStyle.Render("Press any key to close help and return to the wizard."))
		return hsb.String()
	}

	var sb strings.Builder

	// Header banner
	sb.WriteString(TitleStyle.Render(" KGen — Helm Chart Generator "))
	sb.WriteString("\n")

	// Progress indicator
	sb.WriteString(m.renderProgress())
	sb.WriteString("\n\n")

	switch m.Step {
	case StepAppInfo:
		sb.WriteString(HeaderStyle.Render("Step 1: Application Information"))
		sb.WriteString("\n")
		labels := []string{"Application Name  :", "Namespace         :", "Container Image   :", "Container Port    :"}
		for i, input := range m.Inputs {
			prefix := "  "
			if i == m.ActiveInput {
				prefix = ActiveInputStyle.Render("> ")
			}
			sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, labels[i], input.View()))
		}
		sb.WriteString(HelpStyle.Render("Use up/down or tab to navigate. Enter to continue."))

	case StepMode:
		sb.WriteString(HeaderStyle.Render("Step 2: Choose Deployment Mode"))
		sb.WriteString("\n")
		modes := []struct {
			name, desc string
		}{
			{"Simple", "Generates Deployment and Service (ideal for dev/testing)"},
			{"Advanced", "Generates Deployment, Service, Ingress, and HPA (production-ready)"},
			{"Custom", "Pick resources individually"},
		}

		for i, mode := range modes {
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(White)
			if int(m.SelectedMode) == i {
				cursor = ActiveInputStyle.Render("> ")
				nameStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%-10s : %s\n", cursor, nameStyle.Render(mode.name), GrayStyle.Render(mode.desc)))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to select. Esc to go back. (Press '?' for help)"))

	case StepCustomResources:
		sb.WriteString(HeaderStyle.Render("Step 3: Select Resources to Generate"))
		sb.WriteString("\n")
		for i, res := range m.Resources {
			cursor := "  "
			resStyle := lipgloss.NewStyle().Foreground(White)
			if m.ResCursor == i {
				cursor = ActiveInputStyle.Render("> ")
				resStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}

			checked := " "
			if m.SelectedRes[res] {
				checked = CheckboxChecked.Render("✓")
			} else {
				checked = CheckboxUnchecked.Render(" ")
			}

			sb.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, checked, resStyle.Render(res)))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to navigate. Space to select/deselect. Enter to confirm. Esc to go back. (Press '?' for help)"))

	case StepConfirm:
		sb.WriteString(HeaderStyle.Render("Step 4: Confirm Generation Settings"))
		sb.WriteString("\n")

		var resList []string
		for _, r := range m.Resources {
			if m.SelectedRes[r] {
				resList = append(resList, r)
			}
		}

		summary := fmt.Sprintf(
			"Application Name  : %s\nNamespace         : %s\nContainer Image   : %s\nContainer Port    : %s\nProfile           : %s\nResources         : %s",
			m.Inputs[0].Value(),
			m.Inputs[1].Value(),
			m.Inputs[2].Value(),
			m.Inputs[3].Value(),
			m.Profile,
			strings.Join(resList, ", "),
		)
		sb.WriteString(SummaryBox.Render(summary))
		sb.WriteString("\n\n")

		// Buttons
		genBtn := NormalButton.Render("Generate")
		cancelBtn := NormalButton.Render("Cancel")

		if m.ConfirmCursor == 0 {
			genBtn = FocusedButton.Render("Generate")
		} else {
			cancelBtn = FocusedButton.Render("Cancel")
		}

		sb.WriteString(fmt.Sprintf("  %s   %s\n", genBtn, cancelBtn))
		sb.WriteString(HelpStyle.Render("Use left/right to choose button. Enter to execute. Esc to go back. (Press '?' for help)"))
	}

	return sb.String()
}

func (m *WizardModel) renderProgress() string {
	steps := []string{"App Info", "Mode", "Resources", "Confirm"}
	var rendered []string

	currentStepIndex := int(m.Step)
	if m.Step == StepConfirm {
		currentStepIndex = 3
	} else if m.Step == StepCustomResources {
		currentStepIndex = 2
	}

	for i, s := range steps {
		if i == currentStepIndex {
			rendered = append(rendered, StepStyle.Render(fmt.Sprintf("[%s]", s)))
		} else if i < currentStepIndex {
			rendered = append(rendered, SuccessStyle.Render(s))
		} else {
			rendered = append(rendered, GrayStyle.Render(s))
		}
	}
	return "  " + strings.Join(rendered, " ── ")
}

func (m *WizardModel) GetConfig() (generator.Config, string) {
	port, _ := strconv.Atoi(m.Inputs[3].Value())
	if port <= 0 {
		port = 80
	}

	repoAndTag := m.Inputs[2].Value()
	repo := repoAndTag
	tag := "latest"
	if idx := strings.Index(repoAndTag, ":"); idx != -1 {
		repo = repoAndTag[:idx]
		tag = repoAndTag[idx+1:]
	}

	replicaCount := 1
	if m.Profile == "prod" {
		replicaCount = 3
	}

	cfg := generator.Config{
		AppName:            m.Inputs[0].Value(),
		Namespace:          m.Inputs[1].Value(),
		ImageRepository:    repo,
		ImageTag:           tag,
		Port:               port,
		ReplicaCount:       replicaCount,
		IngressEnabled:     m.SelectedRes["Ingress"],
		HPAEnabled:         m.SelectedRes["HPA"],
		HPAMinReplicas:     replicaCount,
		HPAMaxReplicas:     replicaCount * 3,
		ProdProfile:        m.Profile == "prod",
		GenerateDeployment: m.SelectedRes["Deployment"],
		GenerateService:    m.SelectedRes["Service"],
		GenerateIngress:    m.SelectedRes["Ingress"],
		GenerateHPA:        m.SelectedRes["HPA"],
	}

	return cfg, cfg.AppName
}
