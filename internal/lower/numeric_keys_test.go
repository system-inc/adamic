package lower

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// numericKeyProgram is the oracle's element_access_numeric_keys.ts written on fewer lines: a table keyed by two
// enum members and a literal, read only through its keys.
const numericKeyProgram = `enum Kind { A, B, C }
const names = { [Kind.A]: "a", [Kind.B]: "b", [7]: "seven" };
function nameOf(kind: Kind.A | Kind.B): string { return names[kind]; }
console.log(nameOf(Kind.A) + " " + nameOf(Kind.B) + " " + names[7]);
`

// A const table whose numeric computed keys nothing can enumerate lowers as an object whose fields are named by
// the keys' canonical number strings, Adamic's ruling (A) on #d1chbm7 (numeric_keys.go), and its reads name the
// same fields (finiteElementRead). A name in a type, keyof typeof table, is no reference. Mutants, red: numericFieldName
// returning false first, so every row stops at its first key; inTypeQuery dropped from numericKeyReferencesKeyed's
// walk, so the type query row does.
func TestNumericKeyTablesLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source, fields, read string
	}{
		{"keyed by enum members and a literal", numericKeyProgram, "0,1,7", "7"},
		{"read through a type query", "enum Kind { A, B }\nconst table = { [Kind.A]: \"a\", [(Kind.B)]: \"b\" };\nfunction nameOf(key: keyof typeof table): string { return table[key]; }\nconsole.log(nameOf(0) + nameOf(1));\n", "0,1", "1"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowered, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			literals, reads := []string{}, map[string]bool{}
			visit := func(node any) bool {
				switch node := node.(type) {
				case ir.ObjectLiteral:
					if !node.Tuple {
						names := []string{}
						for _, field := range node.Fields {
							names = append(names, field.Name)
						}
						sort.Strings(names)
						literals = append(literals, strings.Join(names, ","))
					}
				case ir.Property:
					reads[node.Name] = true
				}
				return true
			}
			walk(lowered.Main, visit)
			for _, function := range lowered.Functions {
				walk(function.Body, visit)
			}
			found := false
			for _, fields := range literals {
				found = found || fields == probe.fields
			}
			if !found {
				t.Fatalf("no object literal with fields %s, only %q", probe.fields, literals)
			}
			if !reads[probe.read] {
				t.Fatalf("no read of field %q", probe.read)
			}
		})
	}
}

