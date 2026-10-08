Built: 30 callable pair certificates, 158 candidate reads, 90 .a fixtures and eight code-frontier controls.
Commits: based on 926a1d39d1a0d6b4cf49bbccf521a4cc02d10f56; checkpoint one a3dd5fd9ceab3ffc45b4d1b2b19c73f382cc3ac7; checkpoint two follows it on codex/views-callables-a.
Checks: source Node, native release/ASan/UBSan, JavaScript, finishing leak checks, original declarations/read spans/aliases/discriminators and share counts pass; whole-table counts retains outside-share failures.
Mutants: all 60 executable arity mutants are caught by pinned exit/message assertions; no compiler/runtime source is mutated.
Uncovered: remaining assigned pairs are retained in queue.json; eight code-needed pairs are listed below; full tsc execution and exact reachability are unmeasured.

Certified ranks: 102, 120, 276, 291, 435, 450, 453, 456, 465, 357, 363, 444, 429, 285, 306, 333, 354, 360, 366, 369, 372, 375, 396, 399, 402, 426, 477, 486, 489, 513.
The ledger assigns 897 remaining pairs / 2742 candidate reads before exclusions.
This share certifies 30 / 158. Overall lane totals become 186 / 2818 pairs and
2857 / 11063 candidate reads, leaving 2632 / 8206 overall. The share still has
762 pending pairs / 1812 reads; eight code frontiers account for 498 reads.
These reduced original-member contract fixtures do not execute every
inventoried original application context. This delivery is incomplete.

The original member declarations and read expressions are unchanged. Nested
context.factory receivers and parenthesizerRules() are retained. Carrier bodies
and implementations are reduced; PropertyName and MemberName aliases and their
numeric discriminators are verified against the independent upstream pin.
TokenFlags is explicitly reduced to number as adjacent enum representation;
this does not certify enum literal membership. Each original Target signature
has a good control, wrong callable kind, and wrong arity. Both negative cases
pin their full field/type/found messages and exit 70 on all three backend modes.
Good native executions pass the existing independent leak checker.

Each mutant changes only the loaded IR signature's expected arity to the wrong
producer's arity. Recorded producer metadata, result contracts, generated calling
convention and source input remain intact. Both release native and JavaScript
then execute with exit 0 and the source Node output. The independent negative
pin expected exit 70 with its exact signature message, so it catches the mutant.
The logs record all 60 observations individually. No invalid C or sanitizer
failure is used as an arity mutant result.

