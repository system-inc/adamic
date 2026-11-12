package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's getSupportedExtensions (utilities.ts:9993) casts an extension inside a callback past its first
// return. Overload 1's caller passes no extra extensions, so no path under its parameters reaches the
// cast, and the body proof reads only what a path reaches (#z5d0jye step 2b, Correctness's census of
// step 2). Mutant: censusReturnProof walking statements no path reaches, and the declaration is refused.
func TestOverloadProofSkipsCodeNoPathReaches(t *testing.T) {
	t.Parallel()
	source := `type Extension = ".ts" | ".js";
const tsExtensions: readonly (readonly Extension[])[] = [[".ts"]];
const allExtensions: readonly (readonly Extension[])[] = [[".ts"], [".js"]];
interface Options { readonly allowJs?: boolean }
interface Extra { readonly extension: string }
function supported(options?: Options): readonly (readonly Extension[])[];
function supported(options?: Options, extra?: readonly Extra[]): readonly (readonly string[])[];
function supported(options?: Options, extra?: readonly Extra[]): readonly (readonly string[])[] {
    const needJs = options !== undefined && options.allowJs === true;
    if (!extra || extra.length === 0) {
        return needJs ? allExtensions : tsExtensions;
    }
    return extra.map(x => [x.extension as string]);
}
const all: readonly (readonly Extension[])[] = supported({ allowJs: true });
console.log(String(all.length));
console.log(String(supported(undefined, [{ extension: ".vue" }]).length));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n1\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 2 and 1, as Node prints", code, stdout, stderr)
	}
}

// A cast in a statement a path does reach still stops the proof, wherever it could type a value the
// proof reads: here a local a return spreads. The promise is an array, which no call-site check serves.
// Mutant: the reached-statement cast test dropped, and the declaration is accepted.
func TestOverloadProofRefusesACastAPathReaches(t *testing.T) {
	t.Parallel()
	source := `function numbers(xs: readonly (string | number)[]): readonly number[];
function numbers(xs: readonly (string | number)[]): readonly (string | number)[];
function numbers(xs: readonly (string | number)[]): readonly (string | number)[] { const ys = xs as readonly number[]; return [...ys]; }
console.log(String(numbers([1]).length));
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of numbers result") {
		t.Fatalf("want overload 1 of numbers refused, got %v", err)
	}
}
