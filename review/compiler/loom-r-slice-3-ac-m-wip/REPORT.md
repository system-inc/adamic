Built: re-cut compiler slice 3 onto fetched main, preserving cached checker builds and main's release linker selection.
Commit: one commit on 5a94465370ea0172e25aae8c13478e59e40c2bb4; source f75b244b339ac7a0efca162dfc96805547763fe4; the final delivery SHA is in the final response.
Commands and outputs: full recursive suites were run for all 14 touched package roots; 13 package roots have passing full runs; markdown remains red. Final vet and guards pass; all attempts are recorded below.
Mutants: 14/14 recorded production mutants caught; both header mutants and all six direct assignment source mutants caught.
Not covered: no ahra/Loom/full gate, macOS or WASI run; full markdown validation remains red from the recorded cold setup and disk failures.

## Delivery and assumptions

Requested branch: compiler/loom-r-slice-3-ac-m. Delivered as compiler/loom-r-slice-3-ac-m-wip: the code conflict is resolved, but the mandatory full markdown run is not green and its over-budget leaves were not rerun. Conservative assumption: use the authorized WIP form for this validation blocker rather than present a normal delivery as ready. Base: 5a94465370ea0172e25aae8c13478e59e40c2bb4, the newest origin/main returned by the initial fetch, newer than 5179838b3bd78efe21c87cc359b6adf918fd0a09. Source: origin/compiler/loom-r-slice-3-ac at f75b244b339ac7a0efca162dfc96805547763fe4.

Applied only the source commit's own change with git cherry-pick --no-commit, never a main-to-source diff. The final message preserves the complete original message, adds the required Re-cut line and Co-Authored-By trailer, and records the assumptions. Conservative assumption: the fetched main snapshot is the delivery base; main is not chased during validation. The latest request for complete recursive suites takes precedence over the earlier focused-only and 30-minute first-push workflow. Existing main setup tests were kept intact rather than changed to hide cold-run failures. No main behavior was intentionally reverted beyond the source change's replacement of repeated checker-runtime compilation with its cached runtime.

A local candidate commit was necessary before finishing the full suites: main's corpus selector refuses staged GAPS.md changes against HEAD. Final evidence is folded into that same unpushed commit; the delivery still has exactly one parent commit on main.

## Conflicts and automatic merges

One conflicted hunk, internal/native/tsgo.go BuildTSGo: main compiled the runtime's C units fresh and selected linkFlags(tsgoFlags(...), options), including Linux's lld for release ThinLTO. The source reused splitTSGoRuntime, linked relative runtime/archive names, and retained Darwin's separate main.o build and fixed mtime. Kept the source's cache and relative paths, both build/runtime prefix maps, its runtime/header validation, compiler-token limits and Darwin path. Applied main's linkFlags to a clone of flags at the final link only; linker-only flags do not reach Darwin's -c compilation. conflicts.diff and tsgo-resolution.patch preserve the hunk and resolution.

Checked internal/ir/ir.go's automatic merge: Function retains both the source's SourceLength and main's CheckedUnionNarrow. Also reviewed internal/lower/functions.go and cycles.go: source-length recording and library Function recognition coexist with main's newer function lowering and static-global condition. auto-merges.patch records these checks. No other code conflict occurred. counts.md was reset to main's version and generated, not hand-merged.

## Runtime clearance

Extra runtime C/header lines beyond the original commit's own delta: NONE. Signed added/removed lines are identical to the source's own delta in adamic.h, closure.c and union.c. runtime-clearance.json.txt records the exact comparison. Main's newer runtime code stays. No static or runtime feature set was added by this resolution.

## Counts

On Linux with house clang 20.1.8, ran exactly once:

    go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -parallel=4 -timeout=10m -args -update-counts

PASS, 116.660 package seconds. Started with main's table. Changed 33 rows: 22 new and 11 existing; every key is one the source commit itself changes. Thirty-two measured rows equal the source's recorded rows. json_shorthand.a measures 20/20/8/29/4/0 here, versus the source's 18/18/8/27/4/0. Observation: main's JSON runtime has advanced; attributing those extra allocations to a particular newer change was not separately proved. The Linux measurements are retained. No unrelated row changed.

