package json

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestDocumentedStageZeroGaps(t *testing.T) {
	t.Parallel()
	for _, gap := range []struct {
		file, message, output string
		args                  []string
	}{
		{"multiplePush.ts", "push with other than one value", "a,b\n", nil},
		{"emptyFallback.ts", "an array of never", "0\n", nil},
	} {
		t.Run(gap.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("gaps", gap.file))
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), loaded)
			if err == nil || !strings.Contains(err.Error(), gap.message) {
				t.Fatalf("gap changed: %v; update GAPS.md", err)
			}
			result := onNode(t, path, gap.args...)
			if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != gap.output {
				t.Fatalf("Node: %+v", result)
			}
			t.Logf("%s; Node %q", err, gap.output)
		})
	}
}

// Dynamic repeat now has a catchable count/length guard. Keep the original
// gap source as a positive probe rather than expecting an obsolete refusal.
func TestRepeatInTryGapClosed(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join("gaps", "repeatInTry.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	for _, args := range [][]string{nil, {"a", "b", "c"}} {
		want := onNode(t, path, args...)
		got, binary := natively(t, program, args...)
		for name, result := range map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, program, args...)} {
			if result.exitCode != want.exitCode || string(result.stdout) != string(want.stdout) || string(result.stderr) != string(want.stderr) {
				t.Errorf("%s: %+v; Node: %+v", name, result, want)
			}
		}
		if report := leaks(t, program, binary, args...); report != "" {
			t.Error(report)
		}
	}
}
