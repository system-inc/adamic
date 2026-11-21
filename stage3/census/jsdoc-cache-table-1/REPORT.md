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

## Follow-up with the later adapter name

The remote still named `041512efd4d4bd497ccfccf817c05df546cd58ca` as main when I restarted. The renamed directory is `99-correctness-tags-cache-table`. I audited the 22 adapters that sort after it, including their owner manifests and in-script pins. The matching parser entries in `99-method-value-views` pin four individual functions: `parseTag`, `parseJSDocLink`, `parseIntersectionTypeOrHigher`, and `parseUnionTypeOrHigher`. Their references to `currentNode` are calls inside those pins. The generated-name adapter locates particular `autoGenerate` accesses and import specifiers in `utilitiesPublic.ts`. The later pins inspected do not cover the edited declarations. The whole `Parser` pin now runs before the proposed edit.

## Stop at proof 1

I extracted the four affected files from the existing `refs/stage3/99-correctness-supported-extensions-after` snapshot. The adaptation scripts and apply recipe were byte-identical between that snapshot’s Adamic main and the fresh main named above. These commands exercised the draft on that input before any full-tree run.

| Command | Exit | Result |
| --- | ---: | --- |
| `NODE_PATH=/workspace/scratch/jsdoc-scout/adamic/stage3/api/node_modules node stage3/adapt/99-correctness-tags-cache-table/adapt.cjs /workspace/scratch/tags-cache-proof/stage` | 0 | First application wrote four files. |
| Same command, second invocation | 1 | `src/compiler/utilitiesPublic.ts: drift at getJSDocTagsWorker`. |
| Read-only TypeScript 6.0.3 AST probe of the adapted source | 0 | The table declaration exists with numeric kind 244. The enum reverse lookup names that value `FirstStatement`, and the locator returned no match for the requested `VariableStatement` selector. |

The second invocation failed the idempotence check. Comparing `node.kind` with the numeric enum value for the selector would address the observed locator error. The draft remains in the scratch worktree. The restore mutant and proofs 2–6 remain unrun. The ledger entry awaits passing proofs.
