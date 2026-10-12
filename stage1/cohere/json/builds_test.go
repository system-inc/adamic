package json

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
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

func jsonLoweredPortInputs() jsonBuildInputs {
	files := append(compilerinputs.Lowering(), "stage1/cohere/json/builds_test.go")
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	// It lowers in this process and runs clang, never go: its compiler is keyed by its sources (Files) and the Go
	// release, not by go env or GOFLAGS, which differ between the tree builder and a runner (#nm31pcn).
	return jsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append([]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"}, jsonBuildEnvironment("GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT")...), Toolchain: []string{runtime.Version()}}
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
	flags = append(flags, jsonBuildEnvironment("ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
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

func jsonLoweredProducts(t *testing.T, dir, entry string) jsonProducts {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(dir, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	return jsonProducts{entry: filepath.Join(dir, entry), script: filepath.Join(dir, "program.mjs"), c: string(source)}
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
