package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRegExpAdaptGuards(t *testing.T) {
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

// Not parallel: the engine reuses one program artifact, and mutation checks
// deliberately change that program while retaining an untouched original.
func TestRegExpRunnerNode(t *testing.T) {
	e, err := prepare("../..", "testdata/regexp", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.adapt = true
	serial, err := e.runFilter("built-ins/RegExp", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	e.jobs = 2
	parallel, err := e.runFilter("built-ins/RegExp", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(serial, parallel) || parallel.Pass != 3 {
		t.Fatalf("serial/parallel mismatch: %+v / %+v", serial, parallel)
	}
	e.jobs = 1
	serialLimit, err := e.runFilter("built-ins/RegExp", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	e.jobs = 2
	parallelLimit, err := e.runFilter("built-ins/RegExp", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(serialLimit, parallelLimit) || parallelLimit.Pass != 1 || parallelLimit.Unrun != 2 {
		t.Fatalf("attempt limit changed: %+v / %+v", serialLimit, parallelLimit)
	}
	for _, name := range []string{"utilities", "metadata", "construction"} {
		path := "built-ins/RegExp/" + name + ".js"
		data, err := os.ReadFile(filepath.Join(e.test262, "test", path))
		if err != nil {
			t.Fatal(err)
		}
		one := classify(path, string(data), true)
		got := e.attempt(one)
		if got.Kind != outcomePass {
			t.Fatalf("%s: %s: %s", name, got.Kind, got.Reason)
		}
	}
	// A weakened assertion makes both adapted executions pass. Only the
	// independent upstream Node run rejects this false pass.
	one := classify("built-ins/RegExp/false-pass.js", `assert.sameValue(1, 2);`, true)
	one.Program = program(`assertSameValue(1, 1);`)
	if got := e.attempt(one); got.Kind != outcomeFail || !strings.Contains(got.Reason, "original Node") {
		t.Fatalf("false pass accepted: %+v", got)
	}
	// Upstream Node succeeds here, but the old shared harness would accept
	// an Error where the test requires TypeError. That is not a sound pass.
	one = classify("built-ins/RegExp/wrong-constructor.js", `assert.throws(TypeError, () => { throw new TypeError("type"); });`, true)
	one.Program = program(`assertThrows("TypeError", () => { throw new Error("type"); });`)
	if got := e.attempt(one); got.Kind != outcomeFail {
		t.Fatalf("wrong constructor accepted: %+v", got)
	}
	// An unexpected capture must not be ignored by the adapted validator.
	one = classify("built-ins/RegExp/wrong-capture.js", "/*---\nincludes: [regExpUtils.js]\n---*/\nmatchValidator(['x'], 0, 'a')(/a/.exec('a'));", true)
	if got := e.attempt(one); got.Kind == outcomePass || got.Kind == outcomeRefused || got.Kind == outcomeCrashed {
		t.Fatalf("wrong capture was not checked at execution: %+v", got)
	}
}
