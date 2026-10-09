Compiler-area recheck: seven reduced boundary fixtures; no new lowering accepted.
Commits: views correction 6bb75e52; compiler-area merge d0af70b8, incorporating area tip b68b2fe1.
Existing element-access and checked non-null oracles pass; seven boundary probes pass; counts refresh passes with no table change.
Boundary mutant caught: changing NodeArray<Entry> to readonly Entry[] makes lowering succeed and fails the pinned NotYet expectation.
Remaining: 72 derived arrays and two dictionaries; original table rule-shape coverage remains 17/91.

The attempted ffe428ab views merge never completed or produced a merge commit.
It was aborted. After the user's correction, git merge-base verifies ffe428ab
is not an ancestor and c41c0e06 is an ancestor. No revert or force-push was
needed. The merge-attempt report is explicitly superseded; it asks for no ABI
approval. Current origin/area/compiler b68b2fe1 was merged without conflicts.
Neither main nor an area branch was written.

The .ts-only checked non-null prerequisite remains at c41c0e06, merged into
this branch at c7095bc7. The October 8 00:27 ruling governs: ! remains refused
in .a. Focused TestCheckedNonNull/TestNonNull and command explanation tests passed.
The isolated views checkout is unpushed and excluded from this unit. Its fixtures,
certifications and emitter fix are not carried here or counted toward coverage.

Each remaining receiver kind has a reduced .a under
internal/oracle/testdata/notyet_element_access/, registered from the new own
file element_access_boundaries_test.go as deliberately unlowered. The test runs
valid source on Node before pinning the exact named compiler-area stop.
These are boundary witnesses, not runtime certificates for unsupported shapes.
No new backend, IR, runtime C or another worker's lowering function was edited.

| Shape | Raw root sites | Observed reduced stop | Required owner / ruling |
| --- | ---: | --- | --- |
| NodeArray | 62 | an ElementAccessExpression | representation and elementType; metadata-preserving array storage/construction |
| Readonly<PathPathComponents> | 6 | a function returning Path | undefined | representation must first support the branded string Path; then the mapped readonly branded array representation and elementType |
| TemplateStringsArray | 2 | an ElementAccessExpression | representation/elementType and reference-valued raw metadata |
| SortedReadonlyArray | 1 | an ElementAccessExpression | representation/elementType and void brand policy on this area |
| JSDocArray | 1 | an ElementAccessExpression | representation/elementType and optional jsDocCache storage |
| Record<number, FlowGraphNode> | 1 | an ElementAccessExpression | open numeric dictionary representation; index-signature policy remains in force on this area |
| MapLike<string> | 1 | Refused: an index signature | docs/0.1.md:92 ruling; area dictionary representation and operations are not admitted |

TemplateStringsArray's structural declaration is included in its witness so
it does not depend on the restricted library exporting that global interface.
The actual Path brand is string & { __pathBrand: void }, as in types.ts:23.
The path witness correctly records its earlier result-representation stop;
no claim is made that its numeric access was reached. Ownership was refreshed
through origin/codex/notyet-*; representation remains the representations
worker's function, elementType the element-type worker's function. The existing
fields.a write witness still belongs to the setIndex worker's prerequisite.

Both binder examples were replayed after the compiler-area merge. Both exit 0
and reproduce ElementAccessExpression at 1746:21 and 1757:28; later reading-clause
findings are at 1758:18 and 1761:17. Exact commands and findings are in
area-replays.json. No new original roots are claimed lowered.

Commands, with output redirected directly to /tmp/element-access-area-*.log:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestCheckedNonNull|TestNativeAgreesWithNode/internal/oracle/testdata/element_access_' -count=1 -timeout 10m
go test ./internal/lower ./cmd/adamic -run 'TestNonNull|TestCensusSmallFiniteKeyRead|TestExplainChecksDriver' -count=1 -timeout 10m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestElementAccessCompilerAreaBoundaries -v -count=1 -timeout 10m
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
python3 stage3/notyet-element-access/coverage.py
```

Observed: existing oracle ok 1.867s; lower ok 0.105s; command package ok 2.053s;
exact boundary probes ok 0.473s; counts ok 24.251s with no diff. Each boundary
source prints witness and exits 0 on Node. The receiver mutant fails the boundary
test (go test exit 1) because Lower returns nil error for an ordinary readonly
array, not because of a Go/clang failure. Full log and mutation are in
area-boundary-mutant.json. The fixture was restored. Prior nine semantic/check
mutants are unchanged and were not rerun for a merge that changes no lowering
rule. No whole-package test or full gate was run.

The environment resumed ready. The required setup was run in the now-excluded
isolated checkout before the correction: export GOPROXY=https://proxy.golang.org|direct;
bash cloud/setup.sh. It passed with Node ready 0.025s, Go ready 0.024s,
markdown dependencies 0.080s, clang ready 0.176s, submodules ready 291.692s,
build/cache ready 615.054s, done 615.089s; nproc 5, cpu.max 400000 100000.
/workspace/adamic-tools/env.sh was sourced for all compiler-area commands too.
The isolated views counts attempt failed on unrelated integration host/graph
fixtures and is not claimed as a passing area check; the corrected area counts
refresh above passed. Its independent observations are not unit coverage.
