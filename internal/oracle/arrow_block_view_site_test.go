package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/arrow_block_box.a",
		"internal/oracle/testdata/arrow_block_subclass.a",
		"internal/oracle/testdata/arrow_expression_box.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

func TestArrowBlockStructuralViewRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/arrow_block_structural.a"))
	if err != nil {
		t.Fatal(err)
	}
	observation := onNode(t, path)
	if observation.exitCode != 0 || string(observation.stdout) != "2\n" || len(observation.stderr) != 0 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", observation.exitCode, observation.stdout, observation.stderr)
	}
	_, err = lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) {
		t.Fatalf("want nominal refusal, got %v", err)
	}
	if !strings.Contains(refusal.What, "accepted here as the class 'Box'") || !strings.Contains(refusal.What, "not an instance of that class") || !strings.Contains(refusal.Fix, "adamic/nominal-class") {
		t.Fatalf("want cohere's nominal Box refusal, got %v", err)
	}

	if refusal.Where != path+":5:42" {
		t.Fatalf("unexpected refusal site: %s", refusal.Where)
	}
}
