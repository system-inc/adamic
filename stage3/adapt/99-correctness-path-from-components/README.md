# 99-correctness-path-from-components

Type-only adaptation of TypeScript 6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`), final validation on Adamic main `02ebf2b15`. Replace the internal generic `getPathFromPathComponents<T extends string>` with Path/string overloads and a string implementation. Remove only its two `as T` assertions. Each overload is `@internal`, and both precede the existing `//// Path Formatting` comment and JSDoc so emitted comments remain byte-identical. No runtime algorithm changes.

The whole-checkout stock checker measurement finds 15 source calls: 13 string and 2 Path, including five harness alias calls. Neither public declaration inventory contains this function. Main already serves the primitive phantom Path overload result. This removes the generic casts rather than weakening a runtime cast check.

`owners.json` pins the entire before/after top-level owner, including its section comment, documentation and complete overload set, by exact text and SHA-256. The adapter refuses a changed, missing, moved, duplicate or partly adapted owner, and parses and verifies the complete planned file before writing. Unrelated owners are preserved. It rereads the input before its sole write. CRLF is preserved. Replay returns zero files/edits, performs zero writes and preserves bytes and mtime. `contract.cjs` exercises these properties with restoring scratch trees. The adapter also checks complete path.ts transpiled JavaScript equality before writing.

Adaptation 30's declined `pathComponents[0]` read remains exactly as written. Its ledger's owner name, read occurrences and total remain intact. Adapter 30's isolated path.ts check accepts both owner forms with zero edits. Replaying 30 over a fully adapted main tree has a pre-existing addRange refusal in core.ts; that is not apply's path, because apply runs each adapter on its predecessor's output. The replay proofs here are full stage-by-stage apply and this adapter's zero-edit replay over its own output. The existing pin search and historical refusal are retained in the evidence branch.

`forEachAncestorDirectory` is unchanged. Its existing overloads resolve five calls to string (3) or Path (2); one string-overload call passes a generic P argument. Its implementation-only P is not a caller-visible type parameter. Replacing that implementation with a string directory/string callback and removing `as P` produces stock TS2394: the Path callback overload is incompatible with that implementation. Although JavaScript is unchanged, this fails the required zero-diagnostic gate. No broader callback change is approved here.

Measurement output and detailed commands/results belong on [typescript/path-from-components-evidence](https://github.com/system-inc/adamic/tree/typescript/path-from-components-evidence/stage3/census/path-from-components-adapt), outside this hashed adaptation directory. The earlier census is on [typescript/path-from-components](https://github.com/system-inc/adamic/tree/typescript/path-from-components/stage3/census/path-from-components). The evidence records the ten build outputs, both public declarations, stock and Adamic 79-root checks, full apply/replay, exact before/after census changes, restoring mutant, shard generation check and the whole 398-shard Go lane. Consult its REPORT.md for measured exits and any blocker; this README does not substitute for those results.

Reproduce with the pinned stock parser on NODE_PATH:

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-path-from-components/contract.cjs BEFORE_TREE
NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-path-from-components/adapt.cjs TREE
```

Keep every measurement output outside this directory: all its bytes participate in apply's product key. No compiler, oracle, test expectation, public API sanction, or other adapter is changed.
