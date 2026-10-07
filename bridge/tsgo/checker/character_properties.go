package checker

import (
	"fmt"
	"strings"
	"unicode"
)

// characterProperties supplies Unicode scalar categories and simple case maps.
func (p *Program) characterProperties(out *fields, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("character-properties requires text")
	}
	runes := []rune(parts[1])
	out.number(uint64(len(runes)))
	for _, r := range runes {
		out.text(string(r))
		out.yes(unicode.IsLetter(r))
		out.yes(unicode.IsDigit(r))
		out.yes(unicode.IsUpper(r))
		out.yes(unicode.IsLower(r))
		out.text(string(unicode.ToUpper(r)))
		out.text(string(unicode.ToLower(r)))
	}
	return out.String(), nil
}
