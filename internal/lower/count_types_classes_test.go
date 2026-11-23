package lower

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// countTypeClassesSource has three functions of one type, each called: their three records are identical, so one
// source class, and their three callee types are identical, so one target class (#mh5kdde).
const countTypeClassesSource = `function increment(n: number): number { return n + 1; }
function decrement(n: number): number { return n - 1; }
function doubled(n: number): number { return n * 2; }
console.log(String(increment(1) + decrement(2) + doubled(3)));
`

// countTypeClassesCompareSource has two function types of one call shape, (n: number) => number and
// (s: string) => string, each with two implementations and each passed as a callback, so a row shared across the
// shape rather than across identity gives a callback parameter an implementation that doesn't fit it (#mh5kdde).
const countTypeClassesCompareSource = `function increment(n: number): number { return n + 1; }
function doubled(n: number): number { return n * 2; }
function shout(s: string): string { return s + "!"; }
function ask(s: string): string { return s + "?"; }
function apply(f: (n: number) => number, n: number): number { return f(n); }
function echo(f: (s: string) => string, s: string): string { return f(s); }
console.log(String(apply(increment, 1) + apply(doubled, 2)) + echo(shout, "a") + echo(ask, "b"));
`

// countTypeClassesFreshSource has an object literal, whose expression's type is fresh, and a number literal, whose
// fresh type the checker made when it checked it (#mh5kdde).
const countTypeClassesFreshSource = `const point = { x: 1, y: 2 };
console.log(String(point.x + point.y));
`

// countTypeClassesMethodSource is arguments_length_no_reader.a's View cut down: a function and an interface method
// of one type, (value?: number) => string, the method read through a parameter, and a rest function written into a
// value of the method's type, where its parameters relate bivariantly (#mh5kdde).
const countTypeClassesMethodSource = `function optional(value?: number): string { return String(value ?? -1); }
function rest(...items: number[]): string { return items.join(","); }
interface View { read(value?: number): string; }
function viewed(value: View): string { return value.read(); }
const view: View = { read: rest };
console.log(optional() + viewed(view));
`

// countTypeClassesNestedSource has two functions alike but for what they return, a method's type and a function type
// of the same parameters, and a function returning a rest function (#mh5kdde).
const countTypeClassesNestedSource = `interface Holder { take(value?: number): string; }
const holder: Holder = { take: (value?: number) => String(value) };
function viaMethod(): Holder["take"] { return holder.take; }
function viaFunction(): (value?: number) => string { return holder.take; }
function makesRest(): (...items: number[]) => string { return (...items: number[]) => items.join(","); }
console.log(viaMethod()() + viaFunction()() + makesRest()());
`

// countTypeClassesParameterSource has two functions whose parameters differ only in f: a method in one, a property
// of a function type in the other (#mh5kdde).
const countTypeClassesParameterSource = `function viaMethod(o: { f(x?: number): string }): void { console.log(o.f()); }
function viaProperty(o: { f: (x?: number) => string }): void { console.log(o.f()); }
viaMethod({ f: (x?: number) => String(x) });
viaProperty({ f: (x?: number) => String(x) });
`

// countTypeClassesAbstractSource has two abstract constructor types and a concrete one, each of its own type literal,
// so each its own type (#mh5kdde).
const countTypeClassesAbstractSource = `class Box { readonly size: number = 1; }
const abstractMaker: abstract new () => Box = Box;
const otherAbstractMaker: abstract new () => Box = Box;
const maker: new () => Box = Box;
console.log(String(new maker().size));
`

