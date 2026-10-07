Built: rebased wave-02 onto current main and reran the five wave suites; no new rules claimed.
Commits: main e8ba3d5d81de4d3773c723914fccd4c76248b965; rebased implementation ff2c58aec19191fb71798da04a1038935a506efb; evidence commit follows.
Checks: five wave suites, bridge foundation, filtered Node oracle, sanitizers, vet and source gates pass; inherited inventory corpus oracle fails.
Mutants: all wave rule/question mutants, released-registry mutants, seven foundation mutants and ten inventory rule mutants are caught.
Limits: shared binding path equality blocks landing readiness; 14 and 99 JSX controls remain blocked; full repository gate and emitted-JavaScript comparison were not run.

The branch was rebased from 0580de22b27d50d089b918837d3eed98ddb2f906 without conflicts. Only codex/typeaware-wave-02 is pushed, using an exact old-tip force-with-lease because the requested rebase rewrites history. No main or area branch is pushed. This is a blocked landing attempt, not a green landing certification. No new claims or shared source edits were made.

Observed blocker

TestCoverageAgreementAndMutants fails at repository byte 67: Go reports callerDataMutation at bench/spectral_norm.ts bytes 521..528; native omits it. Across the frozen 287-root relative manifest Go reports 60 callerDataMutation findings and native reports zero. Converting the same manifest to absolute paths makes native report 60. Shared bindings.ts lines 39..44 accepts declarations only when checker path equals the manifest path; checker facts.go emits SourceFile.FileName(). The supported wave suites canonicalize their identities independently. Fixing the inherited shared declaration lookup is outside the owned rule directories; Ahra explicitly instructed stopping at such blockers rather than editing shared files. The inventory suite aborts before compiler-corpus and repository sanitizer checks, so those inventory checks are not claimed.

Passing checks

Fresh stage-0 and normal/sanitized C archives were built from the rebased tree. The five wave suites pass with complete finding/fix/suggestion bytes, normal and ASAN/UBSAN/LSAN runs and released-handle checks. Original controls: 28 findings/10576 bytes, compiler 363/142372, repository 3/14259. Continuation 1: 61/25869; continuation 2: 95/58914 plus DOM controls 3/1715; continuation 3: 481 supported controls, 433/221982; continuation 4: 272 supported controls, 2747/655330 and positive relative-root controls 1/375. Other wave corpus results have zero findings, retaining complete file headers: compiler 5318 bytes and repository 13319 bytes.

Go wave tests pass in 417.965s. Bridge foundation passes in 99.716s with 100 ABI queries and 3261 identical Go/native bytes across 162 positions. Filtered uncached Node oracle passes in 41.150s, including its one-byte mutant. Coverage inventory controls pass 53 findings/24202 bytes with sanitizers, followed by the corpus failure. go vet ./... passes; gofmt and git diff --check produce empty logs. Both owned source gates pass formatting and lint through temporary .ts snapshots; actual .a sources compile natively.

Setup used ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh. Timing lines: Go ready 1s; clang ready 1s; Node ready 1s; submodules ready 1s; build/cache warm 133s; done 133s. nproc 5; CPU quota 4. Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0.

Mutant results

All following wave rule mutants compile and exit zero; the independent Go bytes catch them at the indicated byte. Original: logical 316, unsafe-return 4907, nullish 6544. Continuation 1: collection 2949, outcome 20788, pure 12619; ancestry question 20788. Continuation 2: timeout 51957, exit 85, blocking 2073, namespace-kind 61, signature-scope 58902; graph 11647, context 61, signature 15524, modules 39263. Continuation 3: rest 62, regex 15082, supplemental React 155714. Continuation 4: button 818, checked 4739, display 7657. Both owned JSX refusal mutants compile and lose an independently observed Go finding. Original released probes panic 70; omission of released-registry deletion removes that panic and is caught.

Inherited inventory mutants: before 61, cast 1576, methods 2238, coercion 1177, caller 10318, parameter 373, invariant 12854, optional 15637, alias 16123, unused 16955. These all compile and exit zero and are killed by finding comparison. Foundation mutants: input-length and output-length caught by ASAN; retained registry by assertion; wrong source position by bytes; missing linkage by guard; missing C free and region heap allocation by LSAN. Node one-byte mutant is caught by its oracle.

