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
	if _, err := fmt.Sscanf(lines[len(lines)-1], "adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d", &allocations, &frees, &retains, &releases, &peak, &regions); err != nil {
		return "no counts line: " + err.Error()
	}
	if allocations != frees+regions {
		return fmt.Sprintf("heap values leaked: %d (allocations %d, frees %d)", allocations-frees-regions, allocations, frees)
	}
	return ""
}

// generatorRuntimeGuard runs one mode of the hand-written generator harness, a stand-in for what the
// emitter will write, against the runtime with no compiler lowering. The build is both sanitized and
// counted, so a use after free fails under ASan and a leak fails the counts on every platform, Darwin
// included, where LeakSanitizer doesn't run. A non-empty old makes a second run with that exact text
// of generator.c replaced, which must fail with diagnostic.
func generatorRuntimeGuard(t *testing.T, mode, want, old, replacement, diagnostic string) {
	t.Helper()
	files, flags, compiler, version, err := runtimeInputs("", Options{Sanitize: true, Count: true}, []string{"-DADAMIC_CANONICAL_CLOSURES=1"})
	if err != nil {
		t.Fatal(err)
	}
	cache, err := cacheDirectory("generator-runtime")
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("testdata/generator/runtime.c")
	if err != nil {
		t.Fatal(err)
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
					t.Fatalf("mutant seam %q appears %d times", old, strings.Count(text, old))
				}
				files[index].contents = []byte(strings.Replace(text, old, replacement, 1))
				found = true
			}
			if !found {
				t.Fatal("missing runtime file generator.c")
			}
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			t.Fatal(err)
		}
		binary := buildHarnessWithLibrary(t, string(source), flags, compiler, library)
		command, release := boundedrun.Command(5*time.Second, binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1")
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		err = command.Run()
		release()
		exit := 0
		if command.ProcessState != nil {
			exit = command.ProcessState.ExitCode()
		}
		output := stdout.String() + stderr.String()
		if mutant {
			leaked := generatorLeaked(stderr.String())
			if (err == nil && leaked == "") || !strings.Contains(output+leaked, diagnostic) {
				t.Fatalf("mutant survived or failed for another reason: %v %s\n%s", err, leaked, output)
			}
			t.Logf("%s mutant caught by %s: %s", mode, diagnostic, firstSanitizerLine(output))
			continue
		}
		if panic, panics := modePanics[mode]; panics {
			if exit != 70 || !strings.Contains(stderr.String(), panic) {
				t.Fatalf("%s: %v\n%s", mode, err, output)
			}
			continue
		}
		if err != nil || stdout.String() != want {
			t.Fatalf("control: %v\nwant stdout %q\n%s", err, want, output)
		}
		if leaked := generatorLeaked(stderr.String()); leaked != "" {
			t.Fatalf("control: %s\n%s", leaked, output)
		}
	}
}

func TestGeneratorRuntimeComplete(t *testing.T) {
	t.Parallel()
	// finish must NULL what it released, or the drop after completion releases it again.
	generatorRuntimeGuard(t, "complete", "1\n3\n2\n", "cell->value.reference = NULL;\n            adamic_release(owned);", "adamic_release(owned);", "heap-use-after-free")
}

func TestGeneratorRuntimeDropStart(t *testing.T) {
	t.Parallel()
	// A generator never started still owns its parameters.
	generatorRuntimeGuard(t, "drop_start", "", "drop(generator->cells[index].value.reference);", "(void)drop;", "heap values leaked")
}

func TestGeneratorRuntimeDropSuspended(t *testing.T) {
	t.Parallel()
	// Dropped at a yield, the walk frees every live cell.
	generatorRuntimeGuard(t, "drop_suspended", "1\n", "drop(generator->cells[index].value.reference);", "(void)drop;", "heap values leaked")
}

func TestGeneratorRuntimeReturn(t *testing.T) {
	t.Parallel()
	// The finish return runs, retained around it since q3, taken out whole.
	generatorRuntimeGuard(t, "return", "1\n", "adamic_retain(generator);\n            adamic_generator_finish(generator);\n            adamic_release(generator);\n            break;", "break;", "Assertion")
}

