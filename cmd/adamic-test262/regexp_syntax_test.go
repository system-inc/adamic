package main

import "testing"

func TestRegExpSyntaxAssertionGuards(t *testing.T) {
	t.Parallel()
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
