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
// number, no enum's open value, in a member slot, which stops at the target's own kind before its parent's.
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
			"enum member slot Kind.Param"},
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

// viewCastUnbuiltPrelude is binder 3078's shape at small scale: the view reaches Holder's kind, a member slot,
// only through an optional map field whose own contract isn't built (#rgrtt32 measured an optional Map field as
// one: "an unavailable checked union member contract"), as tsc's view reaches HasJSDoc's members only through
// parent.parent, a union no checked contract places.
const viewCastUnbuiltPrelude = "enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other } " +
	"interface Node { readonly kind: Kind; readonly parent: Node | undefined; readonly extra?: ReadonlyMap<string, Node>; } " +
	"interface Holder extends Node { readonly kind: Kind.Name; } " +
	"interface Doc extends Node { readonly kind: Kind.Doc; readonly extra?: ReadonlyMap<string, Holder>; } " +
	"interface Tag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: Doc; } "

// A member slot every visit reaches only past a field whose contract isn't built is walked past, since no read
// or write through the view reaches it: the field's demanded read is NotYet at compile time, and its whole-value
// readers stay stopped (#rgrtt32). So binder 3078's and emitter 1888's casts lower, where c6 alone left the
// slot out and refused them at the relation. The read of the slot's own holder stops where the field is read,
// in #rgrtt32's words. Mutant M5 in view_cast_enum_slots.mutants.md, the bit never set past an unbuilt field,
// refuses the cast again.
func TestViewCastSlotsBehindAnUnbuiltFieldLower(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, viewCastUnbuiltPrelude+"function owner(node: Node): Kind { const tag = node as Tag; return tag.parent.kind; }", ".ts"); err != nil {
		t.Fatal(err)
	}
	_, err := lowerSourceAs(t, viewCastUnbuiltPrelude+"function holder(node: Node): number { const tag = node as Tag; const extra = tag.parent.extra; return extra === undefined ? 0 : extra.size; }", ".ts")
	var notYet *NotYet
	if err == nil || !errors.As(err, &notYet) || !strings.Contains(err.Error(), "whose contract isn't built") {
		t.Fatalf("want the unbuilt field's read NotYet, got %v", err)
	}
}

// A slot whose holder the view has no contract for, reached by a path that crosses no unbuilt field, stays out
// however its holder is named: a generic instance here, whose contract a lookup by the instance's own type id
// wouldn't find either way. Deciding the unbuilt arm by a holder's absence from such a lookup would walk past
// it; the walk decides it on its own path, so it stays refused (Lower's row). Mutant M4, the bit set past every
// field whatever its contract, lowers it and this one with it.
func TestViewCastSlotsUnderAnotherIdentityStayRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, viewCastEnumPrelude+
		"interface Callable<K extends Kind> extends Node { readonly kind: K; (value: number): number; } "+
		"interface GenericTag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: Callable<Kind.Doc>; } "+
		"function f(node: Node): Kind { return (node as GenericTag).parent.kind; }", ".ts")
	var refused *Refused
	var notYet *NotYet
	if err == nil || !errors.As(err, &refused) && !errors.As(err, &notYet) || !strings.Contains(err.Error(), "enum member slot Kind.Doc") {
		t.Fatalf("want a refusal naming %q, got %v", "enum member slot Kind.Doc", err)
	}
}
