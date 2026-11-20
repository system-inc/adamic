package lower

import (
	"errors"
	"strings"
	"testing"
)

// viewCastEnumPrelude is a Node like tsc's: an open kind, and a readonly parent, with Doc a member narrowing
// kind to one literal of it.
const viewCastEnumPrelude = "enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other } " +
	"interface Node { readonly kind: Kind; readonly parent: Node | undefined; } " +
	"interface Doc extends Node { readonly kind: Kind.Doc; } "

// A view cast's member-literal slots, reached through readonly fields its view has a contract for, are checked
// at each read, so tsc's checked downcasts lower (#snypn8c c6, Adamic's ruling under rulings.md:18, Lower's
// seam): a nested parent (binder 3078, emitter 1888), an overlap cast's own kind (binder 3369, moduleSpecifiers
// 915, parser 10804 and 10805, ts.ts 2191), a nested parent that is a union of members (emitter 1868), and a
// member two levels down (declarations 1808's name.expression.kind). Each was refused on main by widened's enum
// arms; viewCastEnumSlots names the slots and widenedPast walks past exactly those. Mutant M1 in
// view_cast_enum_slots.mutants.md, the slots dropped from the walk, turns each red.
func TestViewCastMemberSlotsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"a nested parent", viewCastEnumPrelude +
			"interface PropertyLikeTag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: Doc; readonly name: string; } " +
			"function owner(node: Node): Kind { const tag = node as PropertyLikeTag; return tag.parent.kind; }"},
		{"an overlap's own kind", viewCastEnumPrelude +
			"interface Identifier extends Node { readonly kind: Kind.Name; readonly text: string; readonly escaped: string; } " +
			"interface Wide extends Node { readonly flags: number; } " +
			"interface NameLike extends Node { readonly kind: Kind.Name; readonly text: string; } " +
			"function text(node: Identifier | Wide): string { const named = node as NameLike; return `${named.kind} ${named.text}`; }"},
		{"a union nested parent", viewCastEnumPrelude +
			"interface TypeLiteral extends Node { readonly kind: Kind.Literal; } " +
			"interface Tag extends Node { readonly kind: Kind.Tag | Kind.Param | Kind.Prop; readonly parent: Doc | TypeLiteral; } " +
			"function owner(node: Node): Kind { const tag = node as Tag; return tag.parent.kind; }"},
		{"two levels down", viewCastEnumPrelude +
			"interface Name extends Node { readonly expression: Node | undefined; } " +
			"interface Declaration extends Node { readonly name: Name | undefined; } " +
			"interface EntityName extends Node { readonly kind: Kind.Name; } " +
			"interface ComputedName extends Name { readonly kind: Kind.Computed; readonly expression: EntityName; } " +
			"interface LateBound extends Declaration { readonly name: ComputedName; } " +
			"function late(node: Declaration): Kind { const declaration = node as LateBound; return declaration.name.expression.kind; }"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSourceAs(t, probe.source, ".ts"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// What a view cast's walk leaves refused, each in widened's own words: a member slot reached through a writable
// parent, where a write through the view would store a value no read checks (mutant M3 turns it red); a slot
// typed as a whole enum, here a one-member enum's, which holds any number (mutant M2, every slot walked past,
// turns it red); a slot typed as a whole object type, the enum's object, rather than a member; a member slot
// whose holder the view has no contract for, a callable one, so nothing would check it at the read; and a
// number, no enum's open value, in a member slot.
func TestViewCastSlotsStayRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"a writable parent", "enum Kind { Root, Doc, Name } " +
			"interface WritableNode { readonly kind: Kind; parent: WritableNode | undefined; } " +
			"interface WritableDoc extends WritableNode { readonly kind: Kind.Doc; } " +
			"interface Target { readonly kind: Kind.Name; parent: WritableDoc; } " +
			"interface Member { readonly kind: Kind.Name; parent: WritableDoc; readonly member: number; } " +
			"interface Open { readonly kind: Kind; readonly open: number; } " +
			"interface Wider { readonly kind: Kind.Name; parent: WritableNode; readonly wider: number; } " +
			"function f(x: Member | Open | Wider): Kind { return (x as Target).parent.kind; }",
			"enum member slot Kind.Doc"},
		{"a whole enum slot", viewCastEnumPrelude + "enum Mode { Fast, Slow } enum Lone { Only } " +
			"interface Moded { readonly kind: Kind; readonly parent: Moded | undefined; readonly mode: Mode | Lone; } " +
			"interface LoneDoc extends Moded { readonly kind: Kind.Doc; readonly mode: Lone; } " +
			"interface LoneTag extends Moded { readonly kind: Kind.Param | Kind.Prop; readonly parent: LoneDoc; } " +
			"function f(node: Moded): Lone { return (node as LoneTag).parent.mode; }",
			"enum member slot"},
		{"a whole object slot", viewCastEnumPrelude +
			"interface Tabled { readonly kind: Kind; readonly parent: Tabled | undefined; readonly table: object; } " +
			"interface TableDoc extends Tabled { readonly kind: Kind.Doc; readonly table: typeof Kind; } " +
			"interface TableTag extends Tabled { readonly kind: Kind.Param | Kind.Prop; readonly parent: TableDoc; } " +
			"function f(node: Tabled): Kind { return (node as TableTag).parent.kind; }",
			"typeof Kind"},
		{"a holder with no contract", viewCastEnumPrelude +
			"interface CallableDoc extends Node { readonly kind: Kind.Doc; (value: number): number; } " +
			"interface CallableTag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: CallableDoc; } " +
			"function f(node: Node): Kind { return (node as CallableTag).parent.kind; }",
			"enum member slot Kind.Doc"},
		{"a number in a member slot", "enum Kind { Root, Doc, Param, Prop } " +
			"interface Counted { readonly kind: number; readonly parent: Counted | undefined; } " +
			"interface CountedDoc extends Counted { readonly kind: Kind.Doc; } " +
			"interface CountedTag extends Counted { readonly kind: Kind.Param | Kind.Prop; readonly parent: CountedDoc; } " +
			"function f(node: Counted): Kind { return (node as CountedTag).parent.kind; }",
			"Kind.Doc"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			var notYet *NotYet
			if err == nil || !errors.As(err, &refused) && !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want a refusal naming %q, got %v", probe.want, err)
			}
		})
	}
}
