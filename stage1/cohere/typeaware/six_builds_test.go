package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The oracle recipe records all inputs for internal/buildcache.
type sixBuildInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
}

type sixProduct struct {
	Inputs sixBuildInputs
	Build  func(dir string) error
	File   string
}

// Each product has one callback. Its output and compiler logs stay in dir;
// stage-0 lowering, flags, archive linking and sanitizers are unchanged.
func sixCommandBuild(command *exec.Cmd, file string) func(string) error {
	return func(dir string) error {
		args := append([]string(nil), command.Args[1:]...)
		for i, arg := range args {
			if arg == "-o" {
				args[i+1] = filepath.Join(dir, file)
			}
		}
		cmd := exec.Command(command.Path, args...)
		cmd.Dir = command.Dir
		cmd.Env = command.Env
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := typeAwareRunBuildCommand(cmd)
		if writeErr := os.WriteFile(filepath.Join(dir, "build.stdout"), stdout.Bytes(), 0644); writeErr != nil {
			return writeErr
		}
		if writeErr := os.WriteFile(filepath.Join(dir, "build.stderr"), stderr.Bytes(), 0644); writeErr != nil {
			return writeErr
		}
		if err != nil {
			return fmt.Errorf("%w\n%s\n%s", err, stdout.Bytes(), stderr.Bytes())
		}
		return nil
	}
}