// countTypeTargets asks the checker once per pair of classes. In countTypeClassesSource increment, decrement and
// doubled are three records of one class, and typeof each, the callees, three targets of one class, so every run
// classes away at least two records and two targets, and its asks are its source classes times its target classes.
// Were the program's records only the three, a run asks once per target class: one source class, so 1 × the target
// classes. Mutant: the classing dropped (one class per type), in countTypeClasses
// `if checker.Checker_isTypeIdenticalTo(typeChecker, representatives[held], proven) {` to `if false {`, and records
// and targets each make as many classes as there are of them.
//
// Not parallel: countTypeClassesTally counts every lowering in the package while it's set, and no parallel test runs
// beside a sequential one.
func TestCountTypeTargetsAskOncePerPairOfClasses(t *testing.T) {
	var tally countTypeClassesCounts
	countTypeClassesTally.Store(&tally)
	defer countTypeClassesTally.Store(nil)
	if _, err := lowerSourceAs(t, countTypeClassesSource, ".ts"); err != nil {
		t.Fatalf("Lower: %v", err)
	}
	runs := tally.runs.Load()
	if runs == 0 {
		t.Fatal("the lowering never resolved its count types, so nothing was tallied")
	}
	if asks, pairs := tally.asks.Load(), tally.pairs.Load(); asks != pairs {
		t.Errorf("countTypeTargets asked the checker %d times, not once per pair of classes (%d)", asks, pairs)
	}
	if records, classes := tally.records.Load(), tally.sourceClasses.Load(); records-classes < 2*runs {
		t.Errorf("%d runs made %d source classes of %d records, not at least two fewer a run: increment, decrement and doubled weren't one class", runs, classes, records)
	}
	if targets, classes := tally.targets.Load(), tally.targetClasses.Load(); targets-classes < 2*runs {
		t.Errorf("%d runs made %d target classes of %d targets, not at least two fewer a run: the three callees weren't one class", runs, classes, targets)
	}
}

// With resolveCountTypesCompare on, every run asks the whole matrix again, record by record, and panics where a
// target's row or a pair's overflow isn't its classes'; countTypeClassesCompareSource gives it none. Mutant: a row
// shared between two non-identical targets, in countTypeClasses both guards gone, `buckets[shape]` to `buckets[""]`
// where it's read and written and `if checker.Checker_isTypeIdenticalTo(typeChecker, representatives[held], proven) {`
// to `if true {`, and (s: string) => string takes (n: number) => number's row, which the comparison panics on.
//
// Not parallel: resolveCountTypesCompare changes every lowering in the package while it's set, and no parallel test
// runs beside a sequential one.
func TestCountTypeTargetsCompareFindsNoSharedRowWrong(t *testing.T) {
	resolveCountTypesCompare.Store(true)
	defer resolveCountTypesCompare.Store(false)
	var tally countTypeClassesCounts
	countTypeClassesTally.Store(&tally)
	defer countTypeClassesTally.Store(nil)
	defer func() {
		if failure := recover(); failure != nil {
			t.Fatalf("the comparison found a row or overflow its classes gave wrong: %v", failure)
		}
	}()
	if _, err := lowerSourceAs(t, countTypeClassesCompareSource, ".ts"); err != nil {
		t.Fatalf("Lower: %v", err)
	}
	if tally.runs.Load() == 0 {
		t.Fatal("the lowering never resolved its count types, so nothing was compared")
	}
}