func TestGeneratorRuntimeKill(t *testing.T) {
	t.Parallel()
	// A killed cell left pointing at what it released is released again by the drop walk.
	generatorRuntimeGuard(t, "kill", "1\n", "cell->value.reference = NULL;\n        adamic_release(held);", "adamic_release(held);", "heap-use-after-free")
}

func TestGeneratorRuntimeBorrowed(t *testing.T) {
	t.Parallel()
	// A borrowed alias is never the generator's to drop.
	generatorRuntimeGuard(t, "borrowed", "7\n7\n", "if (generator->cells[index].references && generator->cells[index].value.reference != NULL) {", "if (generator->cells[index].value.reference != NULL) {", "heap-use-after-free")
}

func TestGeneratorRuntimeThrow(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "throw", "5\nError: 6\n", "", "", "")
}

// finish owns an error still set when it runs. Mutant: finish forgets it, and the count leaks.
func TestGeneratorRuntimeFinishError(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "finish_error", "", "generator->error = NULL;\n        adamic_release(error);", "generator->error = NULL;\n        (void)error;", "heap values leaked")
}

// The drop walk drops an error it finds. Mutant: the walk skips error, and the count leaks.
func TestGeneratorRuntimeDropError(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "drop_error", "1\n", "if (generator->error != NULL) { drop(generator->error); }", "(void)0;", "heap values leaked")
}

// A nested generator's carrier, the closure its entry retains into self (Memory's ruling on #qk8rztp
// S3a): finish lets go of it and leaves self `NULL`, so a done generator stops pinning its parent's
// cells, and the harness asserts it `NULL` after the last step. Mutant: finish forgets self's release,
// and the count leaks the closure at completion. (A finish that skips self whole, leaving it set, is
// the assert's: the drop walk would let go of it when the done frame goes, so only the parent's cells
// held for as long as a done frame lives would show it.)
func TestGeneratorRuntimeFinishSelf(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "self_finish", "1\n3\n2\n", "generator->self = NULL;\n        adamic_release(self);", "generator->self = NULL;\n        (void)self;", "heap values leaked")
}

// The drop walk lets go of the carrier at every state: a nested generator dropped before its first
// step releases it with its parameters. Mutant: the walk skips self, and the count leaks.
func TestGeneratorRuntimeDropSelf(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "self_drop_start", "", "if (generator->self != NULL) { drop(generator->self); }", "(void)0;", "heap values leaked")
}

// modePanics is what each mode that reaches a compiler-bug stop must panic with. A call into a
// running generator names the generator and the verb it was called with; a hand-driven next() names
// its site and what it was handed.
var modePanics = map[string]string{
	"reentrant":  "adamic: panic: compiler bug: reentering resumed while running",
	"returning":  "adamic: panic: compiler bug: returning returned while running",
	"next_other": "adamic: panic: compiler bug: the hand-driven next() at harness.ts:3:9 steps a value that is neither a generator nor a collection's iterator",
	"next_null":  "adamic: panic: compiler bug: the hand-driven next() at harness.ts:3:9 steps a value that is neither a generator nor a collection's iterator",
	"next_kind":  "adamic: panic: compiler bug: the hand-driven next() at harness.ts:3:9 steps a generator whose yields no result holds",
}

func TestGeneratorRuntimeReentrant(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "reentrant", "", "", "", "")
}

// return() on a running generator names return (Diagnostics' should 6 on #qk8rztp). Mutant, run by
// hand on generator.c: the return arm's word put back to "resumed"; this mode goes red on the text.
func TestGeneratorRuntimeReturnWhileRunning(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "returning", "", "", "", "")
}

