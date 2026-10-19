Re-cut Views V4 onto current main; preserve callback adaptation, checked finite reads, async equality, and per-call counts.
Source 82a5dffd7654f3eb2cb4652a12b99401210f442a; tested source a93a7bdac8efa6d0079f27a715e368425cf94da8; base 561aeb4b851ca0820f39440a09dc5fb3fd27ba40.
Full packages executed and failed catchers repaired/rerun; full five-way JSON PASS on both trees: base 318.133 s, V4 336.990 s; guards, vet, gofmt and 55-file a-check pass.
Mutants caught: headers by a-check; computed metadata by refusal guard; domains/union reads by oracle; async identity by async oracle; broad count by NoReader; removed storage by statics census; original executable cache/adapter mutants by native and V4 tests.
Not covered: seven source-pending capability/runtime leaves, higher than five-way parallelism, or the missing remote failure text. The first-green 30-minute push target was missed; all failed attempts are retained.

Branch: `compiler/loom-r-v4-2-ac-m`. The resulting full commit SHA is in the submission response (a commit cannot contain its own SHA).

The merge used `git merge-tree --write-tree --merge-base=82a5dffd7654f3eb2cb4652a12b99401210f442a~1 origin/main 82a5dffd7654f3eb2cb4652a12b99401210f442a`. Its conflict tree was materialized on a branch at the fetched main, and resolved as follows. The exact three sides are recorded in conflicts.json.

- internal/lower/object.go, arrayMethodArguments/map: preserve main's callbackType, known-library conditional and adaptArrayCallback; add V4's ViewRead, using the adapted CallbackType.
- internal/lower/object.go, arrayVisit: preserve main's element/index/array callback adaptation and resulting CallbackType; add V4's ViewRead.
- internal/lower/object.go, arrayReduce: preserve main's accumulator/element/index/array callback adaptation and resulting CallbackType; add V4's ViewRead.

Main's newer adaptation came from slice 2, 67733635b0252eb9d3bb3645c72262c7b18361e1. It captures callbacks once and adapts parameter slots/representation. All three merged paths retain that behavior and layer V4 reads over it.

Runtime clearance: no extra C/H lines. The original and re-cut own patches each contain exactly the same 452 added/deleted lines in the same 14 runtime files. runtime-delta.json records an empty differences object; full patches are retained.

Toolchain: Linux, clang 20.1.8, Go 1.27.1, Node 24.19.0. nproc=5, but cpu.max=400000/100000 (4 CPUs); GOMAXPROCS=4 and test parallelism=4. The first setup ran while merge markers remained and failed Go compilation with the recorded syntax diagnostics; after resolving, setup succeeded. Full timing lines are below.

```
setup: go ready (0.020s)
setup: node ready (0.020s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.005s
setup: submodules ready (0.057s)
setup: markdown dependencies ready (0.058s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.153s)
setup: shared cache ready (9.674s)
setup: go build ready (43.997s)
setup: test binaries deferred (use --warm-tests) (44.078s)
setup: build cache warm (44.079s)
setup: build-flags commit=561aeb4b851ca0820f39440a09dc5fb3fd27ba40 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.68 0.20 0.06 1/180 89475 load-after=4.52 1.28 0.44 1/185 91821
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (44.105s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.rOBbEr
```

Counts: restored the main table, then ran `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -parallel=4 -timeout=30m -args -update-counts` initially (PASS, 79.059 seconds), then regenerated after the integration fixture changes (PASS, 103.259 seconds). The final diff is unchanged by that migration. An anchored invocation selected no shards and failed without measuring or writing rows; the successful invocation uses the documented unanchored family prefix. Actual main has the six measurement columns below, not an older digest format. No hand edits. 26 rows added, 3 removed, 1 modified. Full before/after rows:

