package native

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/toolchain"
)

// splitCorpus is the oracle's programs a split build is held to the one call's bytes over: of each family of them, by
// the first word of its name, the first that builds. Async, closures, the node host, regions and the runtime's features
// are among them. None stage 0 lowers links the regexp compiler's archive, so splitRegexpCompiler joins them.
var splitCorpus = []string{
	"async_catch_return.a", "library_array_copy.a", "graph_regions_accessor.a", "node_buffer_bom.a",
	"method_closures.a", "json_decode_coverage_arrays.a", "environment_array.a",
	"borrow_chain_argument.a", "records_census.a", "class_as_interface.a", "nested_array.a", "map_foreach_arrays.a",
	"namespaces.a", "string_append.a", "getters_census_binary_function.a", "census_append_overload.a", "regexp.a",
	"typed_arrays_float64array.a", "optional_after_call_193.a", "host_array_predicate.a", "scout_boxed_scalar_fields.a",
	"cast_class.a", "reuse.a", "enums.a", "search_from.a", "coverage_plain_view.a", "conditions_array.a", "arguments.a",
	"set_foreach_arrays.a", "overload_ancestor_directory.a", "entries_checked_boolean.a", "normalize.a",
	"nbody_base_writes.a", "process_bad_code.a", "label_block_break_async.a", "devirtualize.a",
	"route_targets_accessor.a", "generic_arrow_object_shapes.a", "array_from.a", "weak_narrowed.a",
}

// splitRegexpCompiler is C that compiles a RegExp at run time, so its build links the regexp compiler's archive
// (regexp_constructor_archive_test.go's split source, shortened).
const splitRegexpCompiler = `#define ADAMIC_REGEXP_RUNTIME_COMPILER 1
#include "adamic.h"
#include <stdio.h>
#include <string.h>
int main(int argc, char **argv) {
	adamic_start(argc, argv);
	adamic_string *pattern = adamic_string_allocate(3), *flags = adamic_string_allocate(2);
	memcpy((char *)pattern->bytes, "a/b", 3);
	memcpy((char *)flags->bytes, "ig", 2);
	adamic_object *compiled = adamic_regex_compile_new(pattern, flags);
	adamic_release(pattern);
	adamic_release(flags);
	if (compiled == NULL) {
		adamic_uncaught();
	}
	adamic_release(compiled);
	return 0;
}
`

// splitBuilds are the builds each program of the corpus is made as: a release, a shipped release (ThinLTO, whose code
// generation runs in the link), Debug (absolute inputs, the line tables), and both sanitized builds, whose debug info and
// check metadata the binary carries.
var splitBuilds = []struct {
	name    string
	options Options
}{
	{"release", Options{}},
	{"shipped release", Options{Release: true}},
	{"debug", Options{Debug: true}},
	{"sanitized", Options{Sanitize: true}},
	{"thread sanitized", Options{ThreadSanitize: true}},
}

// splitBuilt is what one build made: the binary's bytes, or the error it failed with and clang's exit code.
type splitBuilt struct {
	binary   []byte
	failure  string
	exitCode int
}

// buildOneCallAndSplit builds source as options say twice, as one clang call that compiles and links (the build off
// macOS before #2r9xhqs, BuildLean's still) and as a compile and then a link, in the same build directory and to the same
// output, so the two differ by the split alone. Each is linkOne as Build reaches it, with Build's flags and its runtime.
func buildOneCallAndSplit(t *testing.T, source string, options Options) (splitBuilt, splitBuilt) {
	t.Helper()
	library, err := RuntimeLibraryForSource("", source, options)
	if err != nil {
		t.Fatal(err)
	}
	jobserver, err := defaultJobserver()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	directory, output := filepath.Join(root, "build"), filepath.Join(root, "program")
	build := func(oneCall bool) splitBuilt {
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(directory)
		defer os.Remove(output)
		err := linkOne(directory, compilerName(options), source, output, options, Flags(options), library, jobserver, func(directory string) ([]string, error) {
			return regexpCompilerLinkFlags(source, options, directory)
		}, oneCall)
		return splitResult(t, output, err)
	}
	return build(true), build(false)
}

