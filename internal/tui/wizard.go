package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ihyamarsdev/kgen/internal/generator"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// validatePort returns a valid port number or the default (80).
func validatePort(s string) int {
	port, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || port < 1 || port > 65535 {
		return 80
	}
	return port
}

// storageSizeRe validates Kubernetes storage size patterns (e.g. "10Gi", "500Mi").
var storageSizeRe = regexp.MustCompile(`^\d+(Gi|Mi|Ti)$`)

// validateStorageSize checks if the value matches a valid Kubernetes storage
// size pattern (e.g. "10Gi", "500Mi"). Returns the cleaned value or the default.
func validateStorageSize(s string, defaultVal string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultVal
	}
	if storageSizeRe.MatchString(s) {
		return s
	}
	return defaultVal
}

// validateImage ensures the image string contains a repository and tag.
// If the user enters just "nginx", it appends ":latest".
func validateImage(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "nginx:latest"
	}
	// If no tag separator found, append :latest.
	if !strings.Contains(s, ":") {
		s += ":latest"
	}
	return s
}

type Step int

const (
	StepAppInfo Step = iota
	StepMode
	StepQuality
	StepCustomResources
	StepStatefulSetStorage
	StepStorageClassInfo
	StepServiceAccountInfo
	StepRbacLevel
	StepIngressTls
	StepNetworkPolicyPreset
	StepSecretBackend
	StepServiceType
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

type TuiState int

const (
	StateCategories TuiState = iota
	StateCategoryResources
)

type RbacState int

const (
	StateRbacLevel RbacState = iota
	StateRbacResources
)

// Categories and Items from PRD
var Categories = []string{
	"Workloads",
	"Storage",
	"Identity & RBAC",
	"Networking",
	"Scaling & Reliability",
	"Secrets & Configuration",
	"Monitoring",
	"GitOps",
}

var CategoryItems = map[string][]string{
	"Workloads": {
		"Deployment",
		"StatefulSet",
		"DaemonSet",
		"Job",
		"CronJob",
	},
	"Storage": {
		"PersistentVolumeClaim",
		"StorageClass Reference",
		"CSI Volume",
	},
	"Identity & RBAC": {
		"ServiceAccount",
		"Role",
		"RoleBinding",
		"ClusterRole",
		"ClusterRoleBinding",
	},
	"Networking": {
		"Service",
		"Ingress",
		"Gateway API",
		"NetworkPolicy",
	},
	"Scaling & Reliability": {
		"HPA",
		"VPA",
		"KEDA",
		"PDB",
		"Pod Anti Affinity",
		"Topology Spread Constraints",
		"Priority Class",
	},
	"Secrets & Configuration": {
		"ConfigMap",
		"Secret",
		"ExternalSecret",
		"SealedSecret",
	},
	"Monitoring": {
		"ServiceMonitor",
		"PodMonitor",
		"PrometheusRule",
		"GrafanaDashboard",
	},
	"GitOps": {
		"ArgoCD Application",
		"ArgoCD ApplicationSet",
		"Flux HelmRelease",
		"Flux Kustomization",
	},
}

type WizardModel struct {
	Step      Step
	Profile   string
	OutputDir string

	// Navigation Queue
	ActiveSteps      []Step
	CurrentStepIndex int

	// Step 1: Inputs
	Inputs      []textinput.Model
	ActiveInput int

	// Step 2: Mode selection
	SelectedMode Mode

	// Step 2.5: Template Quality
	SelectedQuality Quality

	// Step 3: Custom checklist (Categories & Items)
	TuiState         TuiState
	SelectedCategory int
	SubResCursor     int
	Resources        []string
	SelectedRes      map[string]bool

	// Step: StatefulSet Storage Option
	SelectedStatefulSetStorage int // 0: Create PVC, 1: Existing PVC

	// Step: StorageClass / PVC Info
	StorageInputs      []textinput.Model
	ActiveStorageInput int
	SelectedAccessMode int // 0: ReadWriteOnce, 1: ReadWriteMany

	// Step: ServiceAccount Dedicated
	ServiceAccountCreate int // 0: Yes, 1: No
	SaNameInput          textinput.Model

	// Step: RBAC Presets
	SelectedRbacLevel int // 0: Read Only, 1: Namespace Admin, 2: Custom
	RbacResources     []string
	SelectedRbacRes   map[string]bool
	RbacCursor        int
	RbacState         RbacState

	// Step: Ingress TLS suggested
	IngressTlsCreate   int // 0: Yes, 1: No
	IngressTlsProvider int // 0: cert-manager, 1: Existing Secret

	// Step: NetworkPolicy Preset
	SelectedNetPolPreset int // 0: Default Deny, 1: Allow Namespace Only, 2: Custom

	// Step: ExternalSecret backend
	SelectedBackend Backend

	// Step: Service Type (NodePort, LoadBalancer, ClusterIP, ExternalName)
	SelectedServiceType int // 0: ClusterIP, 1: NodePort, 2: LoadBalancer, 3: ExternalName

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
		TuiState:        StateCategories,
		SelectedRes:     map[string]bool{},
		ConfirmCursor:   0,
	}

	// Initialize checklist selections to false
	for _, items := range CategoryItems {
		for _, it := range items {
			m.SelectedRes[it] = false
		}
	}

	// Preset for Simple / Advanced
	if profile == "prod" {
		m.SelectedMode = ModeAdvanced
		m.SelectedQuality = QualityProduction
		m.SelectedRes["Deployment"] = true
		m.SelectedRes["Service"] = true
		m.SelectedRes["Ingress"] = true
		m.SelectedRes["HPA"] = true
		m.SelectedRes["PDB"] = true
		m.SelectedRes["ServiceMonitor"] = true
	} else {
		m.SelectedRes["Deployment"] = true
		m.SelectedRes["Service"] = true
		m.SelectedQuality = QualityBasic
	}

	// Initialize inputs
	m.Inputs = make([]textinput.Model, 4)

	m.Inputs[0] = textinput.New()
	m.Inputs[0].Placeholder = "my-app"
	m.Inputs[0].SetValue(defaultAppName)
	m.Inputs[0].Focus()
	m.Inputs[0].PromptStyle = ActiveInputStyle

	m.Inputs[1] = textinput.New()
	m.Inputs[1].Placeholder = "default"
	m.Inputs[1].SetValue("default")
	m.Inputs[1].PromptStyle = InactiveInputStyle

	m.Inputs[2] = textinput.New()
	m.Inputs[2].Placeholder = "nginx:latest"
	m.Inputs[2].SetValue("nginx:latest")
	m.Inputs[2].PromptStyle = InactiveInputStyle

	m.Inputs[3] = textinput.New()
	m.Inputs[3].Placeholder = "80"
	m.Inputs[3].SetValue("80")
	m.Inputs[3].PromptStyle = InactiveInputStyle

	// Initialize Storage inputs
	m.StorageInputs = make([]textinput.Model, 2)
	m.StorageInputs[0] = textinput.New()
	m.StorageInputs[0].Placeholder = "standard"
	m.StorageInputs[0].SetValue("standard")
	m.StorageInputs[0].PromptStyle = InactiveInputStyle

	m.StorageInputs[1] = textinput.New()
	m.StorageInputs[1].Placeholder = "10Gi"
	m.StorageInputs[1].SetValue("10Gi")
	m.StorageInputs[1].PromptStyle = InactiveInputStyle

	// Initialize ServiceAccount input
	m.SaNameInput = textinput.New()
	m.SaNameInput.Placeholder = defaultAppName
	m.SaNameInput.SetValue(defaultAppName)
	m.SaNameInput.PromptStyle = InactiveInputStyle

	// Initialize Rbac checklist
	m.RbacResources = []string{"ConfigMaps", "Secrets", "Pods", "Deployments"}
	m.SelectedRbacRes = map[string]bool{"ConfigMaps": true, "Secrets": true, "Pods": true}

	// Build default initial steps
	m.ActiveSteps = []Step{StepAppInfo, StepMode}
	m.CurrentStepIndex = 0

	return m
}

