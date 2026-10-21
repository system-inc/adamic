// Package parallel is adamiclint's rule that a test runs in parallel unless it says why.
//
// A top-level test that doesn't call t.Parallel() runs alone, and every parallel test in its package waits for it, so
// one serial test stretches the whole package's run; and the shared state that makes a test serial is what the next
// reader needs to know before touching it. cmd/adamic-gate's TestEveryTestIsParallelOrSaysWhy holds the same line,
// but only when that test runs, so a serial test was found on Loom, after the change was queued: 9 reds and 242 held
// change-minutes trace to this family. As a lint rule the author meets it first. The rule reads tests exactly as that
// test does, so the two agree while both run, and it has no baseline: a test the test's baseline allowed is a finding.
//
// A test is what go test runs as one: a function of a _test.go file named Test, or Test followed by anything but a
// lower-case letter, with no receiver, no type parameters and no results, and one named parameter of type
// *testing.T. Benchmarks, fuzz tests and examples have names of their own, and TestMain takes a *testing.M, so none of
// them is a test here. A test is in parallel when its first statement calls that parameter's own Parallel, the method
// of *testing.T, matched by type, so a Parallel on another *testing.T or on another type is no call to it. Otherwise
// its doc comment must say why, in // Not parallel: <reason> with a reason that isn't empty.
package parallel

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// Message names both cures, so the author can act on the finding without looking anything up.
const Message = "a test runs in parallel unless it says why: call t.Parallel() first, or add " +
	"// Not parallel: <the shared state it touches> directly above it"

var Analyzer = &analysis.Analyzer{
	Name: "parallel",
	Doc:  "flags a top-level test that neither calls t.Parallel() first nor says why in // Not parallel: <reason>",
	Run:  run,
}

const reasonMarker = "Not parallel:"

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// The file's own name, not one a //line directive gives it, since only a _test.go file holds tests.
		if !strings.HasSuffix(pass.Fset.PositionFor(file.Package, false).Filename, "_test.go") {
			continue
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			parameter := testParameter(pass, function)
			if parameter == nil || callsParallelFirst(pass, function, parameter) || saysWhy(function) {
				continue
			}
			pass.Reportf(function.Name.Pos(), "%s", Message)
		}
	}
	return nil, nil
}

// testParameter is the *testing.T parameter of a top-level test, or nil when function is no test.
func testParameter(pass *analysis.Pass, function *ast.FuncDecl) *types.Var {
	if function.Recv != nil || function.Body == nil || !isTestName(function.Name.Name) {
		return nil
	}
	// A parameter with no name, as in func TestX(*testing.T), has nothing to call Parallel on, and the gate's test
	// passes such a test over, so the rule does too.
	if len(function.Type.Params.List) != 1 || len(function.Type.Params.List[0].Names) != 1 {
		return nil
	}
	object, ok := pass.TypesInfo.Defs[function.Name].(*types.Func)
	if !ok {
		return nil
	}
	signature, ok := object.Type().(*types.Signature)
	if !ok || signature.TypeParams().Len() > 0 || signature.Results().Len() > 0 || signature.Params().Len() != 1 {
		return nil
	}
	parameter := signature.Params().At(0)
	pointer, ok := types.Unalias(parameter.Type()).(*types.Pointer)
	if !ok || !isTestingT(pointer.Elem()) {
		return nil
	}
	return parameter
}

// isTestName is go test's own test of a name: Test, or Test followed by anything but a lower-case letter, so
// Testify is a helper and Test_parse a test.
func isTestName(name string) bool {
	rest, ok := strings.CutPrefix(name, "Test")
	if !ok {
		return false
	}
	if rest == "" {
		return true
	}
	first, _ := utf8.DecodeRuneInString(rest)
	return !unicode.IsLower(first)
}

// callsParallelFirst says the test's first statement is parameter.Parallel(), spelled with the parameter itself as the
// receiver and no parentheses around it, as the gate's test reads it. A Parallel later in the body, deferred, or in a
// closure leaves the test serial until then, which is the cost the rule is about.
func callsParallelFirst(pass *analysis.Pass, function *ast.FuncDecl, parameter *types.Var) bool {
	if len(function.Body.List) == 0 {
		return false
	}
	statement, ok := function.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := statement.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	if !ok || pass.TypesInfo.Uses[receiver] != parameter {
		return false
	}
	selection, ok := pass.TypesInfo.Selections[selector]
	if !ok || selection.Kind() != types.MethodVal {
		return false
	}
	method, ok := selection.Obj().(*types.Func)
	return ok && isParallelMethod(method)
}

// isParallelMethod says method is testing's (*T).Parallel, by its package and receiver, so a method of the same name
// on any other type is never mistaken for it.
func isParallelMethod(method *types.Func) bool {
	if method.Name() != "Parallel" || method.Pkg() == nil || method.Pkg().Path() != "testing" {
		return false
	}
	signature, ok := method.Type().(*types.Signature)
	if !ok || signature.Recv() == nil {
		return false
	}
	pointer, ok := types.Unalias(signature.Recv().Type()).(*types.Pointer)
	return ok && isTestingT(pointer.Elem())
}

func isTestingT(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	object := named.Obj()
	return object.Pkg() != nil && object.Pkg().Path() == "testing" && object.Name() == "T"
}

// saysWhy finds // Not parallel: <reason> in the test's doc comment the way the gate's test does: anywhere on any line
// of any comment in the group, with the */ that closes a block comment no part of the reason. The gate's test also
// reads the source line just above the test, but a line there that starts with // always ends the doc comment group,
// so the doc comment holds it already. A blank line between the comment and the test leaves the test with no doc
// comment, so a reason there isn't the test's.
func saysWhy(function *ast.FuncDecl) bool {
	if function.Doc == nil {
		return false
	}
	for _, comment := range function.Doc.List {
		for _, line := range strings.Split(comment.Text, "\n") {
			_, reason, found := strings.Cut(line, reasonMarker)
			// A marker with no reason says nothing to the next reader, so it excuses nothing.
			reason = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(reason), "*/"))
			if found && reason != "" {
				return true
			}
		}
	}
	return false
}
