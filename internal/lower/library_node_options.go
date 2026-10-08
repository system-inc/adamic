package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// nodeHostConsumesArgument proves both consumption and absence. This is not a
// general exemption for Node declarations or for readonly arguments. The fs-file
// lowerer replaces these option objects with scalar arguments; it neither stores
// nor returns the object, nor passes it to the runtime as a wider object view.
func (l *lowering) nodeHostConsumesArgument(node *ast.Node, index int) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	if !l.optionalNodeHostCall(node) {
		return false
	}
	call := node.AsCallExpression()
	if index < 0 || index >= len(call.Arguments.Nodes) {
		return false
	}
	optionIndex := 1
	switch l.nodeLibraryMember(node) {
	case "node:fs.statSync", "node:fs.mkdirSync", "node:fs.readFileSync", "node:fs.rmSync", "node:fs.mkdtempSync":
	case "node:fs.writeFileSync":
		optionIndex = 2
	default:
		// In particular, readdirSync still passes its options object onward.
		return false
	}
	if index != optionIndex {
		return false
	}
	argument := ast.SkipParentheses(call.Arguments.Nodes[index])
	// exactObject follows only unannotated const bindings back to literals.
	// The directory stat lowerer also proves unannotated const literals with as const.
	// Structural annotations, other casts, spreads and parameters can hide a subtype's
	// fields, even when every declared field is read using its declared type.
	if !l.exactObject(argument, 0) && !(l.nodeLibraryMember(node) == "node:fs.statSync" && l.nodeFSDirectoryStatOptions(argument)) {
		return false
	}
	// The consumed fs options are scalar. Do not let an exact outer literal hide
	// a nested optional widening through an object-valued field or a callback.
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(argument)) {
		if !nodeHostScalarOption(l.checker.GetTypeOfSymbol(field)) {
			return false
		}
	}
	return true
}

func nodeHostScalarOption(proven *checker.Type) bool {
	if proven == nil {
		return false
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !nodeHostScalarOption(member) {
				return false
			}
		}
		return true
	}
	return proven.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0
}
