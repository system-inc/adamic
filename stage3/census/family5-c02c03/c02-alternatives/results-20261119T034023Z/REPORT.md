# C02 whole reach and scratch alternatives

Base `fbe5b72ba115a8f3060d4b5229fd578574f4505f`; C02 `12b9e04184663f899acb76505043ad292966f236`. TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. No adaptation, ledger or internal/ edit, and no landing push. All measurements ran sequentially on the same box. Candidate changes exist only in scratch trees; patches and manifests are evidence.

## Whole-entry first-error reach

[Every C02 callable-field site and every new, gone, changed or reached finding, with full before/after words](../whole-entry/REPORT.md). Both raw streams were pushed immediately after completion, before the next side started.

| Side | Real exit | Wall seconds | Stream commit |
|---|---:|---:|---|
| base | 0 | 2537.57 | `2c8ef70a7469d03a02d2fee51fb713ff9220c02c` |
| c02 | 0 | 2648.544 | `0ea786e0eea609ee629911446db28b376f0237a7` |

At C02, `a checked view with callable field getPositionOfLineAndCharacter` appears in 1 events: 1 distinct primary diagnostic positions, 1 positions including Boundary mirrors.

```json
{
  "tsc": {
    "before_events": 3713,
    "after_events": 3699,
    "target_before_distinct_sites": 14,
    "target_after_distinct_sites": 0,
    "after_target_mentions_all_kinds": 0,
    "grades": {
      "changed": 1,
      "gone": 14
    }
  }
}
```

Whole-entry pairing: **new 0, gone 14, changed 1, reached 0**. The callable refusal is at utilities.ts:5339:11 in getFunctionFlags (unit 5317:1), replacing the pre-existing NotYet about checked-view modifiers. This is one observable site in the requested first-error pass, not a global full-mode count.

## Scratch candidates

(a): `/** @internal */ getPositionOfLineAndCharacter?(line: number, character: number, allowEdits?: true): number;` on compiler SourceFile, types.ts:4326 owner.

(b): `/** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;` on that same SourceFile owner.

(c): the 14 source expressions gain `satisfies SourceFileLike`; type-only imports are added in checker, parser and program. No declaration change. Stock TypeScript retains the expression’s source type; the typing evidence below reports it rather than assuming this is a narrower view.

Compiler stock checks use all 79 roots in both the inherited and explicit strict configurations. A separate stock services-source program tests the merge with services/types.ts:172, without project-reference declaration output hiding it. Its baseline and introduced diagnostics are retained. Commented and comment-free ES2024/ESNext JavaScript emits are compared on every changed source; actual emitted bundles and public declarations are compared separately. Missing harness bundles are identified as not built by the local task, not claimed identical.

| Candidate | Stock 79 | Strict 79 | Services probe | Source JS | Emitted JS | Public build | Public cmp |
|---|---:|---:|---:|---:|---:|---:|---:|
| a | 0 | 0 | 1 | 0 | 0 | 1 | — |
| b | 0 | 0 | 1 | 0 | 0 | 1 | — |
| c | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

a / stock: 79 roots, 0 diagnostics.

a / stock-strict: 79 roots, 0 diagnostics.

a / services-merge: 168 roots, 3 diagnostics.

- TS2722 `src/services/formatting/smartIndenter.ts:594:27`: Cannot invoke an object which is possibly 'undefined'.

- TS18048 `src/services/formatting/smartIndenter.ts:594:27`: 'sourceFile.getPositionOfLineAndCharacter' is possibly 'undefined'.

- TS2386 `src/services/types.ts:172:9`: Overload signatures must all be optional or required.

a emitted JavaScript: 17 files compared, 0 differences. Baseline-only files not produced in this build: `["built/local/run.js", "built/local/tsserverlibrary.js"]`. A failed public build does not establish identity for files it failed to produce.

a public API:
```json
{
  "produced": false,
  "build_exit": 1,
  "baseline_sha256": "65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd"
}

```

a services merge relative to the stock baseline:
```json
{
  "baseline_diagnostics": 0,
  "candidate_diagnostics": 3,
  "introduced_diagnostics": [
    [
      2722,
      "src/services/formatting/smartIndenter.ts",
      "Cannot invoke an object which is possibly 'undefined'."
    ],
    [
      18048,
      "src/services/formatting/smartIndenter.ts",
      "'sourceFile.getPositionOfLineAndCharacter' is possibly 'undefined'."
    ],
    [
      2386,
      "src/services/types.ts",
      "Overload signatures must all be optional or required."
    ]
  ]
}

```

