Added numeric SyntaxKind listener declarations to all fifteen implemented wave-11 rules; no new claims.
The branch remains based on current origin/main e8ba3d5d, with the previous pushed tip fab76881.
All five independent Go-byte suites, numeric registration oracle, source Node, sanitizers and package vet passed.
The 32 existing mutant observations passed again; wrong-kind and missing-listener mutants were caught only by Go comparison.
Uncovered: handed-node numeric dispatch, three React HIR/SSA ports, full repository gate and rule emitted-JavaScript comparison.

## Declaration contract and API limit

Each owned rule module exports `listenerKinds: readonly number[]`. The constants
in wave_11_syntax_kinds.a use the pinned typescript-go parser's numeric SyntaxKind
values. An independent Go command reads the unchanged production rules' actual
registration maps with a real checker; it does not copy an expected-kind table.
The native .a probe and original .a sources run on Node match its fifteen lines
byte for byte. Both declarations and constants have comparison-only mutants.
This covers production defaults; configurable listener subsets are not tested.

| Rule | Numeric kinds |
| --- | --- |
| @typescript-eslint/no-for-in-array | 250 |
| @typescript-eslint/prefer-regexp-exec | 214 |
| nexus/correctness-no-collection-misuse | 213,214,227 |
| nexus/correctness-no-discarded-outcome | 245 |
| nexus/correctness-no-discarded-pure-result | 245 |
| nexus/correctness-no-identical-branches | 228,246 |
| no-eval | 79,212,213,214 |
| no-extend-native | 214,227 |
| no-func-assign | 307 |
| no-new-func | 214,215 |
| no-new-native-nonconstructor | 215 |
| no-new-wrappers | 215 |
| no-throw-literal | 258 |
| no-useless-backreference | 13,307 |
| prefer-arrow-callback | 307 |

No shared parser, registration generator, driver or harness was edited. ParseNode
in stage1/typescript/parser/nodes.ts currently exposes only `kind: string`, and
there is no handed-node numeric callback API. These declarations prepare the
forthcoming kind-indexed driver. Existing run() traversals, string-kind reads and
per-rule node fetches are unchanged and do not yet meet the new dispatch speed
rule. Converting that execution path is blocked on the shared numeric parser and
driver API under Ahra's own-file restriction. No speed improvement is claimed.
SourceFile listeners above match Go's registrations and retain their file-level
binding/regexp analyses; they were not replaced with narrower guessed kinds.

The React reservations remain blocked on native HIR lowering, SSA, captures,
post-dominators and memoization erasure/inlining, plus the recorded JSX parser
gap. WAVE_11_SIXTH_REPORT.md maps those prerequisites. No stubs, new reservations
or shared-file changes were made.

## Revalidation commands and output

Reused the already successful cloud setup: ready=0s, cache-warm=86s, total=86s;
nproc printed 5. Source /workspace/adamic-tools/env.sh before every command.

```
ADAMIC_WAVE_11_LISTENER_ARTIFACTS=/workspace/wave-11-listeners-final go test ./stage1/cohere/typeaware -run '^TestWave11ListenerDeclarationsAndMutants$' -count=1 -v
# PASS, 56.100s; 15 sets; wrong-kind byte 79; missing-listener byte 406

go vet ./stage1/cohere/typeaware
# exit 0, empty output

git diff --check
# exit 0, empty output
```

The five existing suites used `go test ./stage1/cohere/typeaware -run <filter>
-count=1 -v` with these exact filters:

```
^TestWave11AgreementAndMutants$
^TestWave11NextAgreementAndMutants$
^TestWave11ThirdAgreementAndMutants$
^TestWave11FourthAgreementAndMutants$
^TestWave11FifthAgreementAndMutants$
```

