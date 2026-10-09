package fuzz

import "testing"

// stopped is an outcome where Node ran to the end, printing nodeStdout, and both of Adamic's backends
// stopped with exit 70 after printing stdout, writing stderr.
func stopped(nodeStdout string, stdout string, stderr string) Outcome {
	adamic := Run{Stdout: []byte(stdout), Stderr: []byte(stderr), ExitCode: 70}
	return Outcome{Node: Run{Stdout: []byte(nodeStdout)}, Native: adamic, Backend: adamic}
}

// A shared stop is an inserted check only when its panic line is one of the checks Adamic inserts
// where JavaScript runs on. Anything else both backends stop at while Node finishes is a finding,
// however little they printed first: an empty stdout is a prefix of everything.
func TestJudgeReadsThePanicLine(t *testing.T) {
	t.Parallel()
	checkout := &Checkout{}
	for _, test := range []struct {
		name    string
		outcome Outcome
		want    Verdict
	}{
		// What JavaScript throws, where Node didn't: both backends stopped where nothing should.
		{"a JavaScript error Node never threw", stopped("", "", "adamic: panic: TypeError: Cannot set properties of undefined\n"), Finding},
		{"a JavaScript error after shared output", stopped("a\nb\n", "a\n", "adamic: panic: TypeError: Cannot read properties of undefined (reading 'value')\n"), Finding},
		{"a ReferenceError Node never threw", stopped("a\n", "", "adamic: panic: ReferenceError: Cannot access 'total' before initialization\n"), Finding},
		{"a program's own panic Node never reached", stopped("a\n", "a\n", "adamic: panic: boom\n"), Finding},
		{"a runtime failure", stopped("a\n", "", "adamic: panic: compiler bug: a function ended without returning\n"), Finding},

		// The checks Adamic inserts where JavaScript goes on with a value the type can't hold.
		{"undefined put back after a narrowing", stopped("a\nundefined\n", "a\n", "adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back\n"), Checked},
		{"null put back after a narrowing", stopped("a\nnull\n", "a\n", "adamic: panic: null where the checker narrowed it away: a call since the narrowing put it back\n"), Checked},
		{"map over an array that shrank", stopped("1,,\n", "", "adamic: panic: map: the array shrank while it was being mapped\n"), Checked},
		{"find over an array that shrank", stopped("-1\n", "", "adamic: panic: find: the array shrank while it was being searched\n"), Checked},
		{"findIndex over an array that shrank", stopped("-1\n", "", "adamic: panic: findIndex: the array shrank while it was being searched\n"), Checked},
		{"find typed search value", stopped("undefined\n", "", "adamic: panic: find: index 2 is undefined; element type number does not admit undefined\n"), Checked},
		{"findIndex typed search value", stopped("undefined\n", "", "adamic: panic: findIndex: index 2 is undefined; element type number does not admit undefined\n"), Checked},
		{"findLast typed search value", stopped("undefined\n", "", "adamic: panic: findLast: index 2 is undefined; element type number does not admit undefined\n"), Checked},
		{"findLastIndex typed search value", stopped("undefined\n", "", "adamic: panic: findLastIndex: index 2 is undefined; element type number does not admit undefined\n"), Checked},
		{"a write past the end", stopped("a\n4\n", "a\n", "adamic: panic: index 3 is outside an array of length 2\n"), Checked},
		{"a write at a fractional index", stopped("2\n", "", "adamic: panic: index 0.5 is outside an array of length 2\n"), Checked},
		{"a cast the discriminant refuses", stopped("circle\n", "", "adamic: panic: cast failed: this Shape is not a Square\n"), Checked},

		// An inserted check's words, but not a shared stop Node's run can explain.
		{"a check after output Node never printed", stopped("a\n", "b\n", "adamic: panic: index 3 is outside an array of length 2\n"), Finding},
		{"a check with more after it", stopped("a\n", "", "adamic: panic: index 3 is outside an array of length 2 and more\n"), Finding},
	} {
		got := checkout.judge(test.outcome, "", "")
		if got.Verdict != test.want {
			t.Errorf("%s: %s (%s), want %s", test.name, got.Verdict, got.Key, test.want)
		}
	}

	// The two backends stopped at different checks: they disagree with each other, whatever the words.
	split := stopped("a\n", "", "adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back\n")
	split.Backend.Stderr = []byte("adamic: panic: index 3 is outside an array of length 2\n")
	if got := checkout.judge(split, "", ""); got.Verdict != Finding {
		t.Errorf("backends stopped at different checks: %s (%s), want %s", got.Verdict, got.Key, Finding)
	}
}
