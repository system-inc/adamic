package lowered

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/system-inc/adamic/internal/toolchain"
)

// loweringEntry is the lowering's two packages, whose Go closure go list computes: what a test that lowers links of the
// compiler, never internal/native or internal/javascript.
var loweringEntry = []string{"./internal/load", "./internal/lower", "./internal/native"}

// ownPackage is this package, which builds the product: its files are keyed beside the closure, its tests are not.
const ownPackage = "./internal/lowered"

// modulePath is the repository's module, the one whose packages the environment guard reads.
const modulePath = "github.com/system-inc/adamic"

// A closurePackage is one package go list names, as much of it as a key reads.
type closurePackage struct {
	ImportPath string
	Dir        string
	Standard   bool
	// Every file the package's build reads, by kind, as go list -json names them.
	GoFiles, CgoFiles, CFiles, CXXFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
	Module                                                                     *closureModule
	Error                                                                      *struct{ Err string }
}

// closureModule is the module go list reports a package in.
type closureModule struct{ Path, Version, GoMod, Dir string }

// A closure is a set of packages and the inputs that key them: the files in the repository, by their path from its root,
// and each file outside it (go's module cache) as an external= flag, by module, version and its path in the module,
// never by where this machine keeps the cache (estree's loom_family_products_test.go keys its closure the same way).
type closure struct {
	packages []closurePackage
	files    []string
	external []string
}

// stage1 is stage 1's closure, computed once a process: go list's answer is the tree's, which a test doesn't change.
var stage1 struct {
	once    sync.Once
	closure closure
	err     error
}

// loweringClosure is stage 1's closure: go list -deps of the lowering's packages, and this package's own files.
func loweringClosure() (closure, error) {
	stage1.once.Do(func() {
		root, err := repositoryRoot()
		if err != nil {
			stage1.err = err
			return
		}
		packages, err := listed(root, nil, true, loweringEntry...)
		if err != nil {
			stage1.err = err
			return
		}
		own, err := listed(root, nil, false, ownPackage)
		if err != nil {
			stage1.err = err
			return
		}
		stage1.closure, stage1.err = closureOf(root, append(packages, own...))
	})
	return stage1.closure, stage1.err
}

// listed is the packages go list names for patterns, with their dependencies when deps is set, run in root under
// toolchain's environment and extra.
func listed(root string, extra []string, deps bool, patterns ...string) ([]closurePackage, error) {
	arguments := []string{"list", "-json"}
	if deps {
		arguments = append(arguments, "-deps")
	}
	command := toolchain.CommandIn(root, "go", append(arguments, patterns...)...)
	command.Env = append(command.Env, extra...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go list %s: %v\n%s", strings.Join(patterns, " "), err, stderr.Bytes())
	}
	var packages []closurePackage
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var item closurePackage
		if err := decoder.Decode(&item); errors.Is(err, io.EOF) {
			return packages, nil
		} else if err != nil {
			return nil, err
		}
		if item.Error != nil {
			return nil, fmt.Errorf("go list %s: %s", item.ImportPath, item.Error.Err)
		}
		if !item.Standard {
			packages = append(packages, item)
		}
	}
}

// closureOf keys packages: every file their builds read, each module file pinning one, its go.sum, and go.work, which
// picks the modules.
func closureOf(root string, packages []closurePackage) (closure, error) {
	keyed := closure{packages: packages}
	files := map[string]bool{"go.mod": true, "go.work": true}
	external := map[string]bool{}
	add := func(path string, module *closureModule) error {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			files[filepath.ToSlash(relative)] = true
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := path
		if module != nil && module.Dir != "" {
			inside, err := filepath.Rel(module.Dir, path)
			if err != nil {
				return err
			}
			name = module.Path + "@" + module.Version + "/" + filepath.ToSlash(inside)
		}
		external[fmt.Sprintf("external=%s:%x", name, sha256.Sum256(data))] = true
		return nil
	}
	for _, item := range packages {
		for _, group := range [][]string{item.GoFiles, item.CgoFiles, item.CFiles, item.CXXFiles, item.HFiles, item.SFiles, item.SysoFiles, item.EmbedFiles} {
			for _, file := range group {
				if err := add(filepath.Join(item.Dir, file), item.Module); err != nil {
					return closure{}, err
				}
			}
		}
		if item.Module == nil || item.Module.GoMod == "" {
			continue
		}
		if err := add(item.Module.GoMod, item.Module); err != nil {
			return closure{}, err
		}
		sum := filepath.Join(filepath.Dir(item.Module.GoMod), "go.sum")
		if _, err := os.Stat(sum); err == nil {
			if err := add(sum, item.Module); err != nil {
				return closure{}, err
			}
		}
	}
	for file := range files {
		keyed.files = append(keyed.files, file)
	}
	for flag := range external {
		keyed.external = append(keyed.external, flag)
	}
	sort.Strings(keyed.files)
	sort.Strings(keyed.external)
	return keyed, nil
}

// repository is the repository's root, found once a process.
var repository struct {
	once      sync.Once
	directory string
	err       error
}

// repositoryRoot is the nearest directory at or above the working directory whose go.mod is the repository's module.
func repositoryRoot() (string, error) {
	repository.once.Do(func() {
		directory, err := os.Getwd()
		if err != nil {
			repository.err = err
			return
		}
		for {
			content, err := os.ReadFile(filepath.Join(directory, "go.mod"))
			if err == nil && strings.HasPrefix(string(content), "module "+modulePath+"\n") {
				repository.directory = directory
				return
			}
			parent := filepath.Dir(directory)
			if parent == directory {
				repository.err = errors.New("lowered: no adamic go.mod above the working directory")
				return
			}
			directory = parent
		}
	})
	return repository.directory, repository.err
}