// Every table the ruling leaves out keeps its stop, "a computed field name", at its first numeric key, each for
// the condition its row names (numeric_keys.go). A method keyed by a number never comes to the field arm: it keeps
// the method arm's own stop. A write through delete isn't a row, since tsc refuses deleting a required field
// (TS2790), and duplicate keys aren't either: TestNumericKeyDuplicatesStopAtTheChecker holds them. Mutants:
// numeric_keys.mutants.md.
func TestNumericKeysKeepTheirStop(t *testing.T) {
	t.Parallel()
	const kinds = "enum Kind { A, B, C }\n"
	for _, probe := range []struct {
		name, source, at, want string
	}{
		{"a fractional key", "const table = { [1.5]: \"x\" };\nconsole.log(table[1.5]);\n", "[1.5]", ""},
		// The checker gives -1 the number literal type -1, a minus applied to the literal 1.
		{"a negative key", "const table = { [-1]: \"x\" };\nconsole.log(table[-1]);\n", "[-1]", ""},
		{"minus zero", "const table = { [-0]: \"x\" };\nconsole.log(table[0]);\n", "[-0]", ""},
		{"a constant that isn't an enum member", "const zero = 0;\nconst table = { [zero]: \"x\" };\nconsole.log(table[zero]);\n", "[zero]", ""},
		{"an exported table", kinds + "export const table = { [Kind.A]: \"a\" };\nconsole.log(table[Kind.A]);\n", "[Kind.A]", ""},
		{"exported by a declaration", kinds + "const table = { [Kind.A]: \"a\" };\nexport { table };\nconsole.log(table[Kind.A]);\n", "[Kind.A]", ""},
		{"a let table", kinds + "let table = { [Kind.A]: \"a\" };\nconsole.log(table[Kind.A]);\n", "[Kind.A]", ""},
		{"passed to Object.keys", kinds + "const table = { [Kind.A]: \"a\" };\nconsole.log(table[Kind.A] + String(Object.keys(table).length));\n", "[Kind.A]", ""},
		{"spread", kinds + "const table = { [Kind.A]: \"a\" };\nconst copy = { ...table };\nconsole.log(copy[Kind.A]);\n", "[Kind.A]", ""},
		{"written", kinds + "const table = { [Kind.A]: \"a\", [Kind.B]: \"b\" };\ntable[Kind.A] = \"z\";\nconsole.log(table[Kind.B]);\n", "[Kind.A]", ""},
		{"printed", kinds + "const table = { [Kind.A]: \"a\" };\nconsole.log(String(table));\n", "[Kind.A]", ""},
		{"aliased", kinds + "const table = { [Kind.A]: \"a\" };\nconst alias = table;\nconsole.log(alias[Kind.A]);\n", "[Kind.A]", ""},
		{"read by a string key", kinds + "const table = { [Kind.A]: \"a\" };\nconsole.log(table[\"0\"]);\n", "[Kind.A]", ""},
		// A call through table[k] hands the function the table as its this, which could enumerate it.
		{"called through", kinds + "const table = { [Kind.A]: () => \"a\" };\nconsole.log(table[Kind.A]());\n", "[Kind.A]", ""},
		{"past the last array index", "const table = { [4294967295]: \"x\" };\nconsole.log(table[4294967295]);\n", "[4294967295]", ""},
		// Language's n01 and n02 (#d1chbm7): an entry handed the table as this by a call through the lookup, and a getter
		// whose body names the table, each of which could enumerate it. n02's getter joins the keys, as n01's entry does,
		// where hers returned JSON.stringify's string | undefined, which this harness's types refuse.
		{"Language's n01, an entry called through the lookup", "enum Kind { A = 1, B = 2, C = 10 }\nconst table = {\n\t[Kind.C]: function (this: object): string {\n\t\treturn Object.keys(this).join(\",\");\n\t},\n\t[Kind.B]: (): string => \"b\",\n\t[Kind.A]: (): string => \"a\",\n};\nfunction visit(kind: Kind.C): string {\n\treturn table[kind]();\n}\nconsole.log(visit(Kind.C));\n", "[Kind.C]", ""},
		{"Language's n02, a getter naming the table, stopped as an accessor first", "enum Kind { A = 1, B = 2 }\nconst table = {\n\t[Kind.B]: (): string => \"b\",\n\t[Kind.A]: (): string => \"a\",\n\tget [3](): string {\n\t\treturn Object.keys(table).join(\",\");\n\t},\n};\nconsole.log(table[3]);\n", "[Kind.B]", "a computed accessor literal name"},
		// n02's point reached by the proof itself: an arrow entry whose body names the table, so the table can reach
		// Object.keys through it, and the walk finds a reference that is no keyed lookup.
		{"an entry naming the table", "enum Kind { A = 1, B = 2 }\nconst table = {\n\t[Kind.B]: (): string => \"b\",\n\t[Kind.A]: (): string => Object.keys(table).join(\",\"),\n};\nconst pick = table[Kind.A];\nconsole.log(pick());\n", "[Kind.B]", ""},
		{"a method entry", kinds + "const table = { [Kind.A]() { return \"a\"; } };\nconsole.log(table[Kind.A]());\n", "[Kind.A]", "an unsupported computed method"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			want := probe.want
			if want == "" {
				want = "a computed field name"
			}
			index := strings.Index(probe.source, probe.at)
			if index < 0 {
				t.Fatalf("%q isn't in the source", probe.at)
			}
			// Where is one-based, by line and column, and every source here is ASCII.
			at := fmt.Sprintf(":%d:%d", strings.Count(probe.source[:index], "\n")+1, index-strings.LastIndex(probe.source[:index], "\n"))
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != want || !strings.HasSuffix(notYet.Where, at) {
				t.Fatalf("want NotYet %q at %s, got %v", want, at, err)
			}
		})
	}
}

