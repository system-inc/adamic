package lower

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"
)

// scanningPipelineExempt are the passes the real pipeline calls that the frozen scanning reference
// doesn't, each with why. Each one only judges the lowered program, by reading it: none of them
// rewrites the IR or adds to its string table, so leaving one out can't make the two modes'
// programs differ, only which of two stops a program reports first.
var scanningPipelineExempt = map[string]string{
	"lowering.checkOverloadResults":            "judges that every call to a narrow overload was settled; reads lowering's notes, rewrites nothing",
	"lowering.checkOverloadVerdictCalls":       "judges that every call to an unserved overload met its gate; reads lowering's notes and the census, rewrites nothing",
	"lowering.checkRecordGuardedReads":         "judges guarded record reads; reads the IR, rewrites nothing",
	"checkFlowLiveness":                        "judges that no function a flow graph took to never run is named after it was built; reads the IR, rewrites nothing",
	"lowering.checkCheckedPredicateParameters": "judges that every predicate parameter admitted for its checked closure was built as one and that no function it can call observes its argument count; reads lowering's notes and the IR, rewrites nothing",
}

// The frozen scanning reference (scanLowerWithElementKey) and the real pipeline
// (lowerWithElementKey) lower the same programs for TestCycleFinderEquality, which compares their
// IR and string tables, so the reference has to call every pass the real one calls, in the same
// order (Lower, on #qk8rztp S2c-1: the reference missing checkIterableViews left a dispatch's panic
// message in one string table only). This reads both function bodies and lists, in source order,
// every call of a lowering method, of flow's and fresh's functions, and of a package function
// (checkGeneratorRoads, readiness). The set is every call either pipeline makes on the lowering or
// to a pass, since any of them may rewrite the program. The reference's own cycle finder
// (scanFindCycles) stands for findCycles, being the thing compared, and scanningPipelineExempt
// names the judging passes the reference has never called. Each pipeline is entered through a
// wrapper that lowers again with every element read checked when the key misses
// (lowerCheckedWithCycleTypes, scanLowerCheckedWithCycleTypes), and the two wrappers make the same
// calls too. Mutants: the reference's checkIterableViews call taken out, or its
// collectViewedElements call, and this goes red.
func TestScanningReferenceCallsEveryPass(t *testing.T) {
	t.Parallel()
	wrapper := pipelineCalls(t, "lower.go", "lowerCheckedWithCycleTypes")
	referenceWrapper := pipelineCalls(t, "cycles_scanning_reference_test.go", "scanLowerCheckedWithCycleTypes")
	for index, call := range referenceWrapper {
		if call == "scanLowerWithElementKey" {
			referenceWrapper[index] = "lowerWithElementKey"
		}
	}
	if !slices.Equal(wrapper, referenceWrapper) {
		t.Fatalf("the scanning reference's wrapper differs from the real pipeline's:\nreal: %v\nreference: %v", wrapper, referenceWrapper)
	}
	pipeline := pipelineCalls(t, "lower.go", "lowerWithElementKey")
	reference := pipelineCalls(t, "cycles_scanning_reference_test.go", "scanLowerWithElementKey")
	for index, call := range reference {
		if call == "lowering.scanFindCycles" {
			reference[index] = "lowering.findCycles"
		}
	}
	var shared []string
	for _, call := range pipeline {
		if _, exempt := scanningPipelineExempt[call]; !exempt {
			shared = append(shared, call)
		}
	}
	for exempt := range scanningPipelineExempt {
		if !slices.Contains(pipeline, exempt) {
			t.Errorf("%s is exempt but the real pipeline no longer calls it; take it out of scanningPipelineExempt", exempt)
		}
	}
	if !slices.Equal(shared, reference) {
		t.Fatalf("the scanning reference's passes differ from the real pipeline's:\nreal (exemptions taken out): %v\nreference: %v", shared, reference)
	}
}

// pipelineCalls lists, in source order, the calls a function's body makes on the lowering, to flow
// and fresh, and to package functions other than Go's builtins.
func pipelineCalls(t *testing.T, path, function string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	builtins := map[string]bool{"append": true, "len": true, "cap": true, "make": true, "new": true, "panic": true, "copy": true, "delete": true, "min": true, "max": true}
	var calls []string
	found := false
	for _, declaration := range file.Decls {
		declared, isFunction := declaration.(*goast.FuncDecl)
		if !isFunction || declared.Name.Name != function || declared.Body == nil {
			continue
		}
		found = true
		// named is a callee's name as the list holds it, and whether the list holds it at all.
		named := func(callee goast.Expr) (string, bool) {
			switch callee := callee.(type) {
			case *goast.SelectorExpr:
				if receiver, isIdentifier := callee.X.(*goast.Ident); isIdentifier && (receiver.Name == "lowering" || receiver.Name == "flow" || receiver.Name == "fresh") {
					return receiver.Name + "." + callee.Sel.Name, true
				}
			case *goast.Ident:
				return callee.Name, !builtins[callee.Name]
			}
			return "", false
		}
		var visit func(node goast.Node) bool
		visit = func(node goast.Node) bool {
			call, isCall := node.(*goast.CallExpr)
			if !isCall {
				return true
			}
			// A first-link lowering's steps (first_link_passes.go) are left out, with what they're handed: the
			// reference never lowers in that mode, and off it each does nothing, firstLinkPass handing back the
			// stop it was given, firstLinkHeldStops returning, firstLinkProgramStop answering false. firstLinkRun
			// runs the pass it's handed, once, as the reference calls it, so its run stands for the pass: a method
			// value or a function by its name, a function literal by the calls inside it.
			if selector, isSelector := call.Fun.(*goast.SelectorExpr); isSelector {
				if receiver, isIdentifier := selector.X.(*goast.Ident); isIdentifier && receiver.Name == "lowering" && strings.HasPrefix(selector.Sel.Name, "firstLink") {
					if selector.Sel.Name == "firstLinkRun" && len(call.Args) == 3 {
						if run, isLiteral := call.Args[2].(*goast.FuncLit); isLiteral {
							goast.Inspect(run.Body, visit)
						} else if name, listed := named(call.Args[2]); listed {
							calls = append(calls, name)
						}
					}
					return false
				}
			}
			if name, listed := named(call.Fun); listed {
				calls = append(calls, name)
			}
			return true
		}
		goast.Inspect(declared.Body, visit)
	}
	if !found {
		t.Fatalf("%s has no function %s", path, function)
	}
	return calls
}
