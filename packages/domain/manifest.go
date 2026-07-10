package domain

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest is the lab run description (YAML/JSON).
type Manifest struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   ManifestMetadata `json:"metadata" yaml:"metadata"`
	Spec       ManifestSpec     `json:"spec" yaml:"spec"`
}

type ManifestMetadata struct {
	Name string `json:"name" yaml:"name"`
}

type ManifestSpec struct {
	Lab     string         `json:"lab" yaml:"lab"`
	Adapter AdapterRef     `json:"adapter" yaml:"adapter"`
	Gates   []Gate         `json:"gates" yaml:"gates"`
	Policy  PolicyRef      `json:"policy" yaml:"policy"`
}

type AdapterRef struct {
	ID     string            `json:"id" yaml:"id"`
	Config map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
}

type PolicyRef struct {
	Pack string `json:"pack" yaml:"pack"`
}

// LoadManifest reads a YAML or JSON manifest from disk.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks required manifest fields.
func (m *Manifest) Validate() error {
	if m.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if m.Kind != "" && m.Kind != "LabManifest" {
		return fmt.Errorf("kind must be LabManifest, got %q", m.Kind)
	}
	if m.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if m.Spec.Lab == "" {
		return fmt.Errorf("spec.lab is required")
	}
	if m.Spec.Adapter.ID == "" {
		return fmt.Errorf("spec.adapter.id is required")
	}
	if len(m.Spec.Gates) == 0 {
		return fmt.Errorf("spec.gates must not be empty")
	}
	for i, g := range m.Spec.Gates {
		if g.ID == "" {
			return fmt.Errorf("spec.gates[%d].id is required", i)
		}
		if len(g.Checks) == 0 {
			return fmt.Errorf("spec.gates[%d].checks must not be empty", i)
		}
		for j, c := range g.Checks {
			if c.ID == "" {
				return fmt.Errorf("spec.gates[%d].checks[%d].id is required", i, j)
			}
			if c.Runner == "" {
				return fmt.Errorf("spec.gates[%d].checks[%d].runner is required", i, j)
			}
		}
	}
	return nil
}
