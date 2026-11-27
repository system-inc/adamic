package overlaytext

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLatentNamespaceValueScanEntryMatchesMain(t *testing.T) {
	t.Parallel()
	// Reads: internal/lower/namespaces.go
	// Reads: internal/lower/namespace_value.go
	// Reads: stage3/census/latent/lower.go.txt
	// Reads: stage3/census/latent/units.go.txt
	// Reads: stage3/census/latent/speculative.go.txt
	repo := filepath.Join("..", "..", "..", "..")
	anchors := []struct{ path, text string }{
		{"internal/lower/namespaces.go", "func (l *lowering) namespaceInitialization(modules []*ast.SourceFile) error {\n\tl.noteNamespaceValues(modules)\n"},
		{"internal/lower/namespace_value.go", "func (l *lowering) noteNamespaceValues(modules []*ast.SourceFile) {"},
		{"internal/lower/namespace_value.go", "return l.namespaceValues[l.symbol(declaration.Name())]"},
		{"stage3/census/latent/lower.go.txt", "l := fresh()\n\t\t\t\tl.latentInstallNamespaceValues(&namespaceFacts)"},
		{"stage3/census/latent/units.go.txt", "base := fresh()\n\t\tbase.latentInstallNamespaceValues(&namespaceFacts)"},
		{"stage3/census/latent/speculative.go.txt", "l.latentInstallNamespaceValues(&namespaceFacts)\n\t\t\tl.latentRegister(program.Files())"},
		{"stage3/census/latent/lower.go.txt", "l.namespaceValues = facts.values\n\tfirst := len(latentFindings)\n\terr := l.latentRefuse(file)"},
	}
	for _, anchor := range anchors {
		source, err := os.ReadFile(filepath.Join(repo, anchor.path))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(source), anchor.text) != 1 {
			t.Errorf("%s: namespace value scan hook moved: %q", anchor.path, anchor.text)
		}
	}
}
