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
	HasLivenessProbe  bool
	HasReadinessProbe bool
	HasLimits         bool
	HasRequests       bool
	HasSecurityCtx    bool
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

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	results := []ValidationResult{}

	// Resource Limits
	if checks.HasLimits {
		results = append(results, ValidationResult{
			Check:   "Resource Limits",
			Status:  "PASS",
			Message: "Resource limit configured",
		})
	} else {
		results = append(results, ValidationResult{
			Check:   "Resource Limits",
			Status:  "WARN",
			Message: "No resource limit configured",
		})
	}

	// Resource Requests
	if checks.HasRequests {
		results = append(results, ValidationResult{
			Check:   "Resource Requests",
			Status:  "PASS",
			Message: "Resource request configured",
		})
	} else {
		results = append(results, ValidationResult{
			Check:   "Resource Requests",
			Status:  "WARN",
			Message: "No resource request configured",
		})
	}

	// Liveness Probe
	if checks.HasLivenessProbe {
		results = append(results, ValidationResult{
			Check:   "Liveness Probe",
			Status:  "PASS",
			Message: "Liveness probe found",
		})
	} else {
		results = append(results, ValidationResult{
			Check:   "Liveness Probe",
			Status:  "WARN",
			Message: "No liveness probe found",
		})
	}

	// Readiness Probe
	if checks.HasReadinessProbe {
		results = append(results, ValidationResult{
			Check:   "Readiness Probe",
			Status:  "PASS",
			Message: "Readiness probe found",
		})
	} else {
		results = append(results, ValidationResult{
			Check:   "Readiness Probe",
			Status:  "WARN",
			Message: "No readiness probe found",
		})
	}

	// Security Context
	if checks.HasSecurityCtx {
		results = append(results, ValidationResult{
			Check:   "Security Context",
			Status:  "PASS",
			Message: "Security context configured",
		})
	} else {
		results = append(results, ValidationResult{
			Check:   "Security Context",
			Status:  "WARN",
			Message: "No security context configured",
		})
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
