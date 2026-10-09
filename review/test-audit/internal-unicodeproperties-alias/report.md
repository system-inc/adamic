# u075 audit

Starting commit: fc31f9f15f95d841d4cd230a37ba55852866f2c8. nproc: 5. Warm env.sh works; core setup skipped.

All seven requested names exist in alias_test.go; none moved or vanished. Each has different assertions, so no scoped family grouping.

Code under test: Lookup, codePoint, expressionOK, splitPair, Property.Contains, Property.ContainsSequence, Set.Contains, Set.Complement, and generated alias/range/sequence tables.

Oracles: Unicode 17 aliases and ECMAScript names, handwritten membership examples, stored census hashes, and self invariants. Independent sampled verification is in authority-check.log.

Full clean baseline timed out at 90.013 seconds with no prior test failure. The bounded inactive-switch baseline passed in 9.985 seconds. The caller search excludes canonicalization scans, which call other production functions. Outside the nine named matrix rows results and uniqueness remain unknown.

All 25 standalone diffs applied to pristine origin source copies and passed go vet ./internal/unicodeproperties/. No test or oracle is mutated in these diffs.

## Rows

```json

[
  {
    "test": "TestBinaryAliases",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.002,
    "oracle": "ECMA-262 binary property table and Unicode 17 PropertyAliases.txt. Checked WSpace=White_Space=space against the downloaded source and newline membership with Node.",
    "oracle_kind": "external-authority",
    "kills": [
      "M14",
      "M18"
    ],
    "unique_kills": [
      "M18"
    ],
    "last_proven_fail": "M18: alias_test.go:170: \\p{WSpace}: got name \"White_Space\" value \"False\" ok true",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M18 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:170: \\p{WSpace}: got name \"White_Space\" value \"False\" ok true"
  },
  {
    "test": "TestGeneralCategoryAliases",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.003,
    "oracle": "Unicode 17 PropertyValueAliases.txt. Checked gc;Lu;Uppercase_Letter against the downloaded source. Shared Set pointer identity is a self expectation.",
    "oracle_kind": [
      "external-authority",
      "self"
    ],
    "kills": [
      "M14",
      "M17",
      "M19"
    ],
    "unique_kills": [
      "M19"
    ],
    "last_proven_fail": "M19: alias_test.go:192: \\p{Lu}: got name \"General_Category\" value \"Lowercase_Letter\" ok true",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M19 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:192: \\p{Lu}: got name \"General_Category\" value \"Lowercase_Letter\" ok true"
  },
  {
    "test": "TestScriptAliasSet",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.002,
    "oracle": "Unicode 17 script aliases and ECMAScript qualified-name rules; checked Latn=Latin against the downloaded source and Node rejecting sc=Hrkt. Count/digest and distinct Set pointers are self expectations, not independently regenerated.",
    "oracle_kind": [
      "external-authority",
      "self"
    ],
    "kills": [
      "M13",
      "M17"
    ],
    "unique_kills": [],
    "last_proven_fail": "M17: alias_test.go:230: script Latin: sc false scx false same set false",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestKnownMembership"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.002,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:230: script Latin: sc false scx false same set false"
  },
  {
    "test": "TestRejectedNames",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.002,
    "oracle": "ECMAScript exact property-name and v-flag rules. Manually checked Node rejects ascii, ASCII=Yes, sc=Hrkt and Basic_Emoji=Yes and accepts RGI_Emoji under v. The test itself does not run Node.",
    "oracle_kind": "external-authority",
    "kills": [
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12: alias_test.go:262: accepted \"RGI_Emoji\" without the v flag",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestKnownMembership"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P1"
    ],
    "subsumer_seconds": 0.002,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:262: accepted \"RGI_Emoji\" without the v flag"
  },
  {
    "test": "TestStringPropertyCensus",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.003,
    "oracle": "Stored counts and SHA256 hashes of our generated sequence output, plus self-written membership checks. This test does not run the separate Node comparison.",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M10",
      "M11",
      "M12"
    ],
    "unique_kills": [],
    "last_proven_fail": "M12: alias_test.go:316: Lookup RGI_Emoji_ZWJ_Sequence: ok false kind 0 len 0",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestKnownMembership"
    ],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P1",
      "P5"
    ],
    "subsumer_seconds": 0.002,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:316: Lookup RGI_Emoji_ZWJ_Sequence: ok false kind 0 len 0"
  },
  {
    "test": "TestSetBoundaries",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.033,
    "oracle": "Self-written ordered-range, endpoint, complement and double-complement invariants. These do not independently establish table membership.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04",
      "M05",
      "M06",
      "M07"
    ],
    "unique_kills": [
      "M04",
      "M05",
      "M06",
      "M07"
    ],
    "last_proven_fail": "M07: alias_test.go:348: set 0 complement agrees at an end",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P2",
      "P3"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:348: set 0 complement agrees at an end"
  },
  {
    "test": "TestKnownMembership",
    "package": "internal/unicodeproperties",
    "file": "internal/unicodeproperties/alias_test.go",
    "seconds": 0.002,
    "oracle": "Handwritten Unicode membership expectations. Checked scx=Latin membership for U+00B7, WSpace newline, gc=Lu A, and RGI_Emoji with Node 24 in this session; the test itself does not run Node.",
    "oracle_kind": "external-authority",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M08",
      "M09",
      "M10",
      "M11",
      "M12",
      "M13",
      "M17"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M17: alias_test.go:378: Lookup gc=Ll failed",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 20,
    "probe_kills": [
      "P1",
      "P4",
      "P5"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestBinaryAliases",
      "TestGeneralCategoryAliases",
      "TestScriptAliasSet",
      "TestRejectedNames",
      "TestStringPropertyCensus",
      "TestSetBoundaries",
      "TestKnownMembership",
      "TestNodeAgrees",
      "TestNodeStringProperties"
    ],
    "evidence": "ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/unicodeproperties/ -run \"^(TestBinaryAliases|TestGeneralCategoryAliases|TestScriptAliasSet|TestRejectedNames|TestStringPropertyCensus|TestSetBoundaries|TestKnownMembership|TestNodeAgrees|TestNodeStringProperties)$\"; alias_test.go:378: Lookup gc=Ll failed"
  }
]

```

