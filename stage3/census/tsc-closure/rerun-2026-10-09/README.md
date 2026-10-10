Hidden share: 54.250473% before; 34.839372% after.
Hidden bytes: 5,430,761/10,010,532 before; 3,488,582/10,013,332 after.
Compiler base: 89ac4a8c1de0b02d95965be72f7f8bf1c92433d2; measured main: 5e33a17b186a8a2218d27b69b21e2de5acc5b750.
Largest three movers: NotYet: a BinaryExpression with &lt;payload&gt; (-459,018 bytes); NotYet: a PrefixUnaryExpression on &lt;payload&gt; (-390,437 bytes); NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class (-329,695 bytes).
Known-answer fixture, both root-drop mutants and new-result byte-mask audit pass; native execution and full gate remain outside coverage.

## Before and after

| Metric | Before, 89ac4a8c | After, 5e33a17b |
| --- | ---: | ---: |
| Reached source files | 81 | 81 |
| Source bytes | 10,010,532 | 10,013,332 |
| Hidden bytes | 5,430,761 | 3,488,582 |
| Hidden share | 54.250473% | 34.839372% |
| Blocked union bytes | 7,543,130 | 6,683,211 |
| Independently examined bytes subtracted | 2,112,369 | 3,194,629 |
| Latent NotYet | 1,473 | 810 |
| Latent Refused | 5,164 | 4,070 |
| Latent SkippedDependency | 3 | 5 |
| Latent panic | 3 | 2 |
| Latent error | 0 | 0 |
| Full NotYet | 15,288 | 7,112 |
| Full Refused | 12,275 | 7,761 |
| Full SkippedDependency | 3 | 15 |
| Full panic | 25 | 60 |
| Full error | 2 | 2 |
| Checker diagnostics | 324 | 320 |

Hidden bytes changed by -1,942,179; share changed by -19.411101 percentage points. Counts retain the original unique `(kind, where, reason, text)` convention, excluding Boundary bookkeeping. Full panic sites increased from 25 to 60: 57 nil-pointer sites and three ComputedPropertyName sites. Dependency skips increased from three to 15. Both ordinary errors remain unattributed; these stops are retained in the report.

New arithmetic: **6,683,211 blocked union bytes minus 3,194,629 independently examined bytes = 3,488,582 hidden bytes (34.839372%)**. The independent byte-mask audit agrees for all 81 files; a one-byte headline mutant fails `headline hidden total`.

Outside compiler: 564/712 hidden bytes before; 0/712 after. Both outside sources remain in the denominator. The two outside Refused sites are refusal-scan findings and remain in the counts. Hidden arithmetic uses blocked boundaries and diagnosed bodies minus independent coverage.

## Twenty largest reason-family movements

Sorted by absolute hidden-byte change. Negative values credit a decrease; positive values record an increase.

| Reason family | Before hidden bytes | After hidden bytes | Change |
| --- | ---: | ---: | ---: |
| NotYet: a BinaryExpression with &lt;payload&gt; | 476,290 | 17,272 | -459,018 |
| NotYet: a PrefixUnaryExpression on &lt;payload&gt; | 390,437 | 0 | -390,437 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 329,695 | 0 | -329,695 |
| Refused: a number as a condition | 308,708 | 0 | -308,708 |
| NotYet: reading &lt;payload&gt; | 994,631 | 729,732 | -264,899 |
| NotYet: a value of type &lt;payload&gt; | 545,281 | 806,996 | +261,715 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 159,168 | 0 | -159,168 |
| NotYet: a union or optional field read in a program with record storage | 0 | 157,122 | +157,122 |
| NotYet: a function returning &lt;payload&gt; | 548,359 | 399,151 | -149,208 |
| NotYet: a NonNullExpression | 119,596 | 0 | -119,596 |
| Refused: a value as a condition | 116,148 | 0 | -116,148 |
| Refused: a cast the runtime can&#x27;t check | 199,241 | 88,460 | -110,781 |
| NotYet: a function without a body | 89,346 | 0 | -89,346 |
| NotYet: a parameter that isn&#x27;t a plain name | 83,515 | 2,927 | -80,588 |
| NotYet: a field of type &lt;payload&gt; | 69,307 | 0 | -69,307 |
| Refused: Object enumeration with getter or accessor descriptors | 0 | 56,663 | +56,663 |
| Refused: overload 1 of visitNode parameter node cannot be served by implementation parameter node | 0 | 38,191 | +38,191 |
| Refused: a string as a condition | 32,003 | 0 | -32,003 |
| NotYet: checked view field modifiers of type NodeArray&lt;Modifier&gt; \| NodeArray&lt;ModifierLike&gt; \| undefined | 0 | 29,667 | +29,667 |
| NotYet: a ModuleDeclaration | 22,557 | 0 | -22,557 |

