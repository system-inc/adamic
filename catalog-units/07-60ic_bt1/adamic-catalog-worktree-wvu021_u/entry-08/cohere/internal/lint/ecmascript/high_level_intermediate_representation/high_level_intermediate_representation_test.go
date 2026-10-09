package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"sort"
	"testing"
)

// TestEachPlaceCoversEveryInstructionValue is the exhaustiveness check Go's type switch cannot do.
//
// It parses THIS PACKAGE'S OWN SOURCE to find every type implementing InstructionValue, then reads
// the case labels out of EachPlace's type switch and compares the two sets. A variant added without
// a case appears as a missing name and the test names it.
//
// Parsing the source rather than reflecting over registered values is deliberate: a reflection
// approach needs a registry, and a variant whose author forgot the registry is exactly the variant
// whose author forgot the visitor. The source cannot be forgotten to update, because the source is
// where the variant was added.
func TestEachPlaceCoversEveryInstructionValue(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "instructionValue")
	if len(declared) < 40 {
		t.Fatalf("found only %d instruction values, expected the full set; the source scan is broken", len(declared))
	}
	covered := typeSwitchCases(t, "visitor.go", "EachPlace")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("EachPlace has no case for %s, so its places are invisible to every pass", name)
		}
	}
}

// TestEachSuccessorCoversEveryTerminal is the same check for the terminal set.
//
// A terminal with no case yields no successors, which makes every block it leads to unreachable and
// silently deletes them from the graph. That failure looks like a lowering bug rather than a
// missing case, which is why this asserts rather than trusting review.
func TestEachSuccessorCoversEveryTerminal(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "terminal")
	if len(declared) < 18 {
		t.Fatalf("found only %d terminals, expected the full set; the source scan is broken", len(declared))
	}
	covered := typeSwitchCases(t, "visitor.go", "EachSuccessor")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("EachSuccessor has no case for %s, so blocks it leads to look unreachable", name)
		}
	}
}

// TestTerminalOrderCoversEveryTerminal guards evaluation order.
//
// A terminal with no case returns order zero, which reads as "before everything" to any pass
// comparing positions, and is wrong in a way that produces plausible answers.
func TestTerminalOrderCoversEveryTerminal(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "terminal")
	covered := typeSwitchCases(t, "visitor.go", "TerminalOrder")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("TerminalOrder has no case for %s, so it reports order zero", name)
		}
	}
}

// TestSetTerminalOrderCoversEveryTerminal guards the writer half of the same field.
func TestSetTerminalOrderCoversEveryTerminal(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "terminal")
	covered := typeSwitchCases(t, "graph.go", "setTerminalOrder")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("setTerminalOrder has no case for %s, so it never receives an order", name)
		}
	}
}

// TestPrintValueCoversEveryInstructionValue keeps the printer honest.
//
// Every lowering test in this package asserts against printed output, so a variant the printer
// renders as "<unknown>" makes those tests unable to distinguish it from any other unprinted
// variant. The printer is test infrastructure and this is what stops it silently degrading.
func TestPrintValueCoversEveryInstructionValue(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "instructionValue")
	covered := typeSwitchCases(t, "print.go", "printValue")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("printValue has no case for %s, so it prints as <unknown>", name)
		}
	}
}

// TestPrintTerminalCoversEveryTerminal is the same for terminals.
func TestPrintTerminalCoversEveryTerminal(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "terminal")
	covered := typeSwitchCases(t, "print.go", "printTerminal")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("printTerminal has no case for %s, so it prints as <unknown terminal>", name)
		}
	}
}

// TestInstructionSetSizeIsWhatWeSaidItIs anchors the count against the number the design was argued
// from.
//
// The whole case for this shape rests on a measurement: 43 instruction values of which 5 are React,
// and 22 terminals of which 2 are React. If the sets drift without the reasoning being revisited,
// the package comment becomes a claim about a program that no longer exists.
//
// The terminal count is 21 rather than upstream's 22 BY DESIGN: `PrunedScope` is deliberately
// absent. `Scope` was absent for the same reason until `scope_terminals.go` landed to construct it,
// and was added in that same commit -- which is what this guard is for. It went red, was read, and
// the count was moved deliberately rather than by a compiler error. See the terminal.go header.
func TestInstructionSetSizeIsWhatWeSaidItIs(t *testing.T) {
	t.Parallel()

	values := markerImplementers(t, "instructionValue")
	if len(values) != 43 {
		t.Errorf("the instruction set holds %d variants, want 43 to match upstream: %v", len(values), values)
	}

	terminals := markerImplementers(t, "terminal")
	if len(terminals) != 21 {
		t.Errorf("the terminal set holds %d variants, want 21 (upstream's 22 minus PrunedScope): %v",
			len(terminals), terminals)
	}

	// The React-specific instruction values are in the core set on purpose. Asserting they are
	// present stops a later cleanup from removing them without reading why they are here.
	for _, name := range []string{"JsxExpression", "JsxFragment", "JsxText", "StartMemoize", "FinishMemoize"} {
		if !contains(values, name) {
			t.Errorf("%s is missing from the instruction set; see the package comment on why it belongs here", name)
		}
	}
	// `Scope` is present because `scope_terminals.go` constructs it. Asserting it is present stops a
	// later cleanup from removing the variant while its producer still builds one.
	if !contains(terminals, "Scope") {
		t.Error("Scope is missing from the terminal set, but scope_terminals.go constructs one")
	}
	// `PrunedScope` is still deliberately absent. The four passes that construct it upstream are not
	// ported, so adding it would put a variant in the set that nothing here produces -- the same
	// shape as `Optional`, which this package already carries as a warning. A stage that ports one of
	// those passes adds the variant in that commit and updates this line.
	if contains(terminals, "PrunedScope") {
		t.Error("PrunedScope was added to the terminal set but nothing in this tree constructs one; " +
			"see ScopeTerminalsGapPrunedScope")
	}
}

