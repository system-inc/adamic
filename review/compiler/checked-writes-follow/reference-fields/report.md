Built step 10 reference-field checks for runtime-made objects and exact slot/value diagnostics; cleared branches untouched.
Commits: based on 2f5a3f9b on compiler/checked-writes-follow; delivery commit is recorded in the push response.
Commands and outputs: checked-write oracle PASS 29.597s; runtime suite PASS 14.290s; reader guard PASS 34.279s; counts PASS 41.409s.
Mutants: all 36 original, the prior literal mutant, and 15 new mutants caught (52 total).
Limits: runtime reference fallback proves broad strings and string arrays, recursively supported structures; no arbitrary literal/branded reference proof, full gate or whole-package test.

Runtime files changed this turn: internal/native/runtime/object.c and internal/native/runtime/directory.c. No other runtime file changed. Runtime-made fields use the actual references bitmap and heap representation before dereferencing; broad numeric/boolean proofs remain based on storage tags. fileStatus now supplies its numeric and boolean tags. String arrays are checked element by element and receive a persistent element contract, so later internal writers cannot invalidate the proof. Existing weaker nullable contracts cannot satisfy a stronger nonnullable contract. Array literal/branded element domains remain refused without directional allocation proof. JavaScript implements the same representation checks and persistent array contracts.

Evidence per example (all use real lowering, not synthetic positive IR; Node stdout is pinned to prevent vacuous success):

- regex: runtime matchAll next result with value string/undefined array; TypeScript adapter agrees with Node in JavaScript, native release and ASAN/UBSAN with no leaks; .a refusal regex.a:5:101, readonly done becomes writable; fix keep done readonly or copy.
- Map: runtime values iterator next result with string value; all three backends agree with Node; .a refusal map.a:5:76, readonly done becomes writable; fix keep done readonly or copy.
- Set: runtime values iterator next result with string value; all three backends agree with Node; .a refusal set.a:5:76, readonly done becomes writable; fix keep done readonly or copy.
- fs: readTextFile result with text reference; all three backends agree with Node and print alpha; .a refusal fs.a:7:35, readonly text becomes writable; fix keep text readonly or copy.
- directory: readDirectory result with names string array; all three backends agree with Node and print alpha.txt; .a refusal directory.a:7:47, readonly names becomes writable; fix keep names readonly or copy.
- file status: fileStatus result with type string, size number and symbolicLink boolean; all three backends agree with Node and print file, 6, false; .a refusal status.a:7:72, readonly type becomes writable; fix keep type readonly or copy.

Each committed source is .a and retains its required refusal header. The test temporarily materializes the same source at the TypeScript-mode frontend boundary, with a readonly receiving interface array exposed to a writable view, forcing the checked-write path. The .a refusal is the intended ruling for that view; it is separately asserted. The previously cleared runtime rejects the fitting TypeScript writes with exit 70. Revert mutants pin that regression. The later-writer witness injects an unsafe internal IR writer and requires exit 70 in both backends, including native sanitizers; it is a negative internal-writer test, not a claim that Node rejects the fitting source.

Diagnostic correction: the requested expected-type substitutions reverse the declarations in the actual upstream sources. 02 allocates never[], 03/16 allocate FlowArrayMutation.node: BinaryExpression, 13 allocates Declaration[], and 15 allocates DiagnosticWithDetachedLocation.file: undefined. number, BindingElement, Node and SourceFile describe the incoming writes. Renaming expects to those incoming types would misdescribe the unchanged check. Exact pins therefore retain the true declared slot type and report the actual value's type:

```
adamic: panic: write failed: array[] expects never, got 1 (number)
adamic: panic: write failed: flow.node expects BinaryExpression, got BindingElement
adamic: panic: write failed: array[] expects Declaration, got Node
adamic: panic: write failed: diagnostic.file expects undefined, got SourceFile
adamic: panic: write failed: flow.node expects BinaryExpression, got BindingElement
```

Allocation-name metadata is diagnostic only; check kinds, allowed values and directional proof tables are unchanged by the message work. All five source fixtures are copied from d339be83 without alteration of their declarations and retain their .a refusal. Both backends stop before the write, with stdout exactly before write and stderr pinned above. Four message-only IR mutants substitute the wider view declaration and fail exact stderr in all backends; separate native/JS formatter mutants restore number instead of never for 02. 31 prior diagnostic golden rows changed only their got type or numeric suffix; diagnostic-row-changes.txt attributes each.

