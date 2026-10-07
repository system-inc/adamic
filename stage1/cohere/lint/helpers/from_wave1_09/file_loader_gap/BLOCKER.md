# Stylesheet loader input blocker

Actual Go stylesheetCollector.loadFile reads raw bytes with os.ReadFile, converts them directly to a Go string, parses CSS and ingests theme values. Adamic's only file-content API is readTextFile (internal/load/prelude.d.ts:13); oracle/adamic.mjs:56 explicitly documents replacement of invalid bytes with U+FFFD. utf8At/utf8Length re-encode already decoded text, so they cannot recover the lost bytes.

The oracle-only wrapper constructs the actual Go collector, calls its unchanged loadFile, and reads the installed --color-probe theme value. The witness is `@theme { --color-probe: BYTE; }`. It establishes that the difference is observable in an actual loaded design system, not merely an unused raw file byte.

| Value on disk | Actual Go loaded theme bytes | Source Node, emitted JS, sanitized native decoded bytes |
|---|---|---|
| A | 41 | 41 |
| é😀 | c3 a9 f0 9f 98 80 | c3 a9 f0 9f 98 80 |
| invalid byte 80 | 80 | ef bf bd |
| invalid byte 81 | 81 | ef bf bd |
| invalid byte ff | ff | ef bf bd |

The Adamic program is a fixed-witness input probe, not a ported CSS loader. It reads the file and extracts the single fixed value; it deliberately does not claim ParseCSS or ingest integration. All three execution paths compile and finish with exit 0 and empty stderr. Actual Go preserves each original invalid byte, while all three Adamic inputs are indistinguishable after decoding. A private helper cannot reconstruct information the shared input API erased.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_09 -run '^TestStylesheetLoaderByteInputGap$' -count=1 -v -timeout=10m > /tmp/lint-helpers-wave109-file-gap.log 2>&1
```

The gap-detection test passes in 2.198s because it verifies valid-text agreement and the required invalid-byte disagreements. It is not a green loader parity gate. Full output is ../evidence/file-loader-gap.log. It fails if the documented gap disappears, prompting replacement by a true loader comparison.

No loader implementation or prerequisite credit is retained. Reservation 2365c934 is withdrawn in the final report commit, and no further helper is claimed. A lossless file-byte input primitive and suitable byte-preserving CSS representation, or an approved restriction of Go's input contract, are needed. Shared runtime/library/compiler changes are outside this unit's permitted territory. This is the exact blocker for stopping; unclaimed helpers still exist.
