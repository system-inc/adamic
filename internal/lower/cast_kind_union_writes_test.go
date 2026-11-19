package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// kindUnionWritesModel is the head 2 probe's shape: an open SyntaxKind on Node, and two kinded interfaces that
// both declare flags as a number and operand as a Node.
const kindUnionWritesModel = `enum SyntaxKind { Identifier = 80, StringLiteral = 11, NumericLiteral = 9, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface Identifier extends Node { readonly kind: SyntaxKind.Identifier; readonly text: string }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flags: number; operand: Node; readonly name: Identifier }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags: number; operand: Node; readonly index: number }
const id: Identifier = { kind: SyntaxKind.Identifier, pos: 1, text: 'a' };
const built: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 0, flags: 0, operand: id, index: 2 };
const held: Node = built;
`

// A named write through a kind-checked cast's union (#b5w3ycg head 2, Adamic's ruling of Nov 17) lowers when every
// declared type carrying one of the cast's kinds declares the field with one type: the probe's w1, through the cast,
// through a local and through an alias, of a number and of an object. A compound or logical assignment, an increment and a decrement are named writes too,
// reading and writing the one slot the ruling's fourth condition holds. Each agrees with Node, and every write of
// the field is a checked SetProperty, main's checked-write road, never a raw store.
func TestKindUnionWriteLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, write string }{
		{"object field", "function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).operand = held; }\nmark(held);\nconsole.log(String(built.operand.pos));\n"},
		{"through the cast", "function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }\nmark(held);\n"},
		{"through a local", "function mark(node: Node): void { const access = node as PropertyAccessExpression | ElementAccessExpression; access.flags = 5; }\nmark(held);\n"},
		{"through an alias", "type AccessExpression = PropertyAccessExpression | ElementAccessExpression;\nfunction mark(node: Node): void { const access = node as AccessExpression; access.flags = 5; }\nmark(held);\n"},
		{"compound", "function mark(node: Node): void { const access = node as PropertyAccessExpression | ElementAccessExpression; access.flags += 3; access.flags *= 2; }\nmark(held);\n"},
		{"increment and decrement", "function mark(node: Node): void { const access = node as PropertyAccessExpression | ElementAccessExpression; access.flags++; ++access.flags; access.flags--; }\nmark(held);\n"},
		{"logical", "function mark(node: Node): void { const access = node as PropertyAccessExpression | ElementAccessExpression; access.flags ||= 3; access.flags &&= 4; }\nmark(held);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, kindUnionWritesModel+probe.write+"console.log(String(built.flags));\n", ".ts")
			writes := 0
			inspect := func(node any) bool {
				if write, ok := node.(ir.SetProperty); ok && (write.Name == "flags" || write.Name == "operand") {
					writes++
					if !program.CheckedField(write.ViewReceiverTypeID, write.Name) {
						t.Errorf("the write of %s through the union isn't a checked write", write.Name)
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if writes == 0 {
				t.Fatal("found no write through the union")
			}
		})
	}
}

// What the ruling's fourth condition leaves NotYet, at compile time: a declared type carrying one of the cast's
// kinds whose layout holds no slot for the field (the probe's w2, BareElement), or holds it optional, waits under
// #x0nd1mh; one declaring the field readonly (Frozen) promised it never changes; two declared types sharing a kind
// and declaring the field with different types (the probe's w3, StringLiteral and NumericText, a valid program
// Node runs, an interface or an intersection alias, or one declaring it unknown, whose storage isn't the number's) can't take the write at all. An element write stays as head 1 left it, since head 2 ruled a named
// write alone; a delete is refused before it's judged, as any delete is.
func TestKindUnionWriteStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"holds no", kindUnionWritesModel + `interface BareElement extends Node { readonly kind: SyntaxKind.ElementAccessExpression }
const bare: BareElement = { kind: SyntaxKind.ElementAccessExpression, pos: 0 };
function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }
mark(bare);
`},
		{"holds it optional", kindUnionWritesModel + `interface LooseElement extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags?: number }
const loose: LooseElement = { kind: SyntaxKind.ElementAccessExpression, pos: 0 };
function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }
mark(loose);
`},
		{"declares it readonly", kindUnionWritesModel + `interface Frozen extends Node { readonly kind: SyntaxKind.ElementAccessExpression; readonly flags: number; operand: Node }
const frozen: Frozen = { kind: SyntaxKind.ElementAccessExpression, pos: 0, flags: 3, operand: id };
function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }
mark(frozen);
`},
		{"declared apart", `enum SyntaxKind { Identifier = 80, StringLiteral = 11, NumericLiteral = 9 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface StringLiteral extends Node { readonly kind: SyntaxKind.StringLiteral; text: string }
interface NumericText extends Node { readonly kind: SyntaxKind.StringLiteral; text: number }
interface NumericLiteral extends Node { readonly kind: SyntaxKind.NumericLiteral; text: string }
const odd: NumericText = { kind: SyntaxKind.StringLiteral, pos: 0, text: 7 };
function retext(node: Node): void { (node as StringLiteral | NumericLiteral).text = 'x'; }
retext(odd);
console.log(String(odd.text + 1));
`},
		{"declared apart by an intersection", `enum SyntaxKind { Identifier = 80, StringLiteral = 11, NumericLiteral = 9 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface StringLiteral extends Node { readonly kind: SyntaxKind.StringLiteral; text: string }
interface NumericLiteral extends Node { readonly kind: SyntaxKind.NumericLiteral; text: string }
type NumericText = Node & { readonly kind: SyntaxKind.StringLiteral; text: number };
const odd: NumericText = { kind: SyntaxKind.StringLiteral, pos: 0, text: 7 };
function retext(node: Node): void { (node as StringLiteral | NumericLiteral).text = 'x'; }
retext(odd);
console.log(String(odd.text + 1));
`},
		{"declared unknown", kindUnionWritesModel + "interface Loose extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags: unknown; operand: Node }\nfunction mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }\nmark(held);\n"},
		{"element write", kindUnionWritesModel + "function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression)['flags'] = 1; }\nmark(held);\n"},
	} {
		want := map[string][]string{
			"holds no":                          {"a write of flags through a kind-checked cast's union, where BareElement carries kind SyntaxKind.ElementAccessExpression but its layout holds no flags"},
			"holds it optional":                 {"a write of flags through a kind-checked cast's union, where LooseElement carries kind SyntaxKind.ElementAccessExpression but its layout holds no flags"},
			"declares it readonly":              {"a write of flags through a kind-checked cast's union, where Frozen carries kind SyntaxKind.ElementAccessExpression but declares flags readonly"},
			"declared apart":                    {"a write of text through a kind-checked cast's union, where ", " share kind SyntaxKind.StringLiteral but declare text as ", "StringLiteral", "NumericText"},
			"declared apart by an intersection": {"a write of text through a kind-checked cast's union, where StringLiteral and NumericText share kind SyntaxKind.StringLiteral but declare text as string and number"},
			"declared unknown":                  {"a write of flags through a kind-checked cast's union, where ", " share kind SyntaxKind.ElementAccessExpression but declare flags as ", "unknown"},
			"element write":                     {"other than a named assignment, increment or decrement, which #b5w3ycg head 2 leaves waiting"},
		}[probe.name]
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet at the write, got %v", err)
			}
			for _, words := range want {
				if !strings.Contains(err.Error(), words) {
					t.Fatalf("want NotYet containing %q, got %v", words, err)
				}
			}
		})
	}
}

