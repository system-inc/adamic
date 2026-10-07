# Slot 01 third batch

One helper per `.a` file, against Go cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. No shared rule harness or registration generator was edited. New tests, runner and Go overlays are slot-owned files.

| File | API | Contract |
|---|---|---|
| tailwind_attribute_values.a | tailwindAttributeValues(nodes, node, attributeNames, classValuesUnder) | Exact name.Text membership by true map value; preserve Attribute-origin initializer delegation, including nil. |
| text_decode_entity.a | textDecodeEntity(item, entities, hexValue) | Nonempty entity body; Go decimal/hex parsing, numeric clamp, surrogate replacement, exact XHTML named lookup, and literal unknown/oversized references. |
| react_likely_component_name.a | reactLikelyComponentName(name) | Nonempty first decoded rune is Unicode category Lu; remainder of the name is irrelevant. |

The attribute arena exposes kind/text/name/initializer, -1 for absent fields. Valid attribute kinds are JsxAttribute; names whose Go Text accessor would assert must be labeled UnsupportedText. Nil/wrong-kind attributes and unsupported name accessors explicitly panic, matching Go's private-helper preconditions. A missing name returns empty lists, while a matching attribute delegates its initializer even when nil. ReadonlyMap entries with false values do not opt in. The classValuesUnder callback belongs to another slot and must preserve the same node identity/origin and complete literal/template payloads.

The entity decoder returns `{replacement, ok}`: malformed bodies return empty replacement/false; unknown alphanumeric names return their literal `&name;`/true; overflow numerics preserve the original reference/true. Lowercase x alone introduces hexadecimal parsing; uppercase X does not. Valid numeric surrogate values become U+FFFD as Go string(rune(...)) does. The empty body panics. The hexValue callback receives decoded Unicode code points, not UTF-16 halves, and must match Go's separately owned helper. No browser or JavaScript HTML-entity implementation replaces Go behavior.

text_xhtml_entities.a contains all 253 named entries generated from pinned cohere's xhtml_entities_generated.go, typescript-estree 8.65.0. Pass textXhtmlEntities to the decoder. This data file defines no additional helper. The tests enumerate every real Go table entry and case variants independently of this ported table.

react_unicode_upper_ranges.a contains all 152 sorted Go unicode.Upper R16/R32 range triples for Unicode 17.0.0, generated under Go 1.27.1. The predicate uses binary search and range stride, preserving titlecase/non-letter exclusions and astral uppercase letters. Regenerate the external data with:

```sh
source /workspace/adamic-tools/env.sh
go run stage1/cohere/lint/helpers/testdata/slot01_upper_ranges.go > /tmp/slot01-upper-ranges.json
```

Each JSON ranges row becomes a readonly `[lo, hi, stride]` in the `.a` data file; version becomes reactUpperUnicodeVersion. A Unicode-table toolchain change must be accompanied by regeneration and the exhaustive actual-Go comparison.

```sh
go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/lint-helpers-01-wave3.log 2>&1
```

The slot-only oracle reads every listed consumer's Go fixture file and statically evaluates complete string expressions, including constant concatenations. Labels/messages are extra inputs; dynamically constructed fixture cases and whole-rule harness execution are not claimed. The attribute capture visits actual JSX attributes with three settings maps, including false entries, empty configuration, custom/case-sensitive names, nil initializers and nested value shapes. Actual Go classValuesUnder provides complete delegate payloads, and the actual reader supplies expected results. Entity capture tests every extracted nonempty string as a body, references found inside each source, all named entries/case variants, numeric samples across the Unicode range, every surrogate numeric value and malformed/overflow controls. Component capture includes every extracted fixture string, every parsed identifier, and every valid Unicode scalar value with a suffix.

Actual Go observations must match Node source, sanitized native and emitted JavaScript running on Node. The seven mutants compile, exit successfully with empty stderr and then differ from Go; panics, compiler refusals or sanitizer errors never count as a semantic catch. Entity observations include UTF-16 code units, because console printing alone hides lone-surrogate versus U+FFFD differences. Exact consumer and residual dependency lists are in slot01_wave3_readiness.json; commands and limits are in SLOT01_WAVE3_REPORT.md.
