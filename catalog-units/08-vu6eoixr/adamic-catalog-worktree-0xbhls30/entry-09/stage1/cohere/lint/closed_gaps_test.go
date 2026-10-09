package lint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Gaps compiler/area-next closed: each lowers now, and native (sanitized) and the JavaScript backend
// print what Node prints. GAPS.md keeps the history.
func TestClosedComparatorGaps(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, answer string }{
		{"3_numeric_or.ts", "2\n"},
		{"2_optional_index.ts", "1\n"},
		{"4_last_index_position.ts", "1\n"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("gaps", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			if answer := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path); string(answer.output) != probe.answer {
				t.Fatalf("Node prints %q, want %q", answer.output, probe.answer)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			lowered, err := lower.Lower(context.Background(), program)
			if err != nil {
				t.Fatalf("closed gap stopped lowering: %v", err)
			}
			binary := filepath.Join(t.TempDir(), "gap")
			if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			if got := execute(t, "", binary); string(got.output) != probe.answer {
				t.Fatalf("native prints %q, Node %q", got.output, probe.answer)
			}
			emitted := filepath.Join(t.TempDir(), "gap.mjs")
			if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0o644); err != nil {
				t.Fatal(err)
			}
			if got := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, emitted); string(got.output) != probe.answer {
				t.Fatalf("JavaScript backend prints %q, Node %q", got.output, probe.answer)
			}
		})
	}
}
