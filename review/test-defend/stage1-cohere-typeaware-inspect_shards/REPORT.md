TestFactsDecoderGuards defended within the bounded matrix.
D1 admits positive integers with leading zeros; only Guard failed among18 observed tests.
Package-wide uniqueness remains unknown; production restored and standalone diff saved.

Starting origin/main: bfe0553300773c0b37db2c10df97adeb909705f8. Audit report, rows, code/oracle notes and matrix copied alongside this report. Current list has578 tests; no top-level test additions/removals against audit branch.

CODE UNDER TEST: native-compiled TypeScript frames/facts decoder, not Go cohere or test preparation. See plan.md for reached decoder functions and predefined attempt. ORACLE: self-written valid stdout64 and invalid exit70 plus specific error text. It checks the reason as well as status. Six family uses external Go cohere agreement.

D1 is a condition operand deletion at frames.ts6 against starting commit: `(negative || start + 1 < last)` becomes `negative`. Native sanitized port build succeeded0.827s. The diff applies to origin/main. Matrix compilation and native execution also succeeded; expected test failure is suite_test.go584: noncanonical-integer escaped:<nil>. All17 passed shards are listed in matrix.json. D2 andD3 remained unused fallback ideas, not executed mutants.

Limits and brief friction:
- Full clean package cooked at90.034s with no assertion failure. Broader clean shard selection cooked at90.282s in costly built-in mutant product setup. Narrowed clean matrix passed27.896s; expanded positive-input matrix passed12.253s. This is not proof the full baseline is green, and not package-wide uniqueness.
- The requested Go coverpkg profiles cannot measure TS production compiled in subprocesses. Both empty profiles and coverage-diff.md are retained. Six000 is a sample, not full family coverage.
- The allowed bounded-matrix rule conflicts with wording that defended must mean no other package row catches it. We use defended only within the explicitly listed bounded matrix. Central replay must settle package-wide uniqueness before any deletion claim.
- A timed-out test can leave native-build child processes running temporarily. Inspected the specific child; it had exited before cleanup. No unrelated processes were killed.
- /usr/bin/time was unavailable; retried using Bash time. Failed timing invocation saved only until successful build log replaced it.
- No TypeScript coverage instrumentation was added, since that would change preparation/harness. No tests changed or weakened.

Toolchain: warm env works, setup skipped, Go1.27.1, Node24.19.0, nproc5. npm ci stage3/api performed before baseline; npm-ci.log retained. Exact npm timing was not recorded. Go compiler build timing was not separately recorded. Mutant native rebuild0.827s, matrix77.127s; Go coverage runs7.008s and53.811s. All test outputs redirected to files. Rough session elapsed15minutes.

Guard name/assertions: genuine decoder admission/refusal checks, including canonical integers; no performance promise. The valid control checks root flags64, not every decoded field. We make no claim about strict/present fields, optional corpora, unsampled Six shards, other TypeAware families, other packages or repo-wide uniqueness.
