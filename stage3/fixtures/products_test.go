package fixtures

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Discovery uses Go's effective dependency lists, including test sources and
// embedded runtime files. Repository inputs are Files; external dependencies,
// compiler binaries and their content hashes identify the toolchain. Each is
// labelled by what it is (a module file by its module-cache path, a tool by its
// role), never by where this machine keeps it (#t37sw0f); go's own locations,
// the GOTOOLCHAIN policy and the python and bootstrap go running this script
// aren't inputs: go version names the release that builds.
const fixtureHookInputsScript = `
import hashlib, json, os, pathlib, runpy, shlex, shutil, subprocess, sys
root = pathlib.Path(sys.argv[1]).resolve()
hook = runpy.run_path(str(root / 'stage3/fixtures/build-hook.py'))
files = {'go.mod', 'stage3/fixtures/build-hook.py', 'stage3/fixtures/products_test.go'}
if (root / 'go.sum').exists(): files.add('go.sum')
external = {}
for package in hook['records'](subprocess.check_output(['go','list','-deps','-test','-json','./internal/oracle'], cwd=root, text=True)):
    if package.get('Error') or package.get('DepsErrors'):
        raise RuntimeError('oracle dependency discovery failed')
    directory = pathlib.Path(package['Dir'])
    paths = [directory / name for field in ['GoFiles','CgoFiles','CFiles','CXXFiles','MFiles','HFiles','FFiles','SFiles','SysoFiles','EmbedFiles'] for name in package.get(field, [])]
    module = package.get('Module', {})
    for dependency in [module, module.get('Replace', {})]:
        if dependency.get('GoMod'):
            path = pathlib.Path(dependency['GoMod'])
            paths += [path, path.with_name('go.sum')]
    for path in paths:
        if not path.exists():
            continue
        try:
            files.add(path.resolve().relative_to(root).as_posix())
        except ValueError:
            external[str(path.resolve())] = None
settings = json.loads(subprocess.check_output(['go','env','-json','GOOS','GOARCH','GOAMD64','GOARM','GOARM64','GO386','GOMIPS','GOMIPS64','GOPPC64','GORISCV64','GOWASM','GOFLAGS','CGO_ENABLED','GOEXPERIMENT','GOWORK','GOTOOLDIR','GOTOOLCHAIN','GOENV','GOROOT','GOPATH','CC','CXX','AR','PKG_CONFIG','GODEBUG','GOFIPS140','CGO_CFLAGS','CGO_CPPFLAGS','CGO_CXXFLAGS','CGO_FFLAGS','CGO_LDFLAGS'], text=True))
for name in ['go.work','go.work.sum']:
    if (root / name).exists(): files.add(name)
if settings['GOWORK'] not in ['', 'off']:
    for path in [pathlib.Path(settings['GOWORK']), pathlib.Path(settings['GOWORK'] + '.sum')]:
        if path.exists():
            try: files.add(path.resolve().relative_to(root).as_posix())
            except ValueError: external[str(path.resolve())] = 'GOWORK ' + path.name
for name in ['compile','link']:
    external[str(pathlib.Path(settings['GOTOOLDIR']) / name)] = 'go tool ' + name
for name in ['CC','CXX','AR','PKG_CONFIG']:
    command = shlex.split(settings[name])
    path = shutil.which(command[0]) if command else None
    if path: external[str(pathlib.Path(path).resolve())] = name + ' ' + pathlib.Path(path).name
tools = [subprocess.check_output(['go','version'], text=True)]
environment = ['CPATH','C_INCLUDE_PATH','CPLUS_INCLUDE_PATH','LIBRARY_PATH','SDKROOT','MACOSX_DEPLOYMENT_TARGET','CGO_CFLAGS_ALLOW','CGO_CFLAGS_DISALLOW','CGO_LDFLAGS_ALLOW','CGO_LDFLAGS_DISALLOW']
flags = [k+'='+os.environ.get(k, '') for k in environment]
cache = pathlib.Path(subprocess.check_output(['go','env','GOCACHE'], text=True).strip()).resolve()
modules = pathlib.Path(subprocess.check_output(['go','env','GOMODCACHE'], text=True).strip()).resolve()
goroot = pathlib.Path(settings['GOROOT']).resolve()
fingerprints = []
for path, label in external.items():
    source = pathlib.Path(path)
    if label is None:
        if source.is_relative_to(cache): label = 'generated-test-main'
        elif source.is_relative_to(goroot): label = 'GOROOT ' + source.relative_to(goroot).as_posix()
        elif source.is_relative_to(modules): label = 'module ' + source.relative_to(modules).as_posix()
        else: label = path
    fingerprints.append(label + ':' + hashlib.sha256(source.read_bytes()).hexdigest())
tools.extend(sorted(fingerprints))
workspace = 'workspace=off' if settings['GOWORK'] in ['', 'off'] else 'workspace=on'
locations = {'GOWORK','GOTOOLDIR','GOTOOLCHAIN','GOENV','GOROOT','GOPATH'}
print(json.dumps({'Name':'stage3-fixture-oracle-hook','Files':sorted(files),'Flags':hook['reproducible_flags'] + ['-c','./internal/oracle','repository='+str(root), workspace] + flags + [k+'='+str(v) for k,v in sorted(settings.items()) if k not in locations], 'Toolchain':tools}))
`

