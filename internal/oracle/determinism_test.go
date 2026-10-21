package oracle

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"os/exec"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

// Stage 0's emission is deterministic (@system_adamic, Oct 9: a 43,250-line main.c on the Threadripper against 5,470
// lines on a Codex box for the same tree). Each fixture's C is emitted twice, each in its own process so Go's
// per-process map order can differ, and the two must match byte for byte, exit code included. This was the gate's
// determinism phase over the 46 smoke programs, which compiled ./cmd/adamic in every unit (128 s of its 129.5 s, Tests'
// measurement on #zw6hb47); here it reads the stage 0 product the tree builds ahead (TestProduct_Adamic) and holds every
// registered fixture to it, lowering or not, as a family like the oracle's others (#86znq4s).

// determinismMutation is the probe's planted nondeterminism: methodList's sort dropped, so a class's dispatch map
// follows map order.
var determinismMutation = buildcache.Mutation{File: "internal/lower/class.go", Before: "\tsort.Strings(names)\n", After: "\t_ = sort.Strings\n"}

// TestProduct_AdamicMapOrderMutant declares the stage 0 compiler with determinismMutation planted, which the probe of
// TestStage0EmitsDeterministically runs, so the tree build builds it ahead.
func TestProduct_AdamicMapOrderMutant(t *testing.T) {
	t.Parallel()
	adamicMapOrderMutant(t)
}

// determinismWitness is the probe's planted case, emitted and never run, so it is one of this family's keys and not a
// registered fixture. Go iterates a small map from a random rotation, so one class's methods come out in only a few
// orders: under determinismMutation the likeliest fixture, class_oct6_deep.a, emitted alike in 3 of 60 pairs (Mac, Oct
// 13), and the probe missed 2 of 130 times. Twelve classes of eight methods each rotate on their own: the mutant
// emitted it alike in 0 of 1,000 pairs, and stage 0 as it is emitted it one way in 20 runs.
const determinismWitness = "internal/oracle/testdata/determinism_dispatch.a"

// determinismKeys is every registered fixture's path, then the witness.
func determinismKeys() []string {
	return append(fixturePaths(), determinismWitness)
}

func adamicMapOrderMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoBuildInMutated(t, "", "adamic", "./cmd/adamic", nil, []buildcache.Mutation{determinismMutation})
}

// emitted is one emission of a fixture's C by a stage 0 binary, in its own process from the repository's root as the
// gate ran it: its exit code and its standard output.
type emitted struct {
	code   int
	output []byte
}

func emit(t *testing.T, binary, path string) emitted {
	t.Helper()
	ctx, command := bounded(t, binary, "c", path)
	command.Dir = repository
	output, _, err := childguard.Capture(ctx, command)
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("running %s c %s: %v", binary, path, err)
	}
	return emitted{code: command.ProcessState.ExitCode(), output: output}
}

// firstDifferingLine is the 1-based line where two emissions first differ, 0 when only the exit codes do.
func firstDifferingLine(first, second []byte) int {
	a, b := bytes.Split(first, []byte("\n")), bytes.Split(second, []byte("\n"))
	for index := range max(len(a), len(b)) {
		if index >= len(a) || index >= len(b) || !bytes.Equal(a[index], b[index]) {
			return index + 1
		}
	}
	return 0
}

func runDeterministicEmissionShard(t *testing.T, shard int) {
	t.Helper()
	binary := buildcache.Adamic(t)
	if _, probing := familyProbe("TestStage0EmitsDeterministically"); probing {
		binary = adamicMapOrderMutant(t)
	}
	keys := determinismKeys()
	for _, index := range familyCases(t, "TestStage0EmitsDeterministically", shard, keys) {
		path := keys[index]
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			first, second := emit(t, binary, path), emit(t, binary, path)
			if first.code != second.code || !bytes.Equal(first.output, second.output) {
				t.Fatalf("stage 0 emitted different C for the same program in two runs: exit %d and %d, first difference at line %d", first.code, second.code, firstDifferingLine(first.output, second.output))
			}
			t.Logf("exit %d, emission sha256 %x", first.code, sha256.Sum256(first.output))
		})
	}
}

// The comparison has to be able to fail on what it reads: two emissions differing only in a byte, only in a line
// added, or only in the exit code each differ, and the line reported is where they do.
func TestFirstDifferingLineFindsWhereEmissionsDiffer(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		first, second string
		line          int
	}{
		{"a\nb\nc", "a\nb\nc", 0},
		{"a\nb\nc", "a\nB\nc", 2},
		{"a\nb", "a\nb\nc", 3},
		{"", "a", 1},
	} {
		if got := firstDifferingLine([]byte(each.first), []byte(each.second)); got != each.line {
			t.Errorf("firstDifferingLine(%q, %q) = %d, want %d", each.first, each.second, got, each.line)
		}
	}
}
