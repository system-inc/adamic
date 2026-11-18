# C02 whole-entry first-error reach

Base `fbe5b72ba115a8f3060d4b5229fd578574f4505f`; C02 `12b9e04184663f899acb76505043ad292966f236`. No landing push. Each side used its own rebuilt entry tool, the pinned v6.0.3 adapted tree, and ran alone. LATENT_FULL=0, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TAIL=1, CommonJSBundle, tsc.ts entry, no LATENT_ONLY_FILE. Each raw stream was pushed before the next side started.

## Counts

C02 phrase: `a checked view with callable field getPositionOfLineAndCharacter`. C02 has 1 matching events, 1 distinct positions including Boundary mirrors, and 1 distinct primary diagnostic positions. Baseline has 0 matching events at 0 distinct primary positions.

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

| Run | Real exit | Wall seconds | Stream commit |
|---|---:|---:|---|
| base | 0 | 2537.57 | `2c8ef70a7469d03a02d2fee51fb713ff9220c02c` |
| c02 | 0 | 2648.544 | `0ea786e0eea609ee629911446db28b376f0237a7` |

Pairs use kind, phase, reason and full message with positions removed and adapted roots normalized. Equal keys pair by unit/site first. A changed primary first-error stop can move inside the same unit. Reached requires an old G02 stop that is cleared in that unit; an unrelated Boundary or a remaining G02 is insufficient. Full raw records and before/after messages are retained.

## Every C02 site naming the callable field

| Site | Unit | Kind | Phase |
|---|---|---|---|
| `/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:11` | `/workspace/family5-c02-tree/src/compiler/utilities.ts:5317:1` | Refused | lowering |

## Every difference, with full words


### 1. changed

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:5339:11`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:5317:1`; NotYet / lowering.

Reason: checked view field modifiers of type NodeArray<Modifier> | NodeArray<ModifierLike> | undefined

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:5339:11: stage 0 can't lower checked view field modifiers of type NodeArray<Modifier> | NodeArray<ModifierLike> | undefined yet
```
after: `/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:11`; unit `/workspace/family5-c02-tree/src/compiler/utilities.ts:5317:1`; Refused / lowering.

Reason: a checked view with callable field getPositionOfLineAndCharacter

```text
/workspace/family5-c02-tree/src/compiler/utilities.ts:5339:11: Adamic 0.1 refuses a checked view with callable field getPositionOfLineAndCharacter; prove the callable body rather than asserting its signature
```

### 2. gone

before: `/tmp/family5-base-tree/src/compiler/checker.ts:49348:52`; unit `/tmp/family5-base-tree/src/compiler/checker.ts:1490:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/checker.ts:49348:52: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 3. gone

before: `/tmp/family5-base-tree/src/compiler/parser.ts:10015:36`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9929:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10015:36: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 4. gone

before: `/tmp/family5-base-tree/src/compiler/parser.ts:10015:48`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9929:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10015:48: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 5. gone

before: `/tmp/family5-base-tree/src/compiler/parser.ts:10087:37`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9929:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10087:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 6. gone

before: `/tmp/family5-base-tree/src/compiler/parser.ts:10244:42`; unit `/tmp/family5-base-tree/src/compiler/parser.ts:9929:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/parser.ts:10244:42: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 7. gone

before: `/tmp/family5-base-tree/src/compiler/program.ts:738:57`; unit `/tmp/family5-base-tree/src/compiler/program.ts:717:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/program.ts:738:57: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 8. gone

before: `/tmp/family5-base-tree/src/compiler/program.ts:739:76`; unit `/tmp/family5-base-tree/src/compiler/program.ts:717:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/program.ts:739:76: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 9. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:1261:24`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:1240:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:1261:24: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 10. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:2532:70`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:2524:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:2532:70: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 11. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7945:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7943:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7945:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 12. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7950:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7949:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7950:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 13. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7960:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7959:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7960:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 14. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7972:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7969:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7972:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

### 15. gone

before: `/tmp/family5-base-tree/src/compiler/utilities.ts:7978:37`; unit `/tmp/family5-base-tree/src/compiler/utilities.ts:7976:1`; Refused / refusal_scan.

Reason: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields

```text
/tmp/family5-base-tree/src/compiler/utilities.ts:7978:37: Adamic 0.1 refuses optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields; declare getPositionOfLineAndCharacter on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```
