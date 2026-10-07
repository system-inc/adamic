package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Signature kinds and return flags are compiler facts, not a lint verdict.
func (p *Program) signatureKinds(out *fields, c *checker.Checker, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("signature-kinds requires an identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	subject := p.typesByID[id-1]
	out.number(uint64(len(c.GetSignaturesOfType(subject, checker.SignatureKindConstruct))))
	calls := c.GetSignaturesOfType(subject, checker.SignatureKindCall)
	out.number(uint64(len(calls)))
	for _, call := range calls {
		result := checker.Checker_getReturnTypeOfSignature(c, call)
		out.yes(result != nil)
		if result != nil {
			out.number(uint64(result.Flags()))
		}
	}
	return out.String(), nil
}
