package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/boundedrun"
	"os"
	"path/filepath"
	"strings"
)

// Observe the registered slice after Go executes all init functions, including conditional
// registrations. An overlay adds only a listing test; repository tests are never edited.
func registeredFixtures(file, variable string) ([]string, error) {
	source, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	root, err := os.MkdirTemp("", "adamic-fixture-discovery-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)
	marker := "adamic-gate registered fixture: "
	source = append(source, []byte(fmt.Sprintf("\nfunc TestAdamicGateRegisteredFixtures(t *testing.T) { for _, row := range %s { t.Log(%q + row.path) } }\n", variable, marker))...)
	replacement := filepath.Join(root, "listing.go")
	if err := os.WriteFile(replacement, source, 0600); err != nil {
		return nil, err
	}
	original, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	overlay := filepath.Join(root, "overlay.json")
	if err := saveJSON(overlay, map[string]any{"Replace": map[string]string{original: replacement}}); err != nil {
		return nil, err
	}
	logPath := filepath.Join(root, "discovery.jsonl")
	cmd, release := boundedrun.Command(boundedrun.Build, "go", "test", "-overlay="+overlay, "-count=1", "-json", "-run", "^TestAdamicGateRegisteredFixtures$", "./"+filepath.ToSlash(filepath.Dir(file)))
	defer release()
	commandErr := runJSONCommand(cmd, logPath)
	bytes, readErr := os.ReadFile(logPath)
	if commandErr != nil || readErr != nil {
		return nil, fmt.Errorf("fixture discovery failed: %v %v\n%s", commandErr, readErr, bytes)
	}
	listing := string(bytes)

	var names []string
	seen := map[string]bool{}
	passed := false
	lines := strings.Split(listing, "\n")
	for index, line := range lines {
		if line == "" && index == len(lines)-1 {
			continue
		}
		var event event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("%s:%d: malformed JSON: %w", logPath, index+1, err)
		}
		if event.Test == "TestAdamicGateRegisteredFixtures" && event.Action == "pass" {
			passed = true
		}
		if at := strings.Index(event.Output, marker); at >= 0 {
			name := strings.TrimSpace(event.Output[at+len(marker):])
			if seen[name] {
				return nil, fmt.Errorf("duplicate registered fixture %s", name)
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	if !passed || len(names) == 0 {
		return nil, fmt.Errorf("fixture discovery incomplete for %s", variable)
	}
	for _, name := range names {
		if _, err := os.Stat(name); err != nil {
			return nil, err
		}
	}
	return names, nil
}
