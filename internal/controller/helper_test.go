package controller

import (
	"reflect"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionStr  string
		expected    Version
		expectError bool
	}{
		{
			name:       "valid three part version",
			versionStr: "1.2.3",
			expected:   Version{Major: 1, Minor: 2, Patch: 3, Raw: "1.2.3"},
		},
		{
			name:       "valid two part version",
			versionStr: "2.5",
			expected:   Version{Major: 2, Minor: 5, Patch: 0, Raw: "2.5"},
		},
		{
			name:       "zero version",
			versionStr: "0.0.0",
			expected:   Version{Major: 0, Minor: 0, Patch: 0, Raw: "0.0.0"},
		},
		{
			name:       "large numbers",
			versionStr: "10.20.30",
			expected:   Version{Major: 10, Minor: 20, Patch: 30, Raw: "10.20.30"},
		},
		{
			name:        "single part version",
			versionStr:  "1",
			expectError: true,
		},
		{
			name:        "four part version",
			versionStr:  "1.2.3.4",
			expectError: true,
		},
		{
			name:        "invalid major version",
			versionStr:  "a.2.3",
			expectError: true,
		},
		{
			name:        "invalid minor version",
			versionStr:  "1.b.3",
			expectError: true,
		},
		{
			name:        "invalid patch version",
			versionStr:  "1.2.c",
			expectError: true,
		},
		{
			name:        "empty string",
			versionStr:  "",
			expectError: true,
		},
		{
			name:        "negative numbers",
			versionStr:  "1.-2.3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVersion(tt.versionStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseVersion() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseVersion() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseVersion() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected int
	}{
		{
			name:     "equal versions",
			v1:       Version{Major: 1, Minor: 2, Patch: 3},
			v2:       Version{Major: 1, Minor: 2, Patch: 3},
			expected: 0,
		},
		{
			name:     "v1 major > v2 major",
			v1:       Version{Major: 2, Minor: 0, Patch: 0},
			v2:       Version{Major: 1, Minor: 9, Patch: 9},
			expected: 1,
		},
		{
			name:     "v1 major < v2 major",
			v1:       Version{Major: 1, Minor: 9, Patch: 9},
			v2:       Version{Major: 2, Minor: 0, Patch: 0},
			expected: -1,
		},
		{
			name:     "v1 minor > v2 minor",
			v1:       Version{Major: 1, Minor: 3, Patch: 0},
			v2:       Version{Major: 1, Minor: 2, Patch: 9},
			expected: 1,
		},
		{
			name:     "v1 minor < v2 minor",
			v1:       Version{Major: 1, Minor: 2, Patch: 9},
			v2:       Version{Major: 1, Minor: 3, Patch: 0},
			expected: -1,
		},
		{
			name:     "v1 patch > v2 patch",
			v1:       Version{Major: 1, Minor: 2, Patch: 4},
			v2:       Version{Major: 1, Minor: 2, Patch: 3},
			expected: 1,
		},
		{
			name:     "v1 patch < v2 patch",
			v1:       Version{Major: 1, Minor: 2, Patch: 3},
			v2:       Version{Major: 1, Minor: 2, Patch: 4},
			expected: -1,
		},
		{
			name:     "zero versions",
			v1:       Version{Major: 0, Minor: 0, Patch: 0},
			v2:       Version{Major: 0, Minor: 0, Patch: 0},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			if result != tt.expected {
				t.Errorf("Version.Compare() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name:     "three part version",
			version:  Version{Major: 1, Minor: 2, Patch: 3, Raw: "1.2.3"},
			expected: "1.2.3",
		},
		{
			name:     "two part version",
			version:  Version{Major: 2, Minor: 5, Patch: 0, Raw: "2.5"},
			expected: "2.5",
		},
		{
			name:     "zero version",
			version:  Version{Major: 0, Minor: 0, Patch: 0, Raw: "0.0.0"},
			expected: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Version.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestVersionSatisfiesConstraint(t *testing.T) {
	v1_2_3, _ := parseVersion("1.2.3")
	v2_0_0, _ := parseVersion("2.0.0")
	v1_5_0, _ := parseVersion("1.5.0")

	tests := []struct {
		name       string
		version    Version
		constraint string
		expected   bool
	}{
		{
			name:       "no constraint",
			version:    v1_2_3,
			constraint: "",
			expected:   true,
		},
		{
			name:       "exact match",
			version:    v1_2_3,
			constraint: "1.2.3",
			expected:   true,
		},
		{
			name:       "exact match with equals",
			version:    v1_2_3,
			constraint: "=1.2.3",
			expected:   true,
		},
		{
			name:       "exact no match",
			version:    v1_2_3,
			constraint: "1.2.4",
			expected:   false,
		},
		{
			name:       "greater than or equal - equal",
			version:    v1_2_3,
			constraint: ">=1.2.3",
			expected:   true,
		},
		{
			name:       "greater than or equal - greater",
			version:    v2_0_0,
			constraint: ">=1.2.3",
			expected:   true,
		},
		{
			name:       "greater than or equal - less",
			version:    v1_2_3,
			constraint: ">=2.0.0",
			expected:   false,
		},
		{
			name:       "less than or equal - equal",
			version:    v1_2_3,
			constraint: "<=1.2.3",
			expected:   true,
		},
		{
			name:       "less than or equal - less",
			version:    v1_2_3,
			constraint: "<=2.0.0",
			expected:   true,
		},
		{
			name:       "less than or equal - greater",
			version:    v2_0_0,
			constraint: "<=1.2.3",
			expected:   false,
		},
		{
			name:       "greater than - true",
			version:    v2_0_0,
			constraint: ">1.2.3",
			expected:   true,
		},
		{
			name:       "greater than - false equal",
			version:    v1_2_3,
			constraint: ">1.2.3",
			expected:   false,
		},
		{
			name:       "greater than - false less",
			version:    v1_2_3,
			constraint: ">2.0.0",
			expected:   false,
		},
		{
			name:       "less than - true",
			version:    v1_2_3,
			constraint: "<2.0.0",
			expected:   true,
		},
		{
			name:       "less than - false equal",
			version:    v1_2_3,
			constraint: "<1.2.3",
			expected:   false,
		},
		{
			name:       "less than - false greater",
			version:    v2_0_0,
			constraint: "<1.2.3",
			expected:   false,
		},
		{
			name:       "invalid constraint",
			version:    v1_2_3,
			constraint: ">=invalid.version",
			expected:   false,
		},
		{
			name:       "constraint with spaces",
			version:    v1_5_0,
			constraint: ">= 1.2.3 ",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.satisfiesConstraint(tt.constraint)
			if result != tt.expected {
				t.Errorf("Version.satisfiesConstraint(%s) = %t, want %t", tt.constraint, result, tt.expected)
			}
		})
	}
}

func TestExtractVersionFromFilename(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		packageName string
		expected    string
		expectError bool
	}{
		{
			name:        "valid filename",
			filename:    "my-package-1.2.3.tar.gz",
			packageName: "my-package",
			expected:    "1.2.3",
		},
		{
			name:        "valid filename with complex name",
			filename:    "awesome-web-app-2.1.0.tar.gz",
			packageName: "awesome-web-app",
			expected:    "2.1.0",
		},
		{
			name:        "valid filename two part version",
			filename:    "simple-pkg-1.0.tar.gz",
			packageName: "simple-pkg",
			expected:    "1.0",
		},
		{
			name:        "package name with numbers",
			filename:    "package123-4.5.6.tar.gz",
			packageName: "package123",
			expected:    "4.5.6",
		},
		{
			name:        "wrong prefix",
			filename:    "other-package-1.2.3.tar.gz",
			packageName: "my-package",
			expectError: true,
		},
		{
			name:        "wrong suffix",
			filename:    "my-package-1.2.3.zip",
			packageName: "my-package",
			expectError: true,
		},
		{
			name:        "no version part",
			filename:    "my-package-.tar.gz",
			packageName: "my-package",
			expected:    "",
		},
		{
			name:        "filename without extension",
			filename:    "my-package-1.2.3",
			packageName: "my-package",
			expectError: true,
		},
		{
			name:        "empty filename",
			filename:    "",
			packageName: "my-package",
			expectError: true,
		},
		{
			name:        "filename equals package name",
			filename:    "my-package",
			packageName: "my-package",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractVersionFromFilename(tt.filename, tt.packageName)

			if tt.expectError {
				if err == nil {
					t.Errorf("extractVersionFromFilename() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("extractVersionFromFilename() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("extractVersionFromFilename() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestPackageCandidate(t *testing.T) {
	// Test that PackageCandidate struct works correctly
	version, err := parseVersion("1.2.3")
	if err != nil {
		t.Fatalf("Failed to parse version: %v", err)
	}

	candidate := PackageCandidate{
		Filename: "test-package-1.2.3.tar.gz",
		Version:  version,
	}

	if candidate.Filename != "test-package-1.2.3.tar.gz" {
		t.Errorf("PackageCandidate.Filename = %s, want %s", candidate.Filename, "test-package-1.2.3.tar.gz")
	}

	if candidate.Version.String() != "1.2.3" {
		t.Errorf("PackageCandidate.Version.String() = %s, want %s", candidate.Version.String(), "1.2.3")
	}
}

// Benchmark tests for performance-critical functions
func BenchmarkParseVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseVersion("1.2.3")
	}
}

func BenchmarkVersionCompare(b *testing.B) {
	v1, _ := parseVersion("1.2.3")
	v2, _ := parseVersion("2.1.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v1.Compare(v2)
	}
}

func BenchmarkVersionSatisfiesConstraint(b *testing.B) {
	version, _ := parseVersion("1.5.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		version.satisfiesConstraint(">=1.2.0")
	}
}

func BenchmarkExtractVersionFromFilename(b *testing.B) {
	for i := 0; i < b.N; i++ {
		extractVersionFromFilename("my-package-1.2.3.tar.gz", "my-package")
	}
}

// Edge case tests
func TestVersionEdgeCases(t *testing.T) {
	t.Run("version with leading zeros", func(t *testing.T) {
		version, err := parseVersion("01.02.03")
		if err != nil {
			t.Errorf("parseVersion() with leading zeros failed: %v", err)
		}
		if version.Major != 1 || version.Minor != 2 || version.Patch != 3 {
			t.Errorf("parseVersion() with leading zeros = %+v, want Major:1 Minor:2 Patch:3", version)
		}
	})

	t.Run("very large version numbers", func(t *testing.T) {
		version, err := parseVersion("999.888.777")
		if err != nil {
			t.Errorf("parseVersion() with large numbers failed: %v", err)
		}
		if version.Major != 999 || version.Minor != 888 || version.Patch != 777 {
			t.Errorf("parseVersion() with large numbers = %+v, want Major:999 Minor:888 Patch:777", version)
		}
	})
}

func TestConstraintEdgeCases(t *testing.T) {
	version, _ := parseVersion("1.2.3")

	t.Run("constraint with extra spaces", func(t *testing.T) {
		result := version.satisfiesConstraint("  >=  1.2.0  ")
		if !result {
			t.Errorf("satisfiesConstraint() with extra spaces should return true")
		}
	})

	t.Run("malformed constraint operator", func(t *testing.T) {
		result := version.satisfiesConstraint(">>1.2.0")
		if result {
			t.Errorf("satisfiesConstraint() with malformed operator should return false")
		}
	})
}
