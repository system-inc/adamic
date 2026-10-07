package lower

import (
	"errors"
	"testing"
)

func TestViewPreflightPreservesOtherRefusals(t *testing.T) {
	for _, source := range []string{
		"interface User { readonly name: string } function read(x: unknown): User { return x as User; }",
		"const n = ('wrong' as unknown) as number;",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("want refusal, got %v", err)
		}
		want := "a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
		if got := refused.What + "; " + refused.Fix; got != want {
			t.Fatalf("changed diagnostic: %s", got)
		}
	}
}
