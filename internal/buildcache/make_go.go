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

// A goBuildPlan is one go product's build with everything it takes decided before its window opens (goRequest.plan):
// the go command resolved, its arguments, directory and environment, the overlay's map, each outside header's bytes
// and each mutated file's text. run, inside the window, writes only what the plan holds and runs go, so the build reads
// no state of buildcache's or toolchain's, which its lean key couldn't name (TestAGoProductsBuildReadsNoStateButItsInitializers,
// Tests' shape and Node's ruling on becf16eeba).
type goBuildPlan struct {
	// failed is why the plan couldn't be made, which fails the build, never the lookup of a product already built. A
	// string, since an error the build is handed would reach every error type of the module (leancensus).
	failed string
	// program is the go command, run in directory with environment, as verb, arguments (no -overlay among them), -o
	// output inside the product, then packages. verb and pkg are the request's, for the build's error.
	program, directory, output, verb, pkg string
	arguments, packages, environment      []string
	// replace is the request's own overlay's map, as it declares it.
	replace map[string]string
	headers []plannedHeader
	mutated []plannedMutation
}

// A plannedHeader is a header the build lays into a package (withOutsideHeaders' reason): target, the path in the
// package it replaces, and content, its bytes.
type plannedHeader struct {
	target  string
	content []byte
}

// A plannedMutation is one mutated file: file, its path in the repository, and text, every mutation of it applied.
// original is the file's absolute path, and pointers the overlay's entries whose replacement is it: run points each at
// the copy, or replaces original by the copy when none is.
type plannedMutation struct {
	file, text, original string
	pointers             []string
}

// run builds the plan's product into directory: each mutated file copied inside the product, under mutations/<its path
// in the repository>, and pointed at by the overlay; each outside header and the overlay itself written into a
// temporary directory, since the overlay names this checkout's paths, which leaves with the build (#tqrqx60); then go.
func (plan goBuildPlan) run(directory string) error {
	if plan.failed != "" {
		return errors.New(plan.failed)
	}
	replace := map[string]string{}
	maps.Copy(replace, plan.replace)
	for _, mutation := range plan.mutated {
		copied := filepath.Join(directory, "mutations", filepath.FromSlash(mutation.file))
		if err := os.MkdirAll(filepath.Dir(copied), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(copied, []byte(mutation.text), 0o644); err != nil {
			return err
		}
		for _, pointer := range mutation.pointers {
			replace[pointer] = copied
		}
		if len(mutation.pointers) == 0 {
			replace[mutation.original] = copied
		}
	}
	arguments := slices.Clone(plan.arguments)
	if len(replace) > 0 || len(plan.headers) > 0 {
		scratch, err := os.MkdirTemp("", "buildcache-overlay-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		for index, header := range plan.headers {
			path := filepath.Join(scratch, fmt.Sprintf("%d.h", index))
			if err := os.WriteFile(path, header.content, 0o644); err != nil {
				return err
			}
			replace[header.target] = path
		}
		encoded, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(scratch, "overlay.json"), encoded, 0o644); err != nil {
			return err
		}
		arguments = append(arguments, "-overlay="+filepath.Join(scratch, "overlay.json"))
	}
	command := exec.Command(plan.program, append(append(arguments, "-o", filepath.Join(directory, plan.output)), plan.packages...)...)
	command.Dir, command.Env = plan.directory, plan.environment
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("go %s %s: %v\n%s", plan.verb, plan.pkg, err, combined)
	}
	return nil
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
