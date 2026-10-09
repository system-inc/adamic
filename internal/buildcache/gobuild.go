package buildcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// GoBuild is the product of 'go build <arguments> -o <output> <pkg>' run at the repository root with environment
// added: the stage 0 compiler, the checker archive, an oracle binary. It returns the built file's path. Its key is
// every file of every package in this repository the build compiles (from go list -deps under the same arguments
// and environment), each module's go.mod and go.sum (which pin every module outside it), the arguments, the build
// environment as go itself resolves it, and the Go and C toolchains. An -overlay reads files from outside the
// repository, which no key here can name, so it is refused: such a build stays the caller's own.
func GoBuild(t testing.TB, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	arguments = reproducible(arguments)
	inputs, err := GoInputs(output, pkg, arguments, environment)
	if err != nil {
		t.Fatalf("build %s: %v", pkg, err)
	}
	root, _ := repositoryRoot()
	directory := Product(t, inputs, func(directory string) error {
		command := exec.Command("go", append(append(append([]string{"build"}, arguments...), "-o", filepath.Join(directory, output)), pkg)...)
		command.Dir = root
		command.Env = append(os.Environ(), environment...)
		if combined, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("go build %s: %v\n%s", pkg, err, combined)
		}
		return nil
	})
	return filepath.Join(directory, output)
}

// reproducible adds what makes a build's bytes independent of where the checkout sits: without them a cgo product
// differs between two checkout paths in Go's build ID alone (measured Oct 9), so a product fetched by another machine
// would fail the audit. Pure Go is already path-independent; -trimpath and an empty build ID make cgo so too.
func reproducible(arguments []string) []string {
	kept := []string{"-trimpath", "-ldflags=-buildid="}
	for _, argument := range arguments {
		if argument == "-trimpath" || strings.HasPrefix(argument, "-ldflags") {
			panic("buildcache.GoBuild sets -trimpath and -ldflags itself, for a product that's the same from any checkout path")
		}
	}
	return append(kept, arguments...)
}

// The go env values a build reads beyond its sources, resolved by go itself so a default counts like a setting.
var goEnvironment = []string{"GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "GOWORK", "CGO_ENABLED", "CC", "CXX",
	"CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOTOOLCHAIN"}

// GoInputs is GoBuild's key: what go list says the build compiles, with everything that can change its output.
func GoInputs(output, pkg string, arguments []string, environment []string) (Inputs, error) {
	for _, argument := range arguments {
		if argument == "-overlay" || strings.HasPrefix(argument, "-overlay=") {
			return Inputs{}, errors.New("an -overlay build reads files outside the repository, so it can't be keyed")
		}
	}
	root, err := repositoryRoot()
	if err != nil {
		return Inputs{}, err
	}
	run := func(arguments ...string) ([]byte, error) {
		command := exec.Command("go", arguments...)
		command.Dir = root
		command.Env = append(os.Environ(), environment...)
		output, err := command.Output()
		if err != nil {
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				return nil, fmt.Errorf("go %s: %v\n%s", strings.Join(arguments, " "), err, exit.Stderr)
			}
			return nil, err
		}
		return output, nil
	}
	listing, err := run(append(append([]string{"list", "-deps", "-json"}, listArguments(arguments)...), pkg)...)
	if err != nil {
		return Inputs{}, err
	}
	files := map[string]bool{}
	add := func(directory string, names []string) error {
		for _, name := range names {
			relative, err := filepath.Rel(root, filepath.Join(directory, name))
			if err != nil {
				return err
			}
			files[filepath.ToSlash(relative)] = true
		}
		return nil
	}
	decoder := json.NewDecoder(strings.NewReader(string(listing)))
	for {
		var listed struct {
			Dir                                                                        string
			Standard                                                                   bool
			Module                                                                     *struct{ GoMod string }
			GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
			CgoCFLAGS, CgoCPPFLAGS, CgoCXXFLAGS                                        []string
		}
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return Inputs{}, err
		}
		if listed.Standard || !inside(root, listed.Dir) {
			continue
		}
		for _, names := range [][]string{listed.GoFiles, listed.CgoFiles, listed.CFiles, listed.CXXFiles, listed.HFiles, listed.SFiles, listed.SysoFiles, listed.EmbedFiles} {
			if err := add(listed.Dir, names); err != nil {
				return Inputs{}, err
			}
		}
		// A header reached through '#cgo CFLAGS: -I<dir>' is an input too: every file of a directory inside the repository.
		for _, flags := range [][]string{listed.CgoCFLAGS, listed.CgoCPPFLAGS, listed.CgoCXXFLAGS} {
			for index, flag := range flags {
				include := strings.TrimPrefix(flag, "-I")
				if flag == "-I" && index+1 < len(flags) {
					include = flags[index+1]
				} else if include == flag {
					continue
				}
				if !filepath.IsAbs(include) {
					include = filepath.Join(listed.Dir, include)
				}
				if inside(root, include) {
					if err := add(root, []string{mustRelative(root, include)}); err != nil {
						return Inputs{}, err
					}
				}
			}
		}
		if listed.Module != nil && inside(root, listed.Module.GoMod) {
			module := filepath.Dir(listed.Module.GoMod)
			if err := add(module, existing(module, "go.mod", "go.sum")); err != nil {
				return Inputs{}, err
			}
		}
	}
	if err := add(root, existing(root, "go.work", "go.work.sum")); err != nil {
		return Inputs{}, err
	}
	resolved, err := run(append([]string{"env"}, goEnvironment...)...)
	if err != nil {
		return Inputs{}, err
	}
	values := strings.Split(strings.TrimRight(string(resolved), "\n"), "\n")
	inputs := Inputs{Name: "go build " + pkg + " " + output, Flags: []string{"arguments " + strings.Join(arguments, " ")}, Toolchain: []string{Tool("go", "version")}}
	for index, name := range goEnvironment {
		if index < len(values) {
			inputs.Flags = append(inputs.Flags, name+"="+values[index])
			if name == "CC" && values[index] != "" {
				inputs.Toolchain = append(inputs.Toolchain, Tool(strings.Fields(values[index])[0], "--version"))
			}
		}
	}
	for name := range files {
		inputs.Files = append(inputs.Files, name)
	}
	sort.Strings(inputs.Files)
	return inputs, nil
}

// listArguments are the build arguments that change which files go list reports: tags and build modes.
func listArguments(arguments []string) []string {
	var kept []string
	for index, argument := range arguments {
		switch {
		case strings.HasPrefix(argument, "-tags=") || strings.HasPrefix(argument, "-buildmode=") || strings.HasPrefix(argument, "-mod="):
			kept = append(kept, argument)
		case (argument == "-tags" || argument == "-mod") && index+1 < len(arguments):
			kept = append(kept, argument, arguments[index+1])
		}
	}
	return kept
}

func inside(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, "../") && !filepath.IsAbs(relative)
}

func mustRelative(root, path string) string {
	relative, _ := filepath.Rel(root, path)
	return relative
}

func existing(directory string, names ...string) []string {
	var found []string
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(directory, name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}
