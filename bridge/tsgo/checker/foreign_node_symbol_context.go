package checker

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A raw symbol lookup in an already-loaded source, including bundled libraries.
func (p *Program) foreignNodeSymbolContext(out *fields, c *checker.Checker, anchor *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 5 || parts[0] != "foreign-node-symbol-context" || anchor.Kind != ast.KindSourceFile || !utf8.ValidString(parts[1]) || strings.ContainsRune(parts[1], 0) {
		return "", fmt.Errorf("invalid foreign-node-symbol-context")
	}
	start, firstErr := strconv.ParseUint(parts[2], 10, 31)
	end, lastErr := strconv.ParseUint(parts[3], 10, 31)
	if firstErr != nil || lastErr != nil || strconv.FormatUint(start, 10) != parts[2] || strconv.FormatUint(end, 10) != parts[3] || start >= end {
		return "", fmt.Errorf("invalid foreign node span")
	}
	source := p.Compiler.GetSourceFile(parts[1])
	if source == nil || end > uint64(len(source.Text())) {
		return "", fmt.Errorf("foreign source or span unavailable")
	}
	var found *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if found != nil {
			return
		}
		if node.Pos() == int(start) && node.End() == int(end) && strings.TrimPrefix(node.Kind.String(), "Kind") == parts[4] {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return found != nil })
	}
	walk(source.AsNode())
	if found == nil {
		return "", fmt.Errorf("foreign node not found")
	}
	return p.symbolContext(out, c, found, "symbol-context")
}
