package validator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ValidationResult struct {
	Check   string
	Status  string // PASS, WARN
	Message string
}

type Checks struct {
	HasLivenessProbe             bool
	HasReadinessProbe            bool
	HasLimits                    bool
	HasRequests                  bool
	HasSecurityCtx               bool
	HasHPA                       bool
	HasPDB                       bool
	HasNetworkPolicy             bool
	HasTopologySpreadConstraints bool
	HasPodAntiAffinity           bool
}

func ValidateDir(dirPath string) ([]ValidationResult, error) {
	fi, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("directory not found: %w", err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	checks := Checks{}

	// Scan templates directory for YAML files indicating resource presence
	templatesDir := filepath.Join(dirPath, "templates")
	if _, err := os.Stat(templatesDir); err == nil {
		_ = filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			name := strings.ToLower(filepath.Base(path))
			switch name {
			case "hpa.yaml", "hpa.yml":
				checks.HasHPA = true
			case "pdb.yaml", "pdb.yml":
				checks.HasPDB = true
			case "networkpolicy.yaml", "networkpolicy.yml":
				checks.HasNetworkPolicy = true
			}
			return nil
		})
	}

	valuesPath := filepath.Join(dirPath, "values.yaml")
	if _, err := os.Stat(valuesPath); err == nil {
		content, err := os.ReadFile(valuesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read values.yaml: %w", err)
		}

		var valMap map[string]any
		if err := yaml.Unmarshal(content, &valMap); err == nil {
			if hasKeyPath(valMap, "livenessProbe") && valMap["livenessProbe"] != nil {
				if m, ok := valMap["livenessProbe"].(map[string]any); !ok || len(m) > 0 {
					checks.HasLivenessProbe = true
				}
			}
			if hasKeyPath(valMap, "readinessProbe") && valMap["readinessProbe"] != nil {
				if m, ok := valMap["readinessProbe"].(map[string]any); !ok || len(m) > 0 {
					checks.HasReadinessProbe = true
				}
			}
			if hasKeyPath(valMap, "securityContext") && valMap["securityContext"] != nil {
				if m, ok := valMap["securityContext"].(map[string]any); !ok || len(m) > 0 {
					checks.HasSecurityCtx = true
				}
			}
			if hasKeyPath(valMap, "resources", "limits") {
				checks.HasLimits = true
			}
			if hasKeyPath(valMap, "resources", "requests") {
				checks.HasRequests = true
			}
			if hasKeyPath(valMap, "autoscaling", "enabled") {
				if autoMap, ok := valMap["autoscaling"].(map[string]any); ok {
					if enabled, ok := autoMap["enabled"]; ok {
						if boolVal, ok := enabled.(bool); ok && boolVal {
							checks.HasHPA = true
						}
					}
				}
			}
			if hasKeyPath(valMap, "pdb", "enabled") {
				if pdbMap, ok := valMap["pdb"].(map[string]any); ok {
					if enabled, ok := pdbMap["enabled"]; ok {
						if boolVal, ok := enabled.(bool); ok && boolVal {
							checks.HasPDB = true
						}
					}
				}
			}
			if hasKeyPath(valMap, "networkPolicy", "enabled") {
				if npMap, ok := valMap["networkPolicy"].(map[string]any); ok {
					if enabled, ok := npMap["enabled"]; ok {
						if boolVal, ok := enabled.(bool); ok && boolVal {
							checks.HasNetworkPolicy = true
						}
					}
				}
			}
			if hasKeyPath(valMap, "topologySpreadConstraints") {
				if constraints, ok := valMap["topologySpreadConstraints"]; ok && constraints != nil {
					switch v := constraints.(type) {
					case []any:
						if len(v) > 0 {
							checks.HasTopologySpreadConstraints = true
						}
					case []map[string]any:
						if len(v) > 0 {
							checks.HasTopologySpreadConstraints = true
						}
					}
				}
			}
			if hasKeyPath(valMap, "affinity", "podAntiAffinity") {
				if affinity, ok := valMap["affinity"].(map[string]any); ok {
					if paa, ok := affinity["podAntiAffinity"]; ok && paa != nil {
						checks.HasPodAntiAffinity = true
					}
				}
			}
		}
	} else {
		err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			strContent := string(content)

			if strings.Contains(strContent, "livenessProbe") {
				checks.HasLivenessProbe = true
			}
			if strings.Contains(strContent, "readinessProbe") {
				checks.HasReadinessProbe = true
			}
			if strings.Contains(strContent, "securityContext") {
				checks.HasSecurityCtx = true
			}
			if strings.Contains(strContent, "limits:") {
				checks.HasLimits = true
			}
			if strings.Contains(strContent, "requests:") {
				checks.HasRequests = true
			}
			if strings.Contains(strContent, "autoscaling:") {
				checks.HasHPA = true
			}
			if strings.Contains(strContent, "pdb:") || strings.Contains(strContent, "kind: PodDisruptionBudget") {
				checks.HasPDB = true
			}
			if strings.Contains(strContent, "networkPolicy:") || strings.Contains(strContent, "kind: NetworkPolicy") {
				checks.HasNetworkPolicy = true
			}
			if strings.Contains(strContent, "topologySpreadConstraints:") {
				checks.HasTopologySpreadConstraints = true
			}
			if strings.Contains(strContent, "podAntiAffinity:") || strings.Contains(strContent, "PodAntiAffinity") {
				checks.HasPodAntiAffinity = true
			}

			// Also check for template file names
			if strings.Contains(strContent, "kind: HorizontalPodAutoscaler") {
				checks.HasHPA = true
			}
			if strings.Contains(strContent, "kind: PodDisruptionBudget") {
				checks.HasPDB = true
			}
			if strings.Contains(strContent, "kind: NetworkPolicy") {
				checks.HasNetworkPolicy = true
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Table-driven result building eliminates repetitive if/else blocks.
	type checkDef struct {
		name    string
		passed  bool
		passMsg string
		warnMsg string
	}

	defs := []checkDef{
		{"Resource Limits", checks.HasLimits, "Resource limit configured", "No resource limit configured"},
		{"Resource Requests", checks.HasRequests, "Resource request configured", "No resource request configured"},
		{"Liveness Probe", checks.HasLivenessProbe, "Liveness probe found", "No liveness probe found"},
		{"Readiness Probe", checks.HasReadinessProbe, "Readiness probe found", "No readiness probe found"},
		{"Security Context", checks.HasSecurityCtx, "Security context configured", "No security context configured"},
		{"HPA (Horizontal Pod Autoscaler)", checks.HasHPA, "HPA configured", "No HPA configured"},
		{"PDB (Pod Disruption Budget)", checks.HasPDB, "PDB configured", "No PDB configured"},
		{"NetworkPolicy", checks.HasNetworkPolicy, "NetworkPolicy configured", "No NetworkPolicy configured"},
		{"Topology Spread Constraints", checks.HasTopologySpreadConstraints, "Topology spread constraints configured", "No topology spread constraints configured"},
		{"Pod Anti Affinity", checks.HasPodAntiAffinity, "Pod anti affinity configured", "No pod anti affinity configured"},
	}

	results := make([]ValidationResult, 0, len(defs))
	for _, d := range defs {
		status, msg := "PASS", d.passMsg
		if !d.passed {
			status, msg = "WARN", d.warnMsg
		}
		results = append(results, ValidationResult{Check: d.name, Status: status, Message: msg})
	}

	return results, nil
}

func hasKeyPath(m map[string]any, path ...string) bool {
	if len(path) == 0 {
		return true
	}
	val, ok := m[path[0]]
	if !ok || val == nil {
		return false
	}
	if len(path) == 1 {
		return true
	}
	nextMap, ok := val.(map[string]any)
	if !ok {
		nextMapAny, ok := val.(map[any]any)
		if !ok {
			return false
		}
		converted := make(map[string]any)
		for k, v := range nextMapAny {
			converted[fmt.Sprintf("%v", k)] = v
		}
		return hasKeyPath(converted, path[1:]...)
	}
	return hasKeyPath(nextMap, path[1:]...)
}
