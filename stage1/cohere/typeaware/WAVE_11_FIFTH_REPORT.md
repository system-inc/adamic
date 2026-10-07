Built native .a ports of no-throw-literal, no-useless-backreference and prefer-arrow-callback.
Claim 5598cab4 was pushed before implementation 9e06b6bb; evidence is committed separately.
Agreement PASS 121.307s: 374 parse-clean controls, 259 findings, 216 repairs; both corpora and sanitizers passed.
Three rule mutants, three native judgment mutants, one raw-query mutant and one released-registry mutant were caught.
Not covered: two strict-parser-rejected inputs, full option matrix, full gate and emitted-JavaScript rule comparison.

## Reservation and scope

All twelve earlier reservations were complete and pushed through c4a9b256.
Fetch audited 347 origin refs, 341 distinct trees and 33 distinct Markdown claim
documents. Descending combined finding volume with lexical ties excluded base
ports and every named origin reservation, including skipped and released entries.
The next three were these zero-volume rules. Main and bridge name matches were
inventories and skipped-type catalogs, not native ports. fifth-selection.json and
its reconstruction script record the preclaim audit.

Production Go is unchanged at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db.
The compiler corpus is TypeScript v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8. Each native rule is in its own .a file;
new support, runner and raw-query consumer files are also .a. No compiler
implementation, shared generator, shared test harness or submodule was edited.

NoThrowLiteral preserves couldBeError's syntactic recursion, operator splits,
whole-throw range, and the special bare-global undefined diagnostic. Parentheses
are deliberately not made transparent where Go does not unwrap them.
NoUselessBackreference performs regex enclosure/path scanning, all five problem
classifications, duplicate named-group reachability, limited Go pattern validation,
constant argument folding and flow-insensitive global RegExp reference tracking
in native Adamic. Writes to a global stop that global's trace, and shorthand
references use the value symbol. Findings retain duplicates and production text.
PreferArrowCallback tracks function ownership of this/super/new.target/arguments,
resolved self-references, callback position and first bind semantics. It preserves
production repair ordering, comment preservation, refusal cases and parentheses.
Both runners use production defaults: named callbacks are allowed to report,
while unbound this-using callbacks are exempt.

## New raw bridge question

symbol_declaration_provenance.go returns each declaration's filename and
IsDeclarationFile flag for a borrowed, program-scoped symbol ID. It accepts zero
as a missing symbol and rejects malformed or out-of-range IDs. It does not choose
lint messages, ranges, repairs, reference paths or shadowing outcomes.
symbol_declaration_provenance.a decodes its frames; native code chooses whether
any declaration is in source. There is one dispatch entry in facts.go (two lines
after gofmt), with no other shared registration edits.

This was needed for the value symbol of a shorthand property. Existing lineage
queries accept a TYPE identity, not a symbol identity, and direct declaration
inspection cannot load the bundled:/ library paths through the path API. Initial
attempts exposed both errors; the isolated raw symbol query resolves the gap.
Direct tests compare source, ambient .d.ts, standard-library shorthand and missing
symbols against the checker itself, plus malformed identity cases. The query's
provenance mutant is caught by the independent Go findings comparator.

## Complete-byte agreement

The independent Go loader and production-rule walk import no bridge code. Every
serialized range, rule/message identity, message byte, fix range/text, suggestion
field and duplicate finding is compared. Sorting is deterministic on both sides;
no finding or fix normalization is applied.

| Population | Roots | Findings | Repairs | Identical bytes | Normal / ASan / UBSan / LSan |
| --- | ---: | ---: | ---: | ---: | --- |
| Parse-clean controls | 374 | 259 | 216 | 95808 | PASS |
| Frozen repository corpus | 287 | 0 | 0 | 18485 | PASS |
| TypeScript src/compiler | 77 | 0 | 0 | 5318 | PASS |

Findings: no-throw-literal 29, no-useless-backreference 127,
prefer-arrow-callback 103. Of the callback findings, 87 carry repairs and 16
withhold them. None of these Go rules offers suggestions; the empty suggestion
fields are compared as well.

There are 20 hand-written controls and 356 literal sources extracted from pinned
upstream-derived Go fixture tables. Expectations are never copied: unchanged
production Go computes all expected bytes. Both runners use default options;
upstream custom settings and configuration environments are not replayed, and
computed/joined fixture source expressions are not extracted.

