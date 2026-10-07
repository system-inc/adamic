package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// numericLiteral returns the checker-held literal, including whether it is held.
func (p *Program) numericLiteral(c *checker.Checker, node *ast.Node, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("numeric-literal requires a type identity")
	}
	id, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	if t.Flags()&(checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) == 0 {
		return "", fmt.Errorf("numeric-literal requires a numeric literal")
	}
	out := &fields{}
	out.number(1)
	out.text("numeric-literal")
	value, held := t.AsLiteralType().Value().(fmt.Stringer)
	out.yes(held)
	text := ""
	if held {
		text = value.String()
	}
	out.text(text)
	return out.String(), nil
}
