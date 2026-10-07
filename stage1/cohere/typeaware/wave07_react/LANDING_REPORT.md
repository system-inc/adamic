Built: rebased wave-07 onto current main and revalidated its nine completed native ports; the three React claims remain unported.
Commits: tested 543efdc23d5c1d1adebea5496314636ddf108130 on main e8ba3d5d81de4d3773c723914fccd4c76248b965; original pushed tip was b9e43baf5abcf7dfc8ee8e3322682a426031847a.
Commands and outputs: all nine production-Go rule gates, sanitized comparisons, released-handle checks, bridge/decoder tests, vet, formatting, 38-file lint and 15 external Node fixtures pass.
Mutants: nine rule mutants, the native graph mutant, five question-fact mutants and the listed buffer, handle, decoder, refusal and Node-byte mutants were caught.
Not covered: three React implementations, four production bridge registrations, shared harness integration, emitted-JavaScript rule comparison, full repository gate and exhaustive options/JSX fixtures.

# Landing refresh

This worker has pushed one branch, codex/typeaware-wave-07. It was rebased without
conflicts onto origin/main at e011f8f6 and all existing gates passed. Main then
advanced with the devirtualization merge. The branch was rebased again without
conflicts onto e8ba3d5d; every completed port was rebuilt and its oracle gates
rerun. No rules were claimed during either pass. The remote main check during
the second pass still returned e8ba3d5d. Publication uses a lease against the
original pushed tip, preserving any unexpected intervening remote update.

The final proof commit adds reports/evidence and isolated verification scripts;
it changes no native rule, checker question or compiler implementation after
these gates. Protected compiler files match origin/main exactly. No shared
parser, registration generator, dispatcher or existing test harness was edited.
The second run and its exact process commands/exits are in landing-evidence/gate.
The earlier run is retained as a separate baseline, not used as the current-main
proof. Final publication SHA is reported after committing this evidence.

## Agreement

| Rule | Control findings | Fix edits | Suggestions |
| --- | ---: | ---: | ---: |
| no-undef-init | 26 | 11 | 0 |
| @typescript-eslint/prefer-for-of | 78 | 0 | 0 |
| @typescript-eslint/consistent-indexed-object-style | 69 | 58 | 7 |
| nexus/correctness-no-uncleared-race-timeout | 11 | 0 | 0 |
| nexus/correctness-no-process-exit-after-output | 16 | 0 | 0 |
| nexus/correctness-require-blocking-standard-streams | 16 | 0 | 0 |
| prefer-rest-params | 9 | 0 | 0 |
| prefer-promise-reject-errors | 22 | 0 | 0 |
| prefer-regex-literals | 41 | 0 | 30 |

The original trio agrees on 283 valid controls (47918 bytes), the frozen
287-source repository corpus (16 findings, 22753 bytes) and TypeScript's 77
compiler sources (24 findings, 9282 bytes). Each continuation rule agrees on
its positive/negative controls, DOM variants and both frozen corpora. Those six
rules have zero corpus findings (18485 repository and 5318 compiler bytes),
backed by their positive controls. Normal and ASan/UBSan/LSan runs match Go on
all controls and both corpora. All canonical findings, byte spans, messages,
fix edits and complete suggestions are compared unchanged. Go-managed heap
internals are outside C sanitizers' instrumentation.

Rest-parameters and promise rejection use the normal production archive.
The three correctness rules and regex use their previously documented private
registration overlays; these four production routes remain pending. A normal
archive still refuses the unregistered regex question with panic 70. These
results establish isolated implementations, not production registration.

## Mutants and guards

Every rule mutant compiles and exits 0 with empty stderr. Only the independent
production Go byte comparison catches it:

- no-undef-init: removal start advanced one byte.
- prefer-for-of: inverted body predicate.
- consistent-indexed-object-style: changed message punctuation.
- the three correctness rules: changed each rule's message ID.
- rest-parameters, promise rejection and regex literals: changed each message ID.

The native reference graph's anchor comparison is inverted, and the Go graph's
symbol identity is zeroed. Both exit normally and differ only in oracle bytes.
The continuation fact mutants zero symbol identity, drop flow edges and erase
module targets; the regex fact mutant drops character spans. All exit normally
and are caught only by Go bytes. Question-probe logs record the first mismatches
at 5390, 618, 9645 and 5854 respectively. Each live question succeeds, and each
post-release query panics 70. Registry-retention mutants in the original trio,
continuation and regex gates exit 0 and fail the required refusal assertion.

