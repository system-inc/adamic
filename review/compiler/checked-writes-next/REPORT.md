Built step 10 checked wider writes as an independent branch on main, retaining its guards.
Code commits: cdea9b22 and 0972c62a; landed-main merge: ea6d8a1c; base: 3a391aff13465375512469aa74f32ea3e7dca8ec.
Proof: 471 checked, 18 proven, 26 refused; 105 runtime witnesses; full admission-delta pass.
Mutants: all 36 detected by the required exit-70 witness becoming a Node-agreeing success.
Limits: namespaces/06 and objects/19 remain NotYet; class optional boolean slots remain unsupported.

The source branch fork is 8cb5e7c1. Its own commits, replayed without integration merges, are:

- 6a093f03 Check wider TypeScript field writes against actual shape contracts
- 5603f5c5 Check diagnostic references against allocation contracts
- 528376cc Check writes into shared never arrays
- b94f41ee Use TypeScript fixtures for checked wider writes
- f84d5e45 Check FlowNode fields and direct container slots through wider TypeScript views
- 674b68b5 Check EmitNode intersections and resolve remaining source census matches
- c9514f34 Check boolean and conditional writes and overload results

Excluded integration merges: 4892bc1b and 830a46e1. The seven commits were replayed on their fork in an isolated worktree, then applied as one net patch to main. No delivery-branch, lowering-chain or side-stack history was imported. The initial base was 946a8f095a7fa419a92117406314b7b3d44630f0. At the final fetch, main advanced to 3a391aff13465375512469aa74f32ea3e7dca8ec by removing audit-approved redundant YAML test checks. Merge ea6d8a1c adopts that landed main update. It changes no compiler, runtime, executable fixture or count baseline, so the count and lower-shard measurements still apply. The full runtime/mutant suite and admission proof were repeated after the merge. Shared conflicts preserve record storage, graph types, V1/V2 view checks, unions, lazy initializer readiness, placeholders, generator refusal, exceptions and existing refusals. Runtime shape initializers gained a fifth contract-pointer field, including main's scalar/string pair shapes in library_object.c.

The final census checks all 603 source hashes from the pinned fixture record and reproduces the historical census byte for byte. None of 471, 18 or 26 moved. The 515 class-d rows are relation decisions on checker-rejected source, not 515 executable programs. Ordinary Load and Lower are explicitly disabled in that measurement overlay. The 105 executable reductions run through source Node, JavaScript, ASan/UBSan native and successful-run leak checks; bad inputs deliberately differ from Node by stopping at the checked write with exit 70. Adamic controls remain refused unless proven. See census.json and checked-proof-final.log.

The pinned source reconstruction ran every adapter through 75-optional-widening. Earlier interrupted reconstruction stopped before the readonly audit; completing only through writable views left types.ts mismatched because the pinned record also contains optional-widening changes. No recorded hash was changed. Final census: 15.70 seconds.

Mutants are listed below and in mutants.json. Each removes a store check, allocation marker, literal domain, nested contract or allocation proof; its witness must lose the required exit 70 and match Node. Native mutants execute uncached, and both generated backends and leak checks are exercised.

- TestCheckedWiderWriteMutants: drop_check, drop_literal_set
- TestCheckedNeverContractMutant: remove never allocation marker
- TestCheckedDiagnosticReferenceMutants: drop_alias_check, drop_nested_contract, drop_allocation_proof
- TestCheckedFlowContainerContractMutants: flow-node-misfit, container-alias-misfit, container-fill-misfit, container-boolean-misfit, container-map-object-misfit, container-nested-misfit, container-object-misfit, container-splice-misfit, container-map-misfit, container-string-misfit, container-number-misfit, flow-array-misfit, flow-undefined-misfit
- TestCheckedEmitContractsMutants: emit-drop-check, emit-resolution-misfit, emit-comment-misfit, emit-callback-misfit, emit-auto-misfit, emit-node-misfit
- TestCheckedViewsNextMutants: optional-boolean-misfit, branch-bottom-number-misfit, branch-file-misfit, branch-start-misfit, optional-false-misfit, branch-length-misfit, required-boolean-misfit, overload-callback-misfit, overload-result-misfit, branch-bottom-string-misfit, generic-site-misfit