// A fresh object literal's type and its regular twin are identical, as are a fresh number literal type and its
// regular one, yet assignability reads an object literal's freshness, so countTypeClasses gives each fresh type a
// class of its own: four types, four classes. Two guards keep the object twins apart, countTypeFresh and the fresh
// literal flag among the key's object flags, and the literal twins are keyed by their own ids. Mutant: freshness not
// kept apart, in countTypeClasses `if countTypeFresh(proven) {` to `if false {` and in countTypeShape the object
// flags read past checker.ObjectFlagsFreshLiteral too, and the fresh object shares its regular twin's class, three in
// all.
func TestCountTypeClassesKeepFreshTypesApart(t *testing.T) {
	t.Parallel()
	program, check, _ := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", countTypeClassesFreshSource))
	entry := program.Files()[0]
	objectLiteral := countTypeClassesNode(entry, ast.KindObjectLiteralExpression)
	numberLiteral := countTypeClassesNode(entry, ast.KindNumericLiteral)
	if objectLiteral == nil || numberLiteral == nil {
		t.Fatal("the fixture has no object literal or no number literal")
	}
	freshObject := check.GetTypeAtLocation(objectLiteral)
	if !countTypeFresh(freshObject) {
		t.Fatalf("the object literal's type %s isn't fresh, so nothing here is kept apart", check.TypeToString(freshObject))
	}
	regularObject := regularTypeOfObjectLiteral(check, freshObject)
	// The checker hands an expression its regular literal type; the fresh one is the type it checked the literal as.
	regularNumber := check.GetTypeAtLocation(numberLiteral)
	if regularNumber.Flags()&checker.TypeFlagsFreshable == 0 {
		t.Fatalf("the number literal's type %s isn't a literal type", check.TypeToString(regularNumber))
	}
	freshNumber := regularNumber.AsLiteralType().FreshType()
	if freshNumber == nil || freshNumber == regularNumber || !countTypeFresh(freshNumber) {
		t.Fatalf("the number literal %s has no fresh type apart from its regular one", check.TypeToString(regularNumber))
	}
	for _, twins := range [][2]*checker.Type{{regularObject, freshObject}, {regularNumber, freshNumber}} {
		if countTypeFresh(twins[0]) {
			t.Fatalf("the regular type %s reads as fresh", check.TypeToString(twins[0]))
		}
		// Identical twins are what the classing would merge, so without identity this test proves nothing.
		if !checker.Checker_isTypeIdenticalTo(check, twins[0], twins[1]) {
			t.Fatalf("the checker doesn't hold %s identical to its fresh twin, so the classing would never merge them", check.TypeToString(twins[0]))
		}
	}
	classOf, representatives := countTypeClasses(check, []*checker.Type{regularObject, freshObject, regularNumber, freshNumber})
	if len(representatives) != 4 || classOf[0] == classOf[1] || classOf[2] == classOf[3] {
		t.Errorf("a regular object, its fresh twin, a regular number literal and its fresh twin made %d classes %v, not four apart", len(representatives), classOf)
	}
}

// A method signature's type and a function's of the same parameters are identical, yet a rest function is assignable
// to the method, whose parameters relate bivariantly, and not to the function, whose parameters relate strictly:
// arguments_length_no_reader.a's View.read lost rest from its row where the two were one class (J-89g9jm1). So
// countTypeTargets gives the method's type a row holding rest and the function's type a row without it. Mutant: the
// declaration kind dropped from countTypeShape's key, `key.number(int64(declaration.Kind))` in signatures deleted,
// together with the symbol's flags, `key.number(int64(symbol.Flags))` deleted: the method's type's symbol is a
// method and the function's a function, so either alone keeps these two apart, and with both gone the two targets
// share one row, whichever was met first.
func TestCountTypeTargetsKeepMethodsApartFromFunctions(t *testing.T) {
	t.Parallel()
	program, check, _ := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", countTypeClassesMethodSource))
	entry := program.Files()[0]
	declared := func(name string) *checker.Type {
		node := countTypeClassesFind(entry, func(node *ast.Node) bool {
			return node.Kind == ast.KindFunctionDeclaration && node.Name() != nil && node.Name().Text() == name
		})
		if node == nil {
			t.Fatalf("the fixture declares no function %s", name)
		}
		return check.GetTypeAtLocation(node.Name())
	}
	function, rest := declared("optional"), declared("rest")
	read := countTypeClassesFind(entry, func(node *ast.Node) bool {
		return node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().Name().Text() == "read"
	})
	if read == nil {
		t.Fatal("the fixture reads no method read")
	}
	method := check.GetTypeAtLocation(read)
	// Without identity the classing would never merge them, and without the relation telling them apart a merge
	// would answer rightly, so either way this test would prove nothing.
	if !checker.Checker_isTypeIdenticalTo(check, function, method) {
		t.Fatalf("the checker doesn't hold %s identical to the method %s", check.TypeToString(function), check.TypeToString(method))
	}
	if !check.IsTypeAssignableTo(rest, method) || check.IsTypeAssignableTo(rest, function) {
		t.Fatalf("%s isn't assignable to the method and not to the function, so nothing tells the two apart", check.TypeToString(rest))
	}
	lowering := &lowering{program: program, checker: check, result: &ir.Program{}, this: -1, functionIndex: -1}
	// The functions' indices are the records' own: optional is 0 and rest is 1 (#mh5kdde).
	lowering.closureRecords = []closureRecord{{proven: function, function: 0}, {proven: rest, function: 1}}
	lowering.countTypeTargets(map[int]*checker.Type{int(function.Id()): function, int(method.Id()): method})
	if row := lowering.result.FunctionTypeTargets[int(method.Id())]; !slices.Equal(row, []int{0, 1}) {
		t.Errorf("the method's type %s has row %v, not optional and rest [0 1]", check.TypeToString(method), row)
	}
	if row := lowering.result.FunctionTypeTargets[int(function.Id())]; !slices.Equal(row, []int{0}) {
		t.Errorf("the function's type %s has row %v, not optional alone [0]", check.TypeToString(function), row)
	}
}

