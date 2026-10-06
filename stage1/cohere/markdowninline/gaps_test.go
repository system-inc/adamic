package markdowninline

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestDelimiterExpressionGap(t *testing.T) {
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
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a RegularExpressionLiteral" {
		t.Fatalf("gap changed: %v; update GAPS.md and remove the scanner if general regex becomes available", err)
	}
}
