// Package refusalrewrite instruments only the refusal walkers used by the latent census.
package refusalrewrite

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

const function = "lowering.refuse"

func fail(message string, args ...any) error {
	return fmt.Errorf("latent refusal rewrite: %s: %s", function, fmt.Sprintf(message, args...))
}

func ident(node ast.Node, name string) bool { n, ok := node.(*ast.Ident); return ok && n.Name == name }
func printed(node ast.Node) string {
	var b bytes.Buffer
	_ = format.Node(&b, token.NewFileSet(), node)
	return b.String()
}

// statements parses our fixed instrumentation, never compiler source substitutions.
func statements(text string) []ast.Stmt {
	f, err := parser.ParseFile(token.NewFileSet(), "instrument.go", "package lower\nfunc instrument(){"+text+"}", 0)
	if err != nil {
		panic(err)
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List
}

// Rewrite retains production refuse and appends a separate measurement-only clone.
// Every supported walker must have its own error binding and exactly one child walk.
func Rewrite(source []byte) ([]byte, error) {
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "refusals.go", source, 0)
	if err != nil {
		return nil, fail("cannot parse refusals.go: %v", err)
	}
	var matches []*ast.FuncDecl
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if ok && fn.Name.Name == "refuse" && fn.Recv != nil && len(fn.Recv.List) == 1 && printed(fn.Recv.List[0].Type) == "*lowering" {
			matches = append(matches, fn)
		}
		if ok && fn.Name.Name == "latentRefuse" {
			return nil, fail("latentRefuse already exists")
		}
	}
	if len(matches) != 1 {
		return nil, fail("expected exactly one function, found %d", len(matches))
	}
	original := matches[0]
	if printed(original.Type) != "func(module *ast.SourceFile) error" || len(original.Recv.List[0].Names) != 1 || original.Recv.List[0].Names[0].Name != "l" {
		return nil, fail("unsupported signature")
	}
	var text bytes.Buffer
	if err := format.Node(&text, fs, original); err != nil {
		return nil, fail("cannot clone: %v", err)
	}
	cloneFile, err := parser.ParseFile(fs, "latent-refuse.go", "package lower\n"+text.String(), 0)
	if err != nil {
		return nil, fail("cannot clone: %v", err)
	}
	clone := cloneFile.Decls[0].(*ast.FuncDecl)
	clone.Name.Name = "latentRefuse"
	var collision bool
	ast.Inspect(clone.Body, func(n ast.Node) bool {
		if ident(n, "latentStop") {
			collision = true
		}
		return true
	})
	if collision {
		return nil, fail("reserved instrumentation name latentStop is already used")
	}
	walkers := map[string]*ast.FuncLit{}
	bindings := map[string]bool{}
	for _, statement := range clone.Body.List {
		if declaration, ok := statement.(*ast.DeclStmt); ok {
			if group, ok := declaration.Decl.(*ast.GenDecl); ok && group.Tok == token.VAR {
				for _, spec := range group.Specs {
					v, ok := spec.(*ast.ValueSpec)
					if ok && printed(v.Type) == "error" {
						for _, name := range v.Names {
							bindings[name.Name] = true
						}
					}
				}
			}
		}
		if assignment, ok := statement.(*ast.AssignStmt); ok && len(assignment.Lhs) == 1 && len(assignment.Rhs) == 1 {
			if literal, ok := assignment.Rhs[0].(*ast.FuncLit); ok {
				name, ok := assignment.Lhs[0].(*ast.Ident)
				if !ok || (name.Name != "visit" && name.Name != "contracts") {
					return nil, fail("unexpected visitor assignment %s", printed(assignment.Lhs[0]))
				}
				if walkers[name.Name] != nil {
					return nil, fail("duplicate visitor %s", name.Name)
				}
				walkers[name.Name] = literal
			}
		}
	}
	if walkers["visit"] == nil || !bindings["found"] {
		return nil, fail("missing visit function or found error binding")
	}
	if (walkers["contracts"] != nil) != bindings["contractError"] {
		return nil, fail("contracts function and contractError binding must occur together")
	}
	unexpectedLiteral := false
	ast.Inspect(clone.Body, func(node ast.Node) bool {
		if literal, ok := node.(*ast.FuncLit); ok {
			if literal != walkers["visit"] && literal != walkers["contracts"] {
				unexpectedLiteral = true
			}
			return false
		}
		return true
	})
	if unexpectedLiteral {
		return nil, fail("unexpected function literal outside refusal visitors")
	}
	for name, walker := range walkers {
		binding := "found"
		if name == "contracts" {
			binding = "contractError"
		}
		if printed(walker.Type) != "func(node *ast.Node) bool" {
			return nil, fail("%s has unsupported signature", name)
		}
		// Removing the original recursive walk ensures our defer walks children exactly once.
		count := 0
		ast.Inspect(walker.Body, func(node ast.Node) bool {
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if expr, ok := node.(*ast.ExprStmt); ok {
				if call, ok := expr.X.(*ast.CallExpr); ok && len(call.Args) == 1 && ident(call.Args[0], name) {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok && ident(sel.X, "node") && sel.Sel.Name == "ForEachChild" {
						expr.X = ast.NewIdent("nil")
						count++
					}
				}
			}
			return true
		})
		if count != 1 {
			return nil, fail("%s expected one child walk, found %d", name, count)
		}
		// Replace the temporary marker structurally, including walks inside conditionals.
		eraseWalks(walker.Body)
		walker.Type.Results.List[0].Names = []*ast.Ident{ast.NewIdent("latentStop")}
		prefix := statements(`
if node.Parent != nil && node.Parent.Kind == ast.KindSourceFile { latentFindingOwner = l.program.Where(node) }
if node.Kind == ast.KindFunctionDeclaration && node.Parent != nil && node.Parent.Kind == ast.KindSourceFile && len(l.program.LatentDiagnosticsIn(node.Body())) > 0 { return false }
` + fmt.Sprintf(`defer func(){ latentRecord(%s); %s = nil; node.ForEachChild(%s); latentStop = false }()`, binding, binding, name))
		walker.Body.List = append(prefix, walker.Body.List...)
	}
	// Module metadata returns are outside the visitors. Collect each directive and pragma.
	directives := 0
	for index, statement := range clone.Body.List {
		conditional, ok := statement.(*ast.IfStmt)
		if !ok || printed(conditional.Cond) != "len(module.CommentDirectives) > 0" {
			continue
		}
		if conditional.Init != nil || conditional.Else != nil || len(conditional.Body.List) < 2 {
			return nil, fail("unsupported directive loop")
		}
		first, ok := conditional.Body.List[0].(*ast.AssignStmt)
		if !ok || first.Tok != token.DEFINE || len(first.Lhs) != 1 || !ident(first.Lhs[0], "directive") || len(first.Rhs) != 1 || printed(first.Rhs[0]) != "module.CommentDirectives[0]" {
			return nil, fail("missing directive binding")
		}
		clone.Body.List[index] = &ast.RangeStmt{Key: ast.NewIdent("_"), Value: ast.NewIdent("directive"), Tok: token.DEFINE, X: statements("_ = module.CommentDirectives")[0].(*ast.AssignStmt).Rhs[0], Body: &ast.BlockStmt{List: conditional.Body.List[1:]}}
		directives++
	}
	if directives != 1 {
		return nil, fail("expected one directive scan, found %d", directives)
	}
	outerReturns := 0
	var badReturn bool
	ast.Inspect(clone.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		if len(ret.Results) != 1 {
			badReturn = true
			return false
		}
		value := ret.Results[0]
		allowed := ident(value, "found") || ident(value, "contractError")
		if unary, ok := value.(*ast.UnaryExpr); ok && unary.Op == token.AND {
			if literal, ok := unary.X.(*ast.CompositeLit); ok && ident(literal.Type, "Refused") {
				allowed = true
			}
		}
		if !allowed {
			badReturn = true
		}
		outerReturns++
		return false
	})
	if badReturn || outerReturns < 2 {
		return nil, fail("unsupported outer refusal returns")
	}
	collectReturns(clone.Body)
	clone.Body.List = append(clone.Body.List, &ast.ReturnStmt{Results: []ast.Expr{ast.NewIdent("nil")}})
	file.Decls = append(file.Decls, clone)
	var output bytes.Buffer
	if err := format.Node(&output, fs, file); err != nil {
		return nil, fail("cannot format: %v", err)
	}
	return output.Bytes(), nil
}

func eraseWalks(block *ast.BlockStmt) {
	rewriteBlocks(block, func(statement ast.Stmt) []ast.Stmt {
		if expr, ok := statement.(*ast.ExprStmt); ok && ident(expr.X, "nil") {
			return nil
		}
		return []ast.Stmt{statement}
	})
}
func collectReturns(block *ast.BlockStmt) {
	rewriteBlocks(block, func(statement ast.Stmt) []ast.Stmt {
		if ret, ok := statement.(*ast.ReturnStmt); ok {
			return []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent("latentRecord"), Args: ret.Results}}}
		}
		return []ast.Stmt{statement}
	})
}

// Only statement blocks are visited; nested function literals are separate scopes.
func rewriteBlocks(block *ast.BlockStmt, rewrite func(ast.Stmt) []ast.Stmt) {
	var list []ast.Stmt
	for _, statement := range block.List {
		ast.Inspect(statement, func(node ast.Node) bool {
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if child, ok := node.(*ast.BlockStmt); ok {
				rewriteBlocks(child, rewrite)
				return false
			}
			return true
		})
		list = append(list, rewrite(statement)...)
	}
	block.List = list
}
