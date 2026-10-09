package unused_exports

import "testing"

// TestIntentionalMarker proves the marker is read where written and not read where it is merely
// resembled, so a word like `cohere-keeper` cannot silently suppress a finding.
func TestIntentionalMarker(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name       string
		text       string
		wantMarked bool
		wantReason string
	}{
		{"bare marker", "// cohere-keep\nexport function f() {}", true, ""},
		{"marker with reason", "// cohere-keep: public API for the SDK\nexport function f() {}", true, "public API for the SDK"},
		{"marker with reason no colon", "// cohere-keep held for an external consumer\nexport function f() {}", true, "held for an external consumer"},
		{"block comment", "/* cohere-keep: parked deliberately */\nexport const x = 1;", true, "parked deliberately"},
		{"no marker", "// an ordinary comment\nexport function f() {}", false, ""},
		{"marker is a prefix of another word", "// cohere-keeper is not a directive\nexport function f() {}", false, ""},
		// Both edges of the word are guarded. `unverify-keeping` embeds the marker but is not one,
		// and reading it as a directive would suppress a finding nobody meant to suppress.
		{"marker inside a longer word", "// unverify-keeping\nexport function f() {}", false, ""},
		// The marker only counts in a comment. A string or identifier that happens to contain it is
		// source, not a directive.
		{"marker in code rather than a comment", "export const label = 'cohere-keep';", false, ""},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := findIntentionalMarker(testCase.text)
			if got.Present != testCase.wantMarked {
				t.Fatalf("marked = %v, want %v", got.Present, testCase.wantMarked)
			}
			if got.Present && got.Text != testCase.wantReason {
				t.Fatalf("reason = %q, want %q", got.Text, testCase.wantReason)
			}
		})
	}
}
