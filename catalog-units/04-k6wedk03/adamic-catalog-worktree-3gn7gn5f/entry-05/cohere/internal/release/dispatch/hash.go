// Package dispatch resolves which cohere binary to run, rebuilding it when its inputs change.
//
// Rules are compiled into the binary, which is what makes them free to run. The cost of that choice
// is that adding a rule means rebuilding. This package pays that cost automatically and exactly
// once per change: the binary is named for a hash of everything that went into it, so a changed
// rule produces a different name, misses the cache, and rebuilds, while an unchanged tree hits and
// execs immediately.
//
// The correctness stake is the hash. A hash that misses an input serves a stale binary running old
// rules, and that looks precisely like success: the gate prints green, the rule file on disk says
// the rule is enforced, and neither is lying about what it knows. So the hash covers every input
// Go itself considers an input, discovered by asking the toolchain rather than by globbing paths.
package dispatch

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// hashLength is how much of the digest ends up in the filename. Eight bytes is 64 bits, which is
// far past collision risk for a cache holding tens of entries, and short enough to read.
const hashLength = 16

// Inputs is everything that can change what the built binary does.
//
// Every field here is a thing that, if it changed without the hash changing, would leave a stale
// binary looking like a fresh one. The set is deliberately wider than "the rules directory": the
// toolchain miscompiles differently across versions, the pinned TypeScript commit is most of
// the code, and the build flags change the binary that comes out.
type Inputs struct {
	// GoVersion is the toolchain identity, as `go env GOVERSION GOOS GOARCH`. A binary built by a
	// different compiler, or for a different platform, is a different binary.
	GoVersion string

	// TypeScriptGoCommit pins the vendored compiler. It is most of the code in the binary, and it
	// moves independently of anything in this repo.
	TypeScriptGoCommit string

	// BuildFlags are the flags handed to `go build`. These change the output, and `-trimpath` in
	// particular is a build-input change rather than a link flag: adding it invalidates the whole
	// compile cache, measured at 33s on a warm 2.0 GB cache.
	BuildFlags []string

	// SourceFiles are the absolute paths of every non-stdlib file the binary compiles, discovered
	// transitively from the entry point. Contents are hashed, not names: a rule edited in place
	// keeps its filename, and a hash over names alone would miss it entirely.
	SourceFiles []string
}

// Compute returns the hash naming the binary these inputs build.
//
// The digest covers file contents, not modification times. Times change when a branch is checked
// out and would cause a rebuild that changes nothing; contents are what actually decides what the
// binary does.
func (inputs Inputs) Compute() (string, error) {
	digest := sha256.New()

	// Each field is written with a label and a terminator so that no two different input sets can
	// serialize to the same bytes. Without the framing, moving a string from one field to another
	// would hash identically.
	fmt.Fprintf(digest, "goVersion\x00%s\x00", inputs.GoVersion)
	fmt.Fprintf(digest, "typeScriptCommit\x00%s\x00", inputs.TypeScriptGoCommit)
	for _, flag := range inputs.BuildFlags {
		fmt.Fprintf(digest, "buildFlag\x00%s\x00", flag)
	}

	// Sorted so the digest does not depend on the order the toolchain happened to list packages in.
	files := append([]string(nil), inputs.SourceFiles...)
	sort.Strings(files)

	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			// A file that was listed as an input and cannot be read is a loud failure. Hashing
			// around it would produce a stable digest for a tree we could not actually see, which
			// is the stale-binary failure wearing a different hat.
			return "", fmt.Errorf("reading build input %s: %w", path, err)
		}
		fmt.Fprintf(digest, "file\x00%s\x00%d\x00", filepath.Base(path), len(contents))
		digest.Write(contents)
	}

	return hex.EncodeToString(digest.Sum(nil))[:hashLength], nil
}

// CollectInputs asks the Go toolchain what the binary at packagePath is built from.
//
// The file list comes from `go list -deps` rather than from walking directories, because the
// toolchain's answer is the correct one by construction: it follows imports, honors build tags,
// and excludes tests. A directory walk would both miss inputs reached through an import we did not
// think of and include files that never compile in, and either error produces a hash that is wrong
// in a way nothing would catch.
func CollectInputs(moduleDirectory string, packagePath string, buildFlags []string) (Inputs, error) {
	goVersion, err := goEnvironment(moduleDirectory)
	if err != nil {
		return Inputs{}, err
	}

	commit, err := typeScriptCommit(moduleDirectory)
	if err != nil {
		return Inputs{}, err
	}

	sourceFiles, err := listSourceFiles(moduleDirectory, packagePath)
	if err != nil {
		return Inputs{}, err
	}
	if len(sourceFiles) == 0 {
		// An empty input set hashes to a perfectly stable digest for a tree containing nothing,
		// and every later run would hit that cache entry. A sweep that measured nothing must never
		// be able to look like a sweep that found nothing.
		return Inputs{}, fmt.Errorf("go list reported no source files for %s, which cannot be right", packagePath)
	}

	return Inputs{
		GoVersion:          goVersion,
		TypeScriptGoCommit: commit,
		BuildFlags:         buildFlags,
		SourceFiles:        sourceFiles,
	}, nil
}

// goEnvironment identifies the toolchain: version, operating system, architecture.
func goEnvironment(moduleDirectory string) (string, error) {
	command := exec.Command("go", "env", "GOVERSION", "GOOS", "GOARCH")
	command.Dir = moduleDirectory

	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("reading the Go toolchain version: %w", err)
	}

	return strings.Join(strings.Fields(string(output)), " "), nil
}

// typeScriptCommit reads the commit the vendored compiler's checkout is at, from its own files.
func typeScriptCommit(moduleDirectory string) (string, error) {
	commit, err := readHead(filepath.Join(moduleDirectory, "TypeScript"))
	if err != nil {
		return "", fmt.Errorf("reading the pinned TypeScript commit: %w", err)
	}
	return commit, nil
}

// listSourceFiles returns every non-stdlib file compiled into packagePath, transitively.
//
// Standard-library packages are excluded because they are already covered by the toolchain version:
// including them would hash thousands of files to learn something GOVERSION already said.
func listSourceFiles(moduleDirectory string, packagePath string) ([]string, error) {
	const template = `{{if not .Standard}}{{$directory := .Dir}}` +
		`{{range .GoFiles}}{{$directory}}/{{.}}{{"\n"}}{{end}}` +
		`{{range .EmbedFiles}}{{$directory}}/{{.}}{{"\n"}}{{end}}` +
		`{{end}}`

	command := exec.Command("go", "list", "-deps", "-f", template, packagePath)
	command.Dir = moduleDirectory

	output, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("listing build inputs for %s: %w\n%s", packagePath, err, exitError.Stderr)
		}
		return nil, fmt.Errorf("listing build inputs for %s: %w", packagePath, err)
	}

	files := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			files = append(files, trimmed)
		}
	}
	return files, nil
}
