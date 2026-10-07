Built: rebased wave 13 onto main b8fb957a and rebuilt native suites after the inherited-static-field fix; no new claims.
Commits: rebased code 5b757a55 replaces pushed aeac7c8e; this report records the renewed landing evidence.
Commands and outputs: original agreement gate PASS 215.974s; next 129/129, more 430/433 controls equal in normal and ASan runs; frozen corpora equal; Node and vet checks pass.
Mutants: nine rule mutants, two export-question mutants, one lifetime mutant and 27 metadata mutants caught again.
Not covered: three shared parser recovery inputs, supplied-node driver integration, full repository gate and emitted-JavaScript rule comparison.

Main added the inherited-static-field native emission fix and its Node fixture.
The rebase completed without conflicts. Stage 0 was rebuilt with go build, and
both continuation suites were freshly built in normal and ASan modes. The original
trio's full gate also rebuilds its native and sanitized suites. The bridge and
cohere pins are unchanged, so the existing normal and ASan C archives were reused.
No protected compiler file was edited by this unit. The main allocator leak-check
helper changes have not appeared in this rebase; none was reverted.

## Renewed checks

TestWave13AgreementAndMutants passes in 215.974s: 18,957 control bytes with 32
findings, 7,087 compiler bytes with four findings and 19,144 repository bytes with
one finding. Normal and sanitized findings, fixes and suggestions agree with the
production Go oracle. Released handles refuse with exit 70.

Original rule mutants unassigned, caught and exports compile and exit 0 with empty
stderr, caught only by byte comparison at offsets 1345, 4136 and 14959. The two
export-question mutants compile and exit normally with empty stderr, caught at
14677 and 17114. The released-registry mutant exits 0 and is caught by the required
released-handle refusal. The original trio has fresh mutant evidence after rebase.

Both continuation suites reran every upstream control under normal and ASan builds.
Next has 129 equal controls. More has 430 equal controls and the same three parser
refusals among 433 controls: 53d1c0a9ffc2fefa, aff6ced2ca61fe89 and defcb4c4ce6a2921.
The first two need recovery for JSX-like syntax in .ts; the third needs a yield
label and break yield. The parser exits 70 before any rule runs. No unsupported
control was omitted or weakened. No sanitizer diagnostics occur in the control
stderr streams. This core rule is not covered by the React parking exception.

Continuation rule mutants process-state, race-handle, blocking-order, namespace-alias,
literal-parentheses and executor-span were rebuilt with current-main stage 0.
Each compiles and exits 0 with empty stderr, caught only by byte comparison.
Both continuation trios agree over all 77 frozen compiler roots (4933 bytes) and
287 repository roots (18485 bytes), with zero findings. Their positive controls
and mutants provide nonempty evidence.

The owned metadata package passes: nine source-name mutants, nine JSON-name mutants
and nine numeric-JSON mutants are rejected. Named ast.Kind declarations and JSON
remain in place; no numeric listener kinds were reintroduced. The filtered Node
oracle passes the new inherited_static_field_read.a and existing class_layouts.a
fixtures, plus TestTheOracleCatchesOneByte. Scoped go vet passes. All commands and
streams are retained under validation/landing-b8fb.

The full repository gate was not run. Other raw-fact and foundation bridge mutants
were not repeated because their code is unchanged; prior evidence remains in
LANDING_F801_REPORT.md. The previous continuation released-handle evidence also
remains applicable to unchanged bridge code; this run repeats the original trio's
released-handle check. The toolchain was reused after its previous setup total of
217s; nproc remains 5.

## Native time against Go

Three alternating full-output rounds ran after all builds and gates completed.
Every round compares canonical bytes. The measurements include checker loading;
native remains slower than Go. No supplied-node driver speed gain is claimed.

| Trio | Population | Native seconds | Go seconds | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 6.747 | 2.247 | 3.00 |
| original | repository | 0.416 | 0.120 | 3.46 |
| next | compiler | 2.802 | 0.485 | 5.78 |
| next | repository | 0.405 | 0.123 | 3.31 |
| more | compiler | 3.194 | 0.427 | 7.48 |
| more | repository | 0.491 | 0.127 | 3.85 |

## Remaining integration limits

Named listener declarations are correct. Own suites still invoke per-file run()
methods over typed bridge contexts, rather than shared supplied-node callbacks.
Legacy relevance checks remain inside those methods; metadata does not prove
kind-indexed dispatch. The missing integration surface is the shared typed driver,
not numeric parser kinds. The ab70f38d4 harness remains outside this main base.
No full syntax-only registry factory/class migration is claimed.

No production Go regex occurs in these nine rules and no hand-rolled regex matcher
was added. The shared regex audit from NAMED_LISTENERS_REPORT.md still applies.
No additional claim was taken while the core parser claim remains partially blocked.
The branch is pushed only to codex/typeaware-wave-13, never main or an area branch.
