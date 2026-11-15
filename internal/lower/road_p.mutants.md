# Road P: mutants

Road P (#p2n17c4, rulings.md:117 applied to the right of an &&) admits a presence test on the right of an && that is
tested for truthiness alone, recursively (truthinessOnly in method_presence.go), and lowers a method's test there to a
boolean (methodPresenceAndRight). Each mutant is one edit, applied alone, then reverted. Run
`go test ./internal/lower -run '^TestRoadP' -count=1 -v` and
`go test ./internal/oracle -run '^TestRoadPAgreesWithNode$' -count=1 -v`. None has been run yet.

## 1. The recursion taken back to one level

File: internal/lower/method_presence.go, truthinessOnly.

Old:

```go
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && truthinessOnly(parent)
```

New:

```go
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.Parent != nil && parent.Parent.Kind == ast.KindIfStatement
```

Expected: TestRoadPRightOfAndLowers' "nested on the right", "nested on the left", "!", "the test of ?:" and "a while"
rows stop in both forms, as a method read as a value. Only an && directly in an if still admits its right.

## 2. Any && operand taken as tested alone

File: internal/lower/method_presence.go, truthinessOnly.

Old:

```go
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && truthinessOnly(parent)
```

New:

```go
		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken
```

Expected: TestRoadPRightOfAndKeepingTheValueIsRefused's initializer, argument, return, typeof and value-left rows
aren't refused as a method read as a value.

## 3. The method's test on the right answering false

File: internal/lower/method_presence.go, methodPresenceAndRight.

Old:

```go
		present = read.then(ir.BooleanConstant{Value: true})
```

New:

```go
		present = read.then(ir.BooleanConstant{Value: false})
```

Expected: TestRoadPAgreesWithNode goes red on every backend: the class's column reads ---- true- where Node prints
abcd falsef and a--- falsef.

## 4. presenceConsumer widened to every right of &&

File: internal/lower/method_presence.go, presenceConsumer. methodPresenceAndRight asks truthinessOnly of the && itself
(Lower's should), so a position added to presenceConsumer can't make an && whose value is read lower to a boolean.

Old:

```go
		if parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Left == node {
```

New:

```go
		if parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken {
```

Expected: TestRoadPRightOfAndKeepingTheValueIsRefused's initializer, argument and return rows, in the class form, stop
as not yet, "a method's presence test on the right of an && whose value is read", where without the guard they would
lower the && to a boolean its reader holds. The rows go red either way, since they want a refusal; the stop's words say
the guard is what held.
