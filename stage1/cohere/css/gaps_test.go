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
		{"gaps/2_array_shift.ts", "a\n1\n", "the library method shift"},
		{"gaps/3_optional_boolean_condition.ts", "important\n", ""},
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
			compiled, err := lower.Lower(context.Background(), program)
			if gap.refusal == "" {
				if err != nil {
					t.Fatal(err)
				}
				nativeRun, binary := natively(t, compiled)
				for _, side := range []struct {
					name   string
					result run
				}{{"native", nativeRun}, {"JavaScript backend", onJavaScriptBackend(t, compiled)}} {
					if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != gap.stdout {
						t.Fatalf("%s: exit %d, stdout %q, stderr %q", side.name, side.result.exitCode, side.result.stdout, side.result.stderr)
					}
				}
				if report := leaks(t, compiled, binary); report != "" {
					t.Fatal(report)
				}
				return
			}
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
