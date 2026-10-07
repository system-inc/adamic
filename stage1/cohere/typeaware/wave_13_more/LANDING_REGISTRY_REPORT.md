Built: rebased wave 13 onto area b4691483, containing current main c7991b90; retained all registry migrations and claimed no new rules.
Commits: rebased code 62ebf352 replaces ada82fba; this evidence commit is pushed only to codex/typeaware-wave-13.
Commands and outputs: original gate PASS 218.610s, next 129/129, more 430/433 in normal and ASan modes; both corpora, metadata, scoped vet and required-input compiler parity pass.
Mutants: nine rule mutants, two export-question mutants, one released-registry mutant and 27 named-listener metadata mutants caught again.
Not covered: three shared parser recovery controls, typed checker-context integration, emitted-JavaScript rule comparisons, full repository gate and the other required-input gates.

The incoming changes migrate legacy syntax rules onto the shared registry and add
syntax helpers to RuleContext. No compiler, bridge, parser or wave 13 source changed
relative to ada82fba. The rebase was clean. The unchanged proof-stage compiler,
bridge archives and continuation binaries were reused; the original test builds
its suites and mutants anew. No shared file was edited by this unit.

The first attempt exhausted the 32 GB workspace. Original controls and sanitized
controls had matched, then mutant setup could not create a directory; the more
validator also stopped on disk exhaustion. Explicitly named obsolete wave 13
scratch binaries were removed, preserving logs and sources. Both interrupted runs
were repeated to completion. This was an infrastructure failure, not a waived test.

The renewed original controls have 19,161 identical bytes and 32 findings; scratch
path headers explain the count difference from the prior report. Repository and
compiler comparisons remain 19,144 bytes/one finding and 7,087 bytes/four findings,
respectively, including sanitized execution. Both continuation trios match Go over
77 compiler roots and 287 repository roots: 4,933 and 18,485 bytes, zero findings.
The normal and sanitized next controls pass all 129; the more controls match 430
of 433. All inputs were retained, including the three failed comparisons.

| Mutant | Required catcher |
| --- | --- |
| unassigned | Go byte comparison at offset 1357 |
| caught | Go byte comparison at offset 4176 |
| exports | Go byte comparison at offset 15111 |
| export-chain-flags | Go byte comparison at offset 14829 |
| export-module-lookup | Go byte comparison at offset 17294 |
| process-state | Go byte comparison at offset 134 |
| race-handle | Go byte comparison at offset 118 |
| blocking-order | Go byte comparison at offset 1690 |
| namespace-alias | Go byte comparison at offset 72 |
| literal-parentheses | Go byte comparison at offset 587 |
| executor-span | Go byte comparison at offset 101 |
| released-registry | required invalid/released-handle refusal |
| nine source listener names | pinned Go listener comparison |
| nine JSON listener names | descriptor comparison |
| nine numeric JSON listeners | named-kind descriptor decoding |

Rule and export-question mutants compile, exit zero and have empty stderr; only
finding bytes catch them. A released handle exits 70 with the expected diagnostic;
its registry mutant exits zero and the refusal assertion catches it. All completed
native control stderr streams were checked for ASan, LeakSanitizer and UBSan errors.

ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus go test
./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$' -count=1
-timeout=15m -v passes in 43.799s. Its required real-compiler input was supplied;
77 files and 28,836,875 expression-tree bytes agree. No selected check skipped.
The original command is TestWave13AgreementAndMutants with the compiler and
repository manifests supplied, -count=1 -v -timeout=30m. Continuation validate.py,
corpora.py and mutants.py run with the pinned production Go adapters. The whole
wave_13_more Go package and go vet ./stage1/cohere/typeaware/... pass. Test output
was written directly to logs. Raw logs, JSON results and streams are retained in
validation/landing-registry.

Single whole-process compiler observations: original native 6.214s versus Go
2.183s, next 3.560s versus 0.571s, more 3.516s versus 0.604s. These runs overlapped
other verification jobs and are not a controlled speed benchmark. The prior
three-round benchmark remains in LANDING_PROOF_REPORT.md. Setup was reused:
previous cloud/setup.sh total 217s and nproc 5.

Remaining failures are exactly 53d1c0a9ffc2fefa (JSX-like slash in .ts),
aff6ced2ca61fe89 (JSX-like element in .ts), and defcb4c4ce6a2921 (yield label and
break yield). Go exits zero; native exits 70 before rules execute. Their exact
inputs remain in the control corpus and diagnostics in the more logs. The shared
RuleContext still has no checker session/bridge-query context for typed factories.
Own suites retain per-file traversal; listener metadata is not proof of supplied-node
integration. Fixes require shared ownership work, which Ahra prohibited this unit
from editing. None of these rules qualifies for the React parking exception.
No new claim was taken and full-green status is not asserted. No production Go
regex occurs in these nine rules; no regex translation or matcher was added.
