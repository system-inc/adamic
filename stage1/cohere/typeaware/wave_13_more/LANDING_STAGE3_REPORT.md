Built: rebased wave 13 onto current main 71d7e491 and retained lint area bb2ece56; no new claims.
Commits: rebased code f119cbd4 replaces pushed 6bc83e02; this evidence commit is pushed only to codex/typeaware-wave-13.
Commands and outputs: original PASS 277.759s; next 129/129 and more 430/433 normal and ASan; both continuation corpora, listener metadata and scoped vet pass.
Mutants: three original rule, two export-question, one released-registry and 27 metadata mutations caught again; six continuation rule mutants retain unchanged-source evidence in LANDING_STACK_REPORT.md.
Not covered: three shared parser recovery inputs, typed supplied-node checker context, emitted-JavaScript typed-rule comparisons, full repository gate and newly integrated Stage 3 tests.

The incoming main changes are Stage 3 adaptations and meter work. Source comparison
confirms cmd, internal, bridge and stage1 are byte-identical from prior main
4e0bfda5 to current main. Own rule sources, compiler/runtime, parser, checker bridge
and shared lint context remain unchanged. The rebase completed cleanly and the
newer lint-area ancestry was retained with a merge. No shared file was edited.
No main or area branch was pushed. Setup and toolchain were reused: previous
cloud/setup.sh total 217s and nproc 5. Explicitly scoped obsolete scratch binaries
were removed to free disk space, preserving source and evidence logs.

The original agreement gate was rerun with -count=1 -v -timeout=30m, both compiler
and repository manifests supplied, and the unchanged stack stage-0 binary and
checker archive. It rebuilds its own normal/sanitized suites and mutants. Controls
match 19,059 bytes and 32 findings, compiler 7,087 bytes/four findings and repository
19,144 bytes/one finding, including sanitized runs. Scratch-path header length
explains the control-byte difference from the preceding report. Released handles
exit 70 with the required invalid/released diagnostic; its registry mutant exits
zero and is caught by the refusal assertion.

| Repeated mutant | Required catcher |
| --- | --- |
| unassigned | Go bytes at offset 1351 |
| caught | Go bytes at offset 4156 |
| exports | Go bytes at offset 15035 |
| export-chain-flags | Go bytes at offset 14753 |
| export-module-lookup | Go bytes at offset 17204 |
| released-registry | required released-handle refusal |
| nine source listener names | production Go listener comparison |
| nine JSON listener names | descriptor comparison |
| nine numeric JSON listeners | named-kind decoding |

The original rule and question mutants compile, exit zero and have empty stderr;
only byte comparison catches them. The six continuation rule mutants were not
repeated because their rule, compiler and bridge sources are identical; exact
freshly compiled mutant evidence from the prior landing remains in
validation/landing-stack and its report. No new mutant success is claimed for them.

Continuation validate.py ran all cases with unchanged-source native and ASan
binaries. Next passes 129/129 in both modes. More matches 430/433 in both modes.
All fixtures ran. The exact failed case set is 53d1c0a9ffc2fefa, aff6ced2ca61fe89,
defcb4c4ce6a2921, checked against retained result JSON. All completed native control
streams were checked for ASan, LeakSanitizer and UBSan errors; none occurred.
Both trios still match Go over all 77 frozen compiler and 287 repository roots:
4,933 and 18,485 bytes, zero findings. Metadata package and scoped go vet pass.
Tests wrote output directly to logs, and no selected check skipped.

Single whole-process compiler observations:

| Trio | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original | 7.502 | 2.638 | 2.84 |
| next | 3.847 | 0.597 | 6.45 |
| more | 4.631 | 0.564 | 8.21 |

Some verification jobs overlapped these observations, so no controlled speed gain
is asserted. Native remains slower. Previous alternating measurements remain in
LANDING_PROOF_REPORT.md. All renewed logs, JSON results, times and raw streams are
retained under validation/landing-stage3.

Required real-compiler parity, filtered Node, runtime and decoded-options checks
were not repeated: their compiler/runtime/parser/harness sources and required
inputs are unchanged, and their passing evidence remains in LANDING_STACK_REPORT.md.
The full repository gate, other required-input gates and newly integrated Stage 3
adaptation/meter tests were not run by this unit. No gate was skipped, weakened or
deleted to obtain green.

The three parser refusals occur before rule execution: JSX-like slash and element
syntax in .ts, and a yield label with break yield. Go exits zero; native exits 70.
Inputs remain in the frozen corpus and full diagnostics in retained logs. The shared
RuleContext still lacks the checker session and bridge-query context for typed
factories. Own suites retain per-file traversal, so metadata alone does not prove
supplied-node dispatch. Shared parser and harness edits are outside this unit's
territory; the React parking exception does not apply. No new claim was taken and
no full-green status is asserted. No production Go regex occurs in the nine rules
and no regex matcher was added.
