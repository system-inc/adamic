package main

import (
	"testing"
)

func TestRegExpConstructorDiagnosticGuards(test *testing.T) {
	test.Parallel()
	valid := `try { throw new Test262Error('failure: ' + (new RegExp('[z-a]').exec('a'))); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`
	got := adaptRegExpConstructorErrors(valid)
	if got.Counts["regex-unreachable-constructor-diagnostic"] != 1 {
		test.Fatal("intrinsic SyntaxError diagnostic not normalized")
	}
	for _, source := range []string{
		`try { throw new Test262Error(effect() + new RegExp('[z-a]')); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`,
		`try { effect(); throw new Test262Error('failure: ' + new RegExp('[z-a]')); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`,
		`try { throw new Test262Error('failure: ' + new RegExp('a')); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`,
		`try { throw new Test262Error('failure: ' + new RegExp('[z-a]')); } catch (e) { assert.sameValue(e instanceof SyntaxError, false, 'syntax'); }`,
		`try { throw new Test262Error('failure: ' + new RegExp('[z-a]')); } catch (e) { effect(); assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`,
		`try { throw new Test262Error('failure: ' + new RegExp('[z-a]')); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); } effect();`,
		`try { throw new Test262Error('failure: ' + new RegExp('a{9223372036854775808,9223372036854775807}')); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`,
	} {
		if got := adaptRegExpConstructorErrors(source); got.Source != source {
			test.Fatalf("unproven diagnostic normalized: %s", got.Source)
		}
	}
}