a full census not run: the compiler input is not checker-clean, or (a) was killed by its introduced merge error, as instructed.

b / stock: 79 roots, 0 diagnostics.

b / stock-strict: 79 roots, 0 diagnostics.

b / services-merge: 168 roots, 4 diagnostics.

- TS2300 `src/compiler/types.ts:4327:22`: Duplicate identifier 'getPositionOfLineAndCharacter'.

- TS2722 `src/services/formatting/smartIndenter.ts:594:27`: Cannot invoke an object which is possibly 'undefined'.

- TS18048 `src/services/formatting/smartIndenter.ts:594:27`: 'sourceFile.getPositionOfLineAndCharacter' is possibly 'undefined'.

- TS2300 `src/services/types.ts:172:9`: Duplicate identifier 'getPositionOfLineAndCharacter'.

b emitted JavaScript: 17 files compared, 0 differences. Baseline-only files not produced in this build: `["built/local/run.js", "built/local/tsserverlibrary.js"]`. A failed public build does not establish identity for files it failed to produce.

b public API:
```json
{
  "produced": false,
  "build_exit": 1,
  "baseline_sha256": "65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd"
}

```

b services merge relative to the stock baseline:
```json
{
  "baseline_diagnostics": 0,
  "candidate_diagnostics": 4,
  "introduced_diagnostics": [
    [
      2300,
      "src/compiler/types.ts",
      "Duplicate identifier 'getPositionOfLineAndCharacter'."
    ],
    [
      2722,
      "src/services/formatting/smartIndenter.ts",
      "Cannot invoke an object which is possibly 'undefined'."
    ],
    [
      18048,
      "src/services/formatting/smartIndenter.ts",
      "'sourceFile.getPositionOfLineAndCharacter' is possibly 'undefined'."
    ],
    [
      2300,
      "src/services/types.ts",
      "Duplicate identifier 'getPositionOfLineAndCharacter'."
    ]
  ]
}

```

c / stock: 79 roots, 0 diagnostics.

c / stock-strict: 79 roots, 0 diagnostics.

c / services-merge: 168 roots, 0 diagnostics.

c emitted JavaScript: 18 files compared, 0 differences. Baseline-only files not produced in this build: `["built/local/run.js"]`. A failed public build does not establish identity for files it failed to produce.

c public API:
```json
{
  "produced": true,
  "build_exit": 0,
  "baseline_sha256": "65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd",
  "candidate_sha256": "65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd",
  "byte_identical": true
}

```

c services merge relative to the stock baseline:
```json
{
  "baseline_diagnostics": 0,
  "candidate_diagnostics": 0,
  "introduced_diagnostics": []
}

```

The first unified services-source probes inherited services’ empty types list and therefore reported 64 missing-Node-global diagnostics at base. Those raw attempts are retained. The table and merge diagnostics above use the corrected Node-global controls when present; their baseline is reported separately below.

Corrected services baseline: 168 roots, 0 diagnostics; real exit 0.

## Every proposed site annotation in (c)

