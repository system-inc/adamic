package oracle

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, name := range []string{
		"uint16",
		"uint16_conversion",
		"class_value",
		"class_cache_or",
		"class_cache_local",
		"class_cache_capture_notyet",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/new_expression_" + name + ".a", name == "uint16" ||
			name == "uint16_conversion" ||
			name == "class_value" ||
			name == "class_cache_or" ||
			name == "class_cache_local", false})
	}
}

// Uint16 constructors lower, so source Node holds their conversions and views.
func TestNewExpressionUint16AgreesWithNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"uint16", "uint16_conversion"} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/new_expression_"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			expected := onNode(test, path)
			if expected.exitCode != 0 || len(expected.stderr) != 0 {
				test.Fatalf("Node constructor: %#v", expected)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			native, _ := natively(test, program)
			for backend, observed := range map[string]run{
				"native":     native,
				"release":    released(test, program),
				"JavaScript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(expected, observed); difference != "" {
					test.Errorf(
						"%s: %s: exit %d stdout %q stderr %q",
						backend,
						difference,
						observed.exitCode,
						observed.stdout,
						observed.stderr,
					)
				}
			}
		})
	}
}

// A generated cache closure may capture only its cache cell. An initializer
// supplied by the caller needs another capture and must remain unsupported.
func TestNewExpressionCacheCaptureRemainsNotYet(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/new_expression_class_cache_capture_notyet.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	node := onNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != "9\n" || len(node.stderr) != 0 {
		test.Fatalf("Node initializer: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	_, loweredError := lowered(test, path)
	var notYet *lower.NotYet
	if !errors.As(loweredError, &notYet) ||
		notYet.What != "a lexical constructor-cache initializer needing additional captures" {
		test.Fatalf("want additional-capture NotYet, got %v", loweredError)
	}
}
