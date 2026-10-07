package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostMetadataWithoutGNUCommandsOnDarwin(t *testing.T) {
	bin := t.TempDir()
	marker := filepath.Join(bin, "nproc-called")
	t.Setenv("GATE_NPROC_MARKER", marker)
	tools := map[string]string{
		"nproc":  "#!/bin/sh\nprintf 'called\\n' > \"$GATE_NPROC_MARKER\"\nprintf '4\\n'\n",
		"sysctl": "#!/bin/sh\ncase \"$*\" in\n '-n hw.logicalcpu') printf '4\\n';;\n '-n vm.loadavg') printf '{ 1.25 0.75 0.50 }\\n';;\n *) exit 2;;\nesac\n",
	}
	for name, source := range tools {
		if err := os.WriteFile(filepath.Join(bin, name), []byte(source), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if processorCount() != "4" {
		t.Fatal("processor count unavailable")
	}
	if hostPlatform == "darwin" {
		if _, err := os.Stat(marker); !os.IsNotExist(err) {
			t.Fatal("Darwin invoked GNU nproc", err)
		}
		if loadAverage() != "1.25 0.75 0.50" || cpuQuota() != "not-applicable" {
			t.Fatal("Darwin metadata used Linux interfaces", loadAverage(), cpuQuota())
		}
		if err := os.Remove(filepath.Join(bin, "nproc")); err != nil {
			t.Fatal(err)
		}
		if processorCount() != "4" {
			t.Fatal("Darwin required nproc")
		}
	}
}

func TestFixtureDiscoveryKeepsGoWarningOffJSON(t *testing.T) {
	root, warningRoot := t.TempDir(), t.TempDir()
	for path, text := range map[string]string{
		filepath.Join(root, "go.mod"):        "module discoveryprobe\n\ngo 1.27\n",
		filepath.Join(root, "fixture.a"):     "fixture\n",
		filepath.Join(warningRoot, "go.mod"): "module ignored\n\ngo 1.27\n",
		filepath.Join(root, "internal/oracle/oracle_test.go"): `package oracle
import "testing"
var fixtures=[]struct{path string}{{"fixture.a"}}
func TestParent(t *testing.T) {}
`,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
	t.Setenv("GOPATH", warningRoot)
	t.Setenv("GOWORK", "off")
	t.Setenv("GO111MODULE", "on")
	names, err := registeredFixtures("internal/oracle/oracle_test.go", "fixtures")
	if err != nil || strings.Join(names, ",") != "fixture.a" {
		t.Fatal("fixture discovery lost stdout purity", names, err)
	}
	listing, err := output("go", "test", "-count=1", "-json", "-run", "^$", "./internal/oracle")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(listing, "\n") {
		var e event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal("structured stdout contains Go stderr", line, err)
		}
	}
}
