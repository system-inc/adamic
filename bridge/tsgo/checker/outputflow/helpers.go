package outputflow

import (
	"math/big"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func isBreakableStatement(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindSwitchStatement, ast.KindWhileStatement, ast.KindDoStatement,
		ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}

// labelsOf collects every label wrapped directly around a loop or switch, so
// `outer: inner: while (...)` resolves `continue outer` and `break outer` as
// well as `inner`.
//
// Keeping the outer label is a known deviation. ESLint attaches only the
// innermost one, so a `continue` naming an outer label finds no loop there and
// its back edge is dropped. It surfaces where a loop's exit hangs off that edge
// - a `do...while`, whose test runs only from it - leaving what follows the loop
// reachable here and unreachable in ESLint, which `no-useless-return`,
// `no-unreachable-loop` and `no-useless-assignment` all read.
func labelsOf(node *ast.Node) []string {
	var labels []string
	current := node
	for parent := current.Parent; parent != nil && parent.Kind == ast.KindLabeledStatement; parent = parent.Parent {
		labeled := parent.AsLabeledStatement()
		if labeled.Statement != current {
			break
		}
		labels = append(labels, labeled.Label.Text())
		current = parent
	}
	return labels
}

// isAlwaysTruthyTest mirrors ESLint's `getBooleanValueIfSimpleConstant`: only a
// literal test can mark the loop infinite, so the code after it is unreachable.
func isAlwaysTruthyTest(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindRegularExpressionLiteral:
		return true
	case ast.KindNumericLiteral:
		// Upstream runs its own `NormalizeNumericLiteral` over this text before comparing. On our
		// substrate that normalizer is a no-op, because typescript-go's scanner has already stored
		// the canonical decimal value rather than the source spelling. Probed on our parser rather
		// than assumed: `0x0`, `0.0`, `0e10`, `0b0`, `0o0`, `0_0`, `00`, and `.0` all arrive as
		// exactly "0", and `0x1`, `1e2`, `1_0`, `1e400` arrive as "1", "100", "10", "Infinity".
		// `internal/rules/core/no_compare_neg_zero.go` reached the same conclusion independently
		// and reads `.Text == "0"` the same way.
		return node.AsNumericLiteral().Text != "0"
	case ast.KindBigIntLiteral:
		// A BigInt literal is NOT normalized the way a numeric one is, which is the asymmetry that
		// makes this branch load-bearing rather than redundant. Probed on our parser: `0n` arrives
		// as "0n" and `0x0n` as "0x0n" - the `n` suffix survives, and a hex or binary spelling
		// survives with it, though `0b0n` and `0o0n` do arrive as "0n". So a bare `.Text != "0"`
		// would call every zero BigInt truthy and mark the code after `while (0n)` unreachable.
		return normalizeBigIntLiteral(node.AsBigIntLiteral().Text) != "0"
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text != ""
	}
	return false
}

// isThrowableIdentifier reports whether completing this identifier is one of
// the points ESLint treats as able to throw inside a `try` block. Names that
// only declare or label something are excluded, as are JSX names, which ESLint
// sees as a separate node type.
func isThrowableIdentifier(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if ast.IsJsxTagName(node) {
		return false
	}
	switch parent.Kind {
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement,
		ast.KindArrayBindingPattern, ast.KindImportSpecifier, ast.KindExportSpecifier,
		ast.KindImportClause, ast.KindNamespaceImport, ast.KindNamespaceExport,
		ast.KindJsxAttribute, ast.KindJsxNamespacedName:
		return false
	case ast.KindBindingElement:
		binding := parent.AsBindingElement()
		if binding.DotDotDotToken != nil || binding.PropertyName == node {
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectBindingPattern
	case ast.KindCatchClause:
		return false
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindClassDeclaration, ast.KindClassExpression, ast.KindVariableDeclaration,
		ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindEnumDeclaration,
		ast.KindModuleDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return parent.Name() != node
	}
	return true
}

// normalizeBigIntLiteral writes a BigInt literal's text as its decimal digits, which is what
// JavaScript's `String(value)` produces and what ESLint compares against.
//
// Inlined from rslint's `utils.NormalizeBigIntLiteral` rather than imported, because that package
// is the linter's whole shared shelf and this graph needs two lines of it. The numeric sibling was
// dropped entirely - see `isAlwaysTruthyTest` for the measurement that says why one survived and
// one did not.
//
// `big.Int` rather than `strconv.ParseInt` because BigInt is arbitrary precision by definition and
// routinely exceeds int64: `0x10000000000000000n` is 2^64 and reaches this function unnormalized.
func normalizeBigIntLiteral(text string) string {
	digits := strings.TrimSuffix(text, "n")
	if value, ok := new(big.Int).SetString(digits, 0); ok {
		return value.String()
	}
	return digits
}
