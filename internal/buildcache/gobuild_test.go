package buildcache

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// GoBuild's key names every file the build compiles in this repository, the module files and a cgo header reached
// through -I, and the product runs.
func TestGoBuildKeysWhatTheBuildCompilesAndRuns(t *testing.T) {
	_, log := cached(t)
	inputs, err := GoInputs("withc", "./internal/buildcache/testdata/withc", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"internal/buildcache/testdata/withc/main.go", "internal/buildcache/testdata/include", "go.mod"} {
		if !slices.Contains(inputs.Files, want) {
			t.Fatalf("%s isn't in the key's files: %v", want, inputs.Files)
		}
	}
	if slices.Contains(inputs.Files, "internal/buildcache/testdata/hello/main.go") {
		t.Fatalf("a package the build doesn't compile is in its key: %v", inputs.Files)
	}
	if !slices.ContainsFunc(inputs.Flags, func(flag string) bool { return strings.HasPrefix(flag, "CGO_CFLAGS=") }) ||
		!slices.ContainsFunc(inputs.Toolchain, func(tool string) bool { return strings.HasPrefix(tool, "go version: ") }) {
		t.Fatalf("flags %v, toolchain %v", inputs.Flags, inputs.Toolchain)
	}
	binary := GoBuild(t, "withc", "./internal/buildcache/testdata/withc", nil)
	// Built path-independent: no Go build ID, which differs between two checkout paths.
	if id, err := exec.Command("go", "tool", "buildid", binary).Output(); err != nil || strings.TrimSpace(string(id)) != "" {
		t.Fatalf("the product's build ID is %q, %v", id, err)
	}
	output, err := exec.Command(binary).Output()
	if err != nil || strings.TrimSpace(string(output)) != "hello from a header" {
		t.Fatalf("the product printed %q, %v", output, err)
	}
	// Built once: the second call is a hit.
	GoBuild(t, "withc", "./internal/buildcache/testdata/withc", nil)
	if lines, _ := os.ReadFile(log); strings.Count(string(lines), " miss ") != 1 || strings.Count(string(lines), " hit ") != 1 {
		t.Fatalf("census: %q", lines)
	}
}

// The arguments and the environment change the key; an -overlay, which reads files outside the repository, is refused.
func TestGoBuildKeysArgumentsAndEnvironment(t *testing.T) {
	cached(t)
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	keyOf := func(arguments []string, environment ...string) string {
		t.Helper()
		inputs, err := GoInputs("hello", "./internal/buildcache/testdata/hello", arguments, environment)
		if err != nil {
			t.Fatal(err)
		}
		return key(t, root, inputs)
	}
	plain := keyOf(nil)
	for name, other := range map[string]string{
		"an argument":     keyOf([]string{"-trimpath"}),
		"the environment": keyOf(nil, "CGO_CFLAGS=-O1 -fsanitize=address"),
		"the compiler":    keyOf(nil, "CC=/usr/bin/no-such-cc"),
	} {
		if other == plain {
			t.Fatalf("%s left the key %s", name, plain)
		}
	}
	if _, err := GoInputs("hello", "./internal/buildcache/testdata/hello", []string{"-overlay", filepath.Join(t.TempDir(), "o.json")}, nil); err == nil {
		t.Fatal("an -overlay build was keyed")
	}
}

// GoBuild builds the same bytes from any checkout path, so a product another machine published audits clean.
func TestGoBuildIsReproducibleAcrossCheckoutPaths(t *testing.T) {
	arguments := reproducible([]string{"-buildmode=c-archive"})
	if !slices.Equal(arguments, []string{"-trimpath", "-ldflags=-buildid=", "-buildmode=c-archive"}) {
		t.Fatalf("arguments %v", arguments)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("a caller's own -ldflags was accepted")
		}
	}()
	reproducible([]string{"-ldflags=-s"})
}
