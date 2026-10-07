package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Type leaf facts expose raw checker operations on a live identity, not lint decisions.
func (p *Program) typeLeafFacts(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("type-leaf-facts requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	t := p.typesByID[id-1]
	out.number(uint64(t.Flags()))
	out.yes(checker.Checker_isArrayOrTupleType(c, t))
	out.number(uint64(len(c.GetSignaturesOfType(t, checker.SignatureKindCall))))
	literal, present := "", false
	if t.Flags()&checker.TypeFlagsStringLiteral != 0 {
		literal, present = t.AsLiteralType().Value().(string)
	}
	out.yes(present)
	out.text(literal)
	writeLeafSymbolOrigin(out, p, t.Symbol())
	out.yes(len(parts) == 3)
	if len(parts) == 3 {
		out.yes(checker.Checker_getPropertyOfType(c, t, parts[2]) != nil)
	}
	return out.String(), nil
}

// Internal anonymous symbol names may contain invalid UTF-8. The wire name is
// display text; property lookup above still uses the original checker string.
func writeLeafSymbolOrigin(out *fields, p *Program, symbol *ast.Symbol) {
	out.yes(symbol != nil)
	if symbol == nil {
		return
	}
	out.text(strings.ToValidUTF8(symbol.Name, "\ufffd"))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			panic("type symbol declaration has no source")
		}
		out.text(source.FileName())
		out.yes(source.IsDeclarationFile)
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
	}
}
