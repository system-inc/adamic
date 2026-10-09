package native

import (
	"path/filepath"
	"strings"
	"testing"
)

// A Node can differ from Adamic in the last bit of Math and parseInt for one reason Adamic accepts:
// its V8 was compiled with multiply-adds contracted. Node v24.14.1 on macOS arm64 is, with clang's
// default, -ffp-contract=on: the runtime built the same way, its FP_CONTRACT OFF pragmas lifted, gives
// that Node's Math and parseInt bit for bit (docs/library-math/contraction.md). Adamic never fuses (Flags,
// contract_test.go), so a program prints one answer on every platform, Node's on x86-64.
//
// So a difference is forgiven only where the same runtime, compiled with contraction allowed
// (Options.FusedRuntime), gives exactly Node's answer. On a machine that can't fuse, that build is
// the unfused one, and nothing is forgiven: the comparison stays bit for bit, as on Linux x86-64, the
// gate of record. A mistake in the runtime's source is in both builds, so it still shows.

// fusedAnswers runs harness on input, its runtime built with contraction allowed, one answer per line.
func fusedAnswers(t *testing.T, harness string, input string, lines int) []string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "fused")
	if err := Build(harness, binary, Options{FusedRuntime: true}); err != nil {
		t.Fatal(err)
	}
	answers := strings.Split(strings.TrimSuffix(runWithInput(t, input, binary), "\n"), "\n")
	if len(answers) != lines {
		t.Fatalf("got %d answers from the fused build for %d questions", len(answers), lines)
	}
	return answers
}

// contraction splits the answers where native and Node differ under same into those a fused build
// explains, which it counts, and those it doesn't, whose indices it returns. It asks for the fused
// build's answers (fusedAnswers) only when there is a difference to explain.
func contraction(native []string, oracle []string, same func(left, right string) bool, fused func() []string) (explained int, unexplained []int) {
	var answers []string
	for index := range native {
		if same(native[index], oracle[index]) {
			continue
		}
		if answers == nil {
			answers = fused()
		}
		if same(answers[index], oracle[index]) {
			explained++
			continue
		}
		unexplained = append(unexplained, index)
	}
	return explained, unexplained
}

func TestContractionForgivesOnlyWhatAFusedBuildReproduces(t *testing.T) {
	t.Parallel()
	equal := func(left, right string) bool { return left == right }
	built := 0
	fused := func() []string { built++; return []string{"a", "B", "c", "d"} }
	// Index 1 differs and the fused build gives Node's answer; index 2 differs and it doesn't.
	explained, unexplained := contraction([]string{"a", "b", "c", "d"}, []string{"a", "B", "C", "d"}, equal, fused)
	if explained != 1 || len(unexplained) != 1 || unexplained[0] != 2 {
		t.Errorf("explained %d, unexplained %v; want 1 explained and index 2 unexplained", explained, unexplained)
	}
	if built != 1 {
		t.Errorf("built the fused harness %d times for one comparison", built)
	}
	built = 0
	if explained, unexplained := contraction([]string{"a"}, []string{"a"}, equal, fused); explained != 0 || len(unexplained) != 0 || built != 0 {
		t.Errorf("with nothing to explain: explained %d, unexplained %v, built %d times", explained, unexplained, built)
	}
}