Two sources, control-150.ts and control-172.ts, use legacy octal escapes rejected
by the strict TypeScript parser: a bare string '\1(a)' and RegExp('\1(a)').
The independent oracle explicitly selects parse-clean roots before comparison
and logs both exclusions. All 376 inputs and their hashes are retained, alongside
both full and parse-clean manifests. Large-corpus roots are never filtered.
Zero findings on those populations alone are not proof of fidelity; positive
controls and mutants exercise the report and repair paths.

## Mutants

All seven diagnostic/raw-query mutants compile, exit zero and emit empty stderr.
Only the independent complete-byte comparison catches their changed answers.

| Mutant | First differing byte |
| --- | ---: |
| no-throw-literal report end +1 | 717 |
| no-useless-backreference report end +1 | 116 |
| prefer-arrow-callback arrow insertion text changed to ` ->` | 10680 |
| Native source/declaration-file provenance inverted | 112 |
| Regex nested problem changed to forward | 4568 |
| Self-reference symbol equality replaced with spelling equality | 10396 |
| Raw symbol provenance query flag inverted | 112 |

A separate scratch registry mutant retains released programs. The real query
panics with exit 70 and `invalid or released checker handle`; the mutant exits
zero and is caught by the required panic check. It exercises the new question.
The production registry is unchanged. Full bridge tests also reprove C-output
free and region ownership mutants through LeakSanitizer.

Initial implementation checks exposed numeric short-circuit and empty-array
inference restrictions, a constructor guard affecting nested helper construction,
a shared fallback array causing duplicate traces, and the symbol/type identity
and bundled-path API mistakes above. Rule-local ternaries, explicit arrays,
constructor dependency injection, separate reference arrays and the isolated raw
query corrected them. These development failures are not counted as mutants.

## Quiet native time against Go

Three alternating rounds ran after all tests finished, without competing builds.
Every timed complete output was checked. Process medians include startup,
loading, rule work, output and teardown. Raw phase timing/query counts and the
benchmark script are retained.

| Population | Go median | Native median | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 131.258 ms | 272.962 ms | 2.080x |
| Compiler | 354.120 ms | 2293.976 ms | 6.478x |

Native is slower in both measured populations. Instrumented validation observed
25 repository checker queries and 460 compiler queries. These worker timings
are observations, not a general speed prediction.

## Commands and logs

Toolchain setup was reused: cloud/setup.sh ready 0s, cache warm 86s, total 86s;
nproc 5. Commands source /workspace/adamic-tools/env.sh. Test output is redirected
to log files, never piped.

```
ADAMIC_WAVE_11_FIFTH_ARTIFACTS=/workspace/wave-11-fifth-validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11FifthAgreementAndMutants$'
# PASS 121.307s; fifth-agreement.log

go test -v -count=1 -timeout 10m ./bridge/tsgo/checker ./bridge/tsgo
# PASS checker 0.218s, bridge 67.350s; fifth-bridge.log
# TestSymbolDeclarationProvenance PASS 0.02s

go vet ./...
# PASS, empty fifth-vet.log

gofmt -l bridge/tsgo/checker/symbol_declaration_provenance.go bridge/tsgo/checker/symbol_declaration_provenance_test.go bridge/tsgo/checker/facts.go stage1/cohere/typeaware/wave_11_fifth_test.go stage1/cohere/typeaware/testdata/oracle_wave_11_fifth.go
# Empty fifth-gofmt.log

git diff --check
# Empty fifth-diff-check.log

python3 /workspace/wave-11-logs/fifth_bench.py
# All complete timed outputs agree; fifth-timing.log and fifth-timing.json
```

validation-wave-11-fifth retains full compressed outputs, hashes, manifests,
compressed control sources, audit/claim evidence, passing logs and quiet timing.
Archives and executables remain in scratch. No PR was opened.

## Limits

The full repository gate and custom-option/profile matrix were not run. Two
strict-parser-rejected controls are explicitly excluded, not treated as passing.
The pinned cohere CLI still cannot resolve .a for the own-lint/format gate;
the native compiler builds these .a files successfully. Shared .a/profile and
emitted-JavaScript comparison integration remains on codex/lint-harness-dot-a
and was not edited or merged here. Applied-fix JavaScript execution and
emitted-JavaScript comparison of the rules are not covered; complete fix bytes
are covered by unchanged production Go. Earlier filtered Node oracle runs are
prior evidence, not rerun checks for this batch. Native production-default parity
has no unresolved blocker in this batch.
