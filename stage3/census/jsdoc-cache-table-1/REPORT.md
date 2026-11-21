# JSDoc cache table owner audit · #3qh3m28

Fresh `origin/main` was `041512efd4d4bd497ccfccf817c05df546cd58ca`, committed 2026-11-21 at 09:31:59 UTC−07:00 (16:31:59 UTC). The commit is “oracle, stage3: fourteen programs that stop where Node runs on, and that no hatch they write excuses, await #00e23th in a pending sidecar, as Adamic ruled (#d358hc1); p67 and field_empty_interface already have theirs on main (da998141)”. The upstream TypeScript pin is v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Stop at the existing owner

`99-correctness-receiver-parser` sorts after `99-correctness-jsdoc-cache-table`. Its first owner pin covers the entire `Parser` module declaration in `src/compiler/parser.ts`. The incremental invalidation is a descendant of that owner, inside `Parser.currentNode`. Editing that function in the proposed earlier adapter would change the text hashed by the later adapter and fail its owner check. The parser edit therefore belongs in an amendment to `99-correctness-receiver-parser`. [Owner pins](../../adapt/99-correctness-receiver-parser/owners.json) · [Hash check](../../adapt/99-correctness-receiver-parser/adapt.cjs)

I searched `stage3/adapt` for `getJSDocTagsWorker`, `currentNode`, `filterOwnedJSDocTags`, `JSDocContainer`, `JSDocArray`, and `jsDocCache`. The follow-up parsed both snapshots of the whole `Parser` owner with TypeScript 6.0.3, resolved its module and child function by AST kind and name, and found the assignment to `jsDocCache` in that function in both snapshots. It also recomputed both whole-owner SHA-256 digests. The AST and digest check exited 0.

| Snapshot | Whole-owner SHA-256 |
| --- | --- |
| Before | `a269b1e9c4aa24067795a33f7f606eb775d97d131d87a90b3050feb82c127742` |
| After | `89b3a88005f8a53e63fbb7f40b211db91e24a2715c849760bd5e07c5700ffd7a` |

The owner gate stopped the task before implementation. Proofs 1–6 require a landed candidate tree and remain unrun. The ledger entry depends on those results.
