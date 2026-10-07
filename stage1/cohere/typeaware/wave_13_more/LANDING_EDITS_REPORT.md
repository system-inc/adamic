Built: retained current main 71d7e491 and incorporated lint area e667e3e1, including multiple automatic edits and path-preserving witnesses; no new claims.
Commits: integrated code c233668f replaces pushed b1b3e96a; this evidence commit is pushed only to codex/typeaware-wave-13.
Commands and outputs: original PASS 331.947s; next 129/129, more 430/433 normal and ASan; both corpora, metadata, scoped vet and shared rule agreement pass.
Mutants: three original rule, two export-question, one released-registry and 27 metadata mutations caught again; six unchanged continuation mutants retain prior evidence.
Not covered: three shared parser recovery inputs, typed supplied-node checker context, emitted-JavaScript typed-rule comparisons, full repository gate and other required-input gates.

The current-main rebase replayed the owned work and prior area patches. Merging the
newer lint area conflicted in fix-proposal collection. The conflict was resolved by
retaining the area's exact expansion of extraFixes; the resulting lint.ts is
byte-identical to origin/area/stage1-lint. This is integration conflict resolution,
not a locally designed shared change. Every other incoming change was retained.
No main or area branch was pushed. Wave 13, compiler/runtime, parser and bridge
sources are unchanged. Existing stage-0, checker archives and continuation binaries
were reused; the original Go test rebuilds its own suites and mutants. Setup was
reused: previous total 217s and nproc 5. Obsolete owned scratch binaries were removed
to free space, preserving source and evidence.

TestWave13AgreementAndMutants was rerun with -count=1 -v -timeout=30m and both
compiler/repository manifests supplied. It passes in 331.947s. Controls match
19,008 bytes and 32 findings, compiler 7,087 bytes/four findings and repository
19,144 bytes/one finding, in normal and sanitized execution. Released handles exit
70 with the required invalid/released diagnostic; the registry mutant exits zero
and is caught by the refusal assertion.

| Repeated mutant | Required catcher |
| --- | --- |
| unassigned | Go bytes at offset 1348 |
| caught | Go bytes at offset 4146 |
| exports | Go bytes at offset 14997 |
| export-chain-flags | Go bytes at offset 14715 |
| export-module-lookup | Go bytes at offset 17159 |
| released-registry | required handle refusal |
| nine source listener names | production Go listener comparison |
| nine JSON listener names | descriptor comparison |
| nine numeric JSON listeners | named-kind decoding |

Rule and question mutants compile, exit zero and have empty stderr; byte comparison
is their sole catcher. Six continuation mutants were not repeated because their
rule, compiler and bridge sources are unchanged; prior fresh compile/run evidence
remains in LANDING_STACK_REPORT.md. No new shared multi-edit or witness-path mutant
proof is claimed in this unit.

Own validate.py reran every continuation control in normal and ASan modes. Next
passes 129/129. More matches 430/433; the failed case set is exactly
53d1c0a9ffc2fefa, aff6ced2ca61fe89, defcb4c4ce6a2921, verified from result JSON.
No fixture was omitted or rewritten. Completed native control stderr streams have
no ASan, LeakSanitizer or UBSan errors. Both trios still match Go across all 77
compiler and 287 repository roots: 4,933 and 18,485 bytes, zero findings. Positive
controls supply nonempty findings and suggestions. Metadata passes in 0.036s and
scoped go vet passes.

The shared gate was run unchanged with go test ./stage1/cohere/lint -run
'^TestRulesAgree$' -count=1 -v -timeout=15m. It passes in 172.213s; Go, Node, emitted
JavaScript and native agree on 13,071,542 bytes for the compared rows. The existing
harness explicitly classifies unsupported parser-recovery rows, as recorded in
shared.log, so this is not a claim of parity on every malformed input. This unit
changed no recovery classification. Own three failed controls remain failing and
retained. No selected test skipped, and output was written directly to logs.

Single compiler whole-process observations:

| Trio | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original | 6.489 | 2.207 | 2.94 |
| next | 3.034 | 0.644 | 4.71 |
| more | 4.783 | 0.900 | 5.31 |

Some verification jobs overlapped these measurements. They are not a controlled
benchmark or evidence of speed gains; native remains slower. Logs, JSON results,
times and raw renewed streams are retained under validation/landing-edits.

Required real-compiler parity and filtered Node/runtime checks were not repeated;
their unchanged-source evidence remains in LANDING_STACK_REPORT.md. The full
repository gate, other required-input gates and Stage 3 tests were not run. No
skip, guard relaxation or deletion was introduced to obtain green.

The three parser refusals occur before rule execution: JSX-like slash and element
syntax in .ts, and a yield label with break yield. Go exits zero; native exits 70.
Their full frozen inputs and retained diagnostics remain the reproducers.
RuleContext still lacks the checker session and bridge-query context for typed
factories. Own suites retain per-file traversal, so metadata alone does not prove
supplied-node or linked-node dispatch. Shared parser and harness feature edits are
outside this unit's territory. The React parking exception does not apply. No new
claims or full-green status are asserted. No Go production regex occurs in the
nine rules and no regex matcher was added.
