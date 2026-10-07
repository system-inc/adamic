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