| Original site | Source expression | Proposed expression | Type-only |
|---|---|---|---|
| `src/compiler/checker.ts:49348:52` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/parser.ts:10015:36` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/parser.ts:10015:48` | `result` | `result satisfies SourceFileLike` | yes |
| `src/compiler/parser.ts:10087:37` | `origSourceFile` | `origSourceFile satisfies SourceFileLike` | yes |
| `src/compiler/parser.ts:10244:42` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/program.ts:738:57` | `file` | `file satisfies SourceFileLike` | yes |
| `src/compiler/program.ts:739:76` | `file` | `file satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:1261:24` | `getSourceFileOfNode(node)` | `getSourceFileOfNode(node) satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:2532:70` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:7945:37` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:7950:37` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:7960:37` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:7972:37` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |
| `src/compiler/utilities.ts:7978:37` | `sourceFile` | `sourceFile satisfies SourceFileLike` | yes |

Stock TypeScript typing for the annotated expressions:
```json
[
  {
    "file": "src/compiler/utilities.ts",
    "line": 1261,
    "column": 24,
    "expression": "getSourceFileOfNode(node)",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/utilities.ts",
    "line": 2532,
    "column": 70,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/utilities.ts",
    "line": 7945,
    "column": 37,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/utilities.ts",
    "line": 7950,
    "column": 37,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/utilities.ts",
    "line": 7960,
    "column": 37,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/utilities.ts",
    "line": 7972,
    "column": 37,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/utilities.ts",
    "line": 7978,
    "column": 37,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/parser.ts",
    "line": 10016,
    "column": 36,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/parser.ts",
    "line": 10016,
    "column": 73,
    "expression": "result",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/parser.ts",
    "line": 10088,
    "column": 37,
    "expression": "origSourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/parser.ts",
    "line": 10245,
    "column": 42,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/checker.ts",
    "line": 49349,
    "column": 52,
    "expression": "sourceFile",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/program.ts",
    "line": 739,
    "column": 57,
    "expression": "file",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  },
  {
    "file": "src/compiler/program.ts",
    "line": 740,
    "column": 76,
    "expression": "file",
    "constraint": "SourceFileLike",
    "source_type": "SourceFile",
    "result_type": "SourceFile"
  }
]
```

## Full single-file censuses

Physical S/T/P site counts are deduplicated by diagnostic position. Pairing grades count primary events by phase/unit; a full census can revisit one position in several units. Boundary mirrors are counted separately.

All full runs use the tsc.ts entry, exact absolute LATENT_ONLY_FILE, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1 and CommonJSBundle. utilities.ts, parser.ts, program.ts and checker.ts are the four distinct files holding the 14 G02 sites; utilities.ts also contains getFunctionFlags. Complete streams, complete stderr, real child exits, source and tool hashes are retained. The previously completed isolated utilities baseline is reused only when its tool, types.ts and utilities.ts hashes match. It is not counted as a new run.

| Candidate | G02 before | G02 after | New | Gone | Changed | Reached | All stock/emit/API prerequisites green |
|---|---:|---:|---:|---:|---:|---:|---|
| a | — | — | — | — | — | — | no completed census comparison |
| b | 14 | 0 | 0 | 14 | 0 | 0 | False |
| c | 14 | 14 | 26 | 0 | 14 | 0 | True |

b:
```json
{
  "checker": {
    "before_events": 8220,
    "after_events": 8219,
    "target_before_distinct_sites": 1,
    "target_after_distinct_sites": 0,
    "after_target_mentions_all_kinds": 0,
    "grades": {
      "gone": 1
    }
  },
  "parser": {
    "before_events": 2512,
    "after_events": 2508,
    "target_before_distinct_sites": 4,
    "target_after_distinct_sites": 0,
    "after_target_mentions_all_kinds": 0,
    "grades": {
      "gone": 4
    }
  },
  "program": {
    "before_events": 1473,
    "after_events": 1471,
    "target_before_distinct_sites": 2,
    "target_after_distinct_sites": 0,
    "after_target_mentions_all_kinds": 0,
    "grades": {
      "gone": 2
    }
  },
  "utilities": {
    "before_events": 1309,
    "after_events": 1302,
    "target_before_distinct_sites": 7,
    "target_after_distinct_sites": 0,
    "after_target_mentions_all_kinds": 0,
    "grades": {
      "gone": 7
    }
  }
}
```

Primary (non-Boundary) event grades: `{"gone": 14}`; Boundary mirror grades: `{}`.

New/changed primary events outside the original G02 sites: `{}`.

### b.1: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/checker.ts:49348:52`; unit `/tmp/family5-base-tree/src/compiler/checker.ts:49295:5`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/checker.ts:49348:52: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.2: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10015:36`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9930:5`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10015:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.3: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10015:48`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9930:5`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10015:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.4: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10087:37`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:10079:9`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10087:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.5: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10244:42`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:10229:9`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10244:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.6: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/program.ts:738:57`; unit `/tmp/family5-base-tree/src/compiler/program.ts:717:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/program.ts:738:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.7: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/program.ts:739:76`; unit `/tmp/family5-base-tree/src/compiler/program.ts:717:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/program.ts:739:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.8: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:1261:24`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:1240:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:1261:24: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.9: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:2532:70`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:2524:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:2532:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.10: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7945:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7943:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7945:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.11: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7950:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7949:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7950:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.12: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7960:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7959:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7960:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.13: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7972:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7969:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7972:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### b.14: gone

Scope: outside original G02 sites.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7978:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7976:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7978:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

