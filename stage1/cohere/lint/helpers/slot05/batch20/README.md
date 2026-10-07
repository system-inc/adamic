# For loops and optional access

One public helper per .a file: forStatement, forInOfStatement and accessOrCall. Pass the typed projection of the handed node, its live FlowState and synchronous FlowOps callbacks. Nodes and blocks remain opaque generic handles. isNil supplies the adapter's nil representation; an absent condition destination and absent jump hook are undefined. Enter and enterDisconnected update state.cur. Projections expose immutable argument and declaration-name lists, normalizing absent parser lists to empty lists.

forStatement preserves initializer routing, the omitted-condition-and-incrementor self-loop, constant-truthy routing, disconnected incrementor traversal before the body and the correct continue target. forInOfStatement evaluates the iterable before creating its head, binds the iteration target after entering the body, and preserves declaration-name order and loop destinations. accessOrCall guards outermost optional-chain allocation, traverses the receiver and computed key or arguments in Go order, visits type arguments, invokes the throwable fork, then pops and enters exactly its own join. Existing jump and chain prefixes remain owned by the caller.

Parser projection, nil/classification/truthiness, graph operations, jump push/pop, expression/body/condition traversal, declaration names, type arguments and fork implementations remain external dependencies. The private oracle executes the real Go dependencies and exports their graph effects. It proves those dependency calls preserve the incoming chain stack before replaying their effects without replacing the port's own stack. No rule, shared harness, compiler or regex matcher is changed.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch20 -count=1 -v -timeout=20m > /tmp/lint05-batch20-helpers.log 2>&1
```

See REPORT.md for actual comparisons, every mutant, consumers and coverage limits.
