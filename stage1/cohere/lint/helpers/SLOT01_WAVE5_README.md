# Slot 01 fifth helper batch

One helper per `.a` file, pinned to cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`.

| File | API | Contract |
|---|---|---|
| `collapse_valid_named_value.a` | `collapseIsValidNamedValue(value)` | Nonempty ASCII a-z, A-Z, 0-9, underscore, dot, percent and hyphen only. |
| `collapse_valid_arbitrary.a` | `collapseIsValidArbitrary(input)` | Actual Go scanner: escape/quote skipping, matching bracket pops, top-level semicolon and closing rejection, permissive final return. |
| `collapse_parse_theme_options.a` | `collapseParseThemeOptions(params, trimmedSegments)` | Four option bits, ignored unknown words, last prefix wins including empty. |

Inputs are valid Unicode strings from the parser/text adapter. ASCII comparisons agree with Go byte comparisons for these strings; malformed raw UTF-8 and isolated UTF-16 surrogates are outside this adapter contract.

The arbitrary scanner follows actual Go code rather than strengthening its validation. It returns true for unfinished opening stacks and quoted strings. A mismatched closing bracket with a nonempty stack leaves that stack intact; opening braces do not push. A closing bracket on an empty stack rejects. A top-level semicolon rejects, while a semicolon inside parentheses or brackets is permitted. Escaped and quoted punctuation is skipped. These observations are independently captured from real Go.

The theme parser's explicit `trimmedSegments` callback must perform Go `strings.TrimSpace(params)` followed by Collapse `segment(trimmed, ' ')`. A plain JavaScript split or trim is not the callback contract. Native integration must supply the separately owned segmentation and Go-equivalent trimming. The helper passes the original params to the callback once, then handles `reference` (2), `inline` (1), `default` (4), `static` (8), and `prefix(...)`. Unknown words are ignored; prefix validity belongs to a separate helper. A later `prefix()` clears an earlier prefix.

Run from the repository root after sourcing `/workspace/adamic-tools/env.sh`:

```sh
go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave5' -count=1 -v -timeout=20m > /tmp/lint-helpers-01-wave5.log 2>&1
```

Slot-owned Go overlays expose the actual private helpers and actual trimming/segmentation. Consumer inputs are complete statically evaluable strings, preserving concatenations; labels and messages are included, dynamic construction is not evaluated. Go callback data is an explicit dependency seam, while independent Go parseThemeOptions supplies expected mask/prefix. Prefix output is observed as UTF-16 code units rather than lossy console text. All three compare actual Go to source Node, sanitized native and emitted JavaScript, with compiling semantic mutants.

Breakpoint bucketing was withdrawn after an earlier slot 04 claim was found. Its file and tests are removed from the delivered tree; its historical successful log is retained only as withdrawn evidence and is not counted.

See [SLOT01_WAVE5_REPORT.md](SLOT01_WAVE5_REPORT.md), `slot01_wave5_readiness.json` and `evidence/slot01-wave5/` for every consumer, remaining dependency and command. No shared registration generator or shared rule harness changes.
