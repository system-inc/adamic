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
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Build inputs use the shared cache API; no package-local disk cache.
type jsonBuildInputs = buildcache.Inputs

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
	settings, err = json.Marshal(resolved)
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
	files := []string{"go.mod", "go.work", "internal", "cohere/TypeScript", "cohere/TypeScript-shim", "cohere/rule_runner", "cohere/static_single_assignment", "cohere/mutation_aliasing", "stage1/cohere/json/builds_test.go"}
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	return jsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append([]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"}, jsonBuildEnvironment("PATH", "GOFLAGS", "GOTOOLCHAIN", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT", "GOENV")...), Toolchain: toolchain}
}

// The immutable lowering product contains the copied TypeScript and both emitted
// backends. Persisting emitted sources avoids serializing IR interface values;
// fetching this product will avoid loading, checking, lowering and emission.
func buildJSONLoweredPort(dir string) error { return buildJSONLoweredPortMutation(dir, nil) }

func buildJSONLoweredPortMutation(dir string, mutation *printerMutation) error {
	for _, name := range portFiles {
		contents, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if mutation != nil && name == mutation.file {
			if strings.Count(string(contents), mutation.from) != 1 {
				return fmt.Errorf("mutant %s must change one place", mutation.name)
			}
			contents = []byte(strings.Replace(string(contents), mutation.from, mutation.to, 1))
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

// jsonRunBuildUnit preserves exact gate selectors: filtered build children still
// prepare their inputs, and local scratch belongs to the parent.
func jsonRunBuildUnit(parent *testing.T, name string, prepare func(testing.TB) string) string {
	var dir string
	called := false
	run := func(t testing.TB) { called = true; dir = prepare(t) }
	if !parent.Run(name, func(t *testing.T) { run(t) }) {
		parent.FailNow()
	}
	if !called {
		run(parent)
	}
	return dir
}

func jsonBuildUnit(parent *testing.T, name string, in buildcache.Inputs, build func(string) error) string {
	return jsonRunBuildUnit(parent, name, func(t testing.TB) string {
		start := time.Now()
		dir := buildcache.Product(t, in, build)
		t.Logf("shared %s %.3fs; inputs %s", name, time.Since(start).Seconds(), in.Name)
		return dir
	})
}

// GoBuild refuses overlays. Until this oracle has a normal repository package,
// build it locally, once per parent, without hand-listing Go cache inputs.
func jsonGoOracleUnit(parent *testing.T) string {
	return jsonRunBuildUnit(parent, "build-go-oracle", func(t testing.TB) string {
		start := time.Now()
		dir := parent.TempDir()
		if err := buildJSONGoOracle(dir); err != nil {
			t.Fatal(err)
		}
		t.Logf("local Go oracle overlay build %.3fs", time.Since(start).Seconds())
		return dir
	})
}

func jsonLoweredProducts(t *testing.T, dir, entry string) jsonProducts {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(dir, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	return jsonProducts{entry: filepath.Join(dir, entry), script: filepath.Join(dir, "program.mjs"), c: string(source)}
}

func jsonPrepareMutation(t *testing.T, mutation printerMutation) jsonProducts {
	tools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	inputs := jsonLoweredPortInputs(tools)
	inputs.Name = "json lowered printer mutant"
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=%+v", mutation))
	dir := jsonBuildUnit(t, "build-lowered-mutant", inputs, func(dir string) error { return buildJSONLoweredPortMutation(dir, &mutation) })
	return jsonLoweredProducts(t, dir, "main.ts")
}

func jsonPrepareFixture(t *testing.T, path string) jsonProducts {
	tools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	inputs := jsonLoweredPortInputs(tools)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(root, absolute)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "json lowered gap fixture"
	inputs.Files = append(inputs.Files, filepath.ToSlash(relative))
	inputs.Flags = append(inputs.Flags, "entry="+filepath.ToSlash(relative))
	dir := jsonBuildUnit(t, "build-lowered-fixture", inputs, func(dir string) error {
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entry := filepath.Join(dir, filepath.Base(path))
		if err := os.WriteFile(entry, source, 0644); err != nil {
			return err
		}
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	})
	return jsonLoweredProducts(t, dir, filepath.Base(path))
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

type jsonProducts struct{ oracle, entry, script, release, sanitized, c string }

func jsonPreparePort(t *testing.T, release, sanitized bool) jsonProducts {
	t.Helper()
	tools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	oracle := jsonGoOracleUnit(t)
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

// A warm product must have the same key despite Go's fresh compiler scratch path.
func TestJSONGoToolchainStable(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var previous string
	for i := 0; i < 2; i++ {
		tools, err := jsonGoToolchain()
		if err != nil {
			t.Fatal(err)
		}
		key, err := buildcache.Key(root, buildcache.Inputs{Name: "json toolchain stability", Files: []string{"go.mod"}, Toolchain: tools})
		if err != nil {
			t.Fatal(err)
		}
		if i > 0 && key != previous {
			t.Fatal("unchanged Go toolchain produced different product keys")
		}
		previous = key
	}
}
