package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Property, index and tuple facts; Adamic decides which a contextual step uses.
func (p *Program) wave19TypeMembers(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 3)
	if len(parts) != 3 {
		return "", fmt.Errorf("wave19-type-members requires identity and property")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	subject := p.typesByID[id-1]
	var property uint64
	if symbol := checker.Checker_getPropertyOfType(c, subject, parts[2]); symbol != nil {
		property = g.add(checker.Checker_getTypeOfSymbol(c, symbol))
	}
	stringIndex := g.add(checker.Checker_getIndexTypeOfType(c, subject, checker.Checker_stringType(c)))
	numberIndex := g.add(checker.Checker_getIndexTypeOfType(c, subject, checker.Checker_numberType(c)))
	tuple := checker.IsTupleType(subject)
	var elements []uint64
	if tuple {
		for _, element := range checker.Checker_getTypeArguments(c, subject) {
			elements = append(elements, g.add(element))
		}
	}
	roots := append([]uint64{property, stringIndex, numberIndex}, elements...)
	out.number(property)
	out.number(stringIndex)
	out.number(numberIndex)
	out.yes(tuple)
	out.ids(elements)
	out.text(p.wave19ContractTypes(g, roots))
	return out.String(), nil
}
