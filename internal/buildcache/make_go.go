package buildcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/system-inc/adamic/internal/toolchain"
)

// A Mutation is one planted change a product is built with: File, a file of the repository the build reads (a
// package's source or an overlay's replacement), relative to the repository, with Before, which must occur in it exactly
// once, replaced by After. It is how a test proves a check can fail, built ahead like any other product.
type Mutation struct{ File, Before, After string }

// A goRequest is one go invocation a product is built by: 'go build', or 'go test -c', of pkg in module (a directory
// of the repository, its root when empty), with arguments and environment added.
type goRequest struct {
	verb, module, output, pkg string
	arguments, environment    []string
	mutations                 []Mutation
}

// build runs the request's go command, writing its output into directory, the product, with arguments as goProduct
// made them, and the request's mutations planted first when it has any.
func (request goRequest) build(root, directory string, arguments []string) error {
	if len(request.mutations) > 0 {
		planted, cleanup, err := request.plant(root, directory, arguments)
		if err != nil {
			return err
		}
		defer cleanup()
		arguments = planted
	}
	command := toolchain.Command("go", append(append(append(strings.Fields(request.verb), arguments...), "-o", filepath.Join(directory, request.output)), strings.Fields(request.pkg)...)...)
	command.Dir = filepath.Join(root, request.module)
	command.Env = toolchain.Env(request.environment...)
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("go %s %s: %v\n%s", request.verb, request.pkg, err, combined)
	}
	return nil
}

