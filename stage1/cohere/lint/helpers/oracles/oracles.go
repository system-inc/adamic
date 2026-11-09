// Package oracles owns the compiler-free Go oracle products for this port.
package oracles

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var producer bool

func Helpers(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../cohere")
	side, _ := filepath.Abs("testdata/oracle.go")
	catalog, _ := filepath.Abs("testdata/catalog.go")
	descriptors, _ := filepath.Abs("testdata/descriptors.go")
	virtual := filepath.Join(root, "adamic_helper_oracle.go")
	// Retain the typed registrations for the descriptor oracle before Entry erases the decoder type.
	registerPath := filepath.Join(root, "internal/lint/rule/register.go")
	register, err := os.ReadFile(registerPath)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "func Register(registrations ...Registrant) {"
	if strings.Count(string(register), anchor) != 1 {
		t.Fatal("registration capture anchor changed")
	}
	captured := strings.Replace(string(register), anchor, anchor+"\n\tAdamicRegistrations = append(AdamicRegistrations, registrations...)", 1)
	capturedPath, _ := filepath.Abs("testdata/register_capture.go")
	stored, err := os.ReadFile(capturedPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != captured {
		t.Fatal("registration capture overlay differs from pinned register.go")
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{registerPath: capturedPath, filepath.Join(root, "internal/lint/rule/adamic_bridge.go"): filepath.Join(filepath.Dir(descriptors), "rule_bridge.go"), virtual: side, filepath.Join(root, "policy/adamic_helpers.go"): catalog, filepath.Join(root, "adamic_helper_descriptors.go"): descriptors}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	// Cohere with the oracle laid over it is a product built ahead, keyed through the overlay (#5qykzj5).
	return stock(t, "go-oracle-owner-v1", virtual+" "+filepath.Join(root, "adamic_helper_descriptors.go"), []string{"-overlay=" + path})
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(t testing.TB, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	if producer {
		return buildcache.GoBuildIn(t, "cohere", output, pkg, arguments, environment...)
	}
	inputs, err := buildcache.GoBuildInInputs("cohere", output, pkg, arguments, nil, environment...)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, "Go oracle", inputs, "stage1/cohere/lint/helpers/oracles owners", func(string) error { return fmt.Errorf("settle the oracle owner first") })
	return filepath.Join(directory, output)
}
