package base

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// An Adamic `.a` file throws under the same rule as a `.ts` one, and `Service.test.a`, Adamic's test
// file, is exempt as `Service.test.ts` is (#kwt1htp).
func TestNoBareThrowReadsAnAdamicFileAsTypeScript(t *testing.T) {
	t.Parallel()
	source := "export function f(): void { throw new Error('x'); }"
	rule_testing.ExpectSameFindings(t,
		rule_testing.RunTyped(t, ConsistencyNoBareThrow, "/repository/source/modules/thing/Service.ts", source),
		rule_testing.RunTyped(t, ConsistencyNoBareThrow, "/repository/source/modules/thing/Service.a", source))
	rule_testing.ExpectClean(t,
		rule_testing.RunTyped(t, ConsistencyNoBareThrow, "/repository/source/modules/thing/Service.test.a", source))
}
