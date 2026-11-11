package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// No source program makes most of the pairs viewPairGuard refuses: every cast lowering turns into a
// view admits only a target assignable to its source (interfaceCast, structuralViewCast, a tag cast's
// members), and the checker holds a type predicate's target assignable to its parameter. Structural
// assignability relates same-named fields too, so a related pair's non-function fields are related.
// These tests lower a real graph program, keep its lowering, and judge views set on it directly.
const graphViewGuardDeclarations = `interface Base { readonly tag: string }
interface Derived extends Base { readonly extra: Link }
interface Left { readonly left: Link }
interface Right { readonly right: Link }
type Bag = Record<string, Left>
interface Wanted { readonly a: Right }
type Holder<T extends Base> = { readonly value: T }
interface MakesDerived { make: () => Derived }
interface MakesBase { make: () => Base }
interface ReadsBase { readonly make: () => Base }
interface HoldsDerived { readonly inner: MakesDerived }
interface HoldsBase { readonly inner: MakesBase }
interface Scalars { readonly count: number; readonly name: string; readonly flag: boolean; readonly missing?: number }
interface Holding { readonly link: Link }
interface Opaque { readonly value: unknown }
type DerivedMakers = Map<string, MakesDerived>
type BaseMakers = Map<string, MakesBase>
type EitherBase = MakesBase | Scalars
type BagMakesDerived = Record<string, MakesDerived>
type BagMakesBase = Record<string, MakesBase>
interface NamedMakes { readonly a?: MakesBase }
class Speaks { speak(): void {} }
interface SpeaksField { speak: () => void }
type BagSpeaks = Record<string, Speaks>
interface NamedSpeaks { readonly a?: SpeaksField }
interface DiscriminantWide { readonly kind: 'a' | 'b'; readonly inner: MakesDerived; readonly other: MakesDerived }
type DiscriminatedBase = { readonly kind: 'a'; readonly inner: MakesBase } | { readonly kind: 'b'; readonly other: MakesBase }
interface PartlyWide { readonly kind: 'a' | 'b'; readonly inner: MakesDerived; readonly z: Link }
type PartlyBase = { readonly kind: 'a'; readonly inner: MakesBase } | { readonly kind: 'a' | 'b'; readonly z: Link }
interface CrossA { readonly kind: 'a'; inner: MakesDerived }
interface CrossB { readonly kind: 'b'; inner: MakesBase }
type Cross = CrossA | CrossB
interface HasInner { readonly tag: string; inner: MakesDerived }
interface NoInner { readonly tag: string }
type EitherInner = HasInner | NoInner
interface ViewsInner { readonly tag: string; readonly inner?: MakesBase }
interface ChoiceA { readonly kind: 'a'; link: Link }
interface ChoiceB { readonly kind: 'b'; other: Link }
type Choice = ChoiceA | ChoiceB
interface HoldsChoice { value: Choice }
interface AlsoChoice { value: Choice }
`

// graphViewGuardLowering lowers graphViewGuardDeclarations into a graph program and keeps the
// lowering, its checker alive until the test ends, with each declared interface's, type alias's and
// type parameter's type and declaration by name.
func graphViewGuardLowering(t *testing.T) (*lowering, map[string]*checker.Type, map[string]*ast.Node) {
	t.Helper()
	loaded, err := load.Load([]string{writeAgreementSource(t, "main.ts", graphViewGuardDeclarations+graphReviewCycle+"\n")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	lowered, err := lowerChecked(loaded, check, entry, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	// The guard runs only where graph types are found, so the views below are judged in a graph program.
	if len(lowered.result.GraphTypes) == 0 {
		t.Fatal("the declarations' program found no graph types")
	}
	types, nodes := map[string]*checker.Type{}, map[string]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindTypeParameter, ast.KindClassDeclaration:
			types[node.Name().Text()] = check.GetTypeAtLocation(node.Name())
			nodes[node.Name().Text()] = node
		}
		return node.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	return lowered, types, nodes
}

// graphViewGuardRow is one view set on the kept lowering: source and target by name, the source
// empty for a view with none, and the NotYet it must give, empty when it passes the guard.
type graphViewGuardRow struct {
	name, source, target, want string
}

// judgeGraphViewGuardRows sets each row's view alone on a kept lowering and runs the guard with a
// new finder, so no row's relation or field caches answer for another's.
func judgeGraphViewGuardRows(t *testing.T, lowered *lowering, types map[string]*checker.Type, nodes map[string]*ast.Node, rows []graphViewGuardRow) {
	t.Helper()
	for _, row := range rows {
		view := viewSource{target: types[row.target], node: nodes[row.target]}
		if row.source != "" {
			view.source = types[row.source]
		}
		if view.target == nil || row.source != "" && view.source == nil {
			t.Fatalf("%s: a type isn't declared", row.name)
		}
		lowered.viewSources = []viewSource{view}
		err := (&cycleFinder{l: lowered, where: map[*checker.Type]*ast.Node{}}).viewPairGuard()
		if row.want == "" {
			if err != nil {
				t.Errorf("%s: got %v, want the view to pass the guard", row.name, err)
			}
			continue
		}
		var notYet *NotYet
		if !errors.As(err, &notYet) || notYet.What != row.want {
			t.Errorf("%s: got %v, want NotYet %q", row.name, err, row.want)
		}
	}
}

// A view between two object types neither assignable to the other is NotYet, and so is a dictionary
// seen as an interface whose field it can't hold: the cycle finder relates neither pair.
func TestGraphViewGuardUnrelatedPairIsNotYet(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	bag := lowered.checker.TypeToString(types["Bag"])
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"unrelated objects", "Left", "Right", "a view of Left as Right in a graph program, which the cycle finder doesn't relate yet"},
		{"a dictionary as an interface", "Bag", "Wanted", "a view of " + bag + " as Wanted in a graph program, which the cycle finder doesn't relate yet"},
	})
}