| Fixture | Before (alloc/free/retain/release/peak/region) | After |
|---|---|---|
| internal/load/testdata/0.1/compile/09_tree.ts | 19/19/28/69/16/0 | 33/33/66/69/16/0 |
| internal/oracle/testdata/047cb0d_n_arrayindex.a | 1/1/1/2/1/0 | 2/1/4/4/1/0 |
| internal/oracle/testdata/047cb0d_n_element.a | 4/3/6/8/4/0 | 5/3/9/10/4/0 |
| internal/oracle/testdata/047cb0d_n_element_method.a | 4/3/6/9/4/0 | 5/3/9/11/4/0 |
| internal/oracle/testdata/047cb0d_n_element_plain.a | 4/3/5/8/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/call_spread_array.a | new | 56/56/112/115/14/0 |
| internal/oracle/testdata/element_access_presence.a | 4/3/5/7/4/0 | 5/3/8/10/4/0 |
| internal/oracle/testdata/json_shorthand.a | new | 20/20/8/29/4/0 |
| internal/oracle/testdata/narrowed_fields.a | 4/3/4/6/4/0 | 5/3/7/8/4/0 |
| internal/oracle/testdata/narrowed_methods.a | 3/3/3/6/3/0 | 4/3/6/8/3/0 |
| internal/oracle/testdata/narrowed_reads.a | 5/5/3/7/4/0 | 6/5/6/9/4/0 |
| internal/oracle/testdata/optional_after_call_193.a | new | 3/3/6/9/3/0 |
| internal/oracle/testdata/optional_after_call_79.a | new | 4/4/6/10/4/0 |
| internal/oracle/testdata/optional_after_call_gaps/computed.a | new | 3/3/1/6/2/0 |
| internal/oracle/testdata/optional_after_call_propagation.a | new | 18/18/13/20/7/0 |
| internal/oracle/testdata/optional_after_call_required.a | new | 3/3/5/7/2/0 |
| internal/oracle/testdata/optional_after_call_variants.a | new | 11/11/10/21/3/0 |
| internal/oracle/testdata/refusal_rulings/function_length.a (TypeScript mode) | new | 20/20/4/29/13/0 |
| internal/oracle/testdata/refusal_rulings/function_signature.a | new | 3/3/0/3/3/0 |
| internal/oracle/testdata/refusal_rulings/function_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 |
| internal/oracle/testdata/refusal_rulings/merge_missing.a (TypeScript mode) | new | 1/0/1/1/1/0 |
| internal/oracle/testdata/refusal_rulings/merge_missing_computed.a (TypeScript mode) | new | 1/0/1/1/1/0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a | new | 3/3/1/4/3/0 |
| internal/oracle/testdata/refusal_rulings/merge_present.a (TypeScript mode) | new | 3/3/1/4/3/0 |
| internal/oracle/testdata/refusal_rulings/merge_unused.a (TypeScript mode) | new | 0/0/0/0/0/0 |
| internal/oracle/testdata/regexp_null_narrowed.a | 4/4/8/5/3/0 | 5/4/11/7/3/0 |
| internal/oracle/testdata/reuse_narrowed.a | 7/5/6/8/5/0 | 8/5/9/10/5/0 |
| stage3/fixtures/assignment-proofs/01_scanner_keyword.a | new | 18/18/18/32/6/0 |
| stage3/fixtures/assignment-proofs/02_if_narrowing.a | new | 6/6/5/13/3/0 |
| stage3/fixtures/assignment-proofs/03_return_defined.a | new | 8/8/6/18/4/0 |
| stage3/fixtures/assignment-proofs/04_chained.a | new | 4/4/6/15/3/0 |
| stage3/fixtures/assignment-proofs/05_compound.a | new | 8/8/0/8/4/0 |
| stage3/fixtures/assignment-proofs/06_wider_target.a | new | 4/4/0/4/4/0 |

## Full package tests

GOMAXPROCS=4 and affinity to four CPUs. Each command had an outer hard limit, -count=1, -json, -parallel=4 and -timeout=11m. Later recursive commands use -p 1 so oracle-owner and reader packages do not compete on the same four-CPU share. Main's compiler-free oracles were prepared with their TestProduct_ functions before subsequent reader runs. suites.jsonl holds every exact command; compressed logs and per-test seconds are alongside it. Failed-node counts include both family parents and leaves.

