Both full runs completed with exit 0 under the 1500-second bound; no scan-only fallback was used. Each used the sys.ts file as the load root and LATENT_ONLY_FILE, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, CommonJSBundle, and the same adapted input. The input is checker-clean in both runs; the tool uses its standard measurement-only label. No usable compiler output was produced.

Findings: {"parent": 279, "head": 278, "unchanged": 278, "gone": 1, "new": 0, "changed": 0}. The unchanged count compares complete finding objects, including unit, phase, kind, location, reason and text. Repeated locations in different nested units are retained. Unambiguous single old/new findings with the same phase/unit/kind are shown as changed; other differences are new or gone.

| Commit | Refusal scan | Lowering | Units | Seconds |
| --- | ---: | ---: | ---: | ---: |
| parent | 19 | 260 | 136 | 55.586 |
| head | 18 | 260 | 136 | 69.62 |

Every new, gone, or changed finding follows. Text is exact except that the common adapted compiler directory prefix is omitted. Raw absolute text is retained in JSON and raw streams.

| Change | Phase | Kind | Unit | Before text | After text |
| --- | --- | --- | --- | --- | --- |
| gone | refusal_scan | Refused | sys.ts:157:9 | sys.ts:160:35: Adamic 0.1 refuses optional property Low in Partial<Levels> absent from structural source {}, which can hide fields; declare Low on the source type, or build a fresh object with known fields (adamic/no-optional-widening) | — |

All 260 lowering findings are unchanged. Both commits record the NotYet at sys.ts:160:17: `stage 0 can't lower assigning an element of a value yet`, including its boundary records in three nested units. This gap is not newly introduced by the head. The ordinary census never reaches it because the refusal scan stops first at sys.ts:272:23.

Adapted line 160, where full mode reaches the lowering gap:

```typescript
                (customLevels || (customLevels = {}))[level] = Number(customLevel);
```

