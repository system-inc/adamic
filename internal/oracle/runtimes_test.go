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
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
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

// Source-profile probes can witness a feature set without a persistent admitted fixture.
// Each witness still lowers a real program; the feature comparison below is unchanged.
var runtimeWitnesses = map[string]func(*testing.T) *ir.Program{}

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
	if suppliesRuntimes(os.Getenv("ADAMIC_STAGE3_FIXTURE")) {
		native.SupplyRuntimes(supplyRuntime)
	}
}

// suppliesRuntimes says whether this process takes its runtime archives from the products, given its
// ADAMIC_STAGE3_FIXTURE. A process that has one is stage3/fixtures' oracle hook, this package's test binary run as
// stage3/fixtures' product, once per fixture in stage3/fixtures' unit, which fetches stage3/fixtures' products and not
// this package's: there the supplier read every archive as not built (lap A.5, a086f7e7f7e4). The hook compiles them
// into native's runtime cache under its unit's build cache instead, once per unit, as it did before they were products.
func suppliesRuntimes(stage3Fixture string) bool {
	return stage3Fixture == ""
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

// runtimeInputs key an archive's product by everything its runtime cache key reads, as every machine spells it: the
// platform, the system headers its compile reads, every flag (which holds the features and the options) and every
// file of the runtime by what it holds (native.RuntimeRequest's Portable), and the compiler and archiver by what they
// report (buildcache.Tool), never by where this machine reached them. Workshop's tree build reaches clang through
// ~/.adamic-tools/bin and a runner through $ADAMIC_TOOLS/bin, and a key that named either would never be found by the
// other (#94aa17s's class).
func runtimeInputs(request native.RuntimeRequest) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "native-runtime",
		Files:     []string{"internal/native/runtime", "internal/native/library.go", "internal/native/native.go", "internal/native/target.go"},
		Flags:     slices.Clone(request.Portable),
		Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")},
	}
}

// runtimeProducts builds ahead the archives of the sets group owns (every third, from group) with option's options,
// one archive at a time, each compiling four units at once, the CPUs a product's unit is given (a traced build gives
// each product the reads of its process, so two at once would each read the other's). A group is a test of its own, so
// the tree build builds an option's archives as three units side by side: all eleven sanitized archives in one were 96
// s traced on Workshop (4.8 s each to compile, the rest keying and the trace), well inside a product's 600 s ceiling
// (Loom's KindCeilings) but one long unit on the tree build's path, where a group is about 25 s.
func runtimeProducts(t *testing.T, name string, group int) {
	t.Helper()
	index := slices.IndexFunc(runtimeOptions, func(option runtimeOption) bool { return option.name == name })
	option := runtimeOptions[index]
	if option.requires != "" && os.Getenv(option.requires) != "1" {
		t.Skipf("set %s=1", option.requires)
	}
	for index, set := range runtimeSets {
		if index%runtimeProductGroups != group {
			continue
		}
		request, err := native.RuntimeRequestFor(option.options, set.features...)
		if err != nil {
			t.Fatal(err)
		}
		buildcache.Product(t, runtimeInputs(request), func(directory string) error { return request.Build(directory, 4) })
	}
}

// runtimeProductGroups is how many product tests each option's archives are cut into.
const runtimeProductGroups = 3

func TestProduct_NativeRuntimesSanitized0(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized", 0)
}
func TestProduct_NativeRuntimesSanitized1(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized", 1)
}
func TestProduct_NativeRuntimesSanitized2(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized", 2)
}
func TestProduct_NativeRuntimesRelease0(t *testing.T) { t.Parallel(); runtimeProducts(t, "release", 0) }
func TestProduct_NativeRuntimesRelease1(t *testing.T) { t.Parallel(); runtimeProducts(t, "release", 1) }
func TestProduct_NativeRuntimesRelease2(t *testing.T) { t.Parallel(); runtimeProducts(t, "release", 2) }
func TestProduct_NativeRuntimesCounted0(t *testing.T) { t.Parallel(); runtimeProducts(t, "counted", 0) }
func TestProduct_NativeRuntimesCounted1(t *testing.T) { t.Parallel(); runtimeProducts(t, "counted", 1) }
func TestProduct_NativeRuntimesCounted2(t *testing.T) { t.Parallel(); runtimeProducts(t, "counted", 2) }
func TestProduct_NativeRuntimesWASI0(t *testing.T)    { t.Parallel(); runtimeProducts(t, "wasm", 0) }
func TestProduct_NativeRuntimesWASI1(t *testing.T)    { t.Parallel(); runtimeProducts(t, "wasm", 1) }
func TestProduct_NativeRuntimesWASI2(t *testing.T)    { t.Parallel(); runtimeProducts(t, "wasm", 2) }

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
		var program *ir.Program
		if witness := runtimeWitnesses[set.witness]; witness != nil {
			program = witness(t)
		} else {
			path, err := filepath.Abs(filepath.Join(repository, set.witness))
			if err != nil {
				t.Fatal(err)
			}
			program, err = lowered(t, path)
			if err != nil {
				t.Errorf("the witness of [%s], %s: %v", name, set.witness, err)
				continue
			}
		}
		if got := native.Features(native.C(program)); !slices.Equal(got, set.features) {
			t.Errorf("the witness of [%s], %s, turns on %v: name a program that needs exactly the set, or drop the set", name, set.witness, got)
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
	key := buildcache.NameKey(root, runtimeInputs(request))
	for index, part := range request.Portable {
		changed := request
		changed.Portable = slices.Clone(request.Portable)
		changed.Portable[index] += "!"
		if buildcache.NameKey(root, runtimeInputs(changed)) == key {
			t.Errorf("changing the part %.60q kept the product key", part)
		}
	}
	// The parts the portable ones leave out are keyed by their tools' reports.
	inputs := runtimeInputs(request)
	if !slices.Equal(inputs.Toolchain, []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")}) {
		t.Errorf("the product isn't keyed by the compiler's and archiver's reports: %q", inputs.Toolchain)
	}
	if len(request.Portable) != len(request.Parts)-3 {
		t.Errorf("the portable parts are %d of the key's %d, want all but the compiler, its version and the archiver", len(request.Portable), len(request.Parts))
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
			fmt.Printf("compiler %s\nkey %v %s\n", request.Compiler, features, buildcache.NameKey(root, runtimeInputs(request)))
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

// The oracle's own tests take their runtime archives from the products, and stage3/fixtures' oracle hook (this
// binary, run with ADAMIC_STAGE3_FIXTURE) doesn't, since its unit holds none of this package's products. Mutants: the
// supplier installed in the hook; the supplier never installed.
func TestTheStage3HookCompilesItsOwnRuntimes(t *testing.T) {
	t.Parallel()
	if !suppliesRuntimes("") {
		t.Error("the oracle's own tests don't take their runtime archives from the products")
	}
	if suppliesRuntimes("stage3/fixtures/real/program.ts") {
		t.Error("stage3/fixtures' oracle hook takes its runtime archives from products its unit doesn't hold")
	}
}
