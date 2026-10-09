Built: slice 2 independently on main 5e33a17b for #wj4pmt1, with fix-forward 2 first.
Commits: a3208531 (2a), ab48779c (2b), 0b663d52 (self-compare), fa6d7689 (callback widening).
Commands and outputs: 274 lowering leaves pass in 12 shards, 141 member/support leaves pass, counts pass with 46 attributed additions, reader and fixture-directory guards pass.
Mutants: all 17 semantic mutants caught, plus the recorded structural inline-definition mutant; every catcher is listed below.
Not covered: per-backend-stops requires step24; full gate and optional WASI execution were not run; two pre-existing inventory tests skip without external corpus configuration.

## Slice 2

| Member | Source SHA | Decision | Dependency evidence | Tasks closed by this delivery, pending integration |
| --- | --- | --- | --- | --- |
| miscompile-fxspptb-2a | 373f0055 | kept | lastFieldAssignment at internal/lower/class.go:716 and classBases at class_inheritance.go:917 exist on main; original census marker fixture exists on main. Error fallback and metadata repair are in its own net. | #63pvx2b, fix-forward 2 portion 2a |
| miscompile-fxspptb-2b | 6de2edb5 | kept | Weak representation, libraryMember, narrowing and method signature helpers exist on main. No callback adapter is needed for its own containment. | #63pvx2b, fix-forward 2 portion 2b |
| self-compare-main | 387c2826 | kept | Existing emitter declarations and Split build support are on main. Repair 9679fbe4 preserves inline definitions; a233eec3 supplies only this directory owner, adapted to main's parallel rule. | #7c6b4pq, roadmap step 04 |
| per-backend-stops | 7a151f7f | dropped | needs step24-parser-main for ADAMIC_CHECK_STACK_NAMED at internal/native/runtime/adamic.h:1074 (e503868c); its backend_stops_test.go pins the named depth diagnostic. Main stack.c:66 emits the unnamed diagnostic. | none; #z1vjxxd remains open |
| callback-widening | 36472529 | kept | Effects, MakeClosure, CallClosure, Environment, fit and contextual checker signatures exist on main. No search-shrink IR field or helper is used. | #1xxqbrk |

The chain head is f5236b48. The source census is e8c4c8e on compiler/chain-slice-1.
Refined source bases and exact extracted patches are saved under ../chain-slice-2/.
The delivery contains no chain or slice-1 merge. Each member has one extraction commit.
The final evidence commit also carries the late-identified self-compare directory-owner repair.
Shared files and ancestry are not counted as dependency edges. The repair ledger is
[repair-attribution.json](../chain-slice-2/repair-attribution.json).

## Proof

All commands source /workspace/adamic-tools/env.sh, use
GOPROXY=https://proxy.golang.org|direct and GOMAXPROCS=4. Output goes directly to logs.

