package json

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

var portMatchesShared struct {
	once                                      sync.Once
	ready                                     bool
	oracle, entry, script, release, sanitized string
}

func TestPortMatchesGoCohereSplit_Setup(test *testing.T) {
	test.Parallel()
	portMatchesPrepare(test)
}

// Shared builds have no test deadline; the process timeout covers preparation.
func portMatchesPrepare(test *testing.T) {
	test.Helper()
	portMatchesShared.once.Do(func() { portMatchesPrepareProducts(test) })
	if !portMatchesShared.ready {
		test.Fatal("shared preparation failed")
	}
}

func portMatchesPrepareProducts(test *testing.T) {
	started := time.Now()
	defer func() { test.Logf("TestPortMatchesGoCohere (setup): %.3fs", time.Since(started).Seconds()) }()
	loweredDir := portMatchesProduct(test, "lowered")
	portMatchesShared.oracle = portMatchesProduct(test, "oracle")
	portMatchesShared.entry = filepath.Join(loweredDir, "main.ts")
	portMatchesShared.script = filepath.Join(loweredDir, "program.mjs")
	portMatchesShared.release = filepath.Join(portMatchesProduct(test, "release"), "port")
	portMatchesShared.sanitized = filepath.Join(portMatchesProduct(test, "sanitized"), "port")
	clangTools, portMatchesJsonClangToolchainError := portMatchesJsonClangToolchain()
	if portMatchesJsonClangToolchainError != nil {
		test.Fatal(portMatchesJsonClangToolchainError)
	}
	// The four-digit port wrappers use the same immutable products.
	// Publish the same products for either independently selected shard family.
	jsonTopCorpus(test)
	jsonTopState.clangTools = clangTools
	jsonTopState.oracle = portMatchesShared.oracle
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

func portMatchesJsonClangToolchain() ([]string, error) {
	clang, lookPathError := exec.LookPath("clang")
	if lookPathError != nil {
		return nil, lookPathError
	}
	if _, outputError := exec.Command(clang, "--version").Output(); outputError != nil {
		return nil, outputError
	}
	archiver := filepath.Join(filepath.Dir(clang), "llvm-ar")
	if info, statError := os.Stat(
		archiver,
	); statError != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		var lookPathError error
		archiver, lookPathError = exec.LookPath("ar")
		if lookPathError != nil {
			return nil, lookPathError
		}
	}
	if _, outputError := exec.Command(archiver, "--version").Output(); outputError != nil {
		return nil, outputError
	}
	// Each tool by its report, which names what it is and not where this machine keeps it (#t37sw0f).
	return []string{
		buildcache.Tool(clang, "--version"),
		buildcache.Tool(archiver, "--version"),
		"GOOS=" + runtime.GOOS,
		"GOARCH=" + runtime.GOARCH,
	}, nil
}

func portMatchesJsonLoweredPortInputs() portMatchesJsonBuildInputs {
	files := append(compilerinputs.Lowering(), "stage1/cohere/json/port_matches_prepared_split_test.go")
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	// It lowers in this process and runs clang, never go: its compiler is keyed by its sources (Files) and the Go
	// release, not by go env or GOFLAGS, which differ between the tree builder and a runner (#nm31pcn).
	return portMatchesJsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append(
		[]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"},
		portMatchesJsonBuildEnvironment("GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT")...,
	), Toolchain: []string{runtime.Version()}}
}

// The immutable lowering product contains the copied TypeScript and both emitted
// backends. Persisting emitted sources avoids serializing IR interface values;
// fetching this product will avoid loading, checking, lowering and emission.
func portMatchesBuildJSONLoweredPort(dir string) error {
	for _, name := range portFiles {
		contents, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		if writeError := os.WriteFile(filepath.Join(dir, name), contents, 0644); writeError != nil {
			return writeError
		}
	}
	program, loadError := load.Load([]string{filepath.Join(dir, "main.ts")})
	if loadError != nil {
		return fmt.Errorf("Load: %w", loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return fmt.Errorf("Lower: %w", lowerError)
	}
	if writeError := os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644); writeError != nil {
		return writeError
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
	flags = append(
		flags,
		fmt.Sprintf("generated-C-sha256=%x", sha256.Sum256([]byte(source))),
		fmt.Sprintf("options=%+v", options),
		"-I=<runtime product>",
		"-o=<dir>/port",
		"-lm",
	)
	flags = append(flags, portMatchesJsonBuildEnvironment(
		"ADAMIC_NATIVE_SPLIT",
		"CPATH",
		"C_INCLUDE_PATH",
		"LIBRARY_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
	)...)
	return portMatchesJsonBuildInputs{
		Name:      name,
		Files:     []string{"internal", "stage1/cohere/json/port_matches_prepared_split_test.go"},
		Flags:     flags,
		Toolchain: toolchain,
	}
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

// portMatchesProduct returns the named product's directory, except the oracle's, which is its binary.
func portMatchesProduct(test *testing.T, name string) string {
	test.Helper()
	value, _ := portMatchesProducts.LoadOrStore(name, &jsonTopBuild{})
	product := value.(*jsonTopBuild)
	product.once.Do(func() {
		switch name {
		case "oracle":
			product.dir = oracles.FormatterOracle(test)
		case "lowered":
			product.dir = buildcache.Product(test, portMatchesJsonLoweredPortInputs(), portMatchesBuildJSONLoweredPort)
		case "release", "sanitized":
			lowered := portMatchesProduct(test, "lowered")
			source, readError := os.ReadFile(filepath.Join(lowered, "main.c"))
			if readError != nil {
				test.Fatal(readError)
			}
			clangTools, portMatchesJsonClangToolchainError := portMatchesJsonClangToolchain()
			if portMatchesJsonClangToolchainError != nil {
				test.Fatal(portMatchesJsonClangToolchainError)
			}
			build := portMatchesBuildJSONReleasePort(string(source))
			if name == "sanitized" {
				build = portMatchesBuildJSONSanitizedPort(string(source))
			}
			product.dir = buildcache.Product(
				test,
				portMatchesJsonNativePortInputs(string(source), name == "sanitized", clangTools),
				build,
			)
		default:
			test.Fatalf("unknown product %q", name)
		}
	})
	if product.dir == "" {
		test.Fatalf("product %s preparation failed", name)
	}
	return product.dir
}

func TestProduct_JsonFindsGoOracle(test *testing.T) {
	test.Parallel()
	portMatchesProduct(test, "oracle")
}
func TestProduct_JSONLoweredPort(test *testing.T) {
	test.Parallel()
	portMatchesProduct(test, "lowered")
}
func TestProduct_JSONNativeRelease(test *testing.T) {
	test.Parallel()
	portMatchesProduct(test, "release")
}
func TestProduct_JSONNativeSanitized(test *testing.T) {
	test.Parallel()
	portMatchesProduct(test, "sanitized")
}