Verification commands (source /workspace/adamic-tools/env.sh first; output directly captured in named logs):

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
nproc
 timeout 89 go test ./internal/oracle -run 'TestCheckedWider|TestCheckedViewsNext|TestCheckedNeverContractMutant|TestCheckedDiagnosticReferenceMutants|TestCheckedFlowContainerContractMutants|TestCheckedEmit|TestCheckedWriteMessage|TestCheckedRuntime' -count=1 -v -timeout=85s
 timeout 89 go test ./internal/native -run 'TestRuntime|TestTypedArrayRuntime|TestNodeBufferRuntime|TestRegExpIteratorResultShape|TestFreedValuesAreCaughtWithSlabs|TestSizeClassesShareTheirChunks' -count=1 -v -timeout=85s
 timeout 89 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout=85s
 timeout 89 go test -overlay review/compiler/checked-writes-follow/literal-mutant-overlay.json ./internal/native -run '^TestRuntimeCheckedLibraryPushRejectsLiteral$' -count=1 -v -timeout=85s
```

The oracle selection reran the original 105 runtime cases and all original 36 mutants. Every new permanent test leaf is below 60 seconds. Positive runtime-field leaves: regex 1.73s, Map 1.19s, Set 1.48s, fs 1.22s, directory 1.14s, status 1.01s; later-writer .98s. Message leaves 02 .68s, 03 .59s, 13 .81s, 15 1.23s, 16 1.61s. Message mutant leaves 03 1.48s, 13 .70s, 15 .82s, 16 .68s. Runtime probe timings and remaining test leaves are in runtime-final.log. Setup: Go .022s, Node .023s, submodules .056s, markdown .069s, clang .158s, shared cache .914s, Go build 20.662s, total 20.796s; nproc 5 with a four-CPU cgroup quota.

Counts were regenerated once after measuring current runtime rows. A first unsplit cold count pass hit its hard limit before writing. Temporary review-only overlays warmed 932 normal/input/fs rows in 24 shards (maximum wall 17.408s), plus the special callback/interface/checked-write rows in a bounded extra leaf (67.49s, review-only; no permanent test added). The final TestCountsAreRecorded uses those freshly measured cache entries and retains the real update-table writer:

```
 timeout 89 go test -overlay review/compiler/checked-writes-follow/reference-fields/counts-shards.json ./internal/oracle -run '^TestCheckedCountsWarmShard$' -count=1 -v -timeout=85s -args -checked-count-shard=<0..23>
 timeout 89 go test -overlay review/compiler/checked-writes-follow/reference-fields/counts-shards.json ./internal/oracle -run '^TestCheckedCountsWarmExtra$' -count=1 -v -timeout=85s
 timeout 89 go test -overlay review/compiler/checked-writes-follow/reference-fields/counts-cached-final.json ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout=85s -args -update-counts