| Path | Before | After |
|---|---|---|
| internal/oracle/testdata/library_array_holes_length.a | None |  /  internal/oracle/testdata/library_array_holes_length.a  /  17  /  17  /  3  /  19  /  6  /  0  /  |
| internal/oracle/testdata/library_array_holes_range.a | None |  /  internal/oracle/testdata/library_array_holes_range.a  /  24  /  24  /  12  /  27  /  4  /  0  /  |
| internal/oracle/testdata/library_array_holes_scanner_probe.a | None |  /  internal/oracle/testdata/library_array_holes_scanner_probe.a  /  4  /  4  /  0  /  3  /  4  /  0  /  |
| internal/oracle/testdata/namespace_method_receiver.a |  /  internal/oracle/testdata/namespace_method_receiver.a  /  7  /  7  /  9  /  16  /  5  /  0  /  |  /  internal/oracle/testdata/namespace_method_receiver.a  /  7  /  7  /  7  /  14  /  5  /  0  /  |
| stage3/interface-downcasts/lane2/array-alias.a | None |  /  stage3/interface-downcasts/lane2/array-alias.a  /  3  /  1  /  9  /  8  /  3  /  0  /  |
| stage3/interface-downcasts/lane2/array-boolean.a | None |  /  stage3/interface-downcasts/lane2/array-boolean.a  /  2  /  0  /  4  /  3  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/array-bounds.a | None |  /  stage3/interface-downcasts/lane2/array-bounds.a  /  5  /  5  /  5  /  10  /  5  /  0  /  |
| stage3/interface-downcasts/lane2/array-element.a | None |  /  stage3/interface-downcasts/lane2/array-element.a  /  4  /  4  /  4  /  7  /  4  /  0  /  |
| stage3/interface-downcasts/lane2/array-field.a | None |  /  stage3/interface-downcasts/lane2/array-field.a  /  4  /  4  /  3  /  6  /  4  /  0  /  |
| stage3/interface-downcasts/lane2/array-holes-view.a | None |  /  stage3/interface-downcasts/lane2/array-holes-view.a  /  7  /  7  /  9  /  15  /  5  /  0  /  |
| stage3/interface-downcasts/lane2/array-iteration-bad.a | None |  /  stage3/interface-downcasts/lane2/array-iteration-bad.a  /  2  /  0  /  4  /  2  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/array-iteration.a | None |  /  stage3/interface-downcasts/lane2/array-iteration.a  /  6  /  6  /  3  /  8  /  4  /  0  /  |
| stage3/interface-downcasts/lane2/array-lazy.a | None |  /  stage3/interface-downcasts/lane2/array-lazy.a  /  4  /  4  /  5  /  7  /  4  /  0  /  |
| stage3/interface-downcasts/lane2/array-length-kind.a | None |  /  stage3/interface-downcasts/lane2/array-length-kind.a  /  2  /  0  /  2  /  1  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/array-map-bad.a | None |  /  stage3/interface-downcasts/lane2/array-map-bad.a  /  4  /  0  /  4  /  1  /  4  /  0  /  |
| stage3/interface-downcasts/lane2/array-map.a | None |  /  stage3/interface-downcasts/lane2/array-map.a  /  7  /  7  /  5  /  12  /  6  /  0  /  |
| stage3/interface-downcasts/lane2/array-missing.a | None |  /  stage3/interface-downcasts/lane2/array-missing.a  /  1  /  0  /  2  /  1  /  1  /  0  /  |
| stage3/interface-downcasts/lane2/array-pop.a | None |  /  stage3/interface-downcasts/lane2/array-pop.a  /  6  /  6  /  5  /  11  /  6  /  0  /  |
| stage3/interface-downcasts/lane2/array-second.a | None |  /  stage3/interface-downcasts/lane2/array-second.a  /  2  /  0  /  6  /  3  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/array-string-literal.a | None |  /  stage3/interface-downcasts/lane2/array-string-literal.a  /  2  /  0  /  4  /  2  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/array-undefined.a | None |  /  stage3/interface-downcasts/lane2/array-undefined.a  /  2  /  0  /  4  /  1  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/generic-array-element.a | None |  /  stage3/interface-downcasts/lane2/generic-array-element.a  /  2  /  1  /  1  /  2  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/non-array.a | None |  /  stage3/interface-downcasts/lane2/non-array.a  /  1  /  0  /  2  /  1  /  1  /  0  /  |
| stage3/interface-downcasts/lane2/nullable-array-element.a | None |  /  stage3/interface-downcasts/lane2/nullable-array-element.a  /  1  /  0  /  4  /  2  /  1  /  0  /  |
| stage3/interface-downcasts/lane2/nullable-array-missing.a | None |  /  stage3/interface-downcasts/lane2/nullable-array-missing.a  /  0  /  0  /  1  /  1  /  0  /  0  /  |
| stage3/interface-downcasts/lane2/nullable-array-string.a | None |  /  stage3/interface-downcasts/lane2/nullable-array-string.a  /  2  /  0  /  4  /  2  /  2  /  0  /  |
| stage3/interface-downcasts/lane2/object-element.a | None |  /  stage3/interface-downcasts/lane2/object-element.a  /  3  /  3  /  9  /  10  /  3  /  0  /  |
| stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a |  /  stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a  /  3  /  0  /  6  /  4  /  3  /  0  /  | None |
| stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a |  /  stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a  /  2  /  0  /  5  /  4  /  2  /  0  /  | None |
| stage3/interface-downcasts/v2/callable-producer-wrong.a |  /  stage3/interface-downcasts/v2/callable-producer-wrong.a  /  2  /  0  /  5  /  4  /  2  /  0  /  | None |