// fixtureHookInputs is the oracle hook's key for repository, discovered once per process: every fixture asks for the
// product, and discovery runs python, go list, go env and go version (16 callers ran it 16 times). Parallel fixtures
// wait on the first discovery; a failed one isn't remembered.
func fixtureHookInputs(t testing.TB, repository string) buildcache.Inputs {
	t.Helper()
	fixtureHookKeys.Lock()
	defer fixtureHookKeys.Unlock()
	if inputs, ok := fixtureHookKeys.inputs[repository]; ok {
		return inputs
	}
	inputs := fixtureHookInputsEnvironment(t, repository, nil)
	if fixtureHookKeys.inputs == nil {
		fixtureHookKeys.inputs = map[string]buildcache.Inputs{}
	}
	fixtureHookKeys.inputs[repository] = inputs
	return inputs
}

var fixtureHookKeys struct {
	sync.Mutex
	inputs map[string]buildcache.Inputs
}

func fixtureHookInputsEnvironment(t testing.TB, repository string, environment []string) buildcache.Inputs {
	t.Helper()
	result := hookPython(t, repository, environment, "-c", fixtureHookInputsScript, repository)
	if result.Exit != 0 {
		t.Fatalf("oracle product input discovery: %s%s", result.Stdout, result.Stderr)
	}
	var inputs buildcache.Inputs
	if err := json.Unmarshal([]byte(result.Stdout), &inputs); err != nil {
		t.Fatal(err)
	}
	return inputs
}

// hookPython runs the oracle hook's python3 under -B, so it never writes bytecode into the tree whatever its caller
// exports: build-hook.py imports stage3/caches.py, and without it python writes stage3/__pycache__ mid-build (lap 1's
// traced fallback, a write into the tree), as stage3/verdict/run.py turns it off for its driver. -I and -S keep it off
// the box's python setup: without -S python imports site, which lists the site directories, and that listing lands in
// the product's reads, so its key turned on which packages a box has installed (lap A.3, a runner's miss on "system
// /usr/lib/python3/dist-packages"); -I drops PYTHON* settings, the user site and the script's directory on sys.path. The
// hook and its discovery need only the standard library, and build-hook.py puts stage3 on sys.path itself.
func hookPython(t testing.TB, directory string, environment []string, arguments ...string) behavior {
	t.Helper()
	return execute(t, directory, environment, "python3", append([]string{"-I", "-S", "-B"}, arguments...)...)
}

