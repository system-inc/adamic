Built: integrated the six external acceptance programs with the existing assignment-proof oracle; no compiler fix was needed.
Commits: implementation ccf8c172263872a93731658d782d0804d20ac294; fixture dependency e98e49ad8c4e8cf95f12f6be82657eb99aa18828; follow-up tip sent with its push.
Commands/results: strict acceptance checker, all six shared fixture tests, both-backend oracle comparisons and counts update and verification passed.
Mutants: all six acceptance source mutants failed at Node stdout byte comparison after zero stock TypeScript diagnostics.
Not covered: full upstream scanner execution, new check elision, performance measurement, whole-package tests or the full gate; previous implementation mutants were not rerun because production code is unchanged.

This follow-up adds acceptance coverage for roadmap step 12, #cvhj5fk, alongside the seven implementation fixtures. All six ruled outcomes already passed on the delivered branch before any follow-up edit. There were no remaining compile or runtime failures to fix.

Only the requested fixture directory was imported from codex/step12-assign-fixtures. No other worker branch was merged. The six `.a` programs and expectations.json remain byte for byte identical to e98e49ad. Their author attribution and NOTICE are retained. The checker report now says “current branch” and records central oracle registration; status.json records six Compiles outcomes. Historical fixture-author observations remain labeled separately in README.md.

| Acceptance program | Initial result | Final result | Observable proof |
|---|---|---|---|
| 01_scanner_keyword | Pass | Pass | Returns/stores 128, 83 and three 80 misses |
| 02_if_narrowing | Pass | Pass | Present/absent branches print `4 1` then `-1 2` |
| 03_return_defined | Pass | Pass | Defined results, cache writes and one call per result |
| 04_chained | Pass | Pass | Both optional targets get `value1`, one call |
| 05_compound | Pass | Pass | Returns/stores updated x: `12 12 1`, then `14 14 2` |
| 06_wider_target | Pass | Pass | Member-typed result and whole-enum target both 128, one call |

`internal/oracle/assignment_proofs_test.go:12` registers all six programs as compiling, with checked=false, so the oracle holds both the generated JavaScript backend and native backend to stock Node, rather than accepting a compiler check as the result. The uncached comparisons include native ASan/UBSan, Linux leak detection, and native release builds. The original seven implementation fixtures also pass. The focused expression selected four existing optional-field/assignment regressions, which passed too: 17 oracle cases total. The shared fixture runner separately passed all six acceptance programs' Node, native and stage0 lanes.

Compiler flags: native sanitizer `clang -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`; release and counted native `clang -O2`. Tool versions and setup results are unchanged from report.md.

**Proof/check inspection**

Emitted C for all six programs is preserved, compressed, in acceptance-emitted/. The defined string in 03 is saved as a string pointer, stored into cache, retained across internal cleanup and returned without a post-store presence check. In 04 the inner saved value is used by the outer assignment without rereading either target. In 06 the member number is saved, written into the full-enum target and returned without an enum-member comparison or tag check. Compound 05 uses the existing computed-target result path and does not treat y as the result of +=.

Assignment-result propagation adds no proof recheck. Existing guards remain: 01 has one RHS unwrap-presence guard after the source keyword guard, and 02 has one undefined dereference guard for x.length. Both guards occur once on the area baseline and once on this branch. They were not introduced by the saved-result path and are not claimed elided. Target global-initialization checks also remain; stores are still lowered through the original independently validated assignment path. No claim is made that these acceptance programs have zero native checks.

**Acceptance mutants**

Each source mutation passes stock TypeScript 6.0.3, then exits 1 at the source Node stdout byte comparison. The final checker's child processes reran all six. These prove that the acceptance goldens can fail; they are distinct from the compiler-implementation mutants in report.md.

| Program | Mutation | Catcher |
|---|---|---|
| 01 | Return/store Identifier in the keyword branch | Node stdout bytes |
| 02 | Return “missing” instead of undefined for the absent branch | Node stdout bytes |
| 03 | Return compute() without writing cache | Node stdout bytes |
| 04 | Omit b's assignment | Node stdout bytes |
| 05 | Replace += with = | Node stdout bytes |
| 06 | Return value without writing wide | Node stdout bytes |

**Commands and output**

Every test command redirected output to a file. Compressed logs, individual source golden bytes, stock TypeScript logs and mutant tracebacks are in logs/acceptance-*.gz.

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
python3 stage3/fixtures/assignment-proofs/check.py --require-adamic > /workspace/scratch/assignment-proofs/acceptance-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/(internal|stage3)/(oracle|fixtures)/(testdata|assignment-proofs)/(assignment_|0[1-6]_)' -count=1 -v -timeout=10m > /workspace/scratch/assignment-proofs/acceptance-oracle.log 2>&1
go test ./stage3/fixtures -run 'TestFixtures/assignment-proofs' -count=1 -v -timeout=5m > /workspace/scratch/assignment-proofs/acceptance-stage3.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m -args -update-counts > /workspace/scratch/assignment-proofs/acceptance-counts-update.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout=30m > /workspace/scratch/assignment-proofs/acceptance-counts-check.log 2>&1
git diff --check > /workspace/scratch/assignment-proofs/acceptance-diff-check.log 2>&1
```

Acceptance checker: `Compiles=6, NotYet=0, Refused=0, Checker=0`; six stock TypeScript checks, six Node goldens and six caught source mutants. Native/JavaScript oracle passed (2.354s); shared fixture test passed (2.981s); counts update passed (33.473s), and counts verification passed (34.126s). These elapsed times describe test runs, not program performance. Counts add exactly six rows to internal/oracle/counts.md; each has equal allocations and frees.
