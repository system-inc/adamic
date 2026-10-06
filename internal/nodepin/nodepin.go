// Package nodepin holds the Node version used as the oracle's truth.
package nodepin

import (
	"fmt"
	"os/exec"
	"strings"
)

const Version = "v24.19.0"

// Check verifies the Node on PATH and returns its path, including on a mismatch.
func Check() (string, error) {
	path, err := exec.LookPath("node")
	if err != nil {
		return "", fmt.Errorf("node: require %s on PATH: %w", Version, err)
	}
	output, err := exec.Command(path, "--version").Output()
	if err != nil {
		return path, fmt.Errorf("node at %s: require %s, --version failed: %w", path, Version, err)
	}
	version := strings.TrimSpace(string(output))
	if version != Version {
		return path, fmt.Errorf("node at %s: got %s, want %s", path, version, Version)
	}
	return path, nil
}