Complete argv vectors and zero exit codes are in results.json. Each used its own
ADAMIC_WAVE_11_ARTIFACTS, ADAMIC_WAVE_11_NEXT_ARTIFACTS,
ADAMIC_WAVE_11_THIRD_ARTIFACTS, ADAMIC_WAVE_11_FOURTH_ARTIFACTS or
ADAMIC_WAVE_11_FIFTH_ARTIFACTS directory and
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript. Output went directly to
log files. These were run concurrently; durations and process timings therefore
include contention and are not a quiet performance comparison.

| Batch | Suite wall seconds |
| --- | ---: |
| first | 159.703 |
| next | 317.890 |
| third | 151.238 |
| fourth | 140.140 |
| fifth | 181.918 |

All five compared complete findings, fixes and suggestion fields on positive
controls and the frozen 287 repository/77 compiler roots, both normal and
ASAN/UBSAN/LSAN execution. Control findings remain 53/67/191/71/259, totaling 641;
all suggestions are empty for these defaults. Fifth still excludes two logged
strict-parser octal inputs (374 of 376 controls); that limitation is unchanged.
Both released-handle panic-70 checks and registry mutants passed in every batch.

## Every mutant observation

| Batch | Observation |
| --- | --- |
| first | regexp mutant: exit 0, empty stderr, independent Go bytes catch byte 469 |
| first | array mutant: exit 0, empty stderr, independent Go bytes catch byte 13891 |
| first | branches mutant: exit 0, empty stderr, independent Go bytes catch byte 16641 |
| first | signature mutant: exit 0, empty stderr, independent Go bytes catch byte 20672 |
| first | index mutant: exit 0, empty stderr, independent Go bytes catch byte 13888 |
| first | syntax mutant: exit 0, empty stderr, independent Go bytes catch byte 1328 |
| first | released-registry mutant exits 0, caught by required panic 70 |
| next | collection mutant: exit 0, empty stderr, independent Go bytes catch byte 71 |
| next | outcome mutant: exit 0, empty stderr, independent Go bytes catch byte 24573 |
| next | pure mutant: exit 0, empty stderr, independent Go bytes catch byte 14725 |
| next | lineage mutant: exit 0, empty stderr, independent Go bytes catch byte 13343 |
| next | awaited mutant: exit 0, empty stderr, independent Go bytes catch byte 25478 |
| next | released-registry mutant exits 0, caught by required panic 70 |
| third | eval mutant: exit 0, empty stderr, independent Go bytes catch byte 68 |
| third | extend mutant: exit 0, empty stderr, independent Go bytes catch byte 19730 |
| third | assign mutant: exit 0, empty stderr, independent Go bytes catch byte 28029 |
| third | global mutant: exit 0, empty stderr, independent Go bytes catch byte 1758 |
| third | anchor mutant: exit 0, empty stderr, independent Go bytes catch byte 35429 |
| third | released-registry mutant exits 0, caught by required panic 70 |
| fourth | func mutant: exit 0, empty stderr, independent Go bytes catch byte 70 |
| fourth | nonconstructor mutant: exit 0, empty stderr, independent Go bytes catch byte 6559 |
| fourth | wrappers mutant: exit 0, empty stderr, independent Go bytes catch byte 7724 |
| fourth | global mutant: exit 0, empty stderr, independent Go bytes catch byte 67 |
| fourth | released-registry mutant exits 0, caught by required panic 70 |
| fifth | throw mutant: exit 0, empty stderr, independent Go bytes catch byte 747 |
| fifth | backreference mutant: exit 0, empty stderr, independent Go bytes catch byte 136 |
| fifth | arrow-fix mutant: exit 0, empty stderr, independent Go bytes catch byte 10790 |
| fifth | provenance mutant: exit 0, empty stderr, independent Go bytes catch byte 132 |
| fifth | regex-path mutant: exit 0, empty stderr, independent Go bytes catch byte 4618 |
| fifth | self-resolution mutant: exit 0, empty stderr, independent Go bytes catch byte 10506 |
| fifth | symbol provenance question mutant: exit 0, empty stderr, independent Go bytes catch byte 132 |
| fifth | released-registry mutant exits 0, caught by required panic 70 |
| listener declarations | wrong-kind: exit 0, empty stderr, Go comparison catches byte 79 |
| listener declarations | missing-listener: exit 0, empty stderr, Go comparison catches byte 406 |

