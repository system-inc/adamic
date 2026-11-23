package lower

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestOptionalIntrinsicKeepsNullableStringRefusal(t *testing.T) {
	t.Parallel()
	for _, receiver := range []string{"string | null", "string | null | undefined"} {
		for _, method := range []string{"trim", "toUpperCase"} {
			source := fmt.Sprintf("function text(value: %s): %s { return value; }\nconsole.log(`${text(null)?.%s()}`);", receiver, receiver, method)
			_, err := lowerSource(t, source)
			var missing *NotYet
			if !errors.As(err, &missing) || missing.What != "a call through ?. (an optional call)" {
				t.Fatalf("%s %s: want unchanged optional-call refusal, got %v", receiver, method, err)
			}
		}
	}
}

func TestOptionalIntrinsicKeepsMaybeScalarRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ receiver, call, want string }{
		{"number", "toFixed(2)", "a call through ?. (an optional call)"},
		{"number", "toString()", "toString on a number | undefined (function source text and mixed value dispatch are not represented)"},
		{"boolean", "toString()", "toString on a boolean | undefined (function source text and mixed value dispatch are not represented)"},
	} {
		t.Run(probe.receiver+"_"+probe.call, func(t *testing.T) {
			t.Parallel()
			source := fmt.Sprintf("function text(value: %s | undefined): string | undefined { return value?.%s; }\nconsole.log(`${text(undefined)}`);", probe.receiver, probe.call)
			_, err := lowerSourceAs(t, source, ".ts")
			var missing *NotYet
			if !errors.As(err, &missing) || missing.What != probe.want {
				t.Fatalf("want unchanged maybe-scalar refusal %q, got %v", probe.want, err)
			}
		})
	}
}

func TestOptionalIntrinsicBooleanPlainToStringLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, "function text(value: boolean): string { return value.toString(); }\nconsole.log(text(true));", ".ts"); err != nil {
		t.Fatal(err)
	}
}

func TestOptionalIntrinsicArgumentsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"delete", "foreach", "effects", "result"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/library_optional_receiver_" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			lowersAndAgreesWithNode(t, string(source))
		})
	}
}

// A spread argument and an optional call of the library method keep optionalCall's stop. A chain read on
// past the call lowers under optionalChain's one guard (#cdpxjmb): property and call, and in TypeScript
// non_null_property, whose ! .a still refuses; and optional_property's second ?. owns its guard around
// Set.add's result, whose ?.size is number | undefined (#8zcm7g4). The oracle's
// testdata/optional_chain/size.a holds the add rows to Node.
func TestOptionalIntrinsicArgumentBoundariesRefuse(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"property", "optional_property", "non_null_property", "call", "spread", "function"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/optional_receiver_refused/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var refused *NotYet
			switch name {
			case "property", "optional_property", "call":
				if err != nil {
					t.Fatalf("want it lowered under the chain's guard, got %v", err)
				}
			case "non_null_property":
				var assertion *Refused
				if !errors.As(err, &assertion) || assertion.What != "the non-null assertion !" {
					t.Fatalf("want the non-null assertion refusal, got %v", err)
				}
				// TypeScript admits !, which runs inside the chain's guard.
				if _, err = lowerTypeScriptAssertionSource(t, string(source)); err != nil {
					t.Fatalf("want the chain through ! lowered under its guard, got %v", err)
				}
			default:
				if !errors.As(err, &refused) || (refused.What != "a call through ?. (an optional call)" && refused.What != "an optional chain longer than one step") {
					t.Fatalf("want optional call/chain refusal, got %v", err)
				}
			}
		})
	}
}

// Library calls have ?. only on the receiver; optional declared methods have it
// on the call. Neither handler may take the other's shape before optionalCall.
// Output alone cannot identify the dispatcher; IR counts pin both call protocols.
func TestOptionalIntrinsicAndDeclaredMethodDispatch(t *testing.T) {
	t.Parallel()
	const source = `interface Runner { run?(value: string): void; }
function library(receiver: Set<string> | undefined): void { receiver?.delete('x'); }
function declared(receiver: Runner | undefined): void { receiver?.run?.('called'); }
const entries = new Set<string>();
entries.add('x');
library(entries);
library(undefined);
console.log(String(entries.has('x')));
declared({ run: (value: string): void => console.log(value) });
declared(undefined);
declared({});
`
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	deletes := 0
	visit := func(node any) bool {
		if _, ok := node.(ir.MapDelete); ok {
			deletes++
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	if deletes != 1 || optionalCalls(program) != 1 {
		t.Fatalf("want one library deletion and one optional declared-method call, got %d and %d", deletes, optionalCalls(program))
	}
	lowersAndAgreesWithNode(t, source)
}
