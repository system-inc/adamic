// Package hostconfiguration selects the Node build recorded by a host fixture.
package hostconfiguration

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

//go:embed configurations.json
var registry embed.FS

type Configuration struct {
	Name      string `json:"name"`
	Node      string `json:"node"`
	Configure string `json:"configure"`
	Ruling    string `json:"ruling"`
}

// Resolve never falls back to stock Node for a configured row. Metadata readers
// use it too: a missing recipe is a visible skip, an unknown recipe is an error.
func Resolve(row int, configuration *Configuration, getenv func(string) string) (binary, skip string, err error) {
	if configuration == nil {
		return "node", "", nil
	}
	data, err := registry.ReadFile("configurations.json")
	if err != nil {
		return "", "", err
	}
	var known map[string]struct{ Node, Configure, Ruling, Environment, Probe, Expect string }
	if err := json.Unmarshal(data, &known); err != nil {
		return "", "", err
	}
	spec, ok := known[configuration.Name]
	if !ok {
		return "", "", fmt.Errorf("row %d: unknown host configuration %q", row, configuration.Name)
	}
	if configuration.Node != spec.Node ||
		configuration.Configure != spec.Configure ||
		configuration.Ruling != spec.Ruling {
		return "", "", fmt.Errorf("row %d: invalid metadata for host configuration %s", row, configuration.Name)
	}
	binary = getenv(spec.Environment)
	if binary == "" {
		skip = fmt.Sprintf("row %d needs %s; set %s", row, configuration.Name, spec.Environment)
	} else if err := probe(binary, spec.Probe, spec.Expect); err != nil {
		return "", "", fmt.Errorf("row %d: %w", row, err)
	}
	return binary, skip, nil
}

// Both successful and failed probes are memoized, once per binary/recipe in a process.
var probes sync.Map

type probeResult struct {
	once sync.Once
	err  error
}

func probe(binary, expression, expected string) error {
	key := struct{ binary, expression, expected string }{binary, expression, expected}
	entry, _ := probes.LoadOrStore(key, &probeResult{})
	result := entry.(*probeResult)
	result.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary, "-p", expression).CombinedOutput()
		actual := strings.TrimSuffix(string(output), "\n")
		if err != nil || actual != expected {
			result.err = fmt.Errorf(
				"configuration probe %q printed %q; expected %q (execution: %v)",
				binary,
				actual,
				expected,
				err,
			)
		}
	})
	return result.err
}

// Skip returns a reason only for a valid configuration whose binary is missing.
// Errors return no skip: callers must continue to Require, which fails loudly.
// Each caller owns its t.Skip in a test file so the census can classify its test.
func Skip(row int, configuration *Configuration) string {
	_, why, err := Resolve(row, configuration, os.Getenv)
	if err != nil {
		return ""
	}
	return why
}

// Require resolves and verifies a binary without skipping. Callers first use
// Skip to classify an unavailable build at their own test's skip site.
func Require(test testing.TB, row int, configuration *Configuration) string {
	test.Helper()
	binary, skip, err := Resolve(row, configuration, os.Getenv)
	if err != nil {
		test.Fatal(err)
	}
	if skip != "" {
		test.Fatal(skip)
	}
	return binary
}
