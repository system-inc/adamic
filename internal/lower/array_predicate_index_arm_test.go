package lower

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// jsonRecovery is tsc's JSON value and record, as stage3's adaptation 52 publishes them, the arm rulings.md:260's
// amendment (Nov 26, @system_adamic on #gcec0bx) admits by its index.
const jsonRecovery = "type AdamicJsonRecoveryValue = string | number | boolean | null | AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject;\ninterface AdamicJsonRecoveryObject { [key: string]: AdamicJsonRecoveryValue }\n"

// isArrayArgument is where a source's last Array.isArray call takes its argument, by line and column, the place
// arrayIsArray's stop names.
func isArrayArgument(source string) string {
	where := ""
	for index, line := range strings.Split(source, "\n") {
		if column := strings.Index(line, "Array.isArray("); column >= 0 {
			where = fmt.Sprintf(":%d:%d", index+1, column+len("Array.isArray(")+1)
		}
	}
	return where
}

// An index-only arm failing the amendment's conditions keeps today's stop and words at isArray's argument, each handed
// a value Node answers for (Diagnostics' pins on #gcec0bx):
//   - unknown-value: the index holds unknown, which takes a function (condition 2);
//   - callable-value: the index's union holds a function type (condition 2);
//   - array-stored-string-index: the JSON record beside an interface stored as an array that declares a string index
//     of its own, which fails every index-only arm closed (condition 3). Its index is unknown here, which .a can hold;
//     one of any would take the array's methods, and the array would be assignable to the JSON record with no hatch.
//
// Mutant (array_predicate_index_arm.mutants.md, M1): condition 2 dropped, every index value admitted. unknown-value and
// callable-value go red, each lowering past isArray; array-stored-string-index stays green, condition 3 still stopping it.
func TestArrayPredicateIndexArmRefusals(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, node string }{
		{"unknown-value", "interface Bag { [key: string]: unknown }\nfunction isList(value: Bag | string[]): boolean { return Array.isArray(value); }\nconsole.log(String(isList(['a'])));\n", "true\n"},
		{"callable-value", "interface Bag { [key: string]: string | (() => void) }\nfunction isList(value: Bag | string[]): boolean { return Array.isArray(value); }\nconsole.log(String(isList(['a'])));\n", "true\n"},
		{"array-stored-string-index", jsonRecovery + "interface Loose extends ReadonlyArray<string> { [key: string]: unknown }\nfunction isList(value: AdamicJsonRecoveryObject | string[]): boolean { return Array.isArray(value); }\nconsole.log(String(isList({ a: 1 })));\n", "false\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if node := runAgreementNode(t, writeAgreementSource(t, "main.a", row.source)); string(node.stdout) != row.node || node.code != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want %q", node.code, node.stdout, node.stderr, row.node)
			}
			_, err := lowerSource(t, row.source)
			var notYet *NotYet
			const what = "Array.isArray on a tuple or an erased object/any/unknown view"
			if where := isArrayArgument(row.source); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, where) || notYet.What != what {
				t.Fatalf("want NotYet at %s, %q, got %v", where, what, err)
			}
		})
	}
}

// arrayExcludedIndexArm on declared types, where a type parameter is still one: lowering reads an instantiation's
// types, so a generic arm a program lowers is read at what its instantiation made its type parameter, and only here,
// on the declaration, is the parameter met as a parameter. Condition 2 reads it at its constraint: Bag's V has none
// and fails, Bounded's holds no callable and passes, Loosely's holds one and fails. Record<string, V> resolves to the
// same index (Counts passes, Calls fails), a property beside the index fails condition 1, and so does an index that
// isn't a string's: Numbered's and Language's NumberOnly (her i01, an arm a string[] is assignable to tsc-clean),
// Pattern's template pattern (her i02) and Sym's symbol (her i05), each of whose value types condition 2 would pass. A closed
// world holding an interface stored as an array with a string index of its own (Loose) fails every arm, condition 3.
//
// Mutants: M1, condition 2 dropped, turns open red on Bag, Loosely, Calls and Unknowns; M4, condition 1 relaxed to any
// one index signature, turns open red on Numbered, NumberOnly, Pattern and Sym.
func TestArrayPredicateIndexArmValueType(t *testing.T) {
	t.Parallel()
	const declarations = jsonRecovery + `interface Bag<V> { [key: string]: V }
interface Bounded<V extends string | number> { [key: string]: V }
interface Loosely<V extends string | (() => void)> { [key: string]: V }
type Counts = Record<string, number>;
type Calls = Record<string, () => void>;
interface Unknowns { [key: string]: unknown }
interface Named { name: string; [key: string]: string }
interface Numbered { [index: number]: string }
type NumberOnly = { [n: number]: string };
` + "type Pattern = { [k: `x${string}`]: number };\ntype Sym = { [s: symbol]: number };\n"
	for _, probe := range []struct {
		name, source string
		admitted     map[string]bool
	}{
		{"open", declarations, map[string]bool{"AdamicJsonRecoveryObject": true, "Bag": false, "Bounded": true, "Loosely": false, "Counts": true, "Calls": false, "Unknowns": false, "Named": false, "Numbered": false, "NumberOnly": false, "Pattern": false, "Sym": false}},
		{"closed-world-string-index", declarations + "interface Loose extends ReadonlyArray<string> { [key: string]: unknown }\n", map[string]bool{"AdamicJsonRecoveryObject": false, "Bounded": false, "Counts": false}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", probe.source)})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked}
			seen := map[string]bool{}
			for _, statement := range file.Statements.Nodes {
				if statement.Kind != ast.KindInterfaceDeclaration && statement.Kind != ast.KindTypeAliasDeclaration {
					continue
				}
				name := statement.Name().Text()
				want, asked := probe.admitted[name]
				if !asked {
					continue
				}
				seen[name] = true
				declared := checked.GetDeclaredTypeOfSymbol(checked.GetSymbolOfDeclaration(statement))
				if got := l.arrayExcludedIndexArm(declared); got != want {
					t.Errorf("%s: arrayExcludedIndexArm %v, want %v", name, got, want)
				}
			}
			if len(seen) != len(probe.admitted) {
				t.Fatalf("asked %d declarations, found %d", len(probe.admitted), len(seen))
			}
		})
	}
}

