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
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache/goplan"
	"github.com/system-inc/adamic/internal/toolchain"
)

// GoBuild is the product of 'go build <arguments> -o <output> <pkg>' run at the repository root with environment
// added: the stage 0 compiler, the checker archive, an oracle binary. It returns the built file's path. Its key is
// every file of every package in this repository the build compiles (from go list -deps under the same arguments
// and environment), each module's go.mod and go.sum (which pin every module outside it), the arguments, the build
// environment as go itself resolves it, and the Go and C toolchains. An -overlay build is keyed too when every file
// its overlay reads is inside the repository (@system_adamic, Oct 9 04:57Z): the key adds the overlay's map and each
// replacement file's content, and leaves out the overlay file's own path, which is a temporary name. One that reads a
// file outside the repository is refused: such a build stays the caller's own.
func GoBuild(t testing.TB, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	return goProduct(t, goRequest{verb: "build", output: output, pkg: pkg, arguments: arguments, environment: environment})
}

// GoTest is GoBuild for a test binary: the product of 'go test -c <arguments> -o <output> <pkg>' run in module, a
// directory of the repository with a go.mod of its own (cohere, cohere/TypeScript/tsc), or the root when empty. It is
// how a port's test builds its Go oracle: a package of the module with a harness laid over it by -overlay, then run.
// Its key is GoBuild's, with the test files go list -test names, and the module's own go.mod, go.sum and go.work. An
// overlay file that adds a path the module doesn't have (a harness, a synthesized main) is keyed through the overlay,
// never as a file on disk.
func GoTest(t testing.TB, module, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	return goProduct(t, goRequest{verb: "test -c", module: module, output: output, pkg: pkg, arguments: arguments, environment: environment})
}

// GoBuildIn is GoBuild run in module, a directory of the repository with a go.mod of its own (cohere): how a port
// builds a Go oracle from a main its overlay synthesizes. pkg may name several files, separated by spaces, as go
// build takes them.
func GoBuildIn(t testing.TB, module, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	return goProduct(t, goRequest{verb: "build", module: module, output: output, pkg: pkg, arguments: arguments, environment: environment})
}

// GoTestMutated is GoTest with mutations planted. Each mutated copy is written inside the product while it builds and
// laid over the original by an overlay of its own, so the product names no machine. The key is GoTest's with each
// mutation's file, before and after: a changed after-string is another product.
func GoTestMutated(t testing.TB, module, output, pkg string, arguments []string, mutations []Mutation, environment ...string) string {
	t.Helper()
	return goProduct(t, goRequest{verb: "test -c", module: module, output: output, pkg: pkg, arguments: arguments, mutations: mutations, environment: environment})
}

// GoBuildInMutated is GoBuildIn with mutations planted, as GoTestMutated plants them.
func GoBuildInMutated(t testing.TB, module, output, pkg string, arguments []string, mutations []Mutation, environment ...string) string {
	t.Helper()
	return goProduct(t, goRequest{verb: "build", module: module, output: output, pkg: pkg, arguments: arguments, mutations: mutations, environment: environment})
}

// GoBuildInInputs is GoBuildInMutated's key (GoBuildIn's when mutations is nil), computed without building: for a
// product that records or runs the binary, which can be keyed by the binary's key rather than its bytes. The arguments
// are made reproducible as the product's own are, so the key is the product's.
func GoBuildInInputs(module, output, pkg string, arguments []string, mutations []Mutation, environment ...string) (Inputs, error) {
	return goRequest{verb: "build", module: module, output: output, pkg: pkg, arguments: reproducible(arguments), mutations: mutations, environment: environment}.inputs()
}

// Adamic is the stage 0 compiler, ./cmd/adamic, as one product every test that runs adamic shares.
func Adamic(t testing.TB) string {
	t.Helper()
	return GoBuild(t, "adamic", "./cmd/adamic", nil)
}

// errBuiltElsewhere is BuiltElsewhere's build: the one product test that builds the product is its owner.
var errBuiltElsewhere = errors.New("built by another package's product test")

// AdamicOwner is the one product test that builds Adamic under the trace: every other package's product test that
// reads it waits for it in the tree's product order (cloud/fast-gate/product-order.json), which Loom's build-tree reads,
// so it has finished before they start.
const AdamicOwner = "internal/native TestProduct_WASICompiler"

