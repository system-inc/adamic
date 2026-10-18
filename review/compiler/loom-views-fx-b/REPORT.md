Built receiver-specific reads, destructured contract metadata, and checked maybe-scalar conversion for 134, 136 and 144.
Commits: one independent delivery commit on compiler/loom-views-fx-b, based on main 4011e74a; its SHA is in the final response.
Commands and outputs: focused Node/backend regressions PASS, counts regenerated once PASS 108.336s, reader guard PASS 10.185s, vet PASS, all 51 reader packages compiled.
Mutants: all six receiver, destructuring, conversion, nullable, generic and union mutations caught by their intended behavioral checks.
Not covered: full package/repository suites, the remaining numbered deliveries, or runtime C changes.

This branch starts at 4011e74a8f59e40ea533f1e0613a2a0aa6b4997b and does not contain branch A. The original-main replay and complete 132-154 table are committed on branch A at 925f65b0d2eef508e07827ab930fba50bec1b693, under review/compiler/loom-views-fx-a/. They cover 91 distinct source programs in Node and both main backends. No old train merge, PR, Loom submission, or main push was made.

Implementation origins: dd5712e2 from compiler/fx6-valid-programs, the receiver safeguards in aaca4be5 from compiler/fx6-candidates-2-fix, and 81cef234 from compiler/fx7-scalar-union-view. Only the needed production hunks and focused fixtures were ported. Main has no shared readViewMember helper, so receiver normalization is applied at its existing object-read sites instead. Existing bare-name record-storage guards remain. The predicate proof test now checks the receiver-specific admission handoff rather than expecting a bare global name. No new IR map field, runtime representation or ABI was introduced.

134 keys checked reads by receiver type and member, including nested contracts. Unrelated same-named properties remain ordinary reads and writes. Nullable receivers retain the present object's identity; generic receivers use their instantiation or constraint; a union receiver preserves the checked common member. 136 records the destructured field's receiver, contract type and location. 144 checks membership and runtime scalar kind before constructing a packed maybe-number or maybe-boolean: undefined is absent while zero and false are present. JavaScript uses the corresponding scalar typeof checks.

Conservative assumption: the original lying callable p51 and invalid object-union p49 should stop at their checked reads. Their unchecked Node outputs are observations, not a reason to invoke an incompatible callable or expose an invalid union payload. Separate checked-stop fixtures pin their diagnostics. The valid p53 and supplemental scalar-union p51 are compared with source Node in both backends. Other fixtures include the original same-name p35-p38 and maybe-number variant, original p54 number/boolean transitions, string control, and wrong-string misfits.

regressions.log.gz records source Node, JavaScript, release native and sanitized native observations, plus ordinary oracle fixture/leak checks. Every new/touched focused leaf is below 60 seconds; leaf-seconds.json records the measured new functions, maximum 10.70s. Existing harness cache/build setup is included in those test durations. The initial cold command reached its 180-second outer limit during oracle compilation after the lowering leaves passed; focused-warm.log.gz is the completed retry, oracle PASS 20.360s. Final regressions passed lower 2.875s and oracle 60.390s; the package elapsed time includes many separate leaves.

Mutants are independent and restored in finally. patches, per-mutant logs and mutants.json record exact commands and exit 1:

| Mutant | Catcher |
|---|---|
| receiver keys collapse to a name-wide key | TestFX6P37, JavaScript stdout differs from Node |
| destructured ViewTypeID removed | TestFX6P53, JavaScript stdout differs from Node |
| native maybe-union validation removed | both TestScalarUnionViewMisfit leaves, native returns exit 0 instead of checked exit 70 |
| nullable receiver normalization removed | TestCheckedViewOptionalReadBoundary, optional checked read escaped |
| generic receiver left rigid | generic/wrong view flow, backend exit codes differ |
| union receiver propagation removed | stage3 assertions/19_identifier_kind.a/stage0, pinned gap changed |

No build failure, C warning or unrelated sanitizer failure is counted as a kill. The restored stage3 witness passed in stage3-restored.log, package 9.010s. run-mutants.py has a 720s outer bound, a 120s subprocess bound and -timeout 90s on every child. All six were caught.

