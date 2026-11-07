package oracle

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
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
func TestProduct_NativeRuntimesSanitized0(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "sanitized", 0)
}
func TestProduct_NativeRuntimesSanitized1(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "sanitized", 1)
}
func TestProduct_NativeRuntimesSanitized2(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "sanitized", 2)
}
func TestProduct_NativeRuntimesRelease0(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "release", 0)
}
func TestProduct_NativeRuntimesRelease1(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "release", 1)
}
func TestProduct_NativeRuntimesRelease2(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "release", 2)
}
func TestProduct_NativeRuntimesCounted0(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "counted", 0)
}
func TestProduct_NativeRuntimesCounted1(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "counted", 1)
}
func TestProduct_NativeRuntimesCounted2(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "counted", 2)
}
func TestProduct_NativeRuntimesWASI0(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "wasm", 0)
}
func TestProduct_NativeRuntimesWASI1(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "wasm", 1)
}
func TestProduct_NativeRuntimesWASI2(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "wasm", 2)
}
func TestProduct_NativeRuntimesWASIRequest0(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "wasm-request", 0)
}
func TestProduct_NativeRuntimesWASIRequest1(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "wasm-request", 1)
}
func TestProduct_NativeRuntimesWASIRequest2(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "wasm-request", 2)
}

// Each set in the table is one some program needs: its witness turns on exactly its features, so a set no program
// needs any more fails here rather than staying to be built ahead for nothing. A program needing a set the table lacks
// fails its own build (runtimeproducts.Declared). Mutants: a set's witness turning on another set; a set removed from
// the table, which its programs' builds name.
func TestRuntimeSetsAreEachUsed(test *testing.T) {
	test.Parallel()
	seen := map[string]bool{}
	for _, set := range runtimeproducts.Sets {
		name := strings.Join(set.Features, "+")
		if seen[name] {
			test.Errorf("the set [%s] is in runtimeproducts.Sets twice", name)
		}
		seen[name] = true
		path, err := filepath.Abs(filepath.Join(repository, set.Witness))
		if err != nil {
			test.Fatal(err)
		}
		program, err := lowered(test, path)
		if err != nil {
			test.Errorf("the witness of [%s], %s: %v", name, set.Witness, err)
			continue
		}
		if got := native.Features(native.C(program)); !slices.Equal(got, set.Features) {
			test.Errorf(
				"the witness of [%s], %s, turns on %v: name a program that needs exactly the set, or drop the set",
				name,
				set.Witness,
				got,
			)
		}
	}
}

// An archive's product key holds every part of its runtime cache key, as every machine spells it: each moves the key
// alone (native's TestRuntimeKeyIncludesEveryInput moves the parts with each input), and the compiler and archiver are
// keyed by their reports. The features and options it was asked with are a test of their own (the next), since each
// request runs clang over the runtime to list the headers its compile reads, and the two in one test were 39.9 s alone
// (#8k8yzn0). Mutant: a part dropped from runtimeInputs.
func TestRuntimeProductKeyHoldsEveryPart(test *testing.T) {
	test.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		test.Fatal(err)
	}
	request, err := native.RuntimeRequestFor(native.Options{}, "ADAMIC_CLOSURE_CONVENTION")
	if err != nil {
		test.Fatal(err)
	}
	key := buildcache.NameKey(root, runtimeproducts.Inputs(request))
	for index, part := range request.Portable {
		changed := request
		changed.Portable = slices.Clone(request.Portable)
		changed.Portable[index] += "!"
		if buildcache.NameKey(root, runtimeproducts.Inputs(changed)) == key {
			test.Errorf("changing the part %.60q kept the product key", part)
		}
	}
	// The parts the portable ones leave out are keyed by their tools' reports.
	inputs := runtimeproducts.Inputs(request)
	if !slices.Equal(
		inputs.Toolchain,
		[]string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")},
	) {
		test.Errorf("the product isn't keyed by the compiler's and archiver's reports: %q", inputs.Toolchain)
	}
	if len(request.Portable) != len(request.Parts)-3 {
		test.Errorf(
			"the portable parts are %d of the key's %d, want all but the compiler, its version and the archiver",
			len(request.Portable),
			len(request.Parts),
		)
	}
}