## Mutants

| ID | Origin file:line | Change | Failed rows |

|---|---|---|---|

| M01 | internal/unicodeproperties/unicodeproperties.go:60 | `ranges[mid].End < cp` -> `ranges[mid].End <= cp` | TestSetBoundaries, TestKnownMembership |

| M02 | internal/unicodeproperties/unicodeproperties.go:66 | `ranges[lo].Start <= cp` -> `ranges[lo].Start < cp` | TestSetBoundaries, TestKnownMembership |

| M03 | internal/unicodeproperties/unicodeproperties.go:52 | `codePoint > 0x10FFFF` -> `codePoint > 0xFFFF` | TestSetBoundaries, TestKnownMembership |

| M04 | internal/unicodeproperties/unicodeproperties.go:79 | `r.Start > next` -> `r.Start >= next` | TestSetBoundaries |

| M05 | internal/unicodeproperties/unicodeproperties.go:80 | `End: r.Start - 1` -> `End: r.Start` | TestSetBoundaries |

| M06 | internal/unicodeproperties/unicodeproperties.go:85 | `next = r.End + 1` -> `next = r.End` | TestSetBoundaries |

| M07 | internal/unicodeproperties/unicodeproperties.go:88 | `End: 0x10FFFF` -> `End: 0x10FFFE` | TestSetBoundaries |

| M08 | internal/unicodeproperties/unicodeproperties.go:123 | `p.Kind != KindCodePoints` -> `p.Kind == KindCodePoints` | TestKnownMembership |

| M09 | internal/unicodeproperties/unicodeproperties.go:131 | `p.Kind != KindStrings` -> `p.Kind == KindStrings` | TestStringPropertyCensus, TestKnownMembership |

| M10 | internal/unicodeproperties/unicodeproperties.go:135 | `p.Sequences[i] >= text` -> `p.Sequences[i] > text` | TestStringPropertyCensus, TestKnownMembership |

| M11 | internal/unicodeproperties/unicodeproperties.go:137 | `p.Sequences[i] == text` -> `p.Sequences[i] != text` | TestStringPropertyCensus, TestKnownMembership |

| M12 | internal/unicodeproperties/unicodeproperties.go:175 | `if unicodeSets {` -> `if !unicodeSets {` | TestRejectedNames, TestStringPropertyCensus, TestKnownMembership, TestNodeStringProperties |

| M13 | internal/unicodeproperties/unicodeproperties.go:160 | `table = scriptExtensionNames` -> `table = scriptNames` | TestScriptAliasSet, TestKnownMembership |

