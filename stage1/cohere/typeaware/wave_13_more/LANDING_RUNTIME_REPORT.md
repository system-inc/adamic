Built: rebased wave 13 onto area/stage1-lint d65a8f93, retaining the shared harness and new runtime optimizations; no new claims.
Commits: rebased code 56e746e6 replaces pushed 724f948d; this report records renewed runtime integration evidence.
Commands and outputs: original agreement gate PASS 245.144s; next 129/129, more 430/433 normal and ASan controls equal; corpora, runtime ownership, Node, metadata and scoped vet checks pass.
Mutants: nine rule, two export-question, one lifetime and 27 metadata mutants caught again.
Not covered: three shared parser recovery inputs, checker-context integration into supplied-node dispatch, full repository gate and emitted-JavaScript rule comparison.

The requested area branch advanced from 7481e032 to d65a8f93 with runtime release,
string equality and backward-search optimizations. The rebase completed without
conflicts, preserving every incoming change. Stage 0 was rebuilt to embed the new
runtime, and all rule native programs were rebuilt in normal and ASan modes.
The bridge source is unchanged and its existing C archives were reused. The branch
contains current main 39638d9e and the requested harness 41eb6eab2. No protected
compiler file, shared harness, registration generator or leak helper was edited
by this unit. No incoming change was reverted. No main or area branch is pushed.

## Renewed lint comparisons

TestWave13AgreementAndMutants passes in 245.144s: 19,110 control bytes and 32 findings,
compiler 7,087 bytes and four findings, repository 19,144 bytes and one finding.
Normal and sanitized findings, fixes and suggestions match production Go exactly.
The extra control bytes relative to the prior report are the three additional
characters in each scratch-path header, not finding changes. Released handles
refuse with exit 70; the registry mutant exits 0 and is caught by that refusal.

Original mutants unassigned, caught and exports compile and exit normally with
empty stderr, caught only by byte comparison. Export-chain-flags and
export-module-lookup likewise require byte comparison. Their exact offsets and
streams are retained in original.log and streams.tar.gz.

Both continuation suites reran every upstream control under normal and ASan builds.
Next has 129 equal controls. More has 430 equal controls and the same three parser
refusals among 433 controls in both modes: 53d1c0a9ffc2fefa, aff6ced2ca61fe89 and
defcb4c4ce6a2921. The first two need JSX-like recovery in .ts; the third needs a
yield label and break yield. They exit 70 before rules run. No control was omitted,
rewritten or weakened, and no sanitizer diagnostics occur in the retained control
stderr streams. The React analysis parking exception does not apply to this core rule.

Continuation mutants process-state, race-handle, blocking-order, namespace-alias,
literal-parentheses and executor-span compile and exit 0 with empty stderr, caught
only by byte comparison. Both continuation trios agree over all 77 frozen compiler
roots (4933 bytes) and 287 repository roots (18485 bytes), with zero findings.
Their positive controls and mutants provide nonempty evidence.

The metadata package passes nine source-name, nine JSON-name and nine numeric-JSON
mutant rejection checks. Listener metadata retains ast.Kind names. Scoped go vet
passes. Other raw-fact and foundation bridge mutants were not repeated because
the bridge is unchanged; their evidence remains in LANDING_F801_REPORT.md. This
run repeats the original trio's lifetime checks; the continuation released-handle
evidence remains unchanged.

## Runtime checks

TestRuntimeReleasePaths and TestRuntimeStringEquality pass in 18.584s. The release
check uses runtime-built strings, shared owners, a 100,000-object chain and explicit
live allocation counts in both normal and sanitizer builds. The equality check
compares same and distinct headers, unequal strings and undefined against Node.
These are observed passes of the inherited tests, not newly authored checks.

The filtered Node gate passes TestRuntimeLastIndexOfMatchesNode, the registered
runtime_last_index_of.a, inherited_static_field_read.a and class_layouts.a fixtures,
and TestTheOracleCatchesOneByte. Search results agree across Node, emitted JavaScript,
release and sanitized native execution. Existing worker caches were used where
applicable. The runtime owner's semantic release, search and equality mutants are
recorded in internal/native/performance/lint-runtime-profile/REPORT.md; those
separate runtime mutants were not rerun by wave 13. All nine lint-rule mutants
were rerun. The complete repository gate was not run.

## Native time against Go

Three alternating full-output rounds ran after all builds and gates completed.
Every round compares canonical bytes. Measurements include checker loading;
native remains slower than Go. These measurements do not establish a controlled
before/after speed gain for the runtime changes.

| Trio | Population | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 7.547 | 2.327 | 3.24 |
| original | repository | 0.424 | 0.133 | 3.18 |
| next | compiler | 2.895 | 0.450 | 6.44 |
| next | repository | 0.409 | 0.126 | 3.25 |
| more | compiler | 3.318 | 0.610 | 5.44 |
| more | repository | 0.513 | 0.140 | 3.67 |

## Remaining blockers

The shared supplied-node harness exists on this branch. Its RuleContext lacks the
checker session and bridge-question context used by Bindings, Rules and OutputContext,
including cross-file queries. Shared checker-context integration is required to
connect these typed factories to its node callbacks. Ahra prohibited shared
registration-generator and harness edits, so none was added. Own suites still
invoke per-file run() methods; legacy relevance checks remain. Named metadata
alone does not prove kind-indexed execution, and no such compliance is claimed.
The finding wire protocol requires no change.

The three parser recovery controls remain blocked. No additional claim was taken.
No production Go regex occurs in these nine rules; the prior regex audit remains
applicable and no hand-rolled matcher was added. The toolchain was reused after its
prior setup total of 217s; nproc remains 5. Build and test logs, raw streams,
comparison results and isolated timings are retained in validation/landing-runtime.
