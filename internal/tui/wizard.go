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
	StepQuality
	StepCustomResources
	StepSecretBackend
	StepConfirm
	StepDone
)

type Mode int

const (
	ModeSimple Mode = iota
	ModeAdvanced
	ModeCustom
)

type Quality int

const (
	QualityBasic Quality = iota
	QualityProduction
	QualityEnterprise
)

type Backend int

const (
	BackendVault Backend = iota
	BackendAWS
	BackendGCP
	BackendAzure
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

	// Step 2.5: Template Quality
	SelectedQuality Quality

	// Step 3: Custom checkboxes
	Resources   []string
	SelectedRes map[string]bool
	ResCursor   int

	// Step 3.5: Secret Backend
	SelectedBackend Backend

	// Confirmation button focus (0: Generate, 1: Cancel)
	ConfirmCursor int

	// Result parameters to return to generator
	Confirmed bool
	Quitted   bool
	ShowHelp  bool
}

func InitialModel(profile string, defaultAppName string) WizardModel {
	m := WizardModel{
		Step:            StepAppInfo,
		Profile:         profile,
		SelectedMode:    ModeSimple,
		SelectedQuality: QualityProduction, // default
		SelectedBackend: BackendVault,      // default
		Resources: []string{
			"Deployment",
			"Service",
			"Ingress",
			"Gateway API",
			"ConfigMap",
			"ExternalSecret",
			"HPA",
			"ServiceMonitor",
			"PDB",
			"VPA",
			"KEDA",
			"StatefulSet",
			"CronJob",
			"ArgoCD Application",
			"Istio VirtualService",
		},
		SelectedRes: map[string]bool{
			"Deployment": true,
			"Service":    true,
		},
		ResCursor:     0,
		ConfirmCursor: 0,
	}

	if profile == "prod" {
		m.SelectedMode = ModeAdvanced
		m.SelectedQuality = QualityProduction
		m.SelectedRes["Ingress"] = true
		m.SelectedRes["HPA"] = true
		m.SelectedRes["PDB"] = true
		m.SelectedRes["ServiceMonitor"] = true
	} else {
		m.SelectedQuality = QualityBasic
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
				switch m.Step {
				case StepMode:
					m.Step = StepAppInfo
				case StepQuality:
					m.Step = StepMode
				case StepCustomResources:
					m.Step = StepQuality
				case StepSecretBackend:
					m.Step = StepCustomResources
				case StepConfirm:
					if m.SelectedMode == ModeCustom {
						if m.SelectedRes["ExternalSecret"] {
							m.Step = StepSecretBackend
						} else {
							m.Step = StepCustomResources
						}
					} else {
						m.Step = StepQuality
					}
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
	case StepQuality:
		return m.updateQuality(msg)
	case StepCustomResources:
		return m.updateCustomResources(msg)
	case StepSecretBackend:
		return m.updateSecretBackend(msg)
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
			m.Step = StepQuality
		}
	}
	return m, nil
}

func (m *WizardModel) updateQuality(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedQuality > QualityBasic {
				m.SelectedQuality--
			} else {
				m.SelectedQuality = QualityEnterprise
			}

		case "down", "j":
			if m.SelectedQuality < QualityEnterprise {
				m.SelectedQuality++
			} else {
				m.SelectedQuality = QualityBasic
			}

		case "enter":
			if m.SelectedMode == ModeCustom {
				m.Step = StepCustomResources
			} else {
				if m.SelectedMode == ModeSimple {
					m.SelectedRes = map[string]bool{
						"Deployment": true,
						"Service":    true,
					}
				} else if m.SelectedMode == ModeAdvanced {
					m.SelectedRes = map[string]bool{
						"Deployment":     true,
						"Service":        true,
						"Ingress":        true,
						"HPA":            true,
						"PDB":            true,
						"ServiceMonitor": true,
					}
				}
				m.Step = StepConfirm
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

			// Smart Dependency Engine
			if res == "ServiceMonitor" && m.SelectedRes["ServiceMonitor"] {
				m.SelectedRes["Service"] = true
			}
			if res == "Service" && !m.SelectedRes["Service"] {
				m.SelectedRes["ServiceMonitor"] = false
			}
			if res == "StatefulSet" && m.SelectedRes["StatefulSet"] {
				m.SelectedRes["PVC"] = true
			}
			if res == "StatefulSet" && !m.SelectedRes["StatefulSet"] {
				m.SelectedRes["PVC"] = false
			}

		case "enter":
			if m.SelectedRes["ExternalSecret"] {
				m.Step = StepSecretBackend
			} else {
				m.Step = StepConfirm
			}
		}
	}
	return m, nil
}

