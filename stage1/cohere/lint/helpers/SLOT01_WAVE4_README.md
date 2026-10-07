# Slot 01 fourth helper batch

One Go helper per `.a` file, against cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`.

| File | API | Contract |
|---|---|---|
| `jsx_string_attribute_value.a` | `jsxStringAttributeValue(nodes, attributes, wanted, attributeName, matches, unescape)` | Ordered JSX attributes; first matching attribute decides. Only a StringLiteral initializer yields a value, including the empty string. |
| `collapse_is_hex_digit.a` | `collapseIsHexDigit(character)` | Exactly ASCII 0-9, a-f and A-F. Input must be an integer byte, otherwise panic. |
| `collapse_node_is_container.a` | `collapseNodeIsContainer(node)` | Exactly rule, at-rule, context and at-root kinds; child-list presence is irrelevant. Non-null typed receiver. |

The JSX adapter preserves parser node identity with arena indices. Negative attribute/initializer references represent nil; nil property lists become empty arrays. Named fields are `kind`, `text`, `initializer` and `properties`. Required kinds are JsxAttributes, JsxAttribute and StringLiteral; other kinds may keep their Go names or Other. Invalid non-null indices panic. Spreads and unnamed/namespaced attributes are skipped. A matching missing initializer or expression initializer stops the search, so later duplicate attributes cannot supply a string. This is the actual Go behavior.

The three JSX callbacks are explicit dependencies belonging to other workers: AttributeName over the same arena, the selected Go-compatible name matcher, and UnescapeStringLiteralText. The oracle passes actual Go callback observations and independently compares the actual Go StringAttributeValue result. Exact matching and ignoring-case matching are both exercised. This does not claim integration with another slot's adapter or decoder.

The hex predicate accepts all 256 byte values, including bytes above ASCII, but cannot silently accept a number outside the Go parameter's domain. The container predicate takes a non-null node because Go's nil receiver panics; it does not inspect children.

Run from the repository root after sourcing `/workspace/adamic-tools/env.sh`:

```sh
go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave4' -count=1 -v -timeout=20m > /tmp/lint-helpers-01-wave4.log 2>&1
```

The slot-owned oracle captures statically evaluable complete string expressions from every consuming rule fixture, including constant concatenations and labels/messages, plus shape controls. The Collapse bridge exposes the actual private Go byte predicate only in a Go build overlay. Actual Go CSS parsing supplies container-node kinds; parse failures are recorded separately. Source Node, sanitized native and emitted JavaScript all compare to actual Go. Five semantic mutants must compile, exit successfully and differ in stdout; the sixth mutant removes the byte-domain refusal and must compile and successfully accept an otherwise refused input. Compiler failures, sanitizer errors and unrelated runtime failures do not count.

See [SLOT01_WAVE4_REPORT.md](SLOT01_WAVE4_REPORT.md), `slot01_wave4_readiness.json` and `evidence/slot01-wave4/` for every consumer, residual blocker and retained evidence. No shared rule harness or registration generator changes are included.
