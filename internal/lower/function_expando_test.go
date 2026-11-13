package lower

import (
	"errors"
	"strings"
	"testing"
)

// Mutant run and restored: returning the old field-store NotYet instead of Refused
// fails these pins and conformance's refusals/function_expando.a (#4g1avf7).
func TestFunctionExpandoRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, function, property string }{
		{"declaration", "function greet(): string { return 'hi'; }\ngreet.loud = true;\nconsole.log(String(greet.loud));\n", "greet", "loud"},
		{"arrow", "const greet = () => 'hi';\ngreet.loud = true;\nconsole.log(String(greet.loud));\n", "greet", "loud"},
		{"expression", "const greet = function(): string { return 'hi'; };\ngreet.loud = true;\nconsole.log(String(greet.loud));\n", "greet", "loud"},
		{"element", "function greet(): string { return 'hi'; }\ngreet['loud'] = true;\nconsole.log(String(greet.loud));\n", "greet", "loud"},
		{"arrow element", "const greet = () => 'hi';\ngreet['loud'] = true;\nconsole.log(String(greet.loud));\n", "greet", "loud"},
		{"names", "function announce(): string { return 'hi'; }\nannounce.volume = 2;\nconsole.log(String(announce.volume));\n", "announce", "volume"},
		{"literal typed key", "const key = 'loud'; function greet(): string { return 'hi'; }\ngreet[key] = true;\nconsole.log(String(greet.loud));\n", "greet", "loud"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "main" + extension + ":2:1: Adamic 0.1 refuses an expando: adding " + probe.property + " to function " + probe.function + " after it's made; declare " + probe.property + " on an object with a call method, or put " + probe.function + " and " + probe.property + " in a namespace (adamic/no-expando)"
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want a refusal ending %q", err, want)
				}
			})
		}
	}
}

func TestFunctionExpandoFixes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"namespace", "function greet(): string { return 'hi'; }\nnamespace greet { export const loud = true; }\nconsole.log(greet());\nconsole.log(String(greet.loud));\n"},
		{"call method", "const greet = { call(): string { return 'hi'; }, loud: true };\nconsole.log(greet.call());\nconsole.log(String(greet.loud));\n"},
		{"declared field write", "const greet = { call: () => 'hi', loud: false };\ngreet.loud = true;\nconsole.log(greet.call());\nconsole.log(String(greet.loud));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				lowersAndAgreesWithNodeAs(t, probe.source, extension)
			})
		}
	}
}

func TestFunctionExpandoBypassesStayRefused(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			for _, probe := range []struct{ source, want string }{
				{"function greet(): string { return 'hi'; }\n(greet as any).loud = true;\n", "a cast the runtime can't check"},
				{"function greet(): string { return 'hi'; }\n// @ts-ignore\ngreet.loud = true;\n", "@ts-ignore suppression directive"},
				{"function greet(): string { return 'hi'; }\nconsole.log(greet.name);\n", "inherited library member name read as an own field"},
			} {
				_, err := lowerSourceAs(t, probe.source, extension)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.want) {
					t.Fatalf("got %v, want the existing refusal saying %q", err, probe.want)
				}
			}
		})
	}
}
