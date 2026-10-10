Built eep-presence containment on main; namespace-value and optional-presence are dropped for real symbol dependencies.
Member commit: ef1d8f4b, extracted from 21ceb23c with its 4478575d repair.
Proof: lower 282 pass / two existing skips; fixtures, reader guard, counts and vet pass.
Mutant: removing the guard is caught by all seven refusal tests; supported control passes.
Not covered: dropped members, runtime support for refused cases, full gate, WASI or other platforms.

## Admission delta

No newly admitted programs and no newly admitted divergence. This is a structural proof, not a finite-corpus claim: the only changed existing production function is property at internal/lower/object.go:341. It first calls eepPropertyPresence, which queries checker types/symbols and syntax without changing lowerer state or IR, returning nil or a located NotYet. If nil, all main code and guards execute unchanged. Admission can only decrease. Other frontend, IR, backend, runtime and dependency inputs are identical to main. See ../chain-slice-7/admission-delta.json.

The seven type-correct witnesses are compile-time refusals, not negative witnesses licensed to diverge. None contains a cast, non-null assertion or declared-type lie. They do not run in the delivered backends. No runtime stop is credited as agreement. The supported neighbor executes and agrees with raw Node on stdout and exit in JavaScript, native and ASan/UBSan/LSan native.

## Extraction and dependency audit

Base is current origin/main 72ad75effe2dbcc0deedc3e7e34a8d5e29670eac. No chain or slice branch was merged. The source plan, repair ledger, merge history, commit history and refined ranges are preserved under ../chain-slice-7/. Slice 6's report was inspected as the model.

| Member | Source SHA | Refined fork point | Decision | Dependency evidence | Tasks |
| --- | --- | --- | --- | --- | --- |
| namespace-value | 26ccff9ca899d97dd1344ff961f1a400fceadf19 | 6f933fa0 | dropped | needs exceptions-21-main (slice 6) for ir.MakeError.Constructor at internal/lower/namespaces.go:444 in repair 71972e18. Main's MakeError at internal/ir/ir.go:573 has only Message and Name. exceptions-21 supplies Constructor at its internal/ir/ir.go:576. Preserving the catchable TypeError constructor requires its builtin error representation and propagation; substituting a named plain Error is not the same meaning. | #67jk7wq remains open; step 15 value escape not delivered |
| optional-presence-next | 897d0e7989ecfc2175f25966b8def3efe51b1ec8 | 2391c655 | dropped | needs namespace-value for namespaceType at internal/lower/library_object.go:95 in its preserved hasOwn guard. namespaceType is defined at internal/lower/namespace_value.go on namespace-value and absent on main. The required member is dropped above. | #r3chqza remains open; #p6a00tt and #k881crd still wait |
| eep-presence | 21ceb23cb860505d94b862bfeeaa79d166e2c22e | 897d0e79 | kept | No optional-presence symbol, fixture or IR dependency. The helper uses main's concrete, notYet, instance.static and accessorNames; ast/checker APIs are present in the pinned checker. staticInitializerCallsThis is defined in this member. Test helper lowerSource exists on main. Ancestry alone is not an edge. | #eep1m5z containment locally complete, awaiting integration acceptance |

The original namespace readiness fields belong to namespace-value, but the exact typed catchable error constructor required by its repair belongs to exceptions-21. This report does not claim that main lacks all readiness. Slice 6 has not landed on main. No outside-member code was copied to remove the edge.

EEP's source range 897d0e79..21ceb23c already includes implementation repair 4478575d; it was applied once, not twice. Its property hunk was placed at main's function entry while preserving the existing staticProperty guard and all following lowering. The generatorResultRead context from the stacked source was not copied. The own-net files are byte-identical to the member except that three-line insertion location in object.go. No other member's production or test changes were applied.

Namespace repairs 71972e18, a233eec3 and 2b2b1095 and optional repairs 163e39f8, d26baa81, 84d03d57, c014d545 and 167aaf0c are excluded with their dropped owners. The merged reader/counts envelope from 167aaf0c is not needed by the kept member: it adds no protected call-target reads. Main's approved reader table remains intact.

## Node and backend evidence

Commands source /workspace/adamic-tools/env.sh, unset GOCACHEPROG, use GOMAXPROCS=4, GOFLAGS='-buildvcs=false -trimpath -p=2' and TMPDIR=/workspace/chain-slice-7-tmp. Output goes to logs under ../chain-slice-7/. Every child has a hard limit. run.py uses raw Node --input-type=module-typescript on source bytes; generated JavaScript uses the repository runtime resolver. Every fixture is compiled separately in release native, --sanitize native and JavaScript modes, with a 60-second child bound.