// splitResult reads what a build at output made, or the failure it returned.
func splitResult(t *testing.T, output string, err error) splitBuilt {
	t.Helper()
	if err != nil {
		built := splitBuilt{failure: err.Error(), exitCode: -1}
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			built.exitCode = exit.ExitCode()
		}
		return built
	}
	binary, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return splitBuilt{binary: binary}
}

// splitDifference says how the split's build differs from the one call's, or "" when they are the same: the same bytes,
// or the same failure with the same exit code.
func splitDifference(oneCall, split splitBuilt) string {
	if oneCall.failure != split.failure || oneCall.exitCode != split.exitCode {
		return fmt.Sprintf("the one call failed with exit code %d and %q, the split with exit code %d and %q", oneCall.exitCode, oneCall.failure, split.exitCode, split.failure)
	}
	if bytes.Equal(oneCall.binary, split.binary) {
		return ""
	}
	first, differing := -1, 0
	for offset := range min(len(oneCall.binary), len(split.binary)) {
		if oneCall.binary[offset] != split.binary[offset] {
			differing++
			if first < 0 {
				first = offset
			}
		}
	}
	return fmt.Sprintf("the binaries differ: %d and %d bytes, %d of the shared ones differing, the first at byte %d", len(oneCall.binary), len(split.binary), differing, first)
}

// splitCorpusSource is the C of the oracle program named name.
func splitCorpusSource(t *testing.T, name string) string {
	t.Helper()
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", name)})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	return C(program)
}

// Compiling main.c to main.o and then linking it makes the one call's binary, byte for byte (#2r9xhqs): the split exists
// so the compile's object can be cached, which it can only be if the split itself changes nothing. Over the oracle's
// programs, one of each family, each as a release, a shipped release, Debug, and both sanitized builds; the whole binary
// equal covers the Debug builds' line and file tables. Mutants, each red: a compile flag left off the compile (the
// prefix maps, so the build directory reaches the debug info), and the compile's flags left off the link.
func TestACompiledThenLinkedProgramIsTheOneCallProgram(t *testing.T) {
	t.Parallel()
	names := append(slices.Clone(splitCorpus), "regexp compiler")
	sources := make([]string, len(names))
	var regexpCompiler, features int
	for index, name := range names {
		sources[index] = splitRegexpCompiler
		if index < len(splitCorpus) {
			sources[index] = splitCorpusSource(t, name)
		}
		if regexpCompilerSource(sources[index]) {
			regexpCompiler++
		}
		if len(Features(sources[index])) > 0 {
			features++
		}
	}
	// The corpus is held to what it says it reaches, so a fixture renamed or rewritten can't quietly narrow it.
	if regexpCompiler == 0 || features == 0 {
		t.Fatalf("of %d programs, %d link the regexp compiler and %d turn on a runtime feature; the corpus needs both", len(names), regexpCompiler, features)
	}
	for index, name := range names {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for _, build := range splitBuilds {
				oneCall, split := buildOneCallAndSplit(t, sources[index], build.options)
				if oneCall.failure != "" {
					t.Errorf("%s: the one call failed: %s", build.name, oneCall.failure)
				}
				if difference := splitDifference(oneCall, split); difference != "" {
					t.Errorf("%s: %s", build.name, difference)
				}
			}
		})
	}
}

// splitFaults are C a program's source ends with that fails its build: under -Werror (a variable used before it may be
// set, from -Wall; a zero-length array, from -pedantic), as plain C (a name nothing declares), and at the link (a
// function nothing defines, called from one kept as used, so ThinLTO can't drop the call as dead).
var splitFaults = []struct {
	name string
	c    string
}{
	{"maybe unset", "int adamic_split_unset(int value) { int unset; if (value) { unset = 1; } return unset; }\n"},
	{"pedantic", "int adamic_split_pedantic[0];\n"},
	{"undeclared", "int adamic_split_undeclared(void) { return adamic_split_nowhere; }\n"},
	{"undefined", "int adamic_split_missing(void);\n__attribute__((used)) int adamic_split_link(void) { return adamic_split_missing(); }\n"},
}

