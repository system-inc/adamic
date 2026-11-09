package lower

import (
	"errors"
	"fmt"
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
