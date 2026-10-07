Delivered control_flow_decorators.a: ordered generic expression delegation, removing one prerequisite from four CFG rules, zero final blockers.
Claim 8256390ae pushed before code; current main b8fb957aa, sibling rule branch c485ad2cb remains pushed and parked.
Complete owned suite PASS 123.469s, vet clean; 29,662 new observations / 59,369 bytes identical on actual Go, source Node, emitted JS and ASan/UBSan native.
Omitting the first decorator and duplicating each expression both compile/run on every backend; only Go comparison catches row 29562.
Full expression walker, graph/finding parity, nil node and concurrent oracle instrumentation are outside this helper; original fixtures delegate zero decorator expressions.

`decorators<E>(node: DecoratedNode<E>, visitExpression: (expression: E) => void)` calls the supplied visitor once per decorator expression in parser list order. It does not inspect relevance kinds or fetch AST nodes. The caller owns the expression walker and its graph state. This is the complete private helper's delegation contract, not a replacement for Builder.expr or the full CFG builder.

Actual Go's method is unchanged. An oracle-only build overlay renames its expression walker entry and wraps it to record direct calls, while invoking the original walker for every call. A depth counter excludes recursive expression calls from the delegation trace. The adapter projects only decorator expression pointer identities, never expected traversal output. All 2,119 captured original input fixtures are parsed; every AST node is offered to the real helper. The original four families produce 29,560 AST nodes and zero direct decorator expression calls: array-callback-return 3,323 nodes, consistent-return 887, no-unreachable-loop 20,964, react-hooks/rules-of-hooks 4,386. Eight explicit decorator controls add 102 nodes and 19 direct calls. These include class/member/parameter decorators, repeated names, Unicode, nested call/short-circuit expressions, computed keys and malformed syntax. This boundary is an observation, not inferred production coverage.

The four consuming rules each lose a sixth dependency delivered by this unit: array-callback-return, consistent-return, no-unreachable-loop, react-hooks/rules-of-hooks. Other shared prerequisites remain. No shared harness, compiler, registry or oracle file is edited.

Commands (environment sourced from /workspace/adamic-tools/env.sh):
```
go test ./stage1/cohere/lint/helpers/wave12 -count=1 -v -timeout=15m > /tmp/wave12-decorators-full.log 2>&1
go vet ./stage1/cohere/lint/helpers/wave12 > /tmp/wave12-decorators-vet.log 2>&1
```
Setup: Go/clang/Node/submodules ready 0s each, build cache warm 122s, total 122s; nproc 5, quota four cores. Evidence logs are in evidence/decorators-*.log. The eight delivered helpers and all 22 existing/new negative controls pass. Earlier controls and their comparison/refusal boundaries remain documented in the earlier reports; the exact counter is still blocked at bigint-return lowering. No full-repository gate or fresh throughput claim is made. docs/parallel-work.md is absent on current main; the existing repository doctrine applies.

Harness 41eb6eab2 is not yet on current origin/main. Its DEDUP_LEDGER.md keeps this unit's ORM and Next document-import copies and retires the three duplicate TypeScript copies to wave1-05/07/08 when rebasing onto the harness. The ledger does not assign a batch-only rule specifically to wave1-12. This helper branch does not merge the harness prematurely.
