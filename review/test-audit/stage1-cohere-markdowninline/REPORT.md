Unit u131: origin/main d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb; nproc 5.
2074 top-level tests form seven rows; 58 raw tests ran in the bounded matrix.
Bounded verdicts: two sacred rows and five setup-check rows.
Four production mutants: three caught, M3 survived; both empty-entry probes caught.
Evidence: review/test-audit/stage1-cohere-markdowninline/ on the requested audit branch.

```json
[
  {
    "test": "TestDelimiterExpressionMatchesNode",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/gaps_test.go",
    "seconds": 0.272,
    "oracle": "Node executes the protected delimiter expression; native bytes must match. A self-written Node output check also requires a\\*b plus newline.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [
      "M4"
    ],
    "last_proven_fail": "M4: gaps_test.go:31: native delimiter expression first byte difference at 0 (lengths 5/5)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P2"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/u131/cache/M4 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > M4.log 2>&1; gaps_test.go:31: native delimiter expression first byte difference at 0 (lengths 5/5)",
    "raw_failing_line": "gaps_test.go:31: native delimiter expression first byte difference at 0 (lengths 5/5)",
    "members_file": "family-members.json",
    "timing_samples": [
      0.292,
      0.268,
      0.272
    ]
  },
  {
    "test": "TestProduct_MarkdownInline family",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/inline_build_products_test.go",
    "seconds": 0.292,
    "oracle": "Self-written product recipes require successful construction of Go, Node, lowered C, native and sanitized products; this is a build/setup check, not a formatting comparison.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S6: inline_products_shards_test.go:318: open /tmp/u131/cache/S6/fba1196956e34be383cd865709e632e206d2ce9fb48abc4a1b0fc72254f7aa8f/program.c: no such file or directory",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/u131/cache/S6 timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u131/evidence/S6.overlay.json ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > S6.log 2>&1; inline_products_shards_test.go:318: open /tmp/u131/cache/S6/fba1196956e34be383cd865709e632e206d2ce9fb48abc4a1b0fc72254f7aa8f/program.c: no such file or directory",
    "raw_failing_line": "inline_products_shards_test.go:315: open /tmp/u131/cache/S6/fba1196956e34be383cd865709e632e206d2ce9fb48abc4a1b0fc72254f7aa8f/program.c: no such file or directory",
    "members_file": "family-members.json",
    "timing_samples": [
      0.276,
      0.292,
      0.307
    ],
    "construction_kills": [
      "S6"
    ]
  },
  {
    "test": "TestMarkdownInline family",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/inline_units_test.go; inline_products_shards_test.go",
    "seconds": 7.52,
    "oracle": "Go cohere overlay, source Node, emitted JavaScript Node, and pinned Prettier 3.9.6 compare bytes. Self-written corpus coverage and planted-disagreement labels also apply. Under TS mutants Node shares the altered source; Go and Prettier remain independent.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M1",
      "M2"
    ],
    "unique_kills": [
      "M1",
      "M2"
    ],
    "last_proven_fail": "M2: inline_products_shards_test.go:560: TestMarkdownInline_0064 native disagreement: case 13377 mode e byte 50388 lengths 171601/171605",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/u131/cache/M2 timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > M2.log 2>&1; inline_products_shards_test.go:560: TestMarkdownInline_0064 native disagreement: case 13377 mode e byte 50388 lengths 171601/171605",
    "raw_failing_line": "inline_products_shards_test.go:560: TestMarkdownInline_0064 native disagreement: case 13377 mode e byte 50388 lengths 171601/171605",
    "members_file": "family-members.json",
    "timing_samples": [
      7.408,
      7.52,
      7.52
    ],
    "witness_evidence": "W1: inline_products_shards_test.go:649: escaped delimiter parity survived (overlay raw line 637); comparison weakened to return nil; W1.log",
    "timing_scope": "32 of 2048 shards plus corpus union; full-family timing unknown"
  },
  {
    "test": "TestMarkdownInlineShardUnion",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/inline_products_shards_test.go",
    "seconds": 0.007,
    "oracle": "Self-written missing/repeated/foreign shard union assertions.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S1: inline_products_shards_test.go:667: invalid union accepted",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u131/evidence/S1.overlay.json ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > S1.log 2>&1; inline_products_shards_test.go:667: invalid union accepted",
    "raw_failing_line": "inline_products_shards_test.go:645: invalid union accepted",
    "members_file": "family-members.json",
    "timing_samples": [
      0.007,
      0.006,
      0.007
    ],
    "construction_kills": [
      "S1"
    ]
  },
  {
    "test": "TestMarkdownInlineShardSelector",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/inline_products_shards_test.go",
    "seconds": 0.007,
    "oracle": "Self-written selector bounds and selected-shard assertions.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S2B: inline_products_shards_test.go:683: accepted \"3/3\"",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u131/evidence/S2B.overlay.json ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > S2B.log 2>&1; inline_products_shards_test.go:683: accepted \"3/3\"",
    "raw_failing_line": "inline_products_shards_test.go:683: accepted \"3/3\"",
    "members_file": "family-members.json",
    "timing_samples": [
      0.007,
      0.007,
      0.007
    ],
    "construction_kills": [
      "S2B"
    ]
  },
  {
    "test": "TestMarkdownInlineProductSourceInputs",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/inline_products_shards_test.go",
    "seconds": 0.008,
    "oracle": "Self-written requirement that changed source content changes the product key.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S3: inline_products_shards_test.go:711: changed prerequisite reused a product key",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u131/evidence/S3.overlay.json ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > S3.log 2>&1; inline_products_shards_test.go:711: changed prerequisite reused a product key",
    "raw_failing_line": "inline_products_shards_test.go:704: changed prerequisite reused a product key",
    "members_file": "family-members.json",
    "timing_samples": [
      0.009,
      0.007,
      0.008
    ],
    "construction_kills": [
      "S3"
    ]
  },
  {
    "test": "TestMarkdownInlineShardAssignmentStable",
    "package": "stage1/cohere/markdowninline",
    "file": "stage1/cohere/markdowninline/inline_products_shards_test.go",
    "seconds": 0.007,
    "oracle": "Self-written stable hash assignment and coverage under corpus insertion.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S4: inline_products_shards_test.go:736: insertion moved text 1 mode w",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestDelimiterExpressionMatchesNode",
      "TestProduct_MarkdownInline family",
      "TestMarkdownInline family",
      "TestMarkdownInlineShardUnion",
      "TestMarkdownInlineShardSelector",
      "TestMarkdownInlineProductSourceInputs",
      "TestMarkdownInlineShardAssignmentStable"
    ],
    "evidence": "ADAMIC_MARKDOWNINLINE_LIBRARY=/tmp/u131/prettier/node_modules/prettier timeout 120 go test -json -count=1 -timeout 90s -overlay=/tmp/u131/evidence/S4.overlay.json ./stage1/cohere/markdowninline/ -run '^(TestDelimiterExpressionMatchesNode|TestProduct_MarkdownInlineGoOverlay|TestProduct_MarkdownInlineNode|TestProduct_MarkdownInlineOriginalLowered|TestProduct_MarkdownInlineOriginalNative|TestProduct_MarkdownInlineOriginalSanitized|TestProduct_MarkdownInlinePlantedLowered|TestProduct_MarkdownInlinePlantedNative|TestProduct_MarkdownInlinePlantedSanitized|TestProduct_MarkdownInlineParityLowered|TestProduct_MarkdownInlineParityNative|TestProduct_MarkdownInlineParitySanitized|TestProduct_MarkdownInlineParityNode|TestProduct_MarkdownInlinePipeLowered|TestProduct_MarkdownInlinePipeNative|TestProduct_MarkdownInlinePipeSanitized|TestProduct_MarkdownInlinePipeNode|TestProduct_MarkdownInlineFenceLowered|TestProduct_MarkdownInlineFenceNative|TestProduct_MarkdownInlineFenceSanitized|TestProduct_MarkdownInlineFenceNode|TestMarkdownInlineShardUnion|TestMarkdownInlineShardSelector|TestMarkdownInlineProductSourceInputs|TestMarkdownInlineShardAssignmentStable|TestMarkdownInlineUnion|TestMarkdownInline_0000|TestMarkdownInline_0064|TestMarkdownInline_0128|TestMarkdownInline_0192|TestMarkdownInline_0256|TestMarkdownInline_0320|TestMarkdownInline_0384|TestMarkdownInline_0448|TestMarkdownInline_0512|TestMarkdownInline_0576|TestMarkdownInline_0640|TestMarkdownInline_0704|TestMarkdownInline_0768|TestMarkdownInline_0832|TestMarkdownInline_0896|TestMarkdownInline_0960|TestMarkdownInline_1024|TestMarkdownInline_1088|TestMarkdownInline_1152|TestMarkdownInline_1216|TestMarkdownInline_1280|TestMarkdownInline_1344|TestMarkdownInline_1408|TestMarkdownInline_1472|TestMarkdownInline_1536|TestMarkdownInline_1600|TestMarkdownInline_1664|TestMarkdownInline_1728|TestMarkdownInline_1792|TestMarkdownInline_1856|TestMarkdownInline_1920|TestMarkdownInline_1984)$' > S4.log 2>&1; inline_products_shards_test.go:736: insertion moved text 1 mode w",
    "raw_failing_line": "inline_products_shards_test.go:736: insertion moved text 1 mode w",
    "members_file": "family-members.json",
    "timing_samples": [
      0.007,
      0.007,
      0.007
    ],
    "construction_kills": [
      "S4"
    ]
  }
]
```