A-check: all 55 changed .a files compiled with this branch compiler; 44 checked, 1 NotYet, 9 refused, 1 TS7009 type error. All matched headers. The p17 review witness gains its measured refusal header; its body is unchanged. No pinned line expectations named these review paths. The unchanged pinned Gate.aCheck from fbac28c62493f27a788edc02a18bc8edb68de5da was executed as the README prescribes; actual diagnostics were collected first. Both header mutants failed; see a-header-mutants.json. No claim of a whole-tree a-check scan.

| Program | Status |
|---|---|
| internal/oracle/testdata/library_array_holes_length.a | checked |
| internal/oracle/testdata/library_array_holes_range.a | checked |
| internal/oracle/testdata/library_array_holes_scanner_probe.a | checked |
| internal/oracle/testdata/review/agree/fxspptb_oct9_views_p64_array_arm_element.a | checked |
| internal/oracle/testdata/review/refused/fxspptb_oct9_views_p17_callable_field_element_call.a | refused |
| internal/oracle/testdata/views_v4_checked/fxspptb_oct9_views_p07_element_access_union_wrong.a | checked |
| internal/oracle/testdata/views_v4_checked/fxspptb_oct9_views_p49_union_destructure_wrong.a | checked |
| internal/oracle/testdata/views_v4_checked/fxspptb_oct9_views_p75_callable_from_map_value.a | checked |
| internal/oracle/testdata/views_v4_refused/construct-no-signature.a | type error |
| internal/oracle/testdata/views_v4_refused/factory-165-wrong-arity.a | refused |
| internal/oracle/testdata/views_v4_refused/view-parameter-wrong-members.a | refused |
| internal/oracle/testdata/views_v4_refused/view-parameter-wrong.a | refused |
| internal/oracle/testdata/views_v4_refused/view-result-wrong-members.a | refused |
| internal/oracle/testdata/views_v4_refused/view-result-wrong.a | refused |
| review/compiler/views-v4/construct-misfit.a | NotYet |
| review/compiler/views-v4/lane5/factory-12-wrong-producer.a | checked |
| review/compiler/views-v4/lane5/factory-174-wrong-producer.a | checked |
| review/compiler/views-v4/lane5/factory-186-wrong-overload.a | checked |
| review/compiler/views-v4/lane5/factory-189-wrong-overload.a | checked |
| review/compiler/views-v4/lane5/factory-24-wrong-overload.a | checked |
| review/compiler/views-v4/lane5/factory-3-wrong-overload.a | checked |
| review/compiler/views-v4/lane5/factory-3-wrong-result.a | checked |
| review/compiler/views-v4/lane5/factory-9-wrong-overload.a | checked |
| stage3/interface-downcasts/lane2/array-alias.a | checked |
| stage3/interface-downcasts/lane2/array-boolean.a | checked |
| stage3/interface-downcasts/lane2/array-bounds.a | checked |
| stage3/interface-downcasts/lane2/array-element.a | checked |
| stage3/interface-downcasts/lane2/array-field.a | checked |
| stage3/interface-downcasts/lane2/array-holes-view.a | checked |
| stage3/interface-downcasts/lane2/array-iteration-bad.a | checked |
| stage3/interface-downcasts/lane2/array-iteration.a | checked |
| stage3/interface-downcasts/lane2/array-lazy.a | checked |
| stage3/interface-downcasts/lane2/array-length-kind.a | checked |
| stage3/interface-downcasts/lane2/array-map-bad.a | checked |
| stage3/interface-downcasts/lane2/array-map.a | checked |
| stage3/interface-downcasts/lane2/array-missing.a | checked |
| stage3/interface-downcasts/lane2/array-pop.a | checked |
| stage3/interface-downcasts/lane2/array-second.a | checked |
| stage3/interface-downcasts/lane2/array-string-literal.a | checked |
| stage3/interface-downcasts/lane2/array-undefined.a | checked |
| stage3/interface-downcasts/lane2/generic-array-element.a | checked |
| stage3/interface-downcasts/lane2/non-array.a | checked |
| stage3/interface-downcasts/lane2/nullable-array-element.a | checked |
| stage3/interface-downcasts/lane2/nullable-array-missing.a | checked |
| stage3/interface-downcasts/lane2/nullable-array-string.a | checked |
| stage3/interface-downcasts/lane2/object-element.a | checked |
| stage3/interface-downcasts/untagged/fixtures/array-union-empty.a | checked |
| stage3/interface-downcasts/untagged/fixtures/array-union-good.a | checked |
| stage3/interface-downcasts/untagged/fixtures/array-union-mixed.a | checked |
| stage3/interface-downcasts/untagged/fixtures/array-union-nested.a | checked |
| stage3/interface-downcasts/untagged/fixtures/array-union-wrong.a | checked |
| stage3/interface-downcasts/untagged/fixtures/callable-union-nested.a | refused |
| stage3/interface-downcasts/untagged/fixtures/callable-union-wrong.a | refused |
| stage3/interface-downcasts/v2/array-arm.a | checked |
| stage3/interface-downcasts/v2/callable-producer-wrong.a | refused |

