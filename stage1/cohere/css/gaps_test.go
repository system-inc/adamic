package css

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestEachGapStandsWhereGapsMdSaysItDoes(test *testing.T) {
	test.Parallel()
	for _, gap := range []struct{ path, stdout, refusal string }{
		{"gaps/2_array_shift.ts", "a\n1\n", "inherited library member shift read as an own field"},
		{"gaps/5_repeat_in_try.ts", "a\n", "a try around repeat"},
	} {
		// Not parallel within the parent: a parallel subtest queues behind every waiting test in the unit, and that
		// wait counts as the parent's own time (212.2 s in the package on Home, for two gaps of 0.15 s each alone on
		// one core).
		test.Run(gap.path, func(test *testing.T) {
			path, err := filepath.Abs(gap.path)
			if err != nil {
				test.Fatal(err)
			}
			result := onNode(test, path)
			if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != gap.stdout {
				test.Fatalf("Node: exit %d, stdout %q, stderr %q", result.exitCode, result.stdout, result.stderr)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				test.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil {
				test.Fatal("gap closed: update GAPS.md and remove the workaround or unblock native composition")
			}
			if !strings.Contains(err.Error(), gap.refusal) {
				test.Fatalf("refusal changed: %v", err)
			}
			test.Logf("Node prints %q; Adamic refuses: %v", gap.stdout, err)
		})
	}
}

// Gap 4 closed on the area-stack slice: an empty array literal assigned into an optional
// number[] slot lowers, so the composition's explicitly typed local is no longer required.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedEmptyArrayUnionGap(test *testing.T) {
	path, _ := filepath.Abs("gaps/4_empty_array_union.ts")
	program := lowered(test, path)
	binary := filepath.Join(programSanitized(test, "gaps/4_empty_array_union.ts"), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(test, path)}, {"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "0\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedParserRegexGap(test *testing.T) {
	path, _ := filepath.Abs("gaps/1_regex_and_value_tree.ts")
	program := lowered(test, path)
	binary := filepath.Join(programSanitized(test, "gaps/1_regex_and_value_tree.ts"), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(test, path)}, {"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "Parsed\nOk\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedOptionalBooleanConditionGap(test *testing.T) {
	path, err := filepath.Abs("gaps/3_optional_boolean_condition.ts")
	if err != nil {
		test.Fatal(err)
	}
	program := lowered(test, path)
	binary := filepath.Join(programSanitized(test, "gaps/3_optional_boolean_condition.ts"), "port")
	nativeRun := sanitizedRun(test, binary)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node", onNode(test, path)},
		{"native ASan/UBSan", nativeRun},
		{"JavaScript backend", onJavaScriptBackend(test, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "important\n" {
			test.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
}
