package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// requireAwaitFacts exposes syntax, signatures and ordinary type operations.
// It never chooses a lint exemption, contextual path, finding or suggestion.
func (p *Program) requireAwaitFacts(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	pieces := strings.SplitN(question, "\n", 5)
	if len(pieces) < 2 || pieces[0] != "require-await" {
		return "", fmt.Errorf("invalid require-await question")
	}
	action := pieces[1]
	out.text(action)
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	present := true
	switch action {
	case "function":
		if len(pieces) != 2 || node.FunctionLikeData() == nil {
			return "", fmt.Errorf("require-await function requires a function-like node")
		}
		out.number(uint64(node.ModifierFlags()))
		out.number(uint64(ast.GetFunctionFlags(node)))
		body := node.Body()
		out.yes(body != nil)
		kind := ""
		statements := 0
		if body != nil {
			kind = strings.TrimPrefix(body.Kind.String(), "Kind")
			if body.Kind == ast.KindBlock && body.AsBlock().Statements != nil {
				statements = len(body.AsBlock().Statements.Nodes)
			}
		}
		out.text(kind)
		out.number(uint64(statements))
		type entry struct {
			node  *ast.Node
			depth uint64
		}
		var nodes []entry
		var walk func(*ast.Node, uint64)
		walk = func(n *ast.Node, depth uint64) {
			nodes = append(nodes, entry{n, depth})
			n.ForEachChild(func(child *ast.Node) bool { walk(child, depth+1); return false })
		}
		if body != nil {
			walk(body, 0)
		}
		out.number(uint64(len(nodes)))
		for _, entry := range nodes {
			out.text(strings.TrimPrefix(entry.node.Kind.String(), "Kind"))
			out.number(entry.depth)
			out.number(uint64(entry.node.Flags))
			out.yes(entry.node.Kind == ast.KindForOfStatement && entry.node.AsForInOrOfStatement().AwaitModifier != nil)
		}
		return out.String(), nil
	case "resolved":
		if len(pieces) != 2 || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
			return "", fmt.Errorf("require-await resolved requires a call or new expression")
		}
		signature := c.GetResolvedSignature(node)
		out.yes(signature != nil)
		target := (*checker.Signature)(nil)
		if signature != nil {
			target = signature.Target()
		}
		out.yes(target != nil)
		var typeParameters []uint64
		if target != nil {
			for _, t := range target.TypeParameters() {
				typeParameters = append(typeParameters, g.add(t))
			}
		}
		out.ids(typeParameters)
		write := func(s *checker.Signature) {
			out.yes(s != nil && s.HasRestParameter())
			var parameters []uint64
			result := uint64(0)
			if s != nil {
				for _, parameter := range s.Parameters() {
					parameters = append(parameters, g.add(checker.Checker_getTypeOfSymbol(c, parameter)))
				}
				result = g.add(c.GetReturnTypeOfSignature(s))
			}
			out.ids(parameters)
			out.number(result)
		}
		write(target)
		write(signature)
	case "heritage":
		if len(pieces) != 2 || !ast.IsClassLike(node) {
			return "", fmt.Errorf("require-await heritage requires a class")
		}
		clauses := node.ClassLikeData().HeritageClauses
		if clauses != nil {
			for _, clause := range clauses.Nodes {
				for _, subject := range clause.AsHeritageClause().Types.Nodes {
					roots = append(roots, g.add(c.GetTypeAtLocation(subject)))
				}
			}
		}
	case "type":
		if len(pieces) < 4 {
			return "", fmt.Errorf("require-await type requires identity and operation")
		}
		id, err := strconv.ParseUint(pieces[2], 10, 64)
		if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != pieces[2] {
			return "", fmt.Errorf("unknown checker type identity")
		}
		subject := p.typesByID[id-1]
		operation := pieces[3]
		switch operation {
		case "signatures":
			if len(pieces) != 4 {
				return "", fmt.Errorf("unexpected signature suffix")
			}
			signatures := c.GetSignaturesOfType(subject, checker.SignatureKindCall)
			out.number(uint64(len(signatures)))
			for _, signature := range signatures {
				out.number(g.add(c.GetReturnTypeOfSignature(signature)))
				parameters := signature.Parameters()
				out.number(uint64(len(parameters)))
				for _, parameter := range parameters {
					out.number(g.add(checker.Checker_getTypeOfSymbol(c, parameter)))
					out.number(g.add(c.GetTypeOfSymbolAtLocation(parameter, node)))
					declaration := parameter.ValueDeclaration
					out.yes(declaration != nil && declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil)
				}
			}
		case "index-number", "index-string":
			if len(pieces) != 4 {
				return "", fmt.Errorf("unexpected index suffix")
			}
			key := checker.Checker_numberType(c)
			if operation == "index-string" {
				key = checker.Checker_stringType(c)
			}
			t := checker.Checker_getIndexTypeOfType(c, subject, key)
			present = t != nil
			if present {
				roots = append(roots, g.add(t))
			}
		case "property-global", "property-location":
			if len(pieces) != 5 {
				return "", fmt.Errorf("property operation requires a name")
			}
			property := checker.Checker_getPropertyOfType(c, subject, pieces[4])
			present = property != nil
			if present {
				t := checker.Checker_getTypeOfSymbol(c, property)
				if operation == "property-location" {
					t = c.GetTypeOfSymbolAtLocation(property, node)
				}
				roots = append(roots, g.add(t))
			}
		default:
			return "", fmt.Errorf("unknown require-await type operation")
		}
	default:
		return "", fmt.Errorf("unknown require-await action")
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(present)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