c:
```json
{
  "checker": {
    "before_events": 8220,
    "after_events": 8224,
    "target_before_distinct_sites": 1,
    "target_after_distinct_sites": 0,
    "target_after_equivalent_fact_distinct_sites": 1,
    "equivalent_relation_proof_distinct_sites": 1,
    "after_target_mentions_all_kinds": 3,
    "grades": {
      "changed": 1,
      "new": 4
    }
  },
  "parser": {
    "before_events": 2512,
    "after_events": 2526,
    "target_before_distinct_sites": 4,
    "target_after_distinct_sites": 0,
    "target_after_equivalent_fact_distinct_sites": 4,
    "equivalent_relation_proof_distinct_sites": 4,
    "after_target_mentions_all_kinds": 18,
    "grades": {
      "changed": 4,
      "new": 14
    }
  },
  "program": {
    "before_events": 1473,
    "after_events": 1493,
    "target_before_distinct_sites": 2,
    "target_after_distinct_sites": 0,
    "target_after_equivalent_fact_distinct_sites": 2,
    "equivalent_relation_proof_distinct_sites": 2,
    "after_target_mentions_all_kinds": 6,
    "grades": {
      "changed": 2,
      "new": 20
    }
  },
  "utilities": {
    "before_events": 1309,
    "after_events": 1323,
    "target_before_distinct_sites": 7,
    "target_after_distinct_sites": 0,
    "target_after_equivalent_fact_distinct_sites": 7,
    "equivalent_relation_proof_distinct_sites": 7,
    "after_target_mentions_all_kinds": 21,
    "grades": {
      "changed": 7,
      "new": 14
    }
  }
}
```

Primary (non-Boundary) event grades: `{"changed": 14, "new": 26}`; Boundary mirror grades: `{"new": 26}`.

New/changed primary events outside the original G02 sites: `{"new": 9}`.

### c.1: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/checker.ts:49348:52`; unit `/tmp/family5-base-tree/src/compiler/checker.ts:49295:5`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/checker.ts:49348:52: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/checker.ts:49349:52`; unit `/workspace/family5-alt-c/src/compiler/checker.ts:49296:5`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/checker.ts:49349:52: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.2: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/checker.ts:49349:9`; unit `/workspace/family5-alt-c/src/compiler/checker.ts:49296:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/checker.ts:49349:52: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.3: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/checker.ts:49350:9`; unit `/workspace/family5-alt-c/src/compiler/checker.ts:49296:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/checker.ts:49350:111: stage 0 can't lower reading start yet
```

### c.4: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/checker.ts:49350:111`; unit `/workspace/family5-alt-c/src/compiler/checker.ts:49296:5`; NotYet / lowering.

Reason: reading start

```text
/workspace/family5-alt-c/src/compiler/checker.ts:49350:111: stage 0 can't lower reading start yet
```

### c.5: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/checker.ts:49349:52`; unit `/workspace/family5-alt-c/src/compiler/checker.ts:49296:5`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/checker.ts:49349:52: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.6: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10015:36`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9930:5`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10015:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10016:36`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:9931:5`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10016:36: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.7: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10015:48`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9930:5`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10015:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10016:73`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:9931:5`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10016:73: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.8: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10087:37`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:10079:9`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10087:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:37`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10080:9`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.9: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/parser.ts:10244:42`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:10229:9`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10244:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10245:42`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10230:9`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10245:42: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.10: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10016:9`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:9931:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10016:36: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.11: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:13`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10069:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.12: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:13`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10070:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.13: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:13`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10071:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.14: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:13`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10080:9`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.15: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10245:17`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10217:5`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10245:42: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.16: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10245:17`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10230:9`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10245:42: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.17: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10016:36`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:9931:5`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10016:36: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.18: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:37`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10069:5`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.19: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:37`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10070:5`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.20: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:37`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10071:5`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.21: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10088:37`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10080:9`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10088:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.22: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10245:42`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10217:5`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10245:42: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.23: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/parser.ts:10245:42`; unit `/workspace/family5-alt-c/src/compiler/parser.ts:10230:9`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/parser.ts:10245:42: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.24: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/program.ts:738:57`; unit `/tmp/family5-base-tree/src/compiler/program.ts:717:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/program.ts:738:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/program.ts:739:57`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/program.ts:739:57: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.25: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/program.ts:739:76`; unit `/tmp/family5-base-tree/src/compiler/program.ts:717:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/program.ts:739:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/program.ts:740:76`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/program.ts:740:76: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.26: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/program.ts:739:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:739:57: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.27: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/program.ts:740:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:740:76: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.28: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:742:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:742:9: stage 0 can't lower assigning to an Identifier yet
```

### c.29: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:743:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:743:9: stage 0 can't lower assigning to an Identifier yet
```

