package fuzz

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMovesFeatureCoverage(t *testing.T) {
	t.Parallel()
	shapes := map[string]int{}
	enabled := 0
	for seed := uint64(1); seed <= 300; seed++ {
		program := GenerateMoves(seed)
		source := program.Source()
		if source != GenerateMoves(seed).Source() {
			t.Fatalf("seed %d is not deterministic", seed)
		}
		for _, line := range strings.Split(source, "\n") {
			if strings.HasPrefix(line, "// moves-shape: ") {
				shapes[strings.TrimPrefix(line, "// moves-shape: ")]++
			}
		}
		what, fix := moveRefusal(source)
		if what != program.Refusal || fix != program.RefusalFix {
			t.Fatalf("seed %d lost its exact expectation", seed)
		}
		clone := program.Clone()
		if clone.Source() != source || clone.RefusalFix != fix || clone.Feature != "moves" {
			t.Fatalf("seed %d clone lost move metadata", seed)
		}
		if strings.Contains(Generate(seed).Source(), "// moves-shape:") {
			enabled++
		}
		for _, disabled := range []string{"moves", "parallel"} {
			if strings.Contains(GenerateWithout(seed, []string{disabled}).Source(), "// moves-shape:") {
				t.Fatalf("seed %d: -without %s still generates moves", seed, disabled)
			}
		}
	}
	for _, kind := range []string{"accepted", "after", "alias", "global", "closure", "nested-object", "nested-array", "nested-map"} {
		if shapes[kind] == 0 {
			t.Fatalf("no %s among 300 move seeds", kind)
		}
		t.Logf("%s: %d", kind, shapes[kind])
	}
	if enabled == 0 {
		t.Fatal("moves never enabled in the ordinary generator")
	}
}

// The default is a smoke run; ADAMIC_FUZZ_MOVES_SEEDS requests a full campaign
// or the same bounded seed search against a compiler mutant in a Go overlay.
func TestMovesCampaign(t *testing.T) {
	t.Parallel()
	count := 8
	if value := os.Getenv("ADAMIC_FUZZ_MOVES_SEEDS"); value != "" {
		var err error
		count, err = strconv.Atoi(value)
		if err != nil || count < 1 {
			t.Fatalf("invalid ADAMIC_FUZZ_MOVES_SEEDS %q", value)
		}
	}
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "linux" && checkout.TSan != "ready" {
		t.Fatalf("TSan required for Linux move campaign: %s", checkout.TSan)
	}
	t.Logf("thread sanitizer: %s", checkout.TSan)
	started := time.Now()
	totals := map[Verdict]int{}
	for seed := uint64(1); seed <= uint64(count); seed++ {
		program := GenerateMoves(seed)
		outcome := checkout.Try(program.Source(), filepath.Join(directory, "program"))
		expected := Agreed
		if program.Refusal != "" {
			expected = Refused
		}
		if outcome.Verdict != expected {
			t.Fatalf("seed %d after %s: %s %s\n%s", seed, time.Since(started), outcome.Verdict, outcome.Key, outcome.Detail)
		}
		totals[outcome.Verdict]++
		t.Logf("seed %d: %s %s", seed, outcome.Verdict, outcome.Key)
	}
	t.Logf("%d move seeds in %s: agreed %d refused %d", count, time.Since(started), totals[Agreed], totals[Refused])
}

func TestMovesRunnerCanFail(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	var accepted, refused *Program
	for seed := uint64(1); seed <= 300 && (accepted == nil || refused == nil); seed++ {
		program := GenerateMoves(seed)
		if program.Refusal == "" {
			accepted = program
		} else {
			refused = program
		}
	}
	if accepted == nil || refused == nil {
		t.Fatal("missing move witnesses")
	}
	for name, source := range map[string]string{
		"path-prefix":   strings.Replace(refused.Source(), refused.Refusal, strings.Split(refused.Refusal, ":")[0], 1),
		"wrong-fix":     strings.Replace(refused.Source(), movesFixPrefix+" "+refused.RefusalFix, movesFixPrefix+" "+oppositeMoveFix(refused.RefusalFix), 1),
		"not-a-refusal": accepted.Source() + "\n" + movesRefusePrefix + " cannot move items: whole reachable ownership is not proven\n" + movesFixPrefix + " return it through the results\n",
	} {
		outcome := checkout.Try(source, filepath.Join(directory, name))
		if outcome.Verdict != Finding {
			t.Fatalf("%s: false success: %+v", name, outcome)
		}
		if name != "not-a-refusal" && outcome.Key != "move refusal differs" {
			t.Fatalf("%s failed for the wrong reason: %+v", name, outcome)
		}
	}
	path := filepath.Join(directory, "finding.a")
	if err := WriteFinding(path, accepted, 1, nil, nil, "example"); err != nil {
		t.Fatal(err)
	}
	bytes, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(bytes), "-only-moves") {
		t.Fatal(fmt.Sprintf("focused replay lost: %v %s", err, bytes))
	}
}

func oppositeMoveFix(fix string) string {
	if fix == "return it through the results" {
		return "don't use it after the parallelMap"
	}
	return "return it through the results"
}

// Even a predicate fooled by the expected-refusal comment must not shrink a
// move finding down to that comment alone, with its task boundary gone.
func TestMovesShrinkKeepsBoundary(t *testing.T) {
	t.Parallel()
	program := GenerateMoves(2)
	if program.RefusalFix == "" {
		t.Fatal("seed 2 is not the alias witness")
	}
	shrunk := Shrink(program, Finding, "accepted refusal", func(candidate *Program) Outcome {
		if strings.Contains(candidate.Source(), movesRefusePrefix) && strings.Contains(candidate.Source(), movesFixPrefix) {
			return Outcome{Verdict: Finding, Key: "accepted refusal"}
		}
		return Outcome{Verdict: Agreed}
	})
	if !strings.Contains(shrunk.Source(), "parallelMap(") {
		t.Fatalf("lost the boundary: %s", shrunk.Source())
	}
}