The fixtures and test harness are share-owned. One test-only counts discovery
hook in checked_views_callable_counts_test.go collects share-*/rank-*/*.a so the
other shares can add families without editing that registry. Exactly 90 count
rows are appended; existing rows remain byte-for-byte unchanged. The required
whole-table update was run twice, first exposing missing stage3 API packages.
Installing the existing pinned dependencies removes that setup issue. The second
update still fails outside this share, including inherited overload/template
lowering, MaybeNumber call conversion and graph free failures. It does not write
a partial global table. The dedicated share updater and subsequent row check pass.

Eight Node-valid original reads require code and are not certified:

| Rank | Pair | Candidate reads | Stop and required change |
| --- | --- | ---: | --- |
| 3 | NodeFactory.createIdentifier | 206 | Unsupported callable contract at factory.createIdentifier; demand-time overload selection with complete recorded producer coverage. |
| 9 | NodeFactory.createStringLiteral | 104 | Unsupported callable contract at factory.createStringLiteral; the same overload coverage, including optional argument domains. |
| 12 | NodeFactory.cloneNode | 91 | Unsupported callable contract at factory.cloneNode; sound generic instantiation preserving parameter/result identity. |
| 24 | NodeFactory.createUniqueName | 58 | Unsupported callable contract at factory.createUniqueName; complete overloaded signature and optional-domain coverage. |
| 165 | TransformationContext.onEmitNode | 11 | Unsupported callable contract at context.onEmitNode for malformed producers; preserve higher-order descriptors independently of compatible producer proof so the runtime guard executes. |
| 174 | NodeFactory.createModifier | 10 | Unsupported callable contract at factory.createModifier; sound generic instantiation of ModifierToken<T>. |
| 186 | NodeFactory.createImportClause | 9 | Unsupported callable contract at nodeFactory.createImportClause; certify both boolean and phase-modifier overload domains. |
| 189 | NodeFactory.createYieldExpression | 9 | Unsupported callable contract at factory.createYieldExpression; preserve correlated asterisk/expression overload domains. |

Every stop is a lowering refusal at the original member read, with source Node
exiting 0 (7 for the seven overload/generic controls, done for rank 165).
Rank 165 has a passing valid native/JavaScript control, but its malformed producer
is refused before runtime; it is excluded from certificates and mutant counts. needs-code.json and logs/lane5-a-frontiers-pinned.log
retain exact declarations, spans, hashes, read names and diagnostic observations.
No runtime/leak result is claimed for refused fixtures. Binding ranks 18 and 294,
Set intrinsics and optional-host families are excluded rather than rewritten.
Untested families retain pending status; no code blocker is inferred for them.

Commands, with all test output redirected directly to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lane5-a-setup.log 2>&1
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane5/share-a/prepare.cjs /workspace/scratch/lane5-a-original /workspace/scratch/lane5-a-api/node_modules/typescript > /tmp/lane5-a-prepare.log 2>&1
node stage3/interface-downcasts/lane5/share-a/prepare-gaps.cjs /workspace/scratch/lane5-a-original /workspace/scratch/lane5-a-api/node_modules/typescript > /tmp/lane5-a-gaps-prepare.log 2>&1
node stage3/interface-downcasts/lane5/share-a/verify.cjs /workspace/scratch/lane5-a-original /workspace/scratch/lane5-a-api/node_modules/typescript > /tmp/lane5-a-verify.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareA$' -count=1 -v -timeout 10m > /tmp/lane5-a-oracle.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableShareACounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-a-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareA$|^TestCheckedViewCallableShareAFrontiers$|^TestCheckedViewCallableShareACounts$' -count=1 -v -timeout 10m > /tmp/lane5-a-restored.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableShareAFrontiers$' -count=1 -v -timeout 5m > /tmp/lane5-a-frontiers-pinned.log 2>&1
go vet ./internal/oracle > /tmp/lane5-a-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-a-counts-global-with-api.log 2>&1
```

First oracle PASS 44.056s; final fixtures/mutants/counts/frontiers PASS 33.489s;
share counts update PASS 13.533s; exact frontier pins PASS 2.032s. Verification:
51 fixtures / 17 pairs / 78 reads, plus seven frontier fixtures. Vet and diff
whitespace checks pass. Whole counts with API exits 1 in 47.532s. No whole-package
test or full gate was run. Earlier harness path/declaration errors were corrected
before the final runs and are not counted as frontier or mutant evidence.

Setup timing: Node ready 0.058s, Go ready 0.070s, clang ready 0.505s, markdown
ready 1.068s, submodules ready 390.925s, build cache warm 602.488s, done
602.518s. nproc 5; cgroup cpu.max 400000 100000; env file
/workspace/adamic-tools/env.sh. Original TypeScript pin is
050880ce59e30b356b686bd3144efe24f875ebc8 in an independent shallow checkout;
no cohere source is copied. The initial full-history original clone was stopped
and replaced by an exact shallow fetch. No other lane was merged.

Second checkpoint commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareA$/^rank-(102|120|276|291|357|363|429|435|444|450|453|456|465)$|^TestCheckedViewCallableShareAFrontiers$' -count=1 -v -timeout 10m > /tmp/lane5-a-batch2-restored.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableShareACounts$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-a-batch2-counts.log 2>&1
node stage3/interface-downcasts/lane5/share-a/verify.cjs /workspace/scratch/lane5-a-original /workspace/scratch/lane5-a-api/node_modules/typescript > /tmp/lane5-a-batch2-verify-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-a-batch2-counts-global.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareA$|^TestCheckedViewCallableShareAFrontiers$|^TestCheckedViewCallableShareACounts$' -count=1 -v -timeout 10m > /tmp/lane5-a-batch2-complete.log 2>&1
```

Second batch restored controls, negative pins and mutants PASS 22.944s; share
counts update PASS 11.295s, adding 39 rows without changing previous rows.
Verification passes 90 fixtures / 30 pairs / 158 reads plus eight frontiers.
Canonical diagnostic signatures follow the checker shim's union order and
explicit optional undefined; original Target declarations remain unchanged.
Additional reduced carriers preserve original union aliases and finite kind
discriminators. Adjacent enums use number representation; broader enum domain
certification is not claimed. No compiler or runtime code was edited.

Second checkpoint whole-table counts exits 1 in 64.907s with the same outside-share failures. The append-only share updater is green; the full table is not certified.
Cumulative fixtures, 60 executable arity mutants, eight frontier pins and all 90 count rows PASS 79.104s.
