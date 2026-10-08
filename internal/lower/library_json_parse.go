package lower

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) isJSONParse(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return false
	}
	callee := node.AsCallExpression().Expression
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "parse" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON")
}
func (l *lowering) jsonParse(node *ast.Node, target *checker.Type) (ir.Expression, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 1 || hasSpread(node) {
		return nil, l.notYet(node, "JSON.parse revivers or non-single arguments (no reviver occurs in the pinned TypeScript src)")
	}
	text, err := l.expression(args[0])
	if err != nil {
		return nil, err
	}
	if text.Type() != ir.String {
		return nil, l.notYet(args[0], "JSON.parse input requiring ToString coercion")
	}
	// A discarded parse still validates the complete grammar and throws real SyntaxErrors.
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindExpressionStatement {
		return ir.JSONParse{Text: text, Of: ir.Object}, nil
	}
	if target == nil {
		target = l.checker.GetContextualType(node, checker.ContextFlagsNone)
	}
	if target == nil || target.Flags()&checker.TypeFlagsAny != 0 {
		target = nil
	}
	check := &ir.JSONParseSchema{Kind: "raw", Name: "unknown", Of: ir.Union}
	if target != nil {
		check, err = l.jsonParseType(node, target, 0)
		if err != nil {
			return nil, err
		}
	}
	layout := check
	if jsonParseNeedsLayout(check) {
		input := ast.SkipParentheses(args[0])
		if input.Kind != ast.KindStringLiteral && input.Kind != ast.KindNoSubstitutionTemplateLiteral {
			return nil, l.notYet(node, "JSON.parse object or unknown result without a constant document proving its complete key layout (dynamic keys, including NUL, need owned runtime shape metadata)")
		}
		decoder := json.NewDecoder(strings.NewReader(input.Text()))
		decoder.UseNumber()
		layout, err = readJSONLayout(decoder, 0)
		if err != nil {
			return nil, l.notYet(node, "JSON.parse result without a provable constant document layout: "+err.Error())
		}
		if _, err = decoder.Token(); err != io.EOF {
			return nil, l.notYet(node, "JSON.parse result without a single constant document layout")
		}
		layout, err = l.jsonParseStorage(node, check, layout)
		if err != nil {
			return nil, err
		}
	}
	l.jsonParseConstants(check)
	l.jsonParseConstants(layout)
	return ir.JSONParse{Text: text, Check: check, Layout: layout, Of: check.Of}, nil
}
func jsonParseNeedsLayout(s *ir.JSONParseSchema) bool {
	if s.Kind == "object" || s.Kind == "raw" {
		return true
	}
	if s.Element != nil && jsonParseNeedsLayout(s.Element) {
		return true
	}
	for _, m := range s.Members {
		if jsonParseNeedsLayout(m) {
			return true
		}
	}
	return false
}
func (l *lowering) jsonParseType(node *ast.Node, t *checker.Type, depth int) (*ir.JSONParseSchema, error) {
	if t.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return nil, l.notYet(node, "JSON.parse generic boundary schemas need per-instantiation contracts")
	}
	t = l.concrete(t)
	if depth > 64 {
		return nil, l.notYet(node, "JSON.parse recursive or deeper-than-64 declared schemas")
	}
	if held, known := l.representation(t); known && held == ir.Object && len(l.checker.GetIndexInfosOfType(t)) != 0 {
		return nil, l.notYet(node, "JSON.parse index-signature schemas require dynamic key storage")
	}
	of, ok := l.representation(t)
	if t.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
		of, ok = ir.Object, true
	}
	if !ok {
		return nil, l.notYet(node, "JSON.parse declared type "+l.checker.TypeToString(t))
	}
	if of == ir.MaybeBoolean {
		return nil, l.notYet(node, "JSON.parse optional booleans require a two-word slot representation")
	}
	s := &ir.JSONParseSchema{Name: l.checker.TypeToString(t), Of: of}
	flags := t.Flags()
	switch {
	case flags&checker.TypeFlagsAny != 0:
		return nil, l.notYet(node, "JSON.parse declared schema containing any")
	case flags&checker.TypeFlagsUnknown != 0:
		s.Kind = "raw"
	case flags&checker.TypeFlagsNull != 0:
		s.Kind = "null"
	case flags&checker.TypeFlagsUndefined != 0:
		s.Kind = "undefined"
	case flags&checker.TypeFlagsUnion != 0:
		s.Kind = "union"
		for _, m := range t.Types() {
			c, e := l.jsonParseType(node, m, depth+1)
			if e != nil {
				return nil, e
			}
			if c.Kind == "object" || c.Kind == "array" {
				return nil, l.notYet(node, "JSON.parse unions containing containers need recursive variant selection")
			}
			s.Members = append(s.Members, c)
		}
	case of == ir.Number:
		s.Kind = "number"
	case of == ir.Boolean:
		s.Kind = "boolean"
	case of == ir.String:
		s.Kind = "string"
	case of == ir.Array:
		elem := l.checker.GetElementTypeOfArrayType(t)
		if elem == nil {
			return nil, l.notYet(node, "JSON.parse tuples")
		}
		s.Kind = "array"
		var err error
		s.Element, err = l.jsonParseType(node, elem, depth+1)
		if err != nil {
			return nil, err
		}
	case of == ir.Object:
		if isClassInstance(t) || l.isLibraryType(t, "RegExp") || l.isLibraryType(t, "Date") {
			return nil, l.notYet(node, "JSON.parse nominal or non-JSON declared type "+s.Name)
		}
		s.Kind = "object"
		for _, f := range l.checker.GetPropertiesOfType(t) {
			if strings.ContainsRune(f.Name, 0) {
				return nil, l.notYet(node, "JSON.parse NUL property names require length-bearing native shapes")
			}
			c, err := l.jsonParseType(node, l.checker.GetTypeOfSymbol(f), depth+1)
			if err != nil {
				return nil, err
			}
			c.Optional = f.Flags&ast.SymbolFlagsOptional != 0
			s.Fields = append(s.Fields, ir.JSONParseField{Name: f.Name, Schema: c})
		}
	default:
		return nil, l.notYet(node, "JSON.parse declared runtime representation "+s.Name)
	}
	if flags&checker.TypeFlagsStringLiteral != 0 {
		s.Literal = t.AsLiteralType().Value().(string)
		s.HasLiteral = true
	}
	if flags&checker.TypeFlagsNumberLiteral != 0 {
		s.Literal = l.checker.TypeToString(t)
		s.HasLiteral = true
	}
	if flags&checker.TypeFlagsBooleanLiteral != 0 {
		s.Literal = s.Name
		s.HasLiteral = true
	}
	return s, nil
}

