package load

import (
	"reflect"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// relationOverflowCode is TS2859, the checker's Excessive_complexity_comparing_types_0_and_1.
const relationOverflowCode = 2859

// RelationOverflow is the first TS2859 the checker holds, formatted as diagnostics formats one, or ""
// when it holds none (#apyq5ve).
//
// A relation query that spends its step budget answers "not related", the checker caches that pair
// as failed by overflow, and it reports TS2859 once: globally when nothing was being checked, as
// after load, or on the file of the node a type query set it checking. Lowering reads a relation's
// answer as a fact, and sometimes as permission, so a program whose checker holds one is never
// lowered. This asks whether the checker holds any, never whether one was added during a lowering,
// because a cached overflow answers false again with no new diagnostic.
//
// It reads the checker's whole collection, global and attached to any file, so a TS2859 the checker
// pinned to a node in the prelude or a lib file is read too (Compiler's question 1). Asking the
// checker file by file would check any lib file load skipped. The collection is unexported, so it's
// read the way class_generic_calls.go reads a signature's mapper: by name, with its existence checked.
// A checker whose collection can't be read is reported as an overflow, which fails closed.
func (p *Program) RelationOverflow(typeChecker *checker.Checker) string {
	field := reflect.ValueOf(typeChecker).Elem().FieldByName("diagnostics")
	if !field.IsValid() || field.Type() != reflect.TypeOf(ast.DiagnosticsCollection{}) {
		return "the checker's diagnostics collection isn't readable, so an overflowed relation can't be ruled out"
	}
	all := (*ast.DiagnosticsCollection)(field.Addr().UnsafePointer()).GetDiagnostics()
	var overflows []string
	for _, diagnostic := range all {
		if diagnostic.Code() == relationOverflowCode {
			overflows = append(overflows, p.formatDiagnostic(diagnostic))
		}
	}
	if len(overflows) == 0 {
		return ""
	}
	sort.Strings(overflows)
	return overflows[0]
}
