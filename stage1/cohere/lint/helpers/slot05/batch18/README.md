# Assignment targets and strict React factories

Three helpers, one production .a file each. patternReads guards nil, unwraps parenthesized, non-null, as, satisfies and type-assertion expressions, reads an identifier before testing whether it can throw and invoking the throwable-fork dependency, and delegates every other node to expression processing. patternWrites guards nil, unwraps the same five forms, writes an identifier and ignores other targets. The ports iterate through transparent wrappers rather than recursing, preserving all direct callbacks and argument identities. Immutable valid parser trees are acyclic.

isEs5ComponentCallStrict first delegates to the separately owned wider ES5 predicate, then skips parentheses around the callee. An identifier must name exactly createReactClass; a property access must have a present identifier name with that same exact text. The initial dependency handles namespace and call validity; the narrower local name check deliberately rejects createClass and React.createClass. No rule, matcher or finding-position code is added.

Parser nil/projection/classification, identifier read/write hooks, throwable classification and fork processing, expression processing, initial ES5 detection and parenthesis skipping remain explicit adapter dependencies. Actual Go dependency executions supply their independent outcomes. This unit ports their composition and observes callback order and argument identity, not their separately owned implementation internals. The flat arena uses opaque node values without object ownership cycles.

The private overlays rename dependency calls only inside the actual pinned Go methods. read/write/fork wrappers execute real builder methods; expression wrappers execute real expression processing but retain only the outer dependency trace. The Go switch discriminator is wrapped to observe one projection per present node, including each transparent wrapper. Production cohere and the shared harness are unchanged. Go pin: 715ba94f3608a6500086b1076ce5cb7e51b836db. [NOTICE.md](NOTICE.md) retains the CFG MIT attribution. The strict React helper is cohere's own implementation.

All test-file string literals are inspected for the four consumers of each helper. They include source, options and expected text. Added controls cover all five wrapper kinds in a TS-parsed string, nested wrappers, member targets, contextual identifiers and both React factory spellings with parenthesized/namespace/computed/case variants. The CFG corpus has 801 distinct strings and 12,508 nodes including nil, used by both walkers; 3,570 read cases invoke the throwable fork. The strict corpus has 630 strings and 9,812 nodes including nil, with 51 positive verdicts. Total: 34,828 case rows. Coverage JSON and corpus hashes are archived.

Go callbacks and verdicts must agree byte-for-byte with source Node, emitted JavaScript and sanitized native. Mutants must compile and exit successfully with empty stderr before their differing output is credited. All eighteen compiling semantic variants pass that requirement and are caught by the independent Go comparison.

Coverage assumes valid immutable parser nodes, faithful adapters and ordinary acyclic wrapper chains. Malformed/forged ASTs, arbitrary external dependency implementations, stack-exhaustion behavior on artificially deep recursive Go inputs and full rule findings/fixes are not covered. Some defensive strict-helper branches are unreachable when the real initial ES5 predicate returns true; arbitrary inconsistent predicate/projection adapters are outside the corpus. This is a bounded helper gate, not the full repository gate or its seventeen required external-input correctness checks. No skipped check is treated as a pass.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH18_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch18/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch18 -count=1 -v -timeout=20m > /tmp/lint05-batch18-final.log 2>&1
```

[REPORT.md](REPORT.md) records commands, witnesses and final landing validation. [CONSUMERS.md](CONSUMERS.md) names each helper's four rules.
