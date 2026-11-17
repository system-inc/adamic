# getParseTreeNode's overload 2, served per call by the node-test proof: mutants (#0kkpw2e set 3)

Each mutant is one exact edit on this head, applied alone, run, then restored from `HEAD`. The pins are
internal/lower/overload_parse_tree_node_test.go: `TestOverloadParseTreeNodeTestedCallsLower` (tsc's shape, calls
with a named predicate erased and a one-argument call on overload 1, agreeing with Node natively and on the
JavaScript backend), `TestOverloadParseTreeNodeAtItsConstraintLowers` (rule 2), `TestOverloadParseTreeNodeBareClaimIsChecked`
and `TestOverloadParseTreeNodeBareClaimLieStopsAtTheCall` (rule 3, checkable), `TestOverloadParseTreeNodeBareClaimIsRefused`
(rule 3, uncheckable in .ts and any bare claim in .a), `TestOverloadParseTreeNodeUndefinedTestStops`,
`TestOverloadParseTreeNodeUnprovenBodiesStayRefused` and `TestOverloadParseTreeNodeCommaCalleeIsNotYet` (rule 4).

## 1. Rule 3's Refused dropped: a bare claim with no check is erased

File: internal/lower/overload_parse_tree_node.go

Old:

```go
	return &Refused{Where: l.program.Where(node), What: "a call to an overload whose result is honest only for a proven node test, made without one", Fix: "pass the node test that proves the result's type"}
```

New:

```go
	return nil
```

Expected: `TestOverloadParseTreeNodeBareClaimIsRefused/uncheckable_instance` and `/checkable_instance_in_.a` red: neither
bare claim is refused at its call, so each lowers with its result erased or stops later on some other road
(the result's representation at censusOverloadFit, a guess). The tested calls, the constraint row, the checked
rows (their road is overloadNodeTestChecked's, which the edit leaves alone), the undefined test, the unproven
bodies and the comma row stay green.

## 2. Proof (a) accepts a parameter returned after a write: the write test dropped

File: internal/lower/overload_parse_tree_node.go

Old:

```go
			if writtenAt[index] <= position {
				return false // Written first, it's no longer the argument the overload typed.
			}
```

New:

```go
```

Expected: `TestOverloadParseTreeNodeUnprovenBodiesStayRefused/returns_the_parameter_reassigned` and
`/shorthand_write_before_the_return` red: each `return node` follows a write of node in an earlier top-level
statement (`node = node.original`, `({ node } = other)`), and with nothing refusing it the census accepts overload 2
by the proof, so the row no longer stops at its declaration. The returns-original row stays green
(`node.original` is none of the proof's shapes), and so does the closure row (proof (b) refuses it by
nestedWrites) and every other pin: tsc's own `return node` comes before its first write.

## 3. The shorthand arm dropped from referencedSymbol: a shorthand write read as the property

File: internal/lower/overload_certificate_array.go

Old:

```go
func (l *lowering) referencedSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		return l.checker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	}
	return l.symbol(identifier)
}
```

New:

```go
func (l *lowering) referencedSymbol(identifier *ast.Node) *ast.Symbol {
	return l.symbol(identifier)
}
```

Expected: `TestOverloadParseTreeNodeUnprovenBodiesStayRefused/shorthand_write_before_the_return` and
`/shorthand_write_the_test_can_run` red: `({ node } = other)`'s name resolves to the object literal's property, so
writtenAt never counts the write before the bare return, and nestedWrites never holds node for the closure the
test can run (main's predicateOverloadBinding resolves it the same way and misses it too), and the census accepts
each body. Certificate A's three shorthand rows, `TestOverloadCertificateArrayNegativesStayRefused/shorthand_alias_over_a_parameter`,
`/spread_shorthand_alias_over_a_parameter` and `/mapper_written_through_a_destructuring_assignment`, red with them,
as its own mutant 3 says. Every other row stays green.
