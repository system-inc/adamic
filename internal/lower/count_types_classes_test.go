package lower

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
// shared between two non-identical targets, in countTypeClasses
// `if checker.Checker_isTypeIdenticalTo(typeChecker, representatives[held], proven) {` to `if true {`, and
// (s: string) => string takes (n: number) => number's row, which the comparison panics on.
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
// class of its own: four types, four classes. Mutant: freshness not kept apart, in countTypeClasses
// `if countTypeFresh(proven) {` to `if false {`, and each fresh type shares its regular twin's class, two in all.
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

// countTypeClassesNode is file's first node of kind, or nil (#mh5kdde).
func countTypeClassesNode(file *ast.SourceFile, kind ast.Kind) *ast.Node {
	var found *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == kind {
			found = node
			return true
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	return found
}