Commands and evidence

Exact command output and streams are preserved in validation/evidence.tar.gz; streams.json lists decompressed hashes, sizes and original paths. The archive contains individually gzipped streams. Reproduction uses source /workspace/adamic-tools/env.sh, the frozen repository manifest, TypeScript v6.0.3 compiler manifest, and ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-02/typescript. The three Go wave tests are TestWave02AgreementAndMutants, TestWave02ContinuationAgreementAndMutants and TestWave02Continuation2AgreementAndMutants in ./stage1/cohere/typeaware. Each uses its ADAMIC_WAVE02[_CONTINUATION[_2]]_ARTIFACTS, _REPOSITORY_MANIFEST and _COMPILER_MANIFEST variables. Continuations 3 and 4 use their owned verify.py with --scratch, --stage0, --archive, --sanitized-archive and --typescript-source. The latter also runs guard_verify.py. The inherited inventory command is go test -v -count=1 -timeout30m ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$' with ADAMIC_COVERAGE_ARTIFACTS, ADAMIC_COVERAGE_REPOSITORY_MANIFEST and ADAMIC_COVERAGE_COMPILER_MANIFEST. TestBridge runs separately from the bridge/question/filtered Node command. Setup, source, vet, format, rebase, fetch and test logs are included.

Native versus Go

Three alternating count-checked rounds include process startup, program creation and traversal. Native is slower in every measured suite. Medians in seconds:

| Suite | Compiler native / Go | Repository native / Go |
| --- | --- | --- |
| Original | 5.324441 / 1.241677 | 0.503506 / 0.200200 |
| Continuation 1 | 3.160669 / 0.874795 | 0.410581 / 0.164975 |
| Continuation 2 | 2.606312 / 0.323902 | 0.355025 / 0.137985 |
| Continuation 3 | 2.435906 / 0.423381 | 0.299610 / 0.141731 |
| Continuation 4 | 1.985432 / 0.313330 | 0.281853 / 0.132076 |

bench.py reproduces timings from the freshly checked binaries; timings.json is archived. Timing rounds were separate from builds; a brief absolute-path diagnosis ran during the first suite. No inventory timing or green inventory corpus claim is made. All existing nondefault-option, JSX, emitted-JavaScript and full-gate exclusions remain. The landing cap prevents further claims until the shared blocker is resolved and the branch is green again.

Landing recheck on current main

Origin main advanced to f8013f0baac41ddc340d76f83bddde38536a8f07. All 21 worker commits rebased cleanly; pre-evidence tip 838b470e3dcf040e66bd9a6dae3688357318b44c. No shared files or claims changed. The inventory test was rebuilt and rerun with the corrected command `go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestCoverageAgreementAndMutants$'`, ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-02/landing-recheck/base-coverage, the frozen repository manifest, the existing compiler manifest and TypeScript source. The initial invocation mistakenly used -timeout30m and failed before running tests; it was corrected.

Controls and ASAN controls pass: 53 findings, 24490 identical bytes. All ten compiling mutants exit zero and are caught by Go bytes: before 69, cast 1608, methods 2294, coercion 1201, caller 10430, parameter 381, invariant 13006, optional 15821, alias 16323, unused 17179. The test fails after 155.382s on the same repository mismatch at byte 67, missing callerDataMutation at bench/spectral_norm.ts bytes 521..528. This fresh failure is in validation/recheck-oracle.log.gz. Later inventory compiler/repository sanitizer/release checks are not reached. The five wave suites and benchmarks were not repeated on this new main because the prerequisite inventory corpus still fails. Their earlier results remain observations on the earlier main only. No green certification, new performance measurement or listener implementation is claimed. The shared-file restriction remains the blocker.

Landing recheck on c01907a7

