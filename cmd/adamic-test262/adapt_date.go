package main

import (
	"context"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// A test often declares its Date or scalar result before assigning it. TypeScript
// accepts that evolving any; stage 0 needs a declared representation. An erasable
// annotation includes the initial undefined and every assignment's proven type.
// Failed checking or any uncertain write leaves the source exactly as it was.
func adaptDateLocals(source string) (string, int) {
	if !strings.Contains(source, "Date") {
		return source, 0
	}
	root, err := os.CreateTemp("", "adamic-date-adapt-*.ts")
	if err != nil {
		return source, 0
	}
	path := root.Name()
	defer os.Remove(path)
	if err := root.Close(); err != nil {
		return source, 0
	}
	prefix := prelude + "\n"
	program, err := load.LoadOverlay([]string{path}, map[string]string{path: prefix + rewriteHarnessCalls(source)})
	if err != nil {
		return source, 0
	}
	// Harness renaming changes byte positions. Edits use the checked body,
	// including its free-function harness spelling, and that same body is returned.
	body := rewriteHarnessCalls(source)
	type local struct {
		name      *ast.Node
		kinds     map[string]bool
		uncertain bool
	}
	locals := map[*ast.Symbol]*local{}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	var declare ast.Visitor
	declare = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			d := node.AsVariableDeclaration()
			if d.Initializer == nil && d.Type == nil && ast.IsIdentifier(d.Name()) && node.Pos() >= len(prefix) {
				locals[checked.GetSymbolAtLocation(d.Name())] = &local{name: d.Name(), kinds: map[string]bool{}}
			}
		}
		node.ForEachChild(declare)
		return false
	}
	file.AsNode().ForEachChild(declare)
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		if node.Kind == ast.KindBinaryExpression {
			b := node.AsBinaryExpression()
			left := ast.SkipParentheses(b.Left)
			if ast.IsIdentifier(left) {
				if target := locals[checked.GetSymbolAtLocation(left)]; target != nil {
					if b.OperatorToken.Kind == ast.KindEqualsToken {
						proven := checked.GetTypeAtLocation(b.Right)
						flags := proven.Flags()
						kind := ""
						switch {
						case flags&checker.TypeFlagsNumberLike != 0:
							kind = "number"
						case flags&checker.TypeFlagsStringLike != 0:
							kind = "string"
						case proven.Symbol() != nil && proven.Symbol().Name == "Date" && len(proven.Symbol().Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(proven.Symbol().Declarations[0])):
							kind = "Date"
						default:
							target.uncertain = true
						}
						if kind != "" {
							if len(target.kinds) > 0 && !target.kinds[kind] {
								target.uncertain = true
							}
							target.kinds[kind] = true
						}
					} else if ast.IsAssignmentOperator(b.OperatorToken.Kind) {
						target.uncertain = true
					}
				}
			}
		}
		node.ForEachChild(writes)
		return false
	}
	file.AsNode().ForEachChild(writes)
	var edits []edit
	for _, target := range locals {
		if target.uncertain || len(target.kinds) != 1 {
			continue
		}
		end := target.name.End() - len(prefix)
		kind := ""
		for name := range target.kinds {
			kind = name
		}
		edits = append(edits, edit{start: end, end: end, text: ": " + kind + " | undefined"})
	}
	if len(edits) == 0 {
		return source, 0
	}
	// Return the free harness names deliberately: rewriteHarnessCalls is idempotent.
	rewritten := applyEdits(body, edits)
	if _, err := load.LoadOverlay([]string{path}, map[string]string{path: prefix + rewritten}); err != nil {
		return source, 0
	}
	return rewritten, len(edits)
}
