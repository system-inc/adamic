package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// The landing gate and the whole-module test run cover every module the repository builds, not only the
// root (#ejcnkja). `go test ./...` from the root reaches the root module's packages and stops at any
// directory with a go.mod of its own, so the TypeScript shims and static_single_assignment, each a module
// so another repository can import it, were never vetted or tested by any gate.
//
// The list is go.work's use directives, read with the parser `go work edit` uses rather than by running go,
// so the land fixture's stand-in go doesn't answer for it. A module inside a submodule (a gitlink in the
// index, by rule, never by name) is not gated: its tests are its upstream's. Everything gated runs in one
// go invocation, so it shares the token's -p budget and Go's test cache, and the gate says what it covered.
//
// A module go.work doesn't name is refused rather than left out: a go.mod anywhere in the tree outside a
// submodule must be one of go.work's uses, since a module missing from that list is a module no gate tests,
// which is how the shims went untested.

// workspaceModule is one module the gate covers: where it is, relative to the root, and its module path.
type workspaceModule struct {
	directory string
	path      string
}

// workspace is what a run covers: the modules it gates, and the uses it leaves to a submodule's upstream.
type workspace struct {
	gated         []workspaceModule
	inSubmodule   []string
	fromWorkspace bool
}

// readWorkspace reads the modules a run in root covers. Without a go.work it is the root module alone, the
// gate as it was.
func readWorkspace(root string) (workspace, error) {
	contents, err := os.ReadFile(filepath.Join(root, "go.work"))
	if errors.Is(err, fs.ErrNotExist) {
		return workspace{gated: []workspaceModule{{directory: "."}}}, nil
	}
	if err != nil {
		return workspace{}, err
	}
	work, err := modfile.ParseWork("go.work", contents, nil)
	if err != nil {
		return workspace{}, err
	}
	links, err := gitlinks(root)
	if err != nil {
		return workspace{}, err
	}

	covered := workspace{fromWorkspace: true}
	for _, use := range work.Use {
		directory := filepath.ToSlash(filepath.Clean(use.Path))
		if insideAny(directory, links) {
			covered.inSubmodule = append(covered.inSubmodule, directory)
			continue
		}
		module := workspaceModule{directory: directory}
		if directory != "." {
			goMod, err := os.ReadFile(filepath.Join(root, directory, "go.mod"))
			if err != nil {
				return workspace{}, fmt.Errorf("go.work uses %s: %w", directory, err)
			}
			module.path = modfile.ModulePath(goMod)
			if module.path == "" {
				return workspace{}, fmt.Errorf("go.work uses %s, whose go.mod names no module", directory)
			}
		}
		covered.gated = append(covered.gated, module)
	}
	if err := refuseUnlistedModules(root, work, links); err != nil {
		return workspace{}, err
	}
	return covered, nil
}

// refuseUnlistedModules walks the tree for go.mod files and refuses any that go.work doesn't use and that
// isn't inside a submodule, naming each with the use line that would list it. A testdata directory holds
// fixtures, not modules, and a dot directory or .cache holds nothing the repository builds, so neither is
// walked; nor is a submodule, whose modules are its upstream's.
func refuseUnlistedModules(root string, work *modfile.WorkFile, links []string) error {
	used := map[string]bool{}
	for _, use := range work.Use {
		used[filepath.ToSlash(filepath.Clean(use.Path))] = true
	}
	var unlisted []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			name := entry.Name()
			if relative != "." && (name == "testdata" || strings.HasPrefix(name, ".") || insideAny(relative, links)) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() != "go.mod" {
			return nil
		}
		directory := filepath.ToSlash(filepath.Dir(relative))
		if !used[directory] {
			unlisted = append(unlisted, directory)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(unlisted) == 0 {
		return nil
	}
	var message strings.Builder
	message.WriteString("go.work does not use every module in the tree, so a gate would leave these untested:")
	for _, directory := range unlisted {
		fmt.Fprintf(&message, "\n  %s/go.mod: add `./%s` to go.work's use", directory, directory)
	}
	return errors.New(message.String())
}

// patterns are the package patterns one go invocation from the root takes to cover every gated module: the
// root's ./..., and each other module's path with /..., which reaches its packages through the workspace.
func (covered workspace) patterns() []string {
	patterns := make([]string, 0, len(covered.gated))
	for _, module := range covered.gated {
		if module.directory == "." {
			patterns = append(patterns, "./...")
			continue
		}
		patterns = append(patterns, module.path+"/...")
	}
	return patterns
}

// report says what a run covers, so a module a gate missed shows as missing from its output.
func (covered workspace) report(output io.Writer) {
	if !covered.fromWorkspace {
		return
	}
	directories := make([]string, 0, len(covered.gated))
	for _, module := range covered.gated {
		directories = append(directories, module.directory)
	}
	fmt.Fprintf(output, "cohere-dev: gating %d modules from go.work: %s\n", len(directories), strings.Join(directories, ", "))
	if len(covered.inSubmodule) > 0 {
		fmt.Fprintf(output, "cohere-dev: not gated, inside a submodule: %s\n", strings.Join(covered.inSubmodule, ", "))
	}
}

// insideAny reports whether directory is one of paths or inside one.
func insideAny(directory string, paths []string) bool {
	for _, path := range paths {
		if directory == path || strings.HasPrefix(directory, path+"/") {
			return true
		}
	}
	return false
}

// gitlink is one submodule of a worktree: its path and the commit its tree pins.
type gitlink struct {
	path, pin string
}

// gitlinksOf lists a worktree's submodules from its index, where a gitlink is an entry of mode 160000.
func gitlinksOf(worktree string) ([]gitlink, error) {
	staged, err := gitOutput(worktree, "ls-files", "--stage")
	if err != nil {
		return nil, fmt.Errorf("listing %s's files: %w", worktree, err)
	}
	var links []gitlink
	for _, line := range strings.Split(staged, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] != "160000" {
			continue
		}
		links = append(links, gitlink{path: strings.Join(fields[3:], " "), pin: fields[1]})
	}
	return links, nil
}

// gitlinks is the paths of a worktree's submodules.
func gitlinks(worktree string) ([]string, error) {
	links, err := gitlinksOf(worktree)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(links))
	for _, link := range links {
		paths = append(paths, link.path)
	}
	return paths, nil
}
