package compilercensus

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Compiler is the compiler's packages, by repository directory: what a compiler consumer depends on. A change to any
// of them selects every consumer.
var Compiler = []string{"internal/load", "internal/lower", "internal/native", "internal/javascript", "internal/ir", "internal/flow", "cmd/adamic"}

// The census's two kinds of evidence, as the gate tools' run.py reads them (compilerMarker and launchMarker): a
// reference to the compiler itself, and launching a process, which may run it. cmd/adamic is bounded at its name
// (compilerReference's check below, since Go's regexp has no lookahead), so cohere's cmd/adamic_<name> harness
// commands are no reference to it.
var (
	compilerImport = regexp.MustCompile(`github\.com/system-inc/adamic/internal/(?:load|lower|native|javascript|ir|flow)|oracle/adamic|ADAMIC_(?:BIN|BINARY)|"adamic"`)
	compilerDriver = regexp.MustCompile(`cmd/adamic`)
	launch         = regexp.MustCompile(`"os/exec"|os\.StartProcess|syscall\.Exec`)
)

// compilerReference is whether source names the compiler: one of compilerImport's forms, or cmd/adamic followed by
// anything but a letter, a digit or an underscore.
func compilerReference(source []byte) bool {
	if compilerImport.Match(source) {
		return true
	}
	for _, found := range compilerDriver.FindAllIndex(source, -1) {
		if found[1] == len(source) {
			return true
		}
		next := source[found[1]]
		if !(next == '_' || '0' <= next && next <= '9' || 'a' <= next && next <= 'z' || 'A' <= next && next <= 'Z') {
			return true
		}
	}
	return false
}

// Dependencies is the compiler dependency census of the repository at root, derived and never declared: each package
// with a tracked Go file that references the compiler or launches a process, and each of the compiler's own packages,
// depends on the whole of Compiler. A Go file belongs to its directory's package, or the nearest package above it (a
// testdata program is its package's). Launching a process counts because the process may be the compiler: a package
// that launches only other tools still depends on Compiler, which selects it more often than it needs and never less.
// The result is the "packages" of cloud/fast-gate/compiler-dependencies.json as it was hand-kept, where every entry
// named all of Compiler (79 on Oct 16), and readers take it in that file's place (#ng0ksed).
func Dependencies(root string) (map[string][]string, error) {
	packages, err := packageDirectories(root)
	if err != nil {
		return nil, err
	}
	tracked, err := output(toolchain.GitCommandIn(root, "ls-files", "-z", "--", "*.go"))
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	consumers := map[string]bool{}
	for _, directory := range Compiler {
		consumers[directory] = true
	}
	for _, name := range strings.Split(strings.TrimSuffix(string(tracked), "\x00"), "\x00") {
		if name == "" {
			continue
		}
		owner := path.Dir(name)
		for owner != "." && !packages[owner] {
			owner = path.Dir(owner)
		}
		if owner == "." && !packages["."] || consumers[owner] {
			continue
		}
		source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return nil, err
		}
		if compilerReference(source) || launch.Match(source) {
			consumers[owner] = true
		}
	}
	dependencies := map[string][]string{}
	for directory := range consumers {
		dependencies[directory] = slices.Clone(Compiler)
	}
	return dependencies, nil
}

// packageDirectories is every package of the repository's own module, by directory relative to root, as go list
// names them: a nested module's (cohere's) are its own, and testdata holds none.
func packageDirectories(root string) (map[string]bool, error) {
	listing, err := output(toolchain.CommandIn(root, "go", "list", "-e", "-f", "{{.Dir}}", "./..."))
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}
	absolute, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	directories := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
		if line == "" {
			continue
		}
		resolved, err := filepath.EvalSymlinks(line)
		if err != nil {
			return nil, err
		}
		relative, err := filepath.Rel(absolute, resolved)
		if err != nil {
			return nil, err
		}
		directories[filepath.ToSlash(relative)] = true
	}
	return directories, nil
}

// output is command's standard output, with its standard error in the error when it fails.
func output(command interface {
	Output() ([]byte, error)
}) ([]byte, error) {
	result, err := command.Output()
	if err != nil {
		var stderr []byte
		if exit, ok := err.(interface{ Stderr() []byte }); ok {
			stderr = exit.Stderr()
		}
		return nil, fmt.Errorf("%w: %s", err, bytes.TrimSpace(stderr))
	}
	return result, nil
}
