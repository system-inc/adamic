package main

import (
	"strings"
	"testing"
)

// These mutations change the executable adaptation, with the native and
// adapted Node executions still agreeing. Only untouched upstream Node
// rejects the changed observation. No compiler or sanitizer failure counts.
func regExpAdapterMutant(t *testing.T, source, old, replacement string) {
	t.Helper()
	t.Parallel()
	e := regExpRunnerEngine(t, "testdata/regexp")
	one := classify("built-ins/RegExp/adapter-mutant.js", source, true)
	if !strings.Contains(one.Program, old) {
		t.Fatalf("mutation site missing: %s", old)
	}
	site := strings.LastIndex(one.Program, old)
	one.Program = one.Program[:site] + replacement + one.Program[site+len(old):]
	got := e.attempt(one)
	if got.Kind != outcomeFail || got.Reason != "original Node test disagrees with adaptation" {
		t.Log(e.compiler.compile(e.work + "/program.a").Stderr)
		t.Fatalf("mutant escaped untouched Node: %+v", got)
	}
}

func TestRegExpMutantLiteralFlags(t *testing.T) {
	regExpAdapterMutant(t, `console.log(/a/g.flags);`, `new RegExp("a", "g")`, `new RegExp("a", "i")`)
}
func TestRegExpMutantNullAssertion(t *testing.T) {
	regExpAdapterMutant(t, `assert.sameValue(/a/.exec('a'), null);`, ` === null`, ` !== null`)
}
func TestRegExpMutantExpectedMetadata(t *testing.T) {
	regExpAdapterMutant(t, `var expected = ['a']; expected.index = 0; expected.input = 'a'; console.log(expected.input);`, `input: 'a'`, `input: 'b'`)
}
func TestRegExpMutantMatchRead(t *testing.T) {
	regExpAdapterMutant(t, `var actual = /a/.exec('ba'); console.log(actual.index.toString());`, `requiredRegexExec(actual).index`, `(requiredRegexExec(actual).index + 1)`)
}
func TestRegExpMutantCodePoint(t *testing.T) {
	regExpAdapterMutant(t, `const symbol = 'a'; console.log(symbol.codePointAt(0).toString(16));`, `return point;`, `return point + 1;`)
}
func TestRegExpMutantStringCollector(t *testing.T) {
	regExpAdapterMutant(t, `const errors = []; errors.push('a'); console.log(errors.join(','));`, `errors.push('a')`, `errors.push('b')`)
}
func TestRegExpMutantHarnessOptionalField(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { testPropertyOfStrings({regExp:/a/,expression:'a',matchStrings:['a']}); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');", `nonMatchStrings: undefined`, `nonMatchStrings: ['a']`)
}
func TestRegExpMutantRangeEndpoint(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nconsole.log(buildString({loneCodePoints:[],ranges:[[48,50]]}));", `point <= end`, `point < end`)
}
func TestRegExpMutantPropertyEscapes(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { testPropertyEscapes(/^[a-c]$/, 'ad', 'letters'); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');", `assert(regExp.test(symbol),`, `assert(true,`)
}
func TestRegExpMutantCaptureValidator(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { matchValidator(['x'],0,'a')(/a/.exec('a')); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');", `assertCompareArray<string | undefined>(match, expectedEntries, "Match entries");`, ``)
}

func TestRegExpMutantMatchIndex(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { matchValidator(['a'],1,'a')(/a/.exec('a')); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');", `assertSameValue(match.index, expectedIndex, "Match index");`, ``)
}
func TestRegExpMutantMatchInput(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { matchValidator(['a'],0,'b')(/a/.exec('a')); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');", `assertSameValue(match.input, expectedInput, "Match input");`, ``)
}
func TestRegExpMutantNegativeStrings(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { testPropertyOfStrings({regExp:/a/,expression:'a',matchStrings:['a'],nonMatchStrings:['x','a']}); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');", `if (regExp.test(negatives.join(""))) {`, `if (!regExp.test(negatives.join(""))) {`)
}
func TestRegExpMutantPrintCodePoint(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nconsole.log(printCodePoint(0x1F30D));", `.toUpperCase()`, `.toLowerCase()`)
}
func TestRegExpMutantPrintStringCodePoints(t *testing.T) {
	regExpAdapterMutant(t, "/*---\nincludes: [regExpUtils.js]\n---*/\nconsole.log(printStringCodePoints('ab'));", `formatted.join(" ")`, `formatted.join(",")`)
}

// Simulate a bad proof discharging a constructor assertion. Both adapted
// executions succeed; the independent upstream assertion must still reject it.
func regExpConstructorProofMutant(t *testing.T, source string) {
	t.Helper()
	t.Parallel()
	e := regExpRunnerEngine(t, "testdata/regexp")
	one := classify("built-ins/RegExp/proof-mutant.js", source, true)
	one.Program = program(`assertSameValue(1, 1);`)
	one.ConstructorAssertion = false
	if got := e.attempt(one); got.Kind != outcomeFail || got.Reason != "original Node test disagrees with adaptation" {
		t.Fatalf("false constructor proof accepted: %+v", got)
	}
}
func TestRegExpMutantSyntaxAssertionProof(t *testing.T) {
	regExpConstructorProofMutant(t, `assert.throws(SyntaxError, () => RegExp('a'));`)
}
func TestRegExpMutantLegacyConstructorProof(t *testing.T) {
	regExpConstructorProofMutant(t, `try { throw new Test262Error('failure: ' + (new RegExp('a').exec('a'))); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`)
}

func TestRegExpMutantStringMatchRead(t *testing.T) {
	regExpAdapterMutant(t, `var actual = 'a'.match(/a/); console.log(actual.length.toString());`, `requiredRegexMatch(actual).length`, `(requiredRegexMatch(actual).length + 1)`)
}
