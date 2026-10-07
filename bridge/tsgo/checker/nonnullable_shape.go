package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

func (p *Program) nonnullableShape(c *checker.Checker, node *ast.Node, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("nonnullable-shape requires a type identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	out := &fields{}
	out.number(1)
	out.text("nonnullable-shape")
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	root := g.add(c.GetNonNullableType(p.typesByID[id-1]))
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids([]uint64{root})
	out.number(0)
	g.write(out)
	return out.String(), nil
}
