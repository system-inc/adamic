package lower

import (
	"errors"
	"strings"
	"testing"
)

// tsc's core.ts identity, the default of arrayToMap's mapper.
const overloadCertificateIdentity = `function identity<T>(x: T): T {
    return x;
}
`

// tsc's core.ts arrayToMap overloads 1 to 3, overload 4, and its implementation, which builds a fresh Map
// in an index loop and returns it, certificate B's shape (#5fws6q8).
const overloadCertificateArrayToMapOverloads = `function arrayToMap<K, V>(array: readonly V[], makeKey: (value: V) => K | undefined): Map<K, V>;
function arrayToMap<K, V1, V2>(array: readonly V1[], makeKey: (value: V1) => K | undefined, makeValue: (value: V1) => V2): Map<K, V2>;
function arrayToMap<T>(array: readonly T[], makeKey: (value: T) => string | undefined): Map<string, T>;
`

const overloadCertificateArrayToMapFourth = `function arrayToMap<T, U>(array: readonly T[], makeKey: (value: T) => string | undefined, makeValue: (value: T) => U): Map<string, U>;
`

const overloadCertificateArrayToMapImplementation = `function arrayToMap<K, V1, V2>(array: readonly V1[], makeKey: (value: V1) => K | undefined, makeValue: (value: V1) => V1 | V2 = identity): Map<K, V1 | V2> {
    const result = new Map<K, V1 | V2>();
    for (let i = 0; i < array.length; i++) {
        const value = array[i]!;
        const key = makeKey(value);
        if (key !== undefined) result.set(key, makeValue(value));
    }
    return result;
}
`

// overloadCertificateArrayToMap is tsc's arrayToMap whole, its four overloads over identity.
const overloadCertificateArrayToMap = overloadCertificateIdentity + overloadCertificateArrayToMapOverloads + overloadCertificateArrayToMapFourth + overloadCertificateArrayToMapImplementation

// overloadCertificateMapShow prints a Map<string, string> in insertion order, as both sides iterate it.
const overloadCertificateMapShow = `const items: string[] = ["a", "bb", "ccc"];
function show(entries: Map<string, string>): void {
    for (const [key, value] of entries) {
        console.log(` + "`${key}=${value}`" + `);
    }
}
`

// tsc's arrayToMap whole, each of its four overloads called directly, lowers and agrees with Node on the
// JavaScript backend and natively. The census accepts overloads 2 and 4, Map<K, V2> and Map<string, U>
// over Map<K, V1 | V2>, by certificate B where it refused them before (census_small.go's result refusal,
// the scout's first arrayToMap failure, and overload 4's parameter makeValue). Overloads 1 and 3 omit the
// mapper; the census accepts them by relation, the implementation's V2 witnessed as their element
// (inferTypes' one missing union binder), before any certificate is asked. Overloads 3 and 4 are reached
// by their type arguments, since overloads 1 and 2 take every call they would. The mapper maps a string
// to a string, so the implementation's instance holds Map<string, string> as the overload promises; rows
// where it doesn't are TestOverloadCertificateMapInstanceStorage's.
func TestOverloadCertificateArrayToMapLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, call, want string }{
		{"overload 1", `show(arrayToMap(items, s => s.length > 2 ? undefined : s.toUpperCase()));`, "A=a\nBB=bb\n"},
		{"overload 2", `show(arrayToMap(items, s => s.length > 2 ? undefined : s.toUpperCase(), s => s + s));`, "A=aa\nBB=bbbb\n"},
		{"overload 3", `show(arrayToMap<string>(items, s => s.length > 2 ? undefined : s.toUpperCase()));`, "A=a\nBB=bb\n"},
		{"overload 4", `show(arrayToMap<string, string>(items, s => s.length > 2 ? undefined : s.toUpperCase(), s => s + s));`, "A=aa\nBB=bbbb\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, overloadCertificateArrayToMap+overloadCertificateMapShow+probe.call+"\n", ".ts")
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.want {
				t.Fatalf("native: exit %d, stdout %q, stderr %q; want %q", code, stdout, stderr, probe.want)
			}
		})
	}
}