The latent controls use the actual stock TypeScript 6.0.3 writers at witness pin 3255eb1e, with explicit pre-store checks. The tsc input driver finishes successfully: flags assignments 61, violations 0; parent assignments 0, violations 0. Both small flags/parent counterexamples exit 70. This is stock compiler execution on Node, not a native build of the whole stock compiler. See latent-controls-final.log and the retained report and witness outputs.

Gap records:

- cycles/05 and cycles/06 now compile. The existing fixture update runner compares each program to Node before writing its Compiles record, using sanitized native. cycles/06 agrees with Node's import-order failure rather than hiding it.
- namespaces/06 remains NotYet at 8:34, a value of type T. The old mutable-view refusal is gone; generic value lowering is absent on this main base.
- objects/19 remains NotYet at 87:33, a class method through a view that erases its prototype origin. The old intersection blocker is gone; the independent prototype guard stays intact.
- JSON emptyFallback, markdownblocks conditional empty array, and values empty-array default were already marked closed on main. Their existing closure tests pass. GAPS.md records this reproof. The values closed probes additionally hold the JavaScript backend to Node. JSON multiplePush/repeatInTry and values' optional boolean class field remain open, with their original assertions preserved.

node_buffer_bom.a was over-refused. Buffer<ArrayBuffer> versus Buffer<ArrayBufferLike> recursed into incompatible intrinsic toString overloads and compared a BufferEncoding parameter with a numeric parameter. Both symbol-provenance-verified Node Buffer types use the same represented byte slots; the structural overload comparison is bypassed only for that intrinsic pair, while nodeBufferUnsupportedUse still rejects unsupported members. All 14 Buffer refusal guards pass. The unchanged buffer program agrees with Node in sanitized native, JavaScript, WASI emission, WASI execution and the flow path graph. Its expected outcome remains compilation, not refusal.

Count attribution found and fixed a regression before delivery: fresh [] inside [[], [1]] or [[1], []] had acquired a retained never[] contract. Fresh allocations now take the enclosing best-common destination, consistently with existing elementType; explicit never[] aliases keep their checks. Both existing empty-first/empty-last fixtures agree with Node on both backends. All 36 mutants and all lower shards were rerun after this correction.

Final count regeneration succeeded in 79.34 seconds. It adds 109 rows (105 checked TypeScript witnesses and four proven Adamic controls), changes 160 existing rows, removes none. Every changed existing row has balanced retain/release increments only; allocations, frees, peak and regions are unchanged. Array.push now pins its receiver before evaluating the stored value. Each row is attributed in count-delta.json. Two cold go-test invocations hit their outer limits before completing. A preliminary successful regeneration exposed the nested-array regression; its table was replaced once after that correction. No counts were hand-edited to make a mismatch pass.

Admission-delta used the tool and generator from compiler/admission-delta 841e335c without importing that dependency's source into this branch. Base is current main above; proven source head is ea6d8a1c6ca6cce626a4de776d61badb3fd838fd. The subsequent delivery evidence/count commit changes no compiler or executable fixture source.

Admission verdict pass: 1029 distinct programs, all 15 changed .a inputs included, 1 newly admitted, 0 omitted, no filtering or sampling budget. proven-emit.a newly lowers and prints 8 with exit 0 on source Node, JavaScript and native. Every newly admitted program agrees. Total proof seconds: 58.910. See admission.json and admission-manifest.json.

