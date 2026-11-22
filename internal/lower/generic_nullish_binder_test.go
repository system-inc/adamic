package lower

import (
	"strings"
	"testing"
)

// A binder held beside only undefined, null or void, callback: () => T | undefined, is absorbed when the call
// gives it undefined, so inference reads nothing back; the checker's mapper holds it, and the instance is
// lowered with it (binderBesideNullish, #hm6sc4z, rulings.md row 68). Each row agrees with Node on the
// JavaScript backend and natively under the sanitizers: an explicit undefined, a callback returning only
// undefined, and tsc's forEach shape with one. Mutant: the gate answering false, and every row stops at main's
// 'a function returning T | undefined'.
func TestNullishBinderIsReadFromTheMapper(t *testing.T) {
	t.Parallel()
	const visit = "function visit<T>(callback: () => T | undefined): T | undefined {\n    return callback();\n}\n"
	const forEach = "function forEach<T, U>(array: readonly T[] | undefined, callback: (element: T, index: number) => U | undefined): U | undefined {\n    if (array !== undefined) {\n        for (let i = 0; i < array.length; i++) {\n            const element = array[i];\n            if (element === undefined) continue;\n            const result = callback(element, i);\n            if (result) {\n                return result;\n            }\n        }\n    }\n    return undefined;\n}\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"explicit undefined", visit + "console.log(typeof visit<undefined>(() => undefined));\n", "undefined\n"},
		{"callback undefined", visit + "console.log(typeof visit(() => undefined));\n", "undefined\n"},
		{"both instances", visit + "console.log(typeof visit<string>(() => 'present'));\nconsole.log(typeof visit<undefined>(() => undefined));\n", "string\nundefined\n"},
		{"forEach undefined", forEach + "console.log(String(forEach([1, 2], () => undefined)));\n", "undefined\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				program := lowersAndAgreesWithNodeAs(t, probe.source, extension)
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
				}
			})
		}
	}
}

// A T given string | undefined reaching a string operation inside the instance is held as possibly
// undefined (row 68): the template reads undefined as Node does, and never hands on a null string.
func TestNullishBinderGivenOptionalStringReachesAStringOperation(t *testing.T) {
	t.Parallel()
	const source = "function shout<T>(get: () => T | undefined): string {\n    const v = get();\n    return `${v}`.toUpperCase();\n}\nconsole.log(shout<string | undefined>(() => undefined));\nconsole.log(shout<string | undefined>(() => 'ab'));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, source, extension)
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "UNDEFINED\nAB\n" {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "UNDEFINED\nAB\n")
			}
		})
	}
}

// Bound to void, as tsc's forEach is by a callback that returns nothing, the instance meets the generic-at-
// void rule (#p7t8tmh), which Lower and Optimizer agreed on #hm6sc4z: forEach's body reads the result, so the
// call is refused there, honestly, where main said NotYet at the declaration.
func TestNullishBinderGivenVoidMeetsTheVoidReadRule(t *testing.T) {
	t.Parallel()
	const source = "function forEach<T, U>(array: readonly T[], callback: (element: T) => U | undefined): U | undefined {\n    for (const element of array) {\n        const result = callback(element);\n        if (result) {\n            return result;\n        }\n    }\n    return undefined;\n}\nlet total = 0;\nforEach([1, 2], (n) => { total += n; });\nconsole.log(String(total));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			if err == nil || !strings.Contains(err.Error(), "a generic instantiated at void whose body reads the value") {
				t.Fatalf("want the void read refused, got %v", err)
			}
		})
	}
}
