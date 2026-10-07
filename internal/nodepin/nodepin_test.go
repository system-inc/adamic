package nodepin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupVersion(t *testing.T) {
	t.Parallel()
	setup, err := os.ReadFile("../../cloud/node-pin.json")
	if err != nil {
		t.Fatal(err)
	}
	var pin struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(setup, &pin); err != nil {
		t.Fatal(err)
	}
	if pin.Version != Version {
		t.Fatalf("setup pins %s, nodepin.Version is %s", pin.Version, Version)
	}
}

func TestCheckUsesPATH(t *testing.T) {
	// Not parallel: the selected Node is process-wide.
	directory := t.TempDir()
	path := filepath.Join(directory, "node")
	t.Setenv("PATH", directory)
	for _, version := range []string{"v24.21.0", Version} {
		if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '"+version+"\\n'\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := Check()
		if got != path {
			t.Fatalf("Node path %q, want %q", got, path)
		}
		if version == Version {
			if err != nil {
				t.Fatal(err)
			}
		} else if err == nil || !strings.Contains(err.Error(), version) || !strings.Contains(err.Error(), Version) || !strings.Contains(err.Error(), path) {
			t.Fatalf("mismatch must name both versions and path, got %v", err)
		}
	}
}
