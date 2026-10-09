package main

import "testing"

func TestRegExpSyntaxAssertionGuards(t *testing.T) {
	for _, source := range []string{
		`globalThis["RegExp"] = factory; assert.throws(SyntaxError, () => RegExp("[z-a]"));`,
		`const ctor = RegExp; assert.throws(SyntaxError, () => RegExp("[z-a]"));`,
		`Object.defineProperty(globalThis, "RegExp", descriptor); assert.throws(SyntaxError, () => RegExp("[z-a]"));`,
		`const SyntaxError = Error; assert.throws(SyntaxError, () => RegExp("[z-a]"));`,
		`const RegExp = factory; assert.throws(SyntaxError, () => RegExp("[z-a]"));`,
		`RegExp.prototype.exec = fn; assert.throws(SyntaxError, () => RegExp("[z-a]"));`,
		`assert.throws(SyntaxError, () => { effect(); RegExp("[z-a]"); });`,
		`assert.throws(SyntaxError, () => RegExp(effect()));`,
		`assert.throws(SyntaxError, () => RegExp("a"));`,
		`assert.throws(SyntaxError, () => RegExp("a{9223372036854775808,9223372036854775807}"));`,
		`assert.throws(SyntaxError, () => RegExp("[z-a]"), effect());`,
	} {
		if got := adaptRegExpSyntaxAssertions(source); got.Source != source {
			t.Fatalf("unproven assertion adapted: %s", got.Source)
		}
	}
}

// Not parallel: the runner reuses its program and original Node artifacts.
func TestRegExpSyntaxAssertionsNode(t *testing.T) {
	e, err := prepare("../..", "testdata/regexp-syntax", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.adapt = true
	got, err := e.runFilter("built-ins/RegExp", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Pass != 2 || got.Fail != 0 || got.Refused != 0 || got.Crashed != 0 {
		t.Fatalf("syntax assertions differ from Node: %+v", got)
	}
}
