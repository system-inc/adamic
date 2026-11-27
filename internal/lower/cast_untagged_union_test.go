package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A cast from a union of plain objects no literal tag tells apart, to one member or narrower (#axqgdck,
// untaggedUnionArm), lowers in .ts to a checked view of the target with no test at the cast (rulings.md:18):
// no CheckedCast, and its view marked untagged. A source holding undefined beside two or more such members
// takes absentArm's presence check first, whose message is the only cast message. tsc's shapes: a member of
// PackageJsonInfo | MissingPackageJsonInfo, Node out of FileReference | Node, and an Identifier out of
// Declaration | Expression, whose kind is open on both. A union target (union target (a), #b5w3ycg) is a view of
// the union read through the fields every target member declares in one representation: two members of three,
// and tsc's objectFlags read through ObjectType | UnionType (utilities.ts:8280). Mutants: untaggedUnionArm returning
// false reds every row; its gate held to one target again reds the two union-target rows alone.
func TestUntaggedUnionCastLowers(t *testing.T) {
	t.Parallel()
	members := "interface A { readonly a: number; readonly shared: string } interface B { readonly b: string; readonly shared: string } "
	for _, probe := range []struct {
		name, source string
		absent       bool
	}{
		{"a member", members + "function f(x: A | B): A { return x as A; } console.log(String(f({ a: 1, shared: 's' }).a));", false},
		{"a member out of three", members + "interface C { readonly c: boolean } function f(x: A | B | C): B { return x as B; } console.log(f({ b: 'b', shared: 's' }).b);", false},
		{"narrower than a member", "interface Declaration { readonly kind: number; readonly pos: number } interface Expression { readonly kind: number; readonly end: number } interface Identifier extends Expression { readonly text: string } function f(x: Declaration | Expression): Identifier { return x as Identifier; } console.log(f({ kind: 80, end: 1, text: 't' } as Identifier).text);", false},
		{"a member beside undefined", members + "function f(x: A | B | undefined): A { return x as A; } console.log(String(f({ a: 1, shared: 's' }).a));", true},
		// An optional literal field is no tag (fieldLiteral answers nil for kind?: 'a'), so the union takes this arm,
		// held by the checked reads; recorded here so a ruling either way moves this row (Lower's should).
		{"a union target", members + "interface C { readonly c: boolean } function f(x: A | B | C): string { return (x as A | B).shared; } console.log(f({ a: 1, shared: 's' }));", false},
		{"tsc's objectFlags through a union target", "interface Type { readonly flags: number } interface ObjectType extends Type { readonly objectFlags: number } interface UnionType extends Type { readonly types: readonly Type[]; readonly objectFlags: number } interface IntrinsicType extends Type { readonly intrinsicName: string } function f(t: ObjectType | UnionType | IntrinsicType): number { return (t as ObjectType | UnionType).objectFlags; } console.log(String(f({ flags: 1, objectFlags: 4 })));", false},
		// The arm judges the widening rule against the cast's whole target (untaggedUnionArm), which on keepsNullishCast's
		// road holds undefined beside the member: it lowers as it did when the rule judged the member alone (Lower's
		// question on #b5w3ycg; J-q94k2e4 ran it at both). The one-target mutant leaves it green.
		{"a member keeping undefined", members + "interface C { readonly c: boolean } function f(x: A | B | C | undefined): A | undefined { return x as A | undefined; } console.log(String(f({ a: 1, shared: 's' })?.a));", false},
		{"an optional literal field", "interface A { readonly kind?: 'a'; readonly a: number } interface B { readonly kind?: 'b'; readonly b: string } function f(x: A | B): A { return x as A; } console.log(String(f({ kind: 'a', a: 1 }).a));", false},
		// One plain object cast to a union target takes the same view, the source as the union's one member (cause U
		// of #eg3krrp, castRoad): tsc's objectFlags read through ObjectType | UnionType of a Type (utilities.ts:8319),
		// and errorCount through a union of build infos of a BuildInfo whose members hold no array field. Mutants: castRoad without its
		// union target arm, or untaggedUnionArm's gate wanting two members again, red both rows alone.
		{"one object to a union target", "interface Type { readonly flags: number } interface ObjectType extends Type { readonly objectFlags: number } interface UnionType extends Type { readonly types: readonly Type[]; readonly objectFlags: number } function f(t: Type): number { return (t as ObjectType | UnionType).objectFlags; } const o: ObjectType = { flags: 1, objectFlags: 4 }; console.log(String(f(o)));", false},
		{"one object to a build info union target", "interface BuildInfo { readonly version: string } interface BundleInfo extends BuildInfo { readonly outSignature?: string; readonly errorCount: number } interface ProjectInfo extends BuildInfo { readonly root: number; readonly errorCount: number } function f(info: BuildInfo): number { return (info as BundleInfo | ProjectInfo).errorCount; } const b: BundleInfo = { version: '1', errorCount: 1 }; console.log(String(f(b)));", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowered, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatal(err)
			}
			for _, function := range lowered.Functions {
				walk(function.Body, func(node any) bool {
					if cast, ok := node.(ir.CheckedCast); ok {
						t.Fatalf("a CheckedCast on %s: an untagged union's cast is a view with no test at the cast", cast.Field)
					}
					return true
				})
			}
			messages := 0
			for _, text := range lowered.Strings {
				if strings.HasPrefix(text, "cast failed:") {
					messages++
				}
			}
			if want := map[bool]int{false: 0, true: 1}[probe.absent]; messages != want {
				t.Fatalf("%d cast messages, want %d: only absentArm's presence check stops at the cast", messages, want)
			}
		})
	}
}

