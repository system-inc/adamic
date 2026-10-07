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
		test := classified{Program: program(one.source)}
		if got := e.attempt(test); got.Kind != one.want {
			t.Fatalf("constructor harness: want %s, got %+v", one.want, got)
		}
	}
}