func (m *WizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *WizardModel) advanceStep() {
	if m.CurrentStepIndex < len(m.ActiveSteps)-1 {
		m.CurrentStepIndex++
		m.Step = m.ActiveSteps[m.CurrentStepIndex]
	}
}

func (m *WizardModel) regressStep() {
	if m.CurrentStepIndex > 0 {
		m.CurrentStepIndex--
		m.Step = m.ActiveSteps[m.CurrentStepIndex]
	}
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
			if m.Step == StepCustomResources && m.TuiState == StateCategoryResources {
				// Escape category items back to categories list
				m.TuiState = StateCategories
				return m, nil
			}
			if m.Step == StepRbacLevel && m.RbacState == StateRbacResources {
				m.RbacState = StateRbacLevel
				return m, nil
			}

			if m.Step > StepAppInfo {
				m.regressStep()
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
	case StepStatefulSetStorage:
		return m.updateStatefulSetStorage(msg)
	case StepStorageClassInfo:
		return m.updateStorageClassInfo(msg)
	case StepServiceAccountInfo:
		return m.updateServiceAccountInfo(msg)
	case StepRbacLevel:
		return m.updateRbacLevel(msg)
	case StepIngressTls:
		return m.updateIngressTls(msg)
	case StepNetworkPolicyPreset:
		return m.updateNetworkPolicyPreset(msg)
	case StepSecretBackend:
		return m.updateSecretBackend(msg)
	case StepServiceType:
		return m.updateServiceType(msg)
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
				// Validate inputs before proceeding.
				appName := strings.TrimSpace(m.Inputs[0].Value())
				if appName == "" {
					m.Inputs[0].SetValue("my-app")
				}
				// Validate image (add :latest if missing tag).
				image := validateImage(m.Inputs[2].Value())
				m.Inputs[2].SetValue(image)
				// Validate port (1-65535, default 80).
				port := validatePort(m.Inputs[3].Value())
				m.Inputs[3].SetValue(strconv.Itoa(port))
				m.ActiveSteps = []Step{StepAppInfo, StepMode}
				m.CurrentStepIndex = 1
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
			if m.SelectedMode == ModeCustom {
				m.ActiveSteps = []Step{StepAppInfo, StepMode, StepCustomResources}
				m.CurrentStepIndex = 2
				m.Step = StepCustomResources
			} else {
				m.ActiveSteps = []Step{StepAppInfo, StepMode, StepQuality, StepConfirm}
				m.CurrentStepIndex = 2
				m.Step = StepQuality
			}
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
			// Set defaults based on simple/advanced mode
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
					"NetworkPolicy":  true,
				}
			}
			m.advanceStep()
		}
	}
	return m, nil
}

