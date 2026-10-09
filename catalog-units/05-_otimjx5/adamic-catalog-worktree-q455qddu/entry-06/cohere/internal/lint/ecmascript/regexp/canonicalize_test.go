package regexp

import (
	"slices"
	"testing"
)

// Every expectation here is what `new RegExp("^"+x+"$", flags).test(y)` answers
// in JavaScript, which is Canonicalize comparing the two characters. The two
// readings are tested together because the whole point of the flag is that they
// disagree.
func TestCanonicalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		left  rune
		right rune
		plain bool // same under `i`
		uFlag bool // same under `iu`
	}{
		{name: "ascii", left: 'a', right: 'A', plain: true, uFlag: true},
		// A final sigma, a small sigma and a capital sigma are one character
		// under either reading.
		{name: "final sigma", left: 0x03C2, right: 0x03A3, plain: true, uFlag: true},
		{name: "small sigma", left: 0x03C3, right: 0x03A3, plain: true, uFlag: true},
		{name: "both sigmas", left: 0x03C2, right: 0x03C3, plain: true, uFlag: true},
		// Without `u` a character outside ASCII never canonicalizes into
		// ASCII. Folding has no such rule, so these join up under `u`.
		{name: "kelvin sign", left: 'k', right: 0x212A, plain: false, uFlag: true},
		{name: "long s", left: 's', right: 0x017F, plain: false, uFlag: true},
		// Neither character is ASCII, so the rule above never applies and only
		// the choice of mapping decides.
		{name: "angstrom sign", left: 0x00E5, right: 0x212B, plain: false, uFlag: true},
		{name: "ohm sign", left: 0x03C9, right: 0x2126, plain: false, uFlag: true},
		// The uppercase of an eszett is two characters, so it stays put; the
		// capital eszett folds onto the small one.
		{name: "eszett", left: 0x00DF, right: 0x1E9E, plain: false, uFlag: true},
		// Three pairs whose members fold together while neither is the other's
		// case, so only the `u` reading joins them.
		{name: "iota with dialytika", left: 0x1FD3, right: 0x0390, plain: false, uFlag: true},
		{name: "upsilon with dialytika", left: 0x1FE3, right: 0x03B0, plain: false, uFlag: true},
		{name: "st ligature", left: 0xFB05, right: 0xFB06, plain: false, uFlag: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, mode := range []struct {
				unicodeMode bool
				want        bool
				flags       string
			}{{false, test.plain, "i"}, {true, test.uFlag, "iu"}} {
				got := Canonicalize(test.left, mode.unicodeMode) == Canonicalize(test.right, mode.unicodeMode)
				if got != mode.want {
					t.Errorf("/%s/ Canonicalize(%U)==Canonicalize(%U) is %v, want %v (got %U and %U)",
						mode.flags, test.left, test.right, got, mode.want,
						Canonicalize(test.left, mode.unicodeMode), Canonicalize(test.right, mode.unicodeMode))
				}
			}
		})
	}
}

func TestCaseEquivalents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		r           rune
		unicodeMode bool
		want        []rune
	}{
		{name: "ascii letter", r: 'k', want: []rune{'K', 'k'}},
		{name: "sigma", r: 0x03A3, want: []rune{0x03A3, 0x03C2, 0x03C3}},
		// Alone in its group, so there is nothing to widen to.
		{name: "kelvin sign", r: 0x212A, want: nil},
		{name: "long s", r: 0x017F, want: nil},
		{name: "digit", r: '4', want: nil},
		// Under `u` the same two characters pull `k` and `s` into their group.
		{name: "ascii letter under u", r: 'k', unicodeMode: true, want: []rune{'K', 'k', 0x212A}},
		{name: "kelvin sign under u", r: 0x212A, unicodeMode: true, want: []rune{'K', 'k', 0x212A}},
		{name: "long s under u", r: 0x017F, unicodeMode: true, want: []rune{'S', 's', 0x017F}},
		{name: "digit under u", r: '4', unicodeMode: true, want: nil},
		// A pair that folds together with no case mapping between its members.
		{name: "st ligature", r: 0xFB05, want: nil},
		{name: "st ligature under u", r: 0xFB05, unicodeMode: true, want: []rune{0xFB05, 0xFB06}},
		{name: "iota with dialytika under u", r: 0x1FD3, unicodeMode: true, want: []rune{0x0390, 0x1FD3}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := CaseEquivalents(test.r, test.unicodeMode)
			if !slices.Equal(got, test.want) {
				t.Errorf("CaseEquivalents(%U, %v) = %U, want %U", test.r, test.unicodeMode, got, test.want)
			}
		})
	}
}