// Both the declared product and every fixture use this one recipe. Preparation
// is allowed on a cache miss, without order dependence or a building sync.Once.
func fixtureOracleHook(t testing.TB, repository string) string {
	t.Helper()
	directory := buildcache.Product(t, fixtureHookInputs(t, repository), func(directory string) error {
		// The helper's store names its manifest, build log and lock by an action key that hashes absolute paths (the
		// checkout, each package's directory, GOTOOLDIR), so they differ by box: it builds into a store of its own here,
		// and only the executable stays (#dhm3wav: Workshop and Adam held the same 47 MB executable under 892955... and
		// a9aa6f... named files).
		store := filepath.Join(directory, "store")
		result := hookPython(t, repository, nil, "stage3/fixtures/build-hook.py", "--prepare", "--store", store)
		if result.Exit != 0 {
			return fmt.Errorf("oracle preparation: %s%s", result.Stdout, result.Stderr)
		}
		// Save the content-addressed executable name. The helper's action
		// manifest includes GOCACHE paths and must not be consulted on fetch.
		name := filepath.Base(strings.TrimSpace(result.Stdout))
		if err := os.Rename(filepath.Join(store, name), filepath.Join(directory, name)); err != nil {
			return err
		}
		entries, err := os.ReadDir(store)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := os.Remove(filepath.Join(store, entry.Name())); err != nil {
				return err
			}
		}
		if err := os.Remove(store); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "product-name"), []byte(name), 0644)
	})
	name, err := os.ReadFile(filepath.Join(directory, "product-name"))
	if err != nil {
		t.Fatal(err)
	}
	if len(name) != 64 || strings.ContainsAny(string(name), "/\\") {
		t.Fatal("invalid oracle product name")
	}
	binary := filepath.Join(directory, string(name))
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != string(name) {
		t.Fatal("oracle build product hash mismatch")
	}
	return binary
}

func TestProduct_FixtureOracleHook(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := fixtureOracleHook(t, repository)
	// The product holds the executable and its name, nothing named by the helper's placeful action key. Mutant: prepare
	// into the product directory itself (<key>.json, <key>.build.log, <key>.lock beside the executable).
	entries, err := os.ReadDir(filepath.Dir(binary))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if want := []string{filepath.Base(binary), "product-name"}; !slices.Equal(names, want) {
		t.Fatalf("oracle hook product holds %q, want only %q", names, want)
	}
}

// The hook's runtime archives are this package's products too: every program it builds links a sanitized archive of
// the runtime for its features, and compiled in the hook, seven of them were 12 s of every unit before its first
// fixture finished, and most of its CPU (#43k99f8). Each test builds one group of them as internal/oracle's own product
// tests do (runtimeproducts.Build), so the tree build builds them ahead under this package too, its unit fetches them,
// and the hook's supplier reads them there. Mutant: no test here, and every fixture's hook reads its archive as not
// built.
func TestProduct_FixtureRuntimesSanitized0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized", 0)
}

func TestProduct_FixtureRuntimesSanitized1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized", 1)
}

func TestProduct_FixtureRuntimesSanitized2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized", 2)
}

