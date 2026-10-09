Unit u076: all 14 requested names exist; grouped into five rows.
Base 60397548dd8a9494a7d2aaa874b7648b8e625607; nproc 5; warm tools.
Whole package and three full-family timings cooked at 90 seconds; bounded baselines passed.
Twelve production mutants killed; closure accepts zero-answer canonicalization probes.
Verdicts: examples/closure sacred, legacy/Unicode family subsumed, batch construction setup-check.

```json
[
  {
    "test": "TestCanonicalizeExamples",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/canonicalize_test.go:18",
    "seconds": 0.002,
    "oracle": "Unicode 17 CaseFolding C/S mappings and ECMA-262 22.2.2.7.3 Canonicalize. Checked A -> a against downloaded CaseFolding.txt: 0041; C; 0061. Legacy a -> A also checked by Node.",
    "oracle_kind": "external-authority",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04",
      "M05",
      "M06",
      "M07",
      "M08",
      "M09",
      "M12"
    ],
    "unique_kills": [
      "M04"
    ],
    "last_proven_fail": "M12: canonicalize_test.go:63: CanonicalizeLegacy(U+0061) = U+0042, want U+0041",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P01",
      "P02",
      "P03",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "entry_vacuity": {
      "P01": false,
      "P02": false,
      "P03": false,
      "P04": false
    },
    "bounded": true,
    "matrix_rows": [
      "TestCanonicalizeExamples",
      "TestEquivalentsAreClosed",
      "TestCanonicalizeLegacyNode",
      "TestCanonicalizeUnicodeNodeRange family",
      "TestUnicodeNodeBatchOrderAndLimit"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run ^(TestCanonicalizeExamples|TestEquivalentsAreClosed|TestCanonicalizeLegacyNode|TestUnicodeNodeBatchOrderAndLimit|TestCanonicalizeUnicodeNodeRange0)$/^0000-0016-; ADAMIC_MUTANT=M12; canonicalize_test.go:63: CanonicalizeLegacy(U+0061) = U+0042, want U+0041"
  },
  {
    "test": "TestEquivalentsAreClosed",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/canonicalize_test.go:96",
    "seconds": 0.064,
    "oracle": "Self-written idempotence, containment, sortedness and class-consistency invariants; not an independent semantic oracle.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M05",
      "M06",
      "M07",
      "M08",
      "M09",
      "M10",
      "M11",
      "M12"
    ],
    "unique_kills": [
      "M10"
    ],
    "last_proven_fail": "M12: canonicalize_test.go:149: legacy U+0061 is listed with U+0041 but maps to U+0042",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P02",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": true,
    "entry_vacuity": {
      "P01": true,
      "P02": false,
      "P03": true,
      "P04": false
    },
    "bounded": true,
    "matrix_rows": [
      "TestCanonicalizeExamples",
      "TestEquivalentsAreClosed",
      "TestCanonicalizeLegacyNode",
      "TestCanonicalizeUnicodeNodeRange family",
      "TestUnicodeNodeBatchOrderAndLimit"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run ^(TestCanonicalizeExamples|TestEquivalentsAreClosed|TestCanonicalizeLegacyNode|TestUnicodeNodeBatchOrderAndLimit|TestCanonicalizeUnicodeNodeRange0)$/^0000-0016-; ADAMIC_MUTANT=M12; canonicalize_test.go:149: legacy U+0061 is listed with U+0041 but maps to U+0042",
    "vacuous_subcases": [
      "Entire row passes P01: CanonicalizeUnicode returns zero.",
      "Entire row passes P03: CanonicalizeLegacy returns zero."
    ]
  },
  {
    "test": "TestCanonicalizeLegacyNode",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/canonicalize_test.go:160",
    "seconds": 4.664,
    "oracle": "Node toUpperCase values and case-insensitive RegExp scans over every UTF-16 code unit; exact equivalence members compared, with scan counts.",
    "oracle_kind": "external-run",
    "kills": [
      "M07",
      "M08",
      "M09",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12: canonicalize_test.go:164: legacy value U+0061: tables U+0042, node U+0041",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEquivalentsAreClosed"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "P04"
    ],
    "subsumer_seconds": 0.064,
    "vacuous": false,
    "entry_vacuity": {
      "P04": false
    },
    "bounded": true,
    "matrix_rows": [
      "TestCanonicalizeExamples",
      "TestEquivalentsAreClosed",
      "TestCanonicalizeLegacyNode",
      "TestCanonicalizeUnicodeNodeRange family",
      "TestUnicodeNodeBatchOrderAndLimit"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run ^(TestCanonicalizeExamples|TestEquivalentsAreClosed|TestCanonicalizeLegacyNode|TestUnicodeNodeBatchOrderAndLimit|TestCanonicalizeUnicodeNodeRange0)$/^0000-0016-; ADAMIC_MUTANT=M12; canonicalize_test.go:164: legacy value U+0061: tables U+0042, node U+0041"
  },
  {
    "test": "TestCanonicalizeUnicodeNodeRange family",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/canonicalize_test.go:203",
    "seconds": null,
    "oracle": "Node RegExp iu/iv scans compared against generated class tables; table input preparation does not call CanonicalizeUnicode or UnicodeEquivalents. Only the first 16-line shard was in the bounded matrix.",
    "oracle_kind": "external-run",
    "kills": [
      "M11"
    ],
    "unique_kills": [],
    "last_proven_fail": "M11: canonicalize_test.go:327: BAD CLASS iu 61 EXTRA 41 MISSING 42",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestEquivalentsAreClosed"
    ],
    "mutants_in_matrix": 12,
    "probe_kills": [],
    "subsumer_seconds": 0.064,
    "vacuous": null,
    "entry_vacuity": {},
    "bounded": true,
    "matrix_rows": [
      "TestCanonicalizeExamples",
      "TestEquivalentsAreClosed",
      "TestCanonicalizeLegacyNode",
      "TestCanonicalizeUnicodeNodeRange family",
      "TestUnicodeNodeBatchOrderAndLimit"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run ^(TestCanonicalizeExamples|TestEquivalentsAreClosed|TestCanonicalizeLegacyNode|TestUnicodeNodeBatchOrderAndLimit|TestCanonicalizeUnicodeNodeRange0)$/^0000-0016-; ADAMIC_MUTANT=M11; canonicalize_test.go:327: BAD CLASS iu 61 EXTRA 41 MISSING 42",
    "members": [
      "TestCanonicalizeUnicodeNodeRange0",
      "TestCanonicalizeUnicodeNodeRange1",
      "TestCanonicalizeUnicodeNodeRange2",
      "TestCanonicalizeUnicodeNodeRange3",
      "TestCanonicalizeUnicodeNodeRange4",
      "TestCanonicalizeUnicodeNodeRange5",
      "TestCanonicalizeUnicodeNodeRange6",
      "TestCanonicalizeUnicodeNodeRange7",
      "TestCanonicalizeUnicodeNodeRange8",
      "TestCanonicalizeUnicodeNodeRange9",
      "TestCanonicalizeUnicodeNodeRange10",
      "TestCanonicalizeUnicodeNodeRange11",
      "TestCanonicalizeUnicodeNodeRange12",
      "TestCanonicalizeUnicodeNodeRange13",
      "TestCanonicalizeUnicodeNodeRange14",
      "TestCanonicalizeUnicodeNodeRange15",
      "TestCanonicalizeUnicodeNodeRange16",
      "TestCanonicalizeUnicodeNodeRange17",
      "TestCanonicalizeUnicodeNodeRange18",
      "TestCanonicalizeUnicodeNodeRange19",
      "TestCanonicalizeUnicodeNodeRange20",
      "TestCanonicalizeUnicodeNodeRange21",
      "TestCanonicalizeUnicodeNodeRange22",
      "TestCanonicalizeUnicodeNodeRange23",
      "TestCanonicalizeUnicodeNodeRange24",
      "TestCanonicalizeUnicodeNodeRange25",
      "TestCanonicalizeUnicodeNodeRange26",
      "TestCanonicalizeUnicodeNodeRange27",
      "TestCanonicalizeUnicodeNodeRange28",
      "TestCanonicalizeUnicodeNodeRange29",
      "TestCanonicalizeUnicodeNodeRange30",
      "TestCanonicalizeUnicodeNodeRange31",
      "TestCanonicalizeUnicodeNodeRange32",
      "TestCanonicalizeUnicodeNodeRange33"
    ],
    "timing_lower_bound_seconds": 90,
    "timing_status": "All three full-family attempts timed out at 90 seconds; no completed median."
  },
  {
    "test": "TestUnicodeNodeBatchOrderAndLimit",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/canonicalize_test.go:428",
    "seconds": 0.003,
    "oracle": "Self-written expected peak=2 and ordered callback-result indices for suite-owned unicodeNodeBatches. S02 rotates slots and fails the order assertion; S03 returns empty results and fails peak=2.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "S03: canonicalize_test.go:460: peak workers 0, want GOMAXPROCS=2",
    "verdict": "setup-check",
    "subsumed_by": [],
    "mutants_in_matrix": 12,
    "probe_kills": [
      "S03"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "entry_vacuity": {
      "S03": false
    },
    "bounded": true,
    "matrix_rows": [
      "TestCanonicalizeExamples",
      "TestEquivalentsAreClosed",
      "TestCanonicalizeLegacyNode",
      "TestCanonicalizeUnicodeNodeRange family",
      "TestUnicodeNodeBatchOrderAndLimit"
    ],
    "evidence": "timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run ^TestUnicodeNodeBatchOrderAndLimit$; canonicalize_test.go:460: peak workers 0, want GOMAXPROCS=2",
    "construction_kills": [
      "S01",
      "S02"
    ]
  }
]
```

