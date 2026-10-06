package nodepin

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestSetupVersion(t *testing.T) {
	t.Parallel()
	setup, err := os.ReadFile("../../cloud/setup.sh")
	if err != nil {
		t.Fatal(err)
	}
	versions := regexp.MustCompile(`(?m)^nodeVersion=(\S+)$`).FindAllSubmatch(setup, -1)
	if len(versions) != 1 {
		t.Fatalf("setup must declare nodeVersion once, got %d declarations", len(versions))
	}
	if version := string(versions[0][1]); version != Version {
		t.Fatalf("setup pins %s, nodepin.Version is %s", version, Version)
	}
}

func TestCheckUsesPATH(t *testing.T) {
	// Not parallel: the selected Node is process-wide.
	directory := t.TempDir()
	path := filepath.Join(directory, "node")
	t.Setenv("PATH", directory)
	for _, version := range []string{"v24.14.1", Version} {
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