The measurement uses the same upstream TypeScript pin `050880ce59e30b356b686bd3144efe24f875ebc8` (6.0.3) and measurement-tool pin `f1502d130bc0b4440f29b93b76b7d18bff3f6a60`. Cohere remains `7945d102a6c18dd36adf9114a758ce646e8b2359`; typescript-go remains `d92d9bfee114c80be2c375d72edae966176e3a4f`. Stock TypeScript independently confirms the same 81-source closure, including two sources outside `src/compiler`.

Only reached source files form the denominator. Unreached JSON assets and other regular directory files remain excluded. Current main's adaptations changed 25 source hashes and added 2,800 bytes. This compares compiler plus adaptation changes, rather than isolating compiler commits. The full closure, edges, source hashes, per-file counts and exact reasons are in [RESULT.json](RESULT.json); raw inputs and logs are under [evidence](evidence/).

Reason credits are descriptive. Each reported hidden byte is assigned once to its smallest blocking span; ties use start, end and lexical family. Diagnostic kinds remain distinct, and payload names/types are folded by the explicit prefixes in [compare.py](compare.py). Checker-owned and dependency-owned diagnosed bodies have separate families. The credits sum to each report's hidden-byte headline. Overlapping stops cannot inflate the table, and changes in credit do not establish causal credit for an individual commit. [MOVERS.json](MOVERS.json) retains every family and signed change.

## Reproduction and compatibility

Run [reproduce.sh](reproduce.sh) from the repository root with a new scratch directory. It follows the original Reproduce sequence: setup, unchanged main adapters, pinned preparation, explicit closure, first-error and full ledgers, exact-manifest stock view, unchanged `measure.py` and pinned `hidden.py`, fixture and independent byte-mask audit. Every expensive command has a hard limit; the full ranges print progress every 30 seconds. Check the latent log manually at least once a minute. Build concurrency is two.

The following failures were observed and minimally adapted:

- The shallow checkout lacked the tool pin. `prepare.py` failed with `fatal: not a tree object`; fetching that exact SHA restored extraction.
- Cold setup exceeded its initial 240-second limit (exit 124). Its bounded retry exited 1 without a failure diagnostic. An adapter retry explicitly failed with `index.lock write error. Out of diskspace`; `/tmp` was full during the retries. Disk exhaustion is a possible explanation for the undiagnosed setup/build failure, not an observed diagnostic from those commands. The final setup passed with `GOFLAGS=-p=2`; the unchanged adapter script passed on workspace disk under a 600-second limit.
- Preparation generated the overlay but its first build exited 1 without a diagnostic. The same overlay and workspace built successfully with concurrency two under a 300-second limit.
- Main added the unexported derived backend cache `ir.Program.argumentFacts`. The pinned reflective snapshot copier produced `latent state copy: unexported IR field argumentFacts`. [compatibility.py](compatibility.py) leaves only this cache nil in copied snapshots. Its accessor recomputes it from copied public state; lowering does not call the backend cache accessors. Other unexported-field guards remain intact. Restoring the rejecting copier fails the known-answer fixture. Production compiler source is unchanged.
- Main added a-check metadata headers to both existing fixture files. The unmodified runner's fixed byte answer failed. The scratch runner strips only those headers, restoring both authored fixture bodies byte-for-byte to commit `002aaaab`; their hashes and identity checks are archived. Existing repository fixtures are untouched.
- Ten-minute trials completed only 28 latent and eight full files. The unchanged latent retry completed all 81 roots under a one-hour limit. Full mode spent about 34 minutes in `checker.ts`. Its nine completed file records were checkpointed before stopping duplicate work. [file-range.py](file-range.py) selects output files at the full-mode file loop, while every worker loads the complete original root manifest and prepares the same fresh per-file lowering state. It changes neither eligibility nor units, boundaries, checker roots or byte arithmetic.
- The remaining full work was scheduled as 19 command-line/factory files, two module files, 14 parser-through-transformer files and 37 transformer-through-end files. Completed workers were stopped before duplicate work. The module range exited 0. A workspace restart preserved workers and files but lost their original tool handles; `/proc` exit fields were inconclusive, so they are not treated as exit-status evidence. The final join requires complete root coverage and identical checker headers. A separate eight-file witness agrees with every original unit, finding and boundary, allowing only finding order. Its equality check and raw records are archived.
- The original byte-mask audit expected a bare compiler-only report. [audit.py](audit.py) unwraps the closure report and reads its full source root; the pinned independent byte-mask algorithm is unchanged. A root-regression mutant fails on the outside file.