func (m *WizardModel) updateCustomResources(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.TuiState == StateCategories {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.SelectedCategory > 0 {
					m.SelectedCategory--
				} else {
					m.SelectedCategory = len(Categories) // Index of [ Continue ]
				}

			case "down", "j":
				if m.SelectedCategory < len(Categories) {
					m.SelectedCategory++
				} else {
					m.SelectedCategory = 0
				}

			case "enter":
				if m.SelectedCategory == len(Categories) {
					// User clicked [ Continue to Confirm ].
					// Build the dynamic steps based on what was selected!
					m.ActiveSteps = []Step{StepAppInfo, StepMode, StepCustomResources}
					if m.SelectedRes["StatefulSet"] {
						m.ActiveSteps = append(m.ActiveSteps, StepStatefulSetStorage)
					}
					if m.SelectedRes["PersistentVolumeClaim"] {
						m.ActiveSteps = append(m.ActiveSteps, StepStorageClassInfo)
					}
					if m.SelectedRes["ServiceAccount"] {
						m.ActiveSteps = append(m.ActiveSteps, StepServiceAccountInfo)
					}
					if m.SelectedRes["Role"] || m.SelectedRes["RoleBinding"] || m.SelectedRes["ClusterRole"] || m.SelectedRes["ClusterRoleBinding"] {
						m.ActiveSteps = append(m.ActiveSteps, StepRbacLevel)
					}
					if m.SelectedRes["Ingress"] {
						m.ActiveSteps = append(m.ActiveSteps, StepIngressTls)
					}
					if m.SelectedRes["Service"] {
						m.ActiveSteps = append(m.ActiveSteps, StepServiceType)
					}
					if m.SelectedRes["NetworkPolicy"] {
						m.ActiveSteps = append(m.ActiveSteps, StepNetworkPolicyPreset)
					}
					if m.SelectedRes["ExternalSecret"] {
						m.ActiveSteps = append(m.ActiveSteps, StepSecretBackend)
					}
					m.ActiveSteps = append(m.ActiveSteps, StepConfirm)
					m.CurrentStepIndex = 2
					m.advanceStep()
				} else {
					m.TuiState = StateCategoryResources
					m.SubResCursor = 0
				}
			}
		}
	} else {
		// Category item Checklist TUI
		currentCat := Categories[m.SelectedCategory]
		items := CategoryItems[currentCat]

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.SubResCursor > 0 {
					m.SubResCursor--
				} else {
					m.SubResCursor = len(items) - 1
				}

			case "down", "j":
				if m.SubResCursor < len(items)-1 {
					m.SubResCursor++
				} else {
					m.SubResCursor = 0
				}

			case " ":
				it := items[m.SubResCursor]
				m.SelectedRes[it] = !m.SelectedRes[it]

				// Smart Dependency Engine
				if it == "ServiceMonitor" && m.SelectedRes["ServiceMonitor"] {
					m.SelectedRes["Service"] = true
				}
				if it == "Service" && !m.SelectedRes["Service"] {
					m.SelectedRes["ServiceMonitor"] = false
				}
				if it == "StatefulSet" && m.SelectedRes["StatefulSet"] {
					m.SelectedRes["PersistentVolumeClaim"] = true
				}
				if it == "StatefulSet" && !m.SelectedRes["StatefulSet"] {
					m.SelectedRes["PersistentVolumeClaim"] = false
				}
				if it == "RoleBinding" && m.SelectedRes["RoleBinding"] {
					m.SelectedRes["Role"] = true
					m.SelectedRes["ServiceAccount"] = true
				}
				if it == "ClusterRoleBinding" && m.SelectedRes["ClusterRoleBinding"] {
					m.SelectedRes["ClusterRole"] = true
					m.SelectedRes["ServiceAccount"] = true
				}

			case "enter":
				m.TuiState = StateCategories
			}
		}
	}
	return m, nil
}

