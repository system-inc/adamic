package fuzz

import (
	"path/filepath"
	"strings"
	"testing"
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
