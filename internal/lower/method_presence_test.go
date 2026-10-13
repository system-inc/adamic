package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMethodPresenceRefusesEscapes(t *testing.T) {
	t.Parallel()
	for _, use := range []string{
		"const saved = value.method;",
		"function check(undefined: number | undefined): boolean { if (typeof undefined === 'undefined') { return value.method === undefined; } return false; }",
		"function take(callback: () => string): void {} take(value.method);",
		"function detach(): () => string { return value.method; }",
		"const saved = (value.method || (() => 'fallback'));",
		"const saved = (true && value.method);",
		"function detach(): (() => string) | undefined { return value.method && value.method; }",
	} {
		t.Run(use, func(t *testing.T) {
			_, err := lowerSource(t, "class Value { method(): string { return 'value'; } } const value = new Value();\n"+use)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "unbound-method") {
				t.Fatalf("got %v, want unbound-method", err)
			}
		})
	}
}

func TestMethodPresenceMakesNoFunctionValue(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "class Value { method(): string { return 'value'; } } const value = new Value(); const present = !!value.method;")
	if err != nil {
		t.Fatal(err)
	}
	for _, local := range program.Locals {
		if local.Type == ir.Closure {
			t.Fatalf("presence made a closure local: %+v", local)
		}
	}
	declaration := program.Main[len(program.Main)-1].(ir.Declare)
	outer := declaration.Value.(ir.Unary)
	inner := outer.Operand.(ir.Unary)
	if _, ok := inner.Operand.(ir.MethodPresence); !ok {
		t.Fatalf("got %T, want MethodPresence", inner.Operand)
	}
}

func TestMethodPresenceHostIsNotYet(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "interface Host { nextTick(): void; } declare const process: Host; const present = !!process.nextTick;")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "ambient host method") {
		t.Fatalf("got %v, want the host descriptor gap", err)
	}
}
