package native

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

var primitiveCases = []string{
	"version-template", "version-add", "declaration-array", "valueOf-first",
	"string-fallback", "objects-throw", "noncallable-skipped", "absent-throw",
	"call-throws", "get-throws", "lookup-after-mutation", "undefined-result",
	"null-result", "boolean-result", "negative-zero-result", "string-result", "number-hint",
	"function-results-throw", "getter-mutates-next", "second-get-throws",
}

func primitiveRun(binary string, index int, environment ...string) primitiveExecution {
	command := exec.Command(binary, fmt.Sprint(index))
	command.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	code := 0
	if runError != nil {
		code = -1
		if command.ProcessState != nil {
			code = command.ProcessState.ExitCode()
		}
	}
	return primitiveExecution{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: code}
}

func primitiveOracle(test *testing.T, index int) []byte {
	test.Helper()
	got, combinedOutputError := exec.Command("node", "testdata/to-primitive/node.a", fmt.Sprint(index)).CombinedOutput()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, got)
	}
	if index == 5 || index == 7 || index == 17 {
		if !bytes.HasPrefix(got, []byte("TypeError: Cannot convert object to primitive value\n")) {
			test.Fatalf("Node's TypeError changed: %s", got)
		}
	}
	return got
}

func TestOrdinaryToPrimitive(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/to-primitive/runtime.c")
	if readError != nil {
		test.Fatal(readError)
	}
	for _, slabs := range []bool{false, true} {
		options := Options{Sanitize: true, slabs: slabs, Count: true}
		binary := filepath.Join(test.TempDir(), "primitive")
		if buildError := Build(string(source), binary, options); buildError != nil {
			test.Fatal(buildError)
		}
		test.Logf("clang flags: %s", strings.Join(Flags(options), " "))
		for index, name := range primitiveCases {
			test.Run(fmt.Sprintf("slabs-%t/%s", slabs, name), func(test *testing.T) {
				want := primitiveOracle(test, index)
				got := primitiveRun(binary, index, "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				if got.ExitCode != 0 || !bytes.Equal(got.Stdout, want) {
					test.Fatalf(
						"C differs from Node: exit %d\n%s%s\nNode: %s",
						got.ExitCode,
						got.Stdout,
						got.Stderr,
						want,
					)
				}
				if report := primitiveUnbalanced(got); report != "" {
					test.Fatal(report)
				}
				if report := primitiveLeaks(test, binary, fmt.Sprint(index)); report != "" {
					test.Fatal(report)
				}
			})
		}
	}
}

// primitiveMutation is one mutant's change to the runtime's exceptions.c: before, which must occur exactly once,
// becomes after. No mutant modifies the checkout or fails compilation.
type primitiveMutation struct{ name, before, after string }

// primitiveMutations is every mutant the to-primitive tests run, TestOrdinaryToPrimitiveMutants' three and
// TestOrdinaryToPrimitiveProtocolGuards' two. A test reaches its binary by the mutant's name.
var primitiveMutations = []primitiveMutation{
	{"wrong-order", "if (hint == adamic_hint_string)", "if (hint != adamic_hint_string)"},
	{
		"missing-TypeError",
		"adamic_thrown = adamic_error_new_kind(name, sizeof name - 1, text);",
		"adamic_release(adamic_error_new_kind(name, sizeof name - 1, text));",
	},
	{"leaked-object-result", "adamic_release(result.value.reference);", "adamic_release(NULL);"},
	{
		"terminal-invalid-hint",
		"if (hint != adamic_hint_default && hint != adamic_hint_number && hint != adamic_hint_string)",
		"if (false)",
	},
	{"preserve-pending-error", "if (adamic_thrown != NULL) { return empty; }\n\tif (hint", "if (hint"},
}

// primitiveMutant is testdata/to-primitive/runtime.c linked against the named mutant's runtime, from the
// primitive-mutants product.
func primitiveMutant(test *testing.T, name string) string {
	test.Helper()
	if !slices.ContainsFunc(
		primitiveMutations,
		func(mutation primitiveMutation) bool { return mutation.name == name },
	) {
		test.Fatalf("no primitive mutant is named %s", name)
	}
	return filepath.Join(primitiveMutants(test), name)
}