// tsc's arrayToMap overload 4 has two binders to the implementation's three. inferTypes witnesses V1 as T
// and K as string but none for V2 (makeValue's V1 | V2 against U, where T doesn't fit U), and bottom made
// the implementation's makeValue (value: T) => T, which overload 4's (value: T) => U doesn't serve, so
// the census refused the set at overload 4's parameter. censusInferRemainder reads V2 as U, makeValue is
// (value: T) => T | U, which (value: T) => U serves, and the result goes to certificate B as overload 2's
// does, so the set lowers whole. Mutant: the remainder read dropped, and overload 4 refuses again.
func TestOverloadCertificateArrayToMapFourthOverloadServed(t *testing.T) {
	t.Parallel()
	source := overloadCertificateArrayToMap + overloadCertificateMapShow + "show(arrayToMap<string, string>(items, s => s.toUpperCase(), s => s + s));\n"
	if _, err := lowerSourceAs(t, source, ".ts"); err != nil {
		t.Fatalf("want tsc's arrayToMap with overload 4 called to lower, got %v", err)
	}
}

// The witness only lets overload 4 reach the result's judgment; the certificate still judges it. An
// implementation that sets the element where tsc's sets makeValue's answer passes the parameter census
// under V2 = U as tsc's does, and is refused at overload 4's declaration for its result, since its body
// isn't certificate B's shape and a T is no U.
func TestOverloadCertificateArrayToMapFourthOverloadElementSetStaysRefused(t *testing.T) {
	t.Parallel()
	source := overloadCertificateIdentity + overloadCertificateArrayToMapFourth + strings.Replace(overloadCertificateArrayToMapImplementation, "result.set(key, makeValue(value))", "result.set(key, value)", 1) + `const made = arrayToMap<number, string>([1], v => "k", v => "s");
const actual = made.get("k");
console.log(` + "`expected string, got ${typeof actual}`" + `);
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of arrayToMap result Map<string, U> ") || !strings.Contains(refused.What, "cannot be served by implementation result") {
		t.Fatalf("want overload 4 of tsc's set, alone here, refused at its declaration for its result, got %v", err)
	}
}

// overloadCertificateMapBad is a set whose one overload omits the mapper and promises Map<string, string>
// over an implementation of certificate B's shape returning Map<string, string | number>, its mapper
// defaulting to defaulted. before goes ahead of the set, body between the Map and the loop, set is the
// loop's guarded statement, and after between the loop and the return.
func overloadCertificateMapBad(before, defaulted, body, set, after string) string {
	return before + `function bad(values: readonly string[], makeKey: (value: string) => string | undefined): Map<string, string>;
function bad(values: readonly string[], makeKey: (value: string) => string | undefined, makeValue: (value: string) => string | number = ` + defaulted + `): Map<string, string | number> {
    const result = new Map<string, string | number>();
` + body + `    for (let i = 0; i < values.length; i++) {
        const value = values[i]!;
        const key = makeKey(value);
        ` + set + `
    }
` + after + `    return result;
}
`
}

// Correctness's Map negatives (stage3/census/overload-body-proof-scout/counterexamples on
// typescript/overload-core-evidence), each still refused at its declaration for its result. Her programs
// loop by for-of over a literal, which the certificate refuses before it reaches the hazard each one
// shows, so each is here as she wrote it, her observation line made one string, and again in the
// canonical index loop over a parameter, so the row stands on its own hazard: the alias, the capture, a
// second set no proof judged (map-wrong-set), a getter read in the set's argument, a mapper the overload
// lets the caller pass as undefined over an identity default, whose element isn't the promised value
// (map-default-undefined), a default other than the identity, an identity that's a rebindable variable,
// and the mapper parameter written.
func TestOverloadCertificateMapNegativesStayRefused(t *testing.T) {
	t.Parallel()
	const observe = "console.log(`expected string, got ${typeof actual}`);\n"
	const guarded = "if (key !== undefined) result.set(key, makeValue(value));"
	const call = "const actual = bad([\"a\"], v => v).get(\"a\");\n" + observe
	for _, probe := range []struct{ name, source string }{
		{"map-alias", `function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    const alias = out;
    for (const value of ["ok"]) out.set("key", value);
    alias.set("key", 0);
    return out;
}
const actual = bad().get("key");
` + observe},
		{"map-capture", `function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    const later = () => out.set("key", 0);
    for (const value of ["ok"]) out.set("key", value);
    later();
    return out;
}
const actual = bad().get("key");
` + observe},
		{"map-default-undefined", `function bad(values: number[], makeValue: ((v: number) => string) | undefined): Map<string, string>;
function bad(values: number[], makeValue: (v: number) => string | number = v => v): Map<string, string | number> {
    const out = new Map<string, string | number>();
    for (const value of values) out.set("key", makeValue(value));
    return out;
}
const actual = bad([0], undefined).get("key");
` + observe},
		{"map-getter", `function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    let reads = 0;
    const source = { get value(): string | number { return ++reads === 1 ? "ok" : 0; } };
    for (const unused of [0]) {
        if (typeof source.value === "string") out.set("key", source.value);
    }
    return out;
}
const actual = bad().get("key");
` + observe},
		{"map-wrong-set", `function bad(): Map<string, string>;
