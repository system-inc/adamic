package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// Each runner case is independently selectable within the gate budget.
func prepareParallelRunner(test *testing.T) (string, *Checkout) {
	test.Helper()
	directory := test.TempDir()
	checkout, prepareBuiltError := PrepareBuilt("../..", buildcache.Adamic(test), filepath.Join(directory, "checkout"))
	if prepareBuiltError != nil {
		test.Fatal(prepareBuiltError)
	}
	test.Logf("thread sanitizer: %s", checkout.TSan)

	return directory, checkout
}

func TestParallelRunnerNumbersAgree(test *testing.T) {
	test.Parallel()
	directory, checkout := prepareParallelRunner(test)
	accepted := "import { parallelMap } from 'adamic';\n" +
		"const items: readonly number[] = [4, 2, 9, 1];\n" +
		"console.log(parallelMap(items, (item, index) => item * 10 + index).join(','));\n"
	if outcome := checkout.Try(accepted, filepath.Join(directory, "accepted")); outcome.Verdict != Agreed {
		test.Fatalf("accepted parallel program: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
}

func TestParallelRunnerStringsAgree(test *testing.T) {
	test.Parallel()
	directory, checkout := prepareParallelRunner(test)
	var generated *Program
	// Search a bounded seed range so added generator features may shift the witness.
	for seed := uint64(1); seed <= 300; seed++ {
		candidate := GenerateWithout(seed, []string{"moves"})
		if generated == nil &&
			candidate.Refusal == "" &&
			strings.Contains(candidate.Source(), "// parallel-shape: strings") {
			generated = candidate
		}
	}
	if generated == nil {
		test.Fatal("no string parallel program in seeds 1..300")
	}
	if outcome := checkout.Try(generated.Source(), filepath.Join(directory, "generated")); outcome.Verdict != Agreed {
		test.Fatalf("generated string parallel program: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
}

func TestParallelRunnerCapturedLetRefused(test *testing.T) {
	test.Parallel()
	directory, checkout := prepareParallelRunner(test)
	var refused *Program
	// Search a bounded seed range so added generator features may shift the witness.
	for seed := uint64(1); seed <= 300; seed++ {
		candidate := GenerateWithout(seed, []string{"moves"})
		if refused == nil && strings.Contains(candidate.Refusal, "not an immutable binding") {
			refused = candidate
		}
	}
	if refused == nil {
		test.Fatal("no captured-let refusal in seeds 1..300")
	}
	if outcome := checkout.Try(refused.Source(), filepath.Join(directory, "refused")); outcome.Verdict != Refused {
		test.Fatalf("captured let: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
}

func TestParallelRunnerCatchesLyingRefusal(test *testing.T) {
	test.Parallel()
	directory, checkout := prepareParallelRunner(test)
	accepted := "import { parallelMap } from 'adamic';\n" +
		"const items: readonly number[] = [4, 2, 9, 1];\n" +
		"console.log(parallelMap(items, (item, index) => item * 10 + index).join(','));\n"
	// Mutant of the check: the comment names a path the program does not break, and the compiler
	// accepts it. The runner must report a finding, not agreement.
	lying := accepted +
		"// parallel-refuse: task capture 'missing' is not shareable: missing is not an immutable binding\n"
	if outcome := checkout.Try(
		lying,
		filepath.Join(directory, "lying"),
	); outcome.Verdict != Finding || outcome.Key != "compiler accepted a program it must refuse" {
		test.Fatalf("lying refusal comment: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
}

func TestParallelRunnerCatchesWrongPath(test *testing.T) {
	test.Parallel()
	directory, checkout := prepareParallelRunner(test)
	// A real refusal whose message names a different path must not count as the expected one.
	wrongPath := "import { parallelMap } from 'adamic';\n" +
		"// parallel-refuse: task capture 'step' is not shareable: step is not an immutable binding\n" +
		"const items: number[] = [1, 2, 3];\n" +
		"parallelMap(items, (item) => item);\n"
	outcome := checkout.Try(wrongPath, filepath.Join(directory, "wrong-path"))
	if outcome.Verdict != Finding || outcome.Key != "refusal did not name the path" {
		test.Fatalf("wrong path: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
	if !strings.Contains(outcome.Detail, "mutable") {
		test.Fatalf("wrong path detail does not show the real refusal:\n%s", outcome.Detail)
	}
}

// A leaked result graph must fail both Linux LSan and Darwin's counted rule.
// Build real task results rather than immortal literal strings or fabricated count lines.
// Run on a thread that exits before leak scanning: with one worker the result pointer
// can survive on the main stack, so LSan can conservatively call it reachable.
// Join the pool too, so no scheduler stack can keep a stale scope pointer alive.
func TestParallelLeakCheckCatchesMutant(test *testing.T) {
	test.Parallel()
	const control = `#include "adamic.h"
#include "parallel.h"
#include <pthread.h>
static adamic_value work(adamic_closure *self,adamic_value *arguments) {
 (void)self;return (adamic_value){.reference=adamic_string_from_number(arguments[0].number)};
}
static void *probe(void *unused) {
 (void)unused;
 adamic_array *items=adamic_array_make(16,adamic_ir_Number);
 for(size_t i=0;i<16;i++) adamic_array_push(items,(adamic_value){.number=(double)i});
 adamic_closure *callback=adamic_closure_new(work,0);
 adamic_array *results=adamic_parallel_map(items,callback,adamic_ir_Union);
 adamic_release(results);adamic_release(callback);adamic_release(items);
 adamic_parallel_shutdown();
 adamic_heap_thread_end();
 return NULL;
}
int main(void) {
 pthread_t thread;
 if (pthread_create(&thread,NULL,probe,NULL) != 0) { return 1; }
 return pthread_join(thread,NULL) != 0;
}
`
	for _, mutant := range []bool{false, true} {
		name := "control"
		code := control
		if mutant {
			name = "missing-results-release"
			code = strings.Replace(code, "adamic_release(results);", "(void)results;", 1)
		}
		test.Run(name, func(test *testing.T) {
			directory := test.TempDir()
			if writeError := os.WriteFile(filepath.Join(directory, "main.c"), []byte(code), 0600); writeError != nil {
				test.Fatal(writeError)
			}
			binary := filepath.Join(directory, "sanitized")
			if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
				test.Fatal(buildError)
			}
			counted := filepath.Join(directory, "counted-proof")
			if buildError := native.Build(code, counted, native.Options{Count: true}); buildError != nil {
				test.Fatal(buildError)
			}
			checkout := &Checkout{}
			for _, threads := range []string{"", "1"} {
				report, leaksError := checkout.leaks(binary, directory, threads)
				if leaksError != nil {
					test.Fatal(leaksError)
				}
				result := executeIsolated(directory, nativeEnvironment(threads), 20*time.Second, counted)
				countedReport := leakcheck.Unbalanced(
					leakcheck.Run{Stdout: result.Stdout, Stderr: result.Stderr, ExitCode: result.ExitCode},
				)
				if mutant {
					if !strings.Contains(report, "LeakSanitizer: detected memory leaks") &&
						!strings.Contains(report, "heap values leaked:") {
						test.Fatalf("leak check missed task result graph: %s", report)
					}
					if !strings.Contains(countedReport, "heap values leaked:") {
						test.Fatalf("counted rule missed task result graph: %s", countedReport)
					}
					test.Logf("threads %q: result graph mutant caught by shared leak check; %s", threads, countedReport)
				} else if report != "" || countedReport != "" {
					test.Fatalf("control leaks: %s; %s", report, countedReport)
				}
			}
		})
	}
}
