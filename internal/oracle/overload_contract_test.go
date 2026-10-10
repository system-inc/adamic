package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Node observes these contracts; Adamic must retain its strict proof obligation.
// A refused contract cannot reach either backend or a sanitizer run.
func TestOverloadContractRulings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, output, refusal string }{
		{"token", "undefined\n", "overload 1 of createToken result"},
		{"trampoline", "node\n", "overload 1 of createBinaryExpressionTrampoline result"},
		{"serializer", "node/arg\n", "overload 2 of setSerializerContextAnd parameter cb"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/overload_contracts/"+test.name+".a"))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != test.output {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q", observed.exitCode, observed.stdout, observed.stderr)
			}
			_, err := lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, test.refusal) {
				t.Fatalf("want refusal %q, got %v", test.refusal, err)
			}
		})
	}
}
