package buildcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

// GoBuild is the product of 'go build <arguments> -o <output> <pkg>' run in directory, a directory of the repository
// ("." for its root, "cohere" for cohere's module), with environment added: the stage 0 compiler, the checker archive,
// an oracle binary. It returns the built file's path. Its key is every file of every package in this repository the
// build compiles (from go list -deps under the same arguments and environment), each module's go.mod and go.sum (which
// pin every module outside it), the arguments, the build environment as go itself resolves it, and the Go and C
// toolchains. An -overlay build is keyed too when every file its overlay reads is inside the repository
// (@system_adamic, Oct 9 04:57Z): the key adds the overlay's map and each replacement file's content, and leaves out
// the overlay file's own path, which is a temporary name. One that reads a file outside the repository is refused:
// such a build stays the caller's own.
func GoBuild(t testing.TB, directory, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	return goProduct(t, false, directory, output, pkg, arguments, environment)
}

// GoTestBinary is GoBuild for 'go test -c': an oracle a test runs as a process, keyed with the package's test files and
// what only its test imports (#hff1651, Oct 9: products that ran go test -c by hand carried their checkout path and were
// keyed without the go flags, so a stored oracle never matched a rebuild from another tree).
func GoTestBinary(t testing.TB, directory, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	return goProduct(t, true, directory, output, pkg, arguments, environment)
}

func goProduct(t testing.TB, test bool, directory, output, pkg string, arguments []string, environment []string) string {
	t.Helper()
	arguments = reproducible(arguments)
	inputs, err := goInputs(test, directory, output, pkg, arguments, environment)
	if err != nil {
		t.Fatalf("build %s: %v", pkg, err)
	}
	root, _ := repositoryRoot()
	if slices.Contains(inputs.Flags, rebuildEverything) {
		arguments = append([]string{"-a"}, arguments...)
	}
	verb := []string{"build"}
	if test {
		verb = []string{"test", "-c"}
	}
	product := Product(t, inputs, func(product string) error {
		command := exec.Command("go", append(append(append(verb, arguments...), "-o", filepath.Join(product, output)), pkg)...)
		command.Dir = filepath.Join(root, directory)
		command.Env = append(os.Environ(), environment...)
		if combined, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("go %s %s: %v\n%s", strings.Join(verb, " "), pkg, err, combined)
		}
		return nil
	})
	return filepath.Join(product, output)
}

// reproducible adds what makes a build's bytes independent of where the checkout sits and which commit it's at:
// without -trimpath and an empty build ID a cgo product differs between two checkout paths in Go's build ID alone
// (measured Oct 9). Without -buildvcs=false every binary built in the repository stamps the commit (vcs.revision, 40
// hex), which the key leaves out on purpose, so two commits with the same inputs shared a key and differed in those
// bytes: floor1's audit rebuild of bridge/tsgo/oracle called it poisoning (Oct 9 04:52Z, same size, other bytes).
func reproducible(arguments []string) []string {
	kept := []string{"-trimpath", "-ldflags=-buildid=", "-buildvcs=false"}
	for _, argument := range arguments {
		if argument == "-trimpath" || strings.HasPrefix(argument, "-ldflags") || strings.HasPrefix(argument, "-buildvcs") {
			panic("buildcache.GoBuild and GoTestBinary set -trimpath, -ldflags and -buildvcs themselves, for a product that's the same from any checkout path and commit")
		}
	}
	return append(kept, arguments...)
}

// rebuildEverything is the flag a key carries when the build reads a header through '#cgo CFLAGS: -I' outside the
// package's own directory. Go's build cache keys a package by its own directory's files, so after such a header
// changes, go build alone links the object compiled from the old one (measured Oct 9 by
// TestProductIdentityAfterAHeaderOutsideThePackageChanges): the key moved, the product didn't. GoBuild builds such a
// product with -a, which compiles every package again, and the flag in the key retires a product built without it.
const rebuildEverything = "go build -a: a header outside its package"

// The go env values a build reads beyond its sources, resolved by go itself so a default counts like a setting.
var goEnvironment = []string{"GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "GOWORK", "CGO_ENABLED", "CC", "CXX",
	"CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOTOOLCHAIN"}

// GoInputs is GoBuild's key: what go list says the build compiles, with everything that can change its output.
func GoInputs(output, pkg string, arguments []string, environment []string) (Inputs, error) {
	return goInputs(false, ".", output, pkg, arguments, environment)
}

