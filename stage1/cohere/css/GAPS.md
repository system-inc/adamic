# CSS slice gaps

The raw and composed CSS/SCSS parsers and printer now compile natively after
merging the regex cycle proof `32f8106`. The remaining workaround and
original-library boundary proofs stay below.

## 1. Native regexp and recursive readonly trees: closed

The former `ir.RegExpCall` cycle-proof refusal is closed by `32f8106`.
[gaps/1_regex_and_value_tree.ts](gaps/1_regex_and_value_tree.ts) still prints
`Parsed` and `Ok`; `TestClosedParserRegexGap` holds it on native ASan/UBSan,
Node, the JavaScript backend and LeakSanitizer. `TestCompositionMatchesGo`
holds the full composed tree and error-position corpus on those same backends.
Readonly recursive ownership is unchanged; no fields were made Weak.

## 2. Array shift

[gaps/2_array_shift.ts](gaps/2_array_shift.ts) prints `a` and `1` on Node. Adamic
refuses `inherited library member shift read as an own field`: this unsupported
method must not become a load from a nonexistent own slot. The parser reads the
front token and removes it with `splice(0, 1)` instead.

## 3. Optional boolean conditions: closed

[gaps/3_optional_boolean_condition.ts](gaps/3_optional_boolean_condition.ts) prints
`important` on source Node, native ASan/UBSan and the JavaScript backend,
with a separate LeakSanitizer check in `TestClosedOptionalBooleanConditionGap`.
Closed by `320b762dfebed000ca07288e51eb6d78cd29b5f7`, which lowers optional
boolean conditions. Raw and composed tree serialization, printer truth predicates and namespace
printing now use the boolean table read directly. Comparisons used to pass a required boolean to
`setBoolean` remain: those normalize a value, rather than work around a condition.

## 4. Empty array assigned into an optional array

[gaps/4_empty_array_union.ts](gaps/4_empty_array_union.ts) prints `0`. Adamic reports
`stage 0 can't lower an array of never yet`. The custom-property composition
creates an explicitly typed `number[]` local before assigning its optional slot.

## 5. Dynamic repeat under catch

[gaps/5_repeat_in_try.ts](gaps/5_repeat_in_try.ts) uses a count from
`programArguments`, and prints `a` with no arguments. Adamic reports:

```text
stage 0 can't lower a try around repeat, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) yet
```

A literal repeat can be folded and does not prove this gap. The composition uses
an explicit loop to repeat spaces or replacement characters. The counts come
from byte lengths or clamped byte-slice boundaries.

## Go and original JavaScript disagree at a surrogate cut

[gaps/upstream_surrogate.mjs](gaps/upstream_surrogate.mjs) runs PostCSS 8.5.16 and
postcss-scss 4.0.9 on `\😀|a`. Both originals report:

```json
{"reason":"Unknown word \\\ud83d","endColumn":3,"endOffset":2}
```

That `endOffset` is the original UTF-16 offset. The oracle's conversion to bytes
turns it into 4, because a lone surrogate encodes as a replacement character.
Go cohere keeps the entire astral character on the left of `input.slice`'s cut:
its reason is `Unknown word \😀`, endColumn is 4, and endOffset is 5 bytes.
The Adamic raw parser follows Go. The upstream comparison admits only this
exact input and the exact two answer strings, separately for CSS and SCSS.
All other differences fail the test. The ordinary corpus contains two
occurrences, one in each dialect; 24,074 other answers agree byte for byte.

## Less scope

Go cohere's `internal/format/css/parser.go` explicitly omits `parseLess` and the
postcss-less branches. This slice likewise exposes CSS and SCSS, not a Less
entry point. Every one of the fork's 43 `.less` files is still tested through
both supported grammars and compared with those Go and JavaScript parsers.
Acceptance by a CSS grammar is not a claim of Less parsing. No postcss-less
oracle or Less composition is implemented here.

## 6. Printer native composition: closed

The former `ir.RegExpNew` cycle-proof refusal is closed by `32f8106`.
[gaps/6_printer_regex_tree.ts](gaps/6_printer_regex_tree.ts) still prints `2`
and `Ok`, held by `TestClosedPrinterRegexGap` on all backends and LeakSanitizer.
The complete printer corpus now runs native ASan/UBSan, source Node and the
JavaScript backend, with separate leak checks and three native output mutants.
See `NATIVE_REPORT.md` for observed results and native throughput.

## Printer boundaries against full Prettier

[gaps/printer_boundaries.ts](gaps/printer_boundaries.ts) and
[gaps/printer_library.mjs](gaps/printer_library.mjs) prove four full-API
boundaries. Go's internal CSS entry receives raw input, whereas Prettier's
full format entry preserves a BOM, normalizes carriage returns, bypasses
parsing for whitespace-only nonbreaking space, and delegates nonempty YAML
front matter to its YAML printer. Go CSS explicitly refuses that delegation.
The Adamic printer follows Go. These are entry-point differences, not newly
claimed Prettier defects.

`testdata/printer_library_gaps.json` records exact input, option mode, oracle
variant and both answers; all other format or acceptance differences fail.
For each option set each original oracle has 4,952 byte-identical formats,
19,080 shared refusals and 44 recorded discrepancy occurrences: BOM 12, CR 2,
nonbreaking space 2, YAML 28. Error messages from Prettier's public API include
code frames and are not compared with Go's internal error API. Go errors and
positions are held exactly. JSON escaping differences are decoded before
comparing formatted string bytes. No arbitrary invalid-UTF-8 output is promised.

## 7. Appending to a shared string slice: closed

The shared-slice append fix included in `50045bd` closed the capacity-underflow
bug. [testdata/shared_slice_append.ts](testdata/shared_slice_append.ts) moved out
of `gaps/` unchanged. `TestSharedSliceAppendAgreesWithNode` now requires `1152`
and exit 0 from Node, native ASan/UBSan and the JavaScript backend, with a
separate LeakSanitizer check.

The printer retains its earlier direct-concatenation workaround pending a new
benchmark. `PERFORMANCE.md` records the historical failure and rejected unsafe
candidate; its old `gaps/7_shared_slice_append.ts` path now refers to the moved
regression above.

Optional boolean closure validation on compiler/area-gaps:

```sh
go test -v ./stage1/cohere/css -run '^(TestClosedOptionalBooleanConditionGap|TestTheCanonicalRangeChecksCanFail|TestCompositionMatchesGo)$' -count=1 -timeout 30m
go test -v ./stage1/cohere/css -run '^TestOptionalBooleanPrinterMatchesGo$' -count=1 -timeout 30m
```

Both passed (180.627s and 60.278s). Composition compares 24,014 cases with Go,
Node and both backends; the tiny printer corpus compares six flag/namespace
cases, including important and SCSS default flags, with all those sides.
Sanitizer and leak checks pass. Existing custom-property, comment and closing
range mutants are caught by Node/Go comparison, and the public Range corruption
mutant is caught on raw and composed backends. Changing the gap's boolean to
false failed on normal Node `ordinary` output. Removing the declaration
semicolon failed the six-case printer test on normal Node output against Go.
Both temporary mutants were restored. The full printer/package corpus was
not run. The first test attempt, before setup completed, failed because the
TypeScript submodule's tsc/go.mod was absent; the logged successful rerun above
followed completed submodule setup.
