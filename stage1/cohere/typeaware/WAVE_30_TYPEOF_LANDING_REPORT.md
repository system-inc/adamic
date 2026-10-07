Built: clean landing-first rebase of 36 own commits onto current main's native typeof fixes through the lint area; no new rules or claims.
Commits: validated source 43a44373d5a1ae32eb61847a917802ffc77c09cc, on lint area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 containing main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06; evidence follows on codex/typeaware-wave-30.
Commands and outputs: twelve own gates PASS 471.300s; required shared compiler parity and comparator mutant PASS 103.152s; inherited typeof mutants PASS 3.796s and seven uncached fixtures PASS 1.143s; vet/diff check PASS; setup 110s, nproc 5.
Mutants: 25 own output mutants, three descriptor-name and six JSX-output checks caught; shared emitted-JavaScript comparator and inherited null, slot-presence, constructor and string-literal mutants caught by Node.
Not covered: full JSX source-rule parity/timings, checker-context integration, parked HIR/SSA/capture analyses, other packages' required external checks and the full repository gate.

The latest area imports main's unified native typeof classification, preserving
null meaning and slot lookup presence. All own commits rebased cleanly. No
protected compiler/runtime or shared harness file was edited by this unit;
integrated changes are accepted unchanged. The exact 36-commit map and verified
gzip logs with uncompressed SHA-256 hashes accompany this report.

All twelve TestWave30 gates pass again, holding the eight standalone behavior
ports to production Go on findings, fixes and suggestions, sanitizer variants,
released bridge handles, positive controls and frozen 77 compiler/287 repository
roots. The partial JSX helpers and parked React helpers also agree with Go,
sanitized native and emitted JavaScript. These helpers do not establish complete
source-rule parity. Registered RuleContext still lacks a checker program handle
and parser/checker-node mapping; parser.path does supply filenames and JSX
parsing works. The three JSX claims remain partial, and HIR/SSA/capture React
claims remain parked. No new claim was made.

The required ADAMIC_TYPESCRIPT_SOURCE points to pinned TypeScript 6.0.3 at
/workspace/wave-30-typescript. Shared compiler/stage1 parity executes without a
skip over all 487 files, matching 21022372 bytes across production Go, source
Node, emitted JavaScript and sanitized native in 83.25s. Its emitted-JavaScript
extra-output mutant is caught by the passing outer test in 19.88s. The child
failure printed by that proof is intentional. The 40 registered mutants and
upstream witnesses were green in the immediately preceding registry landing;
this runtime-only rebase repeats the compiler comparison and comparator mutant.

Inherited typeof mutant tests pass in 3.796s: five null fixtures, slot-presence,
constructor and string-literal output mutations finish cleanly and only Node's
stdout comparison catches them. They produce 10 native and 8 Node uncached
misses. The first combined -run filter selected those mutants but did not select
ordinary fixtures. Its log was inspected, and a separate hierarchical filter
then ran all seven ordinary typeof fixtures uncached in 1.143s, with 21 native
and 14 Node misses. No fixture coverage is inferred from the first filter.

Whole-process timings, including checker load, not benchmark medians:

| Group | Repository native / Go | Compiler native / Go |
| --- | --- | --- |
| Collection and discarded results | 374.063ms / 171.924ms | 2.714s / 902.811ms |
| ISO and callback | 285.177ms / 142.453ms | 1.900s / 301.605ms |
| Timer | 269.734ms / 134.093ms | 1.720s / 324.082ms |

Commands all write to logs, after source /workspace/adamic-tools/env.sh:

- git rebase origin/area/stage1-lint; bash cloud/setup.sh. Setup succeeds in
  110s, with Go/clang/Node/submodules 0s, cache warm 110s, 5 processors and
  cgroup quota 4 cores.
- go test ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v,
  with required TypeScript source and frozen repository/compiler manifests for
  ADAMIC_WAVE_30, NEXT, THIRD and PROCESS.
- go test ./stage1/cohere/lint -run
  '^(TestCompilerAndStage1Agree|TestEmittedJavaScriptMismatch)$'
  -count=1 -timeout 30m -v, with required TypeScript source.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^(TestNativeAgreesWithNode$/internal/oracle/testdata/typeof_.*[.]a$|TestTypeOf.*Mutant)$'
  -count=1 -timeout 10m -v, then separately the ordinary fixtures:
  -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/typeof_.*[.]a$'.
- go vet ./stage1/cohere/typeaware ./bridge/tsgo/...; git diff --check.

Setup cache warming filled the filesystem. Fourteen verified ELF/archive
artifacts from completed own scratch runs (405890975 bytes) and 39 old Go cache
archives (1530525786 bytes) were removed with sources and logs retained, but
usable space remained zero. Removing 38 verified cache-warm archives
(1803092872 bytes) recovered 1.6 GB; scoped validation rebuilt needed packages
and passed. All cleanup paths are recorded. No cache executable was removed
(the executable scan selected zero). The full repository gate and unrelated
required postcss/GraphQL checks were not run. No test was skipped, relaxed or
deleted. Push only the own branch, using an exact lease against old 7cbb62c3c.
