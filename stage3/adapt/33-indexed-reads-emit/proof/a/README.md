# Wave A

62 reviewed indexed-expression occurrences: 25 required assertions, 0 numeric bitwise defaults, 37 declines.

Census before this wave is the tree with 10 and 30; previous waves do not edit these files. Options remain Adamic strictness. All 78 compiler roots are loaded.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 |
| --- | --- | --- | --- | --- | --- |
| emitter.ts | 21 → 3 | 6 → 2 | 3 → 0 | 1 → 1 | 0 → 0 |
| sourcemap.ts | 4 → 1 | 0 → 0 | 1 → 0 | 2 → 1 | 0 → 0 |

Total: 38 → 8. Every remaining finding follows.

- `emitter.ts:1138:5 TS2322`: JSON.stringify has an optional result in the built-in model; not an indexed read.
- `emitter.ts:4111:27 TS2345`: The for-of split result has optional elements in the built-in model; not an indexed read.
- `emitter.ts:4977:46 TS2345`: The split result array has optional elements in the built-in model; not an indexed read.
- `emitter.ts:4979:40 TS18048`: lineText comes from for-of over the modeled split array; not an indexed read.
- `emitter.ts:4980:17 TS18048`: line inherits the optional split element; not an indexed read.
- `emitter.ts:4982:23 TS2345`: write receives the optional split element; not an indexed read.
- `sourcemap.ts:216:58 TS2345`: The adjacent typeof string test handles absence; repeated read is declined without hoisting.
- `sourcemap.ts:83:25 TS2322`: JSON.stringify has an optional result in the built-in model; not an indexed read.

The occurrence ledger is in [sites.json](../../sites.json). Survey coordinates document the original source; addressing uses parsed expressions and occurrence counts.

Commands (stdout/stderr saved directly to logs):

```sh
node adapt.cjs TREE
bash census.sh TREE OUTPUT
node verify.cjs BEFORE TREE
node adapt.cjs TREE # second run: zero edits
node mutant.cjs BEFORE TREE NEW_MUTANT_TREE
stage3/oracle/run.sh TREE OUTPUT # default suites, adaptations 10 + 30 + 33, no 20
```

The mutant changes emitter.ts transform.transformed[0]! to (transform.transformed[0] ?? 0). Both emitted-JavaScript equality and the site contract reject it with exit 1. No mutant oracle was run. CRLF is preserved because edits insert tokens without rewriting lines.

Default oracle: pass; {'passing': 106367, 'failing': 0, 'pending': 0}; 0 baseline differences; 347.566 seconds. See oracle.json and baseline.diff.
