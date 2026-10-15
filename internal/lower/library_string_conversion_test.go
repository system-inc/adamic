package lower

import (
	"errors"
	"testing"
)

func TestStringReferenceAdmissionRefusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function text(value: number[] | undefined): string { return String(value); }",
		"function text(value: Map<string, number> | undefined): string { return String(value); }",
		"function text(value: Map<string, number> | Set<number>): string { return String(value); }",
		"console.log(String([{toString:()=> 'text'}]));",
	} {
		_, err := lowerSource(t, source)
		var refusal *NotYet
		if !errors.As(err, &refusal) {
			t.Fatalf("want a proven refusal for %q, got %v", source, err)
		}
	}
}