// A view's dispatch handed NULL answers "not a generator", and adamic_collection_iterator_is answers
// the same, so the dispatch reaches its loud default naming the site (Memory's should on #qk8rztp
// S2c-1 commit 3). Mutant: the NULL test dropped, and the read faults under ASan.
func TestGeneratorRuntimeIsNull(t *testing.T) {
	t.Parallel()
	// The harness's undefined-behavior sanitizer stops the mutant's read of a `NULL` heap header
	// before it can fault, so the mutant is caught in its words, not as a segmentation fault (Loom's
	// red on chg_rkwdewter).
	generatorRuntimeGuard(t, "is", "0 0 1\n", "return value != NULL && ((const adamic_heap *)value)->kind == adamic_kind_generator;", "return ((const adamic_heap *)value)->kind == adamic_kind_generator;", "member access within null pointer")
}

func TestGeneratorRuntimeInitialize(t *testing.T) {
	t.Parallel()
	// Each cell starts empty and owned by the generator, and the initializer is the only thing that
	// makes it so: _new has no zero fill, since the initializer writes every cell whole.
	generatorRuntimeGuard(t, "drop_start", "", "adamic_environment_initialize_cells(&generator->heap, generator->cells, count);", "(void)0;", "Assertion")
}

// A hand-driven next() through a view a generator feeds (adamic_iterator_next, #qk8rztp S2c-2b), held
// to Memory's ownership contract on #qk8rztp before any compiler code calls it. Each row is a twin of
// one of her pins, and its mutant is one exact text of generator.c.

// P1's twin: a string generator run to done, each result owning the value moved into it. M1: the arm
// retains the yielded value as well, and every string outlives its result in the counts.
func TestGeneratorRuntimeNextStrings(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "next_strings", "1\n2\n", "    return adamic_iterator_result(kind, present, yielded, tag);", "    if (present) { adamic_retain(yielded.reference); }\n    return adamic_iterator_result(kind, present, yielded, tag);", "heap values leaked")
}

// P2's twin: one step, its value read, then the result and the suspended frame dropped. M2: the arm
// treats the value as borrowed and releases it after the build, so the result holds a freed string.
func TestGeneratorRuntimeNextDrop(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "next_drop", "1\n", "    return adamic_iterator_result(kind, present, yielded, tag);", "    adamic_object *built = adamic_iterator_result(kind, present, yielded, tag);\n    adamic_release(yielded.reference);\n    return built;", "heap-use-after-free")
}

// P3's twin: the body throws on the second step, the arm returns NULL and builds nothing, and next()
// after the throw answers done. M3: the builder runs before the throw test, and its result leaks.
func TestGeneratorRuntimeNextThrow(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "next_throw", "5\nError: 6\n", "            return NULL;\n        }\n        return generator_result(", "            (void)generator_result(generator, kind, element_type, yielded);\n            return NULL;\n        }\n        return generator_result(", "heap values leaked")
}

// P4: a number generator stepped to done and once more reads its value present, then absent twice,
// and the done result's keys are value,done, in the order #7nw4kgj gave the shared shapes. M4: done
// passes the frame's zero through as present, and the harness's presence assertion stops it.
func TestGeneratorRuntimeNextDone(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "next_done", "1 10\n0\n0\nvalue,done\n", "bool present = !adamic_generator_done(generator);", "bool present = true;", "Assertion")
}

// P4's boolean row: a boolean generator stepped past done reads its value undefined, under the
// MaybeBoolean byte the builder writes (Memory's ruling on #ak92fcp), and present it reads true under
// Boolean. M6, run by h3-proof.sh on map_set.c, outside this guard's one file: the builder's boolean
// arm skipped, so the slot keeps the absent maybe-number word, main's bytes, whose low byte reads true.
func TestGeneratorRuntimeNextDoneBoolean(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "next_done_boolean", "true 2\nundefined 9\nundefined 9\n", "", "", "")
}

// A collection's iterator steps through the same entry by its own next, and its results are its own.
func TestGeneratorRuntimeNextCollection(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "next_collection", "7\n8\n", "", "", "")
}

// The loud default: a value the proof should have ruled out, a string or NULL, panics naming the site,
// never a fault, and a yield off the allowlist panics before any step.
func TestGeneratorRuntimeNextLoudDefault(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"next_other", "next_null", "next_kind"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			generatorRuntimeGuard(t, mode, "", "", "", "")
		})
	}
}