// goInputs is the key of a go build, or with test of a go test -c, of pkg as seen from directory, a directory of the
// repository: the build runs there, so its module and go.work are the ones go itself resolves.
func goInputs(test bool, directory, output, pkg string, arguments []string, environment []string) (Inputs, error) {
	root, err := repositoryRoot()
	if err != nil {
		return Inputs{}, err
	}
	overlayPath, keyed := "", []string{}
	for index := 0; index < len(arguments); index++ {
		switch argument := arguments[index]; {
		case argument == "-overlay" && index+1 < len(arguments):
			overlayPath = arguments[index+1]
			index++
		case strings.HasPrefix(argument, "-overlay="):
			overlayPath = strings.TrimPrefix(argument, "-overlay=")
		case argument == "-overlay":
			return Inputs{}, errors.New("-overlay names no file")
		default:
			keyed = append(keyed, argument)
		}
	}
	overlay, err := overlayInputs(root, overlayPath)
	if err != nil {
		return Inputs{}, err
	}
	run := func(arguments ...string) ([]byte, error) {
		command := exec.Command("go", arguments...)
		command.Dir = filepath.Join(root, directory)
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
	listed := listArguments(arguments)
	if overlayPath != "" {
		listed = append(listed, "-overlay="+overlayPath)
	}
	command := []string{"list", "-deps", "-json"}
	if test {
		command = append(command, "-test")
	}
	listing, err := run(append(append(command, listed...), pkg)...)
	if err != nil {
		return Inputs{}, err
	}
	files := map[string]bool{}
	outsideHeaders := false
	add := func(directory string, names []string) error {
		for _, name := range names {
			// With -test, a test variant lists the package's _test.go files too, and files go generated (its test main,
			// cgo's output) by absolute path in go's own cache: those follow from the inputs keyed here, so only files
			// inside the repository are named.
			path := filepath.Join(directory, name)
			if filepath.IsAbs(name) {
				if !inside(root, name) {
					continue
				}
				path = name
			}
			relative, err := filepath.Rel(root, path)
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
					if filepath.Clean(include) != filepath.Clean(listed.Dir) {
						outsideHeaders = true
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
	for _, target := range overlay.files {
		files[target] = true
	}
	// A root build keeps the name it always had, so its key doesn't move.
	name := "go build " + pkg + " " + output
	switch {
	case test:
		name = "go test -c " + directory + " " + pkg + " " + output
	case directory != ".":
		name = "go build " + directory + " " + pkg + " " + output
	}
	inputs := Inputs{Name: name, Flags: append([]string{"arguments " + strings.Join(keyed, " ")}, overlay.flags...), Toolchain: []string{Tool("go", "version")}}
	if outsideHeaders {
		inputs.Flags = append(inputs.Flags, rebuildEverything)
	}
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

// overlayKey is what an -overlay adds to a key: one flag per replaced path (the map, as go reads it, made relative to
// the repository) and the replacement files, whose content the key hashes like any other input file.
type overlayKey struct {
	flags, files []string
}

func overlayInputs(root, path string) (overlayKey, error) {
	if path == "" {
		return overlayKey{}, nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return overlayKey{}, fmt.Errorf("overlay %s: %v", path, err)
	}
	var declared struct{ Replace map[string]string }
	if err := json.Unmarshal(content, &declared); err != nil {
		return overlayKey{}, fmt.Errorf("overlay %s: %v", path, err)
	}
	relative := func(name string) string {
		if !filepath.IsAbs(name) {
			name = filepath.Join(root, name)
		}
		if inside(root, name) {
			return filepath.ToSlash(mustRelative(root, name))
		}
		return name
	}
	var key overlayKey
	for original, replacement := range declared.Replace {
		if replacement == "" {
			key.flags = append(key.flags, "overlay "+relative(original)+" deleted")
			continue
		}
		absolute := replacement
		if !filepath.IsAbs(absolute) {
			absolute = filepath.Join(root, absolute)
		}
		if !inside(root, absolute) {
			return overlayKey{}, fmt.Errorf("overlay %s replaces %s with %s, outside the repository, so it can't be keyed", path, original, replacement)
		}
		if _, err := os.Stat(absolute); err != nil {
			return overlayKey{}, fmt.Errorf("overlay %s: %v", path, err)
		}
		key.flags = append(key.flags, "overlay "+relative(original)+" = "+relative(absolute))
		key.files = append(key.files, relative(absolute))
	}
	sort.Strings(key.flags)
	return key, nil
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
