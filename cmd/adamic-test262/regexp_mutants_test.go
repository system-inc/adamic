package main

import (
	"strings"
	"testing"
)

// These mutations change the executable adaptation, with the native and
// adapted Node executions still agreeing. Only untouched upstream Node
// rejects the changed observation. No compiler or sanitizer failure counts.
func regExpAdapterMutant(test *testing.T, source, old, replacement string) {
	test.Helper()
	e := regExpRunnerEngine(test, "testdata/regexp")
	one := classify("built-ins/RegExp/adapter-mutant.js", source, true)
	if !strings.Contains(one.Program, old) {
		test.Fatalf("mutation site missing: %s", old)
	}
	site := strings.LastIndex(one.Program, old)
	one.Program = one.Program[:site] + replacement + one.Program[site+len(old):]
	got := e.attempt(one)
	if got.Kind != outcomeFail || got.Reason != "original Node test disagrees with adaptation" {
		test.Log(e.compiler.compile(e.work + "/program.a").Stderr)
		test.Fatalf("mutant escaped untouched Node: %+v", got)
	}
}

func TestRegExpMutantLiteralFlags(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(test, `console.log(/a/g.flags);`, `new RegExp("a", "g")`, `new RegExp("a", "i")`)
}
func TestRegExpMutantNullAssertion(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(test, `assert.sameValue(/a/.exec('a'), null);`, ` === null`, ` !== null`)
}
func TestRegExpMutantExpectedMetadata(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		`var expected = ['a']; expected.index = 0; expected.input = 'a'; console.log(expected.input);`,
		`input: 'a'`,
		`input: 'b'`,
	)
}
func TestRegExpMutantMatchRead(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		`var actual = /a/.exec('ba'); console.log(actual.index.toString());`,
		`requiredRegexExec(actual).index`,
		`(requiredRegexExec(actual).index + 1)`,
	)
}
func TestRegExpMutantCodePoint(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		`const symbol = 'a'; console.log(symbol.codePointAt(0).toString(16));`,
		`return point;`,
		`return point + 1;`,
	)
}
func TestRegExpMutantStringCollector(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		`const errors = []; errors.push('a'); console.log(errors.join(','));`,
		`errors.push('a')`,
		`errors.push('b')`,
	)
}
func TestRegExpMutantHarnessOptionalField(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { testPropertyOfStrings({regExp:/a/,expression:'a',matchStrings:['a']}); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');",
		`nonMatchStrings: undefined`,
		`nonMatchStrings: ['a']`,
	)
}
func TestRegExpMutantRangeEndpoint(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nconsole.log(buildString({loneCodePoints:[],ranges:[[48,50]]}));",
		`point <= end`,
		`point < end`,
	)
}
func TestRegExpMutantPropertyEscapes(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { testPropertyEscapes(/^[a-c]$/, 'ad', 'letters'); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');",
		`assert(regExp.test(symbol),`,
		`assert(true,`,
	)
}
func TestRegExpMutantCaptureValidator(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { matchValidator(['x'],0,'a')(/a/.exec('a')); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');",
		`assertCompareArray<string | undefined>(match, expectedEntries, "Match entries");`,
		``,
	)
}

func TestRegExpMutantMatchIndex(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { matchValidator(['a'],1,'a')(/a/.exec('a')); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');",
		`assertSameValue(match.index, expectedIndex, "Match index");`,
		``,
	)
}
func TestRegExpMutantMatchInput(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { matchValidator(['a'],0,'b')(/a/.exec('a')); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');",
		`assertSameValue(match.input, expectedInput, "Match input");`,
		``,
	)
}
func TestRegExpMutantNegativeStrings(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nlet rejected = false; try { testPropertyOfStrings({regExp:/a/,expression:'a',matchStrings:['a'],nonMatchStrings:['x','a']}); } catch { rejected = true; } console.log(rejected ? 'yes' : 'no');",
		`if (regExp.test(negatives.join(""))) {`,
		`if (!regExp.test(negatives.join(""))) {`,
	)
}
func TestRegExpMutantPrintCodePoint(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nconsole.log(printCodePoint(0x1F30D));",
		`.toUpperCase()`,
		`.toLowerCase()`,
	)
}
func TestRegExpMutantPrintStringCodePoints(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		"/*---\nincludes: [regExpUtils.js]\n---*/\nconsole.log(printStringCodePoints('ab'));",
		`formatted.join(" ")`,
		`formatted.join(",")`,
	)
}

// Simulate a bad proof discharging a constructor assertion. Both adapted
// executions succeed; the independent upstream assertion must still reject it.
func regExpConstructorProofMutant(test *testing.T, source string) {
	test.Helper()
	e := regExpRunnerEngine(test, "testdata/regexp")
	one := classify("built-ins/RegExp/proof-mutant.js", source, true)
	one.Program = program(`assertSameValue(1, 1);`)
	one.ConstructorAssertion = false
	if got := e.attempt(one); got.Kind != outcomeFail || got.Reason != "original Node test disagrees with adaptation" {
		test.Fatalf("false constructor proof accepted: %+v", got)
	}
}
func TestRegExpMutantSyntaxAssertionProof(test *testing.T) {
	test.Parallel()
	regExpConstructorProofMutant(test, `assert.throws(SyntaxError, () => RegExp('a'));`)
}
func TestRegExpMutantLegacyConstructorProof(test *testing.T) {
	test.Parallel()
	regExpConstructorProofMutant(
		test,
		`try { throw new Test262Error('failure: ' + (new RegExp('a').exec('a'))); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`,
	)
}

func TestRegExpMutantStringMatchRead(test *testing.T) {
	test.Parallel()
	regExpAdapterMutant(
		test,
		`var actual = 'a'.match(/a/); console.log(actual.length.toString());`,
		`requiredRegexMatch(actual).length`,
		`(requiredRegexMatch(actual).length + 1)`,
	)
}
