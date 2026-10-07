# Slot 02 fourth helper batch

One helper per .a file, preserving the pinned Go implementation.

| File | Contract |
|---|---|
| valid_theme_prefix.a | isValidThemePrefix accepts a nonempty sequence of lowercase ASCII letters only. |
| namespace_for_variant_root.a | namespaceForVariantRoot returns --container iff the root begins with @; otherwise --breakpoint. |
| convert_underscores_to_whitespace.a | convertUnderscoresToWhitespace(input, skip) always removes the backslash immediately before an underscore. Bare underscores become spaces only when skip is false. Other backslashes are preserved. |

String inputs are valid Unicode text. Go’s malformed UTF-8 byte strings and isolated UTF-16 surrogates are outside the common string domain. The conversion splits chunks only at ASCII boundaries and preserves astral characters.

The independent Go oracle invokes the actual private helpers using temporary overlays; tracked cohere files stay unchanged. Its inputs include 157 captured runtime rule/file/source cases from every consumer, literals decoded by Go’s real TypeScript parser, supplementary string literals extracted from Go theme and framework-variant tests, actual theme prefixes and 202 successfully parsed framework variants. This produces 1,342 distinct texts, plus every ASCII byte pair and every Unicode code point for prefix/namespace probes (surrogate code points map to Go’s replacement rune). Expected output is removed before Adamic reads the corpus.

The test compares actual Go, Node source, emitted JavaScript and ASan/UBSan native byte for byte. Five compiling semantic mutants independently exercise empty prefixes, the z boundary, namespace position, the skip flag and escaped underscores. Native compilation/runtime failures are not credited as semantic mutant kills.

Regeneration captures all six consumers before external engine skips. Eight known Tailwind live-engine/corpus gates fail in this environment; capture is not represented as a passing rule gate. Missing consumers and unexpected failures reject regeneration. Artifacts reproduce byte for byte.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot02/batch4/testdata/regenerate.py > /tmp/slot02-batch4-capture.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch4$' -count=1 -v > /tmp/slot02-batch4-tests.log 2>&1
```

See REPORT.md for results and limits, RULES.md for consumers and readiness.json for residual dependencies. No shared registration, harness or compiler source is modified.
