package json

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var portMatchesShared struct {
	once                                      sync.Once
	ready                                     bool
	oracle, entry, script, release, sanitized string
}

func TestPortMatchesGoCohereSplit_Setup(t *testing.T) {
	t.Parallel()
	portMatchesPrepare(t)
}

// Shared builds have no test deadline; the process timeout covers preparation.
func portMatchesPrepare(t *testing.T) {
	t.Helper()
	portMatchesShared.once.Do(func() { portMatchesPrepareProducts(t) })
	if !portMatchesShared.ready {
		t.Fatal("shared preparation failed")
	}
}

func portMatchesPrepareProducts(t *testing.T) {
	started := time.Now()
	defer func() { t.Logf("TestPortMatchesGoCohere (setup): %.3fs", time.Since(started).Seconds()) }()
	oracleDir := portMatchesProduct(t, "oracle")
	loweredDir := portMatchesProduct(t, "lowered")
	portMatchesShared.oracle = filepath.Join(oracleDir, "go-cohere")
	portMatchesShared.entry = filepath.Join(loweredDir, "main.ts")
	portMatchesShared.script = filepath.Join(loweredDir, "program.mjs")
	portMatchesShared.release = filepath.Join(portMatchesProduct(t, "release"), "port")
	portMatchesShared.sanitized = filepath.Join(portMatchesProduct(t, "sanitized"), "port")
	goTools, err := portMatchesJsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	clangTools, err := portMatchesJsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	// The four-digit port/upstream wrappers use the same immutable products.
	// Publish the same products for either independently selected shard family.
	jsonTopCorpus(t)
	jsonTopState.goOnce.Do(func() { jsonTopState.goTools = goTools })
	jsonTopState.clangOnce.Do(func() { jsonTopState.clangTools = clangTools })
	jsonTopState.oracle.Do(func() {
		jsonTopState.oracleDir = oracleDir
		jsonTopState.oracleCached = true
	})
	for name, directory := range map[string]string{
		"lowered":   loweredDir,
		"release":   filepath.Dir(portMatchesShared.release),
		"sanitized": filepath.Dir(portMatchesShared.sanitized),
	} {
		product := &jsonTopBuild{dir: directory}
		product.once.Do(func() {})
		jsonTopState.products.Store(name, product)
	}
	portMatchesShared.ready = true
}

// Build inputs use the shared cache API; no package-local disk cache.
type portMatchesJsonBuildInputs = buildcache.Inputs

