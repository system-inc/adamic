package oracle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Every process of this binary takes its runtime archives from their products (internal/native/runtimeproducts),
// stage3/fixtures' oracle hook too, this binary run once per fixture in stage3/fixtures' unit: that package declares
// the sanitized archives its hook links as its own products, which its unit fetches, rather than the hook compiling
// seven archives in every unit, 12 s before its first fixture could finish and most of its CPU (#43k99f8).
func init() {
	native.SupplyRuntimes(runtimeproducts.Supply)
}

// Every variant's archives, built ahead a third of the sets at a time (runtimeproducts.Build).
func TestProduct_NativeRuntimesSanitized0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized", 0)
}
func TestProduct_NativeRuntimesSanitized1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized", 1)
}
func TestProduct_NativeRuntimesSanitized2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized", 2)
}
func TestProduct_NativeRuntimesRelease0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "release", 0)
}
func TestProduct_NativeRuntimesRelease1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "release", 1)
}
func TestProduct_NativeRuntimesRelease2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "release", 2)
}
func TestProduct_NativeRuntimesCounted0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "counted", 0)
}
func TestProduct_NativeRuntimesCounted1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "counted", 1)
}
func TestProduct_NativeRuntimesCounted2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "counted", 2)
}
func TestProduct_NativeRuntimesWASI0(t *testing.T) { t.Parallel(); runtimeproducts.Build(t, "wasm", 0) }
func TestProduct_NativeRuntimesWASI1(t *testing.T) { t.Parallel(); runtimeproducts.Build(t, "wasm", 1) }
func TestProduct_NativeRuntimesWASI2(t *testing.T) { t.Parallel(); runtimeproducts.Build(t, "wasm", 2) }
func TestProduct_NativeRuntimesWASIRequest0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "wasm-request", 0)
}
func TestProduct_NativeRuntimesWASIRequest1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "wasm-request", 1)
}
func TestProduct_NativeRuntimesWASIRequest2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "wasm-request", 2)
}

// Each set in the table is one some program needs: its witness turns on exactly its features, so a set no program
// needs any more fails here rather than staying to be built ahead for nothing. A program needing a set the table lacks
// fails its own build (runtimeproducts.Declared). Mutants: a set's witness turning on another set; a set removed from
// the table, which its programs' builds name.
func TestRuntimeSetsAreEachUsed(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, set := range runtimeproducts.Sets {
		name := strings.Join(set.Features, "+")
		if seen[name] {
			t.Errorf("the set [%s] is in runtimeproducts.Sets twice", name)
		}
		seen[name] = true
		path, err := filepath.Abs(filepath.Join(repository, set.Witness))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Errorf("the witness of [%s], %s: %v", name, set.Witness, err)
			continue
		}
		if got := native.Features(native.C(program)); !slices.Equal(got, set.Features) {
			t.Errorf("the witness of [%s], %s, turns on %v: name a program that needs exactly the set, or drop the set", name, set.Witness, got)
		}
	}
}

// An archive's product key holds every part of its runtime cache key, as every machine spells it, and the features and
// options it was asked with: each moves the key alone (native's TestRuntimeKeyIncludesEveryInput moves the parts with
// each input), and the compiler and archiver are keyed by their reports. Mutant: a part dropped from runtimeInputs.
func TestRuntimeProductKeyHoldsEveryPart(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	request, err := native.RuntimeRequestFor(native.Options{}, "ADAMIC_CLOSURE_CONVENTION")
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.NameKey(root, runtimeproducts.Inputs(request))
	for index, part := range request.Portable {
		changed := request
		changed.Portable = slices.Clone(request.Portable)
		changed.Portable[index] += "!"
		if buildcache.NameKey(root, runtimeproducts.Inputs(changed)) == key {
			t.Errorf("changing the part %.60q kept the product key", part)
		}
	}
	// The parts the portable ones leave out are keyed by their tools' reports.
	inputs := runtimeproducts.Inputs(request)
	if !slices.Equal(inputs.Toolchain, []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")}) {
		t.Errorf("the product isn't keyed by the compiler's and archiver's reports: %q", inputs.Toolchain)
	}
	if len(request.Portable) != len(request.Parts)-3 {
		t.Errorf("the portable parts are %d of the key's %d, want all but the compiler, its version and the archiver", len(request.Portable), len(request.Parts))
	}
	// Each feature alone and each option apart from the release build's, against the plain release archive: a set's
	// key is its features' flags, so these move every set's.
	keys := map[string]string{buildcache.NameKey(root, runtimeproducts.Inputs(plain(t, native.Options{}))): "the plain release archive"}
	asks := map[string]native.RuntimeRequest{}
	for _, feature := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		asks["the feature "+feature] = plain(t, native.Options{}, feature)
	}
	for _, option := range runtimeproducts.Variants {
		if option.Name != "release" && (option.Requires == "" || os.Getenv("WASI_SYSROOT") != "") {
			asks["the "+option.Name+" options"] = plain(t, option.Options)
		}
	}
	for name, asked := range asks {
		key := buildcache.NameKey(root, runtimeproducts.Inputs(asked))
		if other, found := keys[key]; found {
			t.Errorf("%s and %s share a product key", name, other)
		}
		keys[key] = name
	}
}

