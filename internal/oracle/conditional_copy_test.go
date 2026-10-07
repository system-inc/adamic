package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// The refused probe can also run through the ordinary differential harness
// when removing the invocation to measure its pre-refusal behavior.
func init() {
	if os.Getenv("ADAMIC_CONDITIONAL_COPY_ORACLE") == "1" {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/type_rules_conditional_copy.a", true, false,
		})
	}
}

func TestConditionalCopyRefusedByCohere(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/type_rules_conditional_copy.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "1\n" {
		t.Fatalf("Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/invariant-mutable") || !strings.Contains(err.Error(), path+":4:27:") {
		t.Fatalf("want cohere conditional refusal at 4:27, got %v", err)
	}
	t.Logf("Node stdout %q; %v", node.stdout, err)
}
