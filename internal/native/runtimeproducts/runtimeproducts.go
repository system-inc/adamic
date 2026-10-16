// Package runtimeproducts is the native runtime's archives as buildcache products, for the tests that build programs
// natively: internal/oracle's, and stage3/fixtures' oracle hook, which is internal/oracle's test binary.
//
// Every program the oracle builds natively links an archive of the runtime compiled for its features (the leading
// #defines its C turns on) and its options, and an archive is about 56 C files. Compiled where a test first needed it,
// that was 22 s of a 27 s test alone on one core, and 62 to 74 s at a 62-CPU share, where the program's own work is 5
// (#sqmn7jk, lap 3's red): every unit compiled them all, since none was a product. So each archive is a buildcache
// product the tree build builds ahead (Build, from a package's TestProduct_ tests), and native asks a process's
// supplier for it (Supply), which a runner reads and never builds: one it lacks fails the build as not built.
//
// A unit fetches only the products its own package's product tests touch, so each package whose tests read an archive
// has a product test for it: internal/oracle builds every one (TestProduct_NativeRuntimes*, Build), under oracle.test,
// the binary its unit and stage3/fixtures' hook read them as; stage3/fixtures finds the sanitized ones its hook links
// (TestProduct_FixtureRuntimesSanitized*, Find), after oracle's in the tree's product order, and never builds them,
// since one name key two packages build settles twice (buildcache.BuiltElsewhere). The recipe is here, not in oracle's
// tests, so stage3/fixtures can key what it finds (#43k99f8).
package runtimeproducts

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
)

// A Set is one set of features some program the oracle builds turns on, in native's order (native.Features), and
// Witness, a program that turns on exactly these, which internal/oracle's TestRuntimeSetsAreEachUsed lowers so a set no
// program needs any more can't stay in the table.
type Set struct {
	Features []string
	Witness  string
}

// Sets are every feature set the oracle's programs need, counted at 91671a25be over every lowered fixture and review
// program (base 1021, canonical closures 46, convention 18, receivers 14, regexp callback 6, node host 5, the rest 1
// to 4), and the programs stage3/fixtures' oracle hook builds (at ed4fc86eef its 82 that compile need seven, six of
// them these and receivers with node host, which only stage3/fixtures/host's 13_createDirectory.a needs). A program
// whose set isn't here fails its build, naming the set to add.
var Sets = []Set{
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
	{[]string{"ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_NODE_HOST"}, "stage3/fixtures/host/13_createDirectory.a"},
}

// A Variant is one way the oracle builds a program: sanitized for the comparison and the leak check, release as a
// user's build, counted for the counts table and macOS's leak check, and wasm for the WASI families, which run only
// where Requires is set.
type Variant struct {
	Name     string
	Options  native.Options
	Requires string
}

var Variants = []Variant{
	{"sanitized", native.Options{Sanitize: true}, ""},
	{"release", native.Options{}, ""},
	{"counted", native.Options{Count: true}, ""},
	{"wasm", native.Options{Target: "wasm32-wasi"}, "ADAMIC_ORACLE_WASI"},
}

// Groups is how many product tests each variant's archives are cut into.
const Groups = 3

// Supply gives native the archive a build asks for from its product: built here the first time on a machine that
// builds (one unit at a time, as a test waits on it), never on a runner, which reads it or fails as not built.
func Supply(request native.RuntimeRequest) (string, error) {
	if err := Declared(request); err != nil {
		return "", err
	}
	// Lean (#mmpt615): an archive's build takes nothing but its request, made from the runtime's files Inputs names, so
	// a test that asks for it mid-run never keys it by what that test read before. Its compiles are bounded by the
	// process's jobserver, handed to the build as a value (#yzwm4rj).
	jobserver, err := ccjobserver.Process()
	if err != nil {
		return "", err
	}
	return buildcache.LeanGet(Inputs(request), func(directory string) error { return request.Build(directory, 1, jobserver) })
}

