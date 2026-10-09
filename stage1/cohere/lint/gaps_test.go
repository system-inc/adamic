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
