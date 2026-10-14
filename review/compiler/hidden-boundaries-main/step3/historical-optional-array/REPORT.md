Built optional indexed array access and exact caller selection for hidden boundary 04, roadmap step 30.
Base: `dcdbb9098f77f30ad41790c56df1bd63ad462b63`, requested `origin/compiler/area-next-fixtures`; delivery branch: `codex/hidden-04-generic-optional-array`.
Focused Node/native/JavaScript oracle passes, including ASan, UBSan and leak checks; the original signature no longer reproduces, while the next cast does.
Caught mutants: undefined for every present input, eager index evaluation, wrong replay sibling, and a missing region unit.
Own revealed bytes: **0**; the next unsafe cast remains refused, and the second assigned region was not measured.

The delivery SHA is the commit containing this report and is supplied with the
feature-branch push. No main or area branch was modified, and no other worker's
unlanded feature was merged. The requested area tip remains the delivery base.

The historical census pin is `388096e6`. All **82** regular adapted compiler
files match its stock manifest byte lengths and SHA-256 hashes. The verified
existing tree is exposed at `/tmp/hidden-adapted`; [source-hashes.json](evidence/source-hashes.json)
and [provenance.json](evidence/provenance.json) record the inputs and baseline.
The redundant fresh adaptation and all-file census were stopped; their partial
outputs are not used as evidence or represented as complete censuses.

The diagnostic-only selector still selects `core.ts:22:1` and exits 1 without
the historical signature. The new explicit caller selector selects exactly
`checker.ts:10648:13`, `serializeMaybeAliasAssignment`, and also exits 1 because
the historical optional-array signature is absent. No matching rule was weakened:
the diagnostic position, kind and reason remain exact. The independent semantic
fixture is not presented as a reproduction of the historical census failure.
There is no failed boundary containing the recorded head byte `620304` in the
caller attempt. Logs and raw records are retained in `evidence/historical-*`.

The proposed fixture exposed a separate production stop on this area base:
`hidden_boundary_generic_optional_array.a:1:76: stage 0 can't lower ?.[] on a value yet`.
The new lowering saves the array receiver once and branches before evaluating
the index. A present array uses the existing indexed-read representation and
checks; an absent receiver gives undefined. Existing element representation,
unsupported-chain and mutable-generic checks are retained. No erased generic
storage or unchecked cast acceptance was added.

The supplied fixture now prints `7|0\n`, with empty stderr and exit 0, on source
Node, sanitized native code and the JavaScript backend. A separate fixture pins
single receiver evaluation, skipped index side effects, empty arrays, negative
and fractional indices, and string and boolean elements. Its output is:

```text
-1|1|0
8|2|1
-1|-1|-1
word|none
false|true
```

The unsafe generic mutation fixture remains refused by `adamic/invariant-mutable`:
writing `7` through generic mutable storage instantiated for strings would make
the written and read types disagree. Node prints `text|7`. A sound alternative
must use the parameter's element type for the inserted value or prohibit mutation.

The first remaining boundary **inside the assigned largest interval** is
`checker.ts:10667:199`, span `[620603,621236)`:

```text
Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)
```

This is the cast to
`ShorthandPropertyAssignment | PropertyAssignment | PropertyAccessExpression`
inside `getPropertyAssignmentAliasLikeExpression(aliasDecl as ...)`.
Explicit caller replay reproduces that exact Refused signature, exit 0; see
[next-cast.log.txt](evidence/next-cast.log.txt) and `next-cast.json.gz`.
The reduced [program](../../internal/oracle/testdata/hidden_boundary_alias_cast_refused.a)
returns a wide declaration with kind 3 and asserts a union whose only kinds are
1 and 2. Node prints `3`; unchecked acceptance would silently violate the asserted
type. The refusal is sound and necessary for this unchecked shape. A narrow sound
acceptance would prove an exhaustive discriminated source union, or introduce
a runtime checked union view that validates the allowed discriminants and each
required field. Recursive AST fields and callable contracts need their own proof;
checking only a tag is insufficient for a wide structural declaration. This unit
does not lift that refusal.

An earlier failure in the same caller, outside the assigned interval, comes from
`utilitiesPublic.ts:852:44`, the `__String` parameter to
`unescapeLeadingUnderscores`. Its failed declaration can create later `reading`
echoes. Those echoes are retained as observations, not treated as independent
semantic bugs or evidence that the entire caller compiles.

The region census keeps the complete project and attempts all four stock-AST
units overlapping `[620304,627479)`: `checker.ts:1486:1`, `6309:5`, `9346:9` and
`10648:13`. The existing hash-pinned `hidden.py` unions boundaries/checker skips
and subtracts independent attempted coverage. Only the intersection with the
assigned interval is reported. The same partial selection reproduces the original
full census's **7,175-byte** intersection exactly before measuring the current
compiler. This is a partial interval census on a checker-rejected program,
not a whole-file or whole-corpus measurement.

| Configuration | Hidden intersection bytes | Difference from historical pin |
|---|---:|---:|
| Historical census pin | 7,175 | 0 |
| Unchanged area base | 5,213 | 1,962 |
| Delivery compiler | 5,213 | 1,962 |