function bad(): Map<string, string | number> {
    const out = new Map<string, string | number>();
    for (const value of ["ok"]) out.set("key", value);
    out.set("key", 0);
    return out;
}
const actual = bad().get("key");
` + observe},
		{"alias over a parameter", overloadCertificateMapBad(overloadCertificateIdentity, "identity", "    const alias = result;\n", guarded, "    alias.set(\"a\", 0);\n") + call},
		{"capture over a parameter", overloadCertificateMapBad(overloadCertificateIdentity, "identity", "    const later = () => result.set(\"a\", 0);\n", guarded, "    later();\n") + call},
		{"wrong set over a parameter", overloadCertificateMapBad(overloadCertificateIdentity, "identity", "", guarded, "    result.set(\"a\", 0);\n") + call},
		{"getter in the set argument", overloadCertificateMapBad(overloadCertificateIdentity, "identity", "    let reads = 0;\n    const source = { get value(): string | number { return ++reads === 1 ? \"ok\" : 0; } };\n", "if (key !== undefined && typeof source.value === \"string\") result.set(key, source.value);", "") + call},
		{"supplied undefined mapper over identity", overloadCertificateIdentity + `function bad(values: readonly number[], makeKey: (value: number) => string | undefined, makeValue: ((value: number) => string) | undefined): Map<string, string>;
function bad(values: readonly number[], makeKey: (value: number) => string | undefined, makeValue: (value: number) => string | number = identity): Map<string, string | number> {
    const result = new Map<string, string | number>();
    for (let i = 0; i < values.length; i++) {
        const value = values[i]!;
        const key = makeKey(value);
        if (key !== undefined) result.set(key, makeValue(value));
    }
    return result;
}
const actual = bad([0], v => "a", undefined).get("a");
` + observe},
		{"default other than identity", overloadCertificateMapBad("", "() => 0", "", guarded, "") + call},
		{"rebound identity", overloadCertificateMapBad("let identity = (x: string): string | number => x;\n", "identity", "", guarded, "") + "identity = () => 0;\n" + call},
		{"mapper written", overloadCertificateMapBad(overloadCertificateIdentity, "identity", "    makeValue = () => 0;\n", guarded, "") + call},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "overload 1 of bad result Map<string, string> ") || !strings.Contains(refused.What, "cannot be served by implementation result Map<string, string | number>") {
				t.Fatalf("want overload 1 of bad refused at its declaration for its result, got %v", err)
			}
		})
	}
}

// Lower's probe (J-h37mw8d on #5fws6q8) for certificate B: overloads 2 and 4 called where the mapper
// changes the element's type, so the instance stores Map<K, V1 | V2> where the call reads Map<K, V2>.
// Each row stops at the call in the storage condition's words or agrees with Node natively. Overload 4
// from numbers to strings must stop: the call's own reading of the instance (instanceTypes) witnesses no V2
// there, inferTypes' union arm asking a number to fit a string, so the instance's result isn't read at all.
// Mutant: the storage condition dropped, and every row stops elsewhere, inside the instance, where no Map
// of a union lowers.
func TestOverloadCertificateMapInstanceStorage(t *testing.T) {
	t.Parallel()
	show := "function show(entries: Map<string, string>): void {\n    for (const [key, value] of entries) {\n        console.log(`${key}=${value.toUpperCase()}`);\n    }\n}\nconst numbers: number[] = [1, 2, 3];\n"
	for _, probe := range []struct {
		name, source, want string
		stops              bool
	}{
		{"overload 2 number to string", overloadCertificateArrayToMap + show + "show(arrayToMap(numbers, n => `k${n}`, n => `v${n * 2}`));\n", "k1=V2\nk2=V4\nk3=V6\n", false},
		{"overload 2 string to number", overloadCertificateArrayToMap + "const items: string[] = [\"a\", \"bb\"];\nconst lengths = arrayToMap(items, s => s, s => s.length);\nfor (const [key, value] of lengths) console.log(`${key}=${(value + 1).toFixed(1)}`);\n", "a=2.0\nbb=3.0\n", false},
		{"overload 4 number to string", overloadCertificateArrayToMap + show + "show(arrayToMap<number, string>(numbers, n => `k${n}`, n => `v${n * 2}`));\n", "k1=V2\nk2=V4\nk3=V6\n", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			overloadCertificateInstanceRow(t, probe.source, probe.want, probe.stops)
		})
	}
}

// The identity branch accepts what it should: a set whose implementation doesn't relate to any overload's
// result, so every overload is asked of certificate B. Overload 1 omits the mapper, so the default runs and
// identity hands back the element; overload 2 requires it, so the default can't run; overload 3 makes it
// optional, so both run and both are judged. None is refused at its declaration. Whether the calls lower
// past it is the call's own question (the implementation's Map<string, string | number> in another
// representation than an overload's), logged.
func TestOverloadCertificateMapIdentityDefaultAccepted(t *testing.T) {
	t.Parallel()
	source := overloadCertificateIdentity + `function toMap(values: readonly string[], makeKey: (value: string) => string | undefined): Map<string, string>;
