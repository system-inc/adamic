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
	for _, gap := range []struct{ path, stdout, refusal string }{
		{"gaps/1_regex_and_value_tree.ts", "Parsed\nOk\n", "ir.RegExpCall is a node the cycle finder doesn't know"},
		{"gaps/2_array_shift.ts", "a\n1\n", ".shift on a value"},
		{"gaps/3_optional_boolean_condition.ts", "important\n", "a boolean | undefined as a condition"},
		{"gaps/4_empty_array_union.ts", "0\n", "an array of never"},
		{"gaps/5_repeat_in_try.ts", "a\n", "a try around repeat"},
	} {
		t.Run(gap.path, func(t *testing.T) {
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
func TestNativeCompositionIsBlockedByTheRecordedGap(t *testing.T) {
	path, err := filepath.Abs("compose_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil {
		t.Fatal("native composition now lowers: enable the native/JS backend/sanitizer comparison")
	}
	if !strings.Contains(err.Error(), "ir.RegExpCall is a node the cycle finder doesn't know") {
		t.Fatalf("a different blocker appeared: %v", err)
	}
	t.Logf("%v", err)
}