// In .a the cast is refused before any write is judged: its tag test is a check no proof erases, and the fix
// narrows by the tag (head 1's refusal), so a named write through the union never reaches head 2's judgment there.
func TestKindUnionWriteInAStopsAtTheCast(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, kindUnionWritesModel+"function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }\nmark(held);\n", ".a")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "node.kind !== SyntaxKind.") || strings.Contains(err.Error(), "a write of flags") {
		t.Fatalf("want the cast's refusal, narrowing by the tag, got %v", err)
	}
}

// A carrier declared in a module the program only imports a type from is a carrier too: a value can be built at
// Odd though nothing of types.ts runs, and moduleOrder, holding the runnable modules alone, would leave it out, so
// condition four would never see Odd's string beside the members' number (Lower's must on head 2, J-zpw3w0m showed
// it lowering on c6159983). The index reads every file the program holds (load.Program.SourceFiles). Mutant: the
// index walks moduleOrder again, and this program lowers.
func TestKindUnionWriteJudgesTypeOnlyImports(t *testing.T) {
	t.Parallel()
	kinds := `export enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
export interface Node { readonly kind: SyntaxKind; readonly pos: number }
export interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flags: number }
export interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags: number }
`
	types := `import type { Node, SyntaxKind } from './kinds';
export interface Odd extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flags: string }
`
	main := `import { SyntaxKind } from './kinds';
import type { Node, PropertyAccessExpression, ElementAccessExpression } from './kinds';
import type { Odd } from './types';
const odd: Odd = { kind: SyntaxKind.PropertyAccessExpression, pos: 0, flags: 'x' };
function mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }
mark(odd);
console.log(typeof odd.flags);
`
	_, err := lowerModules(t, map[string]string{"kinds.ts": kinds, "types.ts": types, "main.ts": main})
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "Odd share kind SyntaxKind.PropertyAccessExpression but declare flags as") {
		t.Fatalf("want NotYet naming Odd beside PropertyAccessExpression, got %v", err)
	}
}

// A carrier whose field is any is refused before the write is judged, by the rule that every read of an any member
// through a typed view is unchecked, so "one written type" never meets an any (Lower's should on head 2).
func TestKindUnionWriteAnyCarrierIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, kindUnionWritesModel+"interface Loose extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flags: any; operand: Node }\nfunction mark(node: Node): void { (node as PropertyAccessExpression | ElementAccessExpression).flags = 5; }\nmark(held);\n", ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "a member typed any (flags, declared)") {
		t.Fatalf("want the any member's refusal, got %v", err)
	}
}
