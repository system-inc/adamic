package regexp

import "testing"

func TestV8LiteralTestInputProof(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		pattern, flags, input string
		admitted              bool
	}{
		{`[\q{Ss|x}]`, "iv", "sſ", true},
		{`[\q{Ss|x}]`, "iv", "SS", true},
		{`[\q{Ss|x}]`, "iv", "🌍ss", true},
		{`[\q{Ss|x}]`, "iv", "x", true},
		{`[\q{Ss|x}]`, "iv", "", true},
		{`[\q{Ss|x}]`, "iv", "X", false},
		{`[\q{Ss|x}]`, "iv", "Xss", false},
		{`[\q{x}]`, "iv", "x", true},
		{`[\q{x}]`, "iv", "X", false},
		{`[\q{a}]`, "iv", "A", false},
		{`[\q{Ss|x}]`, "giv", "sſ", false},
		{`[\q{Ss|x}]`, "ivy", "sſ", false},
		{`[\q{Ss|x}]+`, "iv", "sſ", false},
		{`[\q{Ss|x}]a`, "iv", "sſa", false},
		{`[^\q{x}]`, "iv", "a", false},
		{`[\q{Ss|x}]`, "iv", string([]byte{0xed, 0xa0, 0x80}), false},
	} {
		t.Run(test.pattern+"/"+test.flags+"/"+test.input, func(t *testing.T) {
			program, err := Compile(test.pattern, test.flags)
			if err != nil {
				t.Fatal(err)
			}
			if program.NativeCompatibility() == nil {
				t.Fatal("control must retain its pattern-wide refusal")
			}
			declarations, err := NativeLiteralTestDeclarations(test.pattern, test.flags, test.input, "proof")
			if test.admitted {
				if err != nil || declarations == "" {
					t.Fatalf("safe literal over-refused: %v", err)
				}
			} else if err == nil {
				t.Fatal("unproven literal admitted")
			}
			if _, err := program.NativeDeclarations("still_refused"); err == nil {
				t.Fatal("literal proof changed the general refusal")
			}
		})
	}
}
