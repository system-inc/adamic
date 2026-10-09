package literal_test

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/ecmascript/literal"
)

// expectMapping asserts the whole mapping, which is what a caller indexes with. Asserting only the
// length would pass for a mapping that is the right size and points at the wrong bytes, which is the
// defect class this function exists to prevent.
func expectMapping(t *testing.T, raw string, cooked string, want []int) {
	t.Helper()
	got := literal.CookedToRaw(raw, cooked)
	if want == nil {
		if got != nil {
			t.Errorf("CookedToRaw(%q, %q) = %v, want nil", raw, cooked, got)
		}
		return
	}
	if got == nil {
		t.Fatalf("CookedToRaw(%q, %q) = nil, want %v", raw, cooked, want)
	}
	if len(got) != len(want) {
		t.Fatalf("CookedToRaw(%q, %q) = %v (len %d), want %v (len %d)",
			raw, cooked, got, len(got), want, len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("CookedToRaw(%q, %q)[%d] = %d, want %d",
				raw, cooked, index, got[index], want[index])
		}
	}
}

// TestIdentityMapping covers the common case: no escapes, so every cooked byte sits where it does in
// the raw text. The slice carries one extra entry so the end offset of the last character is
// addressable.
func TestIdentityMapping(t *testing.T) {
	t.Parallel()
	expectMapping(t, "abc", "abc", []int{0, 1, 2, 3})
	expectMapping(t, "", "", []int{0})
}

// TestEscapesCollapse covers a raw escape that produces fewer cooked bytes than it occupies. Every
// cooked byte an escape produced points back at the escape's first byte, so a span starting inside a
// decoded escape starts at the backslash.
func TestEscapesCollapse(t *testing.T) {
	t.Parallel()
	// `\n` is two raw bytes producing one cooked byte.
	expectMapping(t, `a\nb`, "a\nb", []int{0, 1, 3, 4})
	// `\x41` is four raw bytes producing one cooked `A`.
	expectMapping(t, `\x41b`, "Ab", []int{0, 4, 5})
	// `\u0041` is six raw bytes producing one cooked `A`.
	expectMapping(t, `\u0041b`, "Ab", []int{0, 6, 7})
}

// TestMultiByteRunes covers the case that made every mapping fail before it was fixed: a raw
// character carries its own UTF-8 width, and taking one byte while comparing against a multi-byte
// cooked rune desynced the walk on texts that actually agreed.
func TestMultiByteRunes(t *testing.T) {
	t.Parallel()
	// A two-byte rune, raw and cooked alike.
	expectMapping(t, "á", "á", []int{0, 0, 2})
	// A backslash before a multi-byte character escapes that one character. `\👍` is five raw bytes
	// producing four cooked ones.
	got := literal.CookedToRaw(`\👍`, "👍")
	if got == nil {
		t.Fatal(`CookedToRaw(\👍) = nil, want a mapping`)
	}
	for index := 0; index < 4; index++ {
		if got[index] != 0 {
			t.Errorf("an escaped astral character mapped cooked byte %d to raw %d, want 0", index, got[index])
		}
	}
}

// TestSurrogatePairSpendsMoreRawThanCooked pins the case that makes the strict "raw must also be
// exhausted" check wrong.
//
// `👍` is twelve raw bytes producing one four-byte rune, so cooked legitimately runs out
// with raw still to go. A terminal check requiring both to finish rejects the exact case the
// misleading-character-class rule exists for.
func TestSurrogatePairSpendsMoreRawThanCooked(t *testing.T) {
	t.Parallel()
	if got := literal.CookedToRaw(`👍`, "👍"); got == nil {
		t.Error("a surrogate pair was refused; cooked finishing before raw is legitimate here")
	}
}

// TestLineContinuationIsRefused is the live defect this function's history records.
//
// A trailing line continuation produces no cooked bytes at all, so the walk stops with cooked
// exhausted and a raw tail never compared against anything. The original terminal check confirmed
// only that the cooked text had finished, and returned a complete, well-formed mapping whose every
// offset past the run was wrong. Found by the second caller probing this against its own corpus
// before building on it, which is the argument for the function having a home of its own.
func TestLineContinuationIsRefused(t *testing.T) {
	t.Parallel()
	// `a  b\<newline>` cooks to `a  b`, leaving a raw tail that produced nothing.
	expectMapping(t, "a  b\\\n", "a  b", nil)
	expectMapping(t, "a  b\\\r\n", "a  b", nil)
	// A trailing lone backslash escapes nothing and is refused for the same reason.
	expectMapping(t, `ab\`, "ab", nil)
}

// TestDisagreementIsRefused covers the honest-failure path. A cooked value the walk cannot reconcile
// with the raw text returns nil rather than a wrong answer, because a finding at a guessed span
// points somewhere real and wrong, which is worse than no finding.
func TestDisagreementIsRefused(t *testing.T) {
	t.Parallel()
	expectMapping(t, "abc", "xyz", nil)
	expectMapping(t, "ab", "abcdef", nil)
}

// TestLegacyOctalEscapes covers a shape upstream's corpus uses deliberately: a variation selector
// whose `0` is spelled `\60` so the escape only appears after the string is cooked.
func TestLegacyOctalEscapes(t *testing.T) {
	t.Parallel()
	// `\101` is four raw bytes producing one cooked `A`.
	expectMapping(t, `\101b`, "Ab", []int{0, 4, 5})
	// A leading digit above three caps the run at two digits, so `\41` is three raw bytes.
	expectMapping(t, `\41b`, "!b", []int{0, 3, 4})
}
