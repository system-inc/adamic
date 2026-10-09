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

func TestOptionalDeletionRequiresPlainStorage(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class Source { slot?:number; } const source = new Source(); const alias:{slot?:number} = source; delete alias.slot;`,
		`function remove(source:{slot?:number}):void { delete source.slot; } remove({});`,
		`class Source { slot?:number; } const source = new Source(); const alias = source as {slot?:number}; delete alias.slot;`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "proven plain data origin") {
			t.Fatalf("plain storage boundary: %v", err)
		}
	}
}
