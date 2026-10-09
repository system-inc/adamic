// Package module answers questions about how a file exposes its declarations.
//
// It sits under `internal/utilities/ecmascript/` because a default export is an ECMAScript module
// concept rather than a React or JSX one, and because that prefix is already on the leaf guard's
// allowlist.
package module

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// IsDefaultExported reports whether a declaration carries both the export and default modifiers.
//
// This is lifted on a stronger argument than caller count. Two implementations of it already existed
// in one package, byte-identical in logic under two names: `isDefaultExported` in
// `react_component_no_multiple_primary.go` and `hasDefaultModifier` in
// `next_require_page_default_export.go`. Neither file could see the other, so the second author
// wrote it again rather than found it.
//
// **Two copies that agree today are the state just before drift, not evidence that drift is not a
// risk.** They are free to diverge on the first case nobody has thought about yet, and nothing would
// report it: each rule's fixtures would keep passing against its own copy.
//
// A research pass on `no-page-custom-font` then found a third caller wanting the same predicate from
// `internal/rules/next/`, which is what turned a tidy-up into a lift.
//
// Both modifiers are required rather than either. `export function run() {}` is exported and not
// default; `export default function run() {}` is both.
func IsDefaultExported(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return HasDefaultModifier(node.Modifiers())
}

// HasDefaultModifier is IsDefaultExported for a caller that already holds the modifier list.
//
// Both spellings survive the lift because both call shapes existed in the rules it came from, and
// collapsing them would push one call site into reaching for `node.Modifiers()` on a value it does
// not hold.
func HasDefaultModifier(modifiers *ast.ModifierList) bool {
	if modifiers == nil {
		return false
	}
	sawExport := false
	sawDefault := false
	for _, modifier := range modifiers.Nodes {
		switch modifier.Kind {
		case ast.KindExportKeyword:
			sawExport = true
		case ast.KindDefaultKeyword:
			sawDefault = true
		}
	}
	return sawExport && sawDefault
}

// IsExported reports whether a declaration carries the export keyword, default or not.
//
// Three copies of this existed under three names before the lift: `isExportedDeclaration` in
// `nexus/consistency_no_screaming_snake_case.go`, `isExportedStatement` in
// `structure/react_component_no_multiple_primary.go`, and a third shape below. Two packages, six
// call sites, and no file able to see another's.
//
// The two that agreed are collapsed here. The third is not, and the difference is the reason this
// file has two functions rather than one.
func IsExported(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return HasExportModifier(node.Modifiers())
}

// HasExportModifier is IsExported for a caller holding the modifier list.
func HasExportModifier(modifiers *ast.ModifierList) bool {
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindExportKeyword {
			return true
		}
	}
	return false
}

// IsExportedByName reports an export that is not a default export.
//
// **This is a different question from IsExported and the names are one word apart, which is exactly
// why both belong on a shelf rather than in three rule files.** `export default function Thing()`
// carries both keywords in one modifier list, so a rule asking about named exports has to exclude
// it, and a rule asking about exports at all must not.
//
// The shape it exists for: `react-component-require-named-export` flags a component exported only as
// a default, so reading `export default` as a named export would silence the rule on precisely the
// code it exists to catch.
func IsExportedByName(modifiers *ast.ModifierList) bool {
	if modifiers == nil {
		return false
	}
	hasExport := false
	hasDefault := false
	for _, modifier := range modifiers.Nodes {
		switch modifier.Kind {
		case ast.KindExportKeyword:
			hasExport = true
		case ast.KindDefaultKeyword:
			hasDefault = true
		}
	}
	return hasExport && !hasDefault
}

// PositionAfterModifiers returns the position a finding about a declaration should be scanned from,
// skipping any modifiers that precede the declaration keyword.
//
// TypeScript and ESTree disagree about where `export` lives, and that disagreement is a span defect
// waiting for anyone porting a rule from ESLint. In ESTree an exported function is an
// `ExportNamedDeclaration` **wrapping** a `FunctionDeclaration`, so a rule reporting the inner
// declaration points at `function`. In TypeScript there is no wrapper: `export` is a modifier on the
// declaration itself and is inside its `Pos()`, so the same rule reporting the same conceptual node
// points at `export`, eight columns to the left.
//
// Measured rather than reasoned: `react-component-no-multiple-primary` produced 128 findings on our
// own tree, matching oxlint's count exactly, and **29 of them pointed at the wrong column**. Every
// one was an exported function component, every one was column 1 against oxlint's column 8. The
// counts agreeing is what makes this expensive to notice, since a rule can be right about every file
// and every line while being wrong about where it points, and no count-based check can see it.
//
// This returns a **position to scan from**, not a range, because the caller still has to trim
// leading trivia the way `rule.TokenRange` does. Returning a raw range here skipped that trimming and
// moved all 128 findings onto the line of the preceding comment, which is a worse defect than the one
// being fixed and is why this is shaped as a position.
//
// A declaration carrying no modifiers yields its own `Pos()`, so a caller can reach for this
// unconditionally rather than branching on whether an export is present.
func PositionAfterModifiers(node *ast.Node) int {
	if node == nil {
		return 0
	}
	modifiers := node.Modifiers()
	if modifiers == nil || len(modifiers.Nodes) == 0 {
		return node.Pos()
	}
	last := modifiers.Nodes[len(modifiers.Nodes)-1]
	if last == nil || last.End() >= node.End() {
		return node.Pos()
	}
	return last.End()
}
