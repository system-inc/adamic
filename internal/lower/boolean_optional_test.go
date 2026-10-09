package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestBooleanJSONRequiresCompleteDataOrigin(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function show(value:{flag?:boolean}):void { JSON.stringify(value); } show({flag:true});`,
		`const source = {flag:true, toJSON:():boolean => false}; const alias:{readonly flag:boolean}=source; JSON.stringify(alias);`,
		`const source = {get flag():boolean { return true; }}; const alias:{readonly flag:boolean}=source; JSON.stringify(alias);`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "JSON.stringify object references") {
			t.Fatalf("unproven JSON origin admitted: %v; source: %s", err, source)
		}
	}
}

func TestBooleanOptionalPrototypeSlotStaysNotYet(t *testing.T) {
	_, err := lowerSource(t, `const value:{__proto__?:boolean}={}; value.__proto__=true;`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "__proto__ setter semantics") {
		t.Fatalf("prototype setter must remain refused: %v", err)
	}
}