// What the arm leaves to the stops it found, each by its own words since the cast causes (#b5w3ycg): a member
// that isn't a plain object (a primitive, a class instance, a dictionary), a sub-union target, and a union a tag
// does tell apart, whose payload refinement stays refused (rulings.md, the object-union member entry). A union
// whose tags collide (duplicate enum aliases, mixed tag representations) stays refused too, pinned by
// TestUncheckableCastsStayRefused. In .a the arm's cast is refused as every downcast whose checks no proof
// erases. tsc refuses a sideways cast whose types don't overlap itself (TS2352), so no row holds one. A row
// with no words holds only that the arm leaves it stopped. Mutants: plainViewObject returning true reds the
// class row (a primitive member and a dictionary member are stopped before the arm, J-7j0yf17); unionHoldsLiterals
// returning false reds "a tagged union's refinement" and TestUncheckableCastsStayRefused's colliding-tag rows;
// keptCheckDowncastRefusal without proof.untagged reds "in .a".
func TestUntaggedUnionCastsKeepTheirStops(t *testing.T) {
	t.Parallel()
	members := "interface A { readonly a: number; readonly shared: string } interface B { readonly b: string; readonly shared: string } "
	for _, probe := range []struct {
		name, source, extension string
		refused                 bool
		want                    string
	}{
		{"a primitive member", members + "function f(x: A | string): A { return x as A; }", ".ts", false, "which holds primitives"},
		{"a class instance member", members + "class K { readonly k = 1; } function f(x: A | K): A { return x as A; }", ".ts", false, "whose members no literal tag tells apart"},
		{"a dictionary member", members + "function f(x: A | Record<string, number>): A { return x as A; }", ".ts", false, ""},
		// A target that holds a literal the source doesn't one field down (FlowCall, whose node is a CallExpression),
		// and a readonly field the target makes writable (FlowLabel's antecedent): the unchecked-widening rule would
		// refuse each, so the arm leaves each at the untagged stop and never turns it Refused. A literal at the top,
		// binder.ts's CallExpression, is the cast's tag, tested there, so it lowers (TestUntaggedUnionCastTestsItsTag,
		// #qxt10dd). Mutant: untaggedUnionArm without its write judgment refuses both.
		{"a target with a literal one field down", "enum SyntaxKind { Identifier = 80, CallExpression = 213 } interface Node { readonly kind: SyntaxKind } interface CallExpression extends Node { readonly kind: SyntaxKind.CallExpression } interface FlowStart { readonly flags: number; readonly node: Node | undefined } interface FlowCall { readonly flags: number; readonly node: CallExpression } function f(x: FlowStart | FlowCall): FlowCall { return x as FlowCall; }", ".ts", false, "a checked view of FlowCall over the union FlowCall | FlowStart, whose members no literal tag tells apart"},
		// FlowUnreachable lacks antecedent, so the union isn't assignable to FlowLabel and takes unionMemberProof, as
		// tsc's FlowNode does; with two members it's an upcast in TypeScript's eyes and the relation refuses it first.
		{"a readonly field the target makes writable", "interface FlowStart { readonly flags: number; readonly antecedent: FlowStart | undefined } interface FlowLabel { readonly flags: number; antecedent: FlowStart | undefined } interface FlowUnreachable { readonly flags: number } function f(x: FlowStart | FlowLabel | FlowUnreachable): FlowLabel { return x as FlowLabel; }", ".ts", false, "whose members no literal tag tells apart"},
		// A union target lowers where its members agree on a field's representation (TestUntaggedUnionCastLowers). A
		// field they hold differently stops at its contract (supportsUntaggedRead), union target (b)'s, and a target
		// tagged by a literal every member holds is the tag road's, keeping the untagged stop.
		{"a union target whose shared field differs", "interface A { readonly a: number; readonly shared: string } interface B { readonly b: string; readonly shared: number } interface C { readonly c: boolean } function f(x: A | B | C): string { return String((x as A | B).shared); }", ".ts", false, "checked view field shared of type string | number"},
		{"a tagged union target", "interface A { readonly kind: 'a'; readonly shared: string } interface B { readonly kind: 'b'; readonly shared: string } interface C { readonly c: boolean } function f(x: A | B | C): string { return (x as A | B).shared; }", ".ts", false, "a checked view of A | B over the union A | B | C, whose members no literal tag tells apart"},
		// Runtime's two questions on #axqgdck: a writable field whose type differs by arm is refused at the cast by
		// castUnionWrites (its union type X | Y fits neither arm's slot), so no X is ever written into a B and no
		// cycle the finder can't see closes through one. One nested in a readonly field, which castUnionWrites
		// doesn't walk, is stopped too (overlapViewCast's per-member widening).
		{"a writable field whose type differs by arm", "interface X { readonly x: number } interface Y { readonly y: string } interface A { note: X } interface B { note: Y } function f(u: A | B): A { return u as A; }", ".ts", true, "whose writable field note some member can't hold every value of"},
		{"a writable field nested in a readonly one", "interface X { readonly x: number } interface Y { readonly y: string } interface CX { v: X } interface CY { v: Y } interface A { readonly child: CX } interface B { readonly child: CY } function f(u: A | B): A { return u as A; }", ".ts", false, ""},
		{"a tagged union's refinement", "interface P { readonly kind: 'p'; readonly v: string | number } interface Q { readonly kind: 'q' } interface PN { readonly kind: 'p'; readonly v: number } function f(x: P | Q): PN { return x as PN; }", ".ts", true, "which is narrower than a member of P | Q"},
		{"in .a", members + "function f(x: A | B): A { return x as A; }", ".a", true, "in .a whose checks no proof erases"},
		// One object cast to a union target (cause U of #eg3krrp) keeps its stop where the arm's own tests fail: a
		// target member no narrower than the source is no downcast, and in .a the view's checks are refused as above.
		{"one object to a union target with a sideways member", "interface Base { readonly flags: number } interface A extends Base { readonly a: number } interface Z { readonly z: string } function f(x: Base): A | Z { return x as A | Z; }", ".ts", false, "a checked view of the union A | Z over Base, whose members no literal tag tells apart"},
		// tsc's own build infos hold arrays (fileNames, root), and the target union's view builds a contract for every
		// field its members declare, so the cast stops at an array field's contract, which isn't this arm's: builder.ts:1207
		// moves there from the untagged stop.
		{"one object to a union target holding an array field", "interface BuildInfo { readonly version: string } interface BundleInfo extends BuildInfo { readonly fileNames: readonly string[]; readonly errorCount: number } interface ProjectInfo extends BuildInfo { readonly fileNames: readonly string[]; readonly root: readonly number[]; readonly errorCount: number } function f(info: BuildInfo): number { return (info as BundleInfo | ProjectInfo).errorCount; }", ".ts", false, "checked view field fileNames of type readonly string[]"},
		{"one object to a union target in .a", "interface Base { readonly flags: number } interface A extends Base { readonly a: number } interface B extends Base { readonly b: string } function f(x: Base): number { return (x as A | B).flags; }", ".a", true, "in .a whose checks no proof erases"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var refused *Refused
			var notYet *NotYet
			switch {
			case err == nil:
				t.Fatal("lowered: the arm must leave this cast stopped")
			case probe.want == "":
				if !errors.As(err, &refused) && !errors.As(err, &notYet) {
					t.Fatalf("want a stop, got %v", err)
				}
			case probe.refused && (!errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want a refusal holding %q, got %v", probe.want, err)
			case !probe.refused && (!errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.want)):
				t.Fatalf("want NotYet holding %q, got %v", probe.want, err)
			}
		})
	}
}

