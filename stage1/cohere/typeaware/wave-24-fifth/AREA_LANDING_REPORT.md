Rebased all twelve owned checker-rule ports onto the integrated lint area without changing rule bodies or shared sources.
Final base origin/area/stage1-lint is d65a8f931c98655936ae04c6899f38f14862b73e, containing main 39638d9e and harness 41eb6eab2; push only the worker branch.
All four owned oracle suites and named-listener checks PASS; bridge tests PASS 64.323s/0.249s; vet and diff checks clean.
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
