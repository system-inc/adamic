package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw type links and declaration records for string coercion. No lint verdicts.
func (p *Program) stringificationType(out *fields, c *checker.Checker, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("stringification-type requires an identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	g := &graph{program: p, checker: c}
	out.number(uint64(t.Flags()))
	out.text(strings.ToValidUTF8(c.TypeToString(t), "�"))
	symbol := t.Symbol()
	if alias := checker.Type_alias(t); alias != nil && alias.Symbol() != nil {
		symbol = alias.Symbol()
	}
	genericName := ""
	if symbol != nil && len(symbol.Declarations) > 0 {
		d := symbol.Declarations[0]
		if d.Kind == ast.KindTypeAliasDeclaration || d.Kind == ast.KindInterfaceDeclaration || d.Kind == ast.KindClassDeclaration {
			if params := d.TypeParameterList(); params != nil && len(params.Nodes) > 0 {
				genericName = symbol.Name
			}
		}
	}
	out.text(strings.ToValidUTF8(genericName, "�"))
	out.number(g.id(checker.Checker_getBaseConstraintOfType(c, t)))
	out.yes(checker.Checker_isArrayType(c, t))
	out.yes(checker.IsTupleType(t))
	out.number(g.id(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c))))
	var parts, args, bases []uint64
	if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, v := range t.Types() {
			parts = append(parts, g.id(v))
		}
	}
	if checker.IsTupleType(t) {
		for _, v := range checker.Checker_getTypeArguments(c, t) {
			args = append(args, g.id(v))
		}
	}
	target := t
	if checker.IsNonDeferredTypeReference(t) && t.Target() != nil {
		target = t.Target()
	}
	if target.ObjectFlags()&(checker.ObjectFlagsClass|checker.ObjectFlagsInterface) != 0 {
		for _, v := range checker.Checker_getBaseTypes(c, target) {
			bases = append(bases, g.id(v))
		}
	}
	out.ids(parts)
	out.ids(args)
	out.ids(bases)
	var primitives []*ast.Node
	for _, property := range checker.Checker_getPropertiesOfType(c, t) {
		d := property.ValueDeclaration
		if d != nil && d.Kind == ast.KindMethodSignature {
			primitives = append(primitives, d)
		}
	}
	out.number(uint64(len(primitives)))
	for _, d := range primitives {
		name := d.Name()
		computed := name != nil && name.Kind == ast.KindComputedPropertyName
		out.yes(computed)
		receiver, property := "", ""
		builtin := false
		if computed {
			e := name.AsComputedPropertyName().Expression
			if e != nil && e.Kind == ast.KindPropertyAccessExpression {
				a := e.AsPropertyAccessExpression()
				if a.Expression != nil && a.Expression.Kind == ast.KindIdentifier {
					receiver = a.Expression.Text()
					s := c.GetSymbolAtLocation(a.Expression)
					builtin = false
					if s != nil {
						for _, decl := range s.Declarations {
							file := ast.GetSourceFileOfNode(decl)
							if file != nil && p.Compiler.IsSourceFileDefaultLibrary(file.PathKey()) {
								builtin = true
							}
						}
					}
				}
				if a.Name() != nil && a.Name().Kind == ast.KindIdentifier {
					property = a.Name().Text()
				}
			}
		}
		out.text(receiver)
		out.text(property)
		out.yes(builtin)
	}
	for _, name := range []string{"toLocaleString", "toString", "valueOf"} {
		s := checker.Checker_getPropertyOfType(c, t, name)
		count := 0
		if s != nil {
			count = len(s.Declarations)
		}
		out.number(uint64(count))
		if s != nil {
			for _, d := range s.Declarations {
				kind, parentName := "", ""
				if d.Parent != nil {
					kind = strings.TrimPrefix(d.Parent.Kind.String(), "Kind")
					if d.Parent.Name() != nil {
						parentName = d.Parent.Name().Text()
					}
				}
				out.text(kind)
				out.text(strings.ToValidUTF8(parentName, "�"))
			}
		}
	}
	return out.String(), nil
}