| ID | origin/main location | Change | Failed rows |
|---|---|---|---|
| M01 | internal/unicodeproperties/canonicalize.go:11 | flip condition: `if codePoint >= 0 || codePoint > 0x10FFFF { 		return codePoint` | E, C |
| M02 | internal/unicodeproperties/canonicalize.go:15 | flip comparison: `return unicodeFold[i][0] > cp` | E, C |
| M03 | internal/unicodeproperties/canonicalize.go:17 | change constant index: `return rune(unicodeFold[index][0])` | E, C |
| M04 | internal/unicodeproperties/canonicalize.go:28 | off-by-one bound: `if codePoint < -1 || codePoint > 0x10FFFF { 		return nil` | E |
| M05 | internal/unicodeproperties/canonicalize.go:36 | drop whole copy loop: `` | E, C |
| M06 | internal/unicodeproperties/canonicalize.go:41 | off-by-one value: `return []rune{codePoint + 1}` | E, C |
| M07 | internal/unicodeproperties/canonicalize.go:52 | flip comparison: `return legacyFold[i][0] > codeUnit` | E, L, C |
| M08 | internal/unicodeproperties/canonicalize.go:68 | drop statement: `` | E, L, C |
| M09 | internal/unicodeproperties/canonicalize.go:71 | off-by-one value: `return []uint16{codeUnit + 1}` | E, L, C |
| M10 | internal/unicodeproperties/canonicalize_tables.go:1532 | change last Unicode fold target constant: `{0x1E921, 0x1E944}, }  // unicodeClassCanon[i]` | C |
| M11 | internal/unicodeproperties/canonicalize_tables.go:3022 | change first Unicode class member constant: `var unicodeClass = [][]uint32{ 	{0x0042, 0x0061},` | U, C |
| M12 | internal/unicodeproperties/canonicalize_tables.go:4508 | change first legacy fold target constant: `var legacyFold = [][2]uint16{ 	{0x0061, 0x0042},` | E, L, C |

