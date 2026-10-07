Rebased all twelve owned checker-rule ports onto the integrated lint area without changing rule bodies or shared sources.
Current validated base is origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and harness 41eb6eab2; push only the worker branch.
All four owned byte-oracle suites, named listeners, sanitizer and released-handle checks PASS again; original suite 160.121s; affected Node typeof fixtures and one-byte mutant PASS 18.719s; vet clean.
Caught 12 rule and 11 raw-question byte mutants, released-registry mutation, native listener mutation and two JSON listener mutations.
Default options only, full repository gate and older node-only migration uncovered; React analysis claims remain parked.

## Landing and validation

Rebased cleanly onto the requested current area tip, which includes the integrated finding model, module/suggestion transport, JSX parser and developer-tools changes. Those shared changes are preserved; the worker diff against the area contains no shared lint, parser, native emitter or lowerer changes. The twelve public listener declarations already use validated AST names. No new rule or bridge question is added.

Read the updated CLAUDE.md and docs/lint-registration.md; ran go run ./cmd/lint-registry before building. It validated the 15 integrated registry entries. Generated registries remain ignored and uncommitted. Built a fresh compiler with go build -o /tmp/wave24-area-final-adamic ./cmd/adamic. Setup succeeds: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, cache warm 85s, done 85s, nproc=5; four-CPU quota, 17.6 GB. Removed only named, archived scratch binaries to free disk; source and prior evidence were retained.

Original Go oracle passes in 143.077s with 43 control findings; next has 65, third 295, newest has 125 over 217 valid control roots. Every suite compares complete canonical findings, fixes and ordered suggestions on the frozen 287-root repository and 77-root TypeScript compiler populations, in ordinary and ASan/UBSan modes. Both corpora have zero findings. All released-handle checks require panic 70 and pass. New JSX support causes no comparison drift in these tested TS populations; no new React certification is claimed.

The three parked React ports still require native high-level IR/capture/single-assignment analyses. JSX is now integrated on this base, so JSX availability is no longer listed as the current static-components blocker. Historical pre-JSX refusal evidence remains historical, not a current certification test.