| Package | Attempt | Exit | Wall seconds | Failed nodes | Leaves >=60s |
|---|---:|---:|---:|---:|---:|
| internal/ir | 1 | 1 | 36.367 | 2 | 0 |
| internal/javascript | 1 | 0 | 1.47 | 0 | 0 |
| internal/lower | 1 | 0 | 49.738 | 0 | 0 |
| internal/flow | 1 | 0 | 74.551 | 0 | 0 |
| internal/ir | 2 | 0 | 130.091 | 0 | 0 |
| internal/native | 1 | 0 | 450.148 | 0 | 0 |
| stage1/cohere/cssstrings | 1 | 0 | 38.484 | 0 | 0 |
| stage1/cohere/json | 1 | 1 | 85.329 | 2059 | 0 |
| stage1/cohere/markdownblocks | 1 | 1 | 272.981 | 796 | 4 |
| stage1/cohere/selector | 1 | 1 | 14.872 | 1 | 0 |
| stage1/cohere/yaml | 1 | 0 | 144.545 | 0 | 0 |
| stage1/typescript/parser | 1 | 0 | 149.854 | 0 | 0 |
| stage1/typescript/scanner | 1 | 1 | 39.872 | 17 | 0 |
| stage3/fixtures | 1 | 0 | 80.029 | 0 | 0 |
| internal/oracle | 1 | 1 | 462.282 | 76 | 0 |
| stage1/cohere/json | 2 | 0 | 358.51 | 0 | 0 |
| stage1/typescript/scanner | 2 | 0 | 20.564 | 0 | 0 |
| stage1/cohere/selector | 2 | 0 | 12.658 | 0 | 0 |
| internal/oracle | 2 | 0 | 370.248 | 0 | 0 |

Initial environment failures are preserved, not credited as compiler mismatches:

- IR first failed exporting Go archives to a full root disk; the complete retry passed. Its first raw log was overwritten by the retry; initial-disk-failure.txt and suites.jsonl retain what was observed.
- JSON's pinned historical corpus contains four /workspace/adamic tokens whose original evidence file is no longer tracked. On this builder those happen to name the current tree, so its product relocation check refused them. Built the identical pinned product in an isolated relocated worktree without changing the guard or corpus bytes, then verified the original checkout reads that same portable product (key 05bb4737819b). Both logs are retained. New review data uses .json.txt so it does not change the declared JSON source census.
- Selector initially reported an unsettled Go answers owner; its owner was prepared before retry.
- Scanner's Git helper strips the session network route and its automatic clone could not reach GitHub. Provisioned the exact v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8 with the session route, sparsely retaining src/compiler, and used its supported ADAMIC_TYPESCRIPT_SOURCE setting.
- Oracle's first full run completed with 76 failed nodes after its separate /home/agent/.cache/adamic/gate result cache filled root. Inspected failures say no space left on device. The attempt to stop it found it had already exited. Preserved this unit's result-cache files in a byte-verified archive on task scratch, then retried with ADAMIC_GATE_UNCACHED=1, a task-local XDG_CACHE_HOME, and the old build cache retained as an explicit read source.
- Full markdown ran before the candidate commit and hit the dirty-GAPS corpus guard, disk exhaustion and cold setup overruns. It remains an unresolved full-suite result; no over-budget leaf was rerun and no tests were deleted or loosened.

Timing exceptions: native's existing TestSplitTSGoAgreesUnit00 parent took 167.89s including a cold checker archive (126.89s) and source product (2.86s). The leaf-only summary excludes that parent's setup, so it is stated explicitly. Markdown's over-budget leaves were TestMarkdownListLayout_Setup 81.89s, TestProduct_WholeDocumentPreflightGoProducts 83.44s, TestProduct_MarkdownblocksTextFormatter 68.13s, and TestProduct_GrainMarkdownFormatter 88.62s. These are not represented as satisfying the 60-second requirement. They were not rerun. Do not treat this report as a fully green Loom certification.