// primitiveMutants is the primitive-mutants product, built ahead by TestProduct_PrimitiveMutants: the to-primitive
// harness linked against each mutant's runtime, a binary named for the mutant, with Flags under Sanitize and Count.
// Each mutant used to build a whole sanitized runtime archive inside its test, about 7 core-seconds of clang apiece,
// so TestOrdinaryToPrimitiveMutants took 22 s on one core and 64 s on a loaded box (#knrqkkv). As
// TestProduct_RecordMutants does, the runtime's unchanged translation units compile once, and each mutant compiles
// only its changed exceptions.c and links every object directly.
func primitiveMutants(test *testing.T) string {
	test.Helper()
	flags := Flags(Options{Sanitize: true, Count: true})
	inputs := buildcache.Inputs{
		Name: "primitive-mutants",
		Files: []string{
			"internal/native/runtime",
			"internal/native/testdata/to-primitive/runtime.c",
			"internal/native/to_primitive_test.go",
			"internal/native/split_units_test.go",
			"internal/native/runtime_mutants_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
		},
		Flags:     slices.Clone(flags),
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
	for _, mutation := range primitiveMutations {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf(
			"mutant %s exceptions.c %x %x",
			mutation.name,
			sha256.Sum256([]byte(mutation.before)),
			sha256.Sum256([]byte(mutation.after)),
		))
	}
	harness, readError := os.ReadFile("testdata/to-primitive/runtime.c")
	if readError != nil {
		test.Fatal(readError)
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		files, readRuntimeError := readRuntime(runtime, "runtime")
		if readRuntimeError != nil {
			return readRuntimeError
		}
		var sources []string
		for _, file := range files {
			if strings.HasSuffix(file.name, ".c") {
				sources = append(sources, file.name)
			}
		}
		scratch, temporaryDirectoryError := os.MkdirTemp("", "adamic-primitive-mutants-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(scratch)
		// compile builds a directory's translation unit from inside it, mapped, so no object names the scratch
		// directory (cachedRuntime's way).
		compile := func(within, source string) error {
			arguments := append(
				append(slices.Clone(flags), prefixMaps(within, "/adamic-runtime")...),
				"-c",
				source,
				"-o",
				strings.TrimSuffix(source, ".c")+".o",
			)
			command := runtimeMutantClang(within, arguments...)
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("compiling runtime %s: %w\n%s", source, combinedOutputError, output)
			}
			return nil
		}
		write := func(within string, files []runtimeFile) error {
			if createDirectoryError := os.MkdirAll(within, 0o755); createDirectoryError != nil {
				return createDirectoryError
			}
			for _, file := range files {
				if writeError := os.WriteFile(
					filepath.Join(within, file.name),
					file.contents,
					0o644,
				); writeError != nil {
					return writeError
				}
			}
			return nil
		}
		unchanged := filepath.Join(scratch, "runtime")
		if writeError := write(unchanged, files); writeError != nil {
			return writeError
		}
		if eachInParallelError := eachInParallel(
			len(sources),
			func(index int) error { return compile(unchanged, sources[index]) },
		); eachInParallelError != nil {
			return eachInParallelError
		}
		if writeError := os.WriteFile(filepath.Join(scratch, "main.c"), harness, 0o644); writeError != nil {
			return writeError
		}
		return eachInParallel(len(primitiveMutations), func(index int) error {
			mutation := primitiveMutations[index]
			// The mutant's directory holds the runtime's headers and its changed exceptions.c, so that file compiles
			// as it would among the rest.
			var changed []runtimeFile
			found := false
			for _, file := range files {
				if file.name == "exceptions.c" {
					contents := string(file.contents)
					if strings.Count(contents, mutation.before) != 1 {
						return fmt.Errorf("mutant %s: its anchor is no longer unique in exceptions.c", mutation.name)
					}
					changed = append(
						changed,
						runtimeFile{file.name, []byte(strings.Replace(contents, mutation.before, mutation.after, 1))},
					)
					found = true
				} else if strings.HasSuffix(file.name, ".h") {
					changed = append(changed, file)
				}
			}
			if !found {
				return fmt.Errorf("mutant %s: the runtime has no exceptions.c", mutation.name)
			}
			within := filepath.Join(scratch, "mutant-"+strconv.Itoa(index))
			if writeError := write(within, changed); writeError != nil {
				return writeError
			}
			if compileError := compile(within, "exceptions.c"); compileError != nil {
				return fmt.Errorf("mutant %s: %w", mutation.name, compileError)
			}
			// Linked from the scratch directory by relative names, mapped, so the binary names neither it nor the
			// product.
			arguments := append(
				append(slices.Clone(flags), prefixMaps(scratch, "/adamic-build")...),
				"-I",
				"runtime",
				"-o",
				filepath.Join(directory, mutation.name),
				"main.c",
			)
			for _, source := range sources {
				object := strings.TrimSuffix(source, ".c") + ".o"
				if source == "exceptions.c" {
					arguments = append(arguments, filepath.Join(filepath.Base(within), object))
				} else {
					arguments = append(arguments, filepath.Join("runtime", object))
				}
			}
			arguments = append(arguments, "-lm")
			command := runtimeMutantClang(scratch, arguments...)
			if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
				return fmt.Errorf("mutant %s must compile: %w\n%s", mutation.name, combinedOutputError, output)
			}
			return nil
		})
	})
}

// TestProduct_PrimitiveMutants builds every primitive mutant, so the tree build builds them ahead.
func TestProduct_PrimitiveMutants(test *testing.T) {
	test.Parallel()
	primitiveMutants(test)
}