// Two keys of one literal that spell one string never reach lowering: tsc refuses the pair (TS1117) by the
// name its type gives a computed key, an enum alias's included (grammarchecks.go's checkGrammarObjectLiteralExpression,
// through getEffectivePropertyNameForPropertyNameNode). So the ruling's (1) is held by tsc first and by
// distinctFieldKeys after it (TestDistinctFieldKeys).
func TestNumericKeyDuplicatesStopAtTheChecker(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"an enum alias", "enum Kind { A, B, C = A }\nconst table = { [Kind.A]: \"a\", [Kind.C]: \"c\" };\nconsole.log(table[Kind.A]);\n"},
		{"a named field beside its number", "const table = { \"0\": \"named\", [0]: \"keyed\" };\nconsole.log(table[0]);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := load.Load([]string{writeAgreementSource(t, "main.ts", probe.source)})
			if err == nil || !strings.Contains(err.Error(), "multiple properties with the same name") {
				t.Fatalf("want tsc to refuse the duplicate key, got %v", err)
			}
		})
	}
}

// distinctFieldKeys is the ruling's (1) where lowering stands: a key spelled twice keeps the stop. Mutant, red:
// its seen test dropped, so it answers true for every list.
func TestDistinctFieldKeys(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		keys     []string
		distinct bool
	}{
		{[]string{"0", "1", "7"}, true},
		{[]string{"0", "name", "0"}, false},
		{[]string{}, true},
	} {
		if got := distinctFieldKeys(probe.keys); got != probe.distinct {
			t.Errorf("distinctFieldKeys(%q) = %t, want %t", probe.keys, got, probe.distinct)
		}
	}
}

// A literal with three numeric keys walks the program once for its binding, however many keys ask, and each key
// is named by its canonical number string. A fresh lowering's answer is the held one. Mutant, red: the held answer
// dropped in numericKeyTableClosed, so each key walks again.
func TestNumericKeyTableIsWalkedOncePerBinding(t *testing.T) {
	t.Parallel()
	finder, _ := cycleIndexTypes(t, numericKeyProgram)
	held := finder.l
	var literal *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration && node.Name().Text() == "names" {
			literal = node.Initializer()
			return true
		}
		return node.ForEachChild(visit)
	}
	held.program.Files()[0].AsNode().ForEachChild(visit)
	if literal == nil {
		t.Fatal("no names declaration")
	}
	names := []string{}
	for index, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
		name, named := held.numericFieldName(literal, property)
		if !named {
			t.Fatalf("key %d not named", index)
		}
		names = append(names, name)
	}
	if strings.Join(names, ",") != "0,1,7" {
		t.Fatalf("names %q, want 0, 1 and 7", names)
	}
	if held.numericKeyTableWalks != 1 {
		t.Fatalf("the program walked %d times for one binding asked by three keys", held.numericKeyTableWalks)
	}
	fresh := &lowering{program: held.program, checker: held.checker, result: &ir.Program{}, this: -1, functionIndex: -1}
	if name, named := fresh.numericFieldName(literal, literal.AsObjectLiteralExpression().Properties.Nodes[0]); !named || name != "0" {
		t.Fatalf("a fresh lowering named the first key %q, %t", name, named)
	}
}

func TestNumericKeyEnumerationOrder(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		keys []string
		want bool
	}{
		{[]string{"250", "500", "2000"}, true},
		{[]string{"2000", "500", "250"}, false},
		{[]string{"250", "name", "2000"}, false},
		{[]string{"250", "8"}, false},
		{[]string{"0", "1", "7", "name"}, true},
		{[]string{"name"}, true},
		{nil, true},
		{[]string{"8", "08"}, true},
		{[]string{"08", "8"}, false},
		{[]string{"4294967294", "4294967295"}, true},
		{[]string{"4294967295", "4294967294"}, false},
		{[]string{"8", "8"}, false},
	} {
		if got := keysInEnumerationOrder(row.keys); got != row.want {
			t.Errorf("keysInEnumerationOrder(%q) = %t, want %t", row.keys, got, row.want)
		}
	}
}