function toMap(values: readonly string[], makeKey: (value: string) => string | undefined, makeValue: (value: string) => number): Map<string, number>;
function toMap(values: readonly string[], makeKey: (value: string) => string | undefined, makeValue?: (value: string) => string): Map<string, string>;
function toMap(values: readonly string[], makeKey: (value: string) => string | undefined, makeValue: (value: string) => string | number = identity): Map<string, string | number> {
    const result = new Map<string, string | number>();
    for (let i = 0; i < values.length; i++) {
        const value = values[i]!;
        const key = makeKey(value);
        if (key !== undefined) result.set(key, makeValue(value));
    }
    return result;
}
const plain = toMap(["a", "bb"], s => s.toUpperCase());
console.log(plain.get("A") ?? "missing");
`
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if errors.As(err, &refused) && strings.Contains(refused.What, " of toMap ") {
		t.Fatalf("an overload of toMap refused at its declaration: %v", err)
	}
	t.Logf("lowered whole: %v", err)
}

// Her prototype negative's road for a Map: with the set the library's own by every declaration the
// certificate passes, and the rewrite of Map.prototype's set that would break it stops where it's
// written (library_object.go), as under certificate A.
func TestOverloadCertificateMapPrototypeRewriteStaysRefused(t *testing.T) {
	t.Parallel()
	source := `Object.defineProperty(Map.prototype, "set", {
    value: function(this: Map<string, string | number>) { return this; },
});
` + overloadCertificateMapBad(overloadCertificateIdentity, "identity", "", "if (key !== undefined) result.set(key, makeValue(value));", "") + "const actual = bad([\"a\"], v => v).get(\"a\");\nconsole.log(`expected string, got ${typeof actual}`);\n"
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "Object.defineProperty" {
		t.Fatalf("want the prototype rewrite refused where it's written, got %v", err)
	}
}

// Certificate B rides certificate A's gate: each row enters arrayToMap's overload 2, the one B accepts,
// and its entry is answered by overloadCertificateEntry on its own node. A direct entry never says the
// gate's words, and whether it lowers past the gate is that road's own question, logged; a comma callee
// or a spread argument is NotYet at the call in the gate's words, naming its form. Mutant: the gate
// answering true for every call, and the comma and spread rows go red.
func TestOverloadCertificateMapGate(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, call string
		direct     bool
		form       string
	}{
		{"direct call", "const got = arrayToMap(items, s => s.toUpperCase(), s => s + s);\n", true, ""},
		{"comma callee", "let touched = 0;\nconst got = (touched++, arrayToMap)(items, s => s.toUpperCase(), s => s + s);\n", false, "a comma expression callee"},
		{"spread argument", "const args: [readonly string[], (value: string) => string | undefined, (value: string) => string] = [items, s => s.toUpperCase(), s => s + s];\nconst got = arrayToMap(...args);\n", false, "a spread argument"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			modules := map[string]string{"main.ts": overloadCertificateArrayToMap + "const items: string[] = [\"a\", \"bb\"];\n" + probe.call + "console.log(String(got.size));\n"}
			l, node, overload := overloadCertificateProbe(t, "arrayToMap", modules)
			if erase := l.overloadCertificateEntry(node, overload); erase != probe.direct {
				t.Fatalf("overloadCertificateEntry: got %v, want %v", erase, probe.direct)
			}
			_, err := lowerModules(t, modules)
			if probe.direct {
				if err != nil && strings.Contains(err.Error(), overloadCertificateGateWords) {
					t.Fatalf("a direct entry stopped at the gate: %v", err)
				}
				t.Logf("lowered whole: %v", err)
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != overloadCertificateGateWords+probe.form {
				t.Fatalf("want NotYet %q at the call, got %v", overloadCertificateGateWords+probe.form, err)
			}
		})
	}
}
