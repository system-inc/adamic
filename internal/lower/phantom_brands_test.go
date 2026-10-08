package lower

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func TestPhantomRefusals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, what, fix string }{
		{"prototype", "type Brand = string & { __proto__: void };", "a primitive brand member __proto__ that exists on the primitive", "use a member name the primitive does not have on its own properties or prototype chain"},
		{"length", "type Brand = string & { length: void };", "a primitive brand member length that exists on the primitive", "use a member name the primitive does not have on its own properties or prototype chain"},
		{"nonvoid", "type Brand = string & { __escapedIdentifier: number };", "a primitive brand member __escapedIdentifier whose type is not void", "make __escapedIdentifier void (or optional and typed undefined) so the brand is phantom"},
		{"required undefined", "type Brand = number & { brand: undefined };", "a primitive brand member brand whose type is not void", "make brand void (or optional and typed undefined) so the brand is phantom"},
		{"any Path", "type Path = string & { __pathBrand: any };", "a primitive brand member __pathBrand whose type is not void", "make __pathBrand void (or optional and typed undefined) so the brand is phantom"},
		{"string index", "type Brand = string & { '0': void };", "a primitive brand member 0 that exists on the primitive", "use a member name the primitive does not have on its own properties or prototype chain"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, test.source)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("got %v, want Refused", err)
			}
			if refused.What != test.what || refused.Fix != test.fix {
				t.Fatalf("got %v; want %s; %s", err, test.what, test.fix)
			}
		})
	}
	t.Run("review exact message", func(t *testing.T) {
		source, err := os.ReadFile(filepath.Join("../../review/phantom-brands/prototype-read.a"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = lowerSource(t, string(source))
		want := "main.a:1:14: Adamic 0.1 refuses a primitive brand member __proto__ that exists on the primitive; use a member name the primitive does not have on its own properties or prototype chain"
		if err == nil || !strings.HasSuffix(err.Error(), want) {
			t.Fatalf("got %v, want %s", err, want)
		}
	})
}

func TestPhantomCastsAreErased(t *testing.T) {
	t.Parallel()
	branded := `type Escaped = string & { __escapedIdentifier: void };
function into(value: string): Escaped { return value as Escaped; }
function out(value: Escaped): string { return value as string; }
console.log(out(into('made' + 'now')));`
	plain := `type Escaped = string;
function into(value: string): Escaped { return value; }
function out(value: Escaped): string { return value; }
console.log(out(into('made' + 'now')));`
	got, err := lowerSource(t, branded)
	if err != nil {
		t.Fatal(err)
	}
	want, err := lowerSource(t, plain)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("casts must leave identical IR:\ngot %#v\nwant %#v", got.Functions, want.Functions)
	}
	for _, function := range got.Functions {
		returned, ok := function.Body[0].(ir.Return)
		if !ok {
			t.Fatalf("unexpected body %v", function.Body)
		}
		if _, ok := returned.Value.(ir.Read); !ok {
			t.Fatalf("cast kept runtime operation %T", returned.Value)
		}
	}
}

func TestPhantomPrimitiveNames(t *testing.T) {
	t.Parallel()
	// Node owns the inventory. Test every observed name, including ones absent from Adamic's lib.
	script := `const results = []; for (const value of ['', 0, false]) { let at = Object(value), names = new Set(); while (at !== null) { for (const name of Object.getOwnPropertyNames(at)) names.add(name); at = Object.getPrototypeOf(at); } results.push([...names].sort()); } console.log(JSON.stringify(results));`
	output, err := exec.Command("node", "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v: %s", err, output)
	}
	var inventory [][]string
	if err := json.Unmarshal(output, &inventory); err != nil {
		t.Fatal(err)
	}
	flags := []checker.TypeFlags{checker.TypeFlagsString, checker.TypeFlagsNumber, checker.TypeFlagsBoolean}
	for index, names := range inventory {
		for _, name := range names {
			if !primitiveMember(flags[index], name) {
				t.Errorf("Node's primitive %d has %s, but brand rule accepts it", index, name)
			}
		}
	}
	for _, name := range []string{"0", "1", "9007199254740990"} {
		if !primitiveMember(checker.TypeFlagsString, name) {
			t.Errorf("string own index %s allowed", name)
		}
	}
	for _, name := range []string{"00", "-0", "1.0", "__escapedIdentifier"} {
		if primitiveMember(checker.TypeFlagsString, name) {
			t.Errorf("absent name %s refused", name)
		}
	}
	if primitiveMember(checker.TypeFlagsBoolean, "length") || primitiveMember(checker.TypeFlagsNumber, "slice") {
		t.Error("used another primitive's chain")
	}
}

func TestPhantomLiteralCastsStayRefused(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`type Only = 'yes' & { brand: void }; const value = 'no' as Only;`,
		`type Numeric = number & { brand: void }; function into(value: string | number): Numeric { return value as Numeric; }`,
	} {
		// Stock tsc permits the first assertion but the phantom rule must retain the literal proof.
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Fatalf("got %v, want a refused cast", err)
		}
	}
}

func TestPhantomRealPathRefusal(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../review/phantom-brands/nonvoid-path.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	want := "main.a:2:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or optional and typed undefined) so the brand is phantom"
	if err == nil || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("got %v, want %s", err, want)
	}
}