## Native time against Go

These full-process observations are concurrent, not evidence of a dispatch speed
improvement. Prior implementation reports retain the quiet measurements and
query counts. This run's exact measurements follow.

```
first: repository go process=607.951731ms cohere: load_ns=377019694 rule_ns=1605398 run_ns=163514226
first: repository native process=1.152040537s tsgo: load_ns=407291269 query_ns=0 queries=0 first_query_ns=0 run_ns=723467091
first: compiler go process=563.866803ms cohere: load_ns=453117907 rule_ns=15699285 run_ns=75471169
first: compiler native process=2.895598041s tsgo: load_ns=430646252 query_ns=87414327 queries=74 first_query_ns=8725293 run_ns=2399247358
next: repository go process=474.666853ms cohere: load_ns=193293525 rule_ns=58500531 run_ns=261491783
next: repository native process=676.871893ms tsgo: load_ns=145632871 query_ns=177835068 queries=8128 first_query_ns=277678 run_ns=501269318
next: compiler go process=814.867626ms cohere: load_ns=230827768 rule_ns=505502958 run_ns=563961569
next: compiler native process=39.805094378s tsgo: load_ns=258739743 query_ns=29197747596 queries=91563 first_query_ns=6203698 run_ns=39494358579
third: repository go process=434.181283ms cohere: load_ns=209990169 rule_ns=38689167 run_ns=146580456
third: repository native process=838.999566ms tsgo: load_ns=209887109 query_ns=87072501 queries=643 first_query_ns=175054 run_ns=616512945
third: compiler go process=706.16331ms cohere: load_ns=486893116 rule_ns=88575600 run_ns=165285229
third: compiler native process=4.681104074s tsgo: load_ns=576612111 query_ns=587141966 queries=9190 first_query_ns=3598779 run_ns=4043789805
fourth: repository go process=396.618634ms cohere: load_ns=229683653 rule_ns=5972206 run_ns=116665451
fourth: repository native process=462.303415ms tsgo: load_ns=153559436 query_ns=0 queries=0 first_query_ns=0 run_ns=304625585
fourth: compiler go process=791.908572ms cohere: load_ns=646291951 rule_ns=4563097 run_ns=78793658
fourth: compiler native process=4.167121s tsgo: load_ns=719780790 query_ns=363714682 queries=2 first_query_ns=3723006 run_ns=3423946582
fifth: repository go process=186.538654ms cohere: load_ns=90785457 rule_ns=9017848 run_ns=82829190
fifth: repository native process=418.185373ms tsgo: load_ns=85343077 query_ns=6244981 queries=25 first_query_ns=1148773 run_ns=324149613
fifth: compiler go process=463.927732ms cohere: load_ns=295792780 rule_ns=76393521 run_ns=143396876
fifth: compiler native process=2.838502243s tsgo: load_ns=284870266 query_ns=336318160 queries=460 first_query_ns=9525560 run_ns=2536428769
```

## Evidence and limits

validation-wave-11-listeners retains exact Go/native/sanitizer streams, generated
controls/manifests, all test logs, numeric listener lines, current source hashes,
commands and results. Diagnostic field/path bytes are not normalized. The original
Go Cohere and typescript-go revisions remain unchanged from the landing report.
The original-source Node listener probe passed; emitted JavaScript for this probe
refused the unlinked native checker import at rules.ts:104:16, exit 1. That backend
comparison is not claimed. The full repository gate and inherited baseline rule
suites were not rerun for this metadata-only change. No compiler, runtime or bridge
implementation changed. The prior landing report retains their earlier validation.

The branch is pushed only to codex/typeaware-wave-11, never main or area branches;
integration owns merging. Main was fetched again and remained e8ba3d5d. Commit SHA
is recorded in the final worker response. No PR is opened.