// AdamicBuiltElsewhere is Adamic for a product test whose package reads the stage 0 compiler that AdamicOwner, another
// package's product test, builds (BuiltElsewhere), and outside a trace Adamic. Four product tests each building Adamic in their own process settled
// it under four keys, and a unit of one package missed the key another's set gave it (chg_35f264p7's void on tree
// 29b74b3d, Oct 14: internal/fuzz, cmd/adamic-test262 and internal/oracle against internal/native's).
func AdamicBuiltElsewhere(t testing.TB) string {
	t.Helper()
	if traceDirectory() == "" {
		return Adamic(t)
	}
	request := goRequest{verb: "build", output: "adamic", pkg: "./cmd/adamic"}
	request.arguments = reproducible(request.arguments)
	inputs, err := request.inputs()
	if err != nil {
		t.Fatalf("go build ./cmd/adamic: %v", err)
	}
	return filepath.Join(BuiltElsewhere(t, "the stage 0 compiler", inputs, AdamicOwner, func(string) error { return errBuiltElsewhere }), request.output)
}

// BuiltElsewhere is a product, what, that another package's product test, owner, builds and settles, for a product
// test whose package reads it: found by its name key and never built here, so the tree holds one settlement under one
// key, and its unit fetches that one. Two packages' product tests that each build one name key settle it twice, and the
// tree's read sets keep only the newest, so the other package's units value to a key only the first ships: oracle's
// and stage3/fixtures' runtime archives voided 3 of oracle's 11 units on tree b89a9632 (#43k99f8). The tree's product
// order starts this test only once owner has finished, so under the trace it looks once, and a product that isn't there
// fails it at once, naming the owner: never a wait. Outside a trace (a developer's run, the tree's untraced fallback) it
// is Product, with build, since nothing else will build it.
func BuiltElsewhere(t testing.TB, what string, inputs Inputs, owner string, build func(directory string) error) string {
	t.Helper()
	if traceDirectory() == "" {
		return Product(t, inputs, build)
	}
	directory, err := Get(inputs, func(string) error { return errBuiltElsewhere })
	if err != nil {
		t.Fatalf("%s is %s's to build, and it isn't here: is this test in the tree's product order after it? %v", what, owner, err)
	}
	return directory
}

// GoExports is the product of 'go list -export -deps -test -json <packages>' at the repository root, for a test that
// type-checks against the compiler's export data (go/importer) without running go: listing.json holds one
// ListedPackage per package go list reported, in its order, each Export a file the product holds and each Dir
// relative to the repository. Its key is GoTest's for the same packages: every file they compile, their tests
// included.
func GoExports(t testing.TB, packages ...string) string {
	t.Helper()
	request := goRequest{verb: "list -export", output: "listing.json", pkg: strings.Join(packages, " ")}
	// -buildvcs=false: a test main is package main, and go stamps the commit into every one it compiles, so without it
	// the export data of each test main differs at every commit under the same key (landable-8: internal/ir's product).
	request.arguments = []string{"-trimpath", "-buildvcs=false"}
	inputs, err := request.inputs()
	if err != nil {
		t.Fatalf("go list -export %s: %v", request.pkg, err)
	}
	root, _ := repositoryRoot()
	return Product(t, inputs, func(directory string) error { return goExports(root, packages, request.output, directory) })
}

func goProduct(t testing.TB, request goRequest) string {
	t.Helper()
	request.arguments = reproducible(request.arguments)
	inputs, err := request.inputs()
	if err != nil {
		t.Fatalf("go %s %s: %v", request.verb, request.pkg, err)
	}
	root, _ := repositoryRoot()
	// Lean: go builds from the files its key names, in a process of its own, by a plan made before the window opens.
	// leanBuildOf, never a lean build made here, so Loom's switch reaches it: a go build settled lean under
	// ADAMIC_BUILD_LEAN=off is a set no lookup matches, and every test reading it, its own included, refuses it (tree
	// 62e5604f, Oct 17: internal/native's Adamic settled lean under the switch, and 64 product tests fell back).
	directory := product(t, inputs, request.plan(root, inputs).Run, leanBuildOf(buildcachePackage))
	return filepath.Join(directory, request.output)
}

