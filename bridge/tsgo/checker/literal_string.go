package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) literalString(out *fields, question string) (string, error) {
	split := strings.Split(question, "\n")
	if len(split) != 2 {
		return "", fmt.Errorf("literal-string requires identity")
	}
	id, e := strconv.ParseUint(split[1], 10, 64)
	if e != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != split[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	if t.Flags()&(checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) == 0 {
		return "", fmt.Errorf("literal-string requires numeric literal")
	}
	value, held := t.AsLiteralType().Value().(fmt.Stringer)
	out.yes(held)
	text := ""
	if held {
		text = value.String()
	}
	out.text(text)
	return out.String(), nil
}