// Declared says why a request is for an archive the tree build doesn't build ahead, or nil when it is one.
func Declared(request native.RuntimeRequest) error {
	setDeclared := slices.ContainsFunc(Sets, func(set Set) bool { return slices.Equal(set.Features, request.Features) })
	variantDeclared := slices.ContainsFunc(Variants, func(variant Variant) bool { return variant.Options == request.Options })
	if setDeclared && variantDeclared {
		return nil
	}
	return fmt.Errorf("runtimeproducts: no runtime product for features %v with options %+v: add the set to Sets, or the options to "+
		"Variants (internal/native/runtimeproducts), so the tree build builds it ahead", request.Features, request.Options)
}

// Inputs key an archive's product by everything its runtime cache key reads, as every machine spells it: the
// platform, the system headers its compile reads, every flag (which holds the features and the options) and every
// file of the runtime by what it holds (native.RuntimeRequest's Portable), and the compiler and archiver by what they
// report (buildcache.Tool), never by where this machine reached them. Workshop's tree build reaches clang through
// ~/.adamic-tools/bin and a runner through $ADAMIC_TOOLS/bin, and a key that named either would never be found by the
// other (#94aa17s's class).
func Inputs(request native.RuntimeRequest) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "native-runtime",
		Files:     []string{"internal/native/runtime", "internal/native/library.go", "internal/native/native.go", "internal/native/target.go"},
		Flags:     slices.Clone(request.Portable),
		Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")},
	}
}

// Build builds ahead the archives of the sets group owns (every third, from group) with the named variant's options,
// one archive at a time, each compiling four units at once, the CPUs a product's unit is given (a traced build gives
// each product the reads of its process, so two at once would each read the other's). A group is a test of its own, so
// the tree build builds a variant's archives as three units side by side: all eleven sanitized archives in one were 96
// s traced on Workshop (4.8 s each to compile, the rest keying and the trace), well inside a product's 600 s ceiling
// (Loom's KindCeilings) but one long unit on the tree build's path, where a group is about 25 s.
func Build(t *testing.T, name string, group int) {
	t.Helper()
	variant := variantNamed(t, name)
	for index, set := range Sets {
		if index%Groups != group {
			continue
		}
		request, err := native.RuntimeRequestFor(variant.Options, set.Features...)
		if err != nil {
			t.Fatal(err)
		}
		jobserver, err := ccjobserver.Process()
		if err != nil {
			t.Fatal(err)
		}
		buildcache.LeanProduct(t, Inputs(request), func(directory string) error { return request.Build(directory, 4, jobserver) })
	}
}

// Find has the archives Build builds for the named variant and group, for a package whose tests read them and whose
// product test waits for internal/oracle's TestProduct_NativeRuntimes<Variant><group> in the tree's product order: under
// the trace it finds each by name and fails, naming that test, when one isn't there; outside it builds as Build does.
func Find(t *testing.T, name string, group int, owner string) {
	t.Helper()
	variant := variantNamed(t, name)
	for index, set := range Sets {
		if index%Groups != group {
			continue
		}
		request, err := native.RuntimeRequestFor(variant.Options, set.Features...)
		if err != nil {
			t.Fatal(err)
		}
		jobserver, err := ccjobserver.Process()
		if err != nil {
			t.Fatal(err)
		}
		what := fmt.Sprintf("the %s runtime archive for %v", variant.Name, set.Features)
		buildcache.BuiltElsewhere(t, what, Inputs(request), owner, func(directory string) error { return request.Build(directory, 4, jobserver) })
	}
}

// variantNamed is the variant named name, skipping the test when the variant requires a setting it doesn't have.
func variantNamed(t *testing.T, name string) Variant {
	t.Helper()
	index := slices.IndexFunc(Variants, func(variant Variant) bool { return variant.Name == name })
	if index < 0 {
		t.Fatalf("no runtime variant %q", name)
	}
	variant := Variants[index]
	if variant.Requires != "" && os.Getenv(variant.Requires) != "1" {
		t.Skipf("set %s=1", variant.Requires)
	}
	return variant
}
