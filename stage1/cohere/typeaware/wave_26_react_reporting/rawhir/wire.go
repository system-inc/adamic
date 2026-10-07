package high_level_intermediate_representation

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// EncodeFunctions projects raw lowering/SSA data. It makes no lint decision.
// SourceFile requests include each function lowered independently; function requests
// retain that function's nested arena and capture edges.
func EncodeFunctions(c *checker.Checker, node *ast.Node, memo bool) (string, error) {
	var functions []*Function
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if ast.IsFunctionLike(n) {
			f := Lower(n, c)
			if f != nil {
				Construct(f)
				if memo && mentionsMemo(n) {
					DropManualMemoization(f)
					if InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(f) > 0 {
						Construct(f)
					}
				}
				functions = append(functions, f)
			}
		}
		if node.Kind == ast.KindSourceFile {
			n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
		}
	}
	visit(node)
	data, err := json.Marshal(project(reflect.ValueOf(functions), c))
	return string(data), err
}

// This is the spelling gate used by Cohere's memo-inclusive cache entry.
func mentionsMemo(node *ast.Node) bool {
	sf := ast.GetSourceFileOfNode(node)
	if sf == nil || node.Pos() < 0 || node.End() > len(sf.Text()) || node.Pos() >= node.End() {
		return true
	}
	text := sf.Text()[node.Pos():node.End()]
	if strings.Contains(text, "useMemo") || strings.Contains(text, "useCallback") {
		return true
	}
	if !strings.Contains(text, "\\") {
		return false
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			if n.Text() == "useMemo" || n.Text() == "useCallback" {
				found = true
				return true
			}
		}
		return n.ForEachChild(visit)
	}
	visit(node)
	return found
}

// Only exported graph data is projected. AST pointers become exact syntax
// coordinates; floats become strings so Infinity/NaN cannot corrupt JSON.
func project(v reflect.Value, c *checker.Checker) any {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		return project(v.Elem(), c)
	}
	if v.CanInterface() {
		switch x := v.Interface().(type) {
		case *ast.Node:
			if x == nil {
				return nil
			}
			text := ""
			switch x.Kind {
			case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral:
				text = x.Text()
			}
			file := ""
			if sf := ast.GetSourceFileOfNode(x); sf != nil {
				file = sf.FileName()
			}
			parentKind := ""
			parentPos, parentEnd := 0, 0
			if x.Parent != nil {
				parentKind = strings.TrimPrefix(x.Parent.Kind.String(), "Kind")
				parentPos = x.Parent.Pos()
				parentEnd = x.Parent.End()
			}
			return map[string]any{"Kind": strings.TrimPrefix(x.Kind.String(), "Kind"), "Pos": x.Pos(), "End": x.End(), "Flags": strconv.FormatUint(uint64(x.Flags), 10), "Text": text, "File": file, "ParentKind": parentKind, "ParentPos": parentPos, "ParentEnd": parentEnd}
		case core.TextRange:
			return map[string]any{"Pos": x.Pos(), "End": x.End()}
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return project(v.Elem(), c)
	case reflect.Slice, reflect.Array:
		out := make([]any, v.Len())
		for i := range out {
			out[i] = project(v.Index(i), c)
		}
		return out
	case reflect.Map:
		out := map[string]any{}
		for _, k := range v.MapKeys() {
			out[fmt.Sprint(k.Interface())] = project(v.MapIndex(k), c)
		}
		return out
	case reflect.Struct:
		fields := map[string]any{}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				fields[v.Type().Field(i).Name] = project(v.Field(i), c)
			}
		}
		if v.Type().Name() == "Identifier" {
			alias, symbol, flags := "", "", "0"
			n, _ := v.FieldByName("Node").Interface().(*ast.Node)
			if n != nil {
				if t := c.GetTypeAtLocation(n); t != nil {
					flags = strconv.FormatUint(uint64(t.Flags()), 10)
					if a := checker.Type_alias(t); a != nil && a.Symbol() != nil {
						alias = a.Symbol().Name
					}
					if s := checker.Type_symbol(t); s != nil {
						symbol = s.Name
					}
				}
			}
			fields["TypeAliasName"] = alias
			fields["TypeSymbolName"] = symbol
			fields["TypeFlags"] = flags
		}
		return map[string]any{"Tag": v.Type().Name(), "Fields": fields}
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	}
	panic("unsupported raw HIR field " + v.Type().String())
}
