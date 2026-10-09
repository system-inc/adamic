package react

import "testing"

// The letter after the prefix is the whole test, which is why "used" and "user" are not hooks.
func TestIsHookName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{"useState", true},
		{"useThing", true},
		{"useX", true},
		{"use", false},
		{"used", false},
		{"user", false},
		{"usestate", false},
		{"handleUse", false},
		{"", false},
	}
	for _, testCase := range cases {
		if got := IsHookName(testCase.name); got != testCase.want {
			t.Fatalf("%q: want %v, got %v", testCase.name, testCase.want, got)
		}
	}
}

// The test is unicode rather than an ASCII range, and that is upstream's behavior rather than a
// generalization: oxlint uses `char::is_uppercase` while ESLint's regex is ASCII-only.
func TestIsLikelyComponentName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want bool
	}{
		{"Thing", true},
		{"T", true},
		{"thing", false},
		{"", false},
		{"_Thing", false},
		// A non-ASCII capital. An ASCII-range check answers false here and this answers true,
		// which is the difference between the two upstreams.
		{"Ünnamed", true},
		{"ünnamed", false},
	}
	for _, testCase := range cases {
		if got := IsLikelyComponentName(testCase.name); got != testCase.want {
			t.Fatalf("%q: want %v, got %v", testCase.name, testCase.want, got)
		}
	}
}
