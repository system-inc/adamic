package lower

import (
	"strings"
	"testing"
)

// A lookup whose value is held as a bare reference other than an object, where its type holds both null
// and undefined, is NotYet: one bare pointer can't hold a missing key apart from a stored null (#92cnfpf).
// On main, fxspptb_8e67677_null_undefined_inline.a printed "absent: undefined=false null=true" natively
// and "undefined=false null=false" on the JavaScript backend, where Node prints "undefined=true
// null=false", for a Map<string, RegExpExecArray | null>'s get of a missing key and for an element read
// past the end of a (RegExpExecArray | null)[]. Every bare reference kind but an object and a union is
// judged, never a list of them (Compiler's must). Mutants: the guard back to object-held values alone, and
// the guard back to a list of kinds, and each lowers.
func TestNullableBareLookupStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"map get of a missing key", "const seen = new Map<string, RegExpExecArray | null>();\nconsole.log(String(seen.get('absent') === undefined));\n"},
		{"element read past the end", "const found: (RegExpExecArray | null)[] = [];\nfound.push(/z/.exec('abc'));\nconsole.log(String(found[3] === undefined));\n"},
		{"map of arrays or null", "const seen = new Map<string, number[] | null>();\nconsole.log(String(seen.get('absent') === null));\n"},
		{"map of maps or null", "const seen = new Map<string, Map<string, number> | null>();\nconsole.log(String(seen.get('absent') === undefined));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				if err == nil || !strings.Contains(err.Error(), "a nullable object lookup without separate null/undefined tags") {
					t.Fatalf("want the nullable lookup's NotYet, got %v", err)
				}
			})
		}
	}
}

// The guard asks only a type holding both: a lookup in a map of arrays with no null still tells a missing
// key by its empty pointer, and agrees with Node on the JavaScript backend and natively. A string isn't
// held bare where its type holds null, so an element read past the end of a (string | null)[] keeps null
// and undefined apart and agrees too (J-m0bev1r on main, J-t81cd7q at the head).
func TestNullableBareLookupWithoutNullAgrees(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, "const seen = new Map<string, number[]>();\nseen.set('one', [1]);\nconsole.log(String(seen.get('absent') === undefined) + ' ' + String(seen.get('one') === undefined));\n", ".a")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "true false\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "true false\n")
	}
}

func TestNullableBareLookupStringArrayAgrees(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, "const found: (string | null)[] = [null, 'a'];\nconsole.log(String(found[3] === undefined) + ' ' + String(found[3] === null));\nconsole.log(String(found[0] === undefined) + ' ' + String(found[0] === null));\n", ".a")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "true false\nfalse true\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "true false\nfalse true\n")
	}
}
