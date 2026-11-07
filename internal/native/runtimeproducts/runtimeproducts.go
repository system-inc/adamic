// Package runtimeproducts is the native runtime's archives as buildcache products, for the tests that build programs
// natively: internal/oracle's, stage3/fixtures' oracle hook, which is internal/oracle's test binary, and
// internal/lower's agreement natives, each handed its archive (Archive).
//
// Every program the oracle builds natively links an archive of the runtime compiled for its features (the leading
// #defines its C turns on) and its options, and an archive is about 56 C files. Compiled where a test first needed it,
// that was 22 s of a 27 s test alone on one core, and 62 to 74 s at a 62-CPU share, where the program's own work is 5
// (#sqmn7jk, lap 3's red): every unit compiled them all, since none was a product. So each archive is a buildcache
// product the tree build builds ahead (Build, from a package's TestProduct_ tests), and native asks a process's
// supplier for it (Supply), which a runner reads and never builds: one it lacks fails the build as not built.
//
// A unit fetches only the products its own package's product tests touch, so each package whose tests read an archive
// has a product test for it: internal/oracle builds every one (TestProduct_NativeRuntimes*, Build, and its program
// region tests' archives, TestProduct_ProgramRuntimes*, ProgramArchive), under oracle.test, the binary its unit and
// stage3/fixtures' hook read them as; stage3/fixtures finds the sanitized ones its hook links
// (TestProduct_FixtureRuntimesSanitized*, Find), after oracle's in the tree's product order, and never builds them,
// since one name key two packages build settles twice (buildcache.BuiltElsewhere). cmd/adamic builds the one archive
// only it reads, its debug build's (TestProduct_CommandRuntimeDebug, CommandArchive). The recipe is here, not in
// oracle's tests, so stage3/fixtures can key what it finds (#43k99f8).
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
	{[]string{"ADAMIC_CLOSURE_RECEIVERS"}, "internal/oracle/testdata/census_optional_values_receivers.a"},
	{[]string{"ADAMIC_REGEXP_REPLACE_CALLBACK"}, "internal/oracle/testdata/regexp_replace/argument_guard.a"},
	{[]string{"ADAMIC_NODE_HOST"}, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_fd_ownership.a"},
	{
		[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES"},
		"internal/oracle/testdata/closure_convention_nested.a",
	},
	{
		[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CLOSURE_RECEIVERS"},
		"internal/oracle/testdata/arguments_length_extended.a",
	},
	{
		[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_REGEXP_REPLACE_CALLBACK"},
		"internal/oracle/testdata/regexp_replace/callback.a",
	},
	{
		[]string{"ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS"},
		"internal/oracle/testdata/getters_census_binary_inline.a",
	},
	{
		[]string{"ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST"},
		"internal/oracle/testdata/closure_convention_host24.a",
	},
	{[]string{"ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_NODE_HOST"}, "stage3/fixtures/host/13_createDirectory.a"},
	{[]string{"ADAMIC_CLOSURE_DISCARDS"}, "internal/oracle/testdata/void_discards_value.a"},
	{
		[]string{"ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_DISCARDS"},
		"internal/oracle/testdata/void_discards_canonical.a",
	},
	{
		[]string{"ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_CLOSURE_DISCARDS"},
		"internal/oracle/testdata/census_optional_values_system.a",
	},
	{
		[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CLOSURE_DISCARDS"},
		"internal/oracle/testdata/program_region/parse.a",
	},
	{[]string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_NODE_HOST"}, "internal/oracle/testdata/node_fs_directory_system.a"},
	{[]string{"ADAMIC_CANONICAL_CLOSURES", "ADAMIC_NODE_HOST"}, "internal/oracle/testdata/host_realpath/closure.a"},
}

// A Variant is one way the oracle builds a program: sanitized for the comparison and the leak check, release as a
// user's build, counted for the counts table and macOS's leak check, wasm for the WASI families, and wasm-request for
// the request reactor's shim (TestWASIShimRequest), counted so it can check each request leaves nothing live; the wasm
// variants run only where Requires is set.
type Variant struct {
	Name     string
	Options  native.Options
	Requires string
}

// oracleWASI is the setting the wasm variants require, read by a constant name so the census of an archive's path can
// hold the read to its reason (TestTheArchivesBuildPathReadsOnlyWhatItsKeyHolds).
const oracleWASI = "ADAMIC_ORACLE_WASI"

var Variants = []Variant{
	{"sanitized", native.Options{Sanitize: true}, ""},
	{"release", native.Options{}, ""},
	{"shipped", native.Options{Release: true}, ""},
	{"counted", native.Options{Count: true}, ""},
	{"wasm", native.Options{Target: "wasm32-wasi"}, oracleWASI},
	{"wasm-request", native.Options{Target: "wasm32-wasi", Request: true, Count: true}, oracleWASI},
}

// ProgramVariants are the ways internal/oracle's program region tests build a program, each with its own archive
// (ProgramArchive): a program region sanitized for the comparison and the leak check, released as a user's build, and
// counted for the counts table, and the counted wasm build the host witness compares its default mode with
// (TestProgramRegionHost14DefaultWASIAgreesWithNode), which a program region can't take (native.ValidateOptions) and
// which runs only where Requires is set. They're a list apart from Variants, since Declared reads Variants: a program
// region's options there would have Supply, which refuses any request Declared doesn't declare, build that archive
// mid-test instead.
var ProgramVariants = []Variant{
	{"program-sanitized", native.Options{ProgramRegion: true, Sanitize: true}, ""},
	{"program-release", native.Options{ProgramRegion: true}, ""},
	{"program-counted", native.Options{ProgramRegion: true, Count: true}, ""},
	{"wasm-counted", native.Options{Target: "wasm32-wasi", Count: true}, oracleWASI},
}

// CommandVariants are the ways cmd/adamic builds a program whose archive no other package builds, each with its own
// archive (CommandArchive): `adamic build --debug`, released with its debug information
// (TestProduct_CommandRuntimeDebug). They're a list apart from Variants, as ProgramVariants are, so Supply still
// refuses them.
var CommandVariants = []Variant{
	{"debug", native.Options{Release: true, Debug: true}, ""},
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
	return buildcache.LeanGet(
		Inputs(request),
		func(directory string) error { return request.Build(directory, 1, jobserver) },
	)
}

// Declared says why a request is for an archive the tree build doesn't build ahead, or nil when it is one.
func Declared(request native.RuntimeRequest) error {
	setDeclared := slices.ContainsFunc(Sets, func(set Set) bool { return slices.Equal(set.Features, request.Features) })
	// Split and Jobs govern generated program units, not runtime compilation.
	options := request.Options
	options.Split, options.Jobs = false, 0
	variantDeclared := slices.ContainsFunc(Variants, func(variant Variant) bool { return variant.Options == options })
	if setDeclared && variantDeclared {
		return nil
	}
	return fmt.Errorf(
		"runtimeproducts: no runtime product for features %v with options %+v: add the set to Sets, or the options to "+
			"Variants (internal/native/runtimeproducts), so the tree build builds it ahead",
		request.Features,
		request.Options,
	)
}

// Inputs key an archive's product by everything its runtime cache key reads, as every machine spells it: the
// platform, the system headers its compile reads, every flag (which holds the features and the options) and every
// file of the runtime by what it holds (native.RuntimeRequest's Portable), and the compiler and archiver by what they
// report (buildcache.Tool), never by where this machine reached them. Workshop's tree build reaches clang through
// ~/.adamic-tools/bin and a runner through $ADAMIC_TOOLS/bin, and a key that named either would never be found by the
// other (#94aa17s's class).
func Inputs(request native.RuntimeRequest) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "native-runtime",
		Files: slices.Clone(runtimeFiles),
		Flags: slices.Clone(request.Portable),
		Toolchain: []string{
			buildcache.Tool(request.Compiler, "--version"),
			buildcache.Tool(request.Archiver, "--version"),
		},
	}
}

// runtimeFiles are what an archive's build reads and runs of internal/native, which its lean set keys as files
// (buildcache.LeanProduct, #earlycut): the runtime's sources, and the files defining all the build reaches from
// native.RuntimeRequest.Build, which TestTheRuntimeBuildReachesOnlyItsDeclaredFiles holds to them. Its lean recipe keys
// the rest of internal/native too, the compiler with it, as it keys every package runtimeproducts imports (buildcache's
// leanRecipeOf, Node's must 1 on #gr5y0yf item 1): a change to the tree outside that import closure leaves an archive's
// key alone.
var runtimeFiles = []string{
	"internal/native/runtime",
	"internal/native/jobserver.go",
	"internal/native/library.go",
	"internal/native/native.go",
	"internal/native/target.go",
}

// Build builds ahead the archives of the sets group owns (every third, from group) with the named variant's options,
// one archive at a time, each compiling four units at once, the CPUs a product's unit is given (a traced build gives
// each product the reads of its process, so two at once would each read the other's). A group is a test of its own, so
// the tree build builds a variant's archives as three units side by side: all eleven sanitized archives in one were 96
// s traced on Workshop (4.8 s each to compile, the rest keying and the trace), well inside a product's 600 s ceiling
// (Loom's KindCeilings) but one long unit on the tree build's path, where a group is about 25 s.
func Build(test *testing.T, name string, group int) {
	test.Helper()
	variant := variantNamed(test, name)
	for index, set := range Sets {
		if index%Groups != group {
			continue
		}
		request, err := native.RuntimeRequestFor(variant.Options, set.Features...)
		if err != nil {
			test.Fatal(err)
		}
		jobserver, err := ccjobserver.Process()
		if err != nil {
			test.Fatal(err)
		}
		buildcache.LeanProduct(
			test,
			Inputs(request),
			func(directory string) error { return request.Build(directory, 4, jobserver) },
		)
	}
}

// Find has the archives Build builds for the named variant and group, for a package whose tests read them and whose
// product test waits for internal/oracle's TestProduct_NativeRuntimes<Variant><group> in the tree's product order:
// under the trace it finds each by name and fails, naming that test, when one isn't there; outside it builds as Build
// does.
func Find(test *testing.T, name string, group int, owner string) {
	test.Helper()
	variant := variantNamed(test, name)
	for index, set := range Sets {
		if index%Groups != group {
			continue
		}
		request, err := native.RuntimeRequestFor(variant.Options, set.Features...)
		if err != nil {
			test.Fatal(err)
		}
		jobserver, err := ccjobserver.Process()
		if err != nil {
			test.Fatal(err)
		}
		what := fmt.Sprintf("the %s runtime archive for %v", variant.Name, set.Features)
		buildcache.BuiltElsewhere(
			test,
			what,
			Inputs(request),
			owner,
			func(directory string) error { return request.Build(directory, 4, jobserver) },
		)
	}
}

// Archive is the directory holding the archive Build builds for the named variant and a program turning on features,
// one of Sets, for a package whose tests hand it to a build of their own (internal/lower's agreement natives): found as
// Find finds it, by name under the trace, failing when it isn't there and naming its owner, the test owners names with
// the set's group after it (internal/oracle TestProduct_NativeRuntimesRelease0), and outside the trace built as Build
// builds it. group is the one group whose owner the caller's product test waits for in
// cloud/fast-gate/product-order.json: a set of another group fails, naming that file, since under the trace its archive
// is found only if its owner happened to run first.
func Archive(test testing.TB, name string, features []string, owners string, group int) string {
	test.Helper()
	variant := variantNamed(test, name)
	index := slices.IndexFunc(Sets, func(set Set) bool { return slices.Equal(set.Features, features) })
	if index < 0 {
		test.Fatalf("runtimeproducts: no runtime product for features %v: add the set to Sets", features)
	}
	if index%Groups != group {
		test.Fatalf(
			"runtimeproducts: the %s runtime archive for %v is built by %s%d, and the caller waits for %s%d alone in "+
				"cloud/fast-gate/product-order.json: add %s%d to the caller's entry there, and ask for that group",
			variant.Name,
			features,
			owners,
			index%Groups,
			owners,
			group,
			owners,
			index%Groups,
		)
	}
	request, err := native.RuntimeRequestFor(variant.Options, Sets[index].Features...)
	if err != nil {
		test.Fatal(err)
	}
	jobserver, err := ccjobserver.Process()
	if err != nil {
		test.Fatal(err)
	}
	what := fmt.Sprintf("the %s runtime archive for %v", variant.Name, features)
	return buildcache.BuiltElsewhere(
		test,
		what,
		Inputs(request),
		fmt.Sprintf("%s%d", owners, index%Groups),
		func(directory string) error { return request.Build(directory, 4, jobserver) },
	)
}

// ProgramArchive is the directory holding the archive a program region's build links, for options, one of
// ProgramVariants, and a program turning on features, one of Sets: internal/oracle's TestProduct_ProgramRuntimes shards
// build each ahead with it, and the builds of its tests that link one ask it the same way, so the two can't key one
// archive apart. It's lean (leanArchive): built in oracle, it was keyed by oracle's whole test binary, its test files
// with it, and every edit of an oracle test rebuilt all of them. One unit at a time, as a test waits on it.
func ProgramArchive(test testing.TB, options native.Options, features []string) string {
	test.Helper()
	return leanArchive(test, "Program", ProgramVariants, "ProgramVariants", options, features, 1)
}

// CommandArchive is the directory holding the archive cmd/adamic's TestProduct_CommandRuntime tests build ahead for
// options, one of CommandVariants, and a program turning on features, one of Sets; cmd/adamic's supplier finds it by
// the same Inputs and never builds it. It's lean (leanArchive): built in cmd/adamic, it was keyed by cmd/adamic's whole
// test binary, the compiler with it, and nearly every compiler change rebuilt it. Four units at once, as Build
// compiles.
func CommandArchive(test testing.TB, options native.Options, features []string) string {
	test.Helper()
	return leanArchive(test, "Command", CommandVariants, "CommandVariants", options, features, 4)
}

// leanArchive builds or finds the archive for options, one of variants, the list its caller declares them in, named
// list, and a program turning on features, one of Sets, compiling workers units at once. It's built here as Build
// builds, so its key is runtimeproducts' import closure (buildcache's leanCalled reads the package calling
// LeanProduct), never the test binary of the package asking for it. Options or a feature set it doesn't declare fail
// the test, by kind, naming the list to add them to, before any request is made, and a variant whose setting isn't set
// skips it.
func leanArchive(
	test testing.TB,
	kind string,
	variants []Variant,
	list string,
	options native.Options,
	features []string,
	workers int,
) string {
	test.Helper()
	if !slices.ContainsFunc(Sets, func(set Set) bool { return slices.Equal(set.Features, features) }) {
		test.Fatalf("runtimeproducts: undeclared %s runtime feature set: %v: add the set to Sets", kind, features)
	}
	index := slices.IndexFunc(variants, func(variant Variant) bool { return variant.Options == options })
	if index < 0 {
		test.Fatalf("runtimeproducts: undeclared %s runtime options: %+v: add them to %s", kind, options, list)
	}
	skipUnlessSet(test, variants[index])
	request, err := native.RuntimeRequestFor(options, features...)
	if err != nil {
		test.Fatal(err)
	}
	jobserver, err := ccjobserver.Process()
	if err != nil {
		test.Fatal(err)
	}
	return buildcache.LeanProduct(
		test,
		Inputs(request),
		func(directory string) error { return request.Build(directory, workers, jobserver) },
	)
}

// variantNamed is the variant named name, skipping the test when the variant requires a setting it doesn't have
// (skipUnlessSet).
func variantNamed(test testing.TB, name string) Variant {
	test.Helper()
	index := slices.IndexFunc(Variants, func(variant Variant) bool { return variant.Name == name })
	if index < 0 {
		test.Fatalf("no runtime variant %q", name)
	}
	variant := Variants[index]
	skipUnlessSet(test, variant)
	return variant
}

// skipUnlessSet skips the test when variant requires a setting it doesn't have: the one setting a variant requires is
// oracleWASI, and a variant naming another is skipped until it's read here by name.
func skipUnlessSet(test testing.TB, variant Variant) {
	test.Helper()
	if variant.Requires != "" && (variant.Requires != oracleWASI || os.Getenv(oracleWASI) != "1") {
		test.Skipf("set %s=1", variant.Requires)
	}
}