The full bridge package also catches input/output length errors with ASan,
retained stale handles with its stale-handle assertion, the wrong queried
source position with independent oracle bytes, removed link opt-in with its
refusal assertion, omitted output frees with LSan and heap allocation of an
unowned region result with LSan. The 1600-position bridge oracle agrees on
54982 bytes under sanitizers. The fact decoder catches these 19 wire mutants
with the expected panic: empty, length-negative, length-short, trailing,
version, question, not-integer, rounded-integer, noncanonical-integer,
negative-natural, bad-boolean, zero-id, negative-tuple, missing-root,
missing-constraint, missing-element, duplicate-id, missing-link-17 and
missing-link-18. Wrong-kind and unknown-question guard mutants exit 0 and fail
the required refusal checks. The external Node one-byte mutant is caught too.

The ordinary filtered Node run passes nine native fixtures; an expanded run
passes 15, including all six new devirtualization/direct-call fixtures in main.
Both include the one-byte mutant. Exact selections and oracle cache observations
are retained. Vendored regex tests include all table tests and fuzz seeds;
there was no timed fuzz campaign.

## Reproduction and environment

All test output goes to log files. After sourcing /workspace/adamic-tools/env.sh:

```sh
ADAMIC_WAVE07_LANDING=/workspace/wave-07-landing-current python3 stage1/cohere/typeaware/wave07_react/landing_gate.py > /workspace/wave-07-artifacts/landing-gates-current.log 2>&1
ADAMIC_WAVE07_LANDING=/workspace/wave-07-landing-current python3 stage1/cohere/typeaware/wave07_react/landing_bench.py > /workspace/wave-07-landing-current/bench.log 2>&1
ADAMIC_WAVE07_LANDING=/workspace/wave-07-landing-current python3 stage1/cohere/typeaware/wave07_react/landing_evidence.py > /workspace/wave-07-landing-current/evidence.log 2>&1
```

The runner now includes the expanded Node selection, which was separately run
and asserted during this validation. The existing shared harness is unchanged.
Source lint uses the previously documented isolated .a-capable cohere CLI,
--no-cache --no-fix, over 38 owned native files: 276 rules, zero findings.
Go vet ./... and gofmt checks exit 0 with empty output. Full go test ./... was
not run; the touched bridge packages, nine rule oracles, fact guards and filtered
external Node oracle were run. Setup compiled all packages/tests without running
them before the first pass.

Setup succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0: Go ready 0s,
clang ready 0s, Node ready 0s, submodules ready 0s, cache warm 114s, total 114s.
nproc is 5; cpu.max is 400000 100000; reported memory is 17.6 GB. The same
unchanged toolchain was reused after the second rebase. Oracle cohere remains
715ba94f3608a6500086b1076ce5cb7e51b836db, TypeScript-go remains
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the compiler corpus remains
050880ce59e30b356b686bd3144efe24f875ebc8.

## Fresh native versus Go timing

Three quiet alternating rounds preserve and compare complete outputs. Whole
process medians in seconds, including program loading and traversal:

| Suite | Repository native / Go (s) | Compiler native / Go (s) |
| --- | --- | --- |
| Original trio (combined) | 0.335376 / 0.145892 | 2.180712 / 0.324836 |
| No uncleared race timeout | 0.270944 / 0.144344 | 1.727619 / 0.318342 |
| No process exit after output | 0.281451 / 0.137734 | 1.798318 / 0.335571 |
| Require blocking standard streams | 0.310977 / 0.131662 | 1.744636 / 0.306632 |
| Prefer rest parameters | 0.262383 / 0.128556 | 2.032929 / 0.383173 |
| Prefer promise rejection errors | 0.284467 / 0.144905 | 1.733390 / 0.326889 |
| Prefer regex literals | 0.272352 / 0.136385 | 2.028055 / 0.331854 |

Native remains slower in every measured group. The original trio is timed as
one combined suite; the continuation rows time individual rules. These are
measurements on the frozen corpora, not positive-control-only timings. Raw phase
logs, all rounds and output hashes are retained in landing-evidence/bench.

## Remaining React and integration blockers

The current branch's shared native parser still exits 70 on all three React JSX
positive controls, while the independent Go oracle reports one finding for each
claimed rule. Ordinary TypeScript still parses. The published parser on
origin/codex/stage1-jsx-lint at a8a62d62 accepts all three controls in the isolated
dependency probe; it was retested using the current-main compiler. That parser
has not been integrated into this branch. Under Ahra's instruction to keep
changes in our rule directories and stop rather than edit shared files, this
unit leaves shared parser integration to its owner. The React rules remain
unported and require native value-flow implementations plus their complete
verification after integration. No React rule mutant, sanitizer result or
native/Go rule timing is claimed.

The four pending bridge routes and shared .a harness/profile/suggestion/
emitted-JavaScript integration remain documented in the earlier reports. No
additional rules were claimed. These reservations remain incomplete.
