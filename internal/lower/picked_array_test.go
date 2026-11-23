package lower

import (
	"strings"
	"testing"
)

const pickedArrayD = "interface D { readonly text: string }\n"
const pickedArrayShow = "console.log(String(all.length) + (all.at(-1)?.text ?? 'none'));\n"

// A Pick of an array whose every key is one of Array's methods is held as that array (#kr1v0tn A, Lower's
// ruling), so tsc's JSON helpers' `Pick<DiagnosticWithLocation[], "push">` errors take the array itself. At
// every position a value is held, it agrees with Node on the JavaScript backend and natively under the
// sanitizers, which no object view over a value held as an array would. Mutant: pickedArray answering nil,
// and each stops at the call as main does.
func TestPickedArrayAgrees(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stdout string }{
		{"parameter", pickedArrayD + "function collect(errors: Pick<D[], 'push'>, text: string): void { errors.push({ text }); }\nconst all: D[] = [];\ncollect(all, 'a');\ncollect(all, 'b');\n" + pickedArrayShow, "2b\n"},
		{"generic parameter", pickedArrayD + "function collect<T>(errors: Pick<T[], 'push'>, item: T): void { errors.push(item); }\nconst all: D[] = [];\ncollect(all, { text: 'a' });\n" + pickedArrayShow, "1a\n"},
		{"field", pickedArrayD + "interface Sink { readonly errors: Pick<D[], 'push'> }\nconst all: D[] = [];\nconst sink: Sink = { errors: all };\nsink.errors.push({ text: 'f' });\n" + pickedArrayShow, "1f\n"},
		{"local", pickedArrayD + "const all: D[] = [];\nconst errors: Pick<D[], 'push'> = all;\nerrors.push({ text: 'l' });\n" + pickedArrayShow, "1l\n"},
		{"return", pickedArrayD + "const all: D[] = [];\nfunction sink(): Pick<D[], 'push'> { return all; }\nsink().push({ text: 'r' });\n" + pickedArrayShow, "1r\n"},
		{"array element", pickedArrayD + "const all: D[] = [];\nconst sinks: Pick<D[], 'push'>[] = [all];\nconst first = sinks[0];\nif (first !== undefined) { first.push({ text: 'e' }); }\n" + pickedArrayShow, "1e\n"},
		{"with undefined", pickedArrayD + "const all: D[] = [];\nfunction collect(errors: Pick<D[], 'push'> | undefined): void { errors?.push({ text: 'u' }); }\ncollect(all);\ncollect(undefined);\n" + pickedArrayShow, "1u\n"},
		{"push and pop", pickedArrayD + "const all: D[] = [{ text: 'x' }];\nfunction take(stack: Pick<D[], 'push' | 'pop'>): void { stack.push({ text: 'y' }); stack.pop(); stack.push({ text: 'z' }); }\ntake(all);\n" + pickedArrayShow, "2z\n"},
		{"readonly", pickedArrayD + "function find(xs: Pick<readonly D[], 'indexOf'>, d: D): number { return xs.indexOf(d); }\nconst one: D = { text: 'q' };\nconsole.log(String(find([one], one)));\n", "0\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
			}
		})
	}
}

// A Pick is structural, so an object literal or a class instance with a push method fits its type too. Each
// stops where it goes into the slot held as an array, beside a union with undefined and at a generic
// entry, where T is instantiated (Lower's should). On the one-doorway head
// before the fit check, each lowered and clang refused its native build (J-jjmfxnc). Mutant: the fit check
// dropped, and each lowers.
func TestPickedArrayEntryStops(t *testing.T) {
	t.Parallel()
	const fake = "const fake = { push(...items: D[]): number { return items.length; } };\n"
	for _, probe := range []struct{ name, source string }{
		{"object literal", pickedArrayD + "function collect(errors: Pick<D[], 'push'>): void { errors.push({ text: 'o' }); }\n" + fake + "collect(fake);\nconsole.log('done');\n"},
		{"class instance", pickedArrayD + "class Fake { push(...items: D[]): number { return items.length; } }\nfunction collect(errors: Pick<D[], 'push'>): void { errors.push({ text: 'c' }); }\ncollect(new Fake());\nconsole.log('done');\n"},
		{"object literal with undefined", pickedArrayD + "function collect(errors: Pick<D[], 'push'> | undefined): void { errors?.push({ text: 'o' }); }\n" + fake + "collect(fake);\nconsole.log('done');\n"},
		// Lower's should: the stop holds where T is instantiated, not only at a concrete slot.
		{"generic entry", pickedArrayD + "function collect<T>(errors: Pick<T[], 'push'>, item: T): void { errors.push(item); }\n" + fake + "collect(fake, { text: 'g' });\nconsole.log('done');\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), pickedArrayFitStop) {
				t.Fatalf("want %q, got %v", pickedArrayFitStop, err)
			}
		})
	}
}

// A Pick that isn't one of an array's methods alone keeps the shape it had: length beside a method, a write
// to a method, or a Pick of a Map each stops in today's words, and a Pick of length alone is a plain object
// whose one field agrees with Node, as on main.
func TestPickedArrayOutsideStays(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, stop string }{
		{"push and length", pickedArrayD + "function add(xs: Pick<D[], 'push' | 'length'>): number { xs.push({ text: 'a' }); return xs.length; }\nconst all: D[] = [];\nconsole.log(String(add(all)));\n", "inherited library member push read as an own field"},
		{"write to a method", pickedArrayD + "function swap(errors: Pick<D[], 'push'>): void { errors.push = (...items: D[]): number => items.length; }\nconst all: D[] = [];\nswap(all);\nconsole.log('done');\n", "assigning a field of a value"},
		{"a Map's methods", "function put(m: Pick<Map<string, number>, 'set'>): void { m.set('a', 1); }\nconst map = new Map<string, number>();\nput(map);\nconsole.log(String(map.size));\n", "inherited library member set read as an own field"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), probe.stop) {
				t.Fatalf("want %q, got %v", probe.stop, err)
			}
		})
	}
	t.Run("length alone", func(t *testing.T) {
		t.Parallel()
		program := lowersAndAgreesWithNodeAs(t, pickedArrayD+"function size(xs: Pick<D[], 'length'>): number { return xs.length; }\nconst all: D[] = [];\nconsole.log(String(size(all)));\n", ".ts")
		if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "0\n" {
			t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, "0\n")
		}
	})
}
