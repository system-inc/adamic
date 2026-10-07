Built native .a ports of no-new-func, no-new-native-nonconstructor and no-new-wrappers.
Claim c03f8a68 was pushed before implementation f71a535b; evidence is committed separately.
Agreement PASS 80.663s: 98 controls, 71 findings, 287 repository and 77 compiler roots, normal and sanitized.
Three rule mutants, one provenance mutant and one released-registry mutant were caught.
Not covered: full repository gate, malformed-source recovery, and emitted-JavaScript rule comparison.

## Selection and implementation

All nine earlier reservations were complete and pushed through 2d9d4ad9 before
this batch. Fetch audited 341 origin refs, 336 distinct trees and 33 distinct
Markdown claim documents. Descending combined finding volume with lexical ties
excluded base ports and every named origin reservation, including skipped or
released entries. These three were the first unclaimed rules, all zero-volume.
Main and bridge references were count inventories and skipped-type catalogs,
not ports. The reconstructed preclaim snapshot is fourth-selection.json.

Each rule has its own .a file. Runner and support files are also .a. Existing
declaration-lineage checker facts answer first-declaration source-file provenance;
no new bridge questions, registrations, shared harness edits or compiler edits
were needed. The independent Go oracle calls unchanged production rules at
cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. TypeScript v6.0.3 compiler roots
come from 050880ce59e30b356b686bd3144efe24f875ebc8. Submodules are unchanged.

NoNewFunc checks global Function calls and constructions, including static
apply/bind/call access and parenthesized callees or receivers. Its range is the
whole invoking expression, including a bind call whose result is then called.
NoNewNativeNonconstructor checks global Symbol and BigInt under new, reports the
identifier, and deliberately does not unwrap parentheses, matching Go.
NoNewWrappers checks global String, Number and Boolean under new, unwraps
parentheses, and reports the whole new expression. Aliases, property receivers
and source shadows follow Go's decisions, not a name-only approximation.

## Byte equality and sanitizer checks

The independent comparator includes ranges, message identity and full text,
fixes, suggestions and duplicate findings. These three production rules offer
no fixes or suggestions; their empty fields are still serialized and compared.
Only deterministic diagnostic sorting is applied.

| Population | Roots | Findings | Equal bytes | Normal / ASan / UBSan / LSan |
| --- | ---: | ---: | ---: | --- |
| Controls | 98 | 71 | 30883 | PASS |
| Frozen repository corpus | 287 | 0 | 18485 | PASS |
| TypeScript src/compiler | 77 | 0 | 5318 | PASS |

Controls comprise 15 hand-written sources and 83 literal source cases extracted
from pinned upstream-derived Go tables. Expected diagnostics are not copied:
production Go computes every expected byte. Both sides run production default
options; upstream configuration environments are not replayed and computed/joined
source expressions are not extracted. Findings: no-new-func 36,
no-new-native-nonconstructor 11, no-new-wrappers 24. Shadowing, aliases, optional
calls, parentheses, type assertion wrappers, escaped names, Unicode and CRLF
are exercised. Zero large-corpus findings alone do not establish fidelity;
positive controls and comparison-only mutants exercise report paths.

| Mutant | First changed byte caught by Go comparison |
| --- | ---: |
| no-new-func range end +1 | 60 |
| no-new-native-nonconstructor range end +1 | 6499 |
| no-new-wrappers range end +1 | 7644 |
| Invert declaration-file provenance | 57 |

Each diagnostic mutant compiled, exited zero and emitted empty stderr; only the
independent complete-byte comparison caught it. A scratch Go overlay retaining
released programs also compiled: the real query panics with exit 70 and
`invalid or released checker handle`, while the mutant exits zero, failing the
required panic check. Production bridge sources remain unchanged.

## Native timing against Go

Three alternating rounds ran after validation, without competing builds or tests.
Every timed complete output was verified. These process medians include startup,
loading, rule work, output and teardown. Raw phase timing and query counts are
retained in fourth-timing.json; the script is fourth_bench.py.

| Population | Go median | Native median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 132.986 ms | 270.832 ms | 2.037x |
| Compiler | 310.089 ms | 1805.576 ms | 5.823x |

Native is slower in both measured populations. Instrumented validation observed
zero repository checker queries and two compiler queries, because only matching
constructor candidates ask the checker. These are worker observations, not a
general performance prediction.

## Commands and retained evidence

Toolchain setup is reused from this branch: ready 0s, cache warm 86s, total 86s;
nproc 5. Commands source /workspace/adamic-tools/env.sh. Test output is redirected
to logs, never piped.

```
ADAMIC_WAVE_11_FOURTH_ARTIFACTS=/workspace/wave-11-fourth-validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11FourthAgreementAndMutants$'
# PASS 80.663s; fourth-agreement.log

go vet ./...
# PASS, empty fourth-vet.log

gofmt -l stage1/cohere/typeaware/wave_11_fourth_test.go stage1/cohere/typeaware/testdata/oracle_wave_11_fourth.go
# Empty fourth-gofmt.log

git diff --check
# Empty fourth-diff-check.log

python3 /workspace/wave-11-logs/fourth_bench.py
# All timed outputs agree; fourth-timing.log and fourth-timing.json
```

validation-wave-11-fourth retains compressed complete outputs, output hashes,
source hashes, frozen relative manifests, selection snapshot, claim/push logs,
passing validation logs and timing evidence. Executables and archives stay in
scratch. The native compiler loads these .a sources successfully.

## Limits

The full repository gate, malformed-source parser recovery and full upstream
configuration matrix were not run. The pinned cohere CLI's own-lint/format gate
still cannot resolve .a modules; shared integration belongs to
codex/lint-harness-dot-a and was not edited here. Emitted-JavaScript comparison
of the new rules is not included. The filtered Node oracle and full bridge checks
reported in earlier batch reports are prior validation, not rerun evidence for
this batch. There is no unresolved blocker for these native default-rule ports.
