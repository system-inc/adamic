package wave08core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
)

// AwaitContractFields returns contextual, heritage and resolved/declared signature
// inputs. It does not decide which context is applicable or whether async is required.
func AwaitContractFields(c *checker.Checker, node *ast.Node, id func(*checker.Type) uint64) []string {
	out := []string{"1", "wave08-await-contract"}
	number := func(n int) { out = append(out, strconv.Itoa(n)) }
	typ := func(t *checker.Type) { out = append(out, strconv.FormatUint(id(t), 10)) }
	typ(checker.Checker_getContextualType(c, node, checker.ContextFlagsNone))
	var heritage []*checker.Type
	if parent := node.Parent; parent != nil && (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression) && node.Name() != nil && node.Name().Kind == ast.KindIdentifier {
		var clauses *ast.NodeList
		if parent.Kind == ast.KindClassDeclaration {
			clauses = parent.AsClassDeclaration().HeritageClauses
		} else {
			clauses = parent.AsClassExpression().HeritageClauses
		}
		if clauses != nil {
			for _, clause := range clauses.Nodes {
				for _, base := range clause.AsHeritageClause().Types.Nodes {
					t := c.GetTypeAtLocation(base)
					if t != nil {
						if member := checker.Checker_getPropertyOfType(c, t, node.Name().Text()); member != nil {
							heritage = append(heritage, c.GetTypeOfSymbolAtLocation(member, node))
						}
					}
				}
			}
		}
	}
	number(len(heritage))
	for _, t := range heritage {
		typ(t)
	}
	var call *ast.Node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
			call = parent
			break
		}
	}
	if call == nil {
		number(0)
		return out
	}
	number(1)
	number(int(call.Kind))
	number(call.Pos())
	number(call.End())
	resolved := checker.Checker_getResolvedSignature(c, call, nil, checker.CheckModeNormal)
	var declared *checker.Signature
	if resolved != nil {
		declared = resolved.Target()
	}
	if declared == nil {
		number(0)
		return out
	}
	number(1)
	number(len(declared.TypeParameters()))
	for _, t := range declared.TypeParameters() {
		typ(t)
	}
	parameters := declared.Parameters()
	number(len(parameters))
	for _, parameter := range parameters {
		typ(checker.Checker_getTypeOfSymbol(c, parameter))
	}
	if declared.HasRestParameter() {
		number(1)
	} else {
		number(0)
	}
	var rest *checker.Type
	if declared.HasRestParameter() && len(parameters) > 0 {
		rest = checker.Checker_getIndexTypeOfType(c, checker.Checker_getTypeOfSymbol(c, parameters[len(parameters)-1]), checker.Checker_numberType(c))
	}
	typ(rest)
	typ(checker.Checker_getReturnTypeOfSignature(c, declared))
	parameters = resolved.Parameters()
	number(len(parameters))
	for _, parameter := range parameters {
		typ(checker.Checker_getTypeOfSymbol(c, parameter))
	}
	typ(checker.Checker_getContextualType(c, call, checker.ContextFlagsNone))
	return out
}
