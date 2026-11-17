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

Expected: `TestOverloadParseTreeNodeUnprovenBodiesStayRefused/returns_the_parameter_reassigned` reds: its
`return node` follows `node = node.original` in an earlier top-level statement, and with nothing refusing it the
census accepts overload 2 by the proof, so the row no longer stops at its declaration. The returns-original row
stays green (`node.original` is none of the proof's shapes), and so does every other pin: tsc's own `return node`
comes before its first write.