// The decoder proves layout only. Runtime grammar, values, UTF-16 strings and
// diagnostics come from the V8 parser port, never from Go's decoded values.
func readJSONLayout(d *json.Decoder, depth int) (*ir.JSONParseSchema, error) {
	if depth > 64 {
		return nil, io.ErrUnexpectedEOF
	}
	token, err := d.Token()
	if err != nil {
		return nil, err
	}
	s := &ir.JSONParseSchema{}
	switch v := token.(type) {
	case nil:
		s.Kind = "null"
		s.Of = ir.Object
	case bool:
		s.Kind = "boolean"
		s.Of = ir.Boolean
	case string:
		s.Kind = "string"
		s.Of = ir.String
	case json.Number:
		s.Kind = "number"
		s.Of = ir.Number
	case json.Delim:
		if v == '{' {
			s.Kind = "object"
			s.Of = ir.Object
			for d.More() {
				key, e := d.Token()
				if e != nil {
					return nil, e
				}
				name := key.(string)
				if strings.ContainsRune(name, 0) || strings.ContainsRune(name, '\ufffd') {
					return nil, fmt.Errorf("NUL or replacement-character keys need an exact native key representation")
				}
				child, e := readJSONLayout(d, depth+1)
				if e != nil {
					return nil, e
				}
				found := false
				for i := range s.Fields {
					if s.Fields[i].Name == name {
						s.Fields[i].Schema = child
						found = true
						break
					}
				}
				if !found {
					s.Fields = append(s.Fields, ir.JSONParseField{Name: name, Schema: child})
				}
			}
		} else if v == '[' {
			s.Kind = "array"
			s.Of = ir.Array
			for d.More() {
				child, e := readJSONLayout(d, depth+1)
				if e != nil {
					return nil, e
				}
				s.Members = append(s.Members, child)
			}
		} else {
			return nil, io.ErrUnexpectedEOF
		}
		if _, err = d.Token(); err != nil {
			return nil, err
		}
	default:
		return nil, io.ErrUnexpectedEOF
	}
	s.Name = s.Kind
	return s, nil
}
func (l *lowering) jsonParseStorage(node *ast.Node, check, layout *ir.JSONParseSchema) (*ir.JSONParseSchema, error) {
	result := *layout
	result.Of = check.Of
	if check.Kind == "raw" {
		result.Of = ir.Union
	}
	if check.Kind == "union" {
		for _, m := range check.Members {
			if m.Kind == layout.Kind {
				r, e := l.jsonParseStorage(node, m, layout)
				if e != nil {
					return nil, e
				}
				r.Of = check.Of
				return r, nil
			}
		}
		return &result, nil // A type-wrong document terminates before materialization.
	}
	if layout.Kind == "object" {
		result.Fields = nil
		for _, f := range layout.Fields {
			contract := &ir.JSONParseSchema{Kind: "raw", Of: ir.Union, Name: "unknown"}
			for _, c := range check.Fields {
				if c.Name == f.Name {
					contract = c.Schema
					break
				}
			}
			child, e := l.jsonParseStorage(node, contract, f.Schema)
			if e != nil {
				return nil, e
			}
			result.Fields = append(result.Fields, ir.JSONParseField{Name: f.Name, Schema: child})
		}
	}
	if layout.Kind == "array" {
		// Different per-index layouts require tuple metadata. Homogeneous scalar arrays
		// and arrays with one uniform concrete object layout have the native array ABI.
		contract := check.Element
		if contract == nil {
			contract = &ir.JSONParseSchema{Kind: "raw", Of: ir.Union, Name: "unknown"}
		}
		result.Element = contract
		result.Members = nil
		for _, m := range layout.Members {
			child, e := l.jsonParseStorage(node, contract, m)
			if e != nil {
				return nil, e
			}
			if result.Element == contract {
				result.Element = child
			} else if !sameJSONLayout(result.Element, child) {
				return nil, l.notYet(node, "JSON.parse heterogeneous array layouts")
			}
		}
	}
	return &result, nil
}
func sameJSONLayout(a, b *ir.JSONParseSchema) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}

func (l *lowering) jsonParseConstants(s *ir.JSONParseSchema) {
	if s == nil {
		return
	}
	l.constant(s.Name)
	if s.HasLiteral {
		l.constant(s.Literal)
	}
	l.jsonParseConstants(s.Element)
	for _, f := range s.Fields {
		l.constant(f.Name)
		l.jsonParseConstants(f.Schema)
	}
	for _, m := range s.Members {
		l.jsonParseConstants(m)
	}
}

// The library declaration's any is never trusted. An inferred parse binding or
// its property reads travel as tagged unknown, and operations still need proof.
func (l *lowering) jsonParseRawOrigin(node *ast.Node, seen map[*ast.Symbol]bool) bool {
	node = ast.SkipParentheses(node)
	if l.isJSONParse(node) {
		return true
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return l.jsonParseRawOrigin(node.AsPropertyAccessExpression().Expression, seen)
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || seen[symbol] {
		return false
	}
	seen[symbol] = true
	for _, d := range symbol.Declarations {
		if d.Kind == ast.KindVariableDeclaration && d.Type() == nil && d.AsVariableDeclaration().Initializer != nil {
			return l.jsonParseRawOrigin(d.AsVariableDeclaration().Initializer, seen)
		}
	}
	return false
}
