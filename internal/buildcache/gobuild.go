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
	"sync"
	"testing"
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

// Adamic is the stage 0 compiler, ./cmd/adamic, as one product every test that runs adamic shares.
func Adamic(t testing.TB) string {
	t.Helper()
	return GoBuild(t, "adamic", "./cmd/adamic", nil)
}

// A goRequest is one go invocation a product is built by: 'go build', or 'go test -c', of pkg in module (a directory
// of the repository, its root when empty), with arguments and environment added.
type goRequest struct {
	verb, module, output, pkg string
	arguments, environment    []string
}

func goProduct(t testing.TB, request goRequest) string {
	t.Helper()
	request.arguments = reproducible(request.arguments)
	inputs, err := request.inputs()
	if err != nil {
		t.Fatalf("go %s %s: %v", request.verb, request.pkg, err)
	}
	root, _ := repositoryRoot()
	arguments := request.arguments
	if slices.Contains(inputs.Flags, rebuildEverything) {
		arguments = append([]string{"-a"}, arguments...)
	}
	directory := Product(t, inputs, func(directory string) error {
		command := exec.Command("go", append(append(append(strings.Fields(request.verb), arguments...), "-o", filepath.Join(directory, request.output)), request.pkg)...)
		command.Dir = filepath.Join(root, request.module)
		command.Env = append(os.Environ(), request.environment...)
		if combined, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("go %s %s: %v\n%s", request.verb, request.pkg, err, combined)
		}
		return nil
	})
	return filepath.Join(directory, request.output)
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
			panic("buildcache.GoBuild and GoTest set -trimpath, -ldflags and -buildvcs themselves, for a product that's the same from any checkout path and commit")
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

// The go env values a build reads beyond its sources, resolved by go itself so a default counts like a setting, and
// keyed as goSetting says: GOFLAGS without its -p share (portable), the workspace by content, the release
// GOTOOLCHAIN selects.
var goEnvironment = []string{"GOOS", "GOARCH", "GOAMD64", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "GOWORK", "CGO_ENABLED", "CC", "CXX",
	"CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOTOOLCHAIN"}

// goSettings runs go env for names in directory (the working directory when empty) with environment added, and
// returns each variable as goSetting keys it, in the order asked, with the values go reported. -json among names is
// ignored: the report is the same either way.
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
	command := exec.Command("go", append([]string{"env", "-json", "GOVERSION"}, asked...)...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
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
	output, err := exec.Command("go", "work", "edit", "-json", path).Output()
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

// inputs is the request's key: from the planner's index when it holds the request (precomputed), else as this process
// already computed it, else computed here under keying and remembered for the rest of the process, so a unit asking
// for one product many times keys it once.
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
	if known, ok := computedInputs.Load(identity); ok {
		return known.(Inputs), nil
	}
	var inputs Inputs
	keying(func() { inputs, err = request.keyInputs(root, overlay) })
	if err != nil {
		return Inputs{}, err
	}
	computedInputs.Store(identity, inputs)
	recordInputs(identity, inputs)
	return inputs, nil
}

// computedInputs holds each request's inputs this process computed, by identity.
var computedInputs sync.Map

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
	for _, setting := range request.environment {
		parts = append(parts, "environment "+portable(root, setting))
	}
	return strings.Join(parts, "\n"), overlay, nil
}

// keyInputs computes the request's key with go list and go env, run where the build runs.
func (request goRequest) keyInputs(root string, overlay overlayKey) (Inputs, error) {
	directory := filepath.Join(root, request.module)
	run := func(arguments ...string) ([]byte, error) {
		command := exec.Command("go", arguments...)
		command.Dir = directory
		command.Env = append(os.Environ(), request.environment...)
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
	if request.verb == "test -c" {
		// The test variant of each package, so its _test.go files are keyed too.
		asked = append(asked, "-test")
	}
	listing, err := run(append(append(asked, listed...), request.pkg)...)
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
	outsideHeaders := false
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
						outsideHeaders = true
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
	settings, values, err := goSettings(directory, request.environment, goEnvironment)
	if err != nil {
		return Inputs{}, err
	}
	for _, target := range overlay.files {
		files[target] = true
	}
	name := "go build " + request.pkg + " " + request.output
	if request.verb != "build" {
		name = "go " + request.verb + " " + request.module + " " + request.pkg + " " + request.output
	}
	inputs := Inputs{Name: name, Flags: append([]string{"arguments " + strings.Join(overlay.arguments, " ")}, overlay.flags...), Toolchain: []string{Tool("go", "version")}}
	if outsideHeaders {
		inputs.Flags = append(inputs.Flags, rebuildEverything)
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
