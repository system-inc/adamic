package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestNonNullAssertionLowersToNullishPanic(t *testing.T) {
	t.Parallel()
	program, err := lowerTypeScriptAssertionSource(t, "const map = new Map<string, number>();\nconsole.log(`${map.get('a')!}`);\n")
	if err != nil {
		t.Fatal(err)
	}
	var check ir.Coalesce
	found := false
	walk(program.Main, func(node any) bool {
		if value, ok := node.(ir.Coalesce); ok && value.Panic != nil {
			check = value
			found = true
		}
		return true
	})
	ok := found
	if !ok || check.Type() != ir.Number || check.Value.Type() != ir.MaybeNumber {
		t.Fatalf("want checked maybe-number unwrap, got %#v", program.Main)
	}
	message := check.Panic.(ir.StringConstant)
	if got := program.Strings[message.Index]; !strings.HasSuffix(got, ": map.get('a')! is null or undefined") || !strings.HasPrefix(got, "non-null assertion failed at ") {
		t.Fatalf("unexpected panic text %q", got)
	}
}

func TestNonNullAssertionOnPresentTypeIsErased(t *testing.T) {
	t.Parallel()
	program, err := lowerTypeScriptAssertionSource(t, "const value = 0!;\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := program.Main[0].(ir.Declare).Value.(ir.NumberConstant); !ok {
		t.Fatalf("want a plain number, got %#v", program.Main[0])
	}
	if len(program.Strings) != 0 {
		t.Fatalf("redundant assertion emitted panic text: %q", program.Strings)
	}
}

func lowerTypeScriptAssertionSource(t *testing.T, source string) (*ir.Program, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	return Lower(context.Background(), program)
}
