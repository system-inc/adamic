package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

func (p *Program) indexedTypeShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 3 || (parts[2] != "number" && parts[2] != "string") {
		return "", fmt.Errorf("indexed-type-shape requires identity and index kind")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	key := checker.Checker_numberType(c)
	if parts[2] == "string" {
		key = checker.Checker_stringType(c)
	}
	t := checker.Checker_getIndexTypeOfType(c, p.typesByID[id-1], key)
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	if t != nil {
		roots = append(roots, g.add(t))
	}
	out.yes(true)
	out.yes(t != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
