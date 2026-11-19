package lower

import (
	"errors"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

// slotWritePrelude is viewOptionalWritePrelude's checked view, Element, with name holding undefined too, so a
// literal may write it undefined; Named, a plain interface holding name as Element does, for a literal to be built
// at; and rename, which writes name through the view and reads it back through it.
const slotWritePrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string | undefined; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
interface Named extends Node { name?: string | undefined; }
function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
`

// slotWriteRow is a program judged by checkSlotWrites: want is its NotYet's words exactly, or empty for a
// program that lowers and agrees with Node on the JavaScript backend and natively, sanitized, where it prints
// node.
type slotWriteRow struct{ name, source, want, node string }

// judgeSlotWrites runs each row as its own parallel subtest, lowered with extension.
func judgeSlotWrites(t *testing.T, extension string, rows []slotWriteRow) {
	t.Helper()
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			if row.want == "" {
				program := lowersAndAgreesWithNodeAs(t, row.source, extension)
				if code, out, errs := brandedNative(t, program); code != 0 || out != row.node || errs != "" {
					t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q where Node prints %q", code, out, errs, row.node)
				}
				return
			}
			_, err := lowerSourceAs(t, row.source, extension)
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet %q at the write, got %v", row.want, err)
			}
			if notYet.What != row.want {
				t.Fatalf("want NotYet %q, got %q", row.want, notYet.What)
			}
		})
	}
}

// The view's NotYet opens so, reachHoldsSlot's words following.
const slotWriteViewOpening = "a write to optional field name through a checked view, where "

// A write through a checked view to an optional field is admitted where every value reaching it holds the
// field's slot at the view's type of it (#w8nanzw, Compiler's accept of Nov 19): an instance of a class declaring
// name, or of a class whose base declares it; a literal built at Named that wrote name, undefined or a value. Each
// lowers and agrees with Node, natively too. A literal spreading one that did holds the slot too, but stops at
// main's own whole-object copy of a value that may hold a checked view's. A literal that left name out
// holds no slot, and neither does one spreading such a literal, so each is NotYet at the write in the helper's
// words, and so is a write whose reach isn't known, in a callback handed to forEach or in a function nothing
// calls. A literal built with no contextual type holds name at its own type, a string, which isn't the view's
// string | undefined both ways, so it's NotYet too, the price of judging the slot's type as head 2 does. Mutant:
// reachHoldsSlot's slot test answering true, and literal without goes red.
func TestSlotWritesThroughACheckedView(t *testing.T) {
	t.Parallel()
	prelude := slotWritePrelude
	unknown := slotWriteViewOpening + "what reaches its receiver isn't known"
	judgeSlotWrites(t, ".ts", []slotWriteRow{
		{"class", prelude + "class Box { readonly kind = 171; readonly pos = 1; name?: string | undefined; label = 'l'; }\nconst box = new Box();\nrename(box);\n", "", "w\n"},
		{"base class", prelude + "class Base { readonly kind = 171; readonly pos = 1; name?: string | undefined; }\nclass Derived extends Base { label = 'l'; }\nconst derived = new Derived();\nrename(derived);\n", "", "w\n"},
		{"literal wrote undefined", prelude + "const raw: Named = { kind: 171, pos: 1, name: undefined };\nrename(raw);\n", "", "w\n"},
		{"literal wrote a value", prelude + "const raw: Named = { kind: 171, pos: 1, name: 'a' };\nrename(raw);\n", "", "w\n"},
		// The slot holds, but main's own stop for a whole-object copy of a value that may hold a checked view's
		// value comes after (checkViewCopies), as head 2's spread twin met it; the plain road's spread agrees.
		{"spread holding", prelude + "const base: Named = { kind: 171, pos: 1, name: 'a' };\nconst raw: Named = { ...base };\nrename(raw);\n", "a whole-object copy of a value of type Named that may hold a checked view's value", ""},
		{"built without a context", prelude + "const raw = { kind: 171, pos: 1, name: 'a' };\nrename(raw);\n", slotWriteViewOpening + "a value reaching it is built as { kind: number; pos: number; name: string; }, whose name is string, not string | undefined", ""},
		{"literal without", prelude + "const raw = { kind: 171, pos: 1 };\nrename(raw);\n", slotWriteViewOpening + "a value reaching it is built as { kind: number; pos: number; } without name", ""},
		// Built at Named, which declares name, so the built type holds it and only the literal's own fields and
		// own type show it never wrote it. A literal's own type is its written fields, so the own-type judgment
		// holds this row too; the slot test alone holds a class whose member an interface merges in
		// (view_optional_writes_test.go's merged class member, mutant (a)).
		{"literal at Named without", prelude + "const raw: Named = { kind: 171, pos: 1 };\nrename(raw);\n", "writing a possibly absent optional own field, where a value reaching it is built as Named without name", ""},
		{"spread lacking", prelude + "const bare = { kind: 171, pos: 1 };\nconst raw = { ...bare };\nrename(raw);\n", slotWriteViewOpening + "a value reaching it is built as { kind: number; pos: number; }, whose layout holds no name", ""},
		{"one of two without", prelude + "const held: Named = { kind: 171, pos: 1, name: 'a' };\nconst bare = { kind: 171, pos: 2 };\nrename(held);\nrename(bare);\n", slotWriteViewOpening + "a value reaching it is built as { kind: number; pos: number; } without name", ""},
		{"forEach", prelude + "const raw: Named = { kind: 171, pos: 1, name: 'a' };\n[raw].forEach(rename);\n", unknown, ""},
		// A receiver only undefined reaches: the store would fault where Node throws, so nothing reaching it is NotYet.
		{"nothing reaches", prelude + "const raw = undefined as Named | undefined;\nrename(raw!);\n", slotWriteViewOpening + "nothing the program builds reaches its receiver", ""},
		{"uncalled", prelude + "const raw: Named = { kind: 171, pos: 1, name: 'a' };\nconsole.log(String(raw.name));\n", unknown, ""},
	})
}

// A plain write to an optional own field some literal leaves out (absentOptionalFieldWrite) is admitted where
// every value reaching it holds the slot, in .a as in .ts, since a reach proof is a proof: put writes name on
// held, which wrote it, beside empty, which left it out and never reaches the write, and on a copy spreading held,
// whose slot is held's. Handed empty, or a copy spreading it, the write is NotYet in its own words and the
// helper's.
func TestSlotWritesOfAPlainOptionalField(t *testing.T) {
	t.Parallel()
	const model = "interface Mut { name?: string; }\nconst empty: Mut = {};\nconst held: Mut = { name: 'a' };\nfunction put(x: Mut): void { x.name = 'w'; }\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			judgeSlotWrites(t, extension, []slotWriteRow{
				{"held", model + "put(held);\nconsole.log(String(held.name) + ' ' + String(empty.name));\n", "", "w undefined\n"},
				{"empty", model + "put(empty);\nconsole.log(String(empty.name));\n", "writing a possibly absent optional own field, where a value reaching it is built as Mut without name", ""},
				{"spread of held", model + "const copy: Mut = { ...held };\nput(copy);\nconsole.log(String(copy.name) + ' ' + String(held.name));\n", "", "w a\n"},
				{"spread of empty", model + "const copy: Mut = { ...empty };\nput(copy);\nconsole.log(String(copy.name));\n", "writing a possibly absent optional own field, where a value reaching it is built as Mut without name", ""},
			})
		})
	}
}

// A slot that's there must take the write's type too (Compiler's must on #w8nanzw). count built undefined in a
// literal with no contextual type is an undefined-represented slot, which a number written through `count?:
// number | undefined` doesn't fit. Handed to Counted directly, invariance refuses it first, since Counted can
// write number | undefined where the literal's undefined is read (the invariance row, in .a and .ts), so that
// road is held twice. Through a cast invariance can't see it: seen as the open Node, which declares no count, and
// cast back to Counted by its kind, a checked view, so only the type judgment stops it, NotYet in the helper's
// words, never a stop or a misread at run time. Its honest twin, count 1 built at Counted, lowers and agrees. next
// built undefined at its contextual `next?: Node | undefined` is that reference field's own representation, so a
// Node written there lowers and agrees. Mutant: slotWriteSite skipped for every write but a kind union's, and the
// cast row lowers.
func TestSlotWritesHoldTheFieldsType(t *testing.T) {
	t.Parallel()
	const counted = "interface Counted { count?: number | undefined; }\nconst empty: Counted = {};\nconst raw = { count: undefined };\nfunction bump(x: Counted): void { x.count = 5; }\nbump(raw);\nconsole.log(String(raw.count) + ' ' + String(empty.count));\n"
	const holder = "interface Node { readonly id: number; }\ninterface Holder { next?: Node | undefined; }\nconst empty: Holder = {};\nconst held: Holder = { next: undefined };\nfunction put(holder: Holder, next: Node): void { holder.next = next; }\nput(held, { id: 1 });\nconsole.log(typeof held.next + ' ' + String(held.next?.id) + ' ' + typeof empty.next);\n"
	// cast seen as the open Node and cast back to Counted by its kind, a checked view in .ts, which .a refuses.
	const cast = "interface Node { readonly kind: number; }\ninterface Counted extends Node { readonly kind: 7; count?: number | undefined; }\nfunction bump(node: Node): void { if (node.kind === 7) { (node as Counted).count = 5; } }\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			t.Run("count built undefined", func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, counted, extension)
				var refused *Refused
				const want = "a value of type { count: undefined; } seen as Counted, which can write number | undefined where undefined is read"
				if !errors.As(err, &refused) || !strings.Contains(err.Error(), want) {
					t.Fatalf("want invariance's refusal %q, got %v", want, err)
				}
			})
			rows := []slotWriteRow{{"next built undefined at its context", holder, "", "object 1 undefined\n"}}
			if extension == ".ts" {
				rows = append(rows,
					slotWriteRow{"count built undefined through a cast", cast + "const lie = { kind: 7, count: undefined };\nconst seen: Node = lie;\nbump(seen);\nconsole.log(String(lie.count));\n", "a write to optional field count through a checked view, where a value reaching it is built as { kind: number; count: undefined; }, whose count is undefined, not number | undefined", ""},
					slotWriteRow{"count built 1 at Counted through a cast", cast + "const honest: Counted = { kind: 7, count: 1 };\nconst seen: Node = honest;\nbump(seen);\nconsole.log(String(honest.count));\n", "", "5\n"},
				)
			}
			judgeSlotWrites(t, extension, rows)
		})
	}
}

