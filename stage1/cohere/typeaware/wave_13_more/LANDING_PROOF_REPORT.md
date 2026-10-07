Built: rebased wave 13 onto area/stage1-lint b84a9d93 and main c7991b90, accepting proven predicates, relations and record runtime changes; no new claims.
Commits: rebased code 908cda28 replaces pushed fdc2fdd9; this report records renewed proof integration evidence.
Commands and outputs: original agreement gate PASS 269.959s; next 129/129, more 430/433 normal and ASan controls equal; corpora, proof checks, records, Node and compiler parity pass without skips.
Mutants: nine rule, two export-question, one lifetime, 27 metadata and nine record-runtime mutants caught again.
Not covered: three shared parser recovery inputs, checker-context integration into supplied-node dispatch, full repository gate, the other required-input checks and emitted-JavaScript rule comparison.

Main advanced with predicate-body proofs, satisfies and erased-upcast relation
checks, standalone record tables and Stage 3 measurements. The area incorporated
current main. The rebase completed without conflicts. Stage 0 was rebuilt, all
rule native suites were rebuilt in normal and ASan modes, and the unchanged bridge
C archives were reused. All incoming shared changes were retained; this unit
edited no shared compiler, harness, generator, parser or leak helper file.

## Renewed lint comparisons

TestWave13AgreementAndMutants passes in 269.959s: 19,008 control bytes and 32 findings,
compiler 7,087 bytes and four findings, repository 19,144 bytes and one finding.
Normal and sanitized findings, fixes and suggestions match production Go exactly.
Control byte totals differ from the previous report solely because scratch-path
headers changed length. Released handles refuse with exit 70; the registry mutant
exits 0 and is caught by the required refusal.

Original mutants unassigned, caught and exports compile and exit normally with
empty stderr, caught only by byte comparison at offsets 1348, 4146 and 14997.
Export-chain-flags and export-module-lookup compile and exit normally with empty
stderr and require byte comparison. Exact streams are retained in the evidence.

Both continuation suites reran every upstream control under normal and ASan builds.
Next has 129 equal controls. More has 430 equal controls and the same three parser
refusals among 433 controls in both modes: 53d1c0a9ffc2fefa, aff6ced2ca61fe89 and
defcb4c4ce6a2921. Two require JSX-like recovery in .ts; the third requires a yield
label and break yield. They exit 70 before any rule runs. No control was omitted,
rewritten or weakened. No sanitizer diagnostics occur in the retained control
stderr streams. The React analysis parking exception does not apply.

Continuation mutants process-state, race-handle, blocking-order, namespace-alias,
literal-parentheses and executor-span compile and exit 0 with empty stderr,
caught only by byte comparison. Both trios agree over 77 frozen compiler roots
(4933 bytes) and 287 repository roots (18485 bytes), with zero findings. Positive
controls and mutants provide nonempty evidence. Metadata tests reject nine source
name, nine JSON name and nine numeric-JSON mutations; ast.Kind names remain in use.

## New proof and runtime checks

The four selected lower tests pass in 1.925s: unproven predicate refusal, proven
predicate bodies, invalid relations refusal and valid relation erasure. No
lowering refusal or mismatch appears in any supported wave 13 input.

TestRecordsAgainstNode, TestRecordMutants and TestRecordReadMutants pass in 75.761s.
The independent Node comparisons include own-key order, prototype names, reads,
references, iteration, numeric values and a million-entry workload, with counted
allocation/free evidence. Mutants compile before execution and are caught as follows:

| Record mutant | Catcher |
| --- | --- |
| indices-in-insertion-order | Node stdout comparison |
| uint32-max-as-index | Node stdout comparison |
| deleted-key-iterated | Node stdout comparison |
| overwrite-key-leaked | LeakSanitizer |
| stored-key-freed | AddressSanitizer heap-use-after-free |
| own-slot-null-read | UBSan null member access |
| prototype-membership-restored | exact missing-member diagnostic and exit check |
| missing-read-silent | exact missing-member diagnostic and exit check |
| own-read-checked-as-missing | valid own-hit fixture |

The filtered Node oracle passes proven_guards.a, proven_class_guards.a,
proven_assertions.a, proven_satisfies.a and proven_upcasts.a, plus the one-byte
comparison mutant check, in 26.909s. Required-input compiler parity passes in
43.928s with ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus: all 77 files and
28,836,875 expression-tree bytes agree against real Go TypeScript and Node, with
sanitized native execution. None of these selected checks skips. Scoped go vet
passes. The other required-input checks and complete repository gate were not
run, and no full-gate green is claimed.

Other raw-fact and foundation bridge mutants were not repeated because the bridge
is unchanged; their evidence remains in LANDING_F801_REPORT.md. This run repeats
the original trio's lifetime checks; continuation released-handle evidence remains
unchanged. Toolchain setup was not repeated: its previous reported total is 217s,
and nproc remains 5.

## Native time against Go

Three alternating full-output rounds ran after all builds and gates finished.
Every round compares canonical bytes. Measurements include checker loading;
native remains slower than Go. No controlled before/after speed gain is claimed.

| Trio | Population | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 7.403 | 2.333 | 3.17 |
| original | repository | 0.469 | 0.179 | 2.62 |
| next | compiler | 2.937 | 0.516 | 5.69 |
| next | repository | 0.416 | 0.129 | 3.22 |
| more | compiler | 3.392 | 0.483 | 7.02 |
| more | repository | 0.538 | 0.149 | 3.61 |

## Remaining blockers

The shared supplied-node harness remains present. RuleContext still lacks the
checker session and bridge-query context required by these typed factories and
cross-file questions. Connecting Bindings, Rules and OutputContext to shared node
callbacks requires shared checker-context integration; Ahra prohibited shared
harness and generator edits. None was added. Own suites retain per-file run()
methods and legacy relevance checks, so named metadata alone is not proof of
kind-indexed dispatch. The finding wire protocol needs no change.

The three invalid-syntax recovery controls remain partially blocked, so no new
claim was taken. No production Go regex occurs in these nine rules, and no
hand-rolled regex matcher was added. All renewed commands, output streams,
comparison results and timings are retained in validation/landing-proof. The
branch is pushed only to codex/typeaware-wave-13, never main or an area branch.