func portMatchesJsonBuildEnvironment(names ...string) []string {
	var flags []string
	for _, name := range names {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	return flags
}

func portMatchesJsonGoToolchain() ([]string, error) {
	command := exec.Command("go", "version")
	directory, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	command.Dir = directory
	version, err := command.Output()
	if err != nil {
		return nil, err
	}
	configuration := exec.Command("go", "env", "-json")
	configuration.Dir = directory
	settings, err := configuration.Output()
	if err != nil {
		return nil, err
	}
	// Go generates a fresh scratch directory even for go env. It is mapped to
	// /tmp/go-build in compiler output and cannot affect a product, so normalize
	// only that generated prefix-map argument; retain every real compiler flag.
	var resolved map[string]string
	if err := json.Unmarshal(settings, &resolved); err != nil {
		return nil, err
	}
	compilerFlags := strings.Fields(resolved["GOGCCFLAGS"])
	for i, flag := range compilerFlags {
		if strings.HasPrefix(flag, "-ffile-prefix-map=") && strings.HasSuffix(flag, "=/tmp/go-build") {
			compilerFlags[i] = "-ffile-prefix-map=<Go scratch>=/tmp/go-build"
		}
	}
	resolved["GOGCCFLAGS"] = strings.Join(compilerFlags, " ")
	// Where go keeps things and the policy that picked its release say nothing about a product (the release is go
	// version above), and differ between Workshop and a runner (#t37sw0f).
	for _, name := range []string{"GOROOT", "GOTOOLDIR", "GOPATH", "GOENV", "GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOBIN", "GOCACHEPROG", "GOPROXY", "GOSUMDB", "GOPRIVATE", "GONOPROXY", "GONOSUMDB", "GOINSECURE", "GOAUTH", "GOVCS", "GOTOOLCHAIN", "GOTELEMETRY", "GOTELEMETRYDIR"} {
		delete(resolved, name)
	}
	settings, err = json.Marshal(resolved)
	if err != nil {
		return nil, err
	}
	// Include resolved Go settings as well as explicit environment overrides.
	return []string{"runtime.Version=" + runtime.Version(), string(version), "go env=" + string(settings), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
}

func portMatchesJsonClangToolchain() ([]string, error) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		return nil, err
	}
	if _, err := exec.Command(clang, "--version").Output(); err != nil {
		return nil, err
	}
	archiver := filepath.Join(filepath.Dir(clang), "llvm-ar")
	if info, err := os.Stat(archiver); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		archiver, err = exec.LookPath("ar")
		if err != nil {
			return nil, err
		}
	}
	if _, err := exec.Command(archiver, "--version").Output(); err != nil {
		return nil, err
	}
	// Each tool by its report, which names what it is and not where this machine keeps it (#t37sw0f).
	return []string{buildcache.Tool(clang, "--version"), buildcache.Tool(archiver, "--version"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
}

// portMatchesBuildJSONGoOracle writes the driver and overlay only into its product directory.
// Go's usual module/action caches remain managed by the Go toolchain.
func portMatchesBuildJSONGoOracle(dir string) error {
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return err
	}
	source, err := filepath.Abs("testdata/cohere_driver.go")
	if err != nil {
		return err
	}
	overlay := filepath.Join(dir, "overlay.json")
	encoded, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "command/formatter_comparison/main.go"): source}})
	if err != nil {
		return err
	}
	if err := os.WriteFile(overlay, encoded, 0644); err != nil {
		return err
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(overlay)
	ctx := context.Background()
	command := exec.CommandContext(ctx, "go", "build", "-overlay="+overlay, "-o", filepath.Join(dir, "go-cohere"), "./command/formatter_comparison")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	command.Dir = cohere
	if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
		return fmt.Errorf("Go driver: %w\n%s", err, output)
	}
	return nil
}

func portMatchesJsonLoweredPortInputs() portMatchesJsonBuildInputs {
	files := []string{"go.mod", "go.work", "internal", "cohere/TypeScript", "cohere/TypeScript-shim", "cohere/rule_runner", "cohere/static_single_assignment", "cohere/mutation_aliasing", "stage1/cohere/json/port_matches_prepared_split_test.go"}
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	// It lowers in this process and runs clang, never go: its compiler is keyed by its sources (Files) and the Go
	// release, not by go env or GOFLAGS, which differ between the tree builder and a runner (#nm31pcn).
	return portMatchesJsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append([]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"}, portMatchesJsonBuildEnvironment("GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT")...), Toolchain: []string{runtime.Version()}}
}

// The immutable lowering product contains the copied TypeScript and both emitted
// backends. Persisting emitted sources avoids serializing IR interface values;
// fetching this product will avoid loading, checking, lowering and emission.
func portMatchesBuildJSONLoweredPort(dir string) error {
	for _, name := range portFiles {
		contents, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), contents, 0644); err != nil {
			return err
		}
	}
	program, err := load.Load([]string{filepath.Join(dir, "main.ts")})
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return fmt.Errorf("Lower: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
}

