# CSS slice gaps

The raw CSS and SCSS parsers compile natively. The composed parser matches Go on
Node but does not compile natively. These are separate observations, held by
separate tests. No compiler source or existing slice was changed.

## 1. Native regexp and recursive readonly trees cannot currently compose

[gaps/1_regex_and_value_tree.ts](gaps/1_regex_and_value_tree.ts) imports the existing
selector and value slices and parses `.a` and `red`. Node prints:

```text
Parsed
Ok
```

Adamic refuses `ValueTree.nodes` with:

```text
ir.RegExpCall is a node the cycle finder doesn't know
```

The complete `compose_main.ts` build reaches the same refusal, on
`MediaNode.nodes`, naming the write in `Parser_raw`. The cycle proof falls back
to refusing recursive readonly child storage when it encounters an IR operation
it does not recognize. Both recursive slices compile independently; the
combined native program is refused before clang, rather than compiled incorrectly.

`TestEachGapStandsWhereGapsMdSaysItDoes` holds the small program and its exact
refusal fragment. `TestNativeCompositionIsBlockedByTheRecordedGap` holds the
complete program too. Those tests fail when the compiler gap closes, requiring
native, JavaScript backend, sanitizer and leak agreement to be enabled for
composition. Merely changing the recursive child fields to Weak would change
ownership; this unit does not weaken them or edit the compiler.

## 2. Array shift

[gaps/2_array_shift.ts](gaps/2_array_shift.ts) prints `a` and `1` on Node. Adamic
reports `stage 0 can't lower .shift on a value yet`. The parser reads the front
token and removes it with `splice(0, 1)` instead.

## 3. Optional boolean conditions

[gaps/3_optional_boolean_condition.ts](gaps/3_optional_boolean_condition.ts) prints
`important`. Adamic refuses `a boolean | undefined as a condition`. Reads of the
boolean property tables compare the result explicitly with `true`.

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

## 6. Printer native composition

[gaps/6_printer_regex_tree.ts](gaps/6_printer_regex_tree.ts) constructs a regular
expression and imports the media tree. Node prints `2` and `Ok`. Lowering
refuses `ir.RegExpNew is a node the cycle finder doesn't know`. The complete
`print_main.ts` reaches the same refusal.
`TestNativeCSSPrinterHasTheRecordedCompositionGap` holds both proving programs
and fails when the gap closes. The printer is type checked by Adamic and runs
on Node; native printer output, sanitizer/leak checks and throughput remain
blocked. This cycle-proof work belongs to the other worker.

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
