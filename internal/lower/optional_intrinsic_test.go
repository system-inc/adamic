package lower

import (
	"errors"
	"fmt"
	"os"
	"testing"
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
			if name == "optional_property" {
				// The second ?. owns its guard. The current property representation,
				// rather than the call continuation rule, still refuses Set.add's result.
				if !errors.As(err, &refused) || refused.What != "optional chaining to .size on a value" {
					t.Fatalf("want the optional size property refusal, got %v", err)
				}
			} else if name == "non_null_property" {
				var assertion *Refused
				if !errors.As(err, &assertion) || assertion.What != "the non-null assertion !" {
					t.Fatalf("want the non-null assertion refusal, got %v", err)
				}
				// TypeScript admits !, but property lowering still refuses the chain.
				_, err = lowerTypeScriptAssertionSource(t, string(source))
				if !errors.As(err, &refused) || refused.What != "an optional chain longer than one step" {
					t.Fatalf("want the continuation refusal through !, got %v", err)
				}
			} else if !errors.As(err, &refused) || (refused.What != "a call through ?. (an optional call)" && refused.What != "an optional chain longer than one step") {
				t.Fatalf("want optional call/chain refusal, got %v", err)
			}
		})
	}
}
