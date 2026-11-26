package lower

// Scratch for #xfpsqtn's scout, never to land: traces each decline of neverRead under checkNeverArguments,
// and, under XFP_LIFT=1, lets the call through to see the next stop.

import (
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

var xfpTracing bool

func (l *lowering) xfp(tag string, node *ast.Node, detail string) {
	if !xfpTracing {
		return
	}
	fmt.Fprintf(os.Stderr, "XFP %s at %s: %q %s\n", tag, l.program.Where(node), strings.Join(strings.Fields(sourceExpression(node)), " "), detail)
}