// plan is the request's build as a goplan.Plan, decided before its window opens, so the window holds only go and what
// the plan writes. Each read here is one the build's name key holds, so none is lost to the window: the overlay's map
// and its replacements' content (identity), the mutations' before and after strings and the files they mutate, which
// are the packages' own (Declared) or the overlay's replacements, and the outside headers' key over their include
// directories (outsideHeaders' flag). A plan that can't be made says why, and its build fails with it.
func (request goRequest) plan(root string, inputs Inputs) goplan.Plan {
	parts, err := request.planned(root, inputs)
	if err != nil {
		return goplan.Failed(err.Error())
	}
	return goplan.New(parts.command, parts.replace, parts.headers, parts.mutated)
}

// planParts are what goplan.New takes, as planned makes them.
type planParts struct {
	command goplan.Command
	replace map[string]string
	headers []goplan.Header
	mutated []goplan.Mutation
}

func (request goRequest) planned(root string, inputs Inputs) (planParts, error) {
	program, err := toolchain.Path("go")
	if err != nil {
		return planParts{}, err
	}
	_, overlay, err := request.identity(root)
	if err != nil {
		return planParts{}, err
	}
	parts := planParts{command: goplan.Command{Program: program, Directory: filepath.Join(root, request.module), Output: request.output,
		Verb: request.verb, Package: request.pkg, Arguments: append(strings.Fields(request.verb), overlay.arguments...),
		Packages: strings.Fields(request.pkg), Environment: toolchain.Env(request.environment...)}}
	if overlay.path != "" {
		path := overlay.path
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return planParts{}, err
		}
		var declared struct{ Replace map[string]string }
		if err := json.Unmarshal(content, &declared); err != nil {
			return planParts{}, fmt.Errorf("overlay %s: %v", path, err)
		}
		parts.replace = declared.Replace
	}
	if parts.headers, err = outsideHeaderPlan(root, inputs); err != nil {
		return planParts{}, err
	}
	texts, err := request.mutated(root)
	if err != nil {
		return planParts{}, err
	}
	for _, file := range slices.Sorted(maps.Keys(texts)) {
		mutation := goplan.Mutation{File: file, Text: texts[file], Original: filepath.Join(root, file)}
		for from, to := range parts.replace {
			if !filepath.IsAbs(to) {
				to = filepath.Join(root, request.module, to)
			}
			if filepath.Clean(to) == mutation.Original {
				mutation.Pointers = append(mutation.Pointers, from)
			}
		}
		slices.Sort(mutation.Pointers)
		parts.mutated = append(parts.mutated, mutation)
	}
	return parts, nil
}

// mutationKeys are the request's mutations as its identity and key spell them: the file, then before and after
// quoted, so a changed after-string is another product.
func (request goRequest) mutationKeys() []string {
	var keys []string
	for _, mutation := range request.mutations {
		keys = append(keys, "mutation "+mutation.File+" "+strconv.Quote(mutation.Before)+" -> "+strconv.Quote(mutation.After))
	}
	return keys
}

// outsideHeaders begins the flag a key carries for each package that reads a header through '#cgo CFLAGS: -I' outside
// its own directory, followed by an outsideHeaderPackage as JSON. Go's build cache keys a package by its own directory's
// files, so after such a header changes, go build alone links the object compiled from the old one (measured Oct 9 by
// TestProductIdentityAfterAHeaderOutsideThePackageChanges): the key moved, the product didn't. GoBuild built such a
// product with -a, which compiled the standard library again too, 20.8 s of that test on Cloud and the floor of this
// package's unit (#dzpzyme, #hqxje9h). It now lays a header into the package instead (outsideHeaderPlan), which moves
// that package alone in go's cache.
const outsideHeaders = "cgo headers outside the package: "

// An outsideHeaderPackage is a package directory and the include directories outside it its cgo reads, each relative
// to the repository.
type outsideHeaderPackage struct {
	Package  string   `json:"package"`
	Includes []string `json:"includes"`
}

// outsideHeaderName is the header outsideHeaderPlan lays into a package: read by nothing, so the product's bytes are
// a plain build's (measured on Linux, go1.27.1: byte-identical to a fresh build, as -a's were).
const outsideHeaderName = "zz_buildcache_outside_headers.h"