### c.30: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:747:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:747:20: stage 0 can't lower reading lineContent yet
```

### c.31: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:757:13`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:757:24: stage 0 can't lower reading lineContent yet
```

### c.32: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:758:13`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:758:24: stage 0 can't lower reading lineContent yet
```

### c.33: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:761:13`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:761:24: stage 0 can't lower reading lineContent yet
```

### c.34: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:765:13`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:765:24: stage 0 can't lower reading lineContent yet
```

### c.35: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:741:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/program.ts:741:43: stage 0 can't lower reading lineStart yet
```

### c.36: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:742:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: assigning to an Identifier

```text
/workspace/family5-alt-c/src/compiler/program.ts:742:9: stage 0 can't lower assigning to an Identifier yet
```

### c.37: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:743:9`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: assigning to an Identifier

```text
/workspace/family5-alt-c/src/compiler/program.ts:743:9: stage 0 can't lower assigning to an Identifier yet
```

### c.38: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:747:20`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: reading lineContent

```text
/workspace/family5-alt-c/src/compiler/program.ts:747:20: stage 0 can't lower reading lineContent yet
```

### c.39: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:757:24`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: reading lineContent

```text
/workspace/family5-alt-c/src/compiler/program.ts:757:24: stage 0 can't lower reading lineContent yet
```

### c.40: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:758:24`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: reading lineContent

```text
/workspace/family5-alt-c/src/compiler/program.ts:758:24: stage 0 can't lower reading lineContent yet
```

### c.41: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:761:24`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: reading lineContent

```text
/workspace/family5-alt-c/src/compiler/program.ts:761:24: stage 0 can't lower reading lineContent yet
```

### c.42: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:765:24`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: reading lineContent

```text
/workspace/family5-alt-c/src/compiler/program.ts:765:24: stage 0 can't lower reading lineContent yet
```

### c.43: new

Scope: outside original G02 sites.
after: `/workspace/family5-alt-c/src/compiler/program.ts:741:43`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; NotYet / lowering.

Reason: reading lineStart

```text
/workspace/family5-alt-c/src/compiler/program.ts:741:43: stage 0 can't lower reading lineStart yet
```

### c.44: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/program.ts:739:57`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/program.ts:739:57: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.45: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/program.ts:740:76`; unit `/workspace/family5-alt-c/src/compiler/program.ts:718:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/program.ts:740:76: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.46: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:1261:24`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:1240:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:1261:24: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:1261:24`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:1240:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:1261:24: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.47: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:2532:70`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:2524:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:2532:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:2532:70`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:2524:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:2532:70: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.48: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7945:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7943:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7945:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7945:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7943:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7945:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.49: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7950:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7949:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7950:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7950:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7949:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7950:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.50: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7960:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7959:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7960:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7960:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7959:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7960:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.51: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7972:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7969:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7972:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7972:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7969:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7972:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.52: changed

Scope: original G02 site or its diagnostic mirror.
before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7978:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7976:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7978:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7978:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7976:1`; Refused / refusal_scan.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7978:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.53: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:1261:9`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:1240:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:1261:24: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.54: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:2532:13`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:2524:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:2532:70: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.55: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7945:5`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7943:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7945:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.56: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7950:5`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7949:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7950:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.57: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7960:5`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7959:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7960:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.58: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7972:5`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7969:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7972:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.59: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7978:5`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7976:1`; Boundary / lowering.

Reason: failed statement; state rolled back; continued at next statement or declaration

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7978:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.60: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:1261:24`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:1240:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:1261:24: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.61: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:2532:70`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:2524:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:2532:70: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.62: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7945:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7943:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7945:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.63: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7950:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7949:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7950:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.64: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7960:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7959:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7960:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.65: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7972:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7969:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7972:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

### c.66: new

Scope: original G02 site or its diagnostic mirror.
after: `/workspace/family5-alt-c/src/compiler/utilities.ts:7978:37`; unit `/workspace/family5-alt-c/src/compiler/utilities.ts:7976:1`; Refused / lowering.

Reason: an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type

```text
/workspace/family5-alt-c/src/compiler/utilities.ts:7978:37: Adamic 0.1 refuses an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type; keep compatible optional fields in both views, or construct an object with explicitly compatible fields (adamic/no-optional-widening)
```

## Growth outside G02 in candidate (c)

The nine primary new findings below are outside the original 14 G02 sites. They are new, not reached: the annotated relation still fails, so the G02 presence obligation has not been discharged. The full-event list above also retains their rollback Boundary mirrors. Coordinates below are in the scratch candidate after insertion.

| Site | Kind / phase | Full message |
|---|---|---|
| `/workspace/family5-alt-c/src/compiler/checker.ts:49350:111` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/checker.ts:49350:111: stage 0 can't lower reading start yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:742:9` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:742:9: stage 0 can't lower assigning to an Identifier yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:743:9` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:743:9: stage 0 can't lower assigning to an Identifier yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:747:20` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:747:20: stage 0 can't lower reading lineContent yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:757:24` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:757:24: stage 0 can't lower reading lineContent yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:758:24` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:758:24: stage 0 can't lower reading lineContent yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:761:24` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:761:24: stage 0 can't lower reading lineContent yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:765:24` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:765:24: stage 0 can't lower reading lineContent yet |
| `/workspace/family5-alt-c/src/compiler/program.ts:741:43` | NotYet / lowering | /workspace/family5-alt-c/src/compiler/program.ts:741:43: stage 0 can't lower reading lineStart yet |

