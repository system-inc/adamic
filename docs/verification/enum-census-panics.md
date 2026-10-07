Fixed both census panic paths by retaining only textual field origins; unsupported computed keys reach the existing named NotYet boundary.
Implementation commit: `7b585fcd757c812f4e2d9d5d4d0ec8acd81ea7ea`, following `06e8507356eee09111459c6d7e8a5c0aa952adbd`; report commit is the branch tip.
Verification: full lowering PASS 13.171s, full uncached oracle PASS 204.356s, reduced pins PASS 0.408s; complete main and area censuses each change from two panics to zero.
Mutants: removing the key-kind guard restores each panic in separate runs; discarding every exact field origin wrongly refuses both valid field controls.
Not covered: general computed enum object-key lowering or the outstanding checked tag payload views; the full repository gate was not run.

## Two records, one root cause

Both original records concern
`src/compiler/transformers/declarations/diagnostics.ts`. They are two paths to
one bug, not two independent causes. The original function is
`createGetIsolatedDeclarationErrors` at 606:1. Its diagnostic tables contain
computed enum keys followed by `satisfies Partial<Record<...>>`.

| Original record | Reduced program | Path to the failure | Independent pin |
| --- | --- | --- | --- |
| 1:1, refusal_scan | `internal/lower/testdata/census_panics/refusal_scan.a` | `refuseWidening` -> `provenRelation` -> `optionalRelationFailure` -> `relationFieldExpression` | `TestCensusComputedRelationPanicsStayNamed/refusal_scan/full_lower` |
| 606:1, lowering | `internal/lower/testdata/census_panics/lowering_unit.a` | `enumNeverValue` handles satisfies -> the same relation proof and helper | `TestCensusComputedRelationPanicsStayNamed/lowering_unit/expression_lowering` |

The refusal scan attributes its panic to the file; the unit attempt attributes
it to the containing function. A scratch trace confirmed both stacks on the
unmodified pinned compiler input. It restricted measurement to that file while
loading the full project, then printed recovered stacks. The trace instrumentation
exists only under `/tmp/enum-panic-overlay`, not in production or the census tool.
Log: `/tmp/enum-panic-trace.log`.

The top-level reduction is:

```typescript
enum Kind { A }
const table = { [Kind.A]: 'a' } satisfies Partial<Record<Kind, string>>;
```

The second puts that declaration inside `function make(): void`. Both pass the
source checker. Both independently run on Node with exit 0, empty stdout and
empty stderr. Before the fix, the ordinary compiler and the reduced regression
panic with `Unhandled case in Node.Text: *ast.ComputedPropertyName`.
The first owned failing frame is `relationFieldExpression` in
`internal/lower/proven_relations.go`, formerly line 172.

That helper locates an explicitly written field's initializer so the optional
relation proof can recognize an exact nested literal. It previously called
`property.Name().Text()` on every property assignment. A computed key is an
AST expression wrapped in ComputedPropertyName, and has no Node.Text value.
The assertion on AST shape was missing in Adamic; this is not a checker bug.

## Fix and named outcome

`relationFieldExpression` now compares only identifier and quoted string keys.
Computed keys supply no exact named field origin. Existing object lowering owns
the supported-key decision and returns `NotYet` for these enum-key reductions.
That is a stage-0 limit, not a claim that TypeScript or the language forbids them.
No catch-all panic recovery or permissive proof was added.

Pinned reduced outcomes, with full paths supplied by the harness:

```text
refusal_scan.a:2:17: stage 0 can't lower a computed field name yet
lowering_unit.a:3:19: stage 0 can't lower a computed field name yet
```

The actual CLI exits 1 with those diagnostics, empty stdout and no panic stack.
Both full program lowering and direct expression lowering are tested for each
reduction. The direct path bypasses the up-front scan, so a scan-only workaround
cannot hide the unit failure.

Identifier and quoted-key controls preserve an exact nested literal with a
missing optional field. They must still lower successfully. This holds the fix
to preserving the proof it is meant to support, rather than discarding it for
all object fields.

## Commands and observations

The existing toolchain was reused, with each Go/test shell sourcing
`/workspace/adamic-tools/env.sh`. The original setup timings and nproc 5 are
recorded in [the preceding unit report](enum-tag-final.md).
Every test run wrote output to a log; no test output was piped.

```sh
# Before the fix, FAIL with the owned computed-name panic:
go test ./internal/lower -run '^TestCensusComputedRelationPanicsStayNamed$' -count=1 > /tmp/enum-census-panics-red.log 2>&1
# Final reduced pins and valid-field controls:
go test ./internal/lower -run '^TestCensusComputedRelationPanicsStayNamed$|^TestComputedRelationFixPreservesExactFields$' -count=1 > /tmp/enum-panic-final-pins.log 2>&1
# Complete touched package and independent source/backend oracle:
go test ./internal/lower -count=1 -timeout 30m > /tmp/enum-panics-final-lower.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/enum-panics-final-oracle.log 2>&1
go vet ./... > /tmp/enum-panics-vet.log 2>&1
gofmt -l cmd internal > /tmp/enum-panics-format.log
git diff --check > /tmp/enum-panics-whitespace.log
```