// slotWriteKindModel is head 2's model with flowNode, as tsc declares it, optional on both kinds, and mark
// writing it through the kind-checked cast's union.
const slotWriteKindModel = `enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flowNode?: number | undefined }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flowNode?: number | undefined }
function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flowNode = 5; }
`

// A write through a kind-checked cast's union to a field every member declares optional alike is judged as any
// optional field's is (#b5w3ycg head 2, #w8nanzw): a literal that wrote flowNode, as tsc's factory writes
// `node.flowNode = undefined`, holds its slot, so the write lowers and agrees; one built at Node without it holds
// none, and the write is NotYet at the write.
func TestSlotWritesThroughAKindUnion(t *testing.T) {
	t.Parallel()
	judgeSlotWrites(t, ".ts", []slotWriteRow{
		{"written undefined", slotWriteKindModel + "const odd: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 0, flowNode: undefined };\nmark(odd);\nconsole.log(String(odd.flowNode));\n", "", "5\n"},
		{"built at Node", slotWriteKindModel + "const odd: Node = { kind: SyntaxKind.ElementAccessExpression, pos: 0 };\nmark(odd);\n", "a write of flowNode through a kind-checked cast's union, where a value reaching it is built as Node without flowNode", ""},
	})
}

// Writing undefined into a checked slot that may hold an object stays Rule C's NotYet, whatever reach says: node
// was built with flowNode, so the slot is there, but the checked write can't change the representation its slot
// holds (recordSourceSlotWrite). Mutant: Rule C's test of a field that may hold undefined
// or null dropped, and this lowers.
func TestSlotWriteOfUndefinedIntoAnObjectStaysRuleC(t *testing.T) {
	t.Parallel()
	judgeSlotWrites(t, ".ts", []slotWriteRow{
		{"undefined into an object", `interface FlowNode { readonly id: number }
interface Node { readonly kind: "a" | "b"; flowNode?: FlowNode | undefined }
interface NodeA extends Node { readonly kind: "a" }
const node: Node = { kind: "a", flowNode: { id: 1 } };
const view = node as NodeA;
view.flowNode = undefined;
console.log(String(node.flowNode === undefined));
`, "writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds", ""},
	})
}

