package lower

import "testing"

// A class member the helper can't answer stops only where the hint's order reaches it (#1rxjj3m, Language's
// p18 on #e0cnaa9). Under String's hint a class with a valueOf getter and no toString answers
// Object.prototype.toString's "[object Object]", and Node never reads the getter, so it agrees; so does a
// valueOf getter beside a toString method that answers. A toString getter is what the string hint reads
// first, and it still stops, exit 70, on both backends. On main each agreeing row stopped at run time in
// words no ruling lists. Mutant: every unanswered member stops up front, as before, and the agreeing rows
// stop.
func TestToPrimitiveStopsOnlyWhereReached(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		{"valueOf getter, no toString", "class Late { get valueOf(): () => number { return (): number => 7; } }\nconst v: unknown = new Late();\nconsole.log(String(v));\n", "[object Object]\n"},
		{"valueOf getter beside toString", "class Late { get valueOf(): () => number { return (): number => 7; } toString(): string { return 'late'; } }\nconst v: unknown = new Late();\nconsole.log(String(v));\n", "late\n"},
		{"valueOf field, no toString", "class Late { readonly valueOf = (): number => 9; }\nconst v: unknown = new Late();\nconsole.log(String(v));\n", "[object Object]\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
	t.Run("toString getter", func(t *testing.T) {
		t.Parallel()
		program, err := lowerSourceAs(t, "class Late { get toString(): () => string { return (): string => 'g'; } }\nconst v: unknown = new Late();\nconsole.log(String(v));\n", ".ts")
		if err != nil {
			t.Fatalf("want it lowered to its run-time stop, got %v", err)
		}
		brandedStopsProgram(t, program, "adamic: panic: converting an object whose class has a toString or valueOf accessor to a string is not yet lowered")
	})
}
