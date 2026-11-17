# getProperties overload result (#ev14ch9)

Base: origin/main 932a6344af63dc7c8728fffac05584c176d0475f. TypeScript v6.0.3 pin: 050880ce59e30b356b686bd3144efe24f875ebc8.

Widen only the true overload result from readonly InitializedPropertyDeclaration[] to readonly PropertyDeclaration[]. Its one true caller tests array nonemptiness with some; the other caller passes false and already consumes ordinary properties. No caller needs the initialized-element promise. Parameters, body, returned array identity and runtime values are preserved. This is type-only; both commented and comment-free JavaScript are asserted byte-identical before any write.

The first ownership search found only an esnext.ts import in 99-correctness-receiver-transformers. No existing adapter pins getProperties whole, so this is a new adaptation. owners.json pins the complete three-declaration overload owner before and after. The stock 6.0.3 parser checks it; drift, duplicate owners, parse errors, half-adapted text and mixed line endings are refused. CRLF is preserved, and a replay writes nothing. Alphabetical apply order introduces no owner overlap. mutant.cjs restores the exact original owner for the red proof.

Compiler accepted this candidate under the #ev14ch9 delegation after the original 0-new gate held. The pre-existing walls at transformers/utilities.ts:590:12 (return cast) and binder.ts:2258:13 (checked field write) are newly reached. The edit creates no cast or write. The full latent census on this base is:

| File | Before | After | Newly reached events | Gone events | Changed events |
| --- | ---: | ---: | ---: | ---: | ---: |
| transformers/utilities.ts | 146 | 146 | 6 | 6 | 0 |
| transformers/ts.ts | 658 | 654 | 0 | 4 | 0 |
| transformers/classFields.ts | 633 | 624 | 2 | 11 | 1 |

Counts include Boundary bookkeeping. The owner has three declaration-unit replays of the reached cast. classFields.ts reaches the binder wall and a Boundary whose underlying stop becomes the existing core.ts filter overload. Full before/after JSONL streams and every event delta are on the evidence branch. Each run uses LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, the selected absolute LATENT_ONLY_FILE and CommonJSBundle; every stream finishes normally.

Proof receipts for the whole apply, TestProduct_Stage3AdaptedTree stock build, unfiltered whole lane, restoring mutant red and byte-identical public typescript.d.ts are retained with the [new-base evidence report](https://github.com/system-inc/adamic/blob/typescript/overload-transformers-evidence/stage3/census/overload-transformers/landing-932a6344/REPORT.md). The shard runtime/API inputs are unchanged, so shard-plan.json, shards.json and shards_generated_test.go are preserved. No measurement dump or Go file belongs in this adaptation.
