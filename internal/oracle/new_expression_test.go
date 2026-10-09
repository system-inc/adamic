package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{"uint16", "uint16_notyet", "class_value", "class_cache_or", "class_cache_local", "class_cache_capture_notyet"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/new_expression_" + name + ".a", name == "uint16" || name == "uint16_notyet" || name == "class_value" || name == "class_cache_or" || name == "class_cache_local", false})
	}
}

// The c2 runtime provides Uint16 storage and conversion. Keep the original
// conversion witness in the Node oracle as an admitted program.
func TestNewExpressionUint16IsAdmitted(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_uint16_notyet.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "4464\n" || len(node.stderr) != 0 {
		t.Fatalf("Node conversion: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	if err != nil {
		t.Fatalf("want c2 Uint16Array admission, got %v", err)
	}
}

// A generated cache closure may capture only its cache cell. An initializer
// supplied by the caller needs another capture and must remain unsupported.
func TestNewExpressionCacheCaptureRemainsNotYet(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_class_cache_capture_notyet.a"))
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "9\n" || len(node.stderr) != 0 {
		t.Fatalf("Node initializer: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "a lexical constructor-cache initializer needing additional captures" {
		t.Fatalf("want additional-capture NotYet, got %v", err)
	}
}
