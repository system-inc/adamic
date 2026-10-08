package lint

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestNestedConstructorGap(t *testing.T) {
	path, err := filepath.Abs("gaps/1_nested_constructor.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	answer := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(answer.output) != "1\n" {
		t.Fatalf("Node gap answer %q", answer.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "escaping a constructor before every field is set") {
		t.Fatalf("GAPS.md records a false constructor refusal, got %v", err)
	}
	t.Logf("Node prints 1; native lowering refuses: %v", err)
}

func TestOptionAndComparatorGaps(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, answer, refusal string }{
		{"2_optional_index.ts", "1\n", "?.[] on a value"},
		{"4_last_index_position.ts", "1\n", "lastIndexOf with these arguments"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join("gaps", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			answer := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
			if string(answer.output) != probe.answer {
				t.Fatalf("Node prints %q, want %q", answer.output, probe.answer)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.refusal) {
				t.Fatalf("documented refusal absent: %v", err)
			}
			t.Logf("Node prints %q; stage 0 refuses: %v", answer.output, err)
		})
	}
}
