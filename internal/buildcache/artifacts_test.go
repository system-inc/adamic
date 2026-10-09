package buildcache

import (
	"os"
	"path/filepath"
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