func portMatchesJsonNativePortInputs(source string, sanitize bool, toolchain []string) portMatchesJsonBuildInputs {
	options := native.Options{Sanitize: sanitize}
	name := "json native release port"
	if sanitize {
		name = "json native ASan UBSan port"
	}
	flags := append([]string(nil), native.Flags(options)...)
	// Generated C is a captured input, not a repository file. Its digest includes
	// emitted feature defines; internal/ also covers runtime sources/link policy.
	flags = append(flags, fmt.Sprintf("generated-C-sha256=%x", sha256.Sum256([]byte(source))), fmt.Sprintf("options=%+v", options), "-I=<runtime product>", "-o=<dir>/port", "-lm")
	flags = append(flags, portMatchesJsonBuildEnvironment("ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
	return portMatchesJsonBuildInputs{Name: name, Files: []string{"internal", "stage1/cohere/json/port_matches_prepared_split_test.go"}, Flags: flags, Toolchain: toolchain}
}

// Each factory returns the exact func(dir string) error accepted by buildcache.
// Native.Build retains its existing compiler scratch/runtime action caches.
func portMatchesBuildJSONReleasePort(source string) func(string) error {
	return func(dir string) error { return native.Build(source, filepath.Join(dir, "port"), native.Options{}) }
}
func portMatchesBuildJSONSanitizedPort(source string) func(string) error {
	return func(dir string) error {
		return native.Build(source, filepath.Join(dir, "port"), native.Options{Sanitize: true})
	}
}

// Each independently reported unit owns a full 90-second budget.
func portMatchesDeadline(name string) *time.Timer {
	return time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s exceeded its 90s deadline\n", name)
		os.Exit(124)
	})
}

// The build phase and isolated shards use these exact recipes and keys.
var portMatchesProducts sync.Map

func portMatchesProduct(t *testing.T, name string) string {
	t.Helper()
	value, _ := portMatchesProducts.LoadOrStore(name, &jsonTopBuild{})
	product := value.(*jsonTopBuild)
	product.once.Do(func() {
		goTools, err := portMatchesJsonGoToolchain()
		if err != nil {
			t.Fatal(err)
		}
		switch name {
		case "oracle":
			product.dir = buildcache.Product(t, buildcache.Inputs{
				Name:      "json split Go oracle",
				Files:     []string{"go.mod", "go.work", "cohere", "stage1/cohere/json/testdata/cohere_driver.go", "stage1/cohere/json/port_matches_prepared_split_test.go"},
				Flags:     append([]string{"go build -overlay=<product>/overlay.json -o=<product>/go-cohere ./command/formatter_comparison"}, portMatchesJsonBuildEnvironment("GOFLAGS")...),
				Toolchain: goTools,
			}, portMatchesBuildJSONGoOracle)
		case "lowered":
			product.dir = buildcache.Product(t, portMatchesJsonLoweredPortInputs(), portMatchesBuildJSONLoweredPort)
		case "release", "sanitized":
			lowered := portMatchesProduct(t, "lowered")
			source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
			if err != nil {
				t.Fatal(err)
			}
			clangTools, err := portMatchesJsonClangToolchain()
			if err != nil {
				t.Fatal(err)
			}
			build := portMatchesBuildJSONReleasePort(string(source))
			if name == "sanitized" {
				build = portMatchesBuildJSONSanitizedPort(string(source))
			}
			product.dir = buildcache.Product(t, portMatchesJsonNativePortInputs(string(source), name == "sanitized", clangTools), build)
		default:
			t.Fatalf("unknown product %q", name)
		}
	})
	if product.dir == "" {
		t.Fatalf("product %s preparation failed", name)
	}
	return product.dir
}

func TestProduct_JSONGoOracle(t *testing.T)        { t.Parallel(); portMatchesProduct(t, "oracle") }
func TestProduct_JSONLoweredPort(t *testing.T)     { t.Parallel(); portMatchesProduct(t, "lowered") }
func TestProduct_JSONNativeRelease(t *testing.T)   { t.Parallel(); portMatchesProduct(t, "release") }
func TestProduct_JSONNativeSanitized(t *testing.T) { t.Parallel(); portMatchesProduct(t, "sanitized") }
