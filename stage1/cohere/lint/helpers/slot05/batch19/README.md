# Control flow statement helpers

One public helper per .a file: whileStatement, doStatement and labeledStatement. Generic node and block handles retain identity without requiring a copied AST or owning cyclic graph pointers. Pass a valid typed parser projection for the handed node. FlowState owns the mutable current block and jump stack; FlowOps supplies the existing graph operations, parser predicates and body traversal. Enter must update state.cur. A missing loop false destination is undefined, matching Go nil. Operations are synchronous and observe the live current block after callbacks return.

The while helper allocates its test, after and body in Go order, routes constant truthy tests without a false destination, and continues at the test. The do helper evaluates its body before its test, continues at the test, and leaves the jump's loop hook absent so continue does not emit an iteration. The label helper delegates loops and switches; other labeled statements get a nonbreakable label target and an after block, preserving any existing jump prefix.

Truthiness, parser projection, label lookup, block allocation/link/enter, jump push/pop, condition/expression/body processing and hooks remain external dependencies. The private test adapter executes those dependencies in pinned Go and exports their observed state changes. It does not port or claim their implementations. No rule or shared harness is edited.

Run from the repository root with the setup environment sourced:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch19 -count=1 -v -timeout=20m > /tmp/lint05-batch19-helpers.log 2>&1
```

See REPORT.md for all comparisons, mutants, consumers and coverage limits.
