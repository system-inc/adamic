package regexp

import (
	"github.com/system-inc/adamic/internal/unicodeproperties"
	"sort"
	"testing"
)

// The sparse compile-time maps must agree with the shared runtime Canonicalize.
func TestSharedCanonicalize(test *testing.T) {
	test.Parallel()
	for character := rune(0); character <= 0x10ffff; character++ {
		index := sort.Search(len(simpleCaseFold), func(index int) bool { return simpleCaseFold[index][0] >= character })
		want := character
		if index < len(simpleCaseFold) && simpleCaseFold[index][0] == character {
			want = simpleCaseFold[index][1]
		}
		if got := unicodeproperties.CanonicalizeUnicode(character); got != want {
			test.Fatalf("Unicode U+%04X: shared=%X local=%X", character, got, want)
		}
		if character <= 0xffff {
			index := sort.Search(
				len(legacyUppercase),
				func(index int) bool { return legacyUppercase[index][0] >= character },
			)
			want = character
			if index < len(legacyUppercase) && legacyUppercase[index][0] == character {
				want = legacyUppercase[index][1]
			}
			if got := rune(unicodeproperties.CanonicalizeLegacy(uint16(character))); got != want {
				test.Fatalf("legacy U+%04X: shared=%X local=%X", character, got, want)
			}
		}
	}
}
