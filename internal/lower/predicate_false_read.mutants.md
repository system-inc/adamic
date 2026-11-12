# Mutants for a checked predicate's false side checked only where a read observes it (#39wtqdj)

Each is one edit, run on a house box, then put back.

| Mutant | Edit | Red |
|---|---|---|
| The condition-only arm restored | predicates_checked.go, `checkedPredicateDirections`: main's visitor put back between the read walks and its last `return directions`, marking the false direction for every condition whose false side narrows the argument, read or not (the block below) | measured on the head's proof |
| Types decide alone | predicates_checked.go, `checkedPredicateDirections`: `\|\| l.predicateFalseRead(call, argument)` removed, so only predicateUseDirections' with-and-without type test finds a false read | measured on the head's proof |
| A merge counts | predicates_checked.go, `predicateFalseReaches`: `return falseMet && (!trueMet \|\| reduced)` made `return falseMet`, so a read both sides reach counts as the false side's | measured on the head's proof |
| The shorthand read dropped | predicates_proof.go, `predicateNameReads`: `if parent == nil || parent.Kind != ast.KindShorthandPropertyAssignment || argument.Kind != ast.KindIdentifier {` made `if true {`, so a shorthand `{ items }` reads nothing | TestPredicateFalseRead/shorthand: some's false direction unobservable, want checked |
| Holders read | predicates_proof.go, `predicateNameReads`: `return predicateReadPosition(node) && l.predicateSameReference(node, argument)` made `return l.predicateSameReference(node, argument)`, so an assignment or a declaration the checker matches against the reference, which carries no flow data, counts as a read the graph doesn't place | measured on the head's proof |
| The name after a dot read | predicates_proof.go, `predicateReadPosition`: `return ast.IsExpressionNode(node) && !ast.IsRightSideOfPropertyAccess(node)` made `return ast.IsExpressionNode(node)`, so the xs in holder.xs reads the local xs | measured on the head's proof |
| Labels read | predicates_proof.go, `predicateReadPosition`: `return ast.IsExpressionNode(node) && !ast.IsRightSideOfPropertyAccess(node)` made `return !ast.IsRightSideOfPropertyAccess(node)`, so the label in `items: for (;;) { break items; }` reads the local items | measured on the head's proof |

The condition-only arm, inserted before `checkedPredicateDirections`' last `return directions`:

```go
	conditionNarrows := func(condition *ast.Node, truth bool) bool {
		if condition == nil || l.predicateConditionDirection(condition, call.AsNode(), truth, 0)&predicateFalse == 0 {
			return false
		}
		flags := ast.FlowFlagsFalseCondition
		if truth {
			flags = ast.FlowFlagsTrueCondition
		}
		branch := &ast.FlowNode{Flags: flags, Node: condition, Antecedent: &ast.FlowNode{Flags: ast.FlowFlagsStart}}
		narrowed := predicateFlowType(l.checker, argument, source, source, nil, branch)
		return !checker.Checker_isTypeIdenticalTo(l.checker, source, narrowed)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		var condition *ast.Node
		switch node.Kind {
		case ast.KindIfStatement:
			condition = node.AsIfStatement().Expression
		case ast.KindConditionalExpression:
			condition = node.AsConditionalExpression().Condition
		case ast.KindWhileStatement:
			condition = node.AsWhileStatement().Expression
		case ast.KindDoStatement:
			condition = node.AsDoStatement().Expression
		case ast.KindForStatement:
			condition = node.AsForStatement().Condition
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			if binary.OperatorToken.Kind == ast.KindBarBarToken && conditionNarrows(binary.Left, false) {
				directions |= predicateFalse
			}
		case ast.KindSwitchStatement:
			statement := node.AsSwitchStatement()
			if ast.SkipParentheses(statement.Expression).Kind == ast.KindTrueKeyword {
				for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
					if clause.Kind == ast.KindCaseClause && conditionNarrows(clause.Expression(), false) {
						directions |= predicateFalse
					}
				}
			}
		}
		if conditionNarrows(condition, true) || conditionNarrows(condition, false) {
			directions |= predicateFalse
		}
		node.ForEachChild(visit)
		return false
	}
	ast.GetSourceFileOfNode(call.AsNode()).AsNode().ForEachChild(visit)
```