Verification: compilation-only checks passed all ten packages; go vet passed all ten, gofmt reported no changed Go files requiring formatting. Naming/parallel guards and C compiler census passed. Package result details are appended after completion.

Infrastructure observation: the first full-package pass exhausted this 32 GB filesystem during IR export archive products. IR/lower and the initially started native run reported ENOSPC. 8,633,148,828 bytes of old rebuildable Go cache blobs were removed, preserving sources and evidence; missing blobs are ordinary cache misses per cmd/adamic-gocacheprog. Failed initial logs remain separate from retries.

Necessary integration found by the full package run:

- Main's review refusal family contained three programs V4 admits. The initial oracle run recorded their exact accepted-refusal failures, then was deliberately terminated after identifying those failures so integration could be corrected; it is not a completed successful run.
- p17 (`view['compute'](41)`) exposed a real missing metadata boundary in main's finite-key selection helper. The pre-fix executable check produced ASan SEGV instead of a checked failure. `finiteElementRead` now records each selected property's view metadata and treats a callable returned from the synthetic helper as an escaping value. It reuses V4's escape proof/adapter handling through `prepareViewCallableEscapingValue`, extracted without changing the existing named-read classification. Conservative assumption: in .a, the helper return needs the existing producer-domain escape proof, so this incompatible bracket call remains refused rather than extending the direct-call ABI across an unproven returned closure. Its new first line is `// a-check: refused an unproven escaping callable read of view['compute']`; original body bytes stay intact.
- p64 (the supported array arm) moves, body unchanged, from review/refused to review/agree, so the existing sharded review oracle holds it to Node and both backends.
- p75 (a callable from a Map value) moves, body unchanged, from review/refused to testdata/views_v4_checked. `TestV4ReviewMapValueCallChecksProducerArgument` reads the actual .a fixture, pins Node's `undefined\n`, and requires native/release, ASan/UBSan and JavaScript to stop before entering the string producer with exit70 and the producer-string/view-number argument diagnostic. This is V4's already documented deliberate checked-view divergence, not an output agreement claim.
- The p75 producer-domain-omission IR mutant is caught in all three executions: native signal, sanitized ASan SEGV, JavaScript's later result-domain failure instead of the required argument-domain failure. These are mutant outcomes only; the production case is sanitizer-clean.
- A separate Go-overlay mutant removes all the new computed-property metadata, restoring main's exact pre-integration element_access_fields.go. `TestReviewProgramsRefuse/fxspptb_oct9_views_p17_callable_field_element_call.a` rejects it with `accepted refused review program ...` (exit1); production files are untouched by the overlay. See computed-metadata-mutant.jsonl and its .go.txt source.

