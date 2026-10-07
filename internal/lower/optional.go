package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// optionalStrings are the optional string parameters of the library methods stage 0 lowers, by
// method and position, with what JavaScript uses when the argument is undefined. These are every one
// in lib.es2024 on a method lowered here (checked against the bundled library's declarations:
// padStart and padEnd's padString, normalize's form, and join's separator, which arrayMethod gives
// its default); localeCompare and the
// toLocale methods take others, and aren't lowered.
var optionalStrings = map[string]map[int]string{
	"padStart":  {1: " "},
	"padEnd":    {1: " "},
	"normalize": {0: "NFC"},
}

// orDefault is an argument to an optional string parameter, as JavaScript reads it: undefined is the
// parameter's default. A string that may be undefined is held as a null reference, which the runtime
// would read as a string, so it's given the default here, where undefined is still visible: undefined
// written out becomes the default itself, and a string | undefined becomes value ?? default.
func (l *lowering) orDefault(node *ast.Node, value ir.Expression, fallback string) ir.Expression {
	if _, isUndefined := value.(ir.Undefined); isUndefined {
		return ir.StringConstant{Index: l.constant(fallback)}
	}
	if value.Type() != ir.String || !l.mayBeUndefined(node) {
		return value
	}
	return ir.Coalesce{Value: value, Fallback: ir.StringConstant{Index: l.constant(fallback)}, Of: ir.String}
}

// mayBeUndefined reports whether the checker's type for a node includes undefined.
func (l *lowering) mayBeUndefined(node *ast.Node) bool {
	return l.includesUndefined(l.checker.GetTypeAtLocation(node))
}
