package load

import (
	"context"
	"sort"

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
func (p *Program) RelationOverflow(ctx context.Context, typeChecker *checker.Checker) string {
	// The program's own files were checked at load, so each file's diagnostics are a lookup.
	all := typeChecker.GetGlobalDiagnostics()
	for _, sourceFile := range p.Files() {
		all = append(all, typeChecker.GetDiagnostics(ctx, sourceFile)...)
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