No partial trial is used as RESULT.json. [merge.py](merge.py) rejects incomplete coverage, changed checker headers and disagreeing overlaps before writing a joined ledger. The reproduction recipe uses four disjoint ranges; the actual run also scheduled the two module files separately. All final source identities are checked against the manifest after the runs.

## Checks and mutants

All test processes wrote output to log files. No existing measurement file, compiler implementation, fixture or counts file was edited; no fixture was added.

| Check or mutant | Command / observation |
| --- | --- |
| Known-answer fixture on compatible main binary | `timeout 90 python3 /tmp/closure-fixture-runner/test_fixture.py /tmp/closure-tools-main/census /tmp/closure-tools-main/tree/stage3/census/hidden /tmp/closure-fixture-compatible`: PASS, two sources, TS2322, 54/183 hidden bytes, exact outside range `[33,87)` |
| Same fixture on range binary | Same runner with `/tmp/closure-range/census`: PASS, same answer |
| Real latent root-drop mutant | Outside file remains imported and on disk; omission caught by `closure coverage mismatch: ['src/outside.ts']` |
| Real full root-drop mutant | Same omission in full roots; caught by the same closure-coverage check |
| Restore rejecting IR copier | Known-answer assertion fails; full ledger contains the `argumentFacts` snapshot panic |
| Pinned hidden arithmetic tests | `timeout 90 python3 -m unittest discover -s /tmp/closure-tools-main/tree/stage3/census/hidden -p test_hidden.py`: eight tests, OK, including 500 random byte-set cases |
| Pinned full continuation audit | `timeout 90 python3 /tmp/closure-tools-main/tree/stage3/census/latent/full_audit.py /tmp/closure-tools-main/census /tmp/closure-full-audit`: PASS, three distinct refusal sites, supported string/number conditions, nested checker exclusion, eligible sibling, rollback and no-output guards |
| First-error-only mutant | Full audit catches the missing refusal sites |
| Failed-state-retention mutant | Full audit catches the poison binding left after failure |
| Byte-mask fixture audit | PASS, 54 hidden bytes across both roots |
| Audit compiler-root regression | Fails `file hidden bytes: src/outside.ts` |
| Reason-credit fixture | Exactly 54 bytes credited to the diagnosed-body family; a one-byte headline mutant fails the credit-sum assertion |
| Baseline reason-credit audit | PASS, exclusive credits sum to 5,430,761 |
| File-schedule outside omission | Both checker roots remain in the header; join fails `incomplete root coverage` |
| Corrupted overlap finding | Join fails `overlap mismatch` |
| Changed checker-root header | Join fails `checker program/header mismatch` |
| Eight-file scheduling witness | PASS, all eight original records and complete checker header agree |

The new-result byte-mask audit and its one-byte headline mutant are recorded above with the final arithmetic. The audit independently reconstructs every byte range, union, subtraction, share and largest-region ranking without using `hidden.py`'s union implementation.

Final setup timing: Node 0.021s, Go 0.022s, submodules 0.067s, markdown dependencies 0.074s, clang 0.159s, shared cache 0.452s, Go build 45.728s, cache warm 45.820s, done 45.844s. `nproc` is 5; cgroup CPU quota is 4. Tool versions are Node 24.19.0, Go 1.27.1 and clang 20.1.8. The printed environment path is `/workspace/adamic-tools/env.sh`. Full timing lines, binary and overlay hashes, unchanged compiler diff, trial failures and successful logs are archived.

The 30-minute first-green target was missed. At the dispatcher's explicit request, partial commit `e859bcd2ac4f00331582386199df55b827e225bd` was pushed with the current Reproduce step, compatibility failures and a remaining estimate of 30 to 90 minutes. Its lane checks passed in 0.6 seconds. Work continued to this completed result; no PR was opened.

This measures source coverage on a checker-rejected program. Native execution, JavaScript behavioral equivalence, adaptation-oracle reruns and the full Go gate are outside this unit's coverage. Panics, ordinary errors, dependency skips and unattributed sites remain visible in RESULT.json. The named censuses and fixtures were run; whole-package confirmation was not run.
