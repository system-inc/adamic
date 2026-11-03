// Package hostconfiguration selects the Node build recorded by a host fixture.
package hostconfiguration

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"testing"
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
	var known map[string]struct{ Node, Configure, Ruling, Environment string }
	if err := json.Unmarshal(data, &known); err != nil {
		return "", "", err
	}
	spec, ok := known[configuration.Name]
	if !ok {
		return "", "", fmt.Errorf("row %d: unknown host configuration %q", row, configuration.Name)
	}
	if configuration.Node != spec.Node || configuration.Configure != spec.Configure || configuration.Ruling != spec.Ruling {
		return "", "", fmt.Errorf("row %d: invalid metadata for host configuration %s", row, configuration.Name)
	}
	binary = getenv(spec.Environment)
	if binary == "" {
		skip = fmt.Sprintf("row %d needs %s; set %s", row, configuration.Name, spec.Environment)
	}
	return binary, skip, nil
}

// Require is the only Go skip site for configured host fixtures. Judge keys its
// classification on this condition and Skip line, shared by every Go reader.
func Require(t testing.TB, row int, configuration *Configuration) string {
	t.Helper()
	binary, skip, err := Resolve(row, configuration, os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	if skip != "" {
		t.Skip(skip)
	}
	return binary
}
