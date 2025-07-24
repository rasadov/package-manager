package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rasadov/package-manager/config"
	"github.com/rasadov/package-manager/internal/ssh"
	"github.com/rasadov/package-manager/internal/utils"
)

// PackageCandidate represents a package file with parsed version
type PackageCandidate struct {
	Filename string
	Version  Version
}

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

// parseVersion parses a version string like "1.0.12" into a Version struct
func parseVersion(versionStr string) (Version, error) {
	parts := strings.Split(versionStr, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}
	if major < 0 {
		return Version{}, fmt.Errorf("invalid major version (negative): %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	if minor < 0 {
		return Version{}, fmt.Errorf("invalid minor version (negative): %s", parts[1])
	}

	patch := 0
	if len(parts) == 3 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
		}
		if patch < 0 {
			return Version{}, fmt.Errorf("invalid patch version (negative): %s", parts[2])
		}
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Raw:   versionStr,
	}, nil
}

// Compare compares two versions. Returns:
// -1 if v < other
//
//	0 if v == other
//
//	1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	return 0
}

// String returns the string representation of the version
func (v Version) String() string {
	return v.Raw
}

// satisfiesConstraint checks if version satisfies the given constraint
func (v Version) satisfiesConstraint(constraint string) bool {
	if constraint == "" {
		return true // No constraint means any version is acceptable
	}

	// Parse constraint (e.g., ">=1.0.0", "<=2.0.0", "1.0.0")
	constraint = strings.TrimSpace(constraint)

	var operator string
	var targetVersionStr string

	if strings.HasPrefix(constraint, ">=") {
		operator = ">="
		targetVersionStr = constraint[2:]
	} else if strings.HasPrefix(constraint, "<=") {
		operator = "<="
		targetVersionStr = constraint[2:]
	} else if strings.HasPrefix(constraint, ">") {
		operator = ">"
		targetVersionStr = constraint[1:]
	} else if strings.HasPrefix(constraint, "<") {
		operator = "<"
		targetVersionStr = constraint[1:]
	} else if strings.HasPrefix(constraint, "=") {
		operator = "="
		targetVersionStr = constraint[1:]
	} else {
		// No operator, assume exact match
		operator = "="
		targetVersionStr = constraint
	}

	targetVersion, err := parseVersion(strings.TrimSpace(targetVersionStr))
	if err != nil {
		return false // Invalid constraint
	}

	comparison := v.Compare(targetVersion)

	switch operator {
	case ">=":
		return comparison >= 0
	case "<=":
		return comparison <= 0
	case ">":
		return comparison > 0
	case "<":
		return comparison < 0
	case "=":
		return comparison == 0
	default:
		return false
	}
}

// extractVersionFromFilename extracts version from filename like "package-name-1.0.12.tar.gz"
func extractVersionFromFilename(filename, packageName string) (string, error) {
	prefix := packageName + "-"
	suffix := ".tar.gz"

	if !strings.HasPrefix(filename, prefix) || !strings.HasSuffix(filename, suffix) {
		return "", fmt.Errorf("filename doesn't match expected format")
	}

	// Remove prefix and suffix to get version
	versionStr := filename[len(prefix) : len(filename)-len(suffix)]
	return versionStr, nil
}

// findBestPackageVersion finds the best matching package version on the server
func findBestPackageVersion(sshClient *ssh.Client, pkg config.PackageRequest) (string, error) {
	// List files in remote directory
	files, err := sshClient.ListFiles(sshClient.GetRemoteDir())
	if err != nil {
		return "", fmt.Errorf("failed to list remote files: %w", err)
	}

	// Find matching packages and parse their versions
	var candidates []PackageCandidate
	prefix := pkg.Name + "-"
	suffix := ".tar.gz"

	for _, file := range files {
		if strings.HasPrefix(file, prefix) && strings.HasSuffix(file, suffix) {
			versionStr, err := extractVersionFromFilename(file, pkg.Name)
			if err != nil {
				fmt.Printf("Warning: Could not parse version from %s: %v\n", file, err)
				continue
			}

			version, err := parseVersion(versionStr)
			if err != nil {
				fmt.Printf("Warning: Invalid version format in %s: %v\n", file, err)
				continue
			}

			candidates = append(candidates, PackageCandidate{
				Filename: file,
				Version:  version,
			})
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no packages found for %s", pkg.Name)
	}

	fmt.Printf("Found %d candidate(s) for %s:\n", len(candidates), pkg.Name)
	for _, candidate := range candidates {
		fmt.Printf("  - %s (version %s)\n", candidate.Filename, candidate.Version)
	}

	// Filter candidates that satisfy version constraint
	var validCandidates []PackageCandidate
	for _, candidate := range candidates {
		if candidate.Version.satisfiesConstraint(pkg.Version) {
			validCandidates = append(validCandidates, candidate)
		}
	}

	if len(validCandidates) == 0 {
		return "", fmt.Errorf("no packages found for %s matching constraint %s", pkg.Name, pkg.Version)
	}

	// Sort by version (highest first)
	sort.Slice(validCandidates, func(i, j int) bool {
		return validCandidates[i].Version.Compare(validCandidates[j].Version) > 0
	})

	selected := validCandidates[0]
	fmt.Printf("Selected: %s (version %s) from %d valid candidates\n",
		selected.Filename, selected.Version, len(validCandidates))

	return selected.Filename, nil
}

// downloadAndInstallPackage downloads and extracts a single package
func downloadAndInstallPackage(sshClient *ssh.Client, pkg config.PackageRequest) error {
	// Find the best matching package version on server
	archiveName, err := findBestPackageVersion(sshClient, pkg)
	if err != nil {
		return fmt.Errorf("failed to find package version: %w", err)
	}

	// Create temporary directory for download
	tempDir, err := os.MkdirTemp("", "pm-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download archive
	remotePath := filepath.Join(sshClient.GetRemoteDir(), archiveName)
	localPath := filepath.Join(tempDir, archiveName)

	fmt.Printf("Downloading %s...\n", archiveName)
	if err := sshClient.DownloadFile(remotePath, localPath); err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	// Create installation directory
	installDir := filepath.Join("packages", pkg.Name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Extract archive
	fmt.Printf("Extracting %s to %s...\n", archiveName, installDir)
	if err := utils.ExtractTarGz(localPath, installDir); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	return nil
}