```

counts.md adds 11 measured TypeScript adapter rows (six fitting reference witnesses and five message witnesses). 99 prior rows change only retains/releases, attributed individually in counts-attribution.json to the previous follow-up's removal of the duplicate ArrayPush receiver hold; allocations, frees, peak live and regions are unchanged in those rows. The prior turn had not regenerated these rows. No weakening of expected outcomes or leak checks.

Hot path: 10M pushes, release -O2, Python time.perf_counter_ns wall clock around subprocess.run, one warmup and rotated order, best of five. Main 5e33a17b: 65.942251 ms; previous follow 2f5a3f9b: 68.336267 ms; this head: 67.594049 ms. Head +2.50% versus main and -1.09% versus the cleared follow in the same cohort; the earlier published follow measurement was +5.14%, so this run shows no regression, not an exact reproduction of its noise. AMD EPYC 9V74, Linux 6.18.44, nproc 5, CPU quota 400000/100000; load before/after [1.0562, 2.7949, 2.1621]. All runs are in benchmark.json. No hot-path writer or receiver-hold change in this turn.

Every original mutant (PASS means caught):

- TestCheckedNeverContractMutant: caught (0.87s).
- TestCheckedViewsNextMutants/branch-bottom-number-misfit: caught (0.81s).
- TestCheckedViewsNextMutants/optional-boolean-misfit: caught (1.05s).
- TestCheckedViewsNextMutants/overload-callback-misfit: caught (1.06s).
- TestCheckedViewsNextMutants/overload-result-misfit: caught (1.02s).
- TestCheckedViewsNextMutants/generic-site-misfit: caught (0.99s).
- TestCheckedViewsNextMutants/branch-bottom-string-misfit: caught (0.93s).
- TestCheckedViewsNextMutants/branch-length-misfit: caught (0.88s).
- TestCheckedViewsNextMutants/branch-file-misfit: caught (1.13s).
- TestCheckedViewsNextMutants/branch-start-misfit: caught (1.01s).
- TestCheckedViewsNextMutants/required-boolean-misfit: caught (1.02s).
- TestCheckedViewsNextMutants/optional-false-misfit: caught (1.17s).
- TestCheckedFlowContainerContractMutants/container-object-misfit: caught (0.89s).
- TestCheckedFlowContainerContractMutants/flow-node-misfit: caught (1.21s).
- TestCheckedFlowContainerContractMutants/container-alias-misfit: caught (0.78s).
- TestCheckedFlowContainerContractMutants/container-fill-misfit: caught (0.78s).
- TestCheckedFlowContainerContractMutants/container-boolean-misfit: caught (0.76s).
- TestCheckedFlowContainerContractMutants/container-map-object-misfit: caught (0.90s).
- TestCheckedFlowContainerContractMutants/container-nested-misfit: caught (0.69s).
- TestCheckedFlowContainerContractMutants/container-string-misfit: caught (0.88s).
- TestCheckedFlowContainerContractMutants/container-splice-misfit: caught (0.88s).
- TestCheckedFlowContainerContractMutants/flow-array-misfit: caught (0.84s).
- TestCheckedFlowContainerContractMutants/container-number-misfit: caught (0.81s).
- TestCheckedFlowContainerContractMutants/container-map-misfit: caught (1.01s).
- TestCheckedFlowContainerContractMutants/flow-undefined-misfit: caught (0.85s).
- TestCheckedDiagnosticReferenceMutants/drop_nested_contract: caught (0.87s).
- TestCheckedDiagnosticReferenceMutants/drop_alias_check: caught (0.99s).
- TestCheckedDiagnosticReferenceMutants/drop_allocation_proof: caught (1.04s).
- TestCheckedWiderWriteMutants/drop_literal_set: caught (0.62s).
- TestCheckedWiderWriteMutants/drop_check: caught (0.70s).

- TestCheckedEmitContractsMutants/emit-drop-check: caught (1.03s).
- TestCheckedEmitContractsMutants/emit-callback-misfit: caught (0.99s).
- TestCheckedEmitContractsMutants/emit-resolution-misfit: caught (0.87s).
- TestCheckedEmitContractsMutants/emit-comment-misfit: caught (1.16s).
- TestCheckedEmitContractsMutants/emit-node-misfit: caught (1.00s).
- TestCheckedEmitContractsMutants/emit-auto-misfit: caught (1.05s).

Prior literal mutant: TestRuntimeCheckedLibraryPushRejectsLiteral caught the exit-0 regression (original-literal-mutant.log).

New overlay mutants:

- revert-native-reference-fields: caught by ^TestCheckedRuntime (33.713s); see revert-native-reference-fields.log.
- revert-js-reference-fields: caught by ^TestCheckedRuntime (15.897s); see revert-js-reference-fields.log.
- revert-array-field-schema: caught by ^TestCheckedRuntime(Regex|Directory)Fields$ (22.486s); see revert-array-field-schema.log.
- drop-runtime-array-contract: caught by TestRuntimeCheckedReferenceArrayKeepsContract (15.803s); see drop-runtime-array-contract.log.
- drop-string-heap-kind: caught by TestRuntimeCheckedReferenceArrayKeepsContract (14.579s); see drop-string-heap-kind.log.
- accept-runtime-array-literal: caught by TestRuntimeCheckedReferenceArrayRejectsLiteral (14.809s); see accept-runtime-array-literal.log.
- accept-weaker-array-contract: caught by TestRuntimeCheckedReferenceArrayRejectsWeakerContract (14.184s); see accept-weaker-array-contract.log.
- restore-number-view-native: caught by ^TestCheckedWriteMessage02$ (26.68s); see restore-number-view-native.log.
- restore-number-view-js: caught by ^TestCheckedWriteMessage02$ (14.498s); see restore-number-view-js.log.
- drop-js-runtime-array-contract: caught by ^TestCheckedRuntimeReferenceArrayLaterWriter$ (19.452s); see drop-js-runtime-array-contract.log.
- drop-js-string-heap-kind: caught by ^TestCheckedRuntimeReferenceArrayLaterWriter$ (14.056s); see drop-js-string-heap-kind.log.

Four additional message-only mutants: TestCheckedWriteMessageMutant03, 13, 15 and 16; each exact stderr comparison caught the wider-view declaration in JavaScript, native release and sanitized native (oracle-final.log).

Lane checks after implementation commit 06953dec: PASS in 9.0s; gofmt and tools on 45 Go files, t.Parallel on five test packages, a-check on 16 .a files, vet on five packages. Output is lane-checks.log. Push and delivery SHA are recorded in the final response.
