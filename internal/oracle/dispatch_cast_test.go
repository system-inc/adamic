package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Ruling 17's road (a) (#d1chbm7): each admitted dispatch program lowers (internal/lower/dispatch_lower.go), and
// TestNativeAgreesWithNode holds both backends to Node on it. t00.ts, t05.ts, t22.ts and t24.ts are the brief's
// programs byte for byte; t06.ts is t19.ts with its line 1 `enum Kind { A, B, C }`, so kind C, which no entry names,
// answers "none" through the presence test; minus-zero.ts is t00.ts with its line 15's kind -0, which finds kind 0's
// entry as Node's table[-0] does; optional-call.ts is t00.ts with its line 13 `return fn?.(node) ?? "none";`, the
// optional call with no presence test.
func init() {
	for _, name := range []string{"t00.ts", "t05.ts", "t06.ts", "t22.ts", "t24.ts", "minus-zero.ts", "optional-call.ts"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/dispatch_cast/" + name, true, false})
	}
}

// dispatchAnyWords is today's stop at a declined lookup's assertion, refuseAnyRelation's, where key 0's slot holds
// NodeA: the Record's any parameter related to the table's.
const dispatchAnyWords = "a member typed any (0(0)) related to NodeA at the assertion site"

// Every program that keeps today's stop stops at its lookup's assertion, in today's words, never earlier and never
// at the dispatch: refuseAnyRelation's deferred stop at the asserted table (column 14, after a tab and
// `const fn = (`), or, for a program a stronger refusal stops first, that refusal where it stands (r03.ts's
// primitive cast and g01.ts's and g02.ts's `as unknown as` pairs, as t03.ts's is; c13-prototype.ts's Object.assign
// at :16, which reaches no dispatch). g03.ts, as Language wrote it, is a tsc refusal (TS2322 at :6:7: its Partial
// annotation over constructor and toString), so it pins nothing here and has no row. The twins are t00.ts with one line changed or inserted, as internal/lower/dispatch_cast_test.go writes
// them; t21.ts is g04.ts in the ruled shape, t25.ts is t24.ts with its view and call cut to one parameter, and the
// escapes are t00.ts with fn called through .call, .apply or .bind, stored and called later, or returned.
// Mutant, red: any condition of dispatch_cast.go answering "" for its own pin, which then lowers or stops elsewhere.
func TestDispatchCastKeepsTodaysStops(t *testing.T) {
	t.Parallel()
	union := strings.Replace(dispatchAnyWords, "NodeA", "NodeA | NodeC", 1)
	for _, probe := range []struct {
		file, words, location string
		refused               bool
	}{
		{"t01.ts", dispatchAnyWords, "t01.ts:13:14", false},
		{"t02.ts", dispatchAnyWords, "t02.ts:11:14", false},
		{"t04.ts", dispatchAnyWords, "t04.ts:12:14", false},
		{"t10.ts", dispatchAnyWords, "t10.ts:13:14", false},
		{"t11.ts", dispatchAnyWords, "t11.ts:13:14", false},
		{"t12.ts", dispatchAnyWords, "t12.ts:13:14", false},
		{"t13.ts", dispatchAnyWords, "t13.ts:13:14", false},
		{"t14.ts", dispatchAnyWords, "t14.ts:13:14", false},
		{"t15.ts", dispatchAnyWords, "t15.ts:12:14", false},
		{"t16.ts", dispatchAnyWords, "t16.ts:12:14", false},
		{"t17.ts", dispatchAnyWords, "t17.ts:12:14", false},
		{"t18.ts", dispatchAnyWords, "t18.ts:12:14", false},
		{"t21.ts", dispatchAnyWords, "t21.ts:14:14", false},
		{"t19.ts", "a member typed any (a(0)) related to NodeA at the assertion site", "t19.ts:13:14", false},
		// t25.ts's view takes one parameter where the slot takes two, a cast castRefusal refuses at the lookup first.
		{"t25.ts", "a cast from Table to Record<Kind, (node: any) => string>", "t25.ts:12:14", true},
		{"c06.ts", union, "c06.ts:13:14", false},
		{"c08.ts", union, "c08.ts:15:14", false},
		{"c09.ts", union, "c09.ts:15:14", false},
		{"c10.ts", union, "c10.ts:15:14", false},
		{"c11.ts", dispatchAnyWords, "c11.ts:14:14", false},
		{"c12-trailing.ts", union, "c12-trailing.ts:13:14", false},
		{"c14-trailing-one-member.ts", dispatchAnyWords, "c14-trailing-one-member.ts:12:14", false},
		{"g05.ts", "a member typed any (1(0)) related to NodeA at the assertion site", "g05.ts:14:14", false},
		{"r00.ts", dispatchAnyWords, "r00.ts:12:14", false},
		{"r01.ts", dispatchAnyWords, "r01.ts:13:14", false},
		{"r02.ts", dispatchAnyWords, "r02.ts:11:14", false},
		{"r03.ts", "a cast from Kind to Kind.A, which holds a primitive no view can check", "r03.ts:12:46", true},
		{"g01.ts", "an `as unknown as` pair from", "g01.ts:10:71", true},
		{"g02.ts", "an `as unknown as` pair from", "g02.ts:8:4", true},
		{"g04.ts", dispatchAnyWords, "g04.ts:13:14", false},
		{"escape-call.ts", dispatchAnyWords, "escape-call.ts:12:14", false},
		{"escape-apply.ts", dispatchAnyWords, "escape-apply.ts:12:14", false},
		{"escape-bind.ts", dispatchAnyWords, "escape-bind.ts:12:14", false},
		{"escape-stored.ts", dispatchAnyWords, "escape-stored.ts:12:14", false},
		{"escape-returned.ts", dispatchAnyWords, "escape-returned.ts:12:14", false},
		{"c13-prototype.ts", "Object.assign on a shape not proven by a plain literal or its const binding", "c13-prototype.ts:16:15", false},
	} {
		t.Run(probe.file, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/dispatch_cast", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			where, what := "", ""
			var notYet *lower.NotYet
			var refused *lower.Refused
			switch {
			case !probe.refused && errors.As(err, &notYet):
				where, what = notYet.Where, notYet.What
			case probe.refused && errors.As(err, &refused):
				where, what = refused.Where, refused.What
			}
			if !strings.Contains(what, probe.words) || !strings.Contains(where, probe.location) {
				t.Fatalf("want located stop %q at %s: %v", probe.words, probe.location, err)
			}
			t.Logf("%v", err)
		})
	}
}
