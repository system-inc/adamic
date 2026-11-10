# Mutant for a namespace export replaced by any assignment target

internal/lower/namespaces.go, `namespaceRefusal`. Old:

```go
		if ast.IsAssignmentTarget(node) {
```

New (main's condition before this change):

```go
		if parent := node.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().Left == node && ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) {
```

Run: `go test ./internal/lower -count=1 -run 'TestNamespaceLimitsStayLoud' -v`

Red: MEASURED