- Internal/lower: 276 top-level declarations scheduled across twelve regex shards with go test -count=1 -json -timeout=90s. 274 passed, two existing inventory tests skipped. Largest final shard package duration 37.387 seconds. TestOptionalWideningCensus needs OPTIONAL_WIDENING_CONFIG/OUTPUT; TestOriginalCycleLedger needs a pristine TypeScript corpus through ADAMIC_CYCLE_LEDGER_ROOT. No member test skips. A missing pinned @types/node dependency first failed shard 3; npm ci --prefix stage3/api fixed setup and that shard passed on rerun.
- Member oracle command: go test ./internal/oracle -run '^(TestMiscompile2A|TestWeakDifferential|TestSelfCompare|TestCallbackWidening)' -count=1 -timeout=90s -json. 66 top-level leaves passed against source Node and both backends, native sanitizers and clean-exit leak checks where applicable. Known unsafe programs stop in shared lowering; freed-Weak terminal observations use their explicit native/Node lifetime contracts. Largest leaf 27.31 seconds.
- Support command: go test ./internal/native ./internal/ir ./internal/flow ./internal/fresh -run '^(TestErrorUndefinedMessageHasNoOwnProperty|TestSplitSelfCompareAgreesWithNode|TestCallback|TestClosureTargetsBoundOnlyProvenValues|TestCallTargetReaders)' -count=1 -timeout=90s -json. 75 top-level leaves passed; largest 30.9 seconds. The self-compare directory test passed in 12.91 seconds; its directory-union guard also passed.
- TestCallTargetReaders was rerun separately and passed. No reader allowlist entry was required.
- Admission delta: 1,484 inputs compared using main production-source overlays and the slice, in bounded top-level probe shards. There are 34 observed changes: 16 belong to 2a, 16 to 2b and two to callback widening. Four main computed-member panic witnesses now produce 2a's named NotYet stop. Recovery in the evidence probe records main's panic rather than losing subsequent rows. Every input has exactly one observation per side; every changed row has an owner. This is a corpus proof, not universal admission equivalence. See admission-delta.json and the source/overlay files under ../chain-slice-2/.
- Counts command: go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=600s -v -args -update-counts. PASS in 163.562 seconds. One successful table regeneration: 1,011 to 1,057 rows. All 46 additions are attributed individually in counts-attribution.json; no existing row changed or disappeared. Earlier 90-second and 180-second attempts timed out without writing the table. A main-only diagnostic probe over the timeout candidates passed in 0.661 seconds; no main defect is inferred from those aggregate timeouts.
- go test ./stage3/fixtures -run '^(TestFixturesSelfCompare|TestFixtureDirectoriesHaveTopLevelTests)$' -count=1 -json -timeout=90s passed. The aggregate proof supervisor reached its 600-second limit at completion; the complete final test log independently records PASS. No test process remains running.

The exact commands and individual seconds are in proof-runs.jsonl, the shard logs,
fixtures-retry-seconds.json and support-retry-seconds.json. Failed setup/probe attempts
remain labeled by their original logs, alongside successful retries.

## Mutants

| Mutant | Catcher observed |
| --- | --- |
| 2a error-spread | exact Error and subclass spread stop pins |
| 2a maybe-setter | exact optional numeric setter stop pin |
| 2a long-name | exact long property-name stop pin |
| 2a computed-name | recorded compiler ComputedPropertyName panic returns when containment is removed |
| 2a constructor-capture | both constructor-arrow refusal pins |
| 2a optional-error | UBSan null member read in supplied optional witness and census omitted-argument driver |
| 2a error-own-message | native own-property result differs from Node |
| 2b method-views | wrong native output and sanitizer report, with both view pins detecting admission |
| 2b union-liveness | native exit mismatch and null member read |
| 2b weak-callback | all six exact NotYet pins fail on the combined slice's independent adapter refusal; isolated 2b on main reproduces unexpected admission, ASan heap-buffer-overflow and sanitized exit mismatch |
| callback refusal-revert | eight retired refusal pins |
| callback adaptation-omitted | nine numeric intrinsic/generic-map witnesses |
| callback nullable-token-omitted | nullable object map Node comparison |
| callback unrepresented-parameter-admitted | structural array parameter refusal pin |
| callback adapter-target-omitted | TestCallbackAdapterTargets |
| self-compare constant-fold | successful native run differs from Node at NaN in sweep.a |
| split numeric fold | successful split native run differs from Node |
| static inline definition removed | split build rejects missing inline definition; structural evidence only, not credited as a semantic mutant |

All Go/runtime mutations use overlays and sources named .go.txt. Original recorded
runners are also retained for provenance. The first callback runner discarded setup's
trimpath flag and was stopped during redundant compilation; the successful runner
preserves configured flags. Weak-callback's initial runtime-only validator rejected
the newly redundant containment; the final ledger distinguishes exact diagnostic
containment from the isolated original runtime fault. Source code was never left mutated.

## Toolchain and lanes

nproc=5; cgroup quota is four CPUs. Initial setup reported Go 0.085s, Node 0.087s,
clang 0.522s, markdown 2.026s, submodules 26.500s, shared cache 34.676s. Its bounded
240-second attempt did not reach build/done lines. The environment file was usable;
focused builds warmed compiler dependencies. The retry log records final setup state.
The integration lane command runs on the committed branch and is saved in
../chain-slice-2/lane-checks.log. Git diff --check passes.

No code was copied from cohere. No protected orchestration file changed. The delivery
advances #wj4pmt1's independent green slice work; the supplied brief does not assign
another numbered roadmap step to the fix-forward or adapter members.
