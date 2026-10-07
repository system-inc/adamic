package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// callbackParameters exposes apparent first-parameter types, indexing rest arrays.
func (p *Program) callbackParameters(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("callback-parameters requires a type identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	for _, signature := range c.GetSignaturesOfType(p.typesByID[id-1], checker.SignatureKindCall) {
		parameters := checker.Signature_parameters(signature)
		if len(parameters) == 0 {
			continue
		}
		parameter := parameters[0]
		t := checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(parameter, node))
		if declaration := parameter.ValueDeclaration; declaration != nil && declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
			t = checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c))
		}
		if t != nil {
			roots = append(roots, g.add(t))
		}
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
