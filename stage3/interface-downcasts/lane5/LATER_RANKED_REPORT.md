Built: original rank 67 getLocalName certified; original rank 61 object-union read retained as a refusal witness.
Commits: continued from ea341202ed7e4bc87c587c47f35a7f7a6e2b504e; each completed group pushed separately to codex/views-callables.
Checks: Node, native release/sanitized, JavaScript, leaks, original declarations/read spans and lane counts pass; required whole-table counts still fail inherited cases.
Mutants: native/JavaScript arity, result and directional parameters plus deferred payload registration all caught for rank 67, every mutation restored.
Uncovered: conservative 29/2818 pairs and 1258/11063 candidate reads certified; 2789 pairs and 9805 reads remain. Static totals unchanged at 4/308 pairs and 34/1503 reads.

Reporting date October 12. Original truth: TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8. later-ranked-original-witnesses.json retains exact declarations, read spans, source hashes and declaration file hashes. The verifier resolves Scanner declarations in scanner.ts rather than types.ts. Fixtures preserve the complete declarations and original read expressions; adjacent carriers/helpers are reduced. Candidate read totals are inventory counts, not measured production reachability.

Local-name group: rank 67, 23 candidate reads, seven fixtures. Original getLocalName(node: Declaration, allowComments?: boolean, allowSourceMaps?: boolean, ignoreAssignedName?: boolean): Identifier retained. All three Boolean flags are used in the result, with omitted/undefined, all false, all true and a mixed combination held to Node. The original direct factory.getLocalName read is retained. Oracle passed 3.375s before mutants, and the restored oracle passed (see raw log). Lane counts update passed 25.972s, adding exactly seven new measured rows. Whole-table updater still fails previously recorded predicate/overload/process/nominal-write refusals and native invalid frees.

Rank 61 is not certified: the full createStringLiteralFromNode declaration and original factory member read, with a reached sourceNode.value producer read, refuses unsupported untagged object union contract. Node prints 3. A separate test pins that refusal. No pending Union exception, callback convention or implementation-proof refusal was relaxed. Constant-return substitution to avoid the actual object-union payload read was not used as evidence.

Commands, each test output redirected directly to its log:

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane5/verify-aggregate-witnesses.cjs /workspace/scratch/lane5-ts-pin 61,67,68,69,70,73,74,75,76 later-ranked-original-witnesses.json > /tmp/lane5-later-original.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-local-name-oracle.log 2>&1
python3 stage3/interface-downcasts/lane5/run-later-ranked-mutants.py local-name > /tmp/lane5-local-name-mutants.log 2>&1
node stage3/interface-downcasts/lane5/verify-later-ranked-fixtures.cjs /workspace/scratch/lane5-ts-pin 61,67 > /tmp/lane5-local-name-original.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-local-name-counts.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-local-name-all-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableLaterRanked' -count=1 -timeout 5m > /tmp/lane5-local-name-restored.log 2>&1
```

Mutants: native-arity and javascript-arity admit mismatched arity, caught by the pinned member-read refusal. native-result and javascript-result admit a numeric result into the object result convention, caught by the member-read pin versus later object failure (native ASan diagnostics retained). native-parameters and javascript-parameters admit a string-only node parameter, caught by the pinned early parameter refusal. payload-reads removes aggregate descendant registration and is caught by the node.value read pin versus AddressSanitizer SEGV. Raw outputs retained under logs/later-ranked. No whole-package tests, full gate or PR. No production compiler/runtime changes in this group; only fixtures, oracle/counts registration, evidence and verification tooling.
