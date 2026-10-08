package fuzz

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// The runner's refusal check has to be able to fail, and an accepted parallel program has to agree
// at one thread, at the default, and under ThreadSanitizer when that build exists.
func TestParallelRunnerAgreesAndCanFail(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("thread sanitizer: %s", checkout.TSan)

	accepted := "import { parallelMap } from 'adamic';\n" +
		"const items: readonly number[] = [4, 2, 9, 1];\n" +
		"console.log(parallelMap(items, (item, index) => item * 10 + index).join(','));\n"
	if outcome := checkout.Try(accepted, filepath.Join(directory, "accepted")); outcome.Verdict != Agreed {
		t.Fatalf("accepted parallel program: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}

	var generated *Program
	var refused *Program
	// Main added generator features; the captured-let witness now first appears at seed 114.
	// Search the same 300-seed budget as the move coverage check, keeping both witnesses required.
	for seed := uint64(1); seed <= 300; seed++ {
		candidate := GenerateWithout(seed, []string{"moves"})
		if generated == nil && candidate.Refusal == "" && strings.Contains(candidate.Source(), "// parallel-shape: strings") {
			generated = candidate
		}
		if refused == nil && strings.Contains(candidate.Refusal, "not an immutable binding") {
			refused = candidate
		}
	}
	if generated == nil {
		t.Fatal("no string parallel program in seeds 1..300")
	}
	if outcome := checkout.Try(generated.Source(), filepath.Join(directory, "generated")); outcome.Verdict != Agreed {
		t.Fatalf("generated string parallel program: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
	if refused == nil {
		t.Fatal("no captured-let refusal in seeds 1..300")
	}
	if outcome := checkout.Try(refused.Source(), filepath.Join(directory, "refused")); outcome.Verdict != Refused {
		t.Fatalf("captured let: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}

	// Mutant of the check: the comment names a path the program does not break, and the compiler
	// accepts it. The runner must report a finding, not agreement.
	lying := accepted + "// parallel-refuse: task capture 'missing' is not shareable: missing is not an immutable binding\n"
	if outcome := checkout.Try(lying, filepath.Join(directory, "lying")); outcome.Verdict != Finding || outcome.Key != "compiler accepted a program it must refuse" {
		t.Fatalf("lying refusal comment: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}

	// A real refusal whose message names a different path must not count as the expected one.
	wrongPath := "import { parallelMap } from 'adamic';\n" +
		"// parallel-refuse: task capture 'step' is not shareable: step is not an immutable binding\n" +
		"const items: number[] = [1, 2, 3];\n" +
		"parallelMap(items, (item) => item);\n"
	outcome := checkout.Try(wrongPath, filepath.Join(directory, "wrong-path"))
	if outcome.Verdict != Finding || outcome.Key != "refusal did not name the path" {
		t.Fatalf("wrong path: %s %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
	}
	if !strings.Contains(outcome.Detail, "mutable") {
		t.Fatalf("wrong path detail does not show the real refusal:\n%s", outcome.Detail)
	}
}

// A leaked result graph must fail both Linux LSan and Darwin's counted rule.
// Build real task results rather than immortal literal strings or fabricated count lines.
func TestParallelLeakCheckCatchesMutant(t *testing.T) {
	t.Parallel()
	const control = `#include "adamic.h"
#include "parallel.h"
static adamic_value work(adamic_closure *self,adamic_value *arguments) {
 (void)self;return (adamic_value){.reference=adamic_string_from_number(arguments[0].number)};
}
int main(void) {
 adamic_array *items=adamic_array_new(16,false);
 for(size_t i=0;i<16;i++) adamic_array_push(items,(adamic_value){.number=(double)i});
 adamic_closure *callback=adamic_closure_new(work,0);
 adamic_array *results=adamic_parallel_map(items,callback,true);
 adamic_release(results);adamic_release(callback);adamic_release(items);
 return 0;
}
`
	for _, mutant := range []bool{false, true} {
		name := "control"
		code := control
		if mutant {
			name = "missing-results-release"
			code = strings.Replace(code, "adamic_release(results);", "(void)results;", 1)
		}
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			binary := filepath.Join(directory, "sanitized")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			checkout := &Checkout{}
			for _, threads := range []string{"", "1"} {
				report, err := checkout.leaks(binary, directory, threads)
				if err != nil {
					t.Fatal(err)
				}
				counted := filepath.Join(directory, "counted-proof")
				if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
					t.Fatal(err)
				}
				result := executeIsolated(directory, nativeEnvironment(threads), 20*time.Second, counted)
				countedReport := leakcheck.Unbalanced(leakcheck.Run{Stdout: result.Stdout, Stderr: result.Stderr, ExitCode: result.ExitCode})
				if mutant {
					if !strings.Contains(report, "LeakSanitizer: detected memory leaks") && !strings.Contains(report, "heap values leaked:") {
						t.Fatalf("leak check missed task result graph: %s", report)
					}
					if !strings.Contains(countedReport, "heap values leaked:") {
						t.Fatalf("counted rule missed task result graph: %s", countedReport)
					}
					t.Logf("threads %q: result graph mutant caught by shared leak check; %s", threads, countedReport)
				} else if report != "" || countedReport != "" {
					t.Fatalf("control leaks: %s; %s", report, countedReport)
				}
			}
		})
	}
}