| ID | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/markdowninline/inline.ts:194 | preferred > alternate ? other : quote -> preferred < alternate ? other : quote | TestMarkdownInline family |
| M2 | stage1/cohere/markdowninline/classes.ts:32 | else if(code > end) -> else if(code >= end) | TestMarkdownInline family |
| M3 | stage1/cohere/markdowninline/inline.ts:111 | marker === 61 \|\| marker === 45 -> marker === 62 \|\| marker === 45 | none |
| M4 | internal/native/runtime/regexp.c:912 | next == '$' ? replacement : input -> next != '$' ? replacement : input | TestDelimiterExpressionMatchesNode |
| P1 | stage1/cohere/markdowninline/inline.ts:226 | export function formatLeaf(mode: string, text: string): string { -> export function formatLeaf(mode: string, text: string): string { return ""; | TestMarkdownInline family |
| P2 | internal/native/runtime/regexp.c:919 | adamic_string *replacement, bool require_global) { -> adamic_string *replacement, bool require_global) { return &adamic_string_empty; | TestDelimiterExpressionMatchesNode |
| S1 | stage1/cohere/markdowninline/inline_products_shards_test.go:354 | Return nil from union validation | TestMarkdownInlineShardUnion |
| S2 | stage1/cohere/markdowninline/inline_products_shards_test.go:390 | n < 1 \|\| i < 0 -> n < 0 \|\| i < 0 | none |
| S3 | stage1/cohere/markdowninline/inline_products_shards_test.go:304 | Drop whole source-hashing loop | TestMarkdownInlineProductSourceInputs |
| S4 | stage1/cohere/markdowninline/inline_products_shards_test.go:345 | Use corpus index instead of name for key | TestMarkdownInlineShardAssignmentStable |
| S5 | stage1/cohere/markdowninline/inline_products_shards_test.go:347 | Drop last mode by bound -1 | TestMarkdownInline family, TestMarkdownInlineShardAssignmentStable |
| W1 | stage1/cohere/markdowninline/inline_products_shards_test.go:407 | Return nil from byte comparison | TestMarkdownInline family |
| S6 | stage1/cohere/markdowninline/inline_products_shards_test.go:134 | Drop whole required program.c write | TestMarkdownInline family, TestProduct_MarkdownInline family |
| S2B | stage1/cohere/markdowninline/inline_products_shards_test.go:390 | i >= n { -> i > n { | TestMarkdownInlineShardSelector |

Survivors: M3 leaves === unescaped, whereas clean native and Go cohere emit two backslashes before ===; native execution commands and exact bytes are in survivor-witness.json. S2 is a construction equivalent candidate, with the redundant zero guard explanation in S2-equivalence.txt.

The live scope is 2074 tests, not a file-sized unit. The entire clean package exhausted its 90-second binary budget at 90.049 seconds. Its timeout is not a red baseline. The narrowed, Prettier-enabled baseline passed in 6.853 seconds. The matrix runs all 26 non-shard tests and 32 shards spaced at ordinal multiples of 64, for 58 raw tests and seven grouped rows. The 2016 omitted shards are unknown, including their effects on uniqueness. Sacred here means unique within the bounded matrix only. Neither global package uniqueness nor repository uniqueness is proven.

The family rule combines all 2048 input wrappers with their corpus union. Whole generated wrapper bodies were mechanically checked against the shared-checker template; each differs only by ordinal. The checker also performs raw-input checks and built-in/planted disagreement witnesses at designated inputs. W1 disabled its comparison and made the embedded witness fail. I retain that evidence without turning this mixed family into a separate witness row. Full member lists are in family-members.json. There were no named slice rows to verify against the old 8de93800f4 inventory and no members vanished from the live list.

Timing the complete family three times would repeat the cooked run. Its reported median is explicitly for the bounded family subset plus corpus union. Product recipes were grouped because all call inlineBuild with different construction inputs. Successful compilation is their expected answer, so I judged that family as setup-check by omitting its required generated C file. That construction edit is separate from the four production mutants and does not establish formatting worthiness.

Warm tools did not provide the optional Prettier oracle. Stage3 API npm ci was run before baseline. Prettier 3.9.6 was installed in a dedicated directory using npm ci, then enabled for the bounded baseline, all timings and matrices. The first whole-package baseline had not enabled that optional oracle. No rows skipped in the accepted bounded runs. The TypeScript mutants also alter the source Node product, which would be a shared-error oracle alone; Go cohere and Prettier remain independent. The protected delimiter fixture and its Node execution were not edited. Its hand-written expected Node bytes were checked during the clean passing run, not assigned an external-authority label.

The brief asks for about three mutants per row but caps separately rebuilt port mutants at four. I used the four-rebuild cap, spread over title quoting, punctuation range lookup, pseudo-setext recognition and native RegExp replacement. Each has a private build cache. Construction edits, weakened comparisons and empty answers are labeled separately and contribute no production kills. Their inserted selector switches are absent because these are standalone rebuilds and private Go overlays.

S2 changed a selector zero-bound check but survived. It is an equivalent candidate: the existing index guard rejects every index when n=0. A predicate check over 121 pairs and the logical explanation are retained. S2B then changed the upper index bound and exposed acceptance of 3/3. M3 survives the bounded matrix, but a separately executed native === input differs from the clean native product and Go cohere. That is a bounded coverage gap; the omitted shards may catch it. No wider claim of unguarded package behavior is warranted.

Go overlays shorten function bodies, shifting reported line numbers. The report maps S1 raw line 645 to origin line 667, S3 raw 704 to origin 711, W1 raw 637 to origin 649, and S6 raw 315 to origin 318. Raw logs are retained. Native rebuild timings are product miss durations recorded by the build helper, not guessed clang-only durations; they may overlap due to parallel test execution. The matrix counts JSON fail actions, not intentional disagreement diagnostics printed by passing witnesses.

The tests generated an untracked go.work.sum, which was removed before committing evidence. The repository Markdown corpus is discovered dynamically, so evidence was copied into review only after all audit runs. Replaying after that copy adds audit Markdown to the corpus; the original corpus pin, counts and case IDs remain recorded in the logs. Full transitive runtime call coverage was not collected. The directly mutated functions and listed port functions were read, but no claim is made that every compiler/runtime helper was exhaustively inventoried. No other packages were run and no repository-wide replay was attempted.

Setup was skipped because env.sh worked. API npm ci reported 364 ms, retained in npm.log; Prettier npm ci reported 224 ms. The whole baseline binary cooked at 90.049 s; the accepted baseline took 6.853 s. All seven bounded rows ran alone three times. Sum of recorded audit command wall durations: 219.265 s, excluding initial baseline/dependency setup and analysis.

| Mutation/probe | Command wall s | Binary s | Logged product miss seconds |
|---|---:|---:|---:|
| M1 | 9.507 | 7.775 | 4.64 |
| M2 | 9.993 | 8.161 | 3.63 |
| M3 | 11.608 | 9.975 | 3.81 |
| M4 | 19.037 | 13.682 | 7.2 |
| P1 | 6.012 | 4.398 | 2.49 |
| P2 | 18.797 | 13.29 | 7.0 |

All standalone diffs apply to the starting source and compiled with their relevant build tool. Detailed build products, times, commands, raw failures, family members, probes and construction edits are included. Not covered: full-family cost, omitted shards, full transitive call coverage, repository-wide uniqueness.