func (m *WizardModel) updateStatefulSetStorage(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			m.SelectedStatefulSetStorage = 1 - m.SelectedStatefulSetStorage

		case "enter":
			if m.SelectedStatefulSetStorage == 0 {
				m.SelectedRes["PersistentVolumeClaim"] = true
			} else {
				m.SelectedRes["PersistentVolumeClaim"] = false
			}
			m.advanceStep()
		}
	}
	return m, nil
}

func (m *WizardModel) updateStorageClassInfo(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.ActiveStorageInput == 1 {
				// Validate storage class name before proceeding.
				storageClass := strings.TrimSpace(m.StorageInputs[0].Value())
				if storageClass == "" {
					m.StorageInputs[0].SetValue("standard")
				}
				// Validate storage size before proceeding.
				size := validateStorageSize(m.StorageInputs[1].Value(), "10Gi")
				m.StorageInputs[1].SetValue(size)
				m.advanceStep()
				return m, nil
			}
			m.nextStorageInput()

		case "up", "shift+tab":
			m.prevStorageInput()

		case "down", "tab":
			m.nextStorageInput()

		case "left", "right", "h", "l":
			if m.ActiveStorageInput == 2 {
				m.SelectedAccessMode = 1 - m.SelectedAccessMode
			}
		}
	}

	var cmd tea.Cmd
	if m.ActiveStorageInput < 2 {
		m.StorageInputs[m.ActiveStorageInput], cmd = m.StorageInputs[m.ActiveStorageInput].Update(msg)
	}
	return m, cmd
}

func (m *WizardModel) nextStorageInput() {
	if m.ActiveStorageInput < 2 {
		m.StorageInputs[m.ActiveStorageInput].Blur()
		m.StorageInputs[m.ActiveStorageInput].PromptStyle = InactiveInputStyle
	}
	m.ActiveStorageInput = (m.ActiveStorageInput + 1) % 2
	if m.ActiveStorageInput < 2 {
		m.StorageInputs[m.ActiveStorageInput].Focus()
		m.StorageInputs[m.ActiveStorageInput].PromptStyle = ActiveInputStyle
	}
}

func (m *WizardModel) prevStorageInput() {
	if m.ActiveStorageInput < 2 {
		m.StorageInputs[m.ActiveStorageInput].Blur()
		m.StorageInputs[m.ActiveStorageInput].PromptStyle = InactiveInputStyle
	}
	m.ActiveStorageInput = (m.ActiveStorageInput - 1 + 3) % 2
	if m.ActiveStorageInput < 2 {
		m.StorageInputs[m.ActiveStorageInput].Focus()
		m.StorageInputs[m.ActiveStorageInput].PromptStyle = ActiveInputStyle
	}
}

func (m *WizardModel) updateServiceAccountInfo(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			m.ServiceAccountCreate = 1 - m.ServiceAccountCreate

		case "enter":
			if m.ServiceAccountCreate == 1 { // No
				m.advanceStep()
				return m, nil
			}
			// Focus input if Yes
			if !m.SaNameInput.Focused() {
				m.SaNameInput.Focus()
				m.SaNameInput.PromptStyle = ActiveInputStyle
				return m, nil
			}
			m.advanceStep()
		}
	}

	var cmd tea.Cmd
	if m.ServiceAccountCreate == 0 && m.SaNameInput.Focused() {
		m.SaNameInput, cmd = m.SaNameInput.Update(msg)
	}
	return m, cmd
}

