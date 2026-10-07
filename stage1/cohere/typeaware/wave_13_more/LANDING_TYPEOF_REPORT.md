Built: rebased wave 13 onto area d3a37422 and current main b6b1538b, accepting null typeof and lookup-presence fixes; no new claims.
Commits: rebased code 4b13d7f6 replaces pushed 48a31e6f; the evidence commit is pushed only to codex/typeaware-wave-13.
Commands and outputs: original PASS 261.796s; next 129/129, more 430/433 normal and ASan; corpora, metadata, vet, required compiler parity and typeof Node checks pass.
Mutants: nine rule, two export-question, one released-registry, 27 metadata and four typeof mutant tests caught again.
Not covered: three shared parser recovery inputs, typed supplied-node checker-context integration, emitted-JavaScript lint comparison, full repository gate and the other required-input gates.

The rebase completed without conflicts. Incoming typeof changes affect lowering,
native dispatch, nullable lookup slots and union classification. Stage 0 was
rebuilt and both continuation suites rebuilt in normal and ASan modes; the original
agreement test rebuilds its own suites. The raw checker bridge is unchanged, so
its normal and sanitized archives were reused. No shared file was edited by this
unit. Obsolete explicitly named wave 13 scratch binaries were removed to retain
enough disk space; evidence logs and source were preserved. Setup was reused,
with its previously reported 217s total and nproc 5.

TestWave13AgreementAndMutants passes in 261.796s with the compiler and repository
inputs provided. Controls match 19,059 bytes and 32 findings in normal and sanitized
execution; changed scratch-path headers account for the byte-count difference.
Compiler 7,087 bytes/four findings and repository 19,144 bytes/one finding match Go
in both modes. Released handles produce exit 70 and the required invalid/released
message; the registry mutant exits zero and is caught by the refusal assertion.

Both continuation trios match Go across all 77 compiler and 287 repository roots:
4,933 and 18,485 canonical bytes, respectively, with zero findings. Positive
controls supply nonempty findings and suggestions. Next passes all 129 controls
normally and under ASan. More matches 430 of 433 normally and under ASan. All 433
controls ran; neither failed fixture nor diagnostic was rewritten or omitted.
No completed native control stream has an ASan, LeakSanitizer or UBSan error.

| Mutant | Required catcher |
| --- | --- |
| unassigned | Go byte comparison at offset 1351 |
| caught | Go byte comparison at offset 4156 |
| exports | Go byte comparison at offset 15035 |
| export-chain-flags | Go byte comparison at offset 14753 |
| export-module-lookup | Go byte comparison at offset 17204 |
| process-state | Go byte comparison at offset 134 |
| race-handle | Go byte comparison at offset 118 |
| blocking-order | Go byte comparison at offset 1690 |
| namespace-alias | Go byte comparison at offset 72 |
| literal-parentheses | Go byte comparison at offset 587 |
| executor-span | Go byte comparison at offset 101 |
| released-registry | required handle refusal |
| nine source listener names | production Go listener comparison |
| nine JSON listener names | descriptor comparison |
| nine numeric JSON listeners | named-kind decoding |
| restored null typeof observations | Node stdout across five witnesses |
| missing nullable slot treated as present | Node stdout |
| constructor classified as object | Node stdout |
| string literal classified as undefined | Node stdout |

The nine rule and two question mutants compile, exit zero and have empty stderr;
only byte comparison catches them. The typeof tests compile native semantic
mutants and compare actual Node output. The four tests and all seven typeof
fixtures pass the filtered Node oracle in 23.502s. Ordinary harness observations
may use its valid cache; this is not an uncached integration gate. Cache activity
is recorded in node.log. Normal, sanitized and JavaScript comparisons are held
by that oracle; no selected test skips.

The required-input parser check runs with
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus go test
./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$' -count=1
-timeout=15m -v. It passes in 44.856s with 77 files and 28,836,875 expression-tree
bytes compared against the real compiler and Node, including sanitized native
execution. Metadata package passes in 0.037s; scoped go vet passes. No test output
was piped, and no skip, guard relaxation or deletion was used to obtain green.
Other required-input tests and the complete repository gate were not run.

Single whole-process observations during verification:

| Trio | Compiler native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original | 9.207 | 3.058 | 3.01 |
| next | 4.968 | 0.627 | 7.92 |
| more | 3.964 | 0.561 | 7.06 |

These observations overlapped verification jobs and are not a controlled speed
benchmark. Native remains slower than Go. The earlier alternating benchmark is
retained in LANDING_PROOF_REPORT.md; no speed gain is asserted here.

Remaining native parser refusals are exactly 53d1c0a9ffc2fefa (JSX-like slash in
.ts), aff6ced2ca61fe89 (JSX-like element in .ts), and defcb4c4ce6a2921 (yield label
and break yield). Go exits zero; native exits 70 before rule execution. Their raw
inputs remain in the frozen controls and full diagnostics in the retained streams.
Shared parser code and RuleContext were unchanged in this integration. The typed
checker session and bridge-query context are still absent from shared RuleContext;
own suites retain per-file traversal, and named listener metadata alone does not
prove supplied-node integration. Shared harness/parser edits are prohibited for
this unit. The React parking exception does not apply. No new claim was taken.
No Go production regex occurs in the nine rules, so no regex port or matcher was
introduced. All evidence is under validation/landing-typeof.