## Focused checks, guards and mutants

The checker cache's real Node/native cold/warm witnesses passed: new TestTSGoCacheShippedRelease 4.96s, normal TestTSGoCacheRelease 5.59s, sanitized TestTSGoCacheSanitized 8.72s. They compare program output to source Node, cold/warm binary bytes, absence of builder paths, and the cached archive's identity/mtime.

All 22 changed Adamic programs pass the unchanged pinned a-check predicate using this branch's compiler: 16 clean, six refused with matching first-line reasons, zero NotYet/type-error/unclassified. Both recorded header mutants (removed header and wrong reason) remain genuinely refused and make that predicate exit 1. after.json.txt and mutants.json.txt hold observations.

The carried assignment-proof check.py still asserted the redundant pass headers the source removed. Changed only that assertion to require no a-check header on these six clean programs. Its six proofs compile, match Node, have zero stock TypeScript 6.0.3 diagnostics, and all six source statement/input mutants fail their independent Node stdout goldens. assignment-check.log and assignment-proof-logs/ preserve the outputs. No program body was changed by this re-cut.

IR's sorted-map and call-target reader guards, lower's registered-agreement-program guard, native's statics guard, and oracle's runtime-set witness guard are included in the full suites. Naming and parallelism guards were also run explicitly. No new Go package was introduced, so no compiler-consumer dependency entry was needed. All changed Go files pass gofmt; product/test diff whitespace is clean. Historical review patches/logs carried verbatim from the source contain pre-existing whitespace, so whole-diff whitespace output is not used as product evidence.

| Final focused command | Exit | Wall seconds |
|---|---:|---:|
| timeout 720 python3 /workspace/adamic/review/compiler/loom-r-slice-3-ac-m/run-mutants.py | 1 | 116.508 |
| timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -parallel=4 -timeout=90s | 0 | 2.939 |
| timeout 180 go vet -p 1 ./internal/ir/... ./internal/javascript/... ./internal/lower/... ./internal/flow/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/cssstrings/... ./stage1/cohere/json/... ./stage1/cohere/markdownblocks/... ./stage1/cohere/selector/... ./stage1/cohere/yaml/... ./stage1/typescript/parser/... ./stage1/typescript/scanner/... ./stage3/fixtures/... | 1 | 15.356 |
| timeout 120 go test ./cmd/adamic-gate -run '^TestEveryTest(HasOneDescriptiveName\|IsParallelOrSaysWhy)$' -count=1 -timeout=90s | 0 | 0.768 |
| timeout 120 go test -overlay /workspace/adamic/review/compiler/loom-r-slice-3-ac-m/linker-mutant/overlay.json.txt ./internal/native -run '^TestTSGoCacheShippedRelease$' -count=1 -timeout=90s -v | 1 | 8.621 |
| timeout 120 go test ./stage1/cohere/markdownblocks -run '^TestParserRepresentationProbes$' -count=1 -timeout=90s -v | 0 | 6.997 |
| timeout 600 python3 review/compiler/loom-r-slice-3-ac-m/run-mutants.py function-annotation function-call function-length erase-earlier-throw tsgo-absolute-main tsgo-rewrite-runtime | 0 | not measured |
| timeout 180 go vet -p 1 ./internal/ir/... ./internal/javascript/... ./internal/lower/... ./internal/flow/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/cssstrings/... ./stage1/cohere/json/... ./stage1/cohere/markdownblocks/... ./stage1/cohere/selector/... ./stage1/cohere/yaml/... ./stage1/typescript/parser/... ./stage1/typescript/scanner/... ./stage3/fixtures/... | 1 | 37.998 |
| timeout 120 go test -overlay review/compiler/loom-r-slice-3-ac-m/linker-mutant/overlay.json.txt ./internal/native -run '^TestTSGoCacheShippedRelease$' -count=1 -timeout=90s -v | 1 | 10.873 |
| timeout 180 go vet -p 1 ./internal/ir/... ./internal/javascript/... ./internal/lower/... ./internal/flow/... ./internal/native/... ./internal/oracle/... ./stage1/cohere/cssstrings/... ./stage1/cohere/json/... ./stage1/cohere/markdownblocks/... ./stage1/cohere/selector/... ./stage1/cohere/yaml/... ./stage1/typescript/parser/... ./stage1/typescript/scanner/... ./stage3/fixtures/... | 0 | not measured |

