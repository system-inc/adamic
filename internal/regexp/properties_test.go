package regexp

import (
	"github.com/system-inc/adamic/internal/unicodeproperties"
	"sort"
	"testing"
)

// The sparse compile-time maps must agree with the shared runtime Canonicalize.
func TestSharedCanonicalize(t *testing.T) {
	for c := rune(0); c <= 0x10ffff; c++ {
		i := sort.Search(len(simpleCaseFold), func(i int) bool { return simpleCaseFold[i][0] >= c })
		want := c
		if i < len(simpleCaseFold) && simpleCaseFold[i][0] == c {
			want = simpleCaseFold[i][1]
		}
		if got := unicodeproperties.CanonicalizeUnicode(c); got != want {
			t.Fatalf("Unicode U+%04X: shared=%X local=%X", c, got, want)
		}
		if c <= 0xffff {
			j := sort.Search(len(legacyUppercase), func(i int) bool { return legacyUppercase[i][0] >= c })
			want = c
			if j < len(legacyUppercase) && legacyUppercase[j][0] == c {
				want = legacyUppercase[j][1]
			}
			if got := rune(unicodeproperties.CanonicalizeLegacy(uint16(c))); got != want {
				t.Fatalf("legacy U+%04X: shared=%X local=%X", c, got, want)
			}
		}
	}
}