// In a graph program the untagged view pairs each arm of its source with its target (Memory's item 4 on
// #axqgdck): a B's inner is read and written as A's through the view, so a Y the finder never relates to X
// stops the program NotYet, naming the arm, whatever road a later read or write takes (Language's x05 and
// x06: a parameter and an alias typed A). C has no inner, so the union holds none in common and only the arm
// loop pairs it. The same union whose B holds an X lowers. Mutant: untaggedUnionCast not marking its view
// untagged reds "an arm the finder doesn't relate".
func TestUntaggedUnionCastInAGraphProgram(t *testing.T) {
	t.Parallel()
	// X and Y hold leaves of their own, so the view's contract never reaches the cycle's Link.
	declarations := "interface XLeaf { readonly v: number } interface YLeaf { readonly w: string } interface X { readonly x: XLeaf } interface Y { readonly y: YLeaf } interface A { readonly a: number; readonly inner: X } interface C { readonly c: number } "
	use := "function f(u: A | B | C): X { return (u as A).inner; } console.log(String(f({ a: 1, inner: { x: { v: 4 } } }).x.v) + String(link.next === link));"
	for _, probe := range []struct{ name, b, want string }{
		{"an arm the finder doesn't relate", "interface B { readonly b: number; readonly inner: Y } ", "through an arm no test tells (B as A) through its field inner (Y as X)"},
		{"an arm whose fields pair", "interface B { readonly b: number; readonly inner: X } ", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, graphReviewCycle+declarations+probe.b+use, ".ts")
			if probe.want == "" {
				if err != nil {
					t.Fatalf("an untagged view whose arms pair, in a graph program: %v", err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.want) {
				t.Fatalf("want NotYet holding %q, got %v", probe.want, err)
			}
		})
	}
}

