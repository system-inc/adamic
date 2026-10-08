package main

// This read-only trace follows optionalWidened from the census's f1c9173 implementation.
// It records the nested path which the saved census omitted. No lowering is invoked.
import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func trace(c *checker.Checker, source, target *checker.Type, path string, visited map[[2]*checker.Type]bool) map[string]any {
	if source == nil || target == nil || !c.IsTypeAssignableTo(source, target) || source == target || visited[[2]*checker.Type{source, target}] {
		return nil
	}
	visited[[2]*checker.Type{source, target}] = true
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for i, m := range source.Types() {
			if r := trace(c, m, target, fmt.Sprintf("%s/source-union[%d]", path, i), visited); r != nil {
				return r
			}
		}
		return nil
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for i, m := range target.Types() {
			if r := trace(c, source, m, fmt.Sprintf("%s/target-union[%d]", path, i), visited); r != nil {
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
	if source.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 || target.Flags()&(checker.TypeFlagsObject|checker.TypeFlagsIntersection) == 0 {
		return nil
	}
	if (c.IsArrayType(source) || checker.IsTupleType(source)) && (c.IsArrayType(target) || checker.IsTupleType(target)) {
		a, b := c.GetTypeArguments(source), c.GetTypeArguments(target)
		for i, m := range a {
			j := i
			if !checker.IsTupleType(target) {
				j = 0
			}
			if j < len(b) {
				if r := trace(c, m, b[j], fmt.Sprintf("%s/element[%d]", path, i), visited); r != nil {
					return r
				}
			}
		}
		return nil
	}
	for _, p := range c.GetPropertiesOfType(target) {
		declared := c.GetPropertyOfType(source, p.Name)
		if declared == nil {
			if p.Flags&ast.SymbolFlagsOptional != 0 {
				return map[string]any{"source": c.TypeToString(source), "target": c.TypeToString(target), "flags": uint32(source.Flags()), "property": p.Name, "path": path, "never_flag": source.Flags()&checker.TypeFlagsNever != 0, "shim_flags_match": checker.Type_flags(source) == source.Flags(), "shim_object_flags_match": checker.Type_objectFlags(source) == source.ObjectFlags(), "reduced_never_intersection": source.ObjectFlags()&checker.ObjectFlagsIsNeverIntersection != 0}
			}
			continue
		}
		if r := trace(c, c.GetTypeOfSymbol(declared), c.GetTypeOfSymbol(p), path+"/property:"+p.Name, visited); r != nil {
			return r
		}
	}
	a, b := c.GetSignaturesOfType(source, checker.SignatureKindCall), c.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(a) == 1 && len(b) == 1 && len(a[0].TypeParameters()) == 0 && len(b[0].TypeParameters()) == 0 {
		if r := trace(c, c.GetReturnTypeOfSignature(a[0]), c.GetReturnTypeOfSignature(b[0]), path+"/return", visited); r != nil {
			return r
		}
		x, y := a[0].Parameters(), b[0].Parameters()
		for i := 0; i < len(x) && i < len(y); i++ {
			if r := trace(c, c.GetTypeOfSymbol(y[i]), c.GetTypeOfSymbol(x[i]), fmt.Sprintf("%s/parameter[%d]-contravariant", path, i), visited); r != nil {
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
	if t := c.GetContextualType(n, checker.ContextFlagsNone); t != nil {
		return t
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
func valueTrace(c *checker.Checker, n *ast.Node, target *checker.Type, path string) map[string]any {
	if n.Kind == ast.KindAsExpression {
		n = n.AsAsExpression().Expression
	}
	if n.Kind == ast.KindConditionalExpression {
		a := n.AsConditionalExpression()
		for i, b := range []*ast.Node{a.WhenTrue, a.WhenFalse} {
			if r := valueTrace(c, b, target, fmt.Sprintf("%s/branch[%d]", path, i)); r != nil {
				return r
			}
		}
		return nil
	}
	return trace(c, c.GetTypeAtLocation(n), target, path, map[[2]*checker.Type]bool{})
}