The area-base and delivery JSONL ledgers are byte-identical. Therefore the
historical-to-delivery reduction is **1,962 bytes**, all attributable to prior
area-base progress; this unit's additional reduction is **0 bytes**.
The new hidden intersection is `[620603,621669)`, `[622380,624252)`,
`[624411,624548)`, `[624985,625304)`, `[625583,626740)` and `[626763,627425)`.
[region.json](evidence/region.json) and [base-region.json](evidence/base-region.json)
retain arithmetic, hashes and roster validation. The 13,061-byte group total
was not used as a revealed-byte claim. The second interval and complete parser
or tsc execution were not measured.

Local commands, with output written to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/hidden-04-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/oracle -run '^TestHiddenBoundary04' -count=1 -v > /tmp/hidden-04-oracle.log 2>&1
go test ./internal/oracle -run '^TestHiddenBoundary04PresentArrayMutant$' -count=1 -v > /tmp/hidden-04-present-mutant.log 2>&1
go test ./internal/oracle -run '^TestHiddenBoundary04EagerIndexMutant$' -count=1 -v > /tmp/hidden-04-eager-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/hidden-04-counts.log 2>&1
python3 stage3/hidden-04/replay_check.py /tmp/hidden-04-region-worker /tmp/hidden-04-selector-final > /tmp/hidden-04-selector-final.log 2>&1
```

The guarded worker was built with the existing `make_overlay.py` tool. The
historical caller replay adds
`-attempt /tmp/hidden-adapted/src/compiler/checker.ts:10648:13` to the brief's
unchanged diagnostic selector. The next replay changes only `where`, `kind` and
`reason` to the exact cast signature above. Region invocation:

```sh
LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-04-region-worker \
  -project /tmp/hidden-adapted/src/compiler \
  -region-file /tmp/hidden-adapted/src/compiler/checker.ts \
  -region-start 620304 -region-end 627479 \
  > /tmp/hidden-04-region.jsonl 2> /tmp/hidden-04-region-run.log
python3 stage3/hidden-04/measure_region.py /tmp/hidden-04-source \
  /tmp/hidden-adapted/src/compiler /tmp/hidden-04-region.jsonl \
  stage3/hidden-04/evidence/region.json > /tmp/hidden-04-measure.log 2>&1
```

| Mutant | Intended catcher and observed result |
|---|---|
| Return undefined for every specialized `first` call | Both backends exit cleanly and print `0\|0`, disagreeing with Node's `7\|0`; independent mutant selector passes only when that disagreement is caught |
| Evaluate optional index outside the present branch | Both backends increment the absent-array index counter; Node side-effect pin catches the stdout difference |
| Select caller's sibling | Existing `LATENT_REPLAY_MUTANT_PARENT_SIBLING=1` causes positive caller-selection check to fail with exact attempt mismatch |
| Remove an overlapping region unit | Stock-AST roster assertion fails before reporting byte counts |

Wrong diagnostic reason, explicit wrong caller, and a selector merely inside
the caller also fail the exact reproduction assertion. The measurement worker
checks disabled production Load/Lower and nil-IR guards on every invocation.
No whole packages or full gate were run.

The new check also ran with `/tmp/hidden-04-final-census` as its third argument.
Its caller and partial-region findings equal the unfiltered full census exactly;
the unrelated sibling is excluded. [selector-full.log.txt](evidence/selector-full.log.txt)
records both passing assertions.

The existing `python3 stage3/census/latent/replay/replay_test.py` was run, exit 1
in setup, and is **not claimed passing**. It still expects string and number
conditions to be Refused. Actual findings contain only `var`; a separate replay
with the unchanged area-base worker also observes only `var`, proving this is
a pre-existing stale expectation rather than an optional-array regression.
[existing-replay-tests.log.txt](evidence/existing-replay-tests.log.txt) retains
the failure and `legacy-baseline.json`/`legacy-baseline.log.txt` retain the baseline
check. The new unfiltered-census equality checks do not depend on that expectation.

The counts refresh passes, exit 0 (`ok internal/oracle`, 336.788s). Every table
change is explained below; the negative fixtures have no counts rows.

| Changed row | Explanation |
|---|---|
| `hidden_boundary_generic_optional_array.a` added | New positive fixture: allocations/frees 4/4, retains/releases 4/8, peak 4, regions 0 |
| `hidden_boundary_optional_array_order.a` added | New positive fixture: allocations/frees 20/20, retains/releases 22/43, peak 8, regions 0 |
| `logical_and_reference_maybe.a` moved | Existing row's values remain 8/8/11/22/4/0; regeneration places it at its current fixture-registration position |
| `stage3/fixtures/taste/17_binder_flow.a` removed | The area base already registers this fixture with `lowers=false`; the stale counts row is omitted by the unchanged generator. The base refusal is retained, not lifted by this unit |

The complete counts output is [counts.log.txt](evidence/counts.log.txt).

Setup succeeded: Node ready 0.190s, Go ready 0.262s, markdown step 0.042s and ready
0.660s, submodules ready 0.696s, clang ready 1.683s, Go build ready 491.076s,
test binaries deferred 493.414s, cache warm 493.429s, total 494.237s.
`nproc` is 5, CPU quota 4; Node 24.19.0, Go 1.27.1, clang 20.1.8.
Logs retain the complete timing lines. Initial fixture-path and mutant-ABI test
errors were corrected before the passing focused run. A scratch replay with a
mistyped quoted reason correctly failed signature matching; it is not used as
reproduction evidence.
