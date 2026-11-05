// Read-only call-contract census. This does not invoke Lower or claim native admission.
package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func hasUndefined(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUndefined != 0 {
		return true
	}
	if t.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range t.Types() {
			if hasUndefined(member) {
				return true
			}
		}
	}
	return false
}
func require(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func main() {
	require(len(os.Args) == 2, "usage: call-census AFTER_TREE")
	root, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	paths := []string{}
	err = filepath.WalkDir(filepath.Join(root, "src/compiler"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	sort.Strings(paths)
	require(len(paths) == 79, "want all 79 roots")
	program, err := load.Load(paths)
	if err != nil {
		panic(err)
	}
	var file, coreFile *ast.SourceFile
	for _, f := range program.Files() {
		if program.FileName(f) == filepath.Join(root, "src/compiler/checker.ts") {
			file = f
		}
		if program.FileName(f) == filepath.Join(root, "src/compiler/core.ts") {
			coreFile = f
		}
	}
	require(file != nil && coreFile != nil, "missing source")
	c, release := program.Checker(context.Background(), file)
	defer release()
	owners := []*ast.Node{}
	var find ast.Visitor
	find = func(n *ast.Node) bool {
		if n.Kind == ast.KindFunctionDeclaration && n.Name() != nil && n.Name().Text() == "getDiagnosticsWorker" && n.Body() != nil {
			owners = append(owners, n)
		}
		n.ForEachChild(find)
		return false
	}
	find(file.AsNode())
	require(len(owners) == 1, "owner count drift")
	calls := []*ast.Node{}
	var scan ast.Visitor
	scan = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression {
			call := n.AsCallExpression()
			if ast.IsIdentifier(call.Expression) && call.Expression.Text() == "relativeComplement" {
				calls = append(calls, n)
			}
		}
		n.ForEachChild(scan)
		return false
	}
	scan(owners[0].Body())
	require(len(calls) == 1, "call count drift")
	call := calls[0]
	spelling := strings.TrimSpace(file.Text()[call.Pos():call.End()])
	require(spelling == "relativeComplement(previousGlobalDiagnostics, currentGlobalDiagnostics, compareDiagnostics)", "call text drift")
	declarations := []*ast.Node{}
	var implementation *ast.Node
	for _, n := range coreFile.Statements.Nodes {
		if n.Kind == ast.KindFunctionDeclaration && n.Name() != nil && n.Name().Text() == "relativeComplement" {
			if n.Body() != nil {
				require(implementation == nil, "duplicate implementation")
				implementation = n
			} else {
				declarations = append(declarations, n)
			}
		}
	}
	require(implementation != nil && len(declarations) == 2, "overload inventory drift")
	resolved := c.GetResolvedSignature(call)
	require(resolved != nil && resolved.Declaration() == declarations[0], "defined-array overload not selected")
	promised := c.GetReturnTypeOfSignature(resolved)
	produced := c.GetReturnTypeOfSignature(c.GetSignatureFromDeclaration(implementation))
	require(!hasUndefined(promised) && hasUndefined(produced), "result boundary drift")
	nonNull, assertions := 0, 0
	var body ast.Visitor
	body = func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindNonNullExpression:
			nonNull++
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			assertions++
		}
		n.ForEachChild(body)
		return false
	}
	body(implementation.Body())
	// censusReturnProof scans the whole implementation and returns false for any
	// of these nodes, before its return-path proof. No proof is being simulated.
	require(nonNull+assertions > 0, "main's proof-disabling guard no longer applies")
	result := map[string]any{"measurement": "read-only checker call-contract census, not Lower output", "roots": len(paths), "checkerDiagnostics": 0, "where": strings.TrimPrefix(program.Where(call), root+"/"), "owner": "getDiagnosticsWorker", "callee": "relativeComplement", "call": spelling, "overload": 1, "result": c.TypeToString(promised), "implementationAnnotation": c.TypeToString(produced), "implementationHasUndefined": true, "resultHasUndefined": false, "nonNullAssertionsInBody": nonNull, "castsInBody": assertions, "bodyProof": "main censusReturnProof returns false at its assertion scan", "nativeBoundaryRule": "call-site undefined check if lowering reaches censusOverloadResult; no body-proof erasure", "nativeAdmission": "not claimed; ordinary census stops earlier"}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