// Related at the top, failing through a same-named field: HoldsDerived as HoldsBase relates, as do
// their inner fields, but inner's writable make takes a function inner's source can't hold.
// Mutant: removing viewPairFailure's recursion into fields turns this red, since the top pair and
// its own fields pass alone.
func TestGraphViewGuardNestedPairIsNotYet(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"through a field", "HoldsDerived", "HoldsBase", "a view of HoldsDerived as HoldsBase in a graph program, whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through its field inner (MakesDerived as MakesBase)"},
	})
}

// A function-typed field is the closure targets' to judge, so a view differing from its source only
// there passes, unless the view can write the field with a function its source's field can't hold.
// The direction matters: a view's field assignable to the source's passes though the reverse isn't.
func TestGraphViewGuardFunctionFields(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"writable, the view's function not the source's", "MakesDerived", "MakesBase", "a view of MakesDerived as MakesBase in a graph program, whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder"},
		{"writable, the view's function the source's", "MakesBase", "MakesDerived", ""},
		{"readonly, the view's function not the source's", "MakesDerived", "ReadsBase", ""},
	})
}

// A declared generic pair is judged as any other: T's constraint Base isn't assignable to Derived, so
// T as Derived is NotYet rather than skipped as a template.
func TestGraphViewGuardDeclaredGenericPairIsNotYet(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"type parameter", "T", "Derived", "a view of T as Derived in a graph program, which the cycle finder doesn't relate yet"},
	})
}

// Views the cycle finder relates pass the guard and still reach graphAdoptionBoundary's refusal,
// unchanged: a source program's Node as Identifier, and the same pair set directly.
func TestGraphViewGuardRelatedPairsPassThrough(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, interfaceCastSource+graphReviewCycle+`console.log(identifier({kind:'identifier',pos:0,name:'ok'} as Identifier).name);`, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a checked-view adapter in a graph program" {
		t.Fatalf("got %v, want graphAdoptionBoundary's refusal", err)
	}
	lowered, types, nodes := graphViewGuardLowering(t)
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"a subtype", "Derived", "Base", ""},
		{"a supertype", "Base", "Derived", ""},
	})
}

// A view with no source passes when its target holds only plain scalars, and is NotYet when its
// target holds a reference field, unknown included. Only a speculation's result makes one from
// source, and its fields are scalars, so these are direct.
func TestGraphViewGuardNoSource(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"scalar target", "", "Scalars", ""},
		{"reference field", "", "Holding", "a view with no source type of Holding in a graph program, which holds a reference field the cycle finder can't follow yet"},
		{"unknown field", "", "Opaque", "a view with no source type of Opaque in a graph program, which holds a reference field the cycle finder can't follow yet"},
	})
}