// untaggedTagModel is binder.ts:1416's shape (#qxt10dd): every member of Expression | Statement declares kind as
// the whole open SyntaxKind, through Node, and CallExpression narrows it to one member, so no literal tag tells
// the union's members apart and the target's kind is the tag its cast tests.
const untaggedTagModel = "enum SyntaxKind { Identifier, CallExpression, Block } interface Node { readonly kind: SyntaxKind } interface Expression extends Node {} interface Statement extends Node {} interface CallExpression extends Expression { readonly kind: SyntaxKind.CallExpression; readonly argumentCount: number } "

// A cast from an untagged union to a target that declares a literal tag the members hold open, tsc's `node as
// CallExpression` (#qxt10dd), lowers as a tag test at the cast and the target's checked view: a true claim reads
// through it and agrees with Node, on the sanitized native too. Mutant: proof.untagged left out of castRefusal's
// viewCasts, so stringWideningRefusal judges the cast as a plain widening and refuses the open kind it tests.
func TestUntaggedUnionCastTestsItsTag(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, untaggedTagModel+"function f(x: Expression | Statement): number { return (x as CallExpression).argumentCount; } const call: CallExpression = { kind: SyntaxKind.CallExpression, argumentCount: 2 }; console.log(String(f(call)));", ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "2\n" {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want 2, as Node prints", code, stdout, stderr)
	}
}

