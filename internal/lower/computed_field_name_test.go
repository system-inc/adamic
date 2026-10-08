package lower

import (
	"errors"
	"testing"
)

func TestComputedFieldNameGapsStayExplicit(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function key(): 'label' { console.log('key'); return 'label'; } const value = { [key()]: 1 }; console.log(Object.keys(value).join(','));",
		"const enum Key { Left = 3, Right = 4 } const value = { [Key.Left + Key.Right]: 1 }; console.log(Object.keys(value).join(','));",
	} {
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) || gap.What != "a computed field name" {
			t.Fatalf("got %v, want the computed-field gap", err)
		}
	}
}

func TestComputedDuplicateFieldStaysExplicit(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "const value = { ['a' + 'b']: 1, ['ab']: 2 }; console.log(Object.keys(value).join(','));")
	var gap *NotYet
	if !errors.As(err, &gap) || gap.What != "a repeated object field" {
		t.Fatalf("got %v, want the repeated-field gap", err)
	}
}

func TestComputedFieldStorageNamesStayExplicit(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"computed NUL", `const value={['a'+'\0']:1}; console.log(Object.keys(value).join(','));`, "a computed field name"},
		{"literal NUL", `const value={'a\0':1}; console.log(Object.keys(value).join(','));`, "a field name native storage cannot hold"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || gap.What != probe.reason {
				t.Fatalf("got %v, want the field storage gap %s", err, probe.reason)
			}
		})
	}
}