func (m *WizardModel) updateSecretBackend(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedBackend > BackendVault {
				m.SelectedBackend--
			} else {
				m.SelectedBackend = BackendAzure
			}

		case "down", "j":
			if m.SelectedBackend < BackendAzure {
				m.SelectedBackend++
			} else {
				m.SelectedBackend = BackendVault
			}

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

	case StepQuality:
		sb.WriteString(HeaderStyle.Render("Step 2.5: Choose Template Quality"))
		sb.WriteString("\n")
		qualities := []struct {
			name, desc string
		}{
			{"Basic", "Minimal configuration, low resource footprint"},
			{"Production", "Adds requests/limits, probes, HPA, and PDB"},
			{"Enterprise", "Adds NetworkPolicy, topology constraints, PodSecurityContext, anti-affinity"},
		}

		for i, q := range qualities {
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(White)
			if int(m.SelectedQuality) == i {
				cursor = ActiveInputStyle.Render("> ")
				nameStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%-12s : %s\n", cursor, nameStyle.Render(q.name), GrayStyle.Render(q.desc)))
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
		if m.SelectedRes["HPA"] {
			sb.WriteString("\n" + lipgloss.NewStyle().Foreground(Magenta).Italic(true).Render("  💡 HPA requires Resource Requests (Production/Enterprise Quality) to function correctly."))
		}

	case StepSecretBackend:
		sb.WriteString(HeaderStyle.Render("Step 3.5: Choose Secret Backend for ExternalSecret"))
		sb.WriteString("\n")
		backends := []string{
			"Vault",
			"AWS Secrets Manager",
			"GCP Secret Manager",
			"Azure Key Vault",
		}

		for i, b := range backends {
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(White)
			if int(m.SelectedBackend) == i {
				cursor = ActiveInputStyle.Render("> ")
				nameStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, nameStyle.Render(b)))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to select. Esc to go back. (Press '?' for help)"))

	case StepConfirm:
		sb.WriteString(HeaderStyle.Render("Step 4: Confirm Generation Settings"))
		sb.WriteString("\n")

		var resList []string
		for _, r := range m.Resources {
			if m.SelectedRes[r] {
				resList = append(resList, r)
			}
		}

		var qualityNames = []string{"Basic", "Production", "Enterprise"}
		var backendNames = []string{"Vault", "AWS Secrets Manager", "GCP Secret Manager", "Azure Key Vault"}

		backendLine := ""
		if m.SelectedRes["ExternalSecret"] {
			backendLine = fmt.Sprintf("\nSecret Backend    : %s", backendNames[m.SelectedBackend])
		}

		summary := fmt.Sprintf(
			"Application Name  : %s\nNamespace         : %s\nContainer Image   : %s\nContainer Port    : %s\nProfile           : %s\nTemplate Quality  : %s%s\nResources         : %s",
			m.Inputs[0].Value(),
			m.Inputs[1].Value(),
			m.Inputs[2].Value(),
			m.Inputs[3].Value(),
			m.Profile,
			qualityNames[m.SelectedQuality],
			backendLine,
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
	steps := []string{"App Info", "Mode", "Quality", "Resources", "Confirm"}
	var rendered []string

	currentStepIndex := 0
	switch m.Step {
	case StepAppInfo:
		currentStepIndex = 0
	case StepMode:
		currentStepIndex = 1
	case StepQuality:
		currentStepIndex = 2
	case StepCustomResources, StepSecretBackend:
		currentStepIndex = 3
	case StepConfirm:
		currentStepIndex = 4
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

	var qualityNames = []string{"basic", "production", "enterprise"}
	var backendNames = []string{"vault", "aws", "gcp", "azure"}

	cfg := generator.Config{
		AppName:         m.Inputs[0].Value(),
		Namespace:       m.Inputs[1].Value(),
		ImageRepository: repo,
		ImageTag:        tag,
		Port:            port,
		ReplicaCount:    replicaCount,
		IngressEnabled:  m.SelectedRes["Ingress"],
		HPAEnabled:      m.SelectedRes["HPA"],
		HPAMinReplicas:  replicaCount,
		HPAMaxReplicas:  replicaCount * 3,
		ProdProfile:     m.Profile == "prod",

		TemplateQuality: qualityNames[m.SelectedQuality],
		SecretBackend:   backendNames[m.SelectedBackend],

		GenerateDeployment:     m.SelectedRes["Deployment"],
		GenerateService:        m.SelectedRes["Service"],
		GenerateIngress:        m.SelectedRes["Ingress"],
		GenerateGateway:        m.SelectedRes["Gateway API"],
		GenerateConfigMap:      m.SelectedRes["ConfigMap"],
		GenerateExternalSecret: m.SelectedRes["ExternalSecret"],
		GenerateHPA:            m.SelectedRes["HPA"],
		GenerateServiceMonitor: m.SelectedRes["ServiceMonitor"],
		GeneratePDB:            m.SelectedRes["PDB"],
		GenerateVPA:            m.SelectedRes["VPA"],
		GenerateKEDA:           m.SelectedRes["KEDA"],
		GenerateStatefulSet:    m.SelectedRes["StatefulSet"],
		GenerateCronJob:        m.SelectedRes["CronJob"],
		GenerateArgoCD:         m.SelectedRes["ArgoCD Application"],
		GenerateIstio:          m.SelectedRes["Istio VirtualService"],
		GeneratePVC:            m.SelectedRes["PVC"],
		GenerateNetworkPolicy:  m.SelectedRes["NetworkPolicy"] || (qualityNames[m.SelectedQuality] == "enterprise"),
	}

	return cfg, cfg.AppName
}