func plain(t *testing.T, options native.Options, features ...string) native.RuntimeRequest {
	t.Helper()
	request, err := native.RuntimeRequestFor(options, features...)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

// One archive's product key is one key however the machine reached its compiler: a runner's $ADAMIC_TOOLS/bin/clang
// and Workshop's ~/.adamic-tools/bin/clang (a link to that directory, found on PATH) key the same, since the tree build
// keys on one and a runner reads on the other. Each spelling is a child process of its own, as each machine is.
// Mutant: runtimeInputs keyed by the raw parts, which name where clang was reached.
func TestARuntimeProductKeyIsTheSameThroughEitherLink(t *testing.T) {
	t.Parallel()
	const variable = "ADAMIC_TEST_RUNTIME_PRODUCT_KEY"
	if os.Getenv(variable) == "1" {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		for _, features := range [][]string{nil, {"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_REGEXP_REPLACE_CALLBACK"}} {
			request := plain(t, native.Options{Sanitize: true}, features...)
			fmt.Printf("compiler %s\nkey %v %s\n", request.Compiler, features, buildcache.NameKey(root, runtimeproducts.Inputs(request)))
		}
		return
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	real, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		t.Fatal(err)
	}
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tools := filepath.Join(home, "adamic-tools")
	if err := os.MkdirAll(filepath.Join(tools, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	links := map[string]string{"clang": real}
	if archiver := filepath.Join(filepath.Dir(real), "llvm-ar"); fileExists(archiver) {
		links["llvm-ar"] = archiver
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(tools, "bin", name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("adamic-tools", filepath.Join(home, ".adamic-tools")); err != nil {
		t.Fatal(err)
	}
	keys := func(environment ...string) (string, string) {
		t.Helper()
		command := exec.Command(os.Args[0], "-test.run=^TestARuntimeProductKeyIsTheSameThroughEitherLink$")
		command.Dir = "."
		for _, variable := range os.Environ() {
			if !strings.HasPrefix(variable, "ADAMIC_TOOLS=") && !strings.HasPrefix(variable, "PATH=") {
				command.Env = append(command.Env, variable)
			}
		}
		command.Env = append(command.Env, append(environment, variable+"=1")...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%v\n%s", err, output)
		}
		var compilers, keys []string
		for _, line := range strings.Split(string(output), "\n") {
			if reached, found := strings.CutPrefix(line, "compiler "); found {
				compilers = append(compilers, reached)
			}
			if strings.HasPrefix(line, "key ") {
				keys = append(keys, line)
			}
		}
		if len(keys) != 2 {
			t.Fatalf("the child gave %d keys, want 2:\n%s", len(keys), output)
		}
		return strings.Join(slices.Compact(compilers), " "), strings.Join(keys, "\n")
	}
	path := os.Getenv("PATH")
	runnerCompiler, runner := keys("ADAMIC_TOOLS="+tools, "PATH="+path)
	workshopCompiler, workshop := keys("PATH=" + filepath.Join(home, ".adamic-tools/bin") + string(os.PathListSeparator) + path)
	if runnerCompiler != filepath.Join(tools, "bin/clang") || workshopCompiler != filepath.Join(home, ".adamic-tools/bin/clang") {
		t.Fatalf("the children didn't reach clang through the two links: %s and %s", runnerCompiler, workshopCompiler)
	}
	if runner != workshop {
		t.Fatalf("one archive through two links keyed\n%s\nand\n%s", runner, workshop)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func TestProduct_NativeRuntimesShipped0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "shipped", 0)
}
func TestProduct_NativeRuntimesShipped1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "shipped", 1)
}
func TestProduct_NativeRuntimesShipped2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "shipped", 2)
}
