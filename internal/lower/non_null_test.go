package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNonNullAssertionLowersToNullishPanic(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "const map = new Map<string, number>();\nconst value = map.get('a')!;\n")
	if err != nil {
		t.Fatal(err)
	}
	declaration := program.Main[1].(ir.Declare)
	check, ok := declaration.Value.(ir.Coalesce)
	if !ok || check.Type() != ir.Number || check.Value.Type() != ir.MaybeNumber {
		t.Fatalf("want checked maybe-number unwrap, got %#v", declaration.Value)
	}
	message := check.Panic.(ir.StringConstant)
	if got := program.Strings[message.Index]; got != "non-null assertion failed: map.get('a')! is null or undefined" {
		t.Fatalf("unexpected panic text %q", got)
	}
}

func TestNonNullAssertionOnPresentTypeIsErased(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "const value = 0!;\n")
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