func TestFixtureOracleProductTracksInputs(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	inputs := fixtureHookInputs(t, repository)
	scratch := t.TempDir()
	for _, name := range inputs.Files {
		data, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(scratch, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	before, err := buildcache.Key(scratch, inputs)
	if err != nil {
		t.Fatal(err)
	}
	// Changing the build helper is a changed recipe, even if the Go sources did
	// not change. Dropping this input must serve the old key and fail this check.
	helper := filepath.Join(scratch, "stage3/fixtures/build-hook.py")
	if err := os.MkdirAll(filepath.Dir(helper), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(helper, []byte("# changed oracle build recipe\n"), 0644); err != nil {
		t.Fatal(err)
	}
	after, err := buildcache.Key(scratch, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if after == before {
		t.Fatal("changed oracle build helper reused the same product key")
	}
}

func TestFixtureOracleProductIgnoresGoCacheLocation(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool, err := toolchain.Path("go")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, name := range []string{"first", "second"} {
		cache := filepath.Join(t.TempDir(), name)
		environment := []string{"GOCACHE=" + cache}
		if name == "second" {
			// The same tools by another PATH: go's directory again, before the PATH toolchain gives a tool.
			environment = append(environment, "PATH="+filepath.Dir(goTool)+string(os.PathListSeparator)+strings.TrimPrefix(toolchainPath(), "PATH="))
		}
		inputs := fixtureHookInputsEnvironment(t, repository, environment)
		key, err := buildcache.Key(repository, inputs)
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
	}
	if keys[0] != keys[1] {
		t.Fatalf("Go cache location changed oracle product key: %s != %s", keys[0], keys[1])
	}
}

// The hook's python writes no bytecode into the tree with PYTHONDONTWRITEBYTECODE unset: build-hook.py, run as the
// preparation runs it, in a tree holding it and the stage3/caches.py it imports, leaves no __pycache__ beside either.
// Mutant: hookPython without -B (stage3/__pycache__/caches.cpython-*.pyc).
func TestTheFixtureOracleHookWritesNoBytecode(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	scratch := t.TempDir()
	for _, name := range []string{"stage3/caches.py", "stage3/fixtures/build-hook.py"} {
		data, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(scratch, name)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(scratch, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Empty is unset, as python reads it; --help stops once the imports have run.
	environment := []string{"PYTHONDONTWRITEBYTECODE=", "STAGE3_CACHE=" + t.TempDir()}
	if result := hookPython(t, scratch, environment, "stage3/fixtures/build-hook.py", "--help"); result.Exit != 0 {
		t.Fatalf("build-hook.py --help: %s%s", result.Stdout, result.Stderr)
	}
	for _, directory := range []string{"stage3/__pycache__", "stage3/fixtures/__pycache__"} {
		if _, err := os.Stat(filepath.Join(scratch, directory)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("the hook's python wrote %s (%v)", directory, err)
		}
	}
}

// The hook's python reads no site directory and no PYTHON* setting: it never imports site, so sys.path holds the
// standard library alone, no site-packages or dist-packages, no user site, no PYTHONPATH directory and no current
// directory. Mutants: hookPython without -S (no_site 0, site imported, the box's site-packages on sys.path), without -I
// (isolated 0, PYTHONPATH's directory and the current directory on sys.path).
func TestTheFixtureOracleHooksPythonReadsNoSiteDirectory(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := t.TempDir()
	probe := "import json, os, sys; print(json.dumps({'path': sys.path, 'isolated': sys.flags.isolated, " +
		"'no_site': sys.flags.no_site, 'site': 'site' in sys.modules, 'directory': os.getcwd()}))"
	result := hookPython(t, directory, []string{"PYTHONPATH=" + path}, "-c", probe)
	if result.Exit != 0 {
		t.Fatalf("python probe: %s%s", result.Stdout, result.Stderr)
	}
	var observed struct {
		Path      []string `json:"path"`
		Isolated  int      `json:"isolated"`
		NoSite    int      `json:"no_site"`
		Site      bool     `json:"site"`
		Directory string   `json:"directory"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Isolated != 1 || observed.NoSite != 1 || observed.Site {
		t.Errorf("the hook's python runs with isolated=%d no_site=%d and site imported %v, want isolated, no site", observed.Isolated, observed.NoSite, observed.Site)
	}
	for _, entry := range observed.Path {
		if strings.Contains(entry, "site-packages") || strings.Contains(entry, "dist-packages") || entry == "" ||
			entry == path || entry == observed.Directory {
			t.Errorf("the hook's python has %q on sys.path %q", entry, observed.Path)
		}
	}
}
