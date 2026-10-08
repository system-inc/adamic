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
