package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Presence and flags of the numeric index type, never an array-like verdict.
func (p *Program) numberIndexType(out *fields, c *checker.Checker, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("number-index-type requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("invalid type identity")
	}
	index := checker.Checker_getIndexTypeOfType(c, p.typesByID[id-1], checker.Checker_numberType(c))
	out.yes(index != nil)
	if index != nil {
		out.number(uint64(index.Flags()))
	}
	return out.String(), nil
}