// Two function types alike but for a type nested in them, a method's type returned by one and a function type of the
// same parameters by the other, are identical, yet a function returning a rest function is assignable to the first
// and not to the second: the returned types relate as the targets' declarations say, bivariantly into the method and
// strictly into the function. So the key names nested types by id, and the two are two classes. The nesting is a
// return, not a callback parameter, because compareSignaturesRelated relates a callback parameter's signatures in
// callback mode, where strictVariance is off whatever the declaration (relater.go's
// `checkMode&SignatureCheckModeCallback == 0`), so there a method and a function answer alike. Mutant: the return
// type dropped from countTypeShape's key, `key.typeID(typeChecker.GetReturnTypeOfSignature(signature))` in
// signatures deleted, and the two share one class.
func TestCountTypeClassesKeepNestedMethodsApartFromFunctions(t *testing.T) {
	t.Parallel()
	program, check, _ := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", countTypeClassesNestedSource))
	entry := program.Files()[0]
	viaMethod := countTypeClassesDeclared(t, check, entry, ast.KindFunctionDeclaration, "viaMethod")
	viaFunction := countTypeClassesDeclared(t, check, entry, ast.KindFunctionDeclaration, "viaFunction")
	makesRest := countTypeClassesDeclared(t, check, entry, ast.KindFunctionDeclaration, "makesRest")
	// Without identity the classing would never merge them, and without the relation telling them apart a merge
	// would answer rightly, so either way this test would prove nothing.
	if !checker.Checker_isTypeIdenticalTo(check, viaMethod, viaFunction) {
		t.Fatalf("the checker doesn't hold %s identical to %s", check.TypeToString(viaMethod), check.TypeToString(viaFunction))
	}
	if !check.IsTypeAssignableTo(makesRest, viaMethod) || check.IsTypeAssignableTo(makesRest, viaFunction) {
		t.Fatalf("%s isn't assignable to the one returning the method and not to the one returning the function, so nothing tells the two apart", check.TypeToString(makesRest))
	}
	if classOf, representatives := countTypeClasses(check, []*checker.Type{viaMethod, viaFunction}); len(representatives) != 2 {
		t.Errorf("a function returning a method's type and one returning a function type made %d classes %v, not two", len(representatives), classOf)
	}
}

