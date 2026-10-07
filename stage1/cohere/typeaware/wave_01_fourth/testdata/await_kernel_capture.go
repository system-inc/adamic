package core

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// Test-only observations of unchanged production syntax predicates.
func Wave01AwaitKernelCapture(bodies []*ast.Node) string {
	var out strings.Builder
	for i, body := range bodies {
		fmt.Fprintf(&out, "%d %t %t\n", i, requireAwaitBodyIsEmpty(body), requireAwaitContainsAwait(body))
	}
	return out.String()
}
