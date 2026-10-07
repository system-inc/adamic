# Unknown nodes, optional chains and hook calls

Three helpers in separate .a files. visitUnknown guards nil, then routes a statement to statement processing and returns; every other present node goes to expression processing. Nil detection and classification are explicit opaque-node adapter dependencies. forkOptionalChain short-circuits an empty join stack before querying the optional-chain-root predicate, selects the last join, links the current block to it, allocates a fresh block, links and enters it. Reachability alone does not suppress this helper. It leaves the join stack intact. isHookCall rejects an absent call or callee; identifiers go to text and IsHookName, property accesses go to IsNamespacedMember with the same predicate, and other expression kinds return false.

Parser classification/projection, nil detection, statement and expression processing, optional-root detection, graph transitions, name predicates and namespace detection remain external dependencies. These files port composition, callback order and argument/state identity rather than duplicating those separately owned helpers. No rule or matcher is added. Caller classifiers use the actual typed parser adapter; there is no node-kind string dispatch here.

Private overlays change only dependency call names inside the actual pinned Go methods. Wrappers execute the real dependency and retain only the direct callback trace when nested processing invokes instrumented methods. Production cohere and the shared harness are unchanged. Pin: 715ba94f3608a6500086b1076ce5cb7e51b836db. [NOTICE.md](NOTICE.md) retains the CFG MIT attribution; the hook-call method is Go cohere's own implementation.

visitUnknown uses all 12,452 nodes (including nil) from 801 distinct consumer/control strings. Optional chains use 6,472 states from every parser optional-chain root and one source-file negative per string, stacks of zero through three joins and both current reachability values. Node and block identity is compared through stable arena IDs; the graph callbacks execute real Go transitions. isHookCall uses 36 calls, including absent call/callee, from 355 strings inspected across all four consumers. Controls include direct, namespaced, non-React, computed, parenthesized and optional calls, lowercase/digit failures and Unicode hook names. Name and namespace verdicts are supplied from the real Go leaf dependencies; this is not a fresh implementation of those leaves.

All 18,960 rows agree with Go on source Node, emitted JavaScript and sanitized native. Each mutant must compile and exit successfully with empty stderr before its wrong output is credited. The suite uses sixteen semantic mutants. Coverage and corpus hashes are archived in evidence. Strings extracted from test files include configuration and expected text; this is helper coverage, not complete rule finding/fix replay.

Inputs assume valid immutable parser nodes, dense immutable join lists, distinct block values and a present current block. Invalid ASTs, sparse arrays, arbitrary dependency implementations or callbacks that mutate those projections are not covered. Missing join invariant failures are defensive adapter checks, not upstream successful behavior. The full repository gate and its seventeen required external-input correctness checks were not run in this bounded helper unit; no check was skipped, relaxed or deleted.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH17_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch17/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch17 -count=1 -v -timeout=20m > /tmp/lint05-batch17-final.log 2>&1
```

[REPORT.md](REPORT.md) records commands and every mutation. [CONSUMERS.md](CONSUMERS.md) lists each helper's four consumers.