// Two function types whose parameter objects are alike but for one member, a method f in one and a property f of a
// function type in the other, are identical, and the relater reads the member's declaration where it relates the two
// objects, deeper than any key of the outer type's own facts reaches. So the key names each parameter's type by id,
// and the two are two classes. Mutant: the parameter type ids dropped from countTypeShape's key, the loop
// `for _, parameter := range signature.Parameters() { key.typeID(typeChecker.GetTypeOfSymbol(parameter)) }` in
// signatures deleted, and the two share one class.
func TestCountTypeClassesKeepParameterMembersApart(t *testing.T) {
	t.Parallel()
	program, check, _ := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", countTypeClassesParameterSource))
	entry := program.Files()[0]
	viaMethod := countTypeClassesDeclared(t, check, entry, ast.KindFunctionDeclaration, "viaMethod")
	viaProperty := countTypeClassesDeclared(t, check, entry, ast.KindFunctionDeclaration, "viaProperty")
	// Without identity the classing would never merge them, so the test would prove nothing.
	if !checker.Checker_isTypeIdenticalTo(check, viaMethod, viaProperty) {
		t.Fatalf("the checker doesn't call %s identical to %s, so the classing would never merge them and this test proves nothing", check.TypeToString(viaMethod), check.TypeToString(viaProperty))
	}
	if classOf, representatives := countTypeClasses(check, []*checker.Type{viaMethod, viaProperty}); len(representatives) != 2 {
		t.Errorf("a function taking an object with a method f and one taking an object with a function-typed property f made %d classes %v, not two", len(representatives), classOf)
	}
}

// An abstract construct signature and a concrete one of the same parameters and instance type are identical, since
// identity doesn't read the abstract flag, yet an abstract constructor is assignable to the abstract one and not to
// the concrete one (signaturesRelatedTo). So the two are two classes. Mutant: the signature's flags dropped from
// countTypeShape's key, `key.number(int64(signature.Flags()))` in signatures deleted, and the two share one class.
func TestCountTypeClassesKeepAbstractConstructorsApart(t *testing.T) {
	t.Parallel()
	program, check, _ := viewedElementsProgram(t, writeAgreementSource(t, "main.ts", countTypeClassesAbstractSource))
	entry := program.Files()[0]
	abstractMaker := countTypeClassesDeclared(t, check, entry, ast.KindVariableDeclaration, "abstractMaker")
	otherAbstractMaker := countTypeClassesDeclared(t, check, entry, ast.KindVariableDeclaration, "otherAbstractMaker")
	maker := countTypeClassesDeclared(t, check, entry, ast.KindVariableDeclaration, "maker")
	if !checker.Checker_isTypeIdenticalTo(check, abstractMaker, maker) {
		t.Fatalf("the checker doesn't hold %s identical to %s", check.TypeToString(abstractMaker), check.TypeToString(maker))
	}
	if abstractMaker == otherAbstractMaker || !check.IsTypeAssignableTo(otherAbstractMaker, abstractMaker) || check.IsTypeAssignableTo(otherAbstractMaker, maker) {
		t.Fatalf("a second abstract constructor type %s isn't assignable to the first and not to the concrete one, so nothing tells the two apart", check.TypeToString(otherAbstractMaker))
	}
	if classOf, representatives := countTypeClasses(check, []*checker.Type{abstractMaker, maker}); len(representatives) != 2 {
		t.Errorf("an abstract constructor type and a concrete one made %d classes %v, not two", len(representatives), classOf)
	}
}

// countTypeClassesDeclared is the type of the declaration of kind named name in file, as the checker has it at the
// name (#mh5kdde).
func countTypeClassesDeclared(t *testing.T, check *checker.Checker, file *ast.SourceFile, kind ast.Kind, name string) *checker.Type {
	t.Helper()
	node := countTypeClassesFind(file, func(node *ast.Node) bool {
		return node.Kind == kind && node.Name() != nil && node.Name().Kind == ast.KindIdentifier && node.Name().Text() == name
	})
	if node == nil {
		t.Fatalf("the fixture declares no %s", name)
	}
	return check.GetTypeAtLocation(node.Name())
}

// countTypeClassesNode is file's first node of kind, or nil (#mh5kdde).
func countTypeClassesNode(file *ast.SourceFile, kind ast.Kind) *ast.Node {
	return countTypeClassesFind(file, func(node *ast.Node) bool { return node.Kind == kind })
}

// countTypeClassesFind is file's first node matching, or nil (#mh5kdde).
func countTypeClassesFind(file *ast.SourceFile, matching func(*ast.Node) bool) *ast.Node {
	var found *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if matching(node) {
			found = node
			return true
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	return found
}