## Every child exit

| Receipt | Real exit | Wall seconds |
|---|---:|---:|
| candidates/a/emitted-bundles | 0 | 0.267 |
| candidates/a/javascript | 0 | 1.72 |
| candidates/a/prepare | 0 | 0.768 |
| candidates/a/public-build | 1 | 25.098 |
| candidates/a/services-merge | 1 | 18.747 |
| candidates/a/stock | 0 | 15.151 |
| candidates/a/stock-strict | 0 | 15.077 |
| candidates/b/emitted-bundles | 0 | 0.517 |
| candidates/b/javascript | 0 | 1.839 |
| candidates/b/prepare | 0 | 0.467 |
| candidates/b/public-build | 1 | 27.043 |
| candidates/b/services-merge | 1 | 21.16 |
| candidates/b/stock | 0 | 15.717 |
| candidates/b/stock-strict | 0 | 15.588 |
| candidates/c/emitted-bundles | 0 | 0.217 |
| candidates/c/javascript | 0 | 9.45 |
| candidates/c/prepare | 0 | 3.286 |
| candidates/c/public-build | 0 | 28.1 |
| candidates/c/public-compare | 0 | 0.01 |
| candidates/c/services-merge | 1 | 18.59 |
| candidates/c/stock | 0 | 15.188 |
| candidates/c/stock-strict | 0 | 15.268 |
| comparisons/candidate-b | 0 | 1.216 |
| comparisons/candidate-c | 1 | 2.81 |
| comparisons/whole-entry | 1 | 1.126 |
| entry/base | 0 | 2537.57 |
| entry/c02 | 0 | 2648.544 |
| full/b/checker | 0 | 3532.619 |
| full/b/parser | 0 | 793.073 |
| full/b/program | 0 | 713.393 |
| full/b/utilities | 0 | 573.523 |
| full/base/checker | 0 | 2888.094 |
| full/base/parser | 0 | 706.065 |
| full/base/program | 0 | 713.382 |
| full/base/utilities | 0 | 652.015 |
| full/c/checker | 0 | 3401.747 |
| full/c/parser | 0 | 805.816 |
| full/c/program | 0 | 950.426 |
| full/c/utilities | 0 | 626.352 |
| node-controls/a | 1 | 21.899 |
| node-controls/b | 1 | 21.568 |
| node-controls/base | 0 | 24.72 |
| node-controls/c | 0 | 21.885 |
| publication/whole-entry-pair | 0 | 2.905 |
| setup/build-base | 0 | 8.06 |
| setup/build-c02 | 0 | 3.039 |
| setup/stock-services-baseline | 1 | 22.213 |

For (c), distinguish disappearance of the literal old reason from proof of the shape. The explicit satisfies relation uses `an unproven relation from SourceFile to SourceFileLike: optional field getPositionOfLineAndCharacter has no proven compatible presence/type`. This is the same unresolved S/T/P presence obligation. Such a replacement is graded changed, not reached. The summary G02-after column counts both diagnostic forms; the JSON also retains the literal-old-reason count separately. Stock TypeScript still reports SourceFile as the result type.

## Result

No candidate meets all requested checks. (a) is killed by TS2386; (c) keeps all 14 missing-property obligations under the explicit-relation diagnostic and adds nine primary findings outside G02. Distinguish the lowering result from source/API legality: (b) clears all 14 G02 sites with zero new, changed or reached findings in the four measured files, but its optional property conflicts with the services required method and the public build fails. This is not a landing-ready declaration. C02 does not qualify as an established-shape adaptation under this delegation; return the cause to Compiler for a different road. No landing was pushed.
