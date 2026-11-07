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
