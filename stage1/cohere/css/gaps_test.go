package css

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestEachGapStandsWhereGapsMdSaysItDoes(t *testing.T) {
	t.Parallel()
	for _, gap := range []struct{ path, stdout, refusal string }{
		{"gaps/2_array_shift.ts", "a\n1\n", "inherited library member shift read as an own field"},
		{"gaps/5_repeat_in_try.ts", "a\n", "a try around repeat"},
	} {
		t.Run(gap.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(gap.path)
			if err != nil {
				t.Fatal(err)
			}
			result := onNode(t, path)
			if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != gap.stdout {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", result.exitCode, result.stdout, result.stderr)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil {
				t.Fatal("gap closed: update GAPS.md and remove the workaround or unblock native composition")
			}
			if !strings.Contains(err.Error(), gap.refusal) {
				t.Fatalf("refusal changed: %v", err)
			}
			t.Logf("Node prints %q; Adamic refuses: %v", gap.stdout, err)
		})
	}
}

// Gap 4 closed on the area-stack slice: an empty array literal assigned into an optional
// number[] slot lowers, so the composition's explicitly typed local is no longer required.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedEmptyArrayUnionGap(t *testing.T) {
	path, _ := filepath.Abs("gaps/4_empty_array_union.ts")
	program := lowered(t, path)
	nativeRun, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(t, path)}, {"JavaScript backend", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "0\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedParserRegexGap(t *testing.T) {
	path, _ := filepath.Abs("gaps/1_regex_and_value_tree.ts")
	program := lowered(t, path)
	nativeRun, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native", nativeRun}, {"Node", onNode(t, path)}, {"JavaScript backend", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "Parsed\nOk\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedOptionalBooleanConditionGap(t *testing.T) {
	path, err := filepath.Abs("gaps/3_optional_boolean_condition.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	nativeRun, binary := natively(t, program)
	for _, side := range []struct {
		name   string
		result run
	}{
		{"Node", onNode(t, path)}, {"native ASan/UBSan", nativeRun}, {"JavaScript backend", onJavaScriptBackend(t, program)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != "important\n" {
			t.Fatalf("%s: %d %q %s", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