Survivors: none among the twelve production mutants in the bounded matrix.

Scope, oracles, findings and brief costs:
- E=examples, C=closure, L=legacy Node, U=Unicode Node range family, B=batch order/limit. Both requested and additional range wrappers have identical checker calls with only a range index changed. U has all 34 members; only Range0/0000-0016-* was replayed in the matrix. Requested Range1 through Range9 remain members with unobserved mutation outcomes. No requested name moved or vanished.
- The brief's reference commit 8de93800f4 differs from fetched main 60397548dd8a9494a7d2aaa874b7648b8e625607. Main now has 34 ranges, not just the ten listed. Locations and standalone diffs refer to the actual starting commit.
- Baseline whole package timed out at 90.014 seconds without an observed test assertion failure. Three full-family timings also timed out. No median for a completed full family exists; null is intentional. The 90-second cap prevented completing this census. Repeating family attempts was solely to meet the three timing-run request, not to reuse the full package per mutant.
- Coverage/dispatch and planted-failure tests in shards_test.go assert additional construction and witness properties, so are separate rows rather than interchangeable wrappers of the range checker. They are outside the requested unit and excluded from the bounded mutation matrix. Their production-mutant outcomes are unknown. No package- or repo-wide uniqueness is claimed.
- The bounded matrix includes every direct caller of the four production functions in these tests: examples, closure and the full legacy scan. Unicode table reads additionally run the first complete 16-line range shard. All other package rows, other Unicode shards and unrun family members remain unknown. The family and legacy subsumption are small-set hints, not deletion advice.
- The code-under-test inventory is code-under-test.txt. Four production functions and six generated fold/class arrays are reached. No outside package tests ran. The independent Node scripts and assertions were never mutated. Go vet validated every standalone production and empty-answer diff. The scratch switch was compiled once and its clean bounded baseline passed before matrix runs.
- The Unicode differential family checks generated equivalence tables, not the CanonicalizeUnicode function. M01/M02/M03 break that function and the bounded family still passes. Its table mutation M11 is caught by closure as well. The examples have unique M04; closure has unique M10 within this matrix.
- Closure passes when CanonicalizeUnicode returns zero (P01), and also when CanonicalizeLegacy returns zero (P03). Each causes equivalence lookup to fall back to singleton classes, preserving the self-consistency checks while erasing case folding. This row is vacuous for those entries despite catching meaningful non-probe mutations. It rejects empty equivalence slices P02/P04. The JSON records per-entry vacuity because a single Boolean cannot distinguish these results.
- Examples call all four entries and reject all four empty answers. Legacy Node calls LegacyEquivalents directly, so P04 determines its vacuity. Its transitive CanonicalizeLegacy probe P03 is also observed but is not listed as its own entry probe. The Unicode table family has no production function entry; its vacuity is null. Returning no harness-generated scan lines would mutate preparation, not production, so no such probe was used.
- The batch-order row is a setup-check, not a production test. S01 dropped the worker-cap assignment; S02 rotates result slots; S03 returns empty batch results. All are separate construction evidence, never production kills. S01 reported peak workers 1, not excessive concurrency, so it does not independently prove the cap bound. S02 preserves the comparisons and proves the order check can fail. S03 is the setup entry's empty-answer probe.
- Unicode 17 CaseFolding.txt was downloaded independently and A -> a checked against its 0041; C; 0061 mapping. Node independently reported legacy a -> A and Kelvin matching k under iu. Closure and worker expectations are self, not external authority. Node scans compare exact members as well as scan counts.
- Generated table edits change values only; no generator, Node oracle, assertion or comparison was altered. M05 drops the entire copy loop, keeping declarations used and the package compilable. Switch-only scaffolding is not present in standalone diffs and carries no verdict.
- No opt-in or missing-tool skip event was observed in baseline or bounded logs. A timed-out baseline cannot establish that unrun tests would not skip. The final restored bounded run passed. No broad repository gate was run because the brief forbids other packages.

