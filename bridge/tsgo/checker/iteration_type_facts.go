package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Compiler properties, index types and signature parameters, without lint judgments.
func (p *Program) iterationTypeFacts(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	split := strings.Split(question, "\n")
	if len(split) != 3 {
		return fmt.Errorf("iteration-type-facts requires type identity and operation")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	switch split[2] {
	case "iterator", "asyncIterator", "asyncDispose":
		out.yes(checker.Checker_getPropertyOfType(c, t, checker.Checker_getPropertyNameForKnownSymbolName(c, split[2])) != nil)
		return nil
	case "values":
		out.yes(c.IsArrayLikeType(t))
		if checker.IsTupleType(t) {
			for _, a := range c.GetTypeArguments(t) {
				roots = append(roots, g.add(a))
			}
		} else if c.IsArrayLikeType(t) {
			if a := c.GetNumberIndexType(t); a != nil {
				roots = append(roots, g.add(a))
			}
		} else if t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
			args := c.GetTypeArguments(t)
			if len(args) > 0 {
				roots = append(roots, g.add(args[0]))
			}
		}
	case "callbacks":
		for _, signature := range c.GetSignaturesOfType(t, checker.SignatureKindCall) {
			params := checker.Signature_parameters(signature)
			if len(params) == 0 {
				continue
			}
			param := params[0]
			value := checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(param, node))
			if d := param.ValueDeclaration; d != nil && d.Kind == ast.KindParameter && d.AsParameterDeclaration().DotDotDotToken != nil {
				value = checker.Checker_getIndexTypeOfType(c, value, checker.Checker_numberType(c))
			}
			if value != nil {
				roots = append(roots, g.add(value))
			}
		}
	default:
		return fmt.Errorf("unknown iteration type operation")
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return nil
}
