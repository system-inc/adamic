package native

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The ruling explicitly requires this mutant to die in the C compiler.
func TestClosureConventionDropCount(t *testing.T) {
	checked, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_value_count.a")})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	source := C(program)
	if err := Build(source, filepath.Join(t.TempDir(), "valid"), Options{}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(source, "\n")
	mutated := false
	for i, line := range lines {
		if !strings.Contains(line, "= adamic_closure_call(") {
			continue
		}
		last := strings.LastIndex(line, ", ")
		close := strings.LastIndex(line, ");")
		if last < 0 || close < last {
			t.Fatal("counted call has unexpected form")
		}
		lines[i] = line[:last] + line[close:]
		mutated = true
		break
	}
	if !mutated {
		t.Fatal("no counted call to mutate")
	}
	err = Build(strings.Join(lines, "\n"), filepath.Join(t.TempDir(), "mutant"), Options{})
	if err == nil || !strings.Contains(err.Error(), "too few arguments to function call") || !strings.Contains(err.Error(), "expected 3, have 2") {
		t.Fatalf("drop-count mutant was not rejected for typed arity: %v", err)
	}
	t.Log("drop-count mutant rejected by clang under -Werror: expected 3, have 2")
}
