package estree

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostfixValueGap(t *testing.T) {
	path, err := filepath.Abs("gaps/postfixValue.ts")
	if err != nil {
		t.Fatal(err)
	}
	got := onNode(t, path)
	if string(got) != "7\n1\n" {
		t.Fatalf("Node %q", got)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "a PostfixUnaryExpression") {
		t.Fatalf("postfix value gap changed: %v", err)
	}
	t.Logf("Node 7, 1; lowering: %v", err)
}