func TestOrdinaryToPrimitiveMutants(test *testing.T) {
	test.Parallel()
	test.Run("wrong-order", func(test *testing.T) {
		binary := primitiveMutant(test, "wrong-order")
		for index, name := range primitiveCases {
			test.Run(name, func(test *testing.T) {
				got := primitiveRun(binary, index, "ASAN_OPTIONS=detect_leaks=1")
				if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
					test.Fatalf("behavior mutant must finish leak-clean: %+v", got)
				}
				if bytes.Equal(got.Stdout, primitiveOracle(test, index)) {
					test.Fatal("wrong order survived")
				}
				test.Log("wrong-order mutant caught by Node output and lookup/call trace")
			})
		}
	})
	test.Run("missing-TypeError", func(test *testing.T) {
		binary := primitiveMutant(test, "missing-TypeError")
		for _, index := range []int{5, 7, 17} {
			got := primitiveRun(binary, index, "ASAN_OPTIONS=detect_leaks=1")
			if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
				test.Fatalf("behavior mutant must finish leak-clean: %+v", got)
			}
			if bytes.Equal(got.Stdout, primitiveOracle(test, index)) {
				test.Fatal("missing TypeError survived")
			}
			test.Logf("missing-TypeError mutant caught by %s", primitiveCases[index])
		}
	})
	test.Run("leaked-object-result", func(test *testing.T) {
		binary := primitiveMutant(test, "leaked-object-result")
		got := primitiveRun(binary, 5, "ASAN_OPTIONS=detect_leaks=1", "LSAN_OPTIONS=use_stacks=0:use_registers=0")
		if !bytes.Contains(got.Stderr, []byte("LeakSanitizer: detected memory leaks")) {
			test.Fatalf("leaked object result survived: %+v", got)
		}
		test.Log("rejected-result release mutant caught by LeakSanitizer")
	})
}

// These two runtime ABI guards have no source-level conversion counterpart.
func TestOrdinaryToPrimitiveProtocolGuards(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/to-primitive/runtime.c")
	if readError != nil {
		test.Fatal(readError)
	}
	binary := filepath.Join(test.TempDir(), "primitive")
	if buildError := Build(string(source), binary, Options{Sanitize: true, Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	test.Run("terminal-invalid-hint", func(test *testing.T) {
		got := primitiveRun(binary, 20)
		if got.ExitCode != 70 ||
			!bytes.HasPrefix(got.Stderr, []byte("adamic: panic: compiler bug: invalid primitive hint\n")) {
			test.Fatalf("invalid hint must be terminal: %+v", got)
		}
		mutant := primitiveMutant(test, "terminal-invalid-hint")
		got = primitiveRun(mutant, 20, "ASAN_OPTIONS=detect_leaks=1")
		if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
			test.Fatalf("guard mutant must finish cleanly: %+v", got)
		}
		test.Log("missing invalid-hint guard caught by expected terminal panic")
	})
	test.Run("preserve-pending-error", func(test *testing.T) {
		const want = "Error: kept\n\nafter catch\n"
		got := primitiveRun(binary, 21, "ASAN_OPTIONS=detect_leaks=1")
		if got.ExitCode != 0 || string(got.Stdout) != want || primitiveUnbalanced(got) != "" {
			test.Fatalf("pending error must prevent lookup: %+v", got)
		}
		mutant := primitiveMutant(test, "preserve-pending-error")
		got = primitiveRun(mutant, 21, "ASAN_OPTIONS=detect_leaks=1")
		if got.ExitCode != 0 || primitiveUnbalanced(got) != "" {
			test.Fatalf("guard mutant must finish cleanly: %+v", got)
		}
		if string(got.Stdout) == want {
			test.Fatal("pending-error entry guard mutant survived")
		}
		test.Log("missing pending-error guard caught by an unexpected Get in the trace")
	})
}

// Main keeps its leak gates in the oracle package. These C seam witnesses check
// their own counts and use the same OS leak tools without importing that harness.
type primitiveExecution struct {
	Stdout, Stderr []byte
	ExitCode       int
}

var primitiveCounts = regexp.MustCompile(
	`(?m)^adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)\n$`,
)

func primitiveUnbalanced(run primitiveExecution) string {
	match := primitiveCounts.FindSubmatch(run.Stderr)
	if run.ExitCode != 0 || match == nil {
		return fmt.Sprintf("unfinished counted witness: %+v", run)
	}
	allocations, _ := strconv.ParseUint(string(match[1]), 10, 64)
	frees, _ := strconv.ParseUint(string(match[2]), 10, 64)
	regions, _ := strconv.ParseUint(string(match[6]), 10, 64)
	if allocations != frees+regions {
		return fmt.Sprintf("heap values leaked: %d", allocations-frees-regions)
	}
	return ""
}
func primitiveLeaks(test *testing.T, binary string, arguments ...string) string {
	test.Helper()
	if goruntime.GOOS != "darwin" {
		return ""
	} // Linux runs LSan in primitiveRun.
	command := exec.Command("leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		return fmt.Sprintf("leaks: %v %s", combinedOutputError, output)
	}
	return ""
}
