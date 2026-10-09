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
	"sort"
	"strings"
	"testing"
	"time"
)

// Same shape as the forthcoming internal/buildcache.Inputs. This is an input
// description, not a cache. Replace this type with the shared type when it lands.
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
		err := typeAwareRunCommand(cmd)
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
		cmd := exec.Command("go", append([]string{"list", "-test", "-deps", "-json"}, request.packages...)...)
		cmd.Dir = request.dir
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

// Values are used as build inputs, never printed in timing logs. In particular,
// proxy settings can contain credentials and must not be logged as a manifest.
var sixBuildEnvironment = []string{
	"PATH", "HOME", "USERPROFILE", "XDG_CACHE_HOME", "TMPDIR", "TMP", "TEMP",
	"GODEBUG", "GOFIPS140", "GOAUTH", "GOENV", "GOWORK", "GO111MODULE", "GOFLAGS", "GOTOOLCHAIN", "GOROOT", "GOPATH",
	"GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOPROXY", "GOSUMDB", "GOPRIVATE", "GONOPROXY", "GONOSUMDB",
	"GOOS", "GOARCH", "GOAMD64", "GOARM", "GOARM64", "GOMIPS", "GOMIPS64", "GOWASM", "GOEXPERIMENT",
	"CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "PKG_CONFIG",
	"CPATH", "C_INCLUDE_PATH", "CPLUS_INCLUDE_PATH", "LIBRARY_PATH", "COMPILER_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET",
	"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED", "WASI_SYSROOT",
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
	canonical := func(text string) string {
		text = strings.ReplaceAll(text, h.directory, "$WORK")
		return strings.ReplaceAll(text, h.repository, "$REPO")
	}
	in := sixBuildInputs{Name: "typeaware six " + name, Files: append([]string(nil), h.sixSourceInputs()...), Flags: []string{"six-build-recipe-v1"}}
	if h.sixVersions == nil {
		h.sixVersions = []string{"runtime.Version()=" + runtime.Version()}
		for _, tool := range []struct {
			name string
			args []string
		}{{"go", []string{"version"}}, {"clang", []string{"--version"}}} {
			data, err := typeAwareCommandOutput(exec.Command(tool.name, tool.args...))
			if err != nil {
				h.t.Fatal(err)
			}
			h.sixVersions = append(h.sixVersions, string(data))
		}
	}
	in.Toolchain = append([]string(nil), h.sixVersions...)
	// Include effective Go configuration, not only exported values: GOENV may
	// provide flags and compiler settings. The callback has the same environment.
	env := exec.Command("go", "env", "-json", "GOOS", "GOARCH", "GOAMD64", "GOARM", "GOARM64", "GOEXPERIMENT", "GOFLAGS", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOWORK", "GOMOD")
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
		data, err := typeAwareCommandOutput(exec.Command(cc[0], append(cc[1:], "--version")...))
		if err != nil {
			h.t.Fatal(err)
		}
		in.Toolchain = append(in.Toolchain, string(data))
	}
	for _, key := range sixBuildEnvironment {
		in.Flags = append(in.Flags, "env "+key+"="+sixEnvironment(command, key))
	}
	in.Flags = append(in.Flags, "cwd="+canonical(command.Dir))
	// Go embeds VCS metadata by default. Preserve the existing flags and declare
	// this implicit input too, rather than silently changing -buildvcs behavior.
	revision, err := typeAwareCommandOutput(exec.Command("git", "-C", command.Dir, "rev-parse", "HEAD"))
	if err != nil {
		h.t.Fatal(err)
	}
	status, err := typeAwareCommandOutput(exec.Command("git", "-C", command.Dir, "status", "--porcelain"))
	if err != nil {
		h.t.Fatal(err)
	}
	in.Flags = append(in.Flags, "vcs.revision="+strings.TrimSpace(string(revision)), fmt.Sprintf("vcs.modified=%t", len(status) != 0))

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
	// Once buildcache lands, replace this call with:
	// dir := buildcache.Product(h.t, buildcache.Inputs(product.Inputs), product.Build)
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
	inputs, err := json.Marshal(product.Inputs)
	if err != nil {
		h.t.Fatal(err)
	}
	dir, err := sharedProduct("six "+string(inputs), func(directory string) (string, error) {
		started := time.Now()
		err := product.Build(directory)
		h.t.Logf("six-build unit=%s function=sixCommandBuild elapsed_s=%.6f", strings.TrimPrefix(product.Inputs.Name, "typeaware six "), time.Since(started).Seconds())
		return directory, err
	})
	if err != nil {
		h.t.Fatalf("%s build: %v", product.Inputs.Name, err)
	}
	return dir
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
