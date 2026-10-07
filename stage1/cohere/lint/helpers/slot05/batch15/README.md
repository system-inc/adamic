# Statement lists and continue routing

Two helpers in separate .a files, completing the three-helper batch alongside ../batch14/make_break.a. Both consume opaque caller-supplied node values. makeContinue uses the typed generic JumpState from the preceding owned batch. No rule is registered or implemented here.

statements forwards every item of an immutable parser node list, in order, to its statement callback. An absent list is a no-op; nil node entries are forwarded as values. The callback retains the real builder through its closure. This function does no node-kind dispatch. The private Go wrapper executes the real statement processor and records only calls at this helper's outer dependency seam; nested statement processing is still real Go, but its internal calls belong to the explicit callback dependency.

makeContinue returns immediately on an unreachable current block. It reads an optional label once, skips nil continue destinations, resolves the innermost exact label (or first unlabelled candidate), calls loop then link, and finally calls makeUnreachable. Even a reachable unmatched continue ends the current reachable segment. Nil loop nodes are forwarded to the method dependency, which separately decides whether to invoke a loop hook. No broken flag is changed, and break destinations are retained unchanged. AST text, loop processing, link and fresh unreachable graph transitions remain explicit external dependencies. This file implements routing, not those graph operations.

Actual pinned Go methods are instrumented only by changing dependency call names in a temporary overlay. Production cohere is unchanged. Go is pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db. All four consumers' test-file string literals are inspected and parsed by the real typescript-go parser. Strings include source, configuration and expected text. Coverage is helper behavior on those inputs, not full rule findings/fix replay.

statements: 806 nil/empty/nil-item/consumer parser lists, compared to real Go outer statement calls. Arena IDs represent exact node identity, including nil (-1) and repeats. makeContinue: 27,648 state/label/target queries over depths zero through five, breakability and existing broken flags, alternating/all nil continue destinations, distinct break and continue blocks, nil/present loop nodes, and both current reachability values. Returned Go graph effects are supplied as external dependency results in the .a driver; the helper's own call order, arguments, target identity and broken state are compared byte for byte.

Supported input assumes immutable dense parser lists, a present current block and contiguous distinct jump records. Arbitrary callbacks that mutate parser lists, repeated aliases of one mutable jump record, invalid UTF-8/lone surrogate labels and arbitrary invalid AST/state representations are not covered. External statement/graph implementation parity and whole-rule diagnostics remain separate work.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH15_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch15/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch15 -count=1 -v -timeout=20m > /tmp/lint05-batch15-complete.log 2>&1
```

Upstream provenance and MIT attribution: [NOTICE.md](../batch14/NOTICE.md).