JSON preparation observation: the initial full JSON package failed before assertions because `TestProduct_JSONPinnedCorpus` rejected a pinned repository record as naming `/workspace/adamic` without a matching declared tracked input. The exact missing tokens came from `stage1/cohere/lint/rules/no-caller/evidence/options-overlay.json` at JSON pin `2b4dbde172c1c2f23c065ecae8b906f8fcb72a20`. Main's `8f287383f0a3718ab6d50f0ce28e252d1a962a7d` removed that report. Neither the pinned-corpus recipe nor the relocation rule is modified by this re-cut. `pinned-json-provenance.json` retains that exact historical JSON from git, byte for byte, as source evidence. Main's existing recipe declares tracked JSON, so this provenance is keyed input for the historical bytes the product actually carries, allowing the unchanged whole-token guard to verify their origin. No compiler/output comparison had run in the failed preparation; all dependent shard failures were `shared preparation failed`.

Cache cleanup also removed another 4,312,406,592 bytes of old Go cache blobs and five obsolete export-archive products totaling 9,366,137,329 bytes. Current export products, sources and evidence remain. Cleanup manifests record every removed object.

Timing observation: three unchanged main runtimeproducts census leaves took 97.51, 97.65 and 97.70 seconds while rebuilding export archives. The unchanged native split-TSGo unit recorded a cold product-fetch phase of 130.512 seconds, followed by a 0.064-second whole-checker run. Counts and the separate overlay-mutant compilation overlapped part of this native cold setup on the 4-CPU cgroup; these are observed local cold-run timings, not isolated steady-state measurements. No main test was edited or split to conceal them. No claim that every pre-existing cold product leaf fits 60 seconds here.

GCC portability observation: one of the completed native run failures was `TestLongUnicodePortabilityGCC` (0.81 s). Exact diagnostic: `host_runtime.c:29:10: fatal error: sanitizer/allocator_interface.h: No such file or directory`. This box has GCC 14.2.0; its sanitizer include directory lacks allocator_interface.h, while house clang 20's resource directory has it. The header's queried allocator symbols exist in this box's libasan. host_runtime.c is byte-identical to main's tip commit 561aeb4b, which added that include; V4 makes no change to that file. This is an observed missing-header/toolchain failure, not an observed output assertion. No runtime lines or tests were changed to bypass it. The system include directory is not writable and sudo is absent; no unkeyed include-path wrapper or header stub was introduced. The inference is that main's new host runtime exposes this GCC installation gap independently of V4; this report does not claim a separately executed whole-base package run.

The a-check file selector was corrected from AM to AMR so all four renamed destination files are included. The final scan is 52, not the intermediate scan's 48. Both header mutants still fail the exact predicate.

Two further main review records expected erased-cast output for wrong untagged-union members: p07 (bracket read) and p49 (destructured read). V4's existing union membership checks now trap both, so their unchanged bodies move from review/agree into views_v4_checked. Two new top-level tests observe Node's `true\n` separately and require the documented checked-view divergence: exit70, no stdout, `matches no member of Target`, sanitizer-clean, on native/release, sanitized native, and JavaScript. Each property-check-omission mutant instead prints Node's true on every backend and remains sanitizer/leak-clean. p07's finite-selection helper contains two copies of its field read (the branch and unreachable final return); the mutant checks it removed exactly two, and exactly one for p49. No existing review predicate is weakened.

GCC repair completed: an unchanged copy of /usr/bin/gcc-14 was installed in a private relocatable prefix, with its existing libraries and libexec helpers linked to the original installation. Its sanitizer include directory adds LLVM 20.1.8's unchanged public allocator_interface.h (all other headers are the original GCC installation). There is no compiler wrapper, injected flag, runtime source edit or header stub. An initial private-prefix retry lacked cc1; adding the original libexec link repaired it. `TestLongUnicodePortabilityGCC` then passed in 11.777 seconds, including both backends against Node with ASan/UBSan/LSan. No system files were changed. gcc-header-repair.json and the failed/successful retry logs retain the details.

Stable product preparation: mutable JSON evidence was temporarily kept out of the index while tests ran, because main's pinned-corpus recipe declares every tracked JSON. Updating a tracked package-results.json during the next test rekeys that product. The initial JSON retry consequently reported missing preparation; its 2,048 shard failures were all `shared preparation failed`, not output mismatches. With immutable provenance alone indexed, the corpus, compiler-free oracle products, and lowered/release/sanitized port products all settled successfully. This changes no recipe or rule. Final evidence is staged after collection.

