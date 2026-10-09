// Package testpolicy holds the house rules for cohere's own tests, checked by tests of its own.
package testpolicy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

// NotParallelMarker opens the comment that excuses a test from running in parallel. It must give the
// reason after the colon: "// Not parallel: it sets GOFLAGS with t.Setenv, which a parallel test may not".
const NotParallelMarker = "Not parallel:"

// Violation is one test or subtest that neither calls Parallel nor says why it cannot.
type Violation struct {
	Position token.Position
	Name     string
}

var reasonAfterMarker = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(NotParallelMarker) + `\s*\S`)

// SerialTests is every top-level Test function and every t.Run subtest in one file's source that does not
// call Parallel on its *testing.T and has no comment directly above it starting with NotParallelMarker
// and a reason.
//
// Parallel is the house default (#nxgt2ca): cohere spreads its work over every core by construction, and
// its tests are held to the same. A test that cannot be parallel, because it sets the environment or the
// working directory, swaps package state, or holds memory a sibling would need, says so where a reader
// will see it, and this is what keeps the default from wearing away one serial test at a time.
func SerialTests(fileName string, source []byte) ([]Violation, error) {
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, fileName, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	excused := map[int]bool{}
	for _, group := range file.Comments {
		if reasonAfterMarker.MatchString(group.Text()) {
			excused[files.Position(group.End()).Line+1] = true
		}
	}
	var violations []Violation
	report := func(node ast.Node, name string) {
		position := files.Position(node.Pos())
		if !excused[position.Line] {
			violations = append(violations, Violation{Position: position, Name: name})
		}
	}
	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*ast.FuncDecl)
		if !isFunction || function.Recv != nil || !isTestName(function.Name.Name) {
			continue
		}
		parameter := testingParameter(function.Type)
		if parameter == "" {
			continue
		}
		// A doc comment is the comment above the func line, and it may carry other prose above the marker.
		if !callsParallel(function.Body, parameter) && !(function.Doc != nil && reasonAfterMarker.MatchString(function.Doc.Text())) {
			report(function, function.Name.Name)
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, isCall := node.(*ast.CallExpr)
			if !isCall || len(call.Args) != 2 {
				return true
			}
			selector, isSelector := call.Fun.(*ast.SelectorExpr)
			if !isSelector || selector.Sel.Name != "Run" {
				return true
			}
			literal, isLiteral := call.Args[1].(*ast.FuncLit)
			if !isLiteral {
				return true
			}
			subtest := testingParameter(literal.Type)
			if subtest != "" && !callsParallel(literal.Body, subtest) {
				report(call, function.Name.Name+"/"+subtestName(call.Args[0]))
			}
			return true
		})
	}
	return violations, nil
}

func isTestName(name string) bool {
	if !strings.HasPrefix(name, "Test") || name == "TestMain" {
		return false
	}
	rest := strings.TrimPrefix(name, "Test")
	return rest == "" || !(rest[0] >= 'a' && rest[0] <= 'z')
}

// testingParameter is the name of a function's only parameter when it is a *testing.T, empty otherwise.
func testingParameter(function *ast.FuncType) string {
	if function.Params == nil || len(function.Params.List) != 1 || len(function.Params.List[0].Names) != 1 {
		return ""
	}
	star, isStar := function.Params.List[0].Type.(*ast.StarExpr)
	if !isStar {
		return ""
	}
	selector, isSelector := star.X.(*ast.SelectorExpr)
	if !isSelector || selector.Sel.Name != "T" {
		return ""
	}
	if packageName, isIdentifier := selector.X.(*ast.Ident); !isIdentifier || packageName.Name != "testing" {
		return ""
	}
	return function.Params.List[0].Names[0].Name
}

// callsParallel is whether one of a body's own statements is <parameter>.Parallel(), not one inside a
// closure or a subtest, which would mark some other test.
func callsParallel(body *ast.BlockStmt, parameter string) bool {
	if body == nil {
		return false
	}
	for _, statement := range body.List {
		expression, isExpression := statement.(*ast.ExprStmt)
		if !isExpression {
			continue
		}
		call, isCall := expression.X.(*ast.CallExpr)
		if !isCall || len(call.Args) != 0 {
			continue
		}
		selector, isSelector := call.Fun.(*ast.SelectorExpr)
		if !isSelector || selector.Sel.Name != "Parallel" {
			continue
		}
		if receiver, isIdentifier := selector.X.(*ast.Ident); isIdentifier && receiver.Name == parameter {
			return true
		}
	}
	return false
}

func subtestName(argument ast.Expr) string {
	if literal, isLiteral := argument.(*ast.BasicLit); isLiteral {
		return strings.Trim(literal.Value, "\"`")
	}
	return "(computed name)"
}