// A lie of another kind fails at the cast, with row 9's message (docs/rulings.md:9): a Block holds the
// argumentCount the view reads, as a number, so only the tag test tells it isn't a CallExpression, and native
// stops before printing where Node prints 3. Mutant: untaggedUnionCast returning the view with no CheckedCast
// around it, and native reads the Block's argumentCount through the view and exits 0.
func TestUntaggedUnionCastFailsAnotherKind(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, untaggedTagModel+"interface Block extends Statement { readonly kind: SyntaxKind.Block; readonly argumentCount: number } function f(x: Expression | Statement): number { return (x as CallExpression).argumentCount; } const block: Block = { kind: SyntaxKind.Block, argumentCount: 3 }; console.log(String(f(block)));", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := brandedNative(t, program); code == 0 || stdout != "" || !strings.Contains(stderr, "cast failed: this Expression | Statement is not a CallExpression") {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want a stop at the cast, in row 9's words", code, stdout, stderr)
	}
}

// A member beside the target's ancestors whose kind stays the open SyntaxKind, as tsc's Declaration does beside
// Expression and Statement, leaves the write judgment where it was: every member's kind is the tag the cast
// tests, so the cast lowers. Mutant: the arm judging each member with plain widened, whose enum arm refuses every
// open kind seen as the target's literal.
func TestUntaggedUnionCastTagBesideAnOpenSibling(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, untaggedTagModel+"interface Declaration extends Node { readonly name: string } function f(x: Expression | Statement | Declaration): number { return (x as CallExpression).argumentCount; } const call: CallExpression = { kind: SyntaxKind.CallExpression, argumentCount: 2 }; console.log(String(f(call)));", ".ts"); err != nil {
		t.Fatalf("an open sibling beside the tag's members: %v", err)
	}
}

// A kind a write can reach is no tag (untaggedTagField), so its enum slot keeps the untagged stop, NotYet and
// never Refused: written on every side, and readonly in the members but made writable by the target, where a
// write through the view could store any SyntaxKind under a CallExpression's kind. The every-side row is declined
// by the arm's downcast test first, so it holds the stop's words alone. The target-writable row is held three ways
// at once, so no one mutant reds it: untaggedTagField names no writable tag, so the arm judges it by widened; the
// enum slot widened then meets is a widening; and widenedPast judges a readonly field seen as a writable one a
// widening too. Each of the last two is red-tested alone, by TestUntaggedUnionCastTagBesideAnOpenSibling and by
// TestUntaggedUnionCastsKeepTheirStops' FlowLabel row, and this pin holds them together (J-5wgaj8z).
func TestUntaggedUnionCastWritableKindStops(t *testing.T) {
	t.Parallel()
	use := "function f(x: Expression | Statement): number { return (x as CallExpression).argumentCount; }"
	for _, probe := range []struct{ name, source string }{
		{"a kind written on every side", "enum SyntaxKind { Identifier, CallExpression, Block } interface Node { kind: SyntaxKind } interface Expression extends Node {} interface Statement extends Node {} interface CallExpression extends Expression { kind: SyntaxKind.CallExpression; readonly argumentCount: number } " + use},
		{"a kind the target makes writable", "enum SyntaxKind { Identifier, CallExpression, Block } interface Node { readonly kind: SyntaxKind } interface Expression extends Node {} interface Statement extends Node {} interface CallExpression extends Expression { kind: SyntaxKind.CallExpression; readonly argumentCount: number } " + use},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			var notYet *NotYet
			if errors.As(err, &refused) || !errors.As(err, &notYet) || !strings.Contains(err.Error(), "no literal tag tells apart") {
				t.Fatalf("want NotYet holding %q, never a refusal, got %v", "no literal tag tells apart", err)
			}
		})
	}
}

