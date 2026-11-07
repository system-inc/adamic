package regexp

import "testing"

func TestV8LiteralTestInputProof(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
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
		test.Run(testCase.pattern+"/"+testCase.flags+"/"+testCase.input, func(test *testing.T) {
			program, compileError := Compile(testCase.pattern, testCase.flags)
			if compileError != nil {
				test.Fatal(compileError)
			}
			if program.NativeCompatibility() == nil {
				test.Fatal("control must retain its pattern-wide refusal")
			}
			declarations, nativeLiteralTestDeclarationsError := NativeLiteralTestDeclarations(
				testCase.pattern,
				testCase.flags,
				testCase.input,
				"proof",
			)
			if testCase.admitted {
				if nativeLiteralTestDeclarationsError != nil || declarations == "" {
					test.Fatalf("safe literal over-refused: %v", nativeLiteralTestDeclarationsError)
				}
			} else if nativeLiteralTestDeclarationsError == nil {
				test.Fatal("unproven literal admitted")
			}
			if _, nativeDeclarationsError := program.NativeDeclarations(
				"still_refused",
			); nativeDeclarationsError == nil {
				test.Fatal("literal proof changed the general refusal")
			}
		})
	}
}
