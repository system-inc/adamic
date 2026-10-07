// Oracle-only access to the unchanged production message builder.
package react

import (
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func Wave26StaticReportingFallback() rule.Message {
	return staticComponentsMessage(rule.Context{}, &high_level_intermediate_representation.Function{}, 0)
}
