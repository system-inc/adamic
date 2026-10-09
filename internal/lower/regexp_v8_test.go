package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRegExpV8Refusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`console.log( /[\q{a}]/iv.test('A') ? 'yes' : 'no');`,
		`console.log( /[\q{Ss|x}]/iv.test('sſ') ? 'yes' : 'no');`,
		`console.log( new RegExp('(?i:a)[b]','v').test('aB') ? 'yes' : 'no');`,
		`console.log( new RegExp('(?-i:a)[b]','iv').test('aB') ? 'yes' : 'no');`,
		`console.log( new RegExp('(?i:a)\\w','u').test('aK') ? 'yes' : 'no');`,
		`console.log( new RegExp('(?-i:^)\\W','iu').test('K') ? 'yes' : 'no');`,
		`console.log( new RegExp('(?i:x|[^a-z])').test('B') ? 'yes' : 'no');`,
		`console.log( new RegExp('(?i:a)|\\P{Ll}','v').test('Σ') ? 'yes' : 'no');`,
		`console.log( /[\q{ab|a|}]/v.test('a') ? 'yes' : 'no');`,
		`const regex = new RegExp('[\\q{Ss|x}]','iv'); console.log(regex.test('sſ') ? 'yes' : 'no');`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "V8") || !strings.Contains(err.Error(), "ECMA-262 22.2.") {
			t.Fatalf("expected named V8/spec refusal: %s: %v", source, err)
		}
	}
}
