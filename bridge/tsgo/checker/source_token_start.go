package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Raw trivia position in the requested source, including out-of-file positions.
// Go's diagnostic caller chooses the source independently of a declaration's provenance.
func (p *Program) sourceTokenStart(out *fields, node *ast.Node, question string) (string, error) {
	value, ok := strings.CutPrefix(question, "source-token-start\n")
	position, err := strconv.ParseUint(value, 10, 31)
	if !ok || err != nil || strconv.FormatUint(position, 10) != value || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-token-start requires a SourceFile and integer position")
	}
	out.number(uint64(scanner.SkipTrivia(node.AsSourceFile().Text(), int(position))))
	return out.String(), nil
}