// plant writes each mutated file inside directory, the product, under mutations/<its path in the repository>, and
// returns arguments whose overlay lays the copies over what they mutate: the request's own overlay with every
// replacement that is a mutated file pointed at its copy, and a mutated package source replaced by its copy. The
// overlay names this checkout's paths, so it sits in a temporary directory and leaves with the build (#tqrqx60).
func (request goRequest) plant(root, directory string, arguments []string) ([]string, func(), error) {
	replace := map[string]string{}
	var kept []string
	for index := 0; index < len(arguments); index++ {
		path := ""
		switch argument := arguments[index]; {
		case argument == "-overlay" && index+1 < len(arguments):
			path = arguments[index+1]
			index++
		case strings.HasPrefix(argument, "-overlay="):
			path = strings.TrimPrefix(argument, "-overlay=")
		default:
			kept = append(kept, argument)
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		var declared struct{ Replace map[string]string }
		if err := json.Unmarshal(content, &declared); err != nil {
			return nil, nil, fmt.Errorf("overlay %s: %v", path, err)
		}
		maps.Copy(replace, declared.Replace)
	}
	texts, err := request.mutated(root)
	if err != nil {
		return nil, nil, err
	}
	for file, text := range texts {
		original := filepath.Join(root, file)
		copied := filepath.Join(directory, "mutations", filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(copied), 0o755); err != nil {
			return nil, nil, err
		}
		if err := os.WriteFile(copied, []byte(text), 0o644); err != nil {
			return nil, nil, err
		}
		pointed := false
		for from, to := range replace {
			if !filepath.IsAbs(to) {
				to = filepath.Join(root, request.module, to)
			}
			if filepath.Clean(to) == original {
				replace[from], pointed = copied, true
			}
		}
		if !pointed {
			replace[original] = copied
		}
	}
	scratch, err := os.MkdirTemp("", "buildcache-mutations-")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { os.RemoveAll(scratch) }
	encoded, err := json.Marshal(map[string]any{"Replace": replace})
	if err == nil {
		err = os.WriteFile(filepath.Join(scratch, "overlay.json"), encoded, 0o644)
	}
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return append(kept, "-overlay="+filepath.Join(scratch, "overlay.json")), cleanup, nil
}

// mutated is each file the request's mutations change, by its path in the repository, with every mutation of it
// applied in order. A mutation whose Before isn't in its file exactly once is refused, since it would plant nothing
// or plant it in the wrong place.
func (request goRequest) mutated(root string) (map[string]string, error) {
	texts := map[string]string{}
	for _, mutation := range request.mutations {
		text, ok := texts[mutation.File]
		if !ok {
			content, err := os.ReadFile(filepath.Join(root, mutation.File))
			if err != nil {
				return nil, fmt.Errorf("mutation of %s: %v", mutation.File, err)
			}
			text = string(content)
		}
		if count := strings.Count(text, mutation.Before); count != 1 {
			return nil, fmt.Errorf("mutation of %s: %q occurs %d times, not once", mutation.File, mutation.Before, count)
		}
		texts[mutation.File] = strings.Replace(text, mutation.Before, mutation.After, 1)
	}
	return texts, nil
}

// ReproducibleGoFlags are the flags GoBuild and GoTest add to every go build (reproducible). A product that runs go
// itself (from an overlay it generates, or through its own harness) adds them to its command and to its key, so its
// bytes name no checkout path or commit and a product built without them is retired: without them every oracle binary,
// and everything built from one, held its checkout's path (#j8wf2w4: 97 of a tree's products differed between Server
// and Workshop, and none at the same paths).
func ReproducibleGoFlags() []string {
	return []string{"-trimpath", "-ldflags=-buildid=", "-buildvcs=false"}
}

// reproducible adds what makes a build's bytes independent of where the checkout sits and which commit it's at:
// without -trimpath and an empty build ID a cgo product differs between two checkout paths in Go's build ID alone
// (measured Oct 9). Without -buildvcs=false every binary built in the repository stamps the commit (vcs.revision, 40
// hex), which the key leaves out on purpose, so two commits with the same inputs shared a key and differed in those
// bytes: floor1's audit rebuild of bridge/tsgo/oracle called it poisoning (Oct 9 04:52Z, same size, other bytes).
func reproducible(arguments []string) []string {
	kept := ReproducibleGoFlags()
	for _, argument := range arguments {
		if argument == "-trimpath" || strings.HasPrefix(argument, "-ldflags") || strings.HasPrefix(argument, "-buildvcs") {
			panic("buildcache.GoBuild and GoTest set -trimpath, -ldflags and -buildvcs themselves, for a product that's the same from any checkout path and commit")
		}
	}
	return append(kept, arguments...)
}

// goExports writes the product GoExports makes into directory: the export data of packages and every package they
// depend on, tests included, each under exports/, and their listing as listingName, every path relative to root.
func goExports(root string, packages []string, listingName, directory string) error {
	command := toolchain.Command("go", append([]string{"list", "-trimpath", "-buildvcs=false", "-export", "-deps", "-test", "-json"}, packages...)...)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return fmt.Errorf("go list -export: %v\n%s", err, exit.Stderr)
		}
		return err
	}
	if err := os.Mkdir(filepath.Join(directory, "exports"), 0o755); err != nil {
		return err
	}
	real := resolved(root)
	var listing []ListedPackage
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var listed ListedPackage
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		// A product holds no absolute path: a directory outside the repository (the standard library's, go's module
		// cache) is left out, since only its export data is read.
		relative := ""
		for _, spelling := range []string{root, real} {
			if spelling != "" && inside(spelling, listed.Dir) {
				relative = filepath.ToSlash(mustRelative(spelling, listed.Dir))
				break
			}
		}
		listed.Dir = relative
		// Likewise a file go generated in its own cache (a test main) is named absolutely and left out.
		for _, names := range []*[]string{&listed.GoFiles, &listed.TestGoFiles, &listed.XTestGoFiles} {
			*names = slices.DeleteFunc(*names, filepath.IsAbs)
		}
		if listed.Export != "" {
			content, err := os.ReadFile(listed.Export)
			if err != nil {
				return err
			}
			listed.Export = fmt.Sprintf("exports/%d.a", len(listing))
			if err := os.WriteFile(filepath.Join(directory, listed.Export), content, 0o644); err != nil {
				return err
			}
		}
		listing = append(listing, listed)
	}
	encoded, err := json.MarshalIndent(listing, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, listingName), encoded, 0o644)
}

// A ListedPackage is one package of a GoExports listing, as go list reports it, except that Dir is relative to the
// repository (empty for a package outside it), a file go generated in its own cache is left out, and Export names a
// file in the product.
type ListedPackage struct {
	ImportPath, Dir, Export, ForTest   string
	Standard                           bool
	GoFiles, TestGoFiles, XTestGoFiles []string
}

func inside(root, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, "../") && !filepath.IsAbs(relative)
}

func mustRelative(root, path string) string {
	relative, _ := filepath.Rel(root, path)
	return relative
}
