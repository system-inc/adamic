package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionalAssignHiddenKeys(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const actual = {slot:'hidden'}; const hidden: {} = actual; const result: {slot?:number} = Object.assign({}, hidden); console.log(String(result.slot));`,
		`const actual = {slot:'hidden'}; const hidden: {} = actual; const result: {slot?:number} = Object.assign({}, {...hidden}); console.log(String(result.slot));`,
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "no-optional-widening") {
			t.Fatalf("hidden keys must remain refused: %v", err)
		}
	}
}