// A cast that keeps undefined over the same union, `node as CallExpression | undefined`, takes the same road on its
// present value: keepsNullishValue lowers the untagged arm's view, and untaggedUnionCast tests the tag there, so a
// lie of another kind fails at the cast in the present types' words and undefined is itself. Mutant: the tag
// test dropped from untaggedUnionCast, and native reads the Block's argumentCount and prints it.
func TestUntaggedUnionCastKeepingUndefinedTestsItsTag(t *testing.T) {
	t.Parallel()
	use := "interface Block extends Statement { readonly kind: SyntaxKind.Block; readonly argumentCount: number } function f(x: Expression | Statement | undefined): number { const call = x as CallExpression | undefined; return call === undefined ? -1 : call.argumentCount; } "
	program, err := lowerSourceAs(t, untaggedTagModel+use+"const block: Block = { kind: SyntaxKind.Block, argumentCount: 3 }; console.log(String(f(undefined))); console.log(String(f(block)));", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := brandedNative(t, program); code == 0 || stdout != "-1\n" || !strings.Contains(stderr, "is not a CallExpression") {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want -1, then a stop at the cast in row 9's words", code, stdout, stderr)
	}
	honest := lowersAndAgreesWithNodeAs(t, untaggedTagModel+use+"const call: CallExpression = { kind: SyntaxKind.CallExpression, argumentCount: 2 }; console.log(String(f(call))); console.log(String(f(undefined)));", ".ts")
	if code, stdout, stderr := brandedNative(t, honest); code != 0 || stdout != "2\n-1\n" {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want 2 and -1, as Node prints", code, stdout, stderr)
	}
}

// A member that stores the tag in another representation names no tag: B's kind is SyntaxKind | string, held boxed
// as a union, where the test at the cast would read the target's literal as a Number. untaggedTagField asks each
// member's field for the literal's representation, so this union keeps the untagged stop, NotYet (Lower's must on
// 22d8b724). It's held twice: with that check dropped the cast still stops, since B's kind isn't the tag's open enum
// and viewCastEnumSlots keeps the tag's slot out, so widenedPast judges it whole (J-txh7eqf, its native probe). No
// one mutant reds this pin; it holds the two together.
func TestUntaggedUnionCastMixedTagRepresentationStops(t *testing.T) {
	t.Parallel()
	source := "enum SyntaxKind { Identifier, CallExpression, Block } interface A { readonly kind: SyntaxKind; readonly a: number } interface B { readonly kind: SyntaxKind | string; readonly b: number } interface CallLike extends A { readonly kind: SyntaxKind.CallExpression; readonly argumentCount: number } function f(x: A | B): number { return (x as CallLike).argumentCount; }"
	_, err := lowerSourceAs(t, source, ".ts")
	var refused *Refused
	var notYet *NotYet
	if errors.As(err, &refused) || !errors.As(err, &notYet) || !strings.Contains(err.Error(), "no literal tag tells apart") {
		t.Fatalf("want NotYet holding %q, never a refusal or a lowering, got %v", "no literal tag tells apart", err)
	}
}

// A mutable alias no member's type names can flip the tag after the test: a Writable whose kind is the whole
// SyntaxKind, seen as an Expression, passes the tag test as a CallExpression, then a call writes Block into its
// kind. Readonly binds only the view and the members' names, so the test is the cast's answer at the cast, and the
// view's own read check is what keeps the flip loud: native stops at the read of kind where Node prints the lie
// (Language's x01 on #qxt10dd, J-v2k63m2). It guards against a change that trusts the tested tag past the cast and
// reads kind unchecked through the view.
func TestUntaggedUnionCastAliasFlipStopsAtTheRead(t *testing.T) {
	t.Parallel()
	source := untaggedTagModel + "interface Writable { kind: SyntaxKind; argumentCount: number } const w: Writable = { kind: SyntaxKind.CallExpression, argumentCount: 2 }; const e: Expression = w; function flip(): void { w.kind = SyntaxKind.Block; } function f(x: Expression | Statement): string { const call = x as CallExpression; flip(); return String(call.kind) + \" \" + String(call.argumentCount); } console.log(f(e));"
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := brandedNative(t, program); code == 0 || stdout != "" || !strings.Contains(stderr, "field read failed") || !strings.Contains(stderr, "kind") {
		t.Fatalf("sanitized native exit %d, stdout %q, stderr %.300q; want a stop at the read of kind, where Node prints the flipped lie", code, stdout, stderr)
	}
}
