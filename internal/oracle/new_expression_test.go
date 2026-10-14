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

// Uint16 constructors lower, so source Node holds their conversions and views.
func TestNewExpressionUint16AgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"uint16", "uint16_notyet"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/new_expression_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			if expected.exitCode != 0 || len(expected.stderr) != 0 {
				t.Fatalf("Node constructor: %#v", expected)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, _ := natively(t, program)
			for backend, observed := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(expected, observed); difference != "" {
					t.Errorf("%s: %s: exit %d stdout %q stderr %q", backend, difference, observed.exitCode, observed.stdout, observed.stderr)
				}
			}
		})
	}
}

// A generated cache closure may capture only its cache cell. An initializer
// supplied by the caller needs another capture and must remain unsupported.
func TestNewExpressionCacheCaptureRemainsNotYet(t *testing.T) {
	t.Parallel()
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
