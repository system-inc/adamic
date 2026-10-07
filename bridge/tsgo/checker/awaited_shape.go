package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// awaitedShape exposes the checker operation on a live type identity, not a lint verdict.
func (p *Program) awaitedShape(out *fields, c *checker.Checker, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("awaited-shape requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	subject := checker.Checker_getAwaitedType(c, p.typesByID[id-1])
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	if subject != nil {
		roots = append(roots, g.add(subject))
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(subject != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
