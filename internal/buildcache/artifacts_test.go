package buildcache

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestRequireArtifacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for _, name := range []string{"missing", "empty", "directory"} {
		if name == "empty" {
			write(t, directory, name, "")
		}
		if name == "directory" {
			if err := os.Mkdir(filepath.Join(directory, name), 0755); err != nil {
				t.Fatal(err)
			}
		}
		if err := requireArtifact(directory, Artifact{Name: name}); err == nil {
			t.Fatalf("accepted %s artifact", name)
		}
	}
	write(t, directory, "data", "built")
	RequireArtifacts(t, directory, Artifact{Name: "data"})
	for _, name := range []string{"", "../data", "/data"} {
		if err := requireArtifact(directory, Artifact{Name: name}); err == nil {
			t.Fatalf("accepted name %q", name)
		}
	}
	if err := requireArtifact("", Artifact{Name: "data"}); err == nil {
		t.Fatal("accepted empty directory")
	}
	if err := requireArtifact(directory, Artifact{Name: "data", Executable: true}); err == nil {
		t.Fatal("accepted unusable executable")
	}
	script := filepath.Join(directory, "probe")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nread value\n[ \"$value\" = built ] || exit 3\nexit 7\n"), 0755); err != nil {
		t.Fatal(err)
	}
	RequireArtifacts(t, directory, Artifact{Name: "probe", Executable: true, Stdin: "built\n", ExitCode: 7})
	if err := requireArtifact(directory, Artifact{Name: "probe", Executable: true, Stdin: "built\n"}); err == nil {
		t.Fatal("accepted wrong exit")
	}
}

func TestWithoutLeakDetection(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ environment, want []string }{
		{[]string{"HOME=/h"}, []string{"HOME=/h", "ASAN_OPTIONS=detect_leaks=0"}},
		{[]string{"ASAN_OPTIONS=halt_on_error=1", "HOME=/h"}, []string{"ASAN_OPTIONS=halt_on_error=1:detect_leaks=0", "HOME=/h"}},
		{[]string{"ASAN_OPTIONS=detect_leaks=1"}, []string{"ASAN_OPTIONS=detect_leaks=1:detect_leaks=0"}},
		{[]string{"ASAN_OPTIONS="}, []string{"ASAN_OPTIONS=detect_leaks=0"}},
	} {
		if got := withoutLeakDetection(test.environment); !slices.Equal(got, test.want) {
			t.Errorf("withoutLeakDetection(%q) = %q, want %q", test.environment, got, test.want)
		}
	}
}
