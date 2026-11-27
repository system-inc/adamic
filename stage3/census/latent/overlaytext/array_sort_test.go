package overlaytext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatentArraySortReadsNamedComparatorSignature(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/object.go
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "internal", "lower", "object.go"))
	if err != nil {
		t.Fatal(err)
	}
	const header = "func (l *lowering) arraySort(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {"
	_, body, found := strings.Cut(string(source), header)
	if !found {
		t.Fatal("arraySort signature hook moved")
	}
	body, _, found = strings.Cut(body, "\n}\n")
	const read = "\n\tdeclared := l.result.Functions[function]\n"
	if !found || strings.Count(body, read) != 1 {
		t.Fatal("arraySort named comparator signature read moved")
	}
}