// Input collection is conservative across the three Go products and native
// ports, and includes actual Go embed/C inputs, local modules and TS sources.
// Runtime/tool executables are described by versions; Files remain repo relative.
func (h *harness) sixSourceInputs() []string {
	h.t.Helper()
	if h.sixFiles != nil {
		return h.sixFiles
	}
	seen := map[string]bool{}
	add := func(path string) {
		relative, err := filepath.Rel(h.repository, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			seen[filepath.ToSlash(relative)] = true
		}
	}
	for _, request := range []struct {
		dir      string
		packages []string
	}{
		{h.repository, []string{"./cmd/adamic", "./bridge/tsgo/archive", "./bridge/tsgo/cost", "./bridge/tsgo/checker"}},
		{filepath.Join(h.repository, "cohere"), []string{"./internal/lint/rules/typescript"}},
	} {
		cmd := toolchain.CommandIn(request.dir, "go", append([]string{"list", "-test", "-deps", "-json"}, request.packages...)...)
		data, err := typeAwareCommandOutput(cmd)
		if err != nil {
			h.t.Fatalf("six build input inventory: %v", err)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		for {
			var pkg struct {
				Dir                                                                                                                         string
				GoFiles, TestGoFiles, XTestGoFiles, CgoFiles, CFiles, CXXFiles, MFiles, HFiles, SFiles, SwigFiles, SwigCXXFiles, EmbedFiles []string
				Module                                                                                                                      *struct{ GoMod string }
			}
			err := decoder.Decode(&pkg)
			if err == io.EOF {
				break
			}
			if err != nil {
				h.t.Fatal(err)
			}
			for _, list := range [][]string{pkg.GoFiles, pkg.TestGoFiles, pkg.XTestGoFiles, pkg.CgoFiles, pkg.CFiles, pkg.CXXFiles, pkg.MFiles, pkg.HFiles, pkg.SFiles, pkg.SwigFiles, pkg.SwigCXXFiles, pkg.EmbedFiles} {
				for _, file := range list {
					path := file
					if !filepath.IsAbs(path) {
						path = filepath.Join(pkg.Dir, path)
					}
					// go list -test's absolute generated main lives in GOCACHE.
					// Its contents derive from the declared tests and Go version.
					add(path)
				}
			}
			if pkg.Module != nil && pkg.Module.GoMod != "" {
				add(pkg.Module.GoMod)
				sum := filepath.Join(filepath.Dir(pkg.Module.GoMod), "go.sum")
				if _, err := os.Stat(sum); err == nil {
					add(sum)
				}
			}
		}
	}

	for file := range seen {
		h.sixFiles = append(h.sixFiles, file)
	}
	sort.Strings(h.sixFiles)
	return h.sixFiles
}

// Values are used as build inputs, never printed in timing logs. What only says where something is (PATH, HOME, the
// temporary directory, go's caches, module proxies, GOROOT and GOPATH, GOENV's file), a parallel share
// (ADAMIC_NATIVE_JOBS), a cache switch (ADAMIC_GATE_UNCACHED) or a release policy (GOTOOLCHAIN, whose release
// runtime.Version and go version name) isn't one: it differs between Workshop and a runner and changes no product
// (#t37sw0f). GOWORK is the workspace the command's own go env reports (sixBuildRecipe), named inside the tree;
// WASI_SYSROOT serves wasm builds, which these aren't.
var sixBuildEnvironment = []string{
	"GODEBUG", "GOFIPS140", "GO111MODULE", "GOFLAGS",
	"GOOS", "GOARCH", "GOAMD64", "GOARM", "GOARM64", "GOMIPS", "GOMIPS64", "GOWASM", "GOEXPERIMENT",
	"CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "PKG_CONFIG",
	"CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "LIBRARY_PATH", "COMPILER_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET",
	"ADAMIC_NATIVE_SPLIT",
}

func sixEnvironment(command *exec.Cmd, name string) string {
	environment := command.Env
	if environment == nil {
		environment = os.Environ()
	}
	for i := len(environment) - 1; i >= 0; i-- {
		if value, ok := strings.CutPrefix(environment[i], name+"="); ok {
			return value
		}
	}
	return "<unset>"
}

var sixImport = regexp.MustCompile(`(?:from\s*|import\s*)['"]([^'"]+)['"]`)

// The generated mutant drivers/overlays are not repo files. Their complete
// contents enter Flags, with staging paths normalized, including transitive
// generated imports. Linked product inputs are included as nested recipes.
func (h *harness) sixBuildRecipe(name string, command *exec.Cmd) sixProduct {
	h.t.Helper()
	if command.Dir == "" {
		command.Dir = h.repository
	}
	if command.Env != nil {
		// exec.Cmd sets PWD only for a command with no Env of its own; toolchain's have one.
		command.Env = append(command.Env, "PWD="+command.Dir)
	}
	unstamped(command)
	canonical := func(text string) string {
		text = strings.ReplaceAll(text, h.directory, "$WORK")
		return strings.ReplaceAll(text, h.repository, "$REPO")
	}
	in := sixBuildInputs{Name: "typeaware six " + name, Files: append([]string(nil), h.sixSourceInputs()...), Flags: []string{"six-build-recipe-v1"}}
	if h.sixVersions == nil {
		h.sixVersions = []string{"runtime.Version()=" + runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}
	}
	in.Toolchain = append([]string(nil), h.sixVersions...)
	// Include effective Go configuration, not only exported values: GOENV may
	// provide flags and compiler settings. The callback has the same environment.
	env := toolchain.Command("go", "env", "-json", "GOOS", "GOARCH", "GOAMD64", "GOARM", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOWORK", "GOMOD")
	env.Dir = command.Dir
	env.Env = command.Env
	data, err := typeAwareCommandOutput(env)
	if err != nil {
		h.t.Fatal(err)
	}
	in.Flags = append(in.Flags, "effective go env="+canonical(string(data)))
	var settings map[string]string
	if err := json.Unmarshal(data, &settings); err != nil {
		h.t.Fatal(err)
	}
	// CC's version also matters for cgo, including its default when CC is unset.
	cc := strings.Fields(settings["CC"])
	if len(cc) > 0 {
		in.Toolchain = append(in.Toolchain, buildcache.Tool(cc[0], append(cc[1:], "--version")...))
	}
	for _, key := range sixBuildEnvironment {
		in.Flags = append(in.Flags, "env "+key+"="+sixEnvironment(command, key))
	}
	in.Flags = append(in.Flags, "cwd="+canonical(command.Dir))
	// Go stamps a build with its commit and whether the checkout is modified unless told not to, so every go build here
	// passes -buildvcs=false itself (unstamped), whatever GOFLAGS says: a stamped product's bytes and key named the commit
	// and vcs.modified, so a second checkout, or the same one with a stray file, missed it and every product reading it
	// (#t37sw0f, typeaware six type-symbol-checker and typeaware-volume-type-symbol-native). Nothing reads the stamp.
	in.Flags = append(in.Flags, "vcs: not stamped, -buildvcs=false")

	file := ""
	seen := map[string]bool{}
	var input func(string)
	input = func(path string) {
		if seen[path] {
			return
		}
		seen[path] = true
		if dependency, ok := h.sixProducts[path]; ok {
			data, err := json.Marshal(dependency)
			if err != nil {
				h.t.Fatal(err)
			}
			in.Flags = append(in.Flags, "linked product="+string(data))
			return
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(command.Dir, path)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return
		}
		if strings.HasPrefix(path, h.directory+string(filepath.Separator)) {
			in.Flags = append(in.Flags, "generated "+canonical(path)+"="+canonical(string(contents)))
		} else if relative, err := filepath.Rel(h.repository, path); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			in.Files = append(in.Files, filepath.ToSlash(relative))
		} else {
			h.t.Fatalf("undeclared six build input outside repository: %s", path)
		}

		switch filepath.Ext(path) {
		case ".json":
			var overlay struct{ Replace map[string]string }
			if json.Unmarshal(contents, &overlay) == nil {
				var keys []string
				for key := range overlay.Replace {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					input(overlay.Replace[key])
				}
			}
		case ".ts", ".a":
			for _, match := range sixImport.FindAllStringSubmatch(string(contents), -1) {
				target := match[1]
				if strings.HasPrefix(target, ".") {
					target = filepath.Join(filepath.Dir(path), target)
				}
				if filepath.IsAbs(target) {
					input(target)
				}
			}
		}
	}
	if _, ok := h.sixProducts[command.Path]; ok {
		input(command.Path)
	}
	for i, arg := range command.Args[1:] {
		if i > 0 && command.Args[i] == "-o" {
			file = filepath.Base(arg)
			in.Flags = append(in.Flags, "output="+file)
			continue
		}
		if dependency, ok := h.sixProducts[arg]; ok {
			in.Flags = append(in.Flags, "arg=product "+dependency.Name)
		} else {
			in.Flags = append(in.Flags, "arg="+canonical(arg))
		}
		if arg != "-o" {
			input(arg)
		}
	}
	if file == "" {
		h.t.Fatalf("six build %s has no output", name)
	}
	sort.Strings(in.Files)
	in.Files = compactSixFiles(in.Files)
	return sixProduct{Inputs: in, Build: sixCommandBuild(command, file), File: file}
}

// unstamped makes a go build or go test -c command pass -buildvcs=false, right after its verb, unless it already says
// how to stamp; any other command is left as it is.
func unstamped(command *exec.Cmd) {
	if len(command.Args) < 2 || filepath.Base(command.Args[0]) != "go" || command.Args[1] != "build" && command.Args[1] != "test" {
		return
	}
	if slices.ContainsFunc(command.Args[2:], func(argument string) bool { return strings.HasPrefix(argument, "-buildvcs") }) {
		return
	}
	command.Args = slices.Insert(command.Args, 2, "-buildvcs=false")
}

func compactSixFiles(files []string) []string {
	var result []string
	for _, file := range files {
		if len(result) == 0 || result[len(result)-1] != file {
			result = append(result, file)
		}
	}
	return result
}

func (h *harness) sixBuildProduct(name string, command *exec.Cmd) string {
	h.t.Helper()
	product := h.sixBuildRecipe(name, command)
	dir := h.sixBuildOnce(product)
	path := filepath.Join(dir, product.File)
	if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
		h.t.Fatalf("missing product %s: %v", path, err)
	}
	if h.sixProducts == nil {
		h.sixProducts = map[string]sixBuildInputs{}
	}
	h.sixProducts[path] = product.Inputs
	return path
}

