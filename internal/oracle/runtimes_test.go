package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Every program the oracle builds natively links an archive of the runtime compiled for its features (the leading
// #defines its C turns on) and its options, and an archive is about 56 C files. Compiled where a test first needed it,
// that was 22 s of a 27 s test alone on one core, and 62 to 74 s at a 62-CPU share, where the program's own work is 5
// (#sqmn7jk, lap 3's red): every unit compiled them all, since none was a product. So each archive is a buildcache
// product the tree build builds ahead (TestProduct_NativeRuntimes*), and native asks this process's supplier for it
// (supplyRuntime), which a runner reads and never builds: one it lacks fails the build as not built.

// A runtimeSet is one set of features some program the oracle builds turns on, in native's order (native.Features),
// and witness, a program that turns on exactly these, which TestRuntimeSetsAreEachUsed lowers so a set no program
// needs any more can't stay in the table.
type runtimeSet struct {
	features []string
	witness  string
}

// runtimeSets are every feature set the oracle's programs need, counted at 91671a25be over every lowered fixture and
// review program (base 1021, canonical closures 46, convention 18, receivers 14, regexp callback 6, node host 5, the
// rest 1 to 4). A program whose set isn't here fails its build, naming the set to add.
var runtimeSets = []runtimeSet{
	{nil, "dedication/dedication.a"},
	{[]string{"ADAMIC_CLOSURE_CONVENTION"}, "internal/oracle/testdata/array_from_undefined.a"},
	{[]string{"ADAMIC_CANONICAL_CLOSURES"}, "internal/oracle/testdata/nested_minimal.a"},
	{[]string{"ADAMIC_CLOSURE_RECEIVERS"}, "internal/oracle/testdata/census_optional_values_system.a"},
	{[]string{"ADAMIC_REGEXP_REPLACE_CALLBACK"}, "internal/oracle/testdata/regexp_replace/argument_guard.a"},
	{[]string{"ADAMIC_NODE_HOST"}, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_fd_ownership.a"},
	{[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES"}, "internal/oracle/testdata/closure_convention_nested.a"},
	{[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CLOSURE_RECEIVERS"}, "internal/oracle/testdata/arguments_length_extended.a"},
	{[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_REGEXP_REPLACE_CALLBACK"}, "internal/oracle/testdata/regexp_replace/callback.a"},
	{[]string{"ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS"}, "internal/oracle/testdata/getters_census_binary_inline.a"},
	{[]string{"ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"}, "internal/oracle/testdata/closure_convention_host24.a"},
}

// A runtimeOption is one way the oracle builds a program: sanitized for the comparison and the leak check, release as
// a user's build, counted for the counts table and macOS's leak check, and wasm for the WASI families, which run
// only where requires is set.
type runtimeOption struct {
	name     string
	options  native.Options
	requires string
}

var runtimeOptions = []runtimeOption{
	{"sanitized", native.Options{Sanitize: true}, ""},
	{"release", native.Options{}, ""},
	{"counted", native.Options{Count: true}, ""},
	{"wasm", native.Options{Target: "wasm32-wasi"}, "ADAMIC_ORACLE_WASI"},
}

func init() {
	native.SupplyRuntimes(supplyRuntime)
}

// supplyRuntime gives native the archive a build asks for from its product: built here the first time on a machine
// that builds (one unit at a time, as a test waits on it), never on a runner, which reads it or fails as not built.
func supplyRuntime(request native.RuntimeRequest) (string, error) {
	if err := runtimeDeclared(request); err != nil {
		return "", err
	}
	return buildcache.Get(runtimeInputs(request), func(directory string) error { return request.Build(directory, 1) })
}

// runtimeDeclared says why a request is for an archive the tree build doesn't build ahead, or nil when it is one.
func runtimeDeclared(request native.RuntimeRequest) error {
	setDeclared := slices.ContainsFunc(runtimeSets, func(set runtimeSet) bool { return slices.Equal(set.features, request.Features) })
	optionDeclared := slices.ContainsFunc(runtimeOptions, func(option runtimeOption) bool { return option.options == request.Options })
	if setDeclared && optionDeclared {
		return nil
	}
	return fmt.Errorf("oracle: no runtime product for features %v with options %+v: add the set to runtimeSets, or the options to "+
		"runtimeOptions (runtimes_test.go), so the tree build builds it ahead", request.Features, request.Options)
}

// runtimeInputs key an archive's product by everything its runtime cache key reads (native.RuntimeRequest's Parts:
// the platform, the compiler and its version, the system headers its compile reads, every flag, which holds the
// features and the options, and every file of the runtime by what it holds), so the two move together.
func runtimeInputs(request native.RuntimeRequest) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "native-runtime",
		Files: []string{"internal/native/runtime", "internal/native/library.go", "internal/native/native.go", "internal/native/target.go"},
		Flags: slices.Clone(request.Parts),
	}
}

// runtimeProducts builds ahead every set's archive with option's options, one archive at a time, each compiling four
// units at once, the CPUs a product's unit is given (a traced build gives each product the reads of its process, so
// two at once would each read the other's).
func runtimeProducts(t *testing.T, name string) {
	t.Helper()
	index := slices.IndexFunc(runtimeOptions, func(option runtimeOption) bool { return option.name == name })
	option := runtimeOptions[index]
	if option.requires != "" && os.Getenv(option.requires) != "1" {
		t.Skipf("set %s=1", option.requires)
	}
	for _, set := range runtimeSets {
		request, err := native.RuntimeRequestFor(option.options, set.features...)
		if err != nil {
			t.Fatal(err)
		}
		buildcache.Product(t, runtimeInputs(request), func(directory string) error { return request.Build(directory, 4) })
	}
}

// TestProduct_NativeRuntimesSanitized builds every set's sanitized archive ahead.
func TestProduct_NativeRuntimesSanitized(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized")
}

// TestProduct_NativeRuntimesRelease builds every set's release archive ahead.
func TestProduct_NativeRuntimesRelease(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "release")
}

// TestProduct_NativeRuntimesCounted builds every set's counted archive ahead.
func TestProduct_NativeRuntimesCounted(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "counted")
}

// TestProduct_NativeRuntimesWASI builds every set's wasm archive ahead, where the WASI families run.
func TestProduct_NativeRuntimesWASI(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "wasm")
}

// Each set in the table is one some program needs: its witness turns on exactly its features, so a set no program
// needs any more fails here rather than staying to be built ahead for nothing. A program needing a set the table lacks
// fails its own build (runtimeDeclared). Mutants: a set's witness turning on another set; a set removed from the table,
// which its programs' builds name.
func TestRuntimeSetsAreEachUsed(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, set := range runtimeSets {
		name := strings.Join(set.features, "+")
		if seen[name] {
			t.Errorf("the set [%s] is in runtimeSets twice", name)
		}
		seen[name] = true
		path, err := filepath.Abs(filepath.Join(repository, set.witness))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Errorf("the witness of [%s], %s: %v", name, set.witness, err)
			continue
		}
		if got := native.Features(native.C(program)); !slices.Equal(got, set.features) {
			t.Errorf("the witness of [%s], %s, turns on %v: name a program that needs exactly the set, or drop the set", name, set.witness, got)
		}
	}
}

// An archive's product key holds every part of its runtime cache key, and the features and options it was asked
// with: each moves the key alone (native's TestRuntimeKeyIncludesEveryInput moves the parts with each input). Mutant:
// a part dropped from runtimeInputs.
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
	key := buildcache.NameKey(root, runtimeInputs(request))
	for index, part := range request.Parts {
		changed := request
		changed.Parts = slices.Clone(request.Parts)
		changed.Parts[index] += "!"
		if buildcache.NameKey(root, runtimeInputs(changed)) == key {
			t.Errorf("changing the part %.60q kept the product key", part)
		}
	}
	// Each feature alone and each option apart from the release build's, against the plain release archive: a set's
	// key is its features' flags, so these move every set's.
	keys := map[string]string{buildcache.NameKey(root, runtimeInputs(plain(t, native.Options{}))): "the plain release archive"}
	asks := map[string]native.RuntimeRequest{}
	for _, feature := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"} {
		asks["the feature "+feature] = plain(t, native.Options{}, feature)
	}
	for _, option := range runtimeOptions {
		if option.name != "release" && (option.requires == "" || os.Getenv("WASI_SYSROOT") != "") {
			asks["the "+option.name+" options"] = plain(t, option.options)
		}
	}
	for name, asked := range asks {
		key := buildcache.NameKey(root, runtimeInputs(asked))
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