func (m *WizardModel) updateRbacLevel(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.RbacState == StateRbacLevel {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.SelectedRbacLevel > 0 {
					m.SelectedRbacLevel--
				} else {
					m.SelectedRbacLevel = 2
				}

			case "down", "j":
				if m.SelectedRbacLevel < 2 {
					m.SelectedRbacLevel++
				} else {
					m.SelectedRbacLevel = 0
				}

			case "enter":
				if m.SelectedRbacLevel == 2 {
					m.RbacState = StateRbacResources
					m.RbacCursor = 0
				} else {
					m.advanceStep()
				}
			}
		}
	} else {
		// Custom RBAC resource select TUI
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.RbacCursor > 0 {
					m.RbacCursor--
				} else {
					m.RbacCursor = len(m.RbacResources) - 1
				}

			case "down", "j":
				if m.RbacCursor < len(m.RbacResources)-1 {
					m.RbacCursor++
				} else {
					m.RbacCursor = 0
				}

			case " ":
				res := m.RbacResources[m.RbacCursor]
				m.SelectedRbacRes[res] = !m.SelectedRbacRes[res]

			case "enter":
				m.advanceStep()
			}
		}
	}
	return m, nil
}

func (m *WizardModel) updateIngressTls(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "down", "j":
			m.IngressTlsCreate = 1 - m.IngressTlsCreate

		case "left", "right", "h", "l":
			if m.IngressTlsCreate == 0 {
				m.IngressTlsProvider = 1 - m.IngressTlsProvider
			}

		case "enter":
			if m.IngressTlsCreate == 1 { // No
				m.advanceStep()
				return m, nil
			}
			// If Yes, choose provider step
			m.advanceStep()
		}
	}
	return m, nil
}

func (m *WizardModel) updateNetworkPolicyPreset(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedNetPolPreset > 0 {
				m.SelectedNetPolPreset--
			} else {
				m.SelectedNetPolPreset = 2
			}

		case "down", "j":
			if m.SelectedNetPolPreset < 2 {
				m.SelectedNetPolPreset++
			} else {
				m.SelectedNetPolPreset = 0
			}

		case "enter":
			m.advanceStep()
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
			m.advanceStep()
		}
	}
	return m, nil
}