All 23 worker commits rebased cleanly onto origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; pre-evidence tip 4f5adfc4dadfb774026223042a7ca6a55033efcc. The rebuilt TestCoverageAgreementAndMutants ran with the same corrected command and manifests as above, using ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-02/landing-c019/base-coverage. Controls and sanitizer controls match 53 findings / 24382 bytes. All ten compiling, exit-zero mutants are caught by Go comparison: before 66, cast 1596, methods 2273, coercion 1192, caller 10388, parameter 378, invariant 12949, optional 15752, alias 16248, unused 17095. The test fails in 141.916s at repository byte 67 on the same missing callerDataMutation, bench/spectral_norm.ts 521..528. See validation/c019-oracle.log.gz. Later inventory checks and the five wave suites were not rerun after this prerequisite failed; no new timings or green certification are claimed. React remains parked; no new claims, shared edits or leak-helper reversions were made.

Landing recheck on b8fb957a

All 24 worker commits rebased cleanly onto origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30; pre-evidence tip 1cd18b3a2148a77fc8ed904ccec574399acd4579. TestCoverageAgreementAndMutants was rebuilt and rerun with the previously documented command and manifests, using ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-02/landing-b8fb/base-coverage. It fails after 149.324s at repository byte 67 on the same missing callerDataMutation at bench/spectral_norm.ts bytes 521..528. All ten inventory mutants are caught and controls pass under sanitizers. Complete fresh output is validation/b8fb-oracle.log.gz. Later inventory checks, five wave suites and timing benchmarks were not repeated after this prerequisite failed. React remains parked. No new claims or shared edits were made; the branch remains blocked rather than landing-ready.

Requested shared-harness area rebase

Per the explicit area rebase instruction, all 25 worker commits rebased cleanly onto origin/area/stage1-lint 7481e0324e34a2537aafa9db7eeacda50405611b. This area tip contains current origin/main 39638d9e278d38bb5aeae887f46d55a70e47aaad; pre-evidence worker tip 26d80a2b84d92c3025eed0a56c188656cebb4ddb. Only the worker branch is pushed. Existing shared changes are accepted without reversions.

The rebuilt TestCoverageAgreementAndMutants ran with the previously documented corrected command and manifests, using ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-02/landing-area/base-coverage. Controls and sanitizer controls match 53 findings / 24382 bytes. All ten inventory mutants compile, exit zero and are caught by Go bytes. The test fails in 157.198s at repository byte 67 on the same missing callerDataMutation in bench/spectral_norm.ts 521..528. Complete output is validation/area-oracle.log.gz. The shared harness does not resolve this separate inherited binding-path lookup. Later inventory checks, five wave suites and timing benchmarks are not rerun after this prerequisite failure. React remains parked; no new claims, shared edits or green certification are made.

Area runtime rebase on d65a8f93

Rebased cleanly onto origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e, retaining the incoming runtime profiling and optimization changes. The area contains current main 39638d9e; pre-evidence worker tip e69a5b8d794076e1c4dc53420bee12ecc188c639. The rebuilt inventory oracle uses the same command and manifests, with ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-02/landing-d65/base-coverage. Controls and sanitizer controls match 53 findings / 24346 bytes; all ten inventory mutants compile, exit zero and are caught. Repository comparison still fails at byte 67 in 157.501s on the missing callerDataMutation at bench/spectral_norm.ts 521..528. Complete fresh evidence is validation/d65-oracle.log.gz. Later inventory checks, five wave suites and benchmarks were not repeated after the prerequisite failure. Shared code was not changed or reverted, React remains parked, and no new rules were claimed. This remains a blocked landing attempt.

Area rebase on b84a9d93

All 27 worker commits rebased cleanly onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da, which contains current origin/main c7991b900362796aefd111474e65eb5398e91953. Pre-evidence worker tip 7d8f64fb2db140c1d3d3a21eb2d2de50c417d665. The inventory test rebuilt from this tree uses the documented command and manifests with ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-02/landing-b84/base-coverage. Controls and sanitizer controls match 53 findings / 24346 bytes, and all ten inventory mutants are caught. Repository comparison fails at byte 67 after 158.601s on the same missing callerDataMutation at bench/spectral_norm.ts bytes 521..528. Complete fresh output is validation/b84-oracle.log.gz. Later inventory checks, five wave suites and timing benchmarks were not repeated after this prerequisite failure. No tests were weakened or bypassed; no shared code was edited, no new rules were claimed, and React remains parked. This is still a blocked landing attempt.
