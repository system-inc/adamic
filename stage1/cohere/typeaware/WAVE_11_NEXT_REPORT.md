Built: native ports of no-collection-misuse, no-discarded-outcome and no-discarded-pure-result.
Commits: original wave pushed through `fca4b6e5`; continuation claim `f2f13d2d` pushed before code; implementation `5ff9e4e6`.
Checks: 67 control findings and both frozen corpora match Go byte for byte, normally and under ASan/UBSan/LeakSanitizer.
Mutants: three rule ranges, two raw-question answers and the released registry each rejected by the intended check.
Not covered: full repository gate, exhaustive upstream fixtures/options, JSX and new-rule emitted-JavaScript comparison; pinned cohere `.a` lint support remains external work.

## Authorization, selection and scope

The original three wave-11 rules were complete and pushed before continuation.
After fetching origin, the claim audit covered 320 refs, 315 distinct trees and
30 claim files. Combined descending compiler/repository volume from the tables
linked by VOLUME_REPORT.md, with lexical ties, excluded the 26 documented base
ports and every name appearing in an origin claim file. Released reservations
and the skipped accessor-pairs entry were left untouched because their names
remain in other workers' claim files. No port of the chosen three exists on
origin/main (`ef3d907ecdc4c771b016f7d9c52372def057a340`) or the current bridge
branch (`5afbdb83da2ed7ad9815657cd3f6ececd5294bf6`).
`validation-wave-11-next/continuation-selection.json` records the ranking and
claim evidence. The first remaining names are the three in this report, all
with zero recorded corpus findings. Claim `f2f13d2d` was pushed successfully
before creating any implementation source.

Ahra's correction arrived after that continuation claim. This batch is now
complete; no further rules were claimed. The original shared harness and
registration generator were not edited. Each new rule, question helper, runner
and test is a separate file. The only shared bridge change is two switch
registrations in facts.go, four lines after gofmt, written before the correction.
No protected emitter, lowerer, native driver or oracle implementation changed.
The branch remains codex/typeaware-wave-11. No pull request was opened.

## Implementation and independent oracle

All new Adamic sources use `.a`. The raw declaration-lineage question exposes
symbol declarations, ancestors and AST identities, child links, source provenance
and a type alias's declared-type link. The awaited-type-shape question exposes
the checker's awaited type graph. Go contains no selection of outcome aliases,
pure methods, collection misuse cases, messages or finding ranges.

Adamic classifies library collection symbols, array/string/typed-array size
comparisons, canonical array indexes and keys, pure methods and callable
arguments. It groups outcome literal declarations by their union, requires every
arm, verifies the exact top-level alias and Nexus file suffix, and chooses the
first production alias name when multiple complete outcomes occur. Findings
use the original production messages and spans. These three rules offer no
fixes or suggestions; both serializations still compare those complete fields.

The independent Go oracle invokes the unchanged production registry rules from
cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. It has a separate loader and
AST traversal and imports no bridge code. Neither the submodule pin nor the
TypeScript-go pin changed. Source and message provenance is in
WAVE_11_NEXT_NOTICE.md.

| Population | Collection misuse | Discarded outcome | Discarded pure result | Equal bytes |
| --- | ---: | ---: | ---: | ---: |
| 28 control roots plus four imported declaration fixtures | 40 | 10 | 17 | 29046 |
| Frozen repository, 287 roots | 0 | 0 | 0 | 18485 |
| TypeScript compiler, 77 roots | 0 | 0 | 0 | 5318 |

TypeScript is v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
The repository and compiler manifests are unchanged from validation-coverage;
new ports did not pollute selection or validation populations. Both normal
and sanitized native outputs exactly match Go for every population. Committed
compressed complete outputs, hashes, source hashes, manifests and logs are in
validation-wave-11-next. Output headings contain absolute workspace paths.

Controls cover all four collection messages (7 membership, 21 size comparisons,
8 bracket accesses and 4 Object methods), mirrored operators, zero, negatives,
hexadecimal, separators, exponent literals, readonly types, constraints, unions,
subclasses, intersections, wide keys and shadowed Object/Map declarations.
Pure-result controls cover library versus project declarations, nullable calls,
parentheses, result use, explicit void, callback methods/arguments and callable
unions. Outcome controls cover all five aliases, imported and instantiated arms,
parenthesized unions, await, optional calls, extra arms, incomplete narrowed
outcomes, lookalike local declarations, precedence and Unicode/CRLF spans.

## Mutants and lifetime evidence

Every final rule/question mutant compiled, exited 0 and emitted empty stderr.
Only complete independent Go diagnostic bytes caught the difference:

