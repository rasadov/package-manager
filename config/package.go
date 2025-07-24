package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type PacketTarget struct {
	Path    string   `json:"path" yaml:"path"`
	Exclude []string `json:"exclude,omitempty" yaml:"exclude,omitempty"`
}

func (pt *PacketTarget) UnmarshalJSON(data []byte) error {
	var pathStr string
	if err := json.Unmarshal(data, &pathStr); err == nil {
		pt.Path = pathStr
		return nil
	}

	type Alias PacketTarget
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(pt),
	}
	return json.Unmarshal(data, &aux)
}

type Dependency struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"ver,omitempty" yaml:"ver,omitempty"`
}

type PacketConfig struct {
	Name         string         `json:"name" yaml:"name"`
	Version      string         `json:"ver" yaml:"ver"`
	Targets      []PacketTarget `json:"targets" yaml:"targets"`
	Dependencies []Dependency   `json:"packets,omitempty" yaml:"packets,omitempty"`
}

type PackageRequest struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"ver,omitempty" yaml:"ver,omitempty"`
}

type PackagesConfig struct {
	Packages []PackageRequest `json:"packages" yaml:"packages"`
}

func LoadPacketConfig(filepath string) (*PacketConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config PacketConfig
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])

	switch ext {
	case "json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return &config, nil
}

func LoadPackagesConfig(filepath string) (*PackagesConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages file: %w", err)
	}

	var config PackagesConfig
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, ".")+1:])

	switch ext {
	case "json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON packages: %w", err)
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML packages: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported packages file format: %s", ext)
	}

	return &config, nil
}