Commands (all output directly to files):

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry > /tmp/wave24-area-registry.log 2>&1
go build -o /tmp/wave24-area-final-adamic ./cmd/adamic > /tmp/wave24-area-build.log 2>&1
ADAMIC_WAVE24_ARTIFACTS=/workspace/wave24-area-original ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-24-corpus ADAMIC_WAVE24_REPOSITORY_MANIFEST=/workspace/wave-24-repository.manifest ADAMIC_WAVE24_COMPILER_MANIFEST=/workspace/wave-24-compiler.manifest go test ./stage1/cohere/typeaware -run TestWave24AgreementAndMutants -count=1 -v > /tmp/wave24-area-original.log 2>&1
python3 stage1/cohere/typeaware/wave-24-next/validate.py /workspace/wave24-area-next --stage0 /tmp/wave24-area-final-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-area-next.log 2>&1
python3 stage1/cohere/typeaware/wave-24-third/validate.py /workspace/wave24-area-third --stage0 /tmp/wave24-area-final-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-area-third.log 2>&1
python3 stage1/cohere/typeaware/wave-24-fifth/validate.py /workspace/wave24-area-fifth --stage0 /tmp/wave24-area-final-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-area-fifth.log 2>&1
python3 stage1/cohere/typeaware/wave-24-fourth/validate-listeners.py /workspace/wave24-area-listeners --stage0 /tmp/wave24-area-final-adamic --checker /workspace/wave24-area-fifth/checker.a > /tmp/wave24-area-listeners.log 2>&1
go test ./bridge/tsgo/... -count=1 > /tmp/wave24-area-bridge.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave24-area-vet.log 2>&1
python3 stage1/cohere/typeaware/wave-24-fifth/timing.py /workspace/wave24-area-fifth --repository /workspace/adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-area-timing.log 2>&1
```

Mutants remain those documented in KIND_NAMES_REPORT.md: original receiver/await/unknown and class/then/handler facts; next exit/timeout/blocking and provenance/callee/module facts; third reject/rest/regex and direct-symbol/count facts; newest await wording/Symbol provenance/typeof wording and generic/index/heritage facts. Each of the 23 compiled byte mutants exits 0 with empty stderr; only Go byte comparison catches it. The released-registry mutant fails the required panic; actual named native and JSON listener mutants fail name comparison.

Three alternating whole-process counts after all builds/tests stop: repository native median 287.329 ms versus Go 100.248 ms (2.866x slower); compiler native 1479.105 ms versus Go 281.614 ms (5.252x slower). These measure the newest three-rule suite, with load/parse/walk overhead and no build time. No native speed advantage is claimed. Samples, normal/sanitized byte outputs, mutants, setup and registry logs are in evidence/area-landing/. The shared registry generation check is not additional shared-driver certification of these standalone legacy ports.

## Final area movement

After the first area-based push (7c4463dc3), area advanced to d65a8f93 with allocator-release and string-search optimization. Rebased again without editing or reverting those shared changes. Regenerated the registry, rebuilt /tmp/wave24-area-final-adamic and repeated all four complete owned suites, both named-listener checks, all 23 byte mutants, registry/native/JSON mutations, sanitizer runs and released handles. Every check passes. Original Go suite passes in 143.077s. Vet remains clean. The bridge implementation did not change, so its passing 64.323s/0.249s package tests from the first area pass were not repeated. The timing figures above are this final optimized baseline after all builds stopped. Final artifacts are in evidence/area-landing/final/.

Ran go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/runtime_last_index_of|TestTheOracleCatchesOneByte' -count=1 -v with direct output in runtime-oracle.log. The exact runtime search fixture and one-byte oracle mutant pass in 11.525s, with the fixture taking 11.49s; worker caching enabled (native hits=1, misses=3). No compiler source was edited by this unit.

Post-push audit of all origin branches inspected 610 refs and 33 distinct claim blobs. The conservative mention-based scan counts 172 claimed rules and leaves no unclaimed rule after the frozen main/bridge exclusions. No next claim is made. Main remains 39638d9e; the final area tip was independently verified as d65a8f93 before publication. The shared registry, parser, heap/string optimization and leak-helper changes are preserved unchanged.

## Latest lowering and record-runtime landing

Rebased cleanly onto b84a9d9314b65d3d0261ee017e233287b4f071da, containing current main c7991b900362796aefd111474e65eb5398e91953. Preserved all integrated proven-predicate, relation and record-runtime changes without editing shared sources. Ran cloud/setup.sh successfully: Go, clang, Node and submodules ready at 0s; cache warm 135s; total 135s; nproc=5. Regenerated the 15-entry shared registry and rebuilt /tmp/wave24-refresh-adamic.

Repeated the four complete owned suites using the commands above with /tmp/wave24-refresh-adamic and logs /tmp/wave24-refresh-{original,next,third,fifth}.log. All pass, including full findings/fixes/suggestions bytes on both frozen corpora, all 23 byte mutants, normal/sanitized controls and exact released-handle panics. Original suite passes in 226.443s. Repeated the nine native and JSON named listeners, their mutants, bridge packages and vet: all pass. Bridge times are 162.084s and 1.060s. The latest three declarations and metadata mutant pass in the fifth suite.

Ran the exact Node oracle filter TestNativeAgreesWithNode/internal/oracle/testdata/proven_(assertions|class_guards|guards|satisfies|upcasts)|TestTheOracleCatchesOneByte. The initial run failed before comparisons because the runtime cache could not allocate disk space. Freed only obsolete large binaries in explicitly named worker scratch directories, retaining source and committed evidence. Repeated the unchanged filter: all five fixtures and the one-byte mutant pass in 22.191s. Both logs are retained. No check was skipped or relaxed.

After all builds stopped, three alternating whole-process timing medians: repository native 283.297ms versus Go 99.580ms (2.845x slower); compiler native 1493.848ms versus Go 307.696ms (4.855x slower). These newest-three-suite measurements supersede earlier timing figures for this base. Full gate and the seventeen newly mandatory external-input correctness checks were not run; no coverage of them is claimed. Default-option and older node-only migration limitations still apply. React analysis claims remain parked.

Fresh all-origin audit: 636 refs, 33 distinct claim blobs, 172 claimed rule mentions, no unclaimed rule. Verified main, area and own remote worker tips before publication. Logs and timing samples are in evidence/area-landing/refresh/. No new claim is made.

## Registry-only area update

Rebased cleanly onto b46914832d70e00847d82d5d221ab7bb24040c53. Main remains c7991b90. The area delta migrates legacy lint rules to the shared registry; compiler, bridge and owned type-aware source files are unchanged. Preserved shared edits without changing them. Regenerated 40 descriptors and built /tmp/wave24-registry-adamic. Reused the already validated toolchain from the preceding 135s setup, nproc=5.

Repeated all four complete owned byte-oracle suites with the preceding commands and new compiler path, writing /tmp/wave24-registry-{original,next,third,fifth}.log. All normal and sanitized controls and both corpus byte streams agree with Go on findings, fixes and ordered suggestions. All 23 compiled byte mutants are caught only by comparison, released-handle checks require panic 70, and registry/native/JSON listener mutants are caught. Original Go suite passes in 151.436s. Named-listener validator and vet pass. The unchanged bridge packages and Node lowering checks were green on the prior base; they were not repeated for this shared-lint-only delta. No shared-driver certification of these standalone ports is claimed.

After builds stopped, three alternating whole-process medians for the newest-three suite: repository native 273.497ms versus Go 101.286ms (2.700x slower), compiler native 1417.679ms versus Go 279.134ms (5.079x slower). Samples and logs are in evidence/area-landing/registry/. Removed only obsolete large worker scratch binaries to maintain disk space; retained source and logs.

Fresh audit of 644 origin refs, 33 distinct claim blobs and 172 rule mentions leaves no unclaimed rule. Verified remote area, main and own branch before publication. React analysis claims remain parked. Default-option, older node-only migration, full-gate and seventeen mandatory-input check limitations remain unchanged; none of those unrun checks is represented as passing.

## Typeof null and dispatch integration

Rebased cleanly onto d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing current main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Preserved the inherited typeof lowering, slot-presence and union-runtime fixes without editing shared compiler sources. Registry generation validates 40 entries; fresh compiler is /tmp/wave24-typeof-adamic. The preceding successful setup remains the toolchain used (135s, nproc=5).

Repeated all four complete owned suites using the commands above with this compiler and direct /tmp/wave24-typeof-{original,next,third,fifth}.log output. Full canonical findings/fixes/ordered suggestions agree with Go in normal and sanitizer modes on controls and both frozen corpora. All 23 compiled byte mutants, released-registry mutation, native and JSON listener mutations are caught. Released handles require exact panic 70. Every suite passes; original suite takes 160.121s. Listener comparison and vet pass. Unchanged bridge packages were not repeated; their previous successful package checks remain separate observations.

Ran go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/typeof_(dispatch|null|string_literal)|TestTheOracleCatchesOneByte' -count=1 -v with direct logging. The affected typeof fixtures and one-byte mutant pass in 18.719s, including native sanitizer and emitted-JavaScript comparison with Node. Worker caching is enabled (native 7 hits/15 misses, Node 0 hits/15 misses). No skip or guard relaxation was introduced.

Disk space filled while mutant binaries accumulated. Removed only explicitly named completed worker mutants after their comparison logs had been recorded, plus obsolete worker scratch binaries. All source, byte observation files and logs were retained; remaining checks passed without a disk failure. Three alternating whole-process medians after builds finished: repository native 272.634ms versus Go 102.984ms (2.647x slower), compiler native 1502.864ms versus Go 298.747ms (5.031x slower). Samples and logs are in evidence/area-landing/typeof/.

Audited 659 origin refs, 33 distinct claim blobs and 172 claimed rule mentions: no unclaimed rule. Verified exact remote integration and own worker tips before publication. React analysis claims remain parked. Nondefault options, older node-only migration, full gate and seventeen mandatory external-input checks remain unverified.

Final evidence-directory creation initially failed with no space left on device after tests passed. Freed completed binaries from the three named older-suite scratch directories and retried the evidence write; no check result was altered.