| M14 | internal/unicodeproperties/unicodeproperties.go:188 | `Value: e.value` -> `Value: e.name` | TestBinaryAliases, TestGeneralCategoryAliases |

| M15 | internal/unicodeproperties/unicodeproperties.go:196 | `if expression == "" || expression == "=" { 		return false` -> `if expression == "" || expression == "=" { 		return true` |  |

| M16 | internal/unicodeproperties/unicodeproperties.go:204 | `equals > 1` -> `equals > 2` |  |

| M17 | internal/unicodeproperties/unicodeproperties.go:218 | `expression[i+1:]` -> `expression[i:]` | TestGeneralCategoryAliases, TestScriptAliasSet, TestKnownMembership, TestNodeAgrees |

| M18 | internal/unicodeproperties/tables.go:24925 | `"WSpace":                       {name: "White_Space", value: "True", set: 50}` -> `"WSpace":                       {name: "White_Space", value: "False", set: 50}` | TestBinaryAliases |

| M19 | internal/unicodeproperties/tables.go:24960 | `"Lu":                    {name: "General_Category", value: "Uppercase_Letter", set: 65}` -> `"Lu":                    {name: "General_Category", value: "Lowercase_Letter", set: 65}` | TestGeneralCategoryAliases |

| M20 | internal/unicodeproperties/tables.go:23 | `{0x0000, 0x007F}` -> `{0x0000, 0x007E}` | TestNodeAgrees |

## Survivors

M15: expressionOK("") changed false to true; Lookup still refuses the empty name. This internal lexical behavior is unguarded in the bounded matrix.

M16: expressionOK("gc=Lu=X") changed false to true; Lookup still refuses the malformed alias. This internal lexical behavior is unguarded in the bounded matrix.

## Probes

P1 Lookup -> (Property{}, false); P2 Set.Contains -> false; P3 Complement -> empty non-nil Set; P4 Property.Contains -> false; P5 ContainsSequence -> false. Probe kills are separate from mutants. For rows calling multiple public entries, vacuous requires all applicable empty-entry probes to pass; individual probe outcomes are retained in matrix.json.

## Brief issues and costs

M13 timed out at 90.016 seconds inside the Node disagreement reporter. All seven scoped rows and TestNodeStringProperties were rerun individually; TestNodeAgrees remains unknown for M13. matrix_rows lists the union of attempted rows, not a claim that every cell completed.

The supplied file reference names commit 8de93800f4, but the required fresh origin/main is fc31f9f1. All diffs and locations use the actual starting commit.

The 90-second whole-package limit prevents an exhaustive baseline from finishing even though the nine production callers fit comfortably. The instructions permit narrowing; package uniqueness is not claimed.

The aim of three mutants per row requests 21 mutations for seven rows while the cap is 20; used exactly 20.

The original planned M15 condition flip was rejected by go vet as contradictory. It was replaced before any mutant result was run by a constant change in the existing lexical rejection branch. All final standalone diffs pass vet.

The selector required explicit uint32 type parameters for generated range constants; the initial scratch compilation failed until those annotations were corrected. This was instrumentation work, not a mutant kill.

Core setup was skipped (0 seconds); npm ci succeeded, but its elapsed time was not separately recorded. Per-diff compilation/vet times are in replay-checks.json; matrix wall times include go invocation and Node work. No native products are built, so native cache isolation and rebuild timings do not apply.

The full baseline was still running when scratch selector construction began. Its already compiled test binary was unaffected; no mutant runs preceded the green bounded baseline.

TestScriptAliasSet only requires a nonempty canonical Value: M14 changed it to the property name and the row passed. TestSetBoundaries and TestKnownMembership both passed M20, which removed ASCII U+007F; their structural invariants and selected examples do not cover that edge, while TestNodeAgrees caught it.

The script census hashes and script-alias digest were not regenerated from upstream data. They remain self snapshot oracles. Manual Node checks do not turn these rows into external-run tests.

Subsumption is only a hint over the caught mutants in this fixed menu, not a deletion recommendation.

## Timing

Standalone vet total seconds: 6.419

Matrix invocation total seconds: 317.72

Three-run isolated test binary elapsed totals: 0.142

Clean whole baseline: 90.013 seconds. Bounded baseline: 9.985 seconds. No tests outside internal/unicodeproperties were run.

## Files

Raw JSON test logs, timing logs, fixed menu, standalone diffs, vet results, scripts and survivor witness logs are stored alongside this report. Production files are restored before the evidence commit.
