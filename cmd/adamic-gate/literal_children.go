package main

import "go/ast"

// Only an unconditional, direct t.Run consuming the range value proves that
// literal rows are child names. Nested backend loops are not registrations.
func literalRunBinding(parent *ast.FuncDecl, loop *ast.RangeStmt) (string, bool) {
	value, ok := loop.Value.(*ast.Ident)
	if !ok || value.Obj == nil || len(loop.Body.List) != 1 || len(parent.Type.Params.List) == 0 || len(parent.Type.Params.List[0].Names) != 1 {
		return "", false
	}
	statement, ok := loop.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return "", false
	}
	call, ok := statement.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		return "", false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Run" {
		return "", false
	}
	receiver, ok := selector.X.(*ast.Ident)
	if !ok || receiver.Obj != parent.Type.Params.List[0].Names[0].Obj {
		return "", false
	}
	if name, ok := call.Args[0].(*ast.Ident); ok && name.Obj == value.Obj {
		return "", true
	}
	if name, ok := call.Args[0].(*ast.SelectorExpr); ok {
		if row, ok := name.X.(*ast.Ident); ok && row.Obj == value.Obj {
			return name.Sel.Name, true
		}
	}
	return "", false
}

func literalNameIndex(array *ast.ArrayType, field string) (int, bool) {
	if field == "" {
		return 0, true
	}
	structure, ok := array.Elt.(*ast.StructType)
	if !ok {
		return 0, false
	}
	index := 0
	for _, entry := range structure.Fields.List {
		for _, name := range entry.Names {
			if name.Name == field {
				return index, true
			}
			index++
		}
	}
	return 0, false
}
