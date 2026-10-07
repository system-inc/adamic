package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Exposes a type's string or number index type, with no lint decision.
func (p *Program) typeIndex(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	parts := strings.Split(question, "\n")
	if len(parts) != 3 {
		return fmt.Errorf("type-index requires identity and index kind")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return fmt.Errorf("unknown checker type identity")
	}
	index := checker.Checker_numberType(c)
	if parts[2] == "string" {
		index = checker.Checker_stringType(c)
	} else if parts[2] != "number" {
		return fmt.Errorf("invalid index kind")
	}
	t := checker.Checker_getIndexTypeOfType(c, p.typesByID[id-1], index)
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(t != nil)
	var roots []uint64
	if t != nil {
		roots = append(roots, g.add(t))
	}
	out.ids(roots)
	out.number(0)
	g.write(out)
	return nil
}
