# Mutant for a namespace export replaced by any assignment target

internal/lower/namespaces.go, `namespaceRefusal`. Old:

```go
		if writtenTarget(node) {
```

New (main's condition before this change):

```go
		if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == node && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) {
```

Run: `go test ./internal/lower -count=1 -run 'TestNamespaceLimitsStayLoud' -v`

Red: MEASURED

A second mutant, the wrappers left out of writtenTarget (written_target.go): its `case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:` made `case ast.KindParenthesizedExpression, ast.KindNonNullExpression:`.

Red: MEASURED2
