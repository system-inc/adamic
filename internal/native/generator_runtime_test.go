package native

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

// generatorLeaked reads the counts line a counted build writes last to stderr (count.c): empty when
// every allocation was freed. leakcheck.Unbalanced says the same, but leakcheck imports native.
func generatorLeaked(stderr string) string {
	lines := strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
	var allocations, frees, retains, releases, peak, regions int
	if _, scanError := fmt.Sscanf(
		lines[len(lines)-1],
		"adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d",
		&allocations,
		&frees,
		&retains,
		&releases,
		&peak,
		&regions,
	); scanError != nil {
		return "no counts line: " + scanError.Error()
	}
	if allocations != frees+regions {
		return fmt.Sprintf(
			"heap values leaked: %d (allocations %d, frees %d)",
			allocations-frees-regions,
			allocations,
			frees,
		)
	}
	return ""
}

// generatorRuntimeGuard runs one mode of the hand-written generator harness, a stand-in for what the
// emitter will write, against the runtime with no compiler lowering. The build is both sanitized and
// counted, so a use after free fails under ASan and a leak fails the counts on every platform, Darwin
// included, where LeakSanitizer doesn't run. A non-empty old makes a second run with that exact text
// of generator.c replaced, which must fail with diagnostic.
func generatorRuntimeGuard(test *testing.T, mode, want, old, replacement, diagnostic string) {
	test.Helper()
	files, flags, compiler, version, runtimeInputsError := runtimeInputs(
		"",
		Options{Sanitize: true, Count: true},
		[]string{"-DADAMIC_CANONICAL_CLOSURES=1"},
	)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	cache, cacheDirectoryError := cacheDirectory("generator-runtime")
	if cacheDirectoryError != nil {
		test.Fatal(cacheDirectoryError)
	}
	source, readError := os.ReadFile("testdata/generator/runtime.c")
	if readError != nil {
		test.Fatal(readError)
	}
	mutants := []bool{false}
	if old != "" {
		mutants = append(mutants, true)
	}
	for _, mutant := range mutants {
		if mutant {
			found := false
			for index := range files {
				if files[index].name != "generator.c" {
					continue
				}
				text := string(files[index].contents)
				if strings.Count(text, old) != 1 {
					test.Fatalf("mutant seam %q appears %d times", old, strings.Count(text, old))
				}
				files[index].contents = []byte(strings.Replace(text, old, replacement, 1))
				found = true
			}
			if !found {
				test.Fatal("missing runtime file generator.c")
			}
		}
		library, cachedRuntimeError := cachedRuntime(files, flags, compiler, version, cache)
		if cachedRuntimeError != nil {
			test.Fatal(cachedRuntimeError)
		}
		binary := buildHarnessWithLibrary(test, string(source), flags, compiler, library)
		command, release := boundedrun.Command(5*time.Second, binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1")
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		runError := command.Run()
		release()
		exit := 0
		if command.ProcessState != nil {
			exit = command.ProcessState.ExitCode()
		}
		output := stdout.String() + stderr.String()
		if mutant {
			leaked := generatorLeaked(stderr.String())
			if (runError == nil && leaked == "") || !strings.Contains(output+leaked, diagnostic) {
				test.Fatalf("mutant survived or failed for another reason: %v %s\n%s", runError, leaked, output)
			}
			test.Logf("%s mutant caught by %s: %s", mode, diagnostic, firstSanitizerLine(output))
			continue
		}
		if panic, panics := runningPanics[mode]; panics {
			if exit != 70 || !strings.Contains(stderr.String(), panic) {
				test.Fatalf("%s: %v\n%s", mode, runError, output)
			}
			continue
		}
		if runError != nil || stdout.String() != want {
			test.Fatalf("control: %v\nwant stdout %q\n%s", runError, want, output)
		}
		if leaked := generatorLeaked(stderr.String()); leaked != "" {
			test.Fatalf("control: %s\n%s", leaked, output)
		}
	}
}

func TestGeneratorRuntimeComplete(test *testing.T) {
	test.Parallel()
	// finish must NULL what it released, or the drop after completion releases it again.
	generatorRuntimeGuard(
		test,
		"complete",
		"1\n3\n2\n",
		"cell->value.reference = NULL;\n            adamic_release(owned);",
		"adamic_release(owned);",
		"heap-use-after-free",
	)
}

func TestGeneratorRuntimeDropStart(test *testing.T) {
	test.Parallel()
	// A generator never started still owns its parameters.
	generatorRuntimeGuard(
		test,
		"drop_start",
		"",
		"drop(generator->cells[index].value.reference);",
		"(void)drop;",
		"heap values leaked",
	)
}

func TestGeneratorRuntimeDropSuspended(test *testing.T) {
	test.Parallel()
	// Dropped at a yield, the walk frees every live cell.
	generatorRuntimeGuard(
		test,
		"drop_suspended",
		"1\n",
		"drop(generator->cells[index].value.reference);",
		"(void)drop;",
		"heap values leaked",
	)
}

func TestGeneratorRuntimeReturn(test *testing.T) {
	test.Parallel()
	// The finish return runs, retained around it since q3, taken out whole.
	generatorRuntimeGuard(
		test,
		"return",
		"1\n",
		"adamic_retain(generator);\n            adamic_generator_finish(generator);\n            adamic_release(generator);\n            break;",
		"break;",
		"Assertion",
	)
}

func TestGeneratorRuntimeKill(test *testing.T) {
	test.Parallel()
	// A killed cell left pointing at what it released is released again by the drop walk.
	generatorRuntimeGuard(
		test,
		"kill",
		"1\n",
		"cell->value.reference = NULL;\n        adamic_release(held);",
		"adamic_release(held);",
		"heap-use-after-free",
	)
}

func TestGeneratorRuntimeBorrowed(test *testing.T) {
	test.Parallel()
	// A borrowed alias is never the generator's to drop.
	generatorRuntimeGuard(
		test,
		"borrowed",
		"7\n7\n",
		"if (generator->cells[index].references && generator->cells[index].value.reference != NULL) {",
		"if (generator->cells[index].value.reference != NULL) {",
		"heap-use-after-free",
	)
}

func TestGeneratorRuntimeThrow(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(test, "throw", "5\nError: 6\n", "", "", "")
}

// finish owns an error still set when it runs. Mutant: finish forgets it, and the count leaks.
func TestGeneratorRuntimeFinishError(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(
		test,
		"finish_error",
		"",
		"generator->error = NULL;\n        adamic_release(error);",
		"generator->error = NULL;\n        (void)error;",
		"heap values leaked",
	)
}

// The drop walk drops an error it finds. Mutant: the walk skips error, and the count leaks.
func TestGeneratorRuntimeDropError(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(
		test,
		"drop_error",
		"1\n",
		"if (generator->error != NULL) { drop(generator->error); }",
		"(void)0;",
		"heap values leaked",
	)
}

// A nested generator's carrier, the closure its entry retains into self (Memory's ruling on #qk8rztp
// S3a): finish lets go of it and leaves self `NULL`, so a done generator stops pinning its parent's
// cells, and the harness asserts it `NULL` after the last step. Mutant: finish forgets self's release,
// and the count leaks the closure at completion. (A finish that skips self whole, leaving it set, is
// the assert's: the drop walk would let go of it when the done frame goes, so only the parent's cells
// held for as long as a done frame lives would show it.)
func TestGeneratorRuntimeFinishSelf(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(
		test,
		"self_finish",
		"1\n3\n2\n",
		"generator->self = NULL;\n        adamic_release(self);",
		"generator->self = NULL;\n        (void)self;",
		"heap values leaked",
	)
}

// The drop walk lets go of the carrier at every state: a nested generator dropped before its first
// step releases it with its parameters. Mutant: the walk skips self, and the count leaks.
func TestGeneratorRuntimeDropSelf(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(
		test,
		"self_drop_start",
		"",
		"if (generator->self != NULL) { drop(generator->self); }",
		"(void)0;",
		"heap values leaked",
	)
}

// runningPanics is what each mode that calls into a running generator must panic with: the
// compiler-bug stop names the generator and the verb it was called with.
var runningPanics = map[string]string{
	"reentrant": "adamic: panic: compiler bug: reentering resumed while running",
	"returning": "adamic: panic: compiler bug: returning returned while running",
}

func TestGeneratorRuntimeReentrant(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(test, "reentrant", "", "", "", "")
}

// return() on a running generator names return (Diagnostics' should 6 on #qk8rztp). Mutant, run by
// hand on generator.c: the return arm's word put back to "resumed"; this mode goes red on the text.
func TestGeneratorRuntimeReturnWhileRunning(test *testing.T) {
	test.Parallel()
	generatorRuntimeGuard(test, "returning", "", "", "", "")
}

// A view's dispatch handed NULL answers "not a generator", and adamic_collection_iterator_is answers
// the same, so the dispatch reaches its loud default naming the site (Memory's should on #qk8rztp
// S2c-1 commit 3). Mutant: the NULL test dropped, and the read faults under ASan.
func TestGeneratorRuntimeIsNull(test *testing.T) {
	test.Parallel()
	// The harness's undefined-behavior sanitizer stops the mutant's read of a `NULL` heap header
	// before it can fault, so the mutant is caught in its words, not as a segmentation fault (Loom's
	// red on chg_rkwdewter).
	generatorRuntimeGuard(
		test,
		"is",
		"0 0 1\n",
		"return value != NULL && ((const adamic_heap *)value)->kind == adamic_kind_generator;",
		"return ((const adamic_heap *)value)->kind == adamic_kind_generator;",
		"member access within null pointer",
	)
}

func TestGeneratorRuntimeInitialize(test *testing.T) {
	test.Parallel()
	// Each cell starts empty and owned by the generator, and the initializer is the only thing that
	// makes it so: _new has no zero fill, since the initializer writes every cell whole.
	generatorRuntimeGuard(
		test,
		"drop_start",
		"",
		"adamic_environment_initialize_cells(&generator->heap, generator->cells, count);",
		"(void)0;",
		"Assertion",
	)
}
