package skipcensus

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// optInAnnotations follows off guards, sequential continuation, nested callbacks
// and direct test-helper calls. It does not use the worker's environment: the
// same declaration must be correct both with the switch off and with it on.
func optInAnnotations(root string, rows []Row) error {
	set := token.NewFileSet()
	functions := map[string]*ast.FuncDecl{}
	locations := map[string]int{}
	files := map[string]bool{}
	for index, row := range rows {
		files[row.File] = true
		locations[fmt.Sprintf("%s:%d:%d", row.File, row.Line, row.column)] = index
	}
	// Include callers in other test files even when those files have no skip.
	for file := range files {
		matches, err := filepath.Glob(filepath.Join(root, filepath.Dir(file), "*_test.go"))
		if err != nil {
			return err
		}
		for _, path := range matches {
			relative, _ := filepath.Rel(root, path)
			files[filepath.ToSlash(relative)] = true
		}
	}
	for file := range files {
		tree, err := parser.ParseFile(set, filepath.Join(root, file), nil, 0)
		if err != nil {
			return err
		}
		for _, declaration := range tree.Decls {
			if f, ok := declaration.(*ast.FuncDecl); ok && f.Body != nil {
				functions[filepath.Dir(file)+":"+f.Name.Name] = f
			}
		}
	}
	visited := map[string]bool{}
	var walk func(ast.Node, string, map[string]bool, map[string]bool)
	walk = func(node ast.Node, directory string, on, off map[string]bool) {
		if node == nil {
			return
		}
		switch n := node.(type) {
		case *ast.BlockStmt:
			active := copySwitches(on)
			for _, statement := range n.List {
				walk(statement, directory, active, off)
				if branch, ok := statement.(*ast.IfStmt); ok && branch.Else == nil && terminatingSkip(branch.Body) {
					if variable := disabledVariable(branch.Cond); variable != "" {
						active[variable] = true
					}
				}
			}
			return
		case *ast.IfStmt:
			variable := disabledVariable(n.Cond)
			bodyOn, bodyOff := copySwitches(on), copySwitches(off)
			if variable != "" {
				delete(bodyOn, variable)
				bodyOff[variable] = true
			}
			walk(n.Init, directory, on, off)
			walk(n.Body, directory, bodyOn, bodyOff)
			elseOn, elseOff := copySwitches(on), copySwitches(off)
			if variable != "" {
				elseOn[variable] = true
				delete(elseOff, variable)
			}
			walk(n.Else, directory, elseOn, elseOff)
			return
		case *ast.CallExpr:
			if _, ok := n.Fun.(*ast.SelectorExpr); ok {
				position := set.Position(n.Pos())
				file, _ := filepath.Rel(root, position.Filename)
				if index, ok := locations[fmt.Sprintf("%s:%d:%d", filepath.ToSlash(file), position.Line, position.Column)]; ok {
					for variable := range on {
						rows[index].OptInOn = append(rows[index].OptInOn, variable)
					}
					for variable := range off {
						rows[index].OptInOff = append(rows[index].OptInOff, variable)
					}
				}
			}
			if identifier, ok := n.Fun.(*ast.Ident); ok {
				if callee := functions[directory+":"+identifier.Name]; callee != nil {
					context := directory + ":" + identifier.Name + "/" + switchKey(on) + "/" + switchKey(off)
					if !visited[context] {
						visited[context] = true
						walk(callee.Body, directory, on, off)
					}
				}
			}
		}
		ast.Inspect(node, func(child ast.Node) bool {
			if child == nil {
				return false
			}
			if child == node {
				return true
			}
			walk(child, directory, on, off)
			return false
		})
	}
	for name, f := range functions {
		walk(f.Body, strings.Split(name, ":")[0], map[string]bool{}, map[string]bool{})
	}
	for index := range rows {
		if len(rows[index].OptInOn) > 0 {
			rows[index].OptInOn = unique(rows[index].OptInOn)
		}
		if len(rows[index].OptInOff) > 0 {
			rows[index].OptInOff = unique(rows[index].OptInOff)
		}
	}
	return nil
}
func isSkipName(name string) bool { return name == "Skip" || name == "Skipf" || name == "SkipNow" }
func terminatingSkip(body *ast.BlockStmt) bool {
	for _, statement := range body.List {
		if expression, ok := statement.(*ast.ExprStmt); ok {
			if call, ok := expression.X.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok && isSkipName(selector.Sel.Name) {
					return true
				}
			}
		}
	}
	return false
}
func copySwitches(input map[string]bool) map[string]bool {
	output := map[string]bool{}
	for key, value := range input {
		output[key] = value
	}
	return output
}
func switchKey(input map[string]bool) string {
	var values []string
	for value := range input {
		values = append(values, value)
	}
	sort.Strings(values)
	return strings.Join(values, ",")
}

func disabledVariable(expression ast.Expr) string {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok {
		return ""
	}
	environment := environmentVariable(binary.X, map[*ast.Object]bool{})
	value, known := stringValue(binary.Y, map[*ast.Object]bool{})
	if environment == "" {
		environment = environmentVariable(binary.Y, map[*ast.Object]bool{})
		value, known = stringValue(binary.X, map[*ast.Object]bool{})
	}
	if known && environment != "" && (binary.Op == token.NEQ && value == "1" || binary.Op == token.EQL && value == "") {
		return environment
	}
	return ""
}
func declarationValue(identifier *ast.Ident) ast.Expr {
	if identifier.Obj == nil {
		return nil
	}
	switch declaration := identifier.Obj.Decl.(type) {
	case *ast.ValueSpec:
		for index, name := range declaration.Names {
			if name.Name == identifier.Name && index < len(declaration.Values) {
				return declaration.Values[index]
			}
		}
	case *ast.AssignStmt:
		for index, name := range declaration.Lhs {
			if id, ok := name.(*ast.Ident); ok && id.Name == identifier.Name && index < len(declaration.Rhs) {
				return declaration.Rhs[index]
			}
		}
	}
	return nil
}
func environmentVariable(expression ast.Expr, seen map[*ast.Object]bool) string {
	switch node := expression.(type) {
	case *ast.ParenExpr:
		return environmentVariable(node.X, seen)
	case *ast.Ident:
		if node.Obj == nil || seen[node.Obj] {
			return ""
		}
		seen[node.Obj] = true
		return environmentVariable(declarationValue(node), seen)
	case *ast.CallExpr:
		if selector, ok := node.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Getenv" && len(node.Args) == 1 {
			if receiver, ok := selector.X.(*ast.Ident); ok && receiver.Name == "os" {
				value, _ := stringValue(node.Args[0], map[*ast.Object]bool{})
				return value
			}
		}
	}
	return ""
}
func stringValue(expression ast.Expr, seen map[*ast.Object]bool) (string, bool) {
	switch node := expression.(type) {
	case *ast.BasicLit:
		if node.Kind == token.STRING {
			value, err := strconv.Unquote(node.Value)
			return value, err == nil
		}
	case *ast.Ident:
		if node.Obj != nil && !seen[node.Obj] {
			seen[node.Obj] = true
			return stringValue(declarationValue(node), seen)
		}
	}
	return "", false
}