func TestNumericKeyOrderedEnumPolling(t *testing.T) {
	t.Parallel()
	source := readEnumInitFixture(t, "../oracle/testdata/numeric_keys_enumeration_order.ts")
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if literal, ok := node.(ir.ObjectLiteral); ok && len(literal.Fields) == 3 {
				found = found || literal.Fields[0].Name == "250" &&
					literal.Fields[1].Name == "500" && literal.Fields[2].Name == "2000"
			}
			return true
		})
	}
	if !found {
		t.Fatal("no returned literal with field order 250,500,2000")
	}
}

func TestNumericKeyOrderedEnumStops(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		name, source string
	}{
		{"descending", `enum PollingInterval { High = 2000, Medium = 500, Low = 250 }
function make() {
    return { [PollingInterval.High]: 256, [PollingInterval.Medium]: 64, [PollingInterval.Low]: 32 };
}
console.log(Object.keys(make()).join(","));
`},
		{"interleaved", `enum PollingInterval { High = 2000, Medium = 500, Low = 250 }
function make() {
    return { [PollingInterval.Low]: 1, name: 2, [PollingInterval.High]: 3 };
}
console.log(Object.keys(make()).join(","));
`},
		{"spelled-index", `enum PollingInterval { High = 2000, Medium = 500, Low = 250 }
function make() {
    return { [PollingInterval.Low]: 1, "8": 2 };
}
console.log(Object.keys(make()).join(","));
`},
		{"numeric-literals", `enum PollingInterval { High = 2000, Medium = 500, Low = 250 }
function make() {
    return { [250]: 1, [500]: 2 };
}
console.log(Object.keys(make()).join(","));
`},
		{"const-enum", `const enum PollingInterval { High = 2000, Medium = 500, Low = 250 }
function make() {
    return { [PollingInterval.Low]: 1, [PollingInterval.Medium]: 2 };
}
console.log(Object.keys(make()).join(","));
`},
		{"string-enum", `enum Label { Name = "name" }
function make() {
    return { [Label.Name]: 1 };
}
console.log(Object.keys(make()).join(","));
`},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != "a computed field name" || !strings.HasSuffix(stop.Where, ":3:14") {
				t.Fatalf("want NotYet a computed field name at :3:14, got %v", err)
			}
		})
	}
}

func TestNumericKeyOrderedEnumReadiness(t *testing.T) {
	t.Parallel()
	const early = `function make() {
    return { [Pending.Low]: 1 };
}
function read(): void {
    const table = make();
    console.log(String(table[250]));
}
read();
enum Pending { Low = 250 }
`
	const unknown = `function make() {
    return { [Pending.Low]: initializeField() };
}
function read(): number {
    const table = make();
    return table[250];
}
function initializeField(): number { console.log("initializer"); return 1; }
function invoke(action: () => number): number { return action(); }
let action = (): number => 7;
action = read;
console.log("before");
console.log(String(invoke(action)));
enum Pending { Low = 250 }
`

	_, err := lowerSourceAs(t, early, ".ts")
	var stop *NotYet
	const reason = "reading an enum before its runtime initialization; move the call after the enum declaration"
	if !errors.As(err, &stop) || stop.What != reason || !strings.HasSuffix(stop.Where, ":2:15") {
		t.Fatalf("want known early enum read stopped at :2:15, got %v", err)
	}

	program, err := lowerSourceAs(t, unknown, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	checkedKey := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			literal, ok := node.(ir.ObjectLiteral)
			if !ok || len(literal.Fields) != 1 || literal.Fields[0].Name != "250" {
				return true
			}
			sequence, sequenced := literal.Fields[0].Value.(ir.Comma)
			if !sequenced {
				return true
			}
			walk(sequence.Left, func(node any) bool {
				if read, ok := node.(ir.Read); ok && read.Checked && program.Locals[read.Local].Name == "Pending" {
					checkedKey = true
				}
				return true
			})
			return true
		})
	}
	if !checkedKey {
		t.Fatal("returned literal must check the Pending enum at its key before the initializer")
	}
}
