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
		if mode == "reentrant" {
			if exit != 70 || !strings.Contains(stderr.String(), "adamic: panic: compiler bug: reentering resumed while running") {
				t.Fatalf("reentrant: %v\n%s", err, output)
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
	generatorRuntimeGuard(t, "return", "1\n", "adamic_generator_finish(generator);\n            break;", "break;", "Assertion")
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

func TestGeneratorRuntimeReentrant(t *testing.T) {
	t.Parallel()
	generatorRuntimeGuard(t, "reentrant", "", "", "", "")
}

func TestGeneratorRuntimeInitialize(t *testing.T) {
	t.Parallel()
	// Each cell starts empty and owned by the generator, and the initializer is the only thing that
	// makes it so: _new has no zero fill, since the initializer writes every cell whole.
	generatorRuntimeGuard(t, "drop_start", "", "adamic_environment_initialize_cells(&generator->heap, generator->cells, count);", "(void)0;", "Assertion")
}
