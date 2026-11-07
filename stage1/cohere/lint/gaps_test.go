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

func TestNestedConstructorGap(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("gaps/1_nested_constructor.ts")
	if err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	answer := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(answer.output) != "1\n" {
		test.Fatalf("Node gap answer %q", answer.output)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "escaping a constructor before every field is set") {
		test.Fatalf("GAPS.md records a false constructor refusal, got %v", err)
	}
	test.Logf("Node prints 1; native lowering refuses: %v", err)
}
