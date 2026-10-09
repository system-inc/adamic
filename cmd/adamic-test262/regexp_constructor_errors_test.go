package main

import (
	"testing"
)

func TestRegExpConstructorDiagnosticGuards(t *testing.T) {
	t.Parallel()
	valid := `try { throw new Test262Error('failure: ' + (new RegExp('[z-a]').exec('a'))); } catch (e) { assert.sameValue(e instanceof SyntaxError, true, 'syntax'); }`
	got := adaptRegExpConstructorErrors(valid)
	if got.Counts["regex-unreachable-constructor-diagnostic"] != 1 {
		t.Fatal("intrinsic SyntaxError diagnostic not normalized")
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
			t.Fatalf("unproven diagnostic normalized: %s", got.Source)
		}
	}
}

// Not parallel: the engine reuses original Node and adapted program artifacts.
func TestRegExpConstructorDiagnosticsNode(t *testing.T) {
	e, err := prepare("../..", "testdata/regexp-constructor-errors", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e.adapt = true
	got, err := e.runFilter("built-ins/RegExp", 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Pass != 3 || got.Fail != 0 || got.Refused != 0 || got.NotTypescript != 0 || got.Crashed != 0 {
		t.Fatalf("constructor errors differ from Node: %+v", got)
	}
}