Linux counts were regenerated exactly once, with the current unanchored sharded selector. Fourteen rows were added, none removed, and six existing rows changed. Existing allocations, frees, peak and region columns are unchanged; retains and releases each fall by 26 in total:

| Existing fixture | Retain/release change |
|---|---|
| visitor.a | -4 / -4 |
| lane1/interfaces-good.a | -1 / -1 |
| lane1/objects-good.a | -2 / -2 |
| untagged/fixtures/binding-name-source-good.a | -12 / -12 |
| untagged/fixtures/option-element-source-good.a | -5 / -5 |
| untagged/fixtures/structural-source-good.a | -2 / -2 |

changed-counts.json records the full before/after rows, emitted-C hashes, and explicit successful source Node, JavaScript and native comparisons for all six. The *-counts.patch files show the emitted-C differences. Receiver-specific guards remove checks on unrelated reads, explaining the equal ownership-pair reductions; the numeric deltas themselves are observations from regeneration. Successful new rows balance allocations/frees; checked-failure rows stop before cleanup by design.

Commands, after sourcing /workspace/adamic-tools/env.sh, with all output redirected under this directory:

- timeout 120 go test ./internal/lower ./internal/oracle -run '^TestFX6|^TestScalarUnionView|^TestPredicateBodyProof$|^TestCheckedView|^TestInterfaceCast|^TestRequiredView|^TestNarrowedField|^TestViewField|^TestDefaultTagged|TestNativeAgreesWithNode/internal/lower/testdata/(fx6|scalar_union_views)' -count=1 -v -timeout 90s: regressions.log.gz.
- timeout 720 python3 review/compiler/loom-views-fx-b/run-mutants.py: mutant-run.log, all six caught.
- timeout 360 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 300s -args -update-counts: counts.log, PASS 108.336s, one regeneration.
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: readers-final.log, PASS 10.185s.
- timeout 90 go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle: vet.log, exit 0, no diagnostics.
- timeout 300 go test $(cat review/compiler/loom-views-fx-b/reader-packages.txt) -run '^TestFX6|^TestScalarUnionView|^TestCallTargetReaders$' -count=1 -timeout 90s: readers-packages.log, all 51 reader packages compiled and matching focused leaves passed. Packages without matching tests explicitly report no tests to run. The import closure is unchanged from branch A and includes test imports.
- timeout 120 go test ./stage3/fixtures -run 'TestFixturesAssertions/assertions/19_identifier_kind.a/stage0' -count=1 -v -timeout 90s: stage3-restored.log, PASS 9.010s.
- timeout 120 go build -o /tmp/loom-views-b ./cmd/adamic: cli-build.log. The six changed-count programs were each run through source Node, js plus Node, and native build plus execution, with 30s bounds; commands and byte outputs are in changed-counts.json.
- timeout 240 python3 review/compiler/loom-views-fx-b/check-combination.py: combination.log and combination-results.json. A Go source overlay combines the two fixes without editing either branch. Both c and js give the expected located compile-time refusals for original key p05, unsupported p37 and unsupported callable p21. The overlay sources end .go.txt and are evidence, never ordinary compilable Go files.

The required setup is shared with branch A. setup.log retains the complete output: GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh; Go 0.080s, Node 0.063s, clang 0.564s, Markdown 1.278s, submodules 213.506s, shared cache 219.469s, build 588.893s, cache warm 589.199s, done 589.339s. nproc=5, four-CPU cgroup quota; Go 1.27.1, clang 20.1.8, Node 24.19.0. The worktree shares the already installed submodules and pinned Node declarations through local links; no gitlink, dependency manifest or link is committed. Git worktree checks ignore these local submodule links and explicitly stage only the unit's files.

Runtime C files changed: none. The only native production change is internal/native/view_unions_read.go.

Deferred deliveries: 133, 135, 137-141, 143, 145-154. The name-family items 147-149 may benefit from receiver keying but their pending review rows are not activated or claimed as extra deliveries. Item 150 has no complete fix in the named historical branches. No code was copied from cohere. No whole-package or full repository gate was run.