Full-CPU comparison requested during this unit: use the affinity/nproc count of 5 for both GOMAXPROCS and -parallel. cpu.max still gives a four-core quota. Each complete JSON-package command runs alone after the other suites stop. The base worktree is exactly 561aeb4b851ca0820f39440a09dc5fb3fd27ba40 with the identical pinned cohere and TypeScript commits; dependency node_modules are shared from the setup installation. The cases remain the fixed 3,919-input corpus. Record memory.current, anon/file components, sum of descendant process RSS at 50 ms intervals, and memory.events before/after. Cgroup usage includes reclaimable file cache; descendant RSS is a sum, not proportional memory. Results follow below when complete.

Additional completed-native-log findings (the earlier one-failure sample was incomplete):

- Ten V4 cache fixture leaves failed to link main's new heap accounting symbols. The fixture now compiles main's actual host_runtime.c and parallel.c together with its original closure/heap/weak/count units; function sections discard unrelated host APIs. Its existing abort-on-unrelated-string destructor remains, with the real literal identity marker supplied as one byte in test support. Real backing accounting and shutdown are linked, not stubbed. All ten cache leaves and their original runtime mutants pass afterward (native-cache-final.jsonl, package 4.252 s).
- TestRuntimeStaticsAreListed named five new runtime files and eight storage entries. storage.txt now audits only V4's own additions. RangeError strings are immortal and immutable under main's string-index rules; the five adapter-cache globals are protected by their existing mutex. The scanner's adamic_array forward-tag finding is explicitly identified as a tag, not storage. This inventory does not claim that the source's pending pool/TSan/region conditions are implemented. Removing view_adapter_used from this inventory makes the unchanged guard fail (storage-audit-mutant.jsonl); inventory restored.
- TestNoReaderCallingConvention found V4's global `|| ClosureConventionNeeded()` fallback sending an unrelated function type through counted dispatch. The fallback now uses main's PackedCountNeededFromCall alone, after V4's adapter/receiver branches, preserving per-call selection. The unchanged guard and closure convention tests pass (1.220 s); a Go-overlay mutant restoring the broad fallback fails that guard. All V4-focused oracle tests then pass: 279 pass events, package 40.176 s. No runtime C/H bytes change.
- The WASI split binding test required the optional setup SDK. `bash cloud/setup.sh --wasi-sdk` succeeded, and the failing WASI leaf, including its binding mutant, passed (7.739 s). setup-wasi.log records all timing lines (SDK ready 4.192 s, complete 46.060 s); native clang remains 20.1.8.

The next complete oracle run found only async_coverage_typeof.a, twice through its two catchers: native matched Node, but JavaScript exited70 with `ReferenceError: adamicViewAdapterUnderlying is not defined`. V4's shared equality emitter reaches main's separate async backend, which lacked that hook. Main's async grammar admits only primitive payloads (lower.asyncType/asyncValue) and no adapters. asyncJavaScript now supplies a documented primitive identity hook. The full JavaScript package, all async fixtures and the complete native-agreement shard 027 pass afterward; the latter combined run has 21 pass events in 7.413 s. A Go-overlay mutant deleting the hook reproduces the async fixture failure. The exact Node stdout was `true true true\nboolean boolean\ntrue function true\n`; the fixed backend agrees. The complete oracle pass before this fix is retained as a failure, not relabelled as a passing command.

Counts were regenerated again after the per-call dispatch correction: PASS, 52.101 s, house Linux clang 20.1.8. The final table and row audit below are authoritative.

The three rows V4 removes from counts are .a callable escapes it now refuses. Their programs had no a-check header even though their bodies were unchanged by the original source commit. I compiled these admission changes too and added the measured refusal headers: callable-union-nested (`view.holder.value`), callable-union-wrong (`view.value`), and v2/callable-producer-wrong (`view.value`). Bodies are byte-identical after the first line is removed. No test pins their numeric diagnostic line (the existing helper checks only path prefix, exact reason and fix). The existing producer mutant/refusal test is rerun. Final a-check therefore covers 55 files, including all renamed destinations and these three newly annotated admission changes.

Timing target: the 30-minute first-green push target was missed while resolving main integration failures, disk exhaustion, and toolchain provisioning and running the requested complete packages. No failed package attempt or cold leaf overrun is hidden; the source's pending skips are listed separately in changed-test-timings.json.

Final collected validation before the full-CPU comparison:

