package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Raw declaration-file facts for a borrowed symbol identity. No lint decisions.
func (p *Program) symbolDeclarationProvenance(out *fields, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("symbol-declaration-provenance requires a symbol identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id > uint64(len(p.symbolsByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker symbol identity")
	}
	var declarations []*ast.Node
	if id != 0 {
		declarations = p.symbolsByID[id-1].Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			panic("symbol declaration has no source")
		}
		out.text(source.FileName().AsString())
		out.yes(source.IsDeclarationFile)
	}
	return out.String(), nil
}
