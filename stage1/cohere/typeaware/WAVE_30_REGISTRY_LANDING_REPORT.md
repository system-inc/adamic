Built: clean landing-first rebase of 35 own commits onto the registry-only lint driver; no new rules or claims.
Commits: validated source 4025a8ce91c61ef646d0f55f4946fd5fa406a992, based on lint area b46914832d70e00847d82d5d221ab7bb24040c53 and containing main c7991b900362796aefd111474e65eb5398e91953; evidence follows on codex/typeaware-wave-30.
Commands and outputs: twelve own gates PASS 451.466s; shared agreement and mutation suite PASS 896.481s; vet and diff check PASS; setup 37s, nproc 5.
Mutants: all 25 own successful-exit output mutants, three descriptor-name checks, six JSX-output checks, 40 shared registered mutants, overlap mutant and emitted-JavaScript comparator mutant caught.
Not covered: complete JSX source-rule parity and timings, checker-context integration, parked HIR/SSA/capture analyses, other packages' required external checks and full repository gate.

The lint area migrated its remaining legacy rules into owned registry directories.
All 35 own commits rebased cleanly onto that integration. No shared harness,
generator, registration list or protected compiler/runtime file was edited by
this unit. The exact old/new map and gzip-preserved logs are beside this report.
Their decompressed lengths and SHA-256 hashes are recorded in streams.json.

The required compiler input was supplied as
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript, pinned TypeScript 6.0.3.
TestCompilerAndStage1Agree now covers 487 files and matches 21022372 bytes across
Go, source Node, emitted JavaScript and sanitized native in 86.68s. Shared
upstream witnesses match 13053452 bytes in 49.01s. TestMutants passes all 40
registered rule mutants in 718.15s; the overlap mutant passes in 19.74s.
TestEmittedJavaScriptMismatch passes in 22.88s by catching its intentional
extra-output child failure. That child failure is expected, not an unresolved
failure. No test skipped or was weakened.

All twelve TestWave30 gates pass again, including the eight complete standalone
behavior ports, exact findings/fixes/suggestions, sanitizer variants, released
bridge handles, partial JSX and parked React components. The frozen own compiler
and repository manifests remain 77 and 287 roots. Zero corpus findings for those
eight ports are accompanied by positive controls. These standalone ports are
not claimed as integrated registry rules. The three partial JSX rules still
lack a registered RuleContext checker program handle and checker-node mapping;
parser.path supplies the filename, and JSX parsing itself works. No full
source-rule parity is claimed from the tested component helpers. Parked React
HIR/SSA/capture claims remain parked. No further claim was made.

Whole-process native versus Go times, including checker load, not benchmark medians:

| Group | Repository native / Go | Compiler native / Go |
| --- | --- | --- |
| Collection and discarded results | 384.788ms / 190.479ms | 2.692s / 784.376ms |
| ISO and callback | 288.216ms / 139.973ms | 1.939s / 309.411ms |
| Timer | 303.555ms / 151.513ms | 1.952s / 302.768ms |

Commands, each redirected to a log after sourcing /workspace/adamic-tools/env.sh:

- git rebase origin/area/stage1-lint; bash cloud/setup.sh. Setup reports Go,
  clang, Node and submodules 0s, cache warm 37s, total 37s, 5 processors with
  cgroup quota 4 cores.
- go test ./stage1/cohere/typeaware -run '^TestWave30' -count=1 -timeout 30m -v,
  with ADAMIC_TYPESCRIPT_SOURCE and both frozen manifest variables for
  ADAMIC_WAVE_30, NEXT, THIRD and PROCESS.
- go test ./stage1/cohere/lint -run
  '^(TestRulesAgree|TestCompilerAndStage1Agree|TestEmittedJavaScriptMismatch|TestLegacyMutants|TestMutants)$'
  -count=1 -timeout 30m -v, with the required TypeScript input.
- go vet ./stage1/cohere/typeaware ./bridge/tsgo/...; git diff --check.

Malformed upstream recovery cases remain explicit parser refusals, with Go
recovery output recorded for comparison; they are not successful lint parity.
Other packages' required postcss/GraphQL inputs were not tested, and no full
repository gate is claimed. No native source-rule timing is certified for the
partial JSX rules. The push target is only codex/typeaware-wave-30, replacing
its old ffdac802d tip with an exact lease under the requested rebase workflow.
