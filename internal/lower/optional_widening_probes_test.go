package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalWideningAdditionalAliases(t *testing.T) {
	t.Parallel()
	const hidden = `const wide: { name: string; extra: number } = { name: 'a', extra: 1 };
const view: { name: string } = wide;
`
	for _, probe := range []struct{ name, source string }{
		{"hidden spread alias", hidden + `const copied = { ...view }; const narrow: { name: string; extra?: string } = copied; console.log(narrow.extra ?? 'none');`},
		{"dynamic key alias", `function key(): string { return 'extra'; } const source = { name: 'a', [key()]: 1 }; const narrow: { name: string; extra?: string } = source; console.log(narrow.extra ?? 'none');`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want optional widening refused, got %v", err)
			}
			for _, want := range []string{"adamic/no-optional-widening", "extra", "declare", "source", "object"} {
				if !strings.Contains(refused.Error(), want) {
					t.Errorf("refusal %q does not contain %q", refused.Error(), want)
				}
			}
		})
	}
}
