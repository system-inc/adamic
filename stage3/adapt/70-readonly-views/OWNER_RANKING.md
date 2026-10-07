# Ranked readonly-owner declines, wave 6

This wave makes no source edit. The scores below come from stock-checker declaration identity on the 349-site wave 5 input. Context coverage is a prioritization score, not a claim of net refusals cleared; a readonly slot can expose a deeper refusal, as earlier waves proved. Overlapping field scores are not added.

The complete field ranking, direct writes and unresolved structural paths are in evidence/wave6/field-ranking.json. The complete checker-context inventory is in owners.json.gz. The field audit is a shallow replacement survey; zero direct writes alone never authorizes an edit. It does not prove away erased consumers, accessor effects, callbacks or container-element mutation.

| Highest groups considered | Coverage | Decision and next missing proof |
|---|---:|---|
| DiagnosticRelatedInformation file/start/length family | 26 shared contexts | No direct base-slot writes, but 30/31/35 structural or erased paths respectively. Derived file/start/length declarations and erased receiver storage need a coupled proof; no readonly edit. |
| SourceFileLike | 24 exact wider-target reasons; 25 lineMap contexts | scanner.ts:504:35 writes number[] into lineMap. Actual write fits SourceFile's original readonly number[] slot, so no wrong-type insertion; wider replacement capacity is latent. Retained. |
| decodeMappings.next / IteratorResult<Mapping, any> | 14 reasons | Contextual return contract comes from the stock iterator protocol. A local readonly return annotation also needs proof of the protocol consumers and wider aliases; no such proof completed. Do not edit the stock npm library. |
| addRelatedInfo rest elements | 13 reasons | utilities.ts:10357:5 stores the references in a mutable diagnostic array. Array readonly alone does not certify element-field views. Depends on the diagnostic-family proof. |
| Inferred Map<string, string[]> / never[] unions | 13 reasons | Map element and retained-array ownership is not certified by shallow readonly fields. No single explicit declaration was established for every first-context view. Retained rather than relabeling a generic or inferred contract. |
| mutateMapSkippingNewValues.map | 9 reasons | utilities.ts:8220:13 deletes entries. This write inserts no wrong-typed value; replacement capacity is latent. The view cannot be ReadonlyMap while its consumer deletes. |
| multiMapSparseArrayAdd.map | 4 reasons | transformers/utilities.ts:380:9 pushes and 383:9 assigns the array slot. Wrong-typed reachability into a particular original was not established; retained pending that proof. |

Diagnostic-family paths inspected through AST and checker records:

- utilities.ts:2486:16 casts to DiagnosticMessageChain to read next; 8638:20 casts to DiagnosticWithDetachedLocation. These are field-identity changes, not observed file writes.
- checker.ts:3617:37, 3620:37, 3623:37, 3628:41, 35099:33 and 35107:33 store results in inferred diagnosticMessage locals with contextual any. SearchResult's bounded alias proof cannot certify them: they forward the whole receiver to diagnostic helpers.
- services/refactors/extractSymbol.ts:1869:33 pushes a diagnostic into an erased array, retaining the receiver.
- builder.ts:1523:16 exposes a reusable serialized diagnostic view. category-only consumers and TextSpan projections also lose the file slot from their declared view. Those paths are unproven, not declared corrupt.
- attachFileToDiagnostic constructs a fresh DiagnosticWithLocation; its object-literal file initializer is not a write through the original detached view. This observation does not settle the remaining aliases.

These declines report what is known without inventing reachability. The exact setTextRange fixture, observed [1,2], and the writing destination candidate remain unchanged. The new direct-file-writer mutant changes the survey from zero writes to one; the real-writing SourceFileLike readonly mutant is caught by upstream TS2540. Complete census, API attribution, JS hashes, idempotence and default oracle are recorded beside this report.

For the next adaptation attempt, the diagnostic family is the highest unresolved shared view. Its casts, inferred locals, erased array and reusable view must be audited together before editing any public slots.