func (h *harness) sixBuildOnce(product sixProduct) string {
	h.t.Helper()
	return buildcache.Product(h.t, buildcache.Inputs(product.Inputs), product.Build)
}

// Every go build or go test -c a six product runs is unstamped, so neither its bytes nor its key name the commit or a
// stray file (#t37sw0f); a command that already says how to stamp, and any other tool, is left as it is. Mutants: no
// -buildvcs=false added; one added to a command that already chose.
func TestSixGoBuildsAreUnstamped(t *testing.T) {
	t.Parallel()
	for _, command := range []struct{ arguments, want []string }{
		{[]string{"go", "build", "-o", "x", "./p"}, []string{"go", "build", "-buildvcs=false", "-o", "x", "./p"}},
		{[]string{"go", "test", "-c", "-o", "x", "./p"}, []string{"go", "test", "-buildvcs=false", "-c", "-o", "x", "./p"}},
		{[]string{"go", "build", "-buildvcs=true", "./p"}, []string{"go", "build", "-buildvcs=true", "./p"}},
		{[]string{"go", "build", "-buildvcs=false", "./p"}, []string{"go", "build", "-buildvcs=false", "./p"}},
		{[]string{"go", "env", "GOFLAGS"}, []string{"go", "env", "GOFLAGS"}},
		{[]string{"node", "build", "x.js"}, []string{"node", "build", "x.js"}},
	} {
		cmd := exec.Command(command.arguments[0], command.arguments[1:]...)
		unstamped(cmd)
		if !slices.Equal(cmd.Args, command.want) {
			t.Errorf("%q became %q, want %q", command.arguments, cmd.Args, command.want)
		}
	}
}

// Exercise the callback boundary with a real child process: the requested
// output must be redirected into the product directory, not the caller's path.
// Not parallel: parent and child share the requested build product path.
func TestSixBuildCallbackUsesProductDirectory(t *testing.T) {
	if os.Getenv("ADAMIC_SIX_PRODUCT_CHILD") == "1" {
		for i, arg := range os.Args {
			if arg == "-o" && i+1 < len(os.Args) {
				if err := os.WriteFile(os.Args[i+1], []byte("product"), 0755); err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		t.Fatal("missing product output")
	}
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "product")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "caller-output")
	cmd := exec.Command(os.Args[0], "-test.run=^TestSixBuildCallbackUsesProductDirectory$", "--", "-o", original)
	cmd.Env = append(os.Environ(), "ADAMIC_SIX_PRODUCT_CHILD=1")
	arguments := strings.Join(cmd.Args, "\x00")
	if err := sixCommandBuild(cmd, "port")(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(original); !os.IsNotExist(err) {
		t.Fatalf("callback wrote outside product directory: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "port"))
	if err != nil || string(data) != "product" {
		t.Fatalf("missing product: %v %q", err, data)
	}
	if strings.Join(cmd.Args, "\x00") != arguments {
		t.Fatal("callback changed the declared command inputs")
	}
}