// slotWriteCallers is every function that records a write for checkSlotWrites, each a road an optional field's
// slot is asked of, and slotWriteJudges every function that asks reachHoldsSlot, the one answer.
var (
	slotWriteCallers = []string{"absentOptionalFieldWrite", "kindUnionWrites", "viewWriteReceiverTypeID"}
	slotWriteJudges  = []string{"checkSlotWrites", "reachHoldsSlot"}
)

// Every write that waits on an optional field's slot asks one judgment, reachHoldsSlot, never a sibling (Compiler's
// condition 1 on #w8nanzw): the three roads record their writes (recordSlotWrite) and nothing else stops them on
// the slot, and only checkSlotWrites asks the helper, the judgment of the type each reaching object holds the field
// at riding on its walk for every write, never skipped (slotWriteSite). The test reads this package's source: a
// fourth road recording a write is named here, a function asking reachHoldsSlot on its own is a second judgment,
// reachHoldsSlot may not skip the type for a write that brings none, and head 2's declared index and the
// built-type test may no longer refuse an optional field themselves, nor absentOptionalFieldWrite stop the write a
// literal left out.
func TestSlotWritesHaveOneJudgment(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	calls := map[string]map[string]bool{}
	optionals := map[string]int{}
	notYets := map[string]int{}
	eachNil := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*goast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			name := function.Name.Name
			goast.Inspect(function.Body, func(node goast.Node) bool {
				switch node := node.(type) {
				case *goast.CallExpr:
					if selector, isSelector := node.Fun.(*goast.SelectorExpr); isSelector {
						if calls[selector.Sel.Name] == nil {
							calls[selector.Sel.Name] = map[string]bool{}
						}
						calls[selector.Sel.Name][name] = true
						if selector.Sel.Name == "notYet" {
							notYets[name]++
						}
					}
				case *goast.SelectorExpr:
					if node.Sel.Name == "SymbolFlagsOptional" {
						optionals[name]++
					}
				case *goast.BinaryExpr:
					left, isLeft := node.X.(*goast.Ident)
					right, isRight := node.Y.(*goast.Ident)
					if isLeft && isRight && left.Name == "each" && right.Name == "nil" {
						eachNil[name] = true
					}
				}
				return true
			})
		}
	}
	callers := func(callee string) []string {
		var names []string
		for name := range calls[callee] {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}
	if got := callers("recordSlotWrite"); !slices.Equal(got, slotWriteCallers) {
		t.Errorf("recordSlotWrite's callers are %v, want %v: a road that waits on an optional field's slot records its write and is named here", got, slotWriteCallers)
	}
	if got := callers("reachHoldsSlot"); !slices.Equal(got, slotWriteJudges) {
		t.Errorf("reachHoldsSlot's callers are %v, want %v: a function asking it on its own is a second slot judgment", got, slotWriteJudges)
	}
	if got := callers("slotWriteSite"); !slices.Equal(got, []string{"checkSlotWrites"}) {
		t.Errorf("slotWriteSite's callers are %v, want checkSlotWrites alone: the type each reaching object holds the field at rides on reachHoldsSlot's walk, for every write", got)
	}
	for _, name := range []string{"kindUnionWriteHeld", "slotWriteBuilt"} {
		if optionals[name] != 0 {
			t.Errorf("%s reads SymbolFlagsOptional again: whether an object holds an optional field's slot is reachHoldsSlot's to say", name)
		}
	}
	if eachNil["reachHoldsSlot"] {
		t.Errorf("reachHoldsSlot asks whether each is nil: the type a reaching object holds the field at is judged for every write, never skipped")
	}
	// The one stop absentOptionalFieldWrite keeps is a field held in a Union, which is representation.
	if notYets["absentOptionalFieldWrite"] != 1 {
		t.Errorf("absentOptionalFieldWrite makes %d NotYets, want 1, its Union field's: a write some literal leaves the field out of is checkSlotWrites' to stop", notYets["absentOptionalFieldWrite"])
	}
}

