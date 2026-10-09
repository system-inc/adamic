package buildcache

import (
	"crypto/sha256"
	"fmt"
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
	// GoBuild built it with the reproducible flags, as its recorded inputs say.
	if described, err := os.ReadFile(filepath.Dir(binary) + ".inputs"); err != nil || !strings.Contains(string(described), "flag arguments -trimpath -ldflags=-buildid=") {
		t.Fatalf("the product's inputs: %q, %v", described, err)
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

// GoBuild builds the same bytes from any checkout path, so a product another machine published audits clean: the cgo
// package, copied into two module directories at different paths and built with GoBuild's flags, is byte-identical.
func TestGoBuildIsReproducibleAcrossCheckoutPathsAndCommits(t *testing.T) {
	root, err := repositoryRoot()
	if err != nil {
		t.Fatal(err)
	}
	var sums []string
	for _, name := range []string{"one", "elsewhere/two"} {
		module := filepath.Join(t.TempDir(), name)
		for _, file := range []string{"withc/main.go", "include/greeting.h"} {
			content, err := os.ReadFile(filepath.Join(root, "internal/buildcache/testdata", file))
			if err != nil {
				t.Fatal(err)
			}
			os.MkdirAll(filepath.Join(module, filepath.Dir(file)), 0o755)
			os.WriteFile(filepath.Join(module, file), content, 0o644)
		}
		os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.com/reproducible\n\ngo 1.21\n"), 0o644)
		// Each checkout is its own repository at its own commit, as two gate trees are: Go stamps a binary built in a
		// repository with the commit unless -buildvcs=false.
		for _, git := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "commit " + name}} {
			if output, err := exec.Command("git", append([]string{"-C", module}, git...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v\n%s", git, err, output)
			}
		}
		command := exec.Command("go", append(append([]string{"build"}, reproducible(nil)...), "-o", filepath.Join(module, "out"), "./withc")...)
		command.Dir = module
		command.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, output)
		}
		content, err := os.ReadFile(filepath.Join(module, "out"))
		if err != nil {
			t.Fatal(err)
		}
		sums = append(sums, fmt.Sprintf("%x", sha256.Sum256(content)))
	}
	if sums[0] != sums[1] {
		t.Fatalf("two checkout paths at two commits built %s and %s", sums[0], sums[1])
	}
	arguments := reproducible([]string{"-buildmode=c-archive"})
	if !slices.Equal(arguments, []string{"-trimpath", "-ldflags=-buildid=", "-buildvcs=false", "-buildmode=c-archive"}) {
		t.Fatalf("arguments %v", arguments)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("a caller's own -ldflags was accepted")
		}
	}()
	reproducible([]string{"-ldflags=-s"})
}
