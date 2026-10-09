package markdowninline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestDelimiterExpressionMatchesNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("gaps/1_delimiter_expression.ts")
	if err != nil {
		t.Fatal(err)
	}
	answer := onNode(t, path)
	clean(t, "Node gap", answer)
	equal(t, "Node gap", answer.stdout, []byte("a\\*b\n"))
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	native, binary := natively(t, lowered)
	clean(t, "native delimiter expression", native)
	equal(t, "native delimiter expression", native.stdout, answer.stdout)
	backend := onJavaScriptBackend(t, lowered)
	clean(t, "backend delimiter expression", backend)
	equal(t, "backend delimiter expression", backend.stdout, answer.stdout)
	leaks(t, lowered, binary)
}

// The narrow canary of tools 26226fda selects this package (developer tools, Oct 9).