func contains(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}

// markerImplementers returns every type in this package with a method of the given name.
func markerImplementers(t *testing.T, marker string) []string {
	t.Helper()
	var names []string
	forEachPackageFile(t, func(file *ast.File) {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Name.Name != marker {
				continue
			}
			receiver := function.Recv.List[0].Type
			star, ok := receiver.(*ast.StarExpr)
			if !ok {
				continue
			}
			identifier, ok := star.X.(*ast.Ident)
			if !ok {
				continue
			}
			names = append(names, identifier.Name)
		}
	})
	sort.Strings(names)
	return names
}

// typeSwitchCases returns the type names appearing as case labels in the named function's type
// switches.
func typeSwitchCases(t *testing.T, fileName, functionName string) map[string]bool {
	t.Helper()
	cases := map[string]bool{}
	found := false

	// fileName is documentation of where the function lives rather than a filter: the function name
	// is unique across the package, so searching every file finds it and a move does not break this.
	_ = fileName

	forEachPackageFile(t, func(file *ast.File) {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Name.Name != functionName {
				continue
			}
			found = true
			ast.Inspect(function, func(node ast.Node) bool {
				clause, ok := node.(*ast.CaseClause)
				if !ok {
					return true
				}
				for _, expression := range clause.List {
					star, ok := expression.(*ast.StarExpr)
					if !ok {
						continue
					}
					if identifier, ok := star.X.(*ast.Ident); ok {
						cases[identifier.Name] = true
					}
				}
				return true
			})
		}
	})

	if !found {
		t.Fatalf("no function named %s was found; the source scan is looking in the wrong place", functionName)
	}
	return cases
}

// forEachPackageFile parses this package's own source.
//
// Located by the working directory, which `go test` sets to the package's directory under any flags.
// It was runtime.Caller, whose path is the import path under -trimpath, so the scan opened a directory
// that does not exist and every test built on it failed on a machine that builds with -trimpath.
func forEachPackageFile(t *testing.T, visit func(*ast.File)) {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot locate the package source: %v", err)
	}

	fileSet := token.NewFileSet()
	packages, err := goparser.ParseDir(fileSet, directory, nil, 0)
	if err != nil {
		t.Fatalf("parsing the package source: %v", err)
	}
	for _, pkg := range packages {
		for _, file := range pkg.Files {
			visit(file)
		}
	}
}

// TestPlaceStringIsStableAcrossIdenticalValues is the property the whole IR exists to provide.
func TestPlaceStringIsStableAcrossIdenticalValues(t *testing.T) {
	t.Parallel()

	function := NewFunction(nil, "test", FunctionKindOther)
	identifier := function.NewIdentifier("x", nil, 0)
	first := Place{Identifier: identifier.Id}
	second := Place{Identifier: identifier.Id, Effect: EffectMutate}

	if function.PlaceString(first) != function.PlaceString(second) {
		t.Error("two places naming one identifier must print alike; the effect is a property of the reference, not the value")
	}
	if first.Identifier != second.Identifier {
		t.Error("places naming one value must share an identifier")
	}
}

// TestEffectLatticeOrdering pins the order a join depends on.
func TestEffectLatticeOrdering(t *testing.T) {
	t.Parallel()

	mutable := []Effect{
		EffectCapture, EffectStore, EffectConditionallyMutate,
		EffectConditionallyMutateIterator, EffectMutate,
	}
	for _, effect := range mutable {
		if !effect.IsMutable() {
			t.Errorf("%s must count as mutable", effect)
		}
	}
	for _, effect := range []Effect{EffectUnknown, EffectFreeze, EffectRead} {
		if effect.IsMutable() {
			t.Errorf("%s must not count as mutable", effect)
		}
	}
}

