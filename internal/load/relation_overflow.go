package load

import (
	"reflect"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// relationOverflowCode is TS2859, the checker's Excessive_complexity_comparing_types_0_and_1.
const relationOverflowCode = 2859

// relationOverflowMark is RelationComparisonResultComplexityOverflow (relater.go:72), the mark the
// checker caches on a pair whose comparison spent its budget, whether or not its diagnostic was kept.
const relationOverflowMark = 1 << 5

// relationTables are the checker's relation caches: the five relations and the enum relation.
var relationTables = []string{"subtypeRelation", "strictSubtypeRelation", "assignableRelation", "comparableRelation", "identityRelation"}

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
	if RelationOverflowMarked(typeChecker) {
		// addDiagnostic drops a diagnostic made while TypeToString is at its deepest nesting
		// (checker.go:14261), and node building runs relations, so a dropped TS2859 is read from the
		// relation caches' marks instead (Compiler's condition (b) on #apyq5ve).
		for _, diagnostic := range all {
			if diagnostic.Code() == relationOverflowCode {
				return p.formatDiagnostic(diagnostic)
			}
		}
		return "a type relation the checker marked as overflowed, whose diagnostic it dropped"
	}
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

// RelationOverflowMarked is whether any of the checker's relation caches holds a pair marked as having
// spent its budget. The caches are unexported, so they're read by name as the diagnostics are, and a
// cache that isn't where it's expected counts as marked, which fails closed.
func RelationOverflowMarked(typeChecker *checker.Checker) bool {
	value := reflect.ValueOf(typeChecker).Elem()
	marked := func(results reflect.Value) bool {
		if !results.IsValid() || results.Kind() != reflect.Map {
			return true
		}
		switch results.Type().Elem().Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		default:
			return true // A result that isn't a flag word can't be read, so it counts as marked.
		}
		for iterator := results.MapRange(); iterator.Next(); {
			if iterator.Value().Uint()&relationOverflowMark != 0 {
				return true
			}
		}
		return false
	}
	for _, name := range relationTables {
		relation := value.FieldByName(name)
		if !relation.IsValid() || relation.Kind() != reflect.Pointer {
			return true
		}
		if relation.IsNil() {
			continue
		}
		if marked(relation.Elem().FieldByName("results")) {
			return true
		}
	}
	return marked(value.FieldByName("enumRelation"))
}