Final commands (each command's output is a file in this directory):

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
# Successful setup retry: total 31.889 seconds; SDK setup total 120.937 seconds.
# nproc=5; cgroup quota=4 CPUs.
timeout 89 go test ./internal/oracle -run 'TestCheckedWider|TestCheckedViewsNext|TestCheckedNeverContractMutant|TestCheckedDiagnosticReferenceMutants|TestCheckedFlowContainerContractMutants|TestCheckedEmit' -count=1 -v -timeout 85s
timeout 89 go test ./cmd/adamic -run '^TestExplainCheckedWritesOutput$' -count=1 -v -timeout 85s
timeout 89 go test ./internal/native -run '^(TestRuntimeFieldLayoutsAreIncluded|TestRegExpIteratorResultShape)$' -count=1 -v -timeout 85s
timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 85s
timeout 89 go test ./stage1/cohere/values -run '^TestEachGapStandsWhereGapsMdSaysItDoes$' -count=1 -v -timeout 85s
# Other JSON, markdown and stage3 focused runs are recorded in gaps-records-final.log,
# closed-cycles.log and later-gap-records.log. Their selected test names and outputs are retained.
ADAMIC_ORACLE_WASI=1 timeout 89 go test ./internal/oracle -run '^TestWASIEmission$/^internal$/^oracle$/^testdata$/^node_buffer_bom.a$' -count=1 -v -timeout 85s
ADAMIC_ORACLE_WASI=1 timeout 89 go test ./internal/oracle -run '^TestWASIAgreesWithNode$/^shard-/^internal$/^oracle$/^testdata$/^node_buffer_bom.a$' -count=1 -v -timeout 85s
timeout 90 go test ./internal/oracle -c -o /tmp/checked-writes-oracle.test
# From internal/oracle, excluding compilation from the test deadline:
timeout 89 /tmp/checked-writes-oracle.test -test.run '^TestCountsAreRecorded$' -test.v -test.timeout 85s -update-counts
timeout 90 go test ./internal/lower -c -o /tmp/checked-writes-lower.test
# Every exact shard regex and command is in lower-shards.json; each has test.timeout=85s
# and subprocess timeout=89s, with two workers. No whole-package test or full gate ran.
CHECKED_WRITES_TREE=/workspace/checked-writes-adapted-complete CHECKED_WRITES_RECORDS=/workspace/writable-views-input/stage3/adapt/71-writable-views/language-decision/results.json CHECKED_WRITES_OUTPUT=/workspace/adamic/review/compiler/checked-writes-next/census.json timeout 89 go test -overlay /tmp/checked-writes-census-overlay/overlay.json ./internal/lower -run '^TestCheckedWritesCensus$' -count=1 -v -timeout 85s
python3 /tmp/checked-writes-manifest.py --sha HEAD --repository .
timeout 600 /tmp/checked-writes-admission-delta --base 3a391aff13465375512469aa74f32ea3e7dca8ec --head HEAD --manifest review/compiler/checked-writes-next/admission-manifest.json --manifest-generator cloud/admission-corpus/manifest.py --manifest-generator-revision origin/compiler/admission-delta --workers 4 --compile-timeout 45s --timeout 10s --json
timeout 89 go vet ./cmd/adamic ./internal/lower ./internal/native ./internal/oracle ./stage1/cohere/values
git fetch -q origin main devtools/fast-gate cloud/merge-tree
git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

Setup initially overlapped the one-shot conflict application and encountered conflict markers during its cold build. This was corrected and setup rerun successfully. Successful timing lines: Go 0.022s, Node 0.019s, submodules 0.056s, Markdown dependencies 0.060s, clang 0.140s, shared cache 0.798s, build 31.776s, total 31.889s. The tool environment is under /workspace/adamic-tools; /opt/adamic-tools was unavailable.

All new/touched checked-write leaves finish below 60 seconds; final runtime/mutant proof suite 12.006s, maximum leaf recorded in test-seconds.json; explanation suite 0.999s; runtime shape guards 0.522s; values gap suite 0.389s; reader guard 14.08s. Top-level test and leaf timings are in test-seconds.json. Lower shards:

- shard 00: pass, 2.276s
- shard 01: pass, 10.591s
- shard 02: pass, 2.749s
- shard 03: pass, 2.115s
- shard 04: pass, 1.533s
- shard 05: pass, 2.642s
- shard 06: pass, 3.070s
- shard 07: pass, 6.063s
- shard 08: pass, 5.652s
- shard 09: pass, 13.742s
- shard 10: pass, 8.202s
- shard 11: pass, 7.757s
- shard 12: pass, 4.893s
- shard 13: pass, 14.840s
- shard 14: pass, 7.359s
- shard 15: pass, 6.801s

Lane checks pass: gofmt and tools on 41 Go files, t.Parallel on five test packages, a-check on 15 .a files, vet on five packages. An earlier cold lane run skipped vet at its 10-second limit; the final warm lane run and separate bounded vet both pass. No runtime clearance is presumed. Changed runtime files requiring clearance:

- internal/native/runtime/adamic.h
- internal/native/runtime/array.c
- internal/native/runtime/directory.c
- internal/native/runtime/exceptions.c
- internal/native/runtime/input.c
- internal/native/runtime/library_object.c
- internal/native/runtime/map.c
- internal/native/runtime/map_set.c
- internal/native/runtime/node_crypto.c
- internal/native/runtime/node_fs_file.c
- internal/native/runtime/node_host.c
- internal/native/runtime/object.c
- internal/native/runtime/record.c
- internal/native/runtime/regexp.c
- internal/native/runtime/tsgo.c
