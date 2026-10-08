package main

import "testing"

// Not parallel: the runner reuses one program artifact across attempts.
// This portable constructor test does not depend on the matcher seat's runner
// adaptation tests, which have not landed on main yet.
func TestRegExpProtocolHarnessConstructors(t *testing.T) {
	e, err := prepare("../..", "testdata", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range []struct {
		source string
		want   outcomeKind
	}{
		{`assertThrows("TypeError", () => { throw new TypeError("type"); });`, outcomePass},
		{`assertThrows("TypeError", () => { throw new Error("type"); });`, outcomeFail},
	} {
		test := classified{Program: program(one.source), ConstructorAssertion: true}
		if got := e.attempt(test); got.Kind != one.want {
			t.Fatalf("constructor harness: want %s, got %+v", one.want, got)
		}
	}
}

func TestProtocolAdaptedAbruptConstructor(t *testing.T) {
	e, err := prepare("../..", "testdata/regexp", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`const o = { toString: function() { throw new Test262Error(); } }; assert.throws(Test262Error, function() { String.prototype.codePointAt.call(o, 1); });`,
		`assert.throws(Test262Error, function() { throw new Test262Error(); });`,
	} {
		test := classify("built-ins/String/prototype/codePointAt/abrupt.js", body, true)
		if got := e.attempt(test); got.Kind != outcomePass {
			t.Fatalf("adapted abrupt constructor: %+v", got)
		}
	}
	// The adapted Error assertion would accept this wrong constructor. The
	// untouched Node harness must reject it independently.
	bad := classify("built-ins/String/wrong.js", `if (false) { throw new Test262Error(); } assert.throws(Test262Error, function() { throw new Error(); });`, true)
	if got := e.attempt(bad); got.Kind != outcomeFail {
		t.Fatalf("original Node constructor: %+v", got)
	}
}