// oneCallObject is the object a one-call build compiles main.c to and links, a temporary file of clang's, which a link's
// error names by its path, or under ThinLTO as the link's output's prefix and the object's name
// under the temporary directory (<TMPDIR>/program.lto.main-<random>.o, where the split's reads program.lto.main.o).
var oneCallObject = regexp.MustCompile(`(?:\S*/)?([^\s/]*)main-[0-9A-Za-z]+\.o\b`)

// A build that fails fails the same way split as in one call: the same error, clang's diagnostics in it, and the same
// exit code, under -Werror, as plain C and at the link (Compiler's condition 3 on #2r9xhqs: a build that fails in one
// call and passes split would be a wrong pass). The one place they may differ is the object a link's error names:
// clang's temporary <TMPDIR>/main-<random>.o in one call, random per build, and main.o split, so that name is read as
// main.o in both. Mutant, red: the compile's -Werror left off (the split compile passes what the one call refuses).
func TestACompiledThenLinkedProgramFailsAsTheOneCallFails(t *testing.T) {
	t.Parallel()
	source := splitCorpusSource(t, "arguments.a")
	for _, fault := range splitFaults {
		t.Run(fault.name, func(t *testing.T) {
			t.Parallel()
			for _, build := range splitBuilds {
				oneCall, split := buildOneCallAndSplit(t, source+fault.c, build.options)
				if oneCall.failure == "" || oneCall.exitCode <= 0 {
					t.Fatalf("%s: the one call didn't fail with an exit code: %d %q", build.name, oneCall.exitCode, oneCall.failure)
				}
				oneCall.failure = oneCallObject.ReplaceAllString(oneCall.failure, "${1}main.o")
				if difference := splitDifference(oneCall, split); difference != "" {
					t.Errorf("%s: %s", build.name, difference)
				}
			}
		})
	}
}

// A checker build (BuildTSGo) compiled to main.o and then linked is the one call's binary too, release, shipped
// release, Debug and sanitized, linking a checker archive of the ABI's stubs. Mutant, red: the compile's flags left off
// the checker's link.
func TestACompiledThenLinkedCheckerIsTheOneCallChecker(t *testing.T) {
	t.Parallel()
	source := "#include \"tsgo_runtime.h\"\n" + splitCorpusSource(t, "arguments_length_extended.a")
	directory := t.TempDir()
	var stubs strings.Builder
	seen := map[string]bool{}
	for _, match := range regexp.MustCompile(`\b(tsgo_[a-z_]+)\(`).FindAllStringSubmatch(string(bridge.Header), -1) {
		if !seen[match[1]] {
			seen[match[1]] = true
			stubs.WriteString("void " + match[1] + "(void) {}\n")
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "stubs.c"), []byte(stubs.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	if output, err := toolchain.CommandIn(directory, compiler, "-g", "-c", "stubs.c", "-o", "stubs.o").CombinedOutput(); err != nil {
		t.Fatalf("stubs: %v\n%s", err, output)
	}
	if output, err := toolchain.CommandIn(directory, archiverName(compiler), "rcs", "checker.a", "stubs.o").CombinedOutput(); err != nil {
		t.Fatalf("archive: %v\n%s", err, output)
	}
	output := filepath.Join(directory, "checker")
	for _, build := range splitBuilds[:4] {
		var built [2]splitBuilt
		for index, oneCall := range []bool{true, false} {
			built[index] = splitResult(t, output, buildTSGo(source, output, filepath.Join(directory, "checker.a"), build.options, oneCall))
			os.Remove(output)
		}
		if built[0].failure != "" {
			t.Errorf("%s: the one call failed: %s", build.name, built[0].failure)
		}
		if difference := splitDifference(built[0], built[1]); difference != "" {
			t.Errorf("%s: %s", build.name, difference)
		}
	}
}
