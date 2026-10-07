package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// The census owns required-input classification. Boolean opt-in values are
// derived from its skip conditions; path inputs remain setup's responsibility.
func requiredGateVariables(root string) ([]string, error) {
	rows, _, err := censusDeclarations(root)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for _, row := range rows {
		if row.Class != "required-input" {
			continue
		}
		expression, err := parser.ParseExpr(row.Condition)
		if err != nil {
			continue
		}
		comparison, ok := expression.(*ast.BinaryExpr)
		if !ok || comparison.Op != token.NEQ {
			continue
		}
		literal, ok := comparison.Y.(*ast.BasicLit)
		if !ok || literal.Value != `"1"` {
			continue
		}
		call, ok := comparison.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			continue
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Getenv" {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, row.File), nil, 0)
		if err != nil {
			return nil, err
		}
		constants, err := packageStringConstants(filepath.Join(root, filepath.Dir(row.File)), f.Name.Name)
		if err != nil {
			return nil, err
		}
		name, ok := constantString(call.Args[0], constants, map[string]bool{})
		if !ok {
			return nil, fmt.Errorf("dynamic required-input gate in census: %s::%s", row.File, row.Test)
		}
		names[name] = true
	}
	result := []string{}
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

type environmentRequirement struct {
	Shard     int
	Variables []string
	Gates     []string
	Command   string
}

type stringConstant struct {
	expression  ast.Expr
	declaration *ast.ValueSpec
}

// Include ordinary Go files and other test files in the same package. Resolve
// string literals, aliases and concatenations only; mutable names remain dynamic.
func packageStringConstants(directory, packageName string) (map[string]stringConstant, error) {
	constants := map[string]stringConstant{}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), filepath.Join(directory, entry.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		if f.Name.Name != packageName {
			continue
		}
		for _, declaration := range f.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.CONST {
				continue
			}
			var previous []ast.Expr
			for _, specification := range group.Specs {
				spec := specification.(*ast.ValueSpec)
				values := spec.Values
				if len(values) == 0 {
					values = previous
				} else {
					previous = values
				}
				for i, name := range spec.Names {
					if i < len(values) {
						if _, duplicate := constants[name.Name]; duplicate {
							constants[name.Name] = stringConstant{} // Ambiguous build-tag variants fail closed when used.
						} else {
							constants[name.Name] = stringConstant{values[i], spec}
						}
					}
				}
			}
		}
	}
	return constants, nil
}

func constantString(expression ast.Expr, constants map[string]stringConstant, seen map[string]bool) (string, bool) {
	switch e := expression.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			value, err := strconv.Unquote(e.Value)
			return value, err == nil
		}
	case *ast.ParenExpr:
		return constantString(e.X, constants, seen)
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			a, ok := constantString(e.X, constants, seen)
			if !ok {
				return "", false
			}
			b, ok := constantString(e.Y, constants, seen)
			return a + b, ok
		}
	case *ast.Ident:
		definition, ok := constants[e.Name]
		if !ok || definition.expression == nil || seen[e.Name] {
			return "", false
		}
		// Parsing other files creates different ast.Objects. Within the current
		// file, reject local shadows by checking their declaration's source position.
		if e.Obj != nil {
			spec, ok := e.Obj.Decl.(*ast.ValueSpec)
			if e.Obj.Kind != ast.Con || !ok || spec.Pos() != definition.declaration.Pos() {
				return "", false
			}
		}
		seen[e.Name] = true
		value, ok := constantString(definition.expression, constants, seen)
		delete(seen, e.Name)
		return value, ok
	}
	return "", false
}

func skipCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && (selector.Sel.Name == "Skip" || selector.Sel.Name == "Skipf" || selector.Sel.Name == "SkipNow")
}

func environmentReady(p plan, index int) error {
	if p.Environment == nil || p.Environment.Shard != index {
		return nil
	}
	for _, variable := range p.Environment.Variables {
		if os.Getenv(variable) != "1" {
			return fmt.Errorf("required environment shard %d refuses to start: %s=1 required", index, variable)
		}
	}
	return nil
}

func requiredEnvironmentSkips(p plan, index int, results []result) []string {
	var failures []string
	for _, r := range results {
		if r.Action != "skip" || r.Test == "" {
			continue
		}
		for _, u := range p.Units {
			if u.Package != r.Package || len(u.RequiredEnvironment) == 0 {
				continue
			}
			parent, _, _ := strings.Cut(u.Test, "/")
			if r.Test == parent || strings.HasPrefix(r.Test, parent+"/") {
				failures = append(failures, fmt.Sprintf("shard %d required environment unit %s skipped: %s (%s); inputs=%s", index, u.key(), r.key(), r.Reason, strings.Join(u.RequiredEnvironment, ",")))
			}
		}
	}
	sort.Strings(failures)
	return failures
}

func wasiEnvironmentName(value string) bool {
	if !strings.Contains(value, "WASI") {
		return false
	}
	for _, ch := range value {
		if !(ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_') {
			return false
		}
	}
	return true
}
