Built: retained current main 71d7e491 and incorporated lint area b28757f3, including fixed-source reparsing and disconnected-node guard; no new claims.
Commits: integrated code f8b21baf replaces pushed d00ff1f2; this evidence commit is pushed only to codex/typeaware-wave-13.
Commands and outputs: original PASS 257.996s; next 129/129, more 430/433 normal and ASan; both corpora, metadata, scoped vet and the linked-node guard pass.
Mutants: three original rule, two export-question, one released-registry and 27 metadata mutations caught again; unchanged six continuation rule mutants retain prior evidence.
Not covered: three shared parser recovery inputs, typed supplied-node checker context, emitted-JavaScript typed-rule comparisons, full repository gate and the broader shared fix-engine suite.

The current-main rebase and area merge completed cleanly. Incoming shared changes
reparse fixed source under its original path, preserve cohere's exhausted-fix-budget
answer, and compare results after inserting disconnected node copies. All changes
were retained. Wave 13, compiler, runtime, parser and checker bridge sources are
unchanged from the previous landing. Stage 0, bridge archives and continuation
binaries were reused; the original Go test rebuilds its suites and mutants. This
unit edited no shared source. Setup was reused: previous total 217s and nproc 5.

The first original attempt failed at 157.618s while creating exports-source:
no space left on device. Its log is retained as original-disk-failure.log. The
continuation run also stopped during disk exhaustion. Obsolete owned scratch ELF
binaries and C archives were removed from explicitly scoped wave 13 directories,
preserving source and evidence. Both comparison runs were repeated to completion.
No test or comparison was waived.

The completed original agreement gate passes in 257.996s. Controls match 19,008
bytes and 32 findings; compiler 7,087 bytes/four findings and repository 19,144
bytes/one finding match Go in normal and sanitized runs. Released handles exit 70
with the required invalid/released diagnostic, and the registry mutant exits zero
and fails the refusal assertion. Original command uses TestWave13AgreementAndMutants,
-count=1 -v -timeout=30m, the stage-0/archive/artifact environment variables and
both frozen corpus manifests.

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

Rule and export-question mutants compile, exit zero and have empty stderr; only
byte comparison catches them. Six continuation mutants were not repeated because
their rule, compiler and bridge sources are unchanged; their exact compile/run
and byte-comparison evidence remains in LANDING_STACK_REPORT.md. No renewed mutant
proof is claimed for the new shared linked-node guard, which was integrated and
run as supplied without modifying its harness.

Own validate.py runs every continuation case normally and under ASan. Next passes
129/129; more matches 430/433. The failed case set is checked against result JSON:
53d1c0a9ffc2fefa, aff6ced2ca61fe89, defcb4c4ce6a2921. No fixture was omitted or
rewritten. Both trios match Go across 77 compiler and 287 repository roots: 4,933
and 18,485 bytes, zero findings. All completed native control stderr streams were
checked for ASan, LeakSanitizer and UBSan diagnostics; none occurred. The metadata
package passes in 0.077s and scoped go vet passes.

The incoming shared guard was run as written:
go test ./stage1/cohere/lint -run '^TestNodeTableIsLinkOnly$' -count=1 -v
-timeout=15m. It passes in 158.342s: 2,154 rows produce 13,069,337 identical bytes
with and without disconnected node copies. This verifies the shared driver, not
integration of wave 13's typed factories, which remains blocked. No selected test
skipped and no guard was relaxed. Test output was written directly to logs.

Single compiler whole-process observations:

| Trio | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original | 6.297 | 2.632 | 2.39 |
| next | 3.330 | 0.620 | 5.37 |
| more | 3.772 | 0.543 | 6.94 |

Some verification jobs overlapped these measurements; they are not a controlled
benchmark or proof of speed gains. Native remains slower than Go. Full renewed
streams, logs, JSON results and times are retained in validation/landing-links.

Required real-compiler parity, filtered Node, runtime and options checks were not
repeated; their passing unchanged-source evidence is retained in
LANDING_STACK_REPORT.md. The complete repository gate, other required-input gates,
Stage 3 tests and broader shared fix-engine checks were not run by this unit.

Remaining parser refusals occur before rule execution: JSX-like slash and element
syntax in .ts, and a yield label with break yield. Go exits zero; native exits 70.
The exact frozen controls and retained logs provide their reproducers. RuleContext
still lacks the checker session and bridge-query context for typed factories.
Own suites retain per-file traversal; named metadata alone does not prove supplied-node
or linked-node dispatch. Fixing shared parser/harness code is outside this unit's
territory. The React parking exception does not apply. No new claims or full-green
status are asserted. No Go production regex occurs in the nine rules and no regex
matcher was introduced.