// The guard's three roads before slice 3 (#3zex15c): a container's elements and a dictionary's index
// values are paired as fields are (Lower's first road; Runtime's index-signature source against a
// named field), each arm of a union target the source relates to (Lower's second), and a writable view
// field over its source's method (Language's third, ruled by Lower on #vqzty8t). Each row passed before
// its road. Mutants, each red on its own rows: viewPairElements returning nil (elements); the index
// loop removed (index values); the held == nil index pairing removed (an index source as a named
// field, both rows); the arm loop removed (an arm); the method or accessor stop removed (a method,
// and the method through an index source); the source arm loop removed (a source arm, Lower's should);
// a split arm skipped, its members never paired (a split arm, Language's must: DiscriminantWide relates
// to DiscriminatedBase by its discriminant and to neither arm alone, and only the arms declare inner);
// a split only where the source arm relates to no arm alone (a split arm beside a related one,
// Language's should: PartlyWide relates whole to PartlyBase's second arm, and kind === 'a' can still
// narrow it to the first).
func TestGraphViewGuardRoads(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	for _, row := range []struct{ name, source, target, want string }{
		// A Map: an array's number index signature would pair its elements through the index road too.
		{"elements", "DerivedMakers", "BaseMakers", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through its elements (MakesDerived as MakesBase)"},
		{"index values", "BagMakesDerived", "BagMakesBase", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through its index values (MakesDerived as MakesBase)"},
		{"an index source as a named field", "BagMakesDerived", "NamedMakes", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through its field a (MakesDerived as "},
		{"an arm", "MakesDerived", "EitherBase", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through an arm (MakesDerived as MakesBase)"},
		{"a method", "Speaks", "SpeaksField", "whose writable field speak is its source's method or accessor, which the cycle finder doesn't walk, so a value written through the view escapes it"},
		{"a source arm", "EitherInner", "ViewsInner", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through a source arm (HasInner as ViewsInner) through its field inner"},
		{"a method through an index source", "BagSpeaks", "NamedSpeaks", "whose writable field speak is its source's method or accessor, which the cycle finder doesn't walk, so a value written through the view escapes it through its field a (Speaks as "},
		{"a split arm", "DiscriminantWide", "DiscriminatedBase", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through a split arm (DiscriminantWide as "},
		{"a split arm beside a related one", "PartlyWide", "PartlyBase", "whose field make takes a function its source's make can't hold, so a closure written through the view escapes the cycle finder through a split arm (PartlyWide as "},
	} {
		if types[row.source] == nil || types[row.target] == nil {
			t.Fatalf("%s: a type isn't declared", row.name)
		}
		lowered.viewSources = []viewSource{{source: types[row.source], target: types[row.target], node: nodes[row.target]}}
		err := (&cycleFinder{l: lowered, where: map[*checker.Type]*ast.Node{}}).viewPairGuard()
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, row.want) {
			t.Errorf("%s: got %v, want NotYet holding %q", row.name, err, row.want)
		}
	}
}

// Correctness's crossing on #3zex15c: at 769b0e53 the arm roads paired whole sides, so a slot of one
// union type on both sides crossed into a pair of disjoint arms no program forms (Choice as Choice
// into ChoiceB as ChoiceA), 289 refusals over tsc. Identity pairs nothing, and arms pair one by one,
// so the view passes. Mutants: the cross product again (a source arm judged against the whole target)
// refuses this; identity replaced by assignability lets MakesDerived as MakesBase through, which
// TestGraphViewGuardFunctionFields reds; the checker's identity in place of one type lets Speaks as
// SpeaksField through, which TestGraphViewGuardRoads's two method rows red (Language's must).
func TestGraphViewGuardIdentityAndArms(t *testing.T) {
	t.Parallel()
	lowered, types, nodes := graphViewGuardLowering(t)
	judgeGraphViewGuardRows(t, lowered, types, nodes, []graphViewGuardRow{
		{"one union type on both sides", "HoldsChoice", "AlsoChoice", ""},
		{"the union itself", "Choice", "Choice", ""},
		// CrossA relates to Cross through its own arm, and its kind 'a' shares no value with CrossB's 'b',
		// so it's never split against CrossB. Mutant: viewPairDisjoint returning false pairs CrossA's inner
		// with CrossB's, MakesDerived as MakesBase, and refuses this.
		{"an arm against a disjoint arm", "CrossA", "Cross", ""},
	})
}

// The pairwise rule leans on two stops (Language's condition on #3zex15c): a source arm reaches a view's
// arm only where it relates to it, and `in` narrowing over a union of object types would break that, so
// it stays refused. The second, a write to a property only an `in` test added, is #yqbg9vz's NotYet,
// pinned there when it lands.
func TestGraphViewGuardLeansOnInRefusal(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `interface A { readonly a: number } interface B { readonly b: number }
function read(value: A | B): number { return 'a' in value ? value.a : value.b; }
console.log(String(read({ a: 1 })));
`, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "in that narrows a union of object types") {
		t.Fatalf("got %v, want the refusal of in over a union of object types", err)
	}
}
