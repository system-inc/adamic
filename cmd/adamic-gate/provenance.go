package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type plainProvenance struct {
	Commit string `json:"commit"`
	Go     string `json:"go"`
	Node   string `json:"node"`
}

func readPlainProvenance(log string) (plainProvenance, error) {
	var p plainProvenance
	data, err := os.ReadFile(log + ".provenance.json")
	if err == nil {
		if err := json.Unmarshal(data, &p); err != nil {
			return p, fmt.Errorf("plain provenance sidecar: %w", err)
		}
		return p, nil
	}
	if !os.IsNotExist(err) {
		return p, err
	}
	data, err = os.ReadFile(filepath.Join(filepath.Dir(log), "run-notes.txt"))
	if err != nil {
		return p, fmt.Errorf("plain provenance requires %s.provenance.json or sibling run-notes.txt: %w", log, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		var dst *string
		switch key {
		case "commit":
			dst = &p.Commit
		case "go", "go_version":
			dst = &p.Go
		case "node", "node_version":
			dst = &p.Node
		default:
			continue
		}
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, `"`) {
			decoded, err := strconv.Unquote(value)
			if err != nil {
				return p, fmt.Errorf("plain %s notes: %w", key, err)
			}
			value = decoded
		}
		if *dst != "" {
			return p, fmt.Errorf("plain notes have duplicate %s records", key)
		}
		*dst = value
	}
	return p, nil
}

// Old plans record toolchain versions in each shard's build-flags line.
// Use that evidence, never the toolchain installed on the comparison machine.
func plannedVersion(m merged, key, explicit string) (string, error) {
	re := regexp.MustCompile(`(?:^|\s)` + key + `=("(?:[^"\\]|\\.)*")`)
	expected := explicit
	for _, flags := range m.BuildFlags {
		match := re.FindStringSubmatch(flags)
		if match == nil {
			return "", fmt.Errorf("plan %s version missing from shard build flags", key)
		}
		value, err := strconv.Unquote(match[1])
		if err != nil {
			return "", err
		}
		if expected != "" && expected != value {
			return "", fmt.Errorf("plan %s versions differ across recorded toolchains", key)
		}
		expected = value
	}
	if expected == "" {
		return "", fmt.Errorf("plan %s version is missing", key)
	}
	return expected, nil
}

func checkPlainProvenance(log string, m merged) error {
	p, err := readPlainProvenance(log)
	if err != nil {
		return err
	}
	if !fullCommit.MatchString(p.Commit) || p.Commit != m.Plan.Commit {
		return fmt.Errorf("plain commit %q differs from plan commit %q", p.Commit, m.Plan.Commit)
	}
	goVersion, err := plannedVersion(m, "go", m.Plan.GoVersion)
	if err != nil {
		return err
	}
	nodeVersion, err := plannedVersion(m, "node", m.Plan.NodeVersion)
	if err != nil {
		return err
	}
	if p.Go == "" || p.Go != goVersion {
		return fmt.Errorf("plain Go version %q differs from plan Go version %q", p.Go, goVersion)
	}
	if p.Node == "" || p.Node != nodeVersion {
		return fmt.Errorf("plain Node version %q differs from plan Node version %q", p.Node, nodeVersion)
	}
	return nil
}
