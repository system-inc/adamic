package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strconv"
	"strings"
)

// otherFile delegates an exact selector to another source in the same program. No new program
// or private checker is created, and the shared RuleContext records the entire request.
func (p *Program) otherFile(node *ast.Node, question string) (string, error) {
	if node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("other-file requires a SourceFile")
	}
	parts := strings.SplitN(question, "\n", 6)
	if len(parts) != 6 {
		return "", fmt.Errorf("other-file requires path, start, end, kind and question")
	}
	start, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil || strconv.FormatUint(start, 10) != parts[2] {
		return "", fmt.Errorf("invalid other-file start")
	}
	end, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil || strconv.FormatUint(end, 10) != parts[3] {
		return "", fmt.Errorf("invalid other-file end")
	}
	if strings.HasPrefix(parts[5], "other-file") {
		return "", fmt.Errorf("recursive other-file question")
	}
	return p.Inspect(parts[1], start, end, parts[4], parts[5])
}
