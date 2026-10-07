# Wave B

32 reviewed indexed-expression occurrences: 17 required assertions, 0 numeric bitwise defaults, 15 declines.

Census before this wave is the tree with 10 and 30; previous waves do not edit these files. Options remain Adamic strictness. All 78 compiler roots are loaded.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 |
| --- | --- | --- | --- | --- | --- |
| transformers/ts.ts | 0 → 0 | 2 → 2 | 0 → 0 | 0 → 0 | 1 → 1 |
| transformers/es2015.ts | 14 → 1 | 11 → 0 | 7 → 4 | 0 → 0 | 0 → 0 |
| transformers/declarations.ts | 1 → 1 | 0 → 0 | 1 → 1 | 0 → 0 | 0 → 0 |

Total: 37 → 10. Every remaining finding follows.

- `transformers/declarations.ts:1465:40 TS2532`: The property-symbol list producer must prove a nonempty namespace list before selecting its parent; that invariant remains unreviewed.
- `transformers/declarations.ts:1678:110 TS2345`: The extends clause type-list producer must prove nonemptiness for synthetic and malformed input; that invariant remains unreviewed.
- `transformers/es2015.ts:2803:38 TS2532`: ContainsBindingPattern is not a local length guard; malformed or synthetic declaration-list nonemptiness has not been established.
- `transformers/es2015.ts:3046:64 TS2532`: flattenDestructuringBinding can expand a pattern; nonempty output for every binding pattern has not been established.
- `transformers/es2015.ts:4437:29 TS2345`: The cross-transformer class constructor position after an optional extends call needs a producer-contract review before assertion.
- `transformers/es2015.ts:4691:34 TS2532`: No local length guard precedes this first-segment read; the nonempty spanMap/flatten producer contract is deferred.
- `transformers/es2015.ts:4693:13 TS2532`: No local length guard precedes this first-segment read; the nonempty spanMap/flatten producer contract is deferred.
- `transformers/ts.ts:1351:45 TS2538`: The super-call path producer must prove this depth and the selected statement index; that cross-function invariant remains unreviewed.
- `transformers/ts.ts:1352:97 TS18048`: The super-call path producer must prove this depth and the selected statement index; that cross-function invariant remains unreviewed.
- `transformers/ts.ts:1379:80 TS18048`: superStatementIndex comes from the declined superPath[superPathDepth] read; its valid path depth remains unproven.

The occurrence ledger is in [sites.json](../../sites.json). Survey coordinates document the original source; addressing uses parsed expressions and occurrence counts.

Commands (stdout/stderr saved directly to logs):

```sh
node adapt.cjs TREE
bash census.sh TREE OUTPUT
node verify.cjs BEFORE TREE
node adapt.cjs TREE # second run: zero edits
MUTANT_FILE=transformers/es2015.ts node mutant.cjs BEFORE TREE NEW_MUTANT_TREE
stage3/oracle/run.sh TREE OUTPUT # default suites, adaptations 10 + 30 + 33, no 20
```

The mutant changes transformers/es2015.ts outParams[i]! to (outParams[i] ?? 0). Both emitted-JavaScript equality and the site contract reject it with exit 1. No mutant oracle was run. CRLF is preserved because edits insert tokens without rewriting lines.

Default oracle: pass; {'passing': 106367, 'failing': 0, 'pending': 0}; 0 baseline differences; 223.871 seconds. See oracle.json and baseline.diff.