func (m *WizardModel) updateServiceType(msg tea.Msg) (tea.Model, tea.Cmd) {
	const serviceTypes = 4 // ClusterIP, NodePort, LoadBalancer, ExternalName
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.SelectedServiceType > 0 {
				m.SelectedServiceType--
			} else {
				m.SelectedServiceType = serviceTypes - 1
			}

		case "down", "j":
			if m.SelectedServiceType < serviceTypes-1 {
				m.SelectedServiceType++
			} else {
				m.SelectedServiceType = 0
			}

		case "enter":
			m.advanceStep()
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
			{"Tab / Shift+Tab", "Navigate text inputs"},
			{"Up / Down (k / j)", "Navigate through list options"},
			{"Space", "Toggle checklist selection"},
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

	sb.WriteString(TitleStyle.Render(" KGen — Helm Chart Generator "))
	sb.WriteString("\n")

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
			{"Simple", "Generates Deployment and Service (dev/testing)"},
			{"Advanced", "Generates Deployment, Service, Ingress, HPA, PDB, ServiceMonitor, NetPol"},
			{"Custom", "Pick resources categorised individually"},
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
		if m.TuiState == StateCategories {
			sb.WriteString(HeaderStyle.Render("Step 3: Select Resource Categories"))
			sb.WriteString("\n")
			for i, cat := range Categories {
				cursor := "  "
				catStyle := lipgloss.NewStyle().Foreground(White)
				if m.SelectedCategory == i {
					cursor = ActiveInputStyle.Render("> ")
					catStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
				}
				selectedCount := 0
				items := CategoryItems[cat]
				for _, it := range items {
					if m.SelectedRes[it] {
						selectedCount++
					}
				}
				countStr := fmt.Sprintf("(%d/%d selected)", selectedCount, len(items))
				sb.WriteString(fmt.Sprintf("%s%-24s %s\n", cursor, catStyle.Render(cat), GrayStyle.Render(countStr)))
			}

			cursor := "  "
			btnStyle := NormalButton
			if m.SelectedCategory == len(Categories) {
				cursor = ActiveInputStyle.Render("> ")
				btnStyle = FocusedButton
			}
			sb.WriteString("\n" + cursor + btnStyle.Render("Continue to Confirm") + "\n")
			sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to inspect/edit category. Esc to go back. (Press '?' for help)"))
		} else {
			currentCat := Categories[m.SelectedCategory]
			sb.WriteString(HeaderStyle.Render(fmt.Sprintf("Step 3: Select %s Resources", currentCat)))
			sb.WriteString("\n")
			items := CategoryItems[currentCat]
			for i, it := range items {
				cursor := "  "
				itStyle := lipgloss.NewStyle().Foreground(White)
				if m.SubResCursor == i {
					cursor = ActiveInputStyle.Render("> ")
					itStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
				}

				checked := " "
				if m.SelectedRes[it] {
					checked = CheckboxChecked.Render("✓")
				} else {
					checked = CheckboxUnchecked.Render(" ")
				}

				sb.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, checked, itStyle.Render(it)))
			}
			sb.WriteString(HelpStyle.Render("Use up/down to navigate. Space to toggle. Enter/Esc to return to Categories."))
		}

	case StepStatefulSetStorage:
		sb.WriteString(HeaderStyle.Render("Step 3.1: StatefulSet Storage Settings"))
		sb.WriteString("\n")
		sb.WriteString("Persistent Storage Required?\n\n")

		opts := []string{"Create PVC", "Existing PVC"}
		for i, opt := range opts {
			cursor := "  "
			optStyle := lipgloss.NewStyle().Foreground(White)
			if m.SelectedStatefulSetStorage == i {
				cursor = ActiveInputStyle.Render("> ")
				optStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, optStyle.Render(opt)))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to select. Esc to go back."))

	case StepStorageClassInfo:
		sb.WriteString(HeaderStyle.Render("Step 3.2: Configure PVC / Storage settings"))
		sb.WriteString("\n")

		labels := []string{"Storage Class Name :", "PVC Size           :"}
		for i, input := range m.StorageInputs {
			prefix := "  "
			if i == m.ActiveStorageInput {
				prefix = ActiveInputStyle.Render("> ")
			}
			sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, labels[i], input.View()))
		}

		// Access Mode radio
		prefix := "  "
		if m.ActiveStorageInput == 2 {
			prefix = ActiveInputStyle.Render("> ")
		}
		modeStr := "ReadWriteOnce"
		if m.SelectedAccessMode == 1 {
			modeStr = "ReadWriteMany"
		}
		sb.WriteString(fmt.Sprintf("%sAccess Mode        : %s (Press left/right to toggle)\n", prefix, ActiveInputStyle.Render(modeStr)))

		sb.WriteString(HelpStyle.Render("Use up/down or tab to navigate. Left/right to toggle mode. Enter to continue."))

	case StepServiceAccountInfo:
		sb.WriteString(HeaderStyle.Render("Step 3.3: Configure Dedicated ServiceAccount"))
		sb.WriteString("\n")
		sb.WriteString("Create Dedicated ServiceAccount?\n\n")

		opts := []string{"Yes", "No"}
		for i, opt := range opts {
			cursor := "  "
			optStyle := lipgloss.NewStyle().Foreground(White)
			if m.ServiceAccountCreate == i {
				cursor = ActiveInputStyle.Render("> ")
				optStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, optStyle.Render(opt)))
		}

		if m.ServiceAccountCreate == 0 {
			sb.WriteString("\n  ServiceAccount Name: " + m.SaNameInput.View() + "\n")
		}
		sb.WriteString(HelpStyle.Render("Use up/down to select option. Enter to confirm. Esc to go back."))

	case StepRbacLevel:
		if m.RbacState == StateRbacLevel {
			sb.WriteString(HeaderStyle.Render("Step 3.4: Configure RBAC Authorization"))
			sb.WriteString("\n")
			sb.WriteString("Select RBAC level:\n\n")

			levels := []struct {
				name, desc string
			}{
				{"Read Only", "Access to pods, configmaps, and secrets (read-only)"},
				{"Namespace Admin", "Full admin permissions inside the namespace"},
				{"Custom", "Pick custom resource list"},
			}

			for i, lvl := range levels {
				cursor := "  "
				lvlStyle := lipgloss.NewStyle().Foreground(White)
				if m.SelectedRbacLevel == i {
					cursor = ActiveInputStyle.Render("> ")
					lvlStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
				}
				sb.WriteString(fmt.Sprintf("%s%-18s : %s\n", cursor, lvlStyle.Render(lvl.name), GrayStyle.Render(lvl.desc)))
			}
			sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to select. Esc to go back."))
		} else {
			sb.WriteString(HeaderStyle.Render("Step 3.4: Select allowed RBAC Resources"))
			sb.WriteString("\n")

			for i, r := range m.RbacResources {
				cursor := "  "
				rStyle := lipgloss.NewStyle().Foreground(White)
				if m.RbacCursor == i {
					cursor = ActiveInputStyle.Render("> ")
					rStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
				}

				checked := " "
				if m.SelectedRbacRes[r] {
					checked = CheckboxChecked.Render("✓")
				} else {
					checked = CheckboxUnchecked.Render(" ")
				}

				sb.WriteString(fmt.Sprintf("%s[%s] %s\n", cursor, checked, rStyle.Render(r)))
			}
			sb.WriteString(HelpStyle.Render("Use up/down to navigate. Space to toggle. Enter to confirm. Esc to go back."))
		}

	case StepIngressTls:
		sb.WriteString(HeaderStyle.Render("Step 3.5: Ingress TLS settings"))
		sb.WriteString("\n")
		sb.WriteString("Suggested: TLS Certificate. Use TLS Certificate?\n\n")

		opts := []string{"Yes", "No"}
		for i, opt := range opts {
			cursor := "  "
			optStyle := lipgloss.NewStyle().Foreground(White)
			if m.IngressTlsCreate == i {
				cursor = ActiveInputStyle.Render("> ")
				optStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, optStyle.Render(opt)))
		}

		if m.IngressTlsCreate == 0 {
			sb.WriteString("\n  TLS Provider:\n")
			providers := []string{"cert-manager", "Existing Secret"}
			for i, p := range providers {
				cursor := "    "
				pStyle := lipgloss.NewStyle().Foreground(White)
				if m.IngressTlsProvider == i {
					cursor = ActiveInputStyle.Render("  > ")
					pStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
				}
				sb.WriteString(fmt.Sprintf("%s%s\n", cursor, pStyle.Render(p)))
			}
			sb.WriteString("\n" + HelpStyle.Render("Use left/right or h/l to toggle provider when Yes is active."))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to select. Enter to confirm. Esc to go back."))

	case StepNetworkPolicyPreset:
		sb.WriteString(HeaderStyle.Render("Step 3.6: Configure NetworkPolicy"))
		sb.WriteString("\n")
		sb.WriteString("Select NetworkPolicy preset rules:\n\n")

		presets := []struct {
			name, desc string
		}{
			{"Default Deny", "Deny all incoming and outgoing traffic by default"},
			{"Allow Namespace Only", "Allow traffic only from pods inside this namespace"},
			{"Custom", "Standard HTTP port access and egress allowed"},
		}

		for i, p := range presets {
			cursor := "  "
			pStyle := lipgloss.NewStyle().Foreground(White)
			if m.SelectedNetPolPreset == i {
				cursor = ActiveInputStyle.Render("> ")
				pStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%-24s : %s\n", cursor, pStyle.Render(p.name), GrayStyle.Render(p.desc)))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to select. Esc to go back."))

	case StepSecretBackend:
		sb.WriteString(HeaderStyle.Render("Step 3.7: Choose Secret Backend for ExternalSecret"))
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

	case StepServiceType:
		sb.WriteString(HeaderStyle.Render("Step 3.8: Choose Service Type"))
		sb.WriteString("\n")
		serviceTypes := []struct {
			name string
			desc string
		}{
			{"ClusterIP", "Default — internal cluster access only"},
			{"NodePort", "Expose on each node's IP: static port"},
			{"LoadBalancer", "Provision external load balancer (cloud)"},
			{"ExternalName", "DNS alias to an external service"},
		}

		for i, st := range serviceTypes {
			cursor := "  "
			stStyle := lipgloss.NewStyle().Foreground(White)
			if m.SelectedServiceType == i {
				cursor = ActiveInputStyle.Render("> ")
				stStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
			}
			sb.WriteString(fmt.Sprintf("%s%-20s : %s\n", cursor, stStyle.Render(st.name), GrayStyle.Render(st.desc)))
		}
		sb.WriteString(HelpStyle.Render("Use up/down to navigate. Enter to select. Esc to go back. (Press '?' for help)"))

	case StepConfirm:
		sb.WriteString(HeaderStyle.Render("Step 4: Confirm Generation Settings"))
		sb.WriteString("\n")

		var resList []string
		for _, cat := range Categories {
			for _, r := range CategoryItems[cat] {
				if m.SelectedRes[r] {
					resList = append(resList, r)
				}
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
	steps := []string{"App Info", "Mode", "Resources", "Confirm"}
	var rendered []string

	currentStepIndex := 0
	switch m.Step {
	case StepAppInfo:
		currentStepIndex = 0
	case StepMode:
		currentStepIndex = 1
	case StepQuality, StepCustomResources, StepStatefulSetStorage, StepStorageClassInfo, StepServiceAccountInfo, StepRbacLevel, StepIngressTls, StepNetworkPolicyPreset, StepSecretBackend:
		currentStepIndex = 2
	case StepConfirm:
		currentStepIndex = 3
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
	var serviceTypeNames = []string{"ClusterIP", "NodePort", "LoadBalancer", "ExternalName"}
	var rbacLevelNames = []string{"readonly", "admin", "custom"}
	var netpolPresetNames = []string{"defaultdeny", "namespaceonly", "custom"}

	// Extract Rbac Custom resources
	var rbacCustomRes []string
	if m.SelectedRbacLevel == 2 {
		for r, checked := range m.SelectedRbacRes {
			if checked {
				rbacCustomRes = append(rbacCustomRes, strings.ToLower(r))
			}
		}
	}

	saName := ""
	if m.SelectedRes["ServiceAccount"] {
		if m.ServiceAccountCreate == 0 {
			saName = m.SaNameInput.Value()
		} else {
			saName = "default"
		}
	}

	appName := strings.TrimSpace(m.Inputs[0].Value())
	if appName == "" {
		appName = "my-app"
	}

	cfg := generator.Config{
		AppName:         appName,
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
		ServiceType:     serviceTypeNames[m.SelectedServiceType],

		// Storage Config
		StorageClass:      m.StorageInputs[0].Value(),
		StorageSize:       m.StorageInputs[1].Value(),
		StorageAccessMode: "ReadWriteOnce",

		// ServiceAccount Name
		ServiceAccountName: saName,

		// RBAC Level
		RbacLevel:           rbacLevelNames[m.SelectedRbacLevel],
		RbacCustomResources: rbacCustomRes,

		// Ingress TLS
		IngressTlsEnabled:  m.SelectedRes["Ingress"] && m.IngressTlsCreate == 0,
		IngressTlsProvider: "cert-manager",

		// NetworkPolicy Preset
		NetworkPolicyPreset: netpolPresetNames[m.SelectedNetPolPreset],

		// File generation toggles
		GenerateDeployment:                m.SelectedRes["Deployment"],
		GenerateService:                   m.SelectedRes["Service"],
		GenerateIngress:                   m.SelectedRes["Ingress"],
		GenerateGateway:                   m.SelectedRes["Gateway API"],
		GenerateConfigMap:                 m.SelectedRes["ConfigMap"],
		GenerateSecret:                    m.SelectedRes["Secret"],
		GenerateExternalSecret:            m.SelectedRes["ExternalSecret"],
		GenerateSealedSecret:              m.SelectedRes["SealedSecret"],
		GenerateHPA:                       m.SelectedRes["HPA"],
		GenerateServiceMonitor:            m.SelectedRes["ServiceMonitor"],
		GeneratePDB:                       m.SelectedRes["PDB"],
		GenerateVPA:                       m.SelectedRes["VPA"],
		GenerateKEDA:                      m.SelectedRes["KEDA"],
		GenerateStatefulSet:               m.SelectedRes["StatefulSet"],
		GenerateCronJob:                   m.SelectedRes["CronJob"],
		GenerateArgoCD:                    m.SelectedRes["ArgoCD Application"],
		GenerateIstio:                     m.SelectedRes["Istio VirtualService"],
		GeneratePVC:                       m.SelectedRes["PersistentVolumeClaim"],
		GenerateNetworkPolicy:             m.SelectedRes["NetworkPolicy"],
		GenerateDaemonSet:                 m.SelectedRes["DaemonSet"],
		GenerateJob:                       m.SelectedRes["Job"],
		GenerateServiceAccount:            m.SelectedRes["ServiceAccount"],
		GenerateRbac:                      m.SelectedRes["Role"] || m.SelectedRes["RoleBinding"] || m.SelectedRes["ClusterRole"] || m.SelectedRes["ClusterRoleBinding"],
		GenerateRole:                      m.SelectedRes["Role"],
		GenerateRoleBinding:               m.SelectedRes["RoleBinding"],
		GenerateClusterRole:               m.SelectedRes["ClusterRole"],
		GenerateClusterRoleBinding:        m.SelectedRes["ClusterRoleBinding"],
		GeneratePriorityClass:             m.SelectedRes["Priority Class"],
		GeneratePodMonitor:                m.SelectedRes["PodMonitor"],
		GeneratePrometheusRule:            m.SelectedRes["PrometheusRule"],
		GenerateGrafanaDashboard:          m.SelectedRes["GrafanaDashboard"],
		GenerateArgoCDSet:                 m.SelectedRes["ArgoCD ApplicationSet"],
		GenerateFlux:                      m.SelectedRes["Flux HelmRelease"] || m.SelectedRes["Flux Kustomization"],
		GeneratePodAntiAffinity:           m.SelectedRes["Pod Anti Affinity"] || qualityNames[m.SelectedQuality] == "enterprise",
		GenerateTopologySpreadConstraints: m.SelectedRes["Topology Spread Constraints"] || qualityNames[m.SelectedQuality] == "enterprise",
	}

	if m.SelectedAccessMode == 1 {
		cfg.StorageAccessMode = "ReadWriteMany"
	}
	if m.IngressTlsProvider == 1 {
		cfg.IngressTlsProvider = "secret"
	}

	return cfg, cfg.AppName
}