Timing and coverage limits:
Toolchain setup skipped; npm ci installed 3 packages in 0.481 s; nproc=5. Package baseline 90.014 binary seconds, over budget. Full family three attempts were capped at 90 seconds, recorded individually. Individual three-run binary seconds follow; they are completed package ok-line seconds, not parent durations that exclude parallel children.
TestCanonicalizeExamples: samples [0.002, 0.002, 0.002]; median 0.002.
TestEquivalentsAreClosed: samples [0.072, 0.064, 0.063]; median 0.064.
TestCanonicalizeLegacyNode: samples [4.634, 4.664, 4.695]; median 4.664.
TestCanonicalizeUnicodeNodeRange family: samples []; median None.
TestUnicodeNodeBatchOrderAndLimit: samples [0.004, 0.003, 0.002]; median 0.003.
Scratch compile: 0.609 s; validation total 2.386 s.
M01: command wall 4.985 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M02: command wall 5.007 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M03: command wall 5.001 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M04: command wall 4.884 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M05: command wall 4.951 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M06: command wall 5.015 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M07: command wall 5.023 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M08: command wall 5.237 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M09: command wall 6.372 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M10: command wall 4.998 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M11: command wall 4.948 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
M12: command wall 4.927 s; binary None s (failed commands have no ok line; elapsed events are in matrix.json).
No native products were built, so no native cache/rebuild timing applies. The mutation switch changes only this Go package and its static data. Full Unicode duration and kills outside the bounded set were not covered.
