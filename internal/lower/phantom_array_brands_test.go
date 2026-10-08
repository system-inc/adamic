package lower

import (
	"encoding/json"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestPhantomArrayRefusals(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"void", "undefined", "optional"} {
		for _, name := range []string{"length", "sort", "__proto__", "constructor", "0"} {
			t.Run(field+"/"+name, func(t *testing.T) {
				t.Parallel()
				optional, of := "", field
				if field == "optional" {
					optional, of = "?", "undefined"
				}
				_, err := lowerSource(t, "type Brand = number[] & { '"+name+"'"+optional+": "+of+" };")
				want := "main.a:1:14: Adamic 0.1 refuses an array brand member " + name + " that exists on the array; use a member name the array does not have on its own properties or prototype chain"
				if err == nil || !strings.HasSuffix(err.Error(), want) {
					t.Fatalf("got %v, want %s", err, want)
				}
			})
		}
	}
}

func TestPhantomArrayNames(t *testing.T) {
	t.Parallel()
	script := `let names = new Set(); for (let at = []; at !== null; at = Object.getPrototypeOf(at)) { for (const name of Object.getOwnPropertyNames(at)) names.add(name); } console.log(JSON.stringify([...names].sort()));`
	output, err := exec.Command("node", "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v: %s", err, output)
	}
	var names []string
	if err := json.Unmarshal(output, &names); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if !arrayBrandMember(name) {
			t.Errorf("Node array has %s, but rule accepts it", name)
		}
	}
	for _, name := range []string{"00", "-0", "charAt", " __sortedArrayBrand"} {
		if arrayBrandMember(name) {
			t.Errorf("absent array name %s refused", name)
		}
	}
}

func TestPhantomArrayCastsAreErased(t *testing.T) {
	t.Parallel()
	branded := `type Brand = number[] & { marker: void };
function into(values: number[]): Brand { return values as Brand; }
function out(values: Brand): number[] { return values as number[]; }
console.log(out(into([3, 1])).join(','));`
	plain := `type Brand = number[];
function into(values: number[]): Brand { return values; }
function out(values: Brand): number[] { return values; }
console.log(out(into([3, 1])).join(','));`
	got, err := lowerSource(t, branded)
	if err != nil {
		t.Fatal(err)
	}
	want, err := lowerSource(t, plain)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("array brands or casts changed the unbranded IR")
	}
}

func TestPhantomArrayProofs(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, reason string }{
		{"mutable widening", `type Brand = { kind: string }[] & { marker: void }; function widen(values: { kind: 'dog' }[]): Brand { return values as Brand; }`, "adamic/invariant-mutable"},
		{"interface mutable widening", `interface Brand extends Array<{ kind: string }> { marker: void; } function widen(values: { kind: 'dog' }[]): Brand { return values as Brand; }`, "adamic/invariant-mutable"},
		{"tuple representation", `type Brand = readonly number[] & { marker?: undefined }; function into(values: [number, number]): Brand { return values as Brand; }`, "not an array"},
		{"element narrowing", `type Brand = readonly number[] & { marker: void }; function narrow(values: readonly (number | string)[]): Brand { return values as Brand; }`, "source is not assignable"},
		{"readonly removal", `type Brand = number[] & { marker: void }; function widen(values: readonly number[]): Brand { return values as Brand; }`, "adamic/invariant-mutable"},
		{"cycle", `interface Links extends Array<Link> { marker: void; } interface Link { readonly links: Links; } const root: Link = { links: [] as Link[] as Links }; root.links.push(root);`, "adamic/cycle-capable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, test.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("got %v, want refusal containing %s", err, test.reason)
			}
		})
	}
}

func TestPhantomArrayNonVoidCastStaysUnsupported(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `type Brand = number[] & { marker: number }; function into(values: number[]): Brand { return values as Brand; }`)
	var notYet *NotYet
	var refused *Refused
	if !errors.As(err, &notYet) && !errors.As(err, &refused) {
		t.Fatalf("got %v, want unsupported data-bearing array view", err)
	}
}

func TestPhantomArrayWeakKeeping(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `import type { Weak } from 'adamic'; type Item = { name: string }; type Brand = readonly Weak<Item>[] & { marker: void }; function into(values: readonly Item[]): Brand { return values as Brand; }`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "weak") {
		t.Fatalf("got %v, want refused physical change from strong elements to weak handles", err)
	}
}

func TestPhantomArrayRequiredCastsAreErased(t *testing.T) {
	t.Parallel()
	branded := `interface Brand<T> extends ReadonlyArray<T> { marker: undefined; }
function into<T>(values: readonly T[]): Brand<T> { return values as Brand<T>; }
function out<T>(values: Brand<T>): readonly T[] { return values as readonly T[]; }
console.log(out(into([3, 1])).join(','));`
	plain := `type Brand<T> = readonly T[];
function into<T>(values: readonly T[]): Brand<T> { return values; }
function out<T>(values: Brand<T>): readonly T[] { return values; }
console.log(out(into([3, 1])).join(','));`
	got, err := lowerSource(t, branded)
	if err != nil {
		t.Fatal(err)
	}
	want, err := lowerSource(t, plain)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("required undefined array brands or casts changed the unbranded IR")
	}
}
