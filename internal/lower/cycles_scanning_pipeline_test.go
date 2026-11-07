package lower

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"
)

// scanningPipelineExempt are the passes the real pipeline calls that the frozen scanning reference
// doesn't, each with why. Each one only judges the lowered program, by reading it: none of them
// rewrites the IR or adds to its string table, so leaving one out can't make the two modes'
// programs differ, only which of two stops a program reports first.
var scanningPipelineExempt = map[string]string{
	"lowering.checkRecordGuardedReads": "judges guarded record reads; reads the IR, rewrites nothing",
	"lowering.checkViewOptionalWrites": "judges writes to optional fields; reads the IR, rewrites nothing",
}

// The frozen scanning reference (scanLowerCheckedWithCycleTypes) and the real pipeline
// (lowerCheckedWithCycleTypes) lower the same programs for TestCycleFinderEquality, which compares
// their IR and string tables, so the reference has to call every pass the real one calls, in the
// same order (Lower, on #qk8rztp S2c-1: the reference missing checkIterableViews left a dispatch's
// panic message in one string table only). This reads both function bodies and lists, in source
// order, every call of a lowering method, of flow's and fresh's functions, and of a package
// function (checkGeneratorRoads, readiness). The set is every call either pipeline makes on the
// lowering or to a pass, since any of them may rewrite the program. The reference's own cycle
// finder (scanFindCycles) stands for findCycles, being the thing compared, and
// scanningPipelineExempt names the judging passes the reference has never called. Mutant: the
// reference's checkIterableViews call taken out, and this goes red.
func TestScanningReferenceCallsEveryPass(test *testing.T) {
	test.Parallel()
	pipeline := pipelineCalls(test, "lower.go", "lowerCheckedWithCycleTypes")
	reference := pipelineCalls(test, "cycles_scanning_reference_test.go", "scanLowerCheckedWithCycleTypes")
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
			test.Errorf(
				"%s is exempt but the real pipeline no longer calls it; take it out of scanningPipelineExempt",
				exempt,
			)
		}
	}
	if !slices.Equal(shared, reference) {
		test.Fatalf(
			"the scanning reference's passes differ from the real pipeline's:\nreal (exemptions taken out): %v\nreference: %v",
			shared,
			reference,
		)
	}
}

// pipelineCalls lists, in source order, the calls a function's body makes on the lowering, to flow
// and fresh, and to package functions other than Go's builtins.
func pipelineCalls(test *testing.T, path, function string) []string {
	test.Helper()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	builtins := map[string]bool{
		"append": true,
		"len":    true,
		"cap":    true,
		"make":   true,
		"new":    true,
		"panic":  true,
		"copy":   true,
		"delete": true,
		"min":    true,
		"max":    true,
	}
	var calls []string
	found := false
	for _, declaration := range file.Decls {
		declared, isFunction := declaration.(*goast.FuncDecl)
		if !isFunction || declared.Name.Name != function || declared.Body == nil {
			continue
		}
		found = true
		goast.Inspect(declared.Body, func(node goast.Node) bool {
			call, isCall := node.(*goast.CallExpr)
			if !isCall {
				return true
			}
			switch callee := call.Fun.(type) {
			case *goast.SelectorExpr:
				if receiver, isIdentifier := callee.X.(*goast.Ident); isIdentifier &&
					(receiver.Name == "lowering" || receiver.Name == "flow" || receiver.Name == "fresh") {
					calls = append(calls, receiver.Name+"."+callee.Sel.Name)
				}
			case *goast.Ident:
				if !builtins[callee.Name] {
					calls = append(calls, callee.Name)
				}
			}
			return true
		})
	}
	if !found {
		test.Fatalf("%s has no function %s", path, function)
	}
	return calls
}
