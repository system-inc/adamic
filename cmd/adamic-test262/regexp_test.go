package main

import (
	"strings"
	"testing"
)

func TestRegExpAdaptGuards(t *testing.T) {
	t.Parallel()
	for _, source := range []string{`const RegExp = factory; /a/;`, `function f(RegExp) { return /a/; }`, `RegExp = factory; /a/;`, `RegExp.prototype.exec = fn; /a/;`, `const f = RegExp => /a/;`} {
		if got := adaptRegExpLiterals(source); got.Source != source {
			t.Fatalf("shadowed intrinsic rewritten: %s", got.Source)
		}
	}
	if got := adaptRegExpLiterals(`const re = /(?i:a)\//u;`); got.Counts["regex-literal-constructor"] != 1 || !strings.Contains(got.Source, `new RegExp("(?i:a)\\/", "u")`) {
		t.Fatalf("literal spelling changed: %+v", got)
	}
	for _, source := range []string{
		`let expected = ['a']; expected.index = effect(); expected.input = 'a'; assert.sameValue(expected[0], 'a');`,
		`let expected = ['a']; expected.index = 0; expected.input = 'a'; consume(expected);`,
		`let expected = ['a']; expected.index = 0; expected.input = 'a'; expected = ['b'];`,
		`let match = /a/.exec('a'); try { assert.sameValue(match.index, 0); } catch (e) {}`,
		`let match = /a/.exec('a'); match = /b/.exec('b'); assert.sameValue(match.index, 0);`,
		`testPropertyOfStrings({regExp: /a/, 'nonMatchStrings': ['a']});`,
		`testPropertyOfStrings({regExp: /a/, ...args});`,
		`object.testPropertyOfStrings({regExp: /a/});`,
		`function testPropertyOfStrings(args) {} testPropertyOfStrings({regExp: /a/});`,
		`try { const text = char.codePointAt(0).toString(16); } catch (error) {}`,
		`const errors = []; errors.push(1); errors.join(',');`,
		`const errors = []; errors.push('x'); consume(errors);`,
	} {
		got := adaptRegExp(source)
		if got.Source != source {
			t.Fatalf("uncertain program adapted:\n%s\n%s", source, got.Source)
		}
	}
	got := adaptRegExp(`assert.sameValue(/a/.exec('b'), null, 'missing'); assert.notSameValue(null, /a/.exec('a'));`)
	if got.Counts["regex-null-assert"] != 2 || !strings.Contains(got.Source, " === null") || !strings.Contains(got.Source, " !== null") {
		t.Fatalf("null assertions: %+v", got)
	}
}
