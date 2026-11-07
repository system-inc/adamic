package main

// This read-only trace follows optionalWidened from the census's f1c9173 implementation.
// It records the nested path which the saved census omitted. No lowering is invoked.
import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func trace(
	c *checker.Checker,
	source, target *checker.Type,
	path string,
	visited map[[2]*checker.Type]bool,
) map[string]any {
	if source == nil ||
		target == nil ||
		!c.IsTypeAssignableTo(source, target) ||
		source == target ||
		visited[[2]*checker.Type{source, target}] {
		return nil
	}
	visited[[2]*checker.Type{source, target}] = true
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for index, memberType := range source.Types() {
			if r := trace(c, memberType, target, fmt.Sprintf("%s/source-union[%d]", path, index), visited); r != nil {
				return r
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for index, memberType := range target.Types() {
			if r := trace(c, source, memberType, fmt.Sprintf("%s/target-union[%d]", path, index), visited); r != nil {
				return r
			}
		}
		return nil
	}
	if source.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return trace(c, c.GetBaseConstraintOfType(source), target, path+"/source-constraint", visited)
	}
	if target.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return trace(c, source, c.GetBaseConstraintOfType(target), path+"/target-constraint", visited)
	}
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 ||
		target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	if (c.IsArrayType(source) || checker.IsTupleType(source)) &&
		(c.IsArrayType(target) || checker.IsTupleType(target)) {
		a, b := c.GetTypeArguments(source), c.GetTypeArguments(target)
		for i, subjectType := range a {
			j := i
			if !checker.IsTupleType(target) {
				j = 0
			}
			if j < len(b) {
				if r := trace(c, subjectType, b[j], fmt.Sprintf("%s/element[%d]", path, i), visited); r != nil {
					return r
				}
			}
		}
		return nil
	}
	for _, symbol := range c.GetPropertiesOfType(target) {
		declared := c.GetPropertyOfType(source, symbol.Name)
		if declared == nil {
			if symbol.Flags&ast.SymbolFlagsOptional != 0 {
				return map[string]any{
					"source":                     c.TypeToString(source),
					"target":                     c.TypeToString(target),
					"flags":                      uint32(source.Flags()),
					"property":                   symbol.Name,
					"path":                       path,
					"never_flag":                 source.Flags()&checker.TypeFlagsNever != 0,
					"shim_flags_match":           checker.Type_flags(source) == source.Flags(),
					"shim_object_flags_match":    checker.Type_objectFlags(source) == source.ObjectFlags(),
					"reduced_never_intersection": source.ObjectFlags()&checker.ObjectFlagsIsNeverIntersection != 0,
				}
			}
			continue
		}
		if r := trace(
			c,
			c.GetTypeOfSymbol(declared),
			c.GetTypeOfSymbol(symbol),
			path+"/property:"+symbol.Name,
			visited,
		); r != nil {
			return r
		}
	}
	a, b := c.GetSignaturesOfType(
		source,
		checker.SignatureKindCall,
	), c.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(a) == 1 && len(b) == 1 && len(a[0].TypeParameters()) == 0 && len(b[0].TypeParameters()) == 0 {
		if r := trace(
			c,
			c.GetReturnTypeOfSignature(a[0]),
			c.GetReturnTypeOfSignature(b[0]),
			path+"/return",
			visited,
		); r != nil {
			return r
		}
		x, y := a[0].Parameters(), b[0].Parameters()
		for index := 0; index < len(x) && index < len(y); index++ {
			if r := trace(
				c,
				c.GetTypeOfSymbol(y[index]),
				c.GetTypeOfSymbol(x[index]),
				fmt.Sprintf("%s/parameter[%d]-contravariant", path, index),
				visited,
			); r != nil {
				return r
			}
		}
	}
	return nil
}
func targetAt(c *checker.Checker, n *ast.Node) *checker.Type {
	if n.Kind == ast.KindAsExpression {
		return c.GetTypeAtLocation(n)
	}
	if subjectType := c.GetContextualType(n, checker.ContextFlagsNone); subjectType != nil {
		return subjectType
	}
	child := n
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	p := child.Parent
	if p == nil {
		return nil
	}
	if p.Kind == ast.KindConditionalExpression || p.Kind == ast.KindBinaryExpression {
		return c.GetTypeAtLocation(p)
	}
	if p.Kind == ast.KindReturnStatement || p.Kind == ast.KindArrowFunction {
		for f := p; f != nil; f = f.Parent {
			if ast.IsFunctionLike(f) {
				return c.GetReturnTypeOfSignature(c.GetSignatureFromDeclaration(f))
			}
		}
	}
	return nil
}
func valueTrace(c *checker.Checker, node *ast.Node, target *checker.Type, path string) map[string]any {
	if node.Kind == ast.KindAsExpression {
		node = node.AsAsExpression().Expression
	}
	if node.Kind == ast.KindConditionalExpression {
		expression := node.AsConditionalExpression()
		for index, node := range []*ast.Node{expression.WhenTrue, expression.WhenFalse} {
			if r := valueTrace(c, node, target, fmt.Sprintf("%s/branch[%d]", path, index)); r != nil {
				return r
			}
		}
		return nil
	}
	return trace(c, c.GetTypeAtLocation(node), target, path, map[[2]*checker.Type]bool{})
}