// outsideHeaderPlan is the header the build lays into each package its key names as reading headers outside itself:
// one holding Key over its include directories, so go's cache, which hashes a cgo package's own headers, compiles that
// package again exactly when one of those directories changes.
func outsideHeaderPlan(root string, inputs Inputs) ([]goplan.Header, error) {
	var headers []goplan.Header
	for _, flag := range inputs.Flags {
		encoded, found := strings.CutPrefix(flag, outsideHeaders)
		if !found {
			continue
		}
		var listed outsideHeaderPackage
		if err := json.Unmarshal([]byte(encoded), &listed); err != nil {
			return nil, fmt.Errorf("%s: %v", flag, err)
		}
		var digest string
		var err error
		keying(func() { digest, err = Key(root, Inputs{Name: outsideHeaders + listed.Package, Files: listed.Includes}) })
		if err != nil {
			return nil, err
		}
		headers = append(headers, goplan.Header{
			Target:  filepath.Join(root, filepath.FromSlash(listed.Package), outsideHeaderName),
			Content: []byte("/* buildcache: the headers outside this package are " + digest + " */\n"),
		})
	}
	return headers, nil
}

// goSettings runs go env for names in directory (the working directory when empty) with environment, the whole
// environment go runs with (toolchain.Env for a product, this process's own for Tool), and returns each variable as
// goSetting keys it, in the order asked, with the values go reported. -json among names is ignored: the report is the
// same either way.
func goSettings(directory string, environment, names []string) (settings []string, values map[string]string, err error) {
	keying(func() { settings, values, err = goSettingsKeyed(directory, environment, names) })
	return settings, values, err
}

func goSettingsKeyed(directory string, environment, names []string) ([]string, map[string]string, error) {
	var asked []string
	for _, name := range names {
		if name != "-json" {
			asked = append(asked, name)
		}
	}
	command := toolchain.Command("go", append([]string{"env", "-json", "GOVERSION"}, asked...)...)
	command.Dir = directory
	command.Env = environment
	output, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, nil, fmt.Errorf("go env: %v\n%s", err, exit.Stderr)
		}
		return nil, nil, err
	}
	values := map[string]string{}
	if err := json.Unmarshal(output, &values); err != nil {
		return nil, nil, fmt.Errorf("go env: %v", err)
	}
	settings := make([]string, 0, len(asked))
	for _, name := range asked {
		settings = append(settings, goSetting(name, values[name], values["GOVERSION"]))
	}
	return settings, values, nil
}

// goLocations say where go keeps or fetches things, not what it builds: the release is keyed by GOVERSION, modules by
// go.sum, so two machines with their caches, module downloads and configuration files in different homes build the
// same bytes.
var goLocations = []string{"GOPATH", "GOENV", "GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOBIN", "GOCACHEPROG", "GOPROXY", "GOSUMDB",
	"GOPRIVATE", "GONOPROXY", "GONOSUMDB", "GOINSECURE", "GOAUTH", "GOVCS"}

// goSetting is one go env variable as a key names it (#t37sw0f). GOTOOLCHAIN is a policy (auto on the cloud boxes,
// local elsewhere) and GOROOT and GOTOOLDIR are where a release sits; what changes a product is the release itself, so
// all three name it. The workspace is goWorkspace's. A location is named as one and never valued. Everything else is
// its value, which Key makes portable.
func goSetting(name, value, release string) string {
	switch {
	case name == "GOTOOLCHAIN":
		return "GOTOOLCHAIN selects " + release
	case name == "GOROOT" || name == "GOTOOLDIR":
		return name + " holds " + release
	case name == "GOWORK":
		return goWorkspace(value)
	case slices.Contains(goLocations, name):
		return name + " is a location, not keyed"
	}
	return name + "=" + value
}

