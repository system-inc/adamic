package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"array_dense", "array_holes_notyet", "array_optional_length_notyet"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/new_expression_" + name + ".a", name == "array_dense", false,
		})
	}
}

func TestNewExpressionArrayHolesRemainNotYet(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_array_holes_notyet.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "3\n0\n,,\n" || len(node.stderr) != 0 {
		t.Fatalf("Node holes: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "an Array constructor creating holes" {
		t.Fatalf("want array-hole NotYet, got %v", err)
	}
}

func TestNewExpressionOptionalArrayLengthRemainsNotYet(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_array_optional_length_notyet.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "2\n,\n" || len(node.stderr) != 0 {
		t.Fatalf("Node optional length: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "an Array constructor creating holes" {
		t.Fatalf("want array-hole NotYet, got %v", err)
	}
}
