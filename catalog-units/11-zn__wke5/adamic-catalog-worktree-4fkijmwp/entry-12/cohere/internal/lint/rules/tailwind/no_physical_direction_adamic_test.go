package tailwind

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// An Adamic `.a` file's class strings are checked as a `.ts` file's are (#kwt1htp).
func TestNoPhysicalDirectionChecksAnAdamicFileAsTypeScript(t *testing.T) {
	t.Parallel()
	source := "export const c = cn('flex ml-4');\n"
	rule_testing.ExpectSameFindings(t,
		rule_testing.Run(t, NoPhysicalDirection, "/repository/source/components/Thing.ts", source),
		rule_testing.Run(t, NoPhysicalDirection, "/repository/source/components/Thing.a", source))
}