Recorded production mutants use Go overlays; sources are .go.txt. A behavioral catch requires a test failure, excluding build/clang failures and timeouts.

| Mutant | Catcher | Caught | Wall seconds |
|---|---|---|---:|
| check-optional-receiver | ^TestOptionalAfterCall79$ | True | 15.021 |
| javascript-required-panic | ^TestOptionalAfterCallRequired$ | True | 10.588 |
| native-required-panic | ^TestOptionalAfterCallRequired$ | True | 10.246 |
| drop-flow-throw-edge | ^TestOptionalAfterCallFlowEdges$ | True | 13.974 |
| drop-required-propagation | ^TestOptionalAfterCallPropagation$ | True | 14.694 |
| merged-computed-read | ^TestMergedFieldTypeScriptComputedChecked$ | True | 14.564 |
| merged-declaration | ^TestMergedFieldAdamicUnusedRefused$ | True | 11.292 |
| merged-read | ^TestMergedFieldTypeScriptMissingChecked$ | True | 14.861 |
| function-annotation | ^TestFunctionAdamicUnusedAnnotationRefused$ | True | 12.382 |
| function-call | ^TestFunctionTypeScriptCallRefused$ | True | 12.047 |
| function-length | ^TestFunctionTypeScriptLengthAgreesWithNode$ | True | 16.985 |
| erase-earlier-throw | ^TestOptionalAfterCallPreservesEarlierThrow$ | True | 8.054 |
| tsgo-absolute-main | ^TestTSGoCacheSanitized$ | True | 10.996 |
| tsgo-rewrite-runtime | ^TestTSGoCacheRelease$ | True | 11.381 |

The eight embedded source mutants are the six TestAssignmentProof*Mutant tests, TestCallSpreadEvaluatedTwiceMutant, and TestJSONStringifyShorthandWrongBindingMutant, covered by the full oracle suite. The final vet run passed after removing only completed task-owned reproducible build products to recover space. Two earlier vet attempts and the first linker-negative attempt hit ENOSPC during Go compilation, and are not credited. All 14 production mutants are caught after the six unfinished ones were retried; completed overlay archive variants were evicted from task scratch between those retries to bound disk use.

A separate linker-selection overlay omits only linkFlags from the resolved final checker link; any link failure is classified as structural/toolchain evidence, not a behavioral compiler mutant. Its successful negative retry fails TestTSGoCacheShippedRelease after 1.08s: GNU ld cannot load LLVMgold.so; the positive witness with main's lld selection passed.

Needed before a green submission: the markdown owner must address the cold setup/product grain and obtain a complete clean-tree run with adequate disk. Its four over-budget functions were not split or rerun in this lowering unit. No runtime clearance is needed for extra resolution lines, because there are none.

## Toolchain and limitations

GOPROXY was exported as https://proxy.golang.org|direct before bash cloud/setup.sh, then /workspace/adamic-tools/env.sh was sourced. nproc=5; cgroup quota is 400000/100000, four CPUs. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup succeeded; timing lines:

```text
submodule cache: off
v24.19.0
go version go1.27.1 linux/amd64
setup: node ready (0.034s)
setup: go ready (0.036s)
setup: submodules ready (0.075s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.013s
setup: markdown dependencies ready (0.090s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.242s)
setup: shared cache ready (0.719s)
setup: go build ready (51.290s)
setup: test binaries deferred (use --warm-tests) (51.375s)
setup: build cache warm (51.377s)
setup: build-flags commit=5a94465370ea0172e25aae8c13478e59e40c2bb4 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.00 0.00 1/138 840 load-after=2.72 0.74 0.25 2/166 1837
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (51.407s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.ktigud
```

No ahra CLI, Loom submission, PR, main push or force push was performed. No new code was copied from cohere. Full markdown validation and its cold setup budgets remain unproved/failed as stated above. Original source evidence is preserved under review/compiler/loom-r-slice-3-ac/ and review/compiler/loom-r-slice-3/.
