# u107: lint suggestion products and witnesses

Starting commit: `16f436a16a8e3b9cd2343f449eb1b0b4577c135c` (fetched origin/main). Branch: `test-audit/stage1-cohere-lint-complete_suggestion_products`.

`report.json` is the row deliverable; `matrix.json` maps compact row IDs to full names and records raw failures. `scope-check.json` verifies all 18 requested functions exist. The audit groups these into 12 rows and adds the shared union member from harness_test.go. Product wrappers share one checker. Shard 000 checks different assertions from its siblings, 001-003 compare answers, and 004-005 witness disagreement, so these three purposes stay separate.

CODE UNDER TEST: the Adamic lint driver's main.ts run and suggestions.a constructors for agreement; internal/testguard.Run for the CPU/wait row; suite construction for the product/fixture/setup rows. ORACLE: unchanged Go cohere output, plus hand-written construction and guard assertions. Node executes the port or child fixtures. No Go cohere source or adapter was mutated.

M1-M4 are four code-derived port mutants fixed before any mutant run. The port was rebuilt per mutant with a separate ADAMIC_BUILD_CACHE_DIR under /workspace/u107-cache. M5-M6 mutate the guard that the CPU/wait row directly tests; they are not used as mutations of the lint port. The two guard failure witnesses were excluded from these production matrices and tested by weakening their error reporting instead. S0-S3 are construction checks. W1-W4 weaken the checks witnessed by the corresponding tests. Probe P1B replaces the whole lint run body, P2 replaces the whole guard Run body. Probe failures are never production kills.

Every *.diff is standalone against the starting commit, without a selector. apply-checks.json records git apply --cached --check against the clean starting index. The port mutants and P1B built sanitized native products, as their product test passes show. Each Go construction/witness/guard diff passed go vet for its mutated package. P1-invalid-probe.diff.txt documents a rejected probe, not a compilable replay candidate: retaining unreachable code destroyed TypeScript's control-flow narrowing. Its failures are excluded.

The package baseline timed out at 90 seconds without an assertion failure. The first bounded baseline also timed out in full-registry emitted-mismatch preparation. Excluding that row produced a clean 51.909-second binary baseline. All production matrices are bounded; no claim of package-wide or repository-wide uniqueness is made. The expensive planted-disagreement witness was excluded from production matrices and audited separately. Matrix caller sets and commands are saved in the time files. A native compilation pass is not a lint-output correctness assertion.

Timings are medians of three independent -count=1 selections using the binary's own ok line. Families select all members together. Setup was skipped because /workspace/adamic-tools/env.sh worked; nproc=5. npm ci in stage3/api took 0.768 seconds. /tmp had only 302 MB free, so TMPDIR was overridden after sourcing env.sh to /workspace/u107-tmp. All test stdout/stderr was directed to files. timing-summary.json separates logged product builds from command durations and warns against double-counting nested builds.

Survivor witnesses: M3 changes /*é*/debugger; from column 7 to column 6 (M3-witness-before.log and after.log). M6 changes a 1.5-second Node CPU job with a 3-second budget from guard error=<nil> to an early CPU-limit failure (M6-witness logs and guard-probe.go). These are non-equivalent survivors in the observed slice; other package rows may catch them.

Brief feedback:

* The header says 13 rows but supplies 18 functions. Applying both the product-wrapper family rule and the distinct-assertion exception gives 12 rows here. Generated shard naming alone does not resolve mixed positive checks, fixture validation and built-in witnesses.
* The unit mixes port behavior, product construction and Go process guards. The instruction never to mutate a harness needs a role-based interpretation: Run is the actual code under test for the CPU/wait test, while it is harness code for lint parity. Production guard mutations were bounded to that direct test. Witness error-reporting edits test the reporting contract, not removal of the OS limit or kill itself.
* The union was in harness_test.go, outside the named files. It was included in the agreement family. The isolated cold-shard test was found but remains outside the requested slice.
* TestCompilerGuardBackend is a subprocess helper. Its compilerShardLauncher adapter has no caller found by git grep at this starting commit. Enabling its environment flag alone would exit 125 for missing child arguments. It has no standalone quality verdict.
* Cold package execution spends most of the budget on unrelated setup, then queues requested parallel tests. Even the first slice run spent its budget on emitted-mismatch preparation. Timeouts leave unexecuted rows unknown, not failed.
* The planted-disagreement witness starts six separate processes, each hashing/verifying/preparing products. Three required timing runs therefore cost roughly 150 seconds even with warm products. A requirement to time all rows three times plus native mutant rebuilds and standalone Go vet makes the budget tight.
* Product cache inputs contain port files even for the unchanged Go oracle, so each port change also rebuilds that oracle. Separate caches are correct but add avoidable cost.
* Warm env.sh does not imply sufficient temporary space. It points TMPDIR at /tmp/adamic-gate, on an almost-full filesystem. The brief should permit moving temporary and build-cache directories explicitly.
* Inserting an early empty return can make a formerly valid TypeScript function untypeable because the remaining code is unreachable and loses narrowing. Replace its entire body when needed. Invalid-probe build failures cannot establish vacuity.
* The witness exception and the package-uniqueness definition need to state how production precondition failures in witnesses are counted. This audit excludes them from verdict kills and retains raw observations separately.
* Individual test-package rebuilds for harness edits were costly. This run exceeded the target budget. Two erroneous source anchors in my runner also stopped execution before applying their edits; those were my mistakes, not defects in the brief. I corrected them, preflighted all remaining anchors, and resumed only unfinished checks.

Not covered: other package rows, repository-wide uniqueness, opt-in suites outside this slice, an exhaustive dynamic transitive call graph of parser/scanner internals, and disabling actual CPU-limit/wall-kill enforcement. reach.md records static local declarations and explicitly marks its overapproximation. No outside-authority expected value was claimed or checked. No PR was opened and no main branch was pushed. Production and harness source edits were restored before publication.

Final emitted-row result: its individual clean run timed out at 90 seconds in preparation. No weakened-check run or median of three completed executions is claimed. Verdict cannot-judge.

Measured go-test command wall total: 1362.311 seconds (22.705 minutes), including the invalid-probe attempt and baseline timeouts. Product build durations are nested; see timing-summary.json. Native rebuilds for M1/M2/M3/M4 logged 40.17/21.53/21.28/22.20 seconds respectively; M1 includes waiting for its lowered prerequisite. Reading, vet, editing and publication add elapsed time and were not separately timed. The overall run exceeded the approximate budget.

The large package baseline is saved as baseline.log.gz. Other test logs remain plain files.