| Mutation | First differing byte |
| --- | ---: |
| Collection finding end +1 | 61 |
| Outcome finding end +1 | 24373 |
| Pure-result finding end +1 | 14605 |
| Declaration provenance always non-library | 13243 |
| Awaited graph replaced by raw Promise graph | 25268 |

An earlier provenance-always-library mutant survived because the separate
source-declaration-file condition still excluded project source declarations.
It is not counted as a successful kill. A first numeric-separator control used
invalid `0_0`; the Go oracle rejected its parse diagnostics, and the control was
corrected to valid `1_0`. Stage 0 refused Number conversion and optional boolean
reads in initial implementations; supported parseFloat and explicit presence
checks replaced them before byte agreement was measured.

A released program queried with declaration-lineage exits 70 with exactly
`adamic: panic: invalid or released checker handle`. Retaining the released
registry entry in a scratch Go overlay makes the same probe exit 0; the required
panic check rejects that mutant. Direct checker tests verify awaited type
identity, declaration/ancestor kinds and identities, children, declared-type
links, library provenance, and malformed/unknown/noncanonical query refusals.

The full bridge package again passes C input/output off-by-one sanitizer
mutants, missing output free and region lifetime leak mutants, wrong-position
byte comparison, link refusal and stale/nonreused handle checks. ASan does not
instrument the Go heap. Native strings and returned C buffers are covered by
the sanitizer and ownership probes.

## Commands and observed output

Toolchain reused the successful setup from this same unit: ready 0s, cache warm
86s, done 86s; nproc 5, four-core CPU quota. All test output went to files.
Commands run from the repository after sourcing `/workspace/adamic-tools/env.sh`:

```sh
ADAMIC_WAVE_11_NEXT_ARTIFACTS=/workspace/wave-11-next-validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11NextAgreementAndMutants$' > /workspace/wave-11-logs/continuation-agreement.log 2>&1
go test -v -count=1 ./bridge/tsgo/checker > /workspace/wave-11-logs/continuation-checker.log 2>&1
go test -v -count=1 -timeout 15m ./bridge/tsgo/... > /workspace/wave-11-logs/continuation-bridge.log 2>&1
go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(sorting|string_index|functions|closures)\.a$' > /workspace/wave-11-logs/continuation-node.log 2>&1
go vet ./... > /workspace/wave-11-logs/continuation-vet.log 2>&1
gofmt -l bridge/tsgo/checker stage1/cohere/typeaware/wave_11_next_test.go stage1/cohere/typeaware/testdata/oracle_wave_11_next.go > /workspace/wave-11-logs/continuation-gofmt.log 2>&1
git diff --check > /workspace/wave-11-logs/continuation-diff-check.log 2>&1
python3 /workspace/wave-11-logs/continuation_bench.py > /workspace/wave-11-logs/continuation-timing.log 2>&1
```

Agreement PASS 258.519s. Direct checker package PASS 0.096s. Full bridge package
PASS 60.246s, checker package in that run PASS 0.107s. Filtered Node oracle PASS
13.250s, including generic_functions and method_closures selected by Go's
unanchored subtest regex and the independent one-byte mutant. Vet, gofmt,
whitespace and staged whitespace checks produced empty logs.

## Time against Go and limits

Three quiet alternating rounds per corpus load and run the full three-rule
suite and emit complete diagnostic streams. Every round checks identical bytes
against the independent Go truth. No build or test runs during these rounds.
Process medians, including loading and teardown:

| Corpus | Go | Native | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 0.144859 s | 0.351257 s | 2.42 |
| Compiler | 0.872701 s | 38.599321 s | 44.23 |

Detailed load/run/query phases and all round hashes are in
continuation-timing.json. The initial instrumented compiler run made 91,563
bridge queries, with 27.956s summed adapter time inside a 37.550s run;
the repository made 7,766 queries with 0.121s adapter time inside 0.298s.
Those are observations, not measurements of bare cgo crossing or a causal
attribution of the whole native cost. Native is slower; no optimization or
Go-speed parity is claimed.

The full repository test gate, exhaustive upstream fixture matrices, JSX and
new-rule emitted-JavaScript comparison were not run. Existing filtered Node
fixtures check both native and JavaScript backends, but do not establish a
JavaScript comparison for these new rules. The pinned cohere CLI's `.a`
selection/resolution/formatting limitation was measured in WAVE_11_REPORT.md;
this batch did not edit the shared harness or generator to address it. Ahra
assigned `.a` module support and emitted-JavaScript comparison to
codex/lint-harness-dot-a. Native compilation, complete Go bytes, positive
controls, per-rule/question mutants, released handles and sanitizers are the
passing gates here; corpus equality does not prove all-project equivalence.
