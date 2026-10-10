package fixtures

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
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
print(json.dumps({'Name':'stage3-fixture-oracle-hook','Files':sorted(files),'Flags':['-buildvcs=false','-c','./internal/oracle','repository='+str(root), workspace] + flags + [k+'='+str(v) for k,v in sorted(settings.items()) if k not in locations], 'Toolchain':tools}))
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
	result := execute(t, repository, environment, "python3", "-c", fixtureHookInputsScript, repository)
	if result.Exit != 0 {
		t.Fatalf("oracle product input discovery: %s%s", result.Stdout, result.Stderr)
	}
	var inputs buildcache.Inputs
	if err := json.Unmarshal([]byte(result.Stdout), &inputs); err != nil {
		t.Fatal(err)
	}
	return inputs
}

// Both the declared product and every fixture use this one recipe. Preparation
// is allowed on a cache miss, without order dependence or a building sync.Once.
func fixtureOracleHook(t testing.TB, repository string) string {
	t.Helper()
	directory := buildcache.Product(t, fixtureHookInputs(t, repository), func(directory string) error {
		result := execute(t, repository, nil, "python3", "stage3/fixtures/build-hook.py", "--prepare", "--store", directory)
		if result.Exit != 0 {
			return fmt.Errorf("oracle preparation: %s%s", result.Stdout, result.Stderr)
		}
		// Save the content-addressed executable name. The helper's action
		// manifest includes GOCACHE paths and must not be consulted on fetch.
		name := filepath.Base(strings.TrimSpace(result.Stdout))
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
	fixtureOracleHook(t, repository)
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
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, name := range []string{"first", "second"} {
		cache := filepath.Join(t.TempDir(), name)
		environment := []string{"GOCACHE=" + cache}
		if name == "second" {
			environment = append(environment, "PATH="+filepath.Dir(goTool)+string(os.PathListSeparator)+os.Getenv("PATH"))
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
