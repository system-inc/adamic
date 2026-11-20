# Paired stage3 census for #ad161ms

Interim report. The head entry stream is complete. The base entry, compiler-directory, and 79-root comparisons are in progress.

## 1. New findings under any reason

The head entry pass has 10 refusal-scan findings on lines introduced by the adaptation. The zero-new criterion fails at these sites. The complete paired count and the compiler-directory count are pending.

- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1302:38` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1302:38: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1302:45` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1302:45: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1303:102` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1303:102: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1303:109` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1303:109: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1939:29` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1939:29: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1940:82` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1940:82: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/esDecorators.ts:298:1`, finding at `src/compiler/transformers/esDecorators.ts:1941:44` (Refused, refusal_scan). src/compiler/transformers/esDecorators.ts:1941:44: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/namedEvaluation.ts:69:1`, finding at `src/compiler/transformers/namedEvaluation.ts:81:25` (Refused, refusal_scan). src/compiler/transformers/namedEvaluation.ts:81:25: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/namedEvaluation.ts:69:1`, finding at `src/compiler/transformers/namedEvaluation.ts:82:76` (Refused, refusal_scan). src/compiler/transformers/namedEvaluation.ts:82:76: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)
- `src/compiler/transformers/namedEvaluation.ts:69:1`, finding at `src/compiler/transformers/namedEvaluation.ts:83:32` (Refused, refusal_scan). src/compiler/transformers/namedEvaluation.ts:83:32: Adamic 0.1 refuses an object refinement using an open numeric enum as a literal tag; use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)

## 2. Gone and changed findings

The head entry pass has 740 distinct family-1 findings. The paired gone and changed inventory and the remaining family-1 counts are pending.

## 3. Units that lowered at base and stop at head, and new panics

Pending the paired latent streams.

## 4. First stop by root

At this checkpoint, 0 first stops moved across 5 completed root pairs out of 79.

## Inputs and evidence

Base `b8f6109bf1bc284de5b3dd0b446c26b2053cadfd`. Head `bc7e1a3940326bfc6b67f6068147e9ef6e1fc76a`. Both were fetched by branch name. Both used their own `stage3/apply.sh` with TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8` after recursive submodule initialization.

The source trees differ only in `src/compiler/utilities.ts`, `src/compiler/transformers/esDecorators.ts`, and `src/compiler/transformers/namedEvaluation.ts`. The generated patch summary also differs. All three tools were built from base.

Both latent passes use first-error mode with CommonJSBundle and the scoped tail. [Head entry stream](evidence/head-entry.jsonl.gz) and [adapted tree comparison](evidence/tree-diff.txt).
