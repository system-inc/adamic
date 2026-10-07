package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

func (p *Program) callableReturnShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("callable-return-shape requires identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	for _, signature := range c.GetSignaturesOfType(p.typesByID[id-1], checker.SignatureKindCall) {
		roots = append(roots, g.add(c.GetReturnTypeOfSignature(signature)))
	}
	out.yes(true)
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
