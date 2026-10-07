package checker

import (
	"fmt"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// unicodeNodeText exports Unicode simple-case mappings, not a naming judgment.
// JavaScript full case conversion can expand a scalar; Go simple case does not.
func (p *Program) unicodeNodeText(out *fields, node *ast.Node, question string) (string, error) {
	if question != "unicode-node-text" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("unicode-node-text requires an Identifier and no arguments")
	}
	runes := []rune(node.Text())
	out.number(uint64(len(runes)))
	for _, r := range runes {
		out.number(uint64(r))
		out.number(uint64(unicode.ToUpper(r)))
		out.number(uint64(unicode.ToLower(r)))
	}
	return out.String(), nil
}
