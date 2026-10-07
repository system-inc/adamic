Built: ran the shared lint compiler/stage1 comparison with its required pinned input, replacing an earlier skipped observation with measured parity; no rule code changes.
Commits: tested unchanged wave-30 tip 310e8a616 on area d65a8f931, containing current main 39638d9e; evidence accompanies it on the own branch.
Commands and outputs: TestCompilerAndStage1Agree PASS 83.420s, 440 files and 21021585 identical bytes; comparator mutant PASS 18.927s; setup 75s, nproc 5.
Mutants: a clean-running emitted-JavaScript extra-output line is caught by the same Go byte comparator; earlier per-rule mutants remain recorded in the area landing report.
Not covered: the other packages' required compiler/CSS/GraphQL checks, full repository gate and full JSX source-rule ports; their checker-context integration gap remains.

Fetched origin before work. Main remains 39638d9e and area/stage1-lint remains
d65a8f931, both ancestors of the own pushed branch. No rebase is required.
No main or area branch is a push target. No new rules were claimed, and no
implementation, shared harness, skip condition or required-input enforcement
was edited. Only owned evidence and claim/report text are added.

The earlier full shared-package gate in WAVE_30_AREA_LANDING_REPORT.md was
reported accurately as passing, but its shell did not supply
ADAMIC_TYPESCRIPT_SOURCE, so TestCompilerAndStage1Agree skipped. Ahra now requires
previously optional correctness inputs. This follow-up explicitly supplies the
pinned TypeScript 6.0.3 checkout at commit
050880ce59e30b356b686bd3144efe24f875ebc8. The existing test validates that exact
pin and gathers every .ts/.a file under compiler src/compiler and repository
stage1. It executes, rather than skips, on 440 files. Unmodified Go cohere,
source Node, Adamic emitted JavaScript on Node and sanitized native produce
21021585 identical bytes. This certifies the shared syntax-lint comparison;
it does not integrate or certify the three partial type-aware JSX rules.

The same compareWithJavaScript function is proven able to reject a mutant by
TestEmittedJavaScriptMismatch. The ordinary witness first agrees across all
four runtimes. The test then appends console.log('planted emitted JavaScript
mismatch') to emitted JavaScript only. The altered program exits normally,
and the comparison rejects the extra bytes against Go. The nested test failure
is intentional and verified by the passing outer test. It is not an unresolved
correctness failure. No new native source-rule mutant was needed for this
evidence-only change; earlier 25 rule/component output mutants, metadata checks
and JSX extraction checks remain documented on unchanged source/runtime inputs.

Commands, each writing exclusively to its retained log:

- bash cloud/setup.sh: Go/clang/Node/submodules 0s, cache warm and total 75s,
  nproc 5, CPU quota 4 cores.
- source /workspace/adamic-tools/env.sh, then
  ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript go test
  ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$' -count=1 -timeout 30m -v.
- go test ./stage1/cohere/lint -run '^TestEmittedJavaScriptMismatch$'
  -count=1 -timeout 10m -v.

Exact logs are gzip-preserved with verified decompression and uncompressed
SHA-256 hashes in validation-wave-30-required-input/streams.json. Only verified
ELF/archive artifacts from six named completed own scratch directories were
removed to free 293109898 bytes; sources and exact observations remain, and the
removal manifest is preserved.

No full repository gate is claimed. External postcss, graphql-js, postcss-parser
and other required correctness checks in packages outside this unit were not
run or represented as green. No such input was omitted from a test claimed as
executed here. The own twelve gates already passed on these unchanged inputs
and were not repeated. Full source-rule JSX findings/fixes/suggestions and their
native/Go timings still require registered checker program/node integration;
parked React analyses still require native HIR/SSA/capture work. Shared lint
comparison elapsed time is a validation cost, not a rule-speed benchmark.
