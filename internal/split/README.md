# Split analysis prerequisite

The requested exact analysis over ir.Program cannot yet distinguish programs
that require different eligibility decisions. This directory contains executable
witnesses, rather than a classifier that guesses at erased facts.

Run:

    source /workspace/adamic-tools/env.sh
    go test -count=1 -v ./internal/split > /tmp/workers-split-erasure.log 2>&1

TestRequiredSplitFactsAreErased lowers testdata/erased.a and three source
variants through load.LoadOverlay at the same path:

| Variant | Required decision for compute | Observed IR |
| --- | --- | --- |
| Original const scalar global and readonly number array | Eligible | Baseline |
| const limit changed to let limit | Impure: mutable global read | Entire program identical |
| readonly number array changed to mutable number array | Signature not crossable | Entire program identical |
| readonly number array changed to readonly boolean array | Signature not crossable | Functions and locals identical |

The boolean caller's literals differ, so that witness compares functions and
locals instead of the whole program. All four sources independently run through
oracle/node.mjs, emitted JavaScript, and sanitized native code. Each exits
successfully and prints exactly 6 followed by a newline.

The erased facts are visible in internal/ir/ir.go:

- Local carries a runtime Type and Global, but no declaration mutability.
  ConstantClosure only proves const bindings initialized with closure literals.
- Function signatures carry local indices and a runtime return Type.
- Array is one runtime representation for all element types and both mutable
  and readonly arrays.
- Object does not carry a signature's field shape, readonly properties, nesting
  depth, optional fields, or nominal class identity.
- String and other reference representations do not preserve whether a
  signature includes undefined.

internal/lower/locals.go records const only for ConstantClosure.
internal/lower/functions.go records signature representations, while checker
types remain private to lowering. internal/flow cannot recover distinctions
between identical inputs.

A prerequisite unit must preserve declaration mutability, compile-time constant
global proofs, complete parameter and return boundary types, and source
locations/order in IR. Lowering must populate that metadata from the checker.
Then this unit can implement purity propagation, signature validation, recursion
and loop propagation, and crossings without guessing.

Those changes touch internal/ir and lowering, outside this unit's assigned
territory. No production analysis or adamic-split command is supplied.

Validation: the witness comparison was mutated by changing the compiled variant's
return representation to ir.Boolean. All three equality checks failed; restoring
the mutation made the test green. The four requested classifier mutants have
not been run because no classifier exists yet.
