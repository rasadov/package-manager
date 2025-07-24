package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPacketTarget_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PacketTarget
		wantErr  bool
	}{
		{
			name:     "string path",
			input:    `"/path/to/target"`,
			expected: PacketTarget{Path: "/path/to/target"},
			wantErr:  false,
		},
		{
			name:  "full object",
			input: `{"path": "/path/to/target", "exclude": ["*.tmp", "*.log"]}`,
			expected: PacketTarget{
				Path:    "/path/to/target",
				Exclude: []string{"*.tmp", "*.log"},
			},
			wantErr: false,
		},
		{
			name:     "object with path only",
			input:    `{"path": "/path/to/target"}`,
			expected: PacketTarget{Path: "/path/to/target"},
			wantErr:  false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pt PacketTarget
			err := json.Unmarshal([]byte(tt.input), &pt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(pt, tt.expected) {
				t.Errorf("got %+v, want %+v", pt, tt.expected)
			}
		})
	}
}

func TestLoadPacketConfig(t *testing.T) {
	// Create temporary JSON config file
	jsonConfig := `{
		"name": "test-packet",
		"ver": "1.0.0",
		"targets": [
			"/path/to/target1",
			{
				"path": "/path/to/target2",
				"exclude": ["*.tmp"]
			}
		],
		"packets": [
			{
				"name": "dependency1",
				"ver": "2.0.0"
			},
			{
				"name": "dependency2"
			}
		]
	}`

	tmpFile := createTempFile(t, "config*.json", jsonConfig)
	defer os.Remove(tmpFile)

	config, err := LoadPacketConfig(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := &PacketConfig{
		Name:    "test-packet",
		Version: "1.0.0",
		Targets: []PacketTarget{
			{Path: "/path/to/target1"},
			{Path: "/path/to/target2", Exclude: []string{"*.tmp"}},
		},
		Dependencies: []Dependency{
			{Name: "dependency1", Version: "2.0.0"},
			{Name: "dependency2"},
		},
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("got %+v, want %+v", config, expected)
	}
}

func TestLoadPacketConfig_Errors(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		content     string
		expectError string
	}{
		{
			name:        "file not found",
			filename:    "nonexistent.json",
			expectError: "failed to read config file",
		},
		{
			name:        "invalid JSON",
			filename:    "invalid*.json",
			content:     `{invalid json}`,
			expectError: "failed to parse JSON config",
		},
		{
			name:        "unsupported format",
			filename:    "config*.txt",
			content:     "some content",
			expectError: "unsupported config file format: txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filepath string
			if tt.content != "" {
				filepath = createTempFile(t, tt.filename, tt.content)
				defer os.Remove(filepath)
			} else {
				filepath = tt.filename
			}

			_, err := LoadPacketConfig(filepath)
			if err == nil {
				t.Errorf("expected error containing %q but got none", tt.expectError)
				return
			}

			if !containsString(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestLoadPackagesConfig(t *testing.T) {
	jsonConfig := `{
		"packages": [
			{
				"name": "package1",
				"ver": "1.0.0"
			},
			{
				"name": "package2"
			}
		]
	}`

	tmpFile := createTempFile(t, "packages*.json", jsonConfig)
	defer os.Remove(tmpFile)

	config, err := LoadPackagesConfig(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := &PackagesConfig{
		Packages: []PackageRequest{
			{Name: "package1", Version: "1.0.0"},
			{Name: "package2"},
		},
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("got %+v, want %+v", config, expected)
	}
}

func TestLoadPackagesConfig_Errors(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		content     string
		expectError string
	}{
		{
			name:        "file not found",
			filename:    "nonexistent.json",
			expectError: "failed to read packages file",
		},
		{
			name:        "invalid JSON",
			filename:    "invalid*.json",
			content:     `{invalid json}`,
			expectError: "failed to parse JSON packages",
		},
		{
			name:        "unsupported format",
			filename:    "packages*.txt",
			content:     "some content",
			expectError: "unsupported packages file format: txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filepath string
			if tt.content != "" {
				filepath = createTempFile(t, tt.filename, tt.content)
				defer os.Remove(filepath)
			} else {
				filepath = tt.filename
			}

			_, err := LoadPackagesConfig(filepath)
			if err == nil {
				t.Errorf("expected error containing %q but got none", tt.expectError)
				return
			}

			if !containsString(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

func TestPacketConfig_Serialization(t *testing.T) {
	config := &PacketConfig{
		Name:    "test-packet",
		Version: "1.0.0",
		Targets: []PacketTarget{
			{Path: "/path/to/target1"},
			{Path: "/path/to/target2", Exclude: []string{"*.tmp"}},
		},
		Dependencies: []Dependency{
			{Name: "dependency1", Version: "2.0.0"},
			{Name: "dependency2"},
		},
	}

	// Test JSON serialization round-trip
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	var jsonConfig PacketConfig
	if err := json.Unmarshal(jsonData, &jsonConfig); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(&jsonConfig, config) {
		t.Errorf("JSON round-trip failed: got %+v, want %+v", &jsonConfig, config)
	}
}

func TestPackagesConfig_Serialization(t *testing.T) {
	config := &PackagesConfig{
		Packages: []PackageRequest{
			{Name: "package1", Version: "1.0.0"},
			{Name: "package2"},
		},
	}

	// Test JSON serialization round-trip
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	var jsonConfig PackagesConfig
	if err := json.Unmarshal(jsonData, &jsonConfig); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(&jsonConfig, config) {
		t.Errorf("JSON round-trip failed: got %+v, want %+v", &jsonConfig, config)
	}
}

func TestEmptyConfigs(t *testing.T) {
	// Test empty PacketConfig
	emptyPacketConfig := `{
		"name": "empty-packet",
		"ver": "1.0.0",
		"targets": []
	}`

	tmpFile := createTempFile(t, "empty*.json", emptyPacketConfig)
	defer os.Remove(tmpFile)

	config, err := LoadPacketConfig(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Targets) != 0 {
		t.Errorf("expected empty targets, got %d", len(config.Targets))
	}

	// Test empty PackagesConfig
	emptyPackagesConfig := `{
		"packages": []
	}`

	tmpFile2 := createTempFile(t, "empty_packages*.json", emptyPackagesConfig)
	defer os.Remove(tmpFile2)

	packagesConfig, err := LoadPackagesConfig(tmpFile2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(packagesConfig.Packages) != 0 {
		t.Errorf("expected empty packages, got %d", len(packagesConfig.Packages))
	}
}

// Helper functions

func createTempFile(t *testing.T, pattern, content string) string {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to write temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests

func BenchmarkLoadPacketConfig(b *testing.B) {
	jsonConfig := `{
		"name": "test-packet",
		"ver": "1.0.0",
		"targets": ["/path/to/target1", "/path/to/target2"],
		"packets": [{"name": "dependency1", "ver": "2.0.0"}]
	}`

	tmpFile, err := os.CreateTemp("", "bench*.json")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(jsonConfig)
	tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadPacketConfig(tmpFile.Name())
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkLoadPackagesConfig(b *testing.B) {
	jsonConfig := `{
		"packages": [
			{"name": "package1", "ver": "1.0.0"},
			{"name": "package2"}
		]
	}`

	tmpFile, err := os.CreateTemp("", "bench_packages*.json")
	if err != nil {
		b.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(jsonConfig)
	tmpFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadPackagesConfig(tmpFile.Name())
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
