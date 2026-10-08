package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestNonNullAssertionLowersToNullishPanic(t *testing.T) {
	t.Parallel()
	program, err := lowerTypeScriptSource(t, "const map = new Map<string, number>();\nconsole.log(`${map.get('a')!}`);\n")
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
	if got := program.Strings[message.Index]; got != "non-null assertion failed: map.get('a')! is null or undefined" {
		t.Fatalf("unexpected panic text %q", got)
	}
}

func TestNonNullAssertionOnPresentTypeIsErased(t *testing.T) {
	t.Parallel()
	program, err := lowerTypeScriptSource(t, "const value = 0!;\n")
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

func TestNonNullAssertionIsRefusedInAdamic(t *testing.T) {
	for _, source := range []string{"const value = 0!;", "const map = new Map<string, number>(); const value = map.get('a')!;"} {
		_, err := lowerSource(t, source)
		var refusal *Refused
		if !errors.As(err, &refusal) || refusal.What != "the non-null assertion !" {
			t.Fatalf("want .a assertion refusal, got %v", err)
		}
	}
}

func TestObjectLiteralNeverMethodUsesOrdinaryCall(t *testing.T) {
	_, err := lowerSource(t, "function fail(label: string): never { throw new Error(label); } const system = { abort(label: string): never { return fail(label); } }; try { system.abort('method'); } catch {}")
	if err != nil {
		t.Fatal(err)
	}
}

func lowerTypeScriptSource(t *testing.T, source string) (*ir.Program, error) {
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
