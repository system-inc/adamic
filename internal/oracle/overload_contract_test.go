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
func TestOverloadContractRulings(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct{ name, output, refusal string }{
		{"token", "undefined\n", "overload 1 of createToken result"},
		{"trampoline", "node\n", "overload 1 of createBinaryExpressionTrampoline result"},
		{"serializer", "node/arg\n", "overload 2 of setSerializerContextAnd parameter cb"},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			path, pathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/overload_contracts/"+testCase.name+".a"),
			)
			if pathError != nil {
				test.Fatal(pathError)
			}
			observed := onNode(test, path)
			if observed.exitCode != 0 || string(observed.stdout) != testCase.output {
				test.Fatalf("Node: exit %d, stdout %q, stderr %q", observed.exitCode, observed.stdout, observed.stderr)
			}
			_, loweredError := lowered(test, path)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) || !strings.Contains(refused.What, testCase.refusal) {
				test.Fatalf("want refusal %q, got %v", testCase.refusal, loweredError)
			}
		})
	}
}
