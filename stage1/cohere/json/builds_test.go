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
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Same field layout as the forthcoming internal/buildcache.Inputs. The caller
// can become buildcache.Product(t, buildcache.Inputs(in), build) without changing
// these builders. There is no package-local disk cache.
type jsonBuildInputs struct {
	Name                    string
	Files, Flags, Toolchain []string
}

func jsonBuildEnvironment(names ...string) []string {
	var flags []string
	for _, name := range names {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	return flags
}

func jsonGoToolchain() ([]string, error) {
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
	// Include resolved Go settings as well as explicit environment overrides.
	return []string{"runtime.Version=" + runtime.Version(), string(version), "go env=" + string(settings), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
}

func jsonClangToolchain() ([]string, error) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		return nil, err
	}
	version, err := exec.Command(clang, "--version").Output()
	if err != nil {
		return nil, err
	}
	archiver := filepath.Join(filepath.Dir(clang), "llvm-ar")
	if info, err := os.Stat(archiver); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		archiver, err = exec.LookPath("ar")
		if err != nil {
			return nil, err
		}
	}
	arVersion, err := exec.Command(archiver, "--version").Output()
	if err != nil {
		return nil, err
	}
	return []string{clang, string(version), archiver, string(arVersion), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
}

func jsonGoOracleInputs(toolchain []string) jsonBuildInputs {
	flags := []string{"go", "build", "-overlay=<dir>/overlay.json", "-o=<dir>/go-cohere", "./command/formatter_comparison"}
	flags = append(flags, jsonBuildEnvironment("PATH", "GOFLAGS", "GOTOOLCHAIN", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT", "GOCACHE", "GOMODCACHE", "GOWORK", "GOPATH", "GOENV", "GOROOT", "GO386", "GOARM", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS")...)
	return jsonBuildInputs{Name: "json Go oracle", Files: []string{"cohere", "stage1/cohere/json/testdata/cohere_driver.go", "stage1/cohere/json/builds_test.go"}, Flags: flags, Toolchain: toolchain}
}

// buildJSONGoOracle writes the driver and overlay only into its product directory.
// Go's usual module/action caches remain managed by the Go toolchain.
func buildJSONGoOracle(dir string) error {
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
	command := exec.Command("go", "build", "-overlay="+overlay, "-o", filepath.Join(dir, "go-cohere"), "./command/formatter_comparison")
	command.Dir = cohere
	if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
		return fmt.Errorf("Go driver: %w\n%s", err, output)
	}
	return nil
}

func jsonLoweredPortInputs(toolchain []string) jsonBuildInputs {
	files := []string{"go.mod", "go.sum", "internal", "cohere/TypeScript", "stage1/cohere/json/builds_test.go"}
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	return jsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append([]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"}, jsonBuildEnvironment("PATH", "GOFLAGS", "GOTOOLCHAIN", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT", "GOENV")...), Toolchain: toolchain}
}

// The immutable lowering product contains the copied TypeScript and both emitted
// backends. Persisting emitted sources avoids serializing IR interface values;
// fetching this product will avoid loading, checking, lowering and emission.
func buildJSONLoweredPort(dir string) error {
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

func jsonNativePortInputs(source string, sanitize bool, toolchain []string) jsonBuildInputs {
	options := native.Options{Sanitize: sanitize}
	name := "json native release port"
	if sanitize {
		name = "json native ASan UBSan port"
	}
	flags := append([]string(nil), native.Flags(options)...)
	// Generated C is a captured input, not a repository file. Its digest includes
	// emitted feature defines; internal/ also covers runtime sources/link policy.
	flags = append(flags, fmt.Sprintf("generated-C-sha256=%x", sha256.Sum256([]byte(source))), fmt.Sprintf("options=%+v", options), "-I=<runtime product>", "-o=<dir>/port", "-lm")
	flags = append(flags, jsonBuildEnvironment("PATH", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED", "XDG_CACHE_HOME", "HOME", "TMPDIR", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
	return jsonBuildInputs{Name: name, Files: []string{"internal", "stage1/cohere/json/builds_test.go"}, Flags: flags, Toolchain: toolchain}
}

// Each factory returns the exact func(dir string) error accepted by buildcache.
// Native.Build retains its existing compiler scratch/runtime action caches.
func buildJSONReleasePort(source string) func(string) error {
	return func(dir string) error { return native.Build(source, filepath.Join(dir, "port"), native.Options{}) }
}
func buildJSONSanitizedPort(source string) func(string) error {
	return func(dir string) error {
		return native.Build(source, filepath.Join(dir, "port"), native.Options{Sanitize: true})
	}
}

// A named build census unit runs once before parallel children. If -run selects
// only a logic child, its fetched inputs still need preparing. TempDir belongs
// to the parent so it outlives the build subtest and all consumers.
func jsonBuildUnit(parent *testing.T, name string, in jsonBuildInputs, build func(string) error) string {
	dir := parent.TempDir()
	called := false
	prepare := func(t testing.TB) {
		called = true
		start := time.Now()
		// Replace this call with: dir = buildcache.Product(t, buildcache.Inputs(in), build).
		if err := build(dir); err != nil {
			t.Fatal(err)
		}
		t.Logf("shared %s %.3fs; inputs %s; files %v; %d flags; %d toolchain entries", name, time.Since(start).Seconds(), in.Name, in.Files, len(in.Flags), len(in.Toolchain))
	}
	if !parent.Run(name, func(t *testing.T) { prepare(t) }) {
		parent.FailNow()
	}
	if !called {
		prepare(parent)
	}
	return dir
}

func jsonOracleAnswers(t *testing.T, oracle string, cases []textCase) []answer {
	t.Helper()
	input, _ := protocol(cases, make([]answer, len(cases)))
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	result := execute(t, nil, oracle, "--cases", path)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("%s Go oracle: exit %d stderr %s", t.Name(), result.exitCode, result.stderr)
	}
	answers, err := jsonPortAnswers(string(result.stdout), len(cases))
	if err != nil {
		t.Fatal(err)
	}
	return answers
}

type jsonProducts struct{ oracle, entry, script, release, sanitized string }

func jsonPreparePort(t *testing.T, release, sanitized bool) jsonProducts {
	t.Helper()
	tools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	oracle := jsonBuildUnit(t, "build-go-oracle", jsonGoOracleInputs(tools), buildJSONGoOracle)
	lowered := jsonBuildUnit(t, "build-lowered-port", jsonLoweredPortInputs(tools), buildJSONLoweredPort)
	products := jsonProducts{oracle: filepath.Join(oracle, "go-cohere"), entry: filepath.Join(lowered, "main.ts"), script: filepath.Join(lowered, "program.mjs")}
	if !release && !sanitized {
		return products
	}
	clang, err := jsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	if release {
		products.release = filepath.Join(jsonBuildUnit(t, "build-release", jsonNativePortInputs(string(source), false, clang), buildJSONReleasePort(string(source))), "port")
	}
	if sanitized {
		products.sanitized = filepath.Join(jsonBuildUnit(t, "build-sanitized", jsonNativePortInputs(string(source), true, clang), buildJSONSanitizedPort(string(source))), "port")
	}
	return products
}