| Command scope | Result | Package seconds |
|---|---|---:|
| complete internal/flow, post integration | PASS | 70.809 |
| complete internal/fresh, post integration | PASS | 11.487 |
| complete internal/ir | PASS after ENOSPC cleanup | 58.866 |
| reader guard, -timeout=90s, final source | PASS | 12.919 |
| complete internal/javascript after async fix | PASS | see internal-javascript-final.jsonl |
| complete internal/load | PASS | 1.281 |
| complete internal/lower, final source | PASS | 36.663 |
| complete internal/native | failed attempts retained; every failed leaf rerun and passed after the listed fixes | 433.033 full attempt |
| complete internal/native/runtimeproducts | PASS; existing cold census leaves exceed 60s as noted | 97.710 |
| complete internal/oracle | all tests run; async failure subsequently fixed, both catchers/all async pass; all V4-focused tests pass | 415.619 full attempt; 40.176 V4 focused |
| complete stage1/cohere/json, four Go CPUs/four-way tests | PASS, all 2,048 shards | 379.879 |
| final counts family with -update-counts | PASS, only original own row keys changed | 52.101 |
| final compiler census | PASS | 0.314 |
| final descriptive-name/parallel guards | PASS | see name-parallel-guards-final.jsonl |
| vet, all ten touched packages | PASS, empty log | see vet-final.log |
| gofmt, changed Go files | PASS, empty log | see gofmt-final.log |
| a-check, 55 modified/renamed/admission-change .a files | PASS, both header mutants caught | see a-check.log |

Exact command arrays, package timings and individual pass/fail events are authoritative in the JSONL logs. All changed-test top-level timings are in changed-test-timings.json; no changed executed leaf exceeded 60s. Seven source-pending leaves remain skipped: four runtime pool/TSan/region cases and three callable capability cases. No claim of coverage for those pending features, no ahra/Loom submission, and no PR.

Measured internal/flow: pass 70.809 s.

Measured internal/fresh: pass 11.487 s.

Measured internal/oracle: fail 415.619 s.

Measured stage1/cohere/json: pass 379.879 s.

Full CPU comparison results (both complete packages, all 2,048 requested shards included):

`GOMAXPROCS=5 go test ./stage1/cohere/json -count=1 -parallel=5 -timeout=30m`

| Tree | go test seconds | End-to-end seconds | Peak cgroup anonymous GiB | Peak summed descendant RSS GiB | Peak cgroup total GiB | OOM / OOM kills |
|---|---:|---:|---:|---:|---:|---|
| base 561aeb4b | 318.133 | 320.607 | 4.084 | 6.595 | 15.942 | 0 / 0 |
| re-cut source a93a7bda | 336.990 | 339.471 | 4.180 | 6.513 | 16.000 | 0 / 0 |

Failing shards: **none on either tree**. Both logs contain only a successful package result, so there is no failure text to report for these measured runs. No assertion mismatch, timeout, killed process, or budget failure was observed. V4's Go-package wall time was 18.857 s (5.93%) higher in this one sequential comparison. Anonymous-memory peak was about 98.5 MiB higher, while summed-RSS peak was about 84.0 MiB lower. This is one observation each, not a population estimate.

Memory pressure: V4 reached the 16 GiB cgroup cap and memory.events.max increased by 1,596; base peaked at about 15.942 GiB with no new max events. There were zero OOM events/kills on both. V4 began with 13,839,249,408 bytes of file cache versus base's 11,611,127,808, so total cgroup usage/reclaim pressure is not an isolated code-memory comparison. Anonymous and RSS peaks are reported separately; component maxima need not coincide. The sampler observed a maximum every 50 ms plus its scan time (4,708 V4 samples, 4,341 base samples). RSS sums count shared pages repeatedly; they are not PSS. No competing test/compile suite ran during either measured command. Each tree's declared prerequisites were settled first, matching the check's served-product setup. Read fullcpu-{v4,base}.json.txt for exact byte counts, CPU affinity, quota, baseline/cache memory and counter deltas; the plain .log files preserve the exact requested command outputs.

Scope of inference: all 2,048 shards pass at this box's full affinity/nproc count of five, with a four-core CPU quota and a 16 GiB memory limit. This does not prove the removed Lunatic/Golem failures were noise, does not recover their missing failure text, and does not cover larger whole-box parallelism or other simultaneous workloads. Their exact cause remains unobserved here. No performance optimization from the separate V4-3 branch is implied by this re-cut.

The measured re-cut source commit is a93a7bdac8efa6d0079f27a715e368425cf94da8. The final submission amends only evidence (not Go/C/.a/counts or declared JSON corpus inputs) into that one commit. The final full SHA is reported in the submission response; source-input identity is checked before push.