// An archive's product key holds the features and options it was asked with: each feature alone and each option apart
// from the release build's keys apart from the plain release archive and from every other, and a set's key is its
// features' flags, so these move every set's. Each request runs clang over the runtime to list the headers its compile
// reads, so the requests are asked at once and their keys compared once all are in, where one after another they were
// most of 39.9 s (#8k8yzn0). Mutant: a part dropped from runtimeInputs.
func TestRuntimeProductKeyHoldsEveryFeatureAndOption(test *testing.T) {
	test.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		test.Fatal(err)
	}
	asks := map[string]func(test *testing.T) native.RuntimeRequest{
		"the plain release archive": func(
			test *testing.T,
		) native.RuntimeRequest {
			return plain(test, native.Options{})
		},
	}
	for _, feature := range []string{
		"ADAMIC_CLOSURE_CONVENTION",
		"ADAMIC_CANONICAL_CLOSURES",
		"ADAMIC_CLOSURE_RECEIVERS",
		"ADAMIC_REGEXP_REPLACE_CALLBACK",
		"ADAMIC_NODE_HOST",
		"ADAMIC_CLOSURE_DISCARDS",
	} {
		asks["the feature "+
			feature] = func(test *testing.T) native.RuntimeRequest { return plain(test, native.Options{}, feature) }
	}
	for _, option := range runtimeproducts.Variants {
		if option.Name != "release" && (option.Requires == "" || os.Getenv("WASI_SYSROOT") != "") {
			asks["the "+
				option.Name+
				" options"] = func(test *testing.T) native.RuntimeRequest { return plain(test, option.Options) }
		}
	}
	var mutex sync.Mutex
	keys := map[string]string{}
	// Serial: a group, which returns only once its parallel requests have ended, so the comparison below reads them all.
	test.Run("requests", func(test *testing.T) {
		for name, ask := range asks {
			test.Run(name, func(test *testing.T) {
				test.Parallel()
				key := buildcache.NameKey(root, runtimeproducts.Inputs(ask(test)))
				mutex.Lock()
				defer mutex.Unlock()
				keys[name] = key
			})
		}
	})
	if test.Failed() {
		return
	}
	names := slices.Sorted(maps.Keys(keys))
	for index, name := range names {
		for _, other := range names[index+1:] {
			if keys[name] == keys[other] {
				test.Errorf("%s and %s share a product key", name, other)
			}
		}
	}
}

func plain(test *testing.T, options native.Options, features ...string) native.RuntimeRequest {
	test.Helper()
	request, err := native.RuntimeRequestFor(options, features...)
	if err != nil {
		test.Fatal(err)
	}
	return request
}

// One archive's product key is one key however the machine reached its compiler: a runner's $ADAMIC_TOOLS/bin/clang
// and Workshop's ~/.adamic-tools/bin/clang (a link to that directory, found on PATH) key the same, since the tree build
// keys on one and a runner reads on the other. Each spelling is a child process of its own, as each machine is.
// Mutant: runtimeInputs keyed by the raw parts, which name where clang was reached.
func TestARuntimeProductKeyIsTheSameThroughEitherLink(test *testing.T) {
	test.Parallel()
	const variable = "ADAMIC_TEST_RUNTIME_PRODUCT_KEY"
	if os.Getenv(variable) == "1" {
		root, err := filepath.Abs(repository)
		if err != nil {
			test.Fatal(err)
		}
		for _, features := range [][]string{nil, {"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_REGEXP_REPLACE_CALLBACK"}} {
			request := plain(test, native.Options{Sanitize: true}, features...)
			fmt.Printf(
				"compiler %s\nkey %v %s\n",
				request.Compiler,
				features,
				buildcache.NameKey(root, runtimeproducts.Inputs(request)),
			)
		}
		return
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		test.Fatal(err)
	}
	real, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		test.Fatal(err)
	}
	home, err := filepath.EvalSymlinks(test.TempDir())
	if err != nil {
		test.Fatal(err)
	}
	tools := filepath.Join(home, "adamic-tools")
	if err := os.MkdirAll(filepath.Join(tools, "bin"), 0o755); err != nil {
		test.Fatal(err)
	}
	links := map[string]string{"clang": real}
	if archiver := filepath.Join(filepath.Dir(real), "llvm-ar"); fileExists(archiver) {
		links["llvm-ar"] = archiver
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(tools, "bin", name)); err != nil {
			test.Fatal(err)
		}
	}
	if err := os.Symlink("adamic-tools", filepath.Join(home, ".adamic-tools")); err != nil {
		test.Fatal(err)
	}
	keys := func(environment ...string) (string, string) {
		test.Helper()
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
			test.Fatalf("%v\n%s", err, output)
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
			test.Fatalf("the child gave %d keys, want 2:\n%s", len(keys), output)
		}
		return strings.Join(slices.Compact(compilers), " "), strings.Join(keys, "\n")
	}
	path := os.Getenv("PATH")
	runnerCompiler, runner := keys("ADAMIC_TOOLS="+tools, "PATH="+path)
	workshopCompiler, workshop := keys(
		"PATH=" + filepath.Join(home, ".adamic-tools/bin") + string(os.PathListSeparator) + path,
	)
	if runnerCompiler != filepath.Join(tools, "bin/clang") ||
		workshopCompiler != filepath.Join(home, ".adamic-tools/bin/clang") {
		test.Fatalf(
			"the children didn't reach clang through the two links: %s and %s",
			runnerCompiler,
			workshopCompiler,
		)
	}
	if runner != workshop {
		test.Fatalf("one archive through two links keyed\n%s\nand\n%s", runner, workshop)
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func TestProduct_NativeRuntimesShipped0(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "shipped", 0)
}
func TestProduct_NativeRuntimesShipped1(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "shipped", 1)
}
func TestProduct_NativeRuntimesShipped2(test *testing.T) {
	test.Parallel()
	runtimeproducts.Build(test, "shipped", 2)
}
