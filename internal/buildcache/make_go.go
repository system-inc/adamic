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
// depend on, tests included, each under exports/, and their listing as listingName, every path relative to root. direct
// keeps only packages themselves, their test variants and external tests, and what any of those imports directly
// (GoDirectExports): the rest of -deps is listed by go and dropped, so neither its file nor its listing entry is kept.
func goExports(root string, packages []string, listingName, directory string, direct bool) error {
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
	keep := func(string) bool { return true }
	if direct {
		if keep, err = directPackages(root, real, packages, output); err != nil {
			return err
		}
	}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var listed ListedPackage
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if !keep(listed.ImportPath) {
			continue
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
		// Likewise a file go generated in its own cache (a test main) is named absolutely and left out, and so is that
		// package's export data: it names the generated file where go's cache sits, which -trimpath doesn't rewrite, so
		// two builders' caches gave a test main's export data two sizes (#dhm3wav: 4 of 339 archives, 2 bytes apart).
		// Nothing imports a test main, so no reader of the listing wants it.
		generated := slices.ContainsFunc(listed.GoFiles, filepath.IsAbs)
		for _, names := range []*[]string{&listed.GoFiles, &listed.TestGoFiles, &listed.XTestGoFiles} {
			*names = slices.DeleteFunc(*names, filepath.IsAbs)
		}
		if generated {
			listed.Export = ""
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

// directPackages says which of a go list -deps -test listing (output) GoDirectExports keeps, by import path: every
// entry whose directory is one of packages (the package, its test variant, its external test, its test main), and every
// path one of those imports, both as listed (a dependency go recompiled for a test, "path [p.test]") and plain, since
// go/importer asks for a plain path.
func directPackages(root, real string, packages []string, output []byte) (func(string) bool, error) {
	named := map[string]bool{}
	for _, pkg := range packages {
		named[filepath.ToSlash(filepath.Clean(pkg))] = true
	}
	kept := map[string]bool{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var listed struct {
			ImportPath, Dir string
			Imports         []string
		}
		if err := decoder.Decode(&listed); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		relative := ""
		for _, spelling := range []string{root, real} {
			if spelling != "" && inside(spelling, listed.Dir) {
				relative = filepath.ToSlash(mustRelative(spelling, listed.Dir))
				break
			}
		}
		if relative == "" || !named[relative] {
			continue
		}
		kept[listed.ImportPath] = true
		for _, imported := range listed.Imports {
			plain, _, _ := strings.Cut(imported, " [")
			kept[imported], kept[plain] = true, true
		}
	}
	if len(kept) == 0 {
		return nil, fmt.Errorf("go list -export listed none of %s", strings.Join(packages, " "))
	}
	return func(importPath string) bool { return kept[importPath] }, nil
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
