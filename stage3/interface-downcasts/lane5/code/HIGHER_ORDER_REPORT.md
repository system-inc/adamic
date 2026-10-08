Built: complete logical producer checks for higher-order callable member reads; 13 pairs / 40 candidate reads certified toward roadmap step 09.
Commits: base 926a1d39; requested 09d5525f and 058635b9 cherry-picked first as 360084ad and 6f7902e0; this report travels with the higher-order delivery commit.
Commands/results: focused lower/native/JavaScript/oracle checks pass, Node comparisons and sanitizer/leak controls pass, 41 new count rows pass; required global updater remains red outside this group.
Mutants: removing either backend's logical producer gate admits all 13 incompatible callback producers; pinned runtime negatives catch both mutants by forbidden completed output.
Uncovered: this is the first completed compiler group, not the entire unit; remaining array, Map, Set, generic, overload, predicate, constructor, namespace and carrier frontiers remain pending.

Per-share certification in this group is a: 0 pairs / 0 reads; b: 2 / 19;
c: 11 / 21. The b pairs are copyPrologue and restoreEnclosingLabel. The c
pairs are createCallBinding, the two getSourceFile receivers, the two writeFile
receivers, build, ModeAwareCache.forEach, the parenthesizer callback factory,
watchDirectoryOfFailedLookupLocation and the two watchDirectory receivers.
The pinned member inventory, fixture hashes and exact declarations are in
higher-order-certificates.json. Reads are census candidate reads, not measured
dynamic executions. Adjacent carriers are explicitly reduced; complete upstream
carrier types and complete TypeScript execution are not certified.

A closure pointer establishes a representation, not a callback's logical
parameter or result type. Lowering now retains descriptors for complete supported
higher-order member signatures even when the producer is malformed. After all
bodies are lowered, the existing checker signature identity registry certifies
compatible producers. Both emitters select that certificate by immutable generated
code identity before the member call. This does not alter contextual types or use
method bivariance as evidence. Empty producer sets fail at the read.

Conservative assumption: only identical complete checker signatures certify this
path. Broader assignability, recursive callable signatures, null-bearing callbacks,
predicates, generics, rest parameters, constructors and method thunks retain their
existing refusal or fail closed. No protected compiler file is edited and no
cohere code is copied. Native method thunks lack this closure certificate and
cannot borrow the new path's representation proof.

Each pair has its original member declaration, a completing producer, a wrong
arity producer and a producer whose nested callback signature differs while the
outer physical closure representation is unchanged. Source Node completes all
three. The two malformed producers stop with exit 70, no stdout and the complete
member/type diagnostic pinned in expected.stderr. Completing controls agree with
Node in native release, native ASan/UBSan and generated JavaScript, and pass the
independent finishing leak checker. Two additional controls actually invoke an
optional callback with present and absent arguments, and invoke a returned closure.
They also agree with Node in all modes and pass leak checks.

The two actual compiler mutants return early from the logical producer gate in
each emitter. All 13 same-representation wrong-callback controls then print
completed with exit 0 in the mutated backend. The runner rejects build failures
as evidence and restores both source files in finally blocks. Ordinary arity
checks remain active; their changed diagnostics are not the logical mutant proof.

The eight factory-owned frontiers are excluded using the factory worker's own
498-read report: ranks 3, 9, 12, 24, 165, 174, 186 and 189. That list includes
TransformationContext.onEmitNode alongside the seven NodeFactory frontiers;
the conservative ownership decision avoids duplicating that worker's assignment.
The three share branches are read only, and none is merged. Delivery is based
on the specified area tip, plus the two explicitly requested Set cherry-picks.
The inherited Set number/boolean/string add/has, allocation identity and bad
signature controls pass again here; this alone does not certify the shares'
remaining Set pairs.

Exact commands, each test writing directly to its log:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lane5-code-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --ignore-scripts # in stage3/api; captured in api-setup.log
python3 stage3/interface-downcasts/lane5/code/run-mutants.py > /tmp/lane5-code-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestPrepareViewCallableRead$|^TestViewCallableShapeNative$|^TestViewCallableShapeNode$|^TestCheckedViewCallableCodeHigherOrder$|^TestCheckedViewCallableCodeCounts$|^TestCheckedViewSetIntrinsics$|^TestCheckedViewSetIntrinsicSignatures$' -count=1 -v -timeout 10m > /tmp/lane5-code-first-group-final.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCodeCallbackExecution$' -count=1 -v -timeout 5m > /tmp/lane5-code-callback-execution.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/lane5-code-counts-global-with-api.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCodeCounts$' -count=1 -timeout 5m -args -update-counts > /tmp/lane5-code-counts-owned.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableCodeCounts$' -count=1 -timeout 5m > /tmp/lane5-code-counts-check.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$/^fixtures$/internal/oracle/testdata/(census_overload_contracts|census_small_boolean|maybe_number_slots|graph_regions_regression_07)[.]a$' -count=1 -timeout 5m > /tmp/lane5-code-counts-baseline.log 2>&1 # isolated pinned-base worktree
git diff --check
```

Focused results: lower 0.284s, native 3.019s, JavaScript 0.482s, oracle 26.509s;
callback execution 1.249s; final recorded count check 4.215s. All PASS. The final focused oracle rerun, including both execution controls and all
41 count rows, also passes in 26.710s (logs/oracle-final.log). The
global updater was rerun after installing its missing pinned Node typings and
still exits 1 in 46.899s. Its remaining failures include existing overloaded-call
and template refusals, maybe-number C type mismatch, graph-region invalid frees
and unsupported process.exit values. It does not certify the whole count table.
Four representative failures reproduce unchanged on the pinned 926a1d39 base:
census_overload_contracts, census_small_boolean, maybe_number_slots and
graph_regions_regression_07. The isolated baseline log is counts-baseline.log;
other global failures are not independently triaged. No whole-package test or
full gate is run.

Setup timings: Node 0.025s, Go 0.027s, submodules 0.078s, markdown 0.079s,
clang 0.188s, build 83.857s, cache warm 84.016s, done 84.041s. nproc is 5;
cpu.max is 400000 100000. Environment: /workspace/adamic-tools/env.sh.