// TestCaseEquivalenceGroupsAgree covers the two views of the same table: every
// member of every group has to look the group back up, and canonicalize onto
// the same character as the rest of it.
func TestCaseEquivalenceGroupsAgree(t *testing.T) {
	t.Parallel()
	for _, unicodeMode := range []bool{false, true} {
		groups := CaseEquivalenceGroups(unicodeMode)
		if len(groups) == 0 {
			t.Fatalf("CaseEquivalenceGroups(%v) is empty", unicodeMode)
		}
		for _, members := range groups {
			if len(members) < 2 {
				t.Errorf("group %U has nothing to widen to", members)
			}
			canonical := Canonicalize(members[0], unicodeMode)
			for _, member := range members {
				if Canonicalize(member, unicodeMode) != canonical {
					t.Errorf("group %U holds %U, which canonicalizes to %U not %U",
						members, member, Canonicalize(member, unicodeMode), canonical)
				}
				if !slices.Equal(CaseEquivalents(member, unicodeMode), members) {
					t.Errorf("CaseEquivalents(%U, %v) = %U, want the group %U",
						member, unicodeMode, CaseEquivalents(member, unicodeMode), members)
				}
			}
		}
	}
}

// TestCanonicalizeUnicode17 covers what the delta in [unicode17] is for:
// JavaScript compares each of the pairs it names equal under a regexp's `iu`
// flags, and Go's own tables would compare them unequal. The `i` flag on its
// own canonicalizes one UTF-16 code unit at a time, so it joins the pairs
// inside the basic plane and leaves the two supplementary-plane scripts apart —
// which is checked too, because the delta must not reach further than
// JavaScript does.
func TestCanonicalizeUnicode17(t *testing.T) {
	t.Parallel()
	for _, pair := range unicode17Pairs {
		lower, upper := pair[0], pair[1]
		group := []rune{min(lower, upper), max(lower, upper)}
		joinsWithoutU := lower <= 0xFFFF

		if got := Canonicalize(lower, true); got != group[0] {
			t.Errorf("Canonicalize(%U, true) = %U, want %U", lower, got, group[0])
		}
		if got := Canonicalize(upper, true); got != group[0] {
			t.Errorf("Canonicalize(%U, true) = %U, want %U", upper, got, group[0])
		}
		if got := CaseEquivalents(lower, true); !slices.Equal(got, group) {
			t.Errorf("CaseEquivalents(%U, true) = %U, want %U", lower, got, group)
		}

		wantCanonical, wantEquivalents := lower, []rune(nil)
		if joinsWithoutU {
			wantCanonical, wantEquivalents = upper, group
		}
		if got := Canonicalize(lower, false); got != wantCanonical {
			t.Errorf("Canonicalize(%U, false) = %U, want %U", lower, got, wantCanonical)
		}
		if got := Canonicalize(upper, false); got != upper {
			t.Errorf("Canonicalize(%U, false) = %U, want it to stand for itself", upper, got)
		}
		if got := CaseEquivalents(lower, false); !slices.Equal(got, wantEquivalents) {
			t.Errorf("CaseEquivalents(%U, false) = %U, want %U", lower, got, wantEquivalents)
		}
	}
}

// unicode17Pairs are the {lower, upper} pairs Unicode 16 and 17 added on top of the edition Go
// shipped when this parser was written, spelled out here rather than read from a table.
//
// Only case pairs belong here. The delta also carried three fold-only pairs, U+0390/U+1FD3,
// U+03B0/U+1FE3 and U+FB05/U+FB06, which have no case mapping in either direction and fold together
// instead. Including them here fails, correctly: Canonicalize without the `u` flag leaves a
// character with no uppercase standing for itself, so a fold pair asserted as a case pair asks for
// a mapping that does not exist. That failure is how this list was found to be wrong.
//
// They used to come from an internal delta package that carried the mappings Go's tables lacked.
// That package deleted itself on Go 1.27, which ships Unicode 17.0.0: measured across all 116 runes
// it special-cased, it returned exactly what the standard library returns, with a corrupted control
// confirming the comparison could see a difference. So the data moved into the test rather than
// leaving with it, because the behaviour it pins is the parser's, not the delta's: JavaScript
// compares each pair equal under `iu`, and a toolchain whose tables regressed would break this
// whether or not a delta package existed to blame.
var unicode17Pairs = [][2]rune{
	{0x019B, 0xA7DC}, {0x0264, 0xA7CB}, {0x1C8A, 0x1C89}, {0xA7CD, 0xA7CC},
	{0xA7CF, 0xA7CE}, {0xA7D3, 0xA7D2}, {0xA7D5, 0xA7D4}, {0xA7DB, 0xA7DA},
}
