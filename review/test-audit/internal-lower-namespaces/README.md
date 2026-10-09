# u038: internal/lower namespaces test audit

Starting origin/main: `1f34d0d300301faebc94d397adee1a090923d2c7`.
All nine requested names exist in `go test -list . ./internal/lower/` and remain
in namespaces_test.go. No requested row moved or vanished. There are no families
or witness rows in this unit. The package lists 238 top-level tests.

Code under test: Go internal/lower namespace preflight, call graph, readiness
IR, returned assignment handling, and mutable-view invariance. Oracles are self:
handwritten diagnostic types/text and IR binding assertions. No outside runtime
is run by these nine rows. lowerSource calls Load to prepare checked input, then
Lower, which is the production entry these rows judge. Only Lower received an
empty-answer probe; Load is outside this unit's code under test.

The complete reached-function inventory was produced before mutation, using a
coverage run over all nine rows. functions-reached.txt lists 307 functions with
origin/main locations. functions-all.txt and coverage.out include the remaining
unreached functions and the underlying coverage evidence.

Sixteen mutations were fixed before checking any outcome: fifteen ordinary
production mutations and E1, the Lower empty-answer probe. A preliminary set of
20 was reduced before any mutant test to omit M03, M04, M09 and M19. The retained
IDs therefore have gaps. No inserted-statement supplemental mutant was used.
Early-return patches use `if true` to remain acceptable to go vet. The switched
source was built once. Each independent .diff applies to the starting commit
without any selector or harness change, and every diff passed go vet in an
isolated checkout. M11 and M12 drop complete loops, so no unused variables remain.
No test, checker oracle, fixture, or comparison harness was changed.

The warm env.sh worked: Go 1.27.1, Node 24.19.0, nproc=5. Setup was skipped;
mandatory stage3/api npm ci passed in 1.276 seconds. The clean full baseline
passed in 48.788 test-binary seconds and 52.117 outer wall seconds. Each requested
row then ran three times alone with count=1. report.json contains the median of
the package's own printed `ok ... seconds` line, not Go command startup time.

M01 completed a full-package matrix and failed only TestNamespaceTypesErase.
M02's full matrix reached the 90-second outer budget and was stopped, leaving
uncompleted observations unknown. All nine rows were then rerun individually.
The remaining matrices were narrowed to the nine rows to honor the per-step and
unit budget, and M02 was repeated over that bounded set. report.json sets bounded
true for all rows and lists the complete nine-row matrix. Unique kills other
than M01 are claims only about this bounded set. M14 and M15 are bounded survivors;
we did not establish package-wide survival for them. Package and repo-wide
uniqueness remain central replay work.

E1 returned nil, nil from Lower. It caused TestNamespaceTypesErase to panic;
all nine rows were rerun individually after that aborted matrix. Each failed.
All row-level vacuous fields are therefore false. The nil panic is retained as
observed evidence, not silently treated as a diagnostic comparison. The positive
debug_log.a and debug_class.a subcases accepted E1, but their mixed top-level row
failed its negative class_merge.a subcase. Row-level vacuity masks that distinction.

M14 and M15 have changed-output witnesses. The ready-bits probe emits a stable
Lower summary. Two baseline runs were byte-identical. M14 changes
product.main[0].Value.Value from false to true; M15 changes
product.main[2].Value.Value from true to false. Native execution was not needed
for these IR observations and was not performed. Every compiler mutant run set
ADAMIC_NATIVE_SPLIT=u038-ID so any native products have separate cache identities.

Commands, exact failure lines, passed and failed top-level names, panics, and
outer timings are in the per-run JSON files. Raw output is preserved as .log.gz.
The first E1 matrix is partial; use its individual reruns for row conclusions.
The interrupted M05 full-package log and cooked M02 full matrix are retained
without claiming their unobserved rows passed. The individual .diff files are
central replay inputs. switched-production.patch preserves the scratch switch;
production source was restored before committing this evidence.

A replay, from the starting commit, can apply one independent patch:

    git apply review/test-audit/internal-lower-namespaces/M14.diff
    go vet ./internal/lower/ > M14-vet.log 2>&1
    ADAMIC_NATIVE_SPLIT=u038-M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > M14.log 2>&1

## Brief feedback and execution costs

1. The new median definition, explicit family counting, supplemental-menu rule,
   bounded flag, and panic reruns resolved ambiguities from the previous unit.
2. The verdict taxonomy still misses a slower sole subsumer. Tracing's two kills
   are also caught by Limits, whose median is 0.851 versus 0.096 seconds, about
   8.86 times slower. It is not subsumed within 20%, but literal overlapping
   excludes it because one other row catches everything. We label it overlapping
   with an explicit timing-aware interpretation and retain it. Define overlapping
   as absence of one eligible subsumer, including the timing requirement.
3. The full baseline was below 90 seconds, but M02's full mutant run was not.
   The baseline-only narrowing condition does not state what to do here. We used
   the general stop-and-narrow budget instruction and marked all later results
   bounded. Explicitly allow narrowing after a mutant run cooks too.
4. A compiler slice still reaches hundreds of common lowering helpers. The
   instruction to list every function is feasible with coverage; reading every
   reached implementation in full is not a realistic 20-minute requirement.
   The functions selected for mutation and their supporting namespace/invariance
   code were inspected, with the complete reachability inventory retained.
5. Standalone go vet validation means additional builds even when the matrix
   source compiles once. Initial isolated verification used symlinked relative
   dependency paths, which missed warm caches. Three vet attempts cooked at
   90 seconds. Canonical absolute dependency paths and warmed caches resolved
   this; all final vet runs passed. Dependency identity was preserved, and no
   modified go.mod is part of any replay patch. Budget verification separately
   or recommend an overlay/canonical checkout method.
6. The brief asks for timeout 120 but also stopping a step at 90 seconds. The
   driver enforced the stricter 90-second limit and killed the audit processes
   when narrowing. State whether the 120 seconds is only a backstop.
7. The entry-probe rule is ambiguous for a helper that calls Load and then Lower.
   Mutating Load would leave the unit and its stated code under test. We chose
   Lower and documented that choice before mutation. Specify semantic entry
   rather than every transitive preparation API.
8. A nil pointer is the literal empty return value, but the erasure row detects
   it by panic. This does not establish rejection of an allocated empty IR.
   Consider distinguishing nil-answer probes from structurally empty answers.
9. A mixed positive/negative row can have vacuous false while its positive
   subcases accept an empty answer. E1 demonstrates this in the Debug row.
   The requested row-level count is followed; a subcase annotation would retain
   the additional finding without splitting the row.
10. Survivor and unique-kill statements must always carry their bounded scope.
    The central replay can resolve the package rows outside this unit; this run
    cannot establish that M14 or M15 survive the complete package.
11. I initially ran verification concurrently with full matrices. Cold compiler
    work made those runs contend. The logged measurements and cooked runs remain
    visible. This was an execution choice that cost time, not a test defect.
12. The audit took about 27 minutes, exceeding the approximate 20-minute target.
    metadata.json records measured setup, builds, baselines, matrices, and final
    vet timings. Partial and cooked attempts are retained rather than omitted.

No other packages' tests, native runtime comparisons, external authorities,
sanitizers, or repository-wide replay were run. No pull request was opened.
