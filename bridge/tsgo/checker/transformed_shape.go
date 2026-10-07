package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) transformedShape(out *fields, c *checker.Checker, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 3 || (split[2] != "non-nullable" && split[2] != "constraint") {
		return "", fmt.Errorf("transformed shape requires identity")
	}
	id, e := strconv.ParseUint(split[1], 10, 64)
	if e != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	if split[2] == "non-nullable" {
		t = c.GetNonNullableType(t)
	} else {
		t = checker.Checker_getBaseConstraintOfType(c, t)
	}
	g := &graph{program: p, checker: c, seen: map[*checker.Type]bool{}}
	var roots []uint64
	if t != nil {
		roots = append(roots, g.add(t))
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(t != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
