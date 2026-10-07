package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// typeArguments exposes reference and alias arguments without choosing wrappers.
func (p *Program) typeArguments(out *fields, c *checker.Checker, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("type-arguments requires a type identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	symbolName, aliasName := "", ""
	if symbol := t.Symbol(); symbol != nil {
		symbolName = symbol.Name
	}
	var arguments, aliases []uint64
	if t.Flags()&checker.TypeFlagsObject != 0 && t.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		for _, argument := range checker.Checker_getTypeArguments(c, t) {
			arguments = append(arguments, g.add(argument))
		}
	}
	if alias := checker.Type_alias(t); alias != nil {
		if symbol := alias.Symbol(); symbol != nil {
			aliasName = symbol.Name
		}
		for _, argument := range alias.TypeArguments() {
			aliases = append(aliases, g.add(argument))
		}
	}
	out.text(symbolName)
	out.text(aliasName)
	out.ids(arguments)
	out.ids(aliases)
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids(append(append([]uint64{}, arguments...), aliases...))
	out.number(0)
	g.write(out)
	return out.String(), nil
}
