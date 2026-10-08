Adapted two private JSON diagnostic sinks to the push operation they actually perform.
This supersedes the generic diagnostic edits in 487f0d6c, whose narrower aliases escaped into writable storage.
Only the sink edits receive final credit; their complete census delta is nine sites, with no runtime change.
Undoing convertConfigFileToObject's sink restores two rows; adding errors.push(undefined) is rejected before source writes.
Related-information and collection storage views remain unadapted; generic-storage.a demonstrates why capture alone is insufficient.

convertConfigFileToObject and convertToJson only append values produced by createDiagnosticForNodeInSourceFile. Their private errors parameter is Pick<DiagnosticWithLocation[], "push">. This exposes neither a writable element index nor a push accepting a diagnostic without a file. The caller's location-diagnostic array can hold every appended value. A broader Diagnostic[] also provides that append operation. Both reviewed bodies are pinned, including nested callbacks. The public convertToObject signature remains unchanged.

These edits remove two DiagnosticWithLocation[] -> Diagnostic[] caller views and seven DiagnosticWithLocation -> Diagnostic append-argument views inside those helpers. The latter become exact location-diagnostic arguments rather than wider field views. They introduce no extra factory allocation, copy or branch. diagnostics.json contains only the two sink edits and their type import; it adds no any or unknown keyword to source.

Withdrawn generic edits

The earlier 75-row diagnostic experiment also changed addRelatedInfo, createDiagnosticMessageChainFromDiagnostic, addErrorOrSuggestion, invocationErrorRecovery and addTypeOnlyDeclarationRelatedInfo to captured generic inputs. The final source restores all five signatures. Only nine sink removals survive; 66 removals from the generic experiment are withdrawn.

In checker.ts:2561, addErrorOrSuggestion still forwards the original object to DiagnosticCollection.add. In utilities.ts:10357, addRelatedInfo still pushes the related-information references into a mutable DiagnosticRelatedInformation[] slot. Even a readonly rest array of captured R values does not stop those references from being exposed through the broader writable elements. A runtime-body fingerprint proves that no new algorithm was added; it does not prove downstream aliases readonly. The earlier claim combined those different propositions incorrectly.

language-questions/generic-storage.a is the minimal counterexample to that workaround. A generic helper appends a DiagnosticWithLocation into Diagnostic[]; a subsequent wider view writes undefined, and the original holder still promises SourceFile. Stock TypeScript reports zero diagnostics and Node prints 1 when the promise is violated. Its mutant restores the original file instead; Node prints 0 and the witness assertion fails with exit 1. This is a counterexample to the proposed adaptation, not evidence that tsc actually performs the later undefined assignment. It does not reclassify the supplied table row as a language question.

The historical census, oracle and lane output in evidence/diagnostics/ is retained as rejected experiment evidence. Its observed counts are reproducible, but its 75 removals are not the final proven adaptation count. The collection ReturnType experiment was also rejected because its exported factory fails stock TypeScript's isolatedDeclarations requirement. Neither experiment is included in the final source.

Remaining diagnostic views

The final exact-table scalar DiagnosticWithLocation -> Diagnostic reason is 93 -> 86; DiagnosticWithLocation -> DiagnosticRelatedInformation stays 58 -> 58. The location-array -> diagnostic-array reason is 9 -> 7. Detached-location -> related-information remains five. Readonly array containers can still expose writable diagnostic fields, so merely adding readonly to an array parameter is insufficient.

The retained sites cross mutable arrays, collections and public Diagnostic/DiagnosticRelatedInformation contracts. To protect the narrower holder, a type-only approach must propagate readonly element fields through those storage contracts and their aliases. That audit is incomplete here, and updating the lane's API sanction is outside the unit's territory. A copy would be a runtime edit if the existing writable API must be retained. No copy was made. This conditional API limitation is distinct from proof of a genuine writer in tsc.

All retained current sites are in REMAINING.md and evidence/reconciliation.json. The unit leaves these coupled storage views alone and does not claim that a lack of proof establishes a language question. The original position-setter witnesses remain the actual writer examples.
