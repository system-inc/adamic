package flow_test

import (
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/lower"
)

// The graph's tests lower every program they build graphs of, and lower depends on this package, so
// lowering reaches them from here.
func init() {
	flow.Lower = lower.Lower
}