// goWorkspace keys the workspace a build reads by what it says, never by which file said it (#t37sw0f): none
// (GOWORK=off and no go.work found are the same build), or the file as go reads it (go work edit -json), with each
// local path in it (a use, a replacement's directory) resolved against the file's own directory and made portable.
// So the tree's own go.work, found from two checkouts, and a go.work outside the tree that uses the same directories
// are one workspace; another go line, toolchain, use or replacement is another.
func goWorkspace(path string) string {
	if path == "" || path == "off" {
		return "GOWORK=off"
	}
	output, err := toolchain.Command("go", "work", "edit", "-json", path).Output()
	if err != nil {
		return "GOWORK unreadable: " + err.Error()
	}
	var workspace struct {
		Go, Toolchain string
		Godebug       []struct{ Key, Value string }
		Use           []struct{ DiskPath, ModulePath string }
		Replace       []struct {
			Old, New struct{ Path, Version string }
		}
	}
	if err := json.Unmarshal(output, &workspace); err != nil {
		return "GOWORK unreadable: " + err.Error()
	}
	root, _ := repositoryRoot()
	local := func(name string) string {
		if !filepath.IsAbs(name) {
			name = filepath.Join(filepath.Dir(path), name)
		}
		return portable(root, filepath.Clean(name))
	}
	for index := range workspace.Use {
		workspace.Use[index].DiskPath = local(workspace.Use[index].DiskPath)
	}
	for index := range workspace.Replace {
		if workspace.Replace[index].New.Version == "" {
			workspace.Replace[index].New.Path = local(workspace.Replace[index].New.Path)
		}
	}
	canonical, err := json.Marshal(workspace)
	if err != nil {
		return "GOWORK unreadable: " + err.Error()
	}
	return "GOWORK " + string(canonical)
}

// GoInputs is GoBuild's key: what go list says the build compiles, with everything that can change its output.
func GoInputs(output, pkg string, arguments []string, environment []string) (Inputs, error) {
	return goRequest{verb: "build", output: output, pkg: pkg, arguments: arguments, environment: environment}.inputs()
}

// GoTestInputs is GoTestMutated's key (GoTest's, with no mutations), computed without building: for a product made by
// running the test binary, which can be keyed by the binary's key rather than its bytes, so a reader keys it without
// holding the binary (#4ycgm2n: tsprinter's mutant oracle corpora). The arguments are made reproducible as the
// product's own are, so the key is the product's.
func GoTestInputs(module, output, pkg string, arguments []string, mutations []Mutation, environment ...string) (Inputs, error) {
	return goRequest{verb: "test -c", module: module, output: output, pkg: pkg, arguments: reproducible(arguments), mutations: mutations, environment: environment}.inputs()
}

// inputs is the request's key: from the planner's index when it holds the request (precomputed), else as this process
// already computed it, else computed here under keying and remembered for the rest of the process, so a unit asking
// for one product many times keys it once, its parallel tests included: they wait on the first one's go list rather
// than each running their own. A failure isn't remembered, so the next ask tries again.
func (request goRequest) inputs() (Inputs, error) {
	root, err := repositoryRoot()
	if err != nil {
		return Inputs{}, err
	}
	identity, overlay, err := request.identity(root)
	if err != nil {
		return Inputs{}, err
	}
	if inputs, ok := precomputed()[identity]; ok {
		return inputs, nil
	}
	entry, _ := computedInputs.LoadOrStore(identity, &keyedRequest{})
	keyed := entry.(*keyedRequest)
	keyed.once.Do(func() {
		keying(func() { keyed.inputs, keyed.err = request.keyInputs(root, overlay) })
		if keyed.err != nil {
			computedInputs.CompareAndDelete(identity, keyed)
			return
		}
		recordInputs(identity, keyed.inputs)
	})
	return keyed.inputs, keyed.err
}

// computedInputs holds each request this process keyed, by identity, as a keyedRequest.
var computedInputs sync.Map

// A keyedRequest is one request's key, computed once however many tests ask for it at the same time.
type keyedRequest struct {
	once   sync.Once
	inputs Inputs
	err    error
}

// identity spells the request the same from every checkout of the tree: the verb, module, package and output, each
// argument, an -overlay as the map it declares (overlayKey's flags, never the overlay file's temporary path), and each
// environment value, made portable. It names the request in the planner's index and in this process's memo.
func (request goRequest) identity(root string) (string, overlayKey, error) {
	if request.module != "" && (filepath.IsAbs(request.module) || !inside(root, filepath.Join(root, request.module))) {
		return "", overlayKey{}, fmt.Errorf("module %s isn't a directory of the repository", request.module)
	}
	overlayPath, keyed := "", []string{}
	for index := 0; index < len(request.arguments); index++ {
		switch argument := request.arguments[index]; {
		case argument == "-overlay" && index+1 < len(request.arguments):
			overlayPath = request.arguments[index+1]
			index++
		case strings.HasPrefix(argument, "-overlay="):
			overlayPath = strings.TrimPrefix(argument, "-overlay=")
		case argument == "-overlay":
			return "", overlayKey{}, errors.New("-overlay names no file")
		default:
			keyed = append(keyed, argument)
		}
	}
	overlay, err := overlayInputs(root, overlayPath)
	if err != nil {
		return "", overlayKey{}, err
	}
	overlay.path, overlay.arguments = overlayPath, keyed
	parts := []string{"go " + request.verb, "module " + request.module, "package " + portable(root, request.pkg), "output " + request.output}
	for _, argument := range keyed {
		parts = append(parts, "argument "+portable(root, argument))
	}
	parts = append(parts, overlay.flags...)
	parts = append(parts, request.mutationKeys()...)
	for _, setting := range request.environment {
		parts = append(parts, "environment "+portable(root, setting))
	}
	return strings.Join(parts, "\n"), overlay, nil
}

