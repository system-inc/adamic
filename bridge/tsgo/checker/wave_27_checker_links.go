package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Raw symbol, type and signature links. No syntax traversal or lint verdicts.
func (p *Program) wave27CheckerLinks(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) < 2 {
		return "", fmt.Errorf("checker-links requires an operation")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	switch parts[1] {
	case "symbol":
		if len(parts) != 2 {
			return "", fmt.Errorf("unexpected symbol suffix")
		}
		symbol := c.GetSymbolAtLocation(node)
		out.yes(symbol != nil)
		if symbol != nil {
			out.number(uint64(len(symbol.Declarations)))
			for _, decl := range symbol.Declarations {
				f := ast.GetSourceFileOfNode(decl)
				out.yes(f != nil)
				out.yes(f != nil && f.IsDeclarationFile)
			}
		}
	case "heritage":
		if len(parts) != 2 || (node.Kind != ast.KindClassDeclaration && node.Kind != ast.KindClassExpression) {
			return "", fmt.Errorf("heritage links require a class")
		}
		var clauses *ast.NodeList
		if node.Kind == ast.KindClassDeclaration {
			clauses = node.AsClassDeclaration().HeritageClauses
		} else {
			clauses = node.AsClassExpression().HeritageClauses
		}
		if clauses != nil {
			for _, clause := range clauses.Nodes {
				for _, base := range clause.AsHeritageClause().Types.Nodes {
					roots = append(roots, g.add(c.GetTypeAtLocation(base)))
				}
			}
		}
	case "call":
		if len(parts) != 2 || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
			return "", fmt.Errorf("call links require a call")
		}
		resolved := c.GetResolvedSignature(node)
		out.yes(resolved != nil)
		if resolved != nil {
			declared := resolved.Target()
			out.yes(declared != nil)
			if declared != nil {
				ids := []uint64{}
				for _, t := range declared.TypeParameters() {
					ids = append(ids, g.add(t))
				}
				out.ids(ids)
				for _, sig := range []*checker.Signature{declared, resolved} {
					out.yes(sig.HasRestParameter())
					ids = nil
					for _, param := range sig.Parameters() {
						ids = append(ids, g.add(checker.Checker_getTypeOfSymbol(c, param)))
					}
					out.ids(ids)
					roots = append(roots, g.add(checker.Checker_getReturnTypeOfSignature(c, sig)))
				}
				roots = append(roots, g.add(checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)))
			}
		}
	case "type":
		if len(parts) != 4 {
			return "", fmt.Errorf("type links require identity and property")
		}
		id, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[2] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		t := p.typesByID[id-1]
		// Both property APIs are exposed because location-dependent thenability and a
		// declared container's property type ask distinct compiler questions.
		roots = append(roots, g.add(t), g.add(checker.Checker_getApparentType(c, t)))
		property := checker.Checker_getPropertyOfType(c, t, parts[3])
		var raw, location *checker.Type
		if property != nil {
			raw = checker.Checker_getTypeOfSymbol(c, property)
			location = c.GetTypeOfSymbolAtLocation(property, node)
		}
		roots = append(roots, g.add(raw), g.add(location), g.add(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c))), g.add(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_stringType(c))))
		out.yes(checker.Checker_isArrayOrTupleType(c, t) && !checker.Checker_isArrayType(c, t))
		ids := []uint64{}
		if checker.Checker_isArrayOrTupleType(c, t) {
			for _, arg := range checker.Checker_getTypeArguments(c, t) {
				ids = append(ids, g.add(arg))
			}
		}
		out.ids(ids)
		signatures := checker.Checker_getSignaturesOfType(c, t, checker.SignatureKindCall)
		out.number(uint64(len(signatures)))
		for _, sig := range signatures {
			out.number(g.add(checker.Checker_getReturnTypeOfSignature(c, sig)))
			params := sig.Parameters()
			out.number(uint64(len(params)))
			for _, param := range params {
				out.number(g.add(checker.Checker_getTypeOfSymbol(c, param)))
				out.number(g.add(checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(param, node))))
				decl := param.ValueDeclaration
				out.yes(decl != nil && ast.IsParameterDeclaration(decl) && decl.AsParameterDeclaration().DotDotDotToken != nil)
			}
		}
	default:
		return "", fmt.Errorf("unknown checker-links operation")
	}
	out.ids(roots)
	g.write(out)
	return out.String(), nil
}
