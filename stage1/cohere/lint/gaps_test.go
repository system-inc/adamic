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
	path, absolutePathError := filepath.Abs("gaps/1_nested_constructor.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	answer := execute(test, "", "node", "--disable-warning=ExperimentalWarning", runner, path)
	if string(answer.output) != "1\n" {
		test.Fatalf("Node gap answer %q", answer.output)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(lowerError, &refused) ||
		!strings.Contains(lowerError.Error(), "escaping a constructor before every field is set") {
		test.Fatalf("GAPS.md records a false constructor refusal, got %v", lowerError)
	}
	test.Logf("Node prints 1; native lowering refuses: %v", lowerError)
}