// A member an interface merges into a class has no slot in the class's instances, which lay out only the fields a
// class declares, so a write of it is judged by reach and stops, written through an interface or through the class's
// own type (#86cmqrp, Language's attack on #w8nanzw: it lowered, and native panicked "a field the checker proved is
// there is missing" where Node runs on). A member the class declares optional is written as before.
func TestSlotWritesOfAMergedClassMember(t *testing.T) {
	t.Parallel()
	const merged = "class Merged { a = 1; }\ninterface Merged { f?: string; }\ninterface X { a: number; f?: string; }\n"
	for _, row := range []struct{ name, source string }{
		{"through an interface", merged + "function put(m: X): void { m.f = 'x'; }\nconst m = new Merged();\nput(m);\nconsole.log(String(m.f));\n"},
		{"through the class", merged + "function put(m: Merged): void { m.f = 'x'; }\nconst m = new Merged();\nput(m);\nconsole.log(String(m.f));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "an instance of Merged, whose class holds no f") {
				t.Fatalf("want NotYet naming Merged's missing slot for f, got %v", err)
			}
		})
	}
	t.Run("declared by the class", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, "class Declared { a = 1; f?: string; }\ninterface X { a: number; f?: string; }\nfunction put(m: X): void { m.f = 'x'; }\nconst m = new Declared();\nput(m);\nconsole.log(String(m.f));\n", ".ts")
	})
}