// Language's i01 on #gcec0bx: a number-only index arm takes a string[] tsc-clean, since a number index constrains only
// numeric keys and the array's methods never meet it, and Node calls the value an array, so the arm must stay out of
// the domain. (Her i02 and i05, a template-pattern and a symbol index alone, tsc refuses as written; they're held at
// condition 1 on their declarations in TestArrayPredicateIndexArmValueType.) The program must stop before it lowers
// whole, by its own first stop: a type indexed by a number alone has no representation natively (records.go's
// recordElement takes a string index only), so it may stop at the arm's first value before isArray's argument, and the
// row asks only that it stops, logging where. Condition 1's own test of the key is TestArrayPredicateIndexArmValueType's,
// which M4 turns red.
func TestArrayPredicateIndexArmKeyRefusals(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, node string }{
		{"i01-number-only", "type NumberOnly = { [n: number]: string };\ntype Value = string | NumberOnly;\nfunction show(v: Value): string {\n\tif (typeof v === \"object\" && !Array.isArray(v)) return \"record \" + String(v[0]);\n\treturn \"other\";\n}\nconst arr: string[] = [\"a\"];\nconst asArm: NumberOnly = arr;\nconsole.log(show(asArm));\n", "other\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if node := runAgreementNode(t, writeAgreementSource(t, "main.ts", row.source)); string(node.stdout) != row.node || node.code != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want %q", node.code, node.stdout, node.stderr, row.node)
			}
			_, err := lowerSourceAs(t, row.source, ".ts")
			var notYet *NotYet
			var refused *Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("want a stop, got %v", err)
			}
			t.Logf("stops: %v", err)
		})
	}
}

// Condition 4 (Lower's probe on #9z2ksv7, read for the index arm): a checked view checks at a read, never at its cast,
// so a view whose target the JSON record takes can put a tuple held as an object behind the arm with nothing read.
// Named's object literal type takes an implicit index, so it's assignable to AdamicJsonRecoveryObject, and the overlap
// view of Named over `[number, number] | Other` reaches the arm; checkArrayArmViews stops the isArray over it, by its
// own words, at its argument. Node prints false for the record it's handed.
//
// Mutant (M3): the index arm not recorded in arrayArmSites. The program lowers past isArray and this goes red.
func TestArrayPredicateIndexArmViewStops(t *testing.T) {
	t.Parallel()
	source := jsonRecovery + `type Named = { readonly name: string };
interface Other { readonly name: string; readonly kind: 'o' }
const pair: [number, number] = [1, 2];
function hold(): [number, number] | Other { return pair; }
const u = hold();
const named = u as Named;
function isList(value: AdamicJsonRecoveryObject | string[]): boolean { return Array.isArray(value); }
console.log(String(isList({ a: 1 })));
`
	path := writeAgreementSource(t, "main.ts", source)
	if node := runAgreementNode(t, path); string(node.stdout) != "false\n" || node.code != 0 {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want false", node.code, node.stdout, node.stderr)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_, err = Lower(context.Background(), program)
	var notYet *NotYet
	const opening = "Array.isArray over AdamicJsonRecoveryObject, which the checked view of Named over "
	if !errors.As(err, &notYet) || notYet.Where != path+isArrayArgument(source) || !strings.HasPrefix(notYet.What, opening) || !strings.HasSuffix(notYet.What, " can hold an array behind, unread") {
		t.Fatalf("want checkArrayArmViews' NotYet at %s, got %v", isArrayArgument(source), err)
	}
}