| Source fixture | Raw Node stdout / exit | Native / sanitized / JS compilation |
| --- | --- | --- |
| eep1m5z.a | `1 2\n` / 0 | located NotYet, all exit 1 |
| eep1m5z-93122fa_t_t5.a | `1 2\n` / 0 | located NotYet, all exit 1 |
| eep1m5z-93122fa_t_v1.a | `2 -1\n` / 0 | located NotYet, all exit 1 |
| eep1m5z-93122fa_t_v2.a | `2 3\n` / 0 | located NotYet, all exit 1 |
| eep1m5z-93122fa_t_v5.a | `2\n2\n` / 0 | located NotYet, all exit 1 |
| eep1m5z-classfeat_method_view.a | `worker one1\nlazy\n` / 0 | located NotYet, all exit 1 |
| eep1m5z-classfeat_static_virtual.a | empty / 1 | located NotYet, all exit 1 |
| supported neighbor | `2\n2\nok\n` / 0 | all compile and execute; identical stdout, exit 0, empty stderr |

results.json records complete stdout, stderr, exit, commands and seconds. Sanitized execution enables ASAN_OPTIONS=detect_leaks=1 and UBSAN_OPTIONS=halt_on_error=1. The static witness's raw Node exception is retained, including its engine-specific stderr. An initial source observation through oracle/node.mjs remapped that throw to exit 70 and was replaced by raw Node. A direct generated-JS attempt could not resolve the adamic runtime package; final execution uses the resolver and passes. Neither attempt is credited as a semantic result.

## Tests and mutant

| Proof | Exact command / result | Evidence |
| --- | --- | --- |
| Lower shards | timeout 600 python3 review/compiler/chain-slice-7/run-lower-shards.py; 284 top-level leaves in 18 disjoint selectors, each go test ./internal/lower -run exact-selector -count=1 -json -timeout 80s with outer 85-second limit. 282 pass, two existing skips, zero missing/duplicate verdicts. Maximum shard 9.828 s. | lower-results.json, lower-union.json, lower-00..17.jsonl |
| Own fixtures | timeout 180 python3 review/compiler/chain-slice-7/run.py; seven REFUSED, supported FIXED. Source Node plus both backends, native sanitizers/leaks. | results.json, fixtures-final.log and per-mode logs |
| Revert mutant | timeout 90 go test -overlay=review/compiler/chain-slice-7/revert-overlay.json ./internal/lower -run '^TestEEPPresence' -v -count=1 -timeout 80s; exits 1. TestEEPPresence0..6 each report got <nil>; supported control passes. One shared guard-removal mutant, seven catchers; no build/warning kill. | object-mutant.go.txt, revert.patch, revert-overlay.json, revert-mutant.log |
| Reader guard | timeout 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 80s -json; pass 22.720 s. | readers-retry.jsonl |
| Counts | timeout 180 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 150s -json -args -update-counts; pass 99.402 s. Table regenerated once successfully; all 1,011 rows byte/value unchanged, zero new/removed rows. Every row attributed to unchanged admitted IR. Refusal fixtures are outside the executable oracle census. | counts-retry.jsonl, counts-attribution.json |
| Vet | timeout 90 go vet ./internal/lower; exit 0, no output. | vet.log |
| Lane | Required fetch and origin/cloud/merge-tree lane script after member commit: exit 0, gofmt/tools on three Go files, t.Parallel on one test package. Initial vet exceeded its short budget; separate vet passes above. Final delivery lane: exit 0, lane checks 0.8 s, gofmt/tools on three Go files, t.Parallel on one test package, vet one package. | lane-fetch.log, lane-checks.log, lane-final.log |

New test leaf seconds in the full shard proof: TestEEPPresence0 0.06, 1 0.08, 2 0.08, 3 0.03, 4 0.09, 5 0.06, 6 0.06, Supported 0.10. Existing skips are TestOriginalCycleLedger and TestOptionalWideningCensus. No added test skips; no whole-package tests or full gate ran.

## Setup and limits

GOPROXY=https://proxy.golang.org|direct was exported before setup. First timeout 180 bash cloud/setup.sh reached Go 0.062 s, Node 0.085 s, clang 0.504 s, Markdown 1.123 s, submodules 16.728 s and shared cache 23.111 s, then expired during warming. Subsequent build/test attempts hit cold-cache limits and then no space left on device in /tmp. Only two explicitly identified abandoned scratch directories from this unit were removed; subsequent scratch moved to /workspace.

Retry with ADAMIC_GOCACHE_OFF=1 completed: Go 0.021 s, Node 0.021 s, submodules 0.054 s, Markdown 0.061 s, clang 0.160 s, shared cache off 0.161 s, build 150.934 s, deferred test binaries 151.059 s, cache warm 151.060 s, done 151.086 s. nproc=5, cgroup quota four CPUs. Go 1.27.1, Node 24.19.0, clang 20.1.8. Pinned Node declarations were installed using npm ci --prefix stage3/api. Failed infrastructure attempts are retained and not counted as passes.

No runtime file changed against main. @system_adamic_runtime has no runtime delta to clear for this slice. No forbidden orchestration/emission file was edited. Static refusal conservatively overapproximates reachability; runtime tuple lengths and method/getter dispatch support remain outside scope. This advances sound compile-time containment for #eep1m5z. Step 15 namespace value escape remains blocked by the recorded dependency; no task waiting on optional presence is claimed closed. Only compiler/chain-slice-7 is pushed, with no PR or main/area push.