// keyInputs computes the request's key with go list and go env, run where the build runs.
func (request goRequest) keyInputs(root string, overlay overlayKey) (Inputs, error) {
	directory := filepath.Join(root, request.module)
	run := func(arguments ...string) ([]byte, error) {
		command := toolchain.Command("go", arguments...)
		command.Dir = directory
		command.Env = toolchain.Env(request.environment...)
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
	listed := listArguments(request.arguments)
	if overlay.path != "" {
		listed = append(listed, "-overlay="+overlay.path)
	}
	asked := []string{"list", "-deps", "-json"}
	if request.verb == "test -c" || request.verb == "list -export" {
		// The test variant of each package, so its _test.go files are keyed too.
		asked = append(asked, "-test")
	}
	listing, err := run(append(append(asked, listed...), strings.Fields(request.pkg)...)...)
	if err != nil {
		return Inputs{}, err
	}
	// go reports paths as the system resolves them, and the repository's root may be spelled through a symbolic link
	// (macOS's /var is /private/var): a path under either spelling is the repository's, respelled under root, so no
	// file of the tree drops out of the key unseen.
	real := resolved(root)
	respell := func(path string) string {
		if real != "" && real != root && inside(real, path) {
			return filepath.Join(root, mustRelative(real, path))
		}
		return path
	}
	files := map[string]bool{}
	// Each package that reads headers outside its own directory, with those include directories (outsideHeaders).
	outside := map[string][]string{}
	add := func(directory string, names []string) error {
		for _, name := range names {
			// go list -test names some files absolutely, in go's own cache (a test variant's generated sources): not the
			// repository's, so not keyed.
			path := name
			if !filepath.IsAbs(path) {
				path = filepath.Join(directory, name)
			}
			path = respell(path)
			if !inside(root, path) {
				continue
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			// A path the overlay replaces or adds is keyed by the overlay's map and the replacement's content; one the
			// overlay adds isn't on disk to be hashed at all.
			if !slices.Contains(overlay.originals, relative) {
				files[relative] = true
			}
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
		listed.Dir = respell(listed.Dir)
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
				include = respell(include)
				if inside(root, include) {
					if err := add(root, []string{mustRelative(root, include)}); err != nil {
						return Inputs{}, err
					}
					if filepath.Clean(include) != filepath.Clean(listed.Dir) {
						pkg := filepath.ToSlash(mustRelative(root, listed.Dir))
						if relative := filepath.ToSlash(mustRelative(root, include)); !slices.Contains(outside[pkg], relative) {
							outside[pkg] = append(outside[pkg], relative)
						}
					}
				}
			}
		}
		if listed.Module != nil && inside(root, respell(listed.Module.GoMod)) {
			module := filepath.Dir(respell(listed.Module.GoMod))
			if err := add(module, existing(module, "go.mod", "go.sum")); err != nil {
				return Inputs{}, err
			}
		}
	}
	for _, place := range []string{root, directory} {
		if err := add(place, existing(place, "go.work")); err != nil {
			return Inputs{}, err
		}
	}
	// A module the build's module or the workspace replaces with a directory of the tree selects versions through its
	// go.mod whether or not the build compiles a package of it (the 17 cohere/TypeScript-shim modules, #c2chyab).
	replaced, err := localReplacements(root, root, directory)
	if err != nil {
		return Inputs{}, err
	}
	for _, module := range replaced {
		if err := add(module, existing(module, "go.mod", "go.sum")); err != nil {
			return Inputs{}, err
		}
	}
	// The go env values a build reads beyond its sources (toolchain.GoSettings, each of which toolchain.Env passes), asked
	// of go under the environment the build gets, so a default counts like a setting, keyed as goSetting says: GOFLAGS
	// without its -p share (portable), the workspace by content, the release GOTOOLCHAIN selects.
	settings, values, err := goSettings(directory, toolchain.Env(request.environment...), toolchain.GoSettings)
	if err != nil {
		return Inputs{}, err
	}
	for _, target := range overlay.files {
		files[target] = true
	}
	// A mutation plants into a file the build reads, as a package's source or an overlay's replacement, with a Before
	// that occurs once: anything else would build the unmutated product under a mutant's key.
	for _, mutation := range request.mutations {
		if !files[mutation.File] {
			return Inputs{}, fmt.Errorf("mutation of %s: the build doesn't read it (a package's source the overlay doesn't replace, or an overlay's replacement)", mutation.File)
		}
	}
	if _, err := request.mutated(root); err != nil {
		return Inputs{}, err
	}
	name := "go build " + request.pkg + " " + request.output
	if request.verb != "build" || request.module != "" {
		name = "go " + request.verb + " " + request.module + " " + request.pkg + " " + request.output
	}
	inputs := Inputs{Name: name, Flags: append(append([]string{"arguments " + strings.Join(overlay.arguments, " ")}, overlay.flags...), request.mutationKeys()...), Toolchain: []string{Tool("go", "version")}}
	for _, pkg := range slices.Sorted(maps.Keys(outside)) {
		includes := slices.Sorted(slices.Values(outside[pkg]))
		encoded, err := json.Marshal(outsideHeaderPackage{Package: pkg, Includes: includes})
		if err != nil {
			return Inputs{}, err
		}
		inputs.Flags = append(inputs.Flags, outsideHeaders+string(encoded))
	}
	inputs.Flags = append(inputs.Flags, settings...)
	if compiler := strings.Fields(values["CC"]); len(compiler) > 0 {
		inputs.Toolchain = append(inputs.Toolchain, Tool(compiler[0], "--version"))
	}
	for name := range files {
		inputs.Files = append(inputs.Files, name)
	}
	sort.Strings(inputs.Files)
	return inputs, nil
}

// The planner's index of go requests' keys (#pc0jvv4): ADAMIC_BUILD_INPUTS names a file of JSON lines, each a request's
// identity and its Inputs, so a unit reading products runs no go to key them. A tree build writes those lines as it
// keys (ADAMIC_BUILD_INPUTS_RECORD, appended one line per request), and a planner hands the file on. A request the
// index doesn't hold is keyed here, as before.
type indexedInputs struct {
	Request string `json:"request"`
	Inputs  Inputs `json:"inputs"`
}

// precomputed is the index ADAMIC_BUILD_INPUTS names, read once per file.
func precomputed() map[string]Inputs {
	path := os.Getenv("ADAMIC_BUILD_INPUTS")
	if path == "" {
		return nil
	}
	if index, ok := indexes.Load(path); ok {
		return index.(map[string]Inputs)
	}
	index := map[string]Inputs{}
	content, err := os.ReadFile(path)
	if err != nil {
		note("build inputs index %s: %v", path, err)
		return index
	}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	for {
		var line indexedInputs
		if err := decoder.Decode(&line); err == io.EOF {
			break
		} else if err != nil {
			note("build inputs index %s: %v", path, err)
			return map[string]Inputs{}
		}
		index[line.Request] = line.Inputs
	}
	indexes.Store(path, index)
	return index
}

var indexes sync.Map

func recordInputs(identity string, inputs Inputs) {
	path := os.Getenv("ADAMIC_BUILD_INPUTS_RECORD")
	if path == "" {
		return
	}
	line, err := json.Marshal(indexedInputs{Request: identity, Inputs: inputs})
	if err != nil {
		return
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		note("build inputs record %s: %v", path, err)
		return
	}
	file.Write(append(line, '\n'))
	file.Close()
}

// overlayKey is what an -overlay adds to a key: one flag per replaced path (the map, as go reads it, made relative to
// the repository) and the replacement files, whose content the key hashes like any other input file.
type overlayKey struct {
	flags, files []string
	// originals are the paths the overlay replaces or adds, relative to the repository; path is the overlay file the
	// request names and arguments the request's other arguments.
	originals, arguments []string
	path                 string
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
		key.originals = append(key.originals, relative(original))
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

func existing(directory string, names ...string) []string {
	var found []string
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(directory, name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}