func TestClassifyFunction(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want FunctionKind
	}{
		{"Component", FunctionKindComponent},
		{"useState", FunctionKindHook},
		// React's own convention is `use` followed by a capital OR a digit, so `use2` is a hook.
		// Asserting the digit case explicitly because it is the half a reimplementation forgets.
		{"use2", FunctionKindHook},
		{"useSomething", FunctionKindHook},
		{"user", FunctionKindOther},
		{"helper", FunctionKindOther},
		{"", FunctionKindOther},
	}
	for _, test := range cases {
		if got := classifyFunction(test.name); got != test.want {
			t.Errorf("classifyFunction(%q) = %s, want %s", test.name, got, test.want)
		}
	}
}

// TestEachBlockReferencePointerCoversEveryTerminal guards the rewriting walker.
//
// A terminal with no case keeps whatever block ids it held, which is invisible to a caller: the
// rewrite reports success, the graph still builds, and the terminal points at a block from a
// different id space. That is the same failure mode `EachSuccessor` is guarded against, one step
// worse, because a missed edge reads as unreachable while a missed rewrite reads as reachable and
// wrong.
func TestEachBlockReferencePointerCoversEveryTerminal(t *testing.T) {
	t.Parallel()

	declared := markerImplementers(t, "terminal")
	if len(declared) < 18 {
		t.Fatalf("found only %d terminals, expected the full set; the source scan is broken",
			len(declared))
	}
	covered := typeSwitchCases(t, "visitor_mutate.go", "EachBlockReferencePointer")

	for _, name := range declared {
		if !covered[name] {
			t.Errorf("EachBlockReferencePointer has no case for %s, so a pass renaming blocks "+
				"leaves it pointing into the old id space", name)
		}
	}
}

// TestEachBlockReferencePointerSeesEveryReadOnlyReference is the cross-check against its siblings.
//
// The arm list above is guarded against the declared terminals; this is guarded against what the
// read-only walkers actually yield, which is the stronger question. A terminal can have a case here
// and still miss one of its own fields.
func TestEachBlockReferencePointerSeesEveryReadOnlyReference(t *testing.T) {
	t.Parallel()

	terminals := everyTerminalSample(t)
	if len(terminals) < 18 {
		t.Fatalf("only %d terminal samples; the comparison below would not cover the set",
			len(terminals))
	}
	for _, terminal := range terminals {
		expected := map[static_single_assignment.BlockId]bool{}
		EachSuccessorAndFallthrough(terminal, func(block static_single_assignment.BlockId) { expected[block] = true })

		seen := map[static_single_assignment.BlockId]bool{}
		EachBlockReferencePointer(terminal, func(block *static_single_assignment.BlockId) { seen[*block] = true })

		for block := range expected {
			if !seen[block] {
				t.Errorf("%T names block %d through the read-only walkers and "+
					"EachBlockReferencePointer does not reach it", terminal, block)
			}
		}
	}
}

// everyTerminalSample builds one of each terminal with a distinct block id in every slot.
//
// Distinct ids are the whole point: a walker that reads `t.Consequent` twice instead of reading
// `t.Alternate` passes any sample where the two are equal.
func everyTerminalSample(t *testing.T) []Terminal {
	t.Helper()
	next := static_single_assignment.BlockId(0)
	block := func() static_single_assignment.BlockId {
		next++
		return next
	}
	return []Terminal{
		&Return{}, &Throw{}, &Unreachable{}, &Unsupported{},
		&Goto{Block: block()},
		&If{Consequent: block(), Alternate: block(), Fallthrough: block()},
		&Branch{Consequent: block(), Alternate: block(), Fallthrough: block()},
		&Switch{Cases: []SwitchCase{{Block: block()}, {Block: block()}}, Fallthrough: block()},
		&While{Test: block(), Loop: block(), Fallthrough: block()},
		&DoWhile{Loop: block(), Test: block(), Fallthrough: block()},
		&For{Init: block(), Test: block(), Loop: block(), Update: block(), Fallthrough: block()},
		&ForOf{Init: block(), Test: block(), Loop: block(), Fallthrough: block()},
		&ForIn{Init: block(), Loop: block(), Fallthrough: block()},
		&Logical{Test: block(), Fallthrough: block()},
		&Ternary{Test: block(), Fallthrough: block()},
		&Optional{Test: block(), Fallthrough: block()},
		&Sequence{Block: block(), Fallthrough: block()},
		&Label{Block: block(), Fallthrough: block()},
		&Try{Block: block(), Handler: block(), Fallthrough: block()},
		&MaybeThrow{Continuation: block(), Handler: block()},
		&Scope{Block: block(), Fallthrough: block()},
	}
}
