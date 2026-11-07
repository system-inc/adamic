package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRegExpLiteralTestInputRefusals(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const input = 'sſ'; /[\q{Ss|x}]/iv.test(input);`,
		`function input(): string { return 'sſ'; } /[\q{Ss|x}]/iv.test(input());`,
		`const regex = /[\q{Ss|x}]/iv; regex.test('sſ');`,
		`/[\q{Ss|x}]/iv.exec('sſ');`,
		`/[\q{Ss|x}]/iv.test('X');`,
		`/[\q{Ss|x}]/iv.test('\u0058');`,
		`/[\q{Ss|x}]/giv.test('sſ');`,
		`/[\q{Ss|x}]/ivy.test('sſ');`,
		`/[\q{Ss|x}]/iv.test('s' + 'ſ');`,
		`/[\q{Ss|x}]/iv.test('\ud800');`,
	} {
		_, err := lowerSource(test, source)
		var refusal *NotYet
		if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "singleton") {
			test.Fatalf("want named singleton refusal for %s: %v", source, err)
		}
	}
}