Observed PASS: lower 13.171s, full uncached oracle 204.356s, final pins 0.408s.
Vet, formatting and whitespace checks pass with empty logs; final repeats are
`/tmp/enum-panics-final-{vet,format,whitespace}.log`. No backend or
runtime implementation changed, and no oracle counts rows were added: the new
fixtures are lowering refusals under `internal/lower/testdata`.
The successful oracle programs remain held to independent Node, release,
ASan/UBSan and leak checks through the existing harness.

CLI and Node observations, including their full diagnostics, are preserved in
`/tmp/enum-panic-reductions-results.json` and in the checked JSON below.
The pre-fix ordinary CLI traces are `/tmp/enum-panic-{top,function}-before.log`.

## Mutant evidence

There are two distinct mutations and three separate failing test invocations.
No Go or clang build failure is counted as a catch. Production files were not
modified by the overlays.

| Mutation | Invocation and catch |
| --- | --- |
| Restore unconditional key.Text | `/tmp/enum-panic-relation-mutant.json`, `TestCensusComputedRelationPanicsStayNamed/refusal_scan/full_lower`: panic in relationFieldExpression, logged in `/tmp/enum-panic-scan-mutant.log`. |
| Same mutation, independently exercise the unit path | Same overlay, `TestCensusComputedRelationPanicsStayNamed/lowering_unit/expression_lowering`: the same owned panic without a refusal scan, logged in `/tmp/enum-panic-unit-mutant.log`. |
| Return no exact field origins at all | `/tmp/enum-panic-no-origin-mutant.json`, `TestComputedRelationFixPreservesExactFields`: both identifier and quoted fields wrongly receive adamic/no-optional-widening refusals, logged in `/tmp/enum-panic-no-origin-mutant.log`. |

## Census evidence

[enum-census-panics.json](enum-census-panics.json) records compiler commits,
binary hash, unchanged input hashes, before/after totals, the two original
phase-attributed panic records and their named replacement boundary.
The same pinned adapted main 48c05d09 and area b2c4549f inputs were reused, with
all 79 source files validated by the meter's `latent_summary` on each input.
Both complete runs exit 0. Each reports zero panic sites, replacing the
original function failure with `NotYet: a computed field name` at 608:9.
Main totals: NotYet 1560 -> 1561, Refused 3735 -> 3745, panic 2 -> 0.
Area totals: NotYet 1558 -> 1559, Refused 3894 -> 3904, panic 2 -> 0.
The extra refusal findings are observations made reachable after the scan
stopped panicking, not new source edits or a claim that those sites compile.
The old enum-tag refusal row stays zero and the checked tag-view NotYet count
stays 90 on both inputs.

The latent binary cannot expose usable IR and has no backend output path.

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/enum-panics-fixed-overlay > /tmp/enum-panics-fixed-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/enum-panics-fixed-overlay/overlay.json -o /tmp/enum-panics-fixed-meter ./stage3/census/latent/tool > /tmp/enum-panics-fixed-meter-build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/enum-panics-fixed-meter /tmp/adamic-gate/stage3-meter.RLIiIU/main-adapted/src/compiler /tmp/enum-panics-final-meter/main/latent.jsonl > /tmp/enum-panics-final-meter/main/latent.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/enum-panics-fixed-meter /tmp/adamic-gate/stage3-meter.RLIiIU/area-adapted/src/compiler /tmp/enum-panics-final-meter/area/latent.jsonl > /tmp/enum-panics-final-meter/area/latent.log 2>&1
```

These are measurements on checker-rejected programs, not claims that the
TypeScript compiler now builds. The prior numeric-tag checked-view limits are
separate from the panic repair.

## Files and integration

Production: `internal/lower/proven_relations.go`.
Regression: `internal/lower/census_panics_test.go`.
Reduced `.a` programs: `internal/lower/testdata/census_panics/refusal_scan.a`
and `internal/lower/testdata/census_panics/lowering_unit.a`.
Records: `docs/verification/enum-census-panics.md`,
`docs/verification/enum-census-panics.json`; links added to `docs/enums.md`
and `docs/verification/enum-tag-final.md`.

The four prohibited files remain untouched. No cohere source was copied or
changed. Current origin/main is ce0750f2; fetching and merging it reported
Already up to date. Only `codex/enum-tag-narrowing` is pushed, with no PR.
