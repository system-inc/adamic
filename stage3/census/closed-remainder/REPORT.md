# Closed-remainder paired census — w3e1vr8

**For the 48 canonical findings: 37 retained units change first stop — 3 lower outright, 0 stop at kind-through-Node, and 34 stop elsewhere. No newly stopped units or new panics occur.**

Head fetched and measured: `efd9bff92ce62bc92463917c6b72882adf32280e`. Its own main base is `4bba06cb2dc4a0046c11fc97e9953ec85ed0e638` (merge base with fetched main `3007955b6db96ae5f65cedb2e680d77ad1b88664`). Merging head onto its own base yields `9d9feff4b2e8d37287de934ed72a06557fdb98b1`, exactly head's tree: **no difference**. Head changes only `internal/lower/predicates_checked.go` and its remainder tests; **no adaptation changes**.

## 1. The exact supplied list

The source is **`typescript/overload-proof-census` at `481552841c9b5e1496fc6237da10501b707178d0`**, files `stage3/census/overload-phantom-78933e52/head/shard-0.jsonl.gz` through `shard-7.jsonl.gz`. These exact compressed files are retained under [evidence/authoritative-inventory/](evidence/authoritative-inventory/). No daily or standalone-78933e52 rerun supplies this list.

The archived `merge.py` restores shard order. Findings are filtered by the exact reason `checked predicate membership over an open tag domain`, then deduplicated using **exactly pair's distinct key**: `((source path, source SHA-256, outermost contiguous access-chain UTF-8 byte start), exact reason)`. The first representative in merged order is retained, including its unit. Unit identity uses the unnormalized byte offset, as pair does. The unmodified pair comparator independently confirms 48 raw-key and 48 distinct-key findings. The normalizer and pair code at the source and measured base are byte-identical; all relevant current adapted source hashes match the archived manifest.

**Observed: 50 raw occurrences in 50 owning attempts become 48 raw-key findings and 48 distinct-key findings in 48 retained units. All 48 retained findings are their unit's first lowering/deferred stop. I do not reproduce 28 units or 12 first stops. There is no dedup-key difference from pair.** The two removed representatives are duplicate reports of `utilities.ts:4229:13` under overload owners `utilities.ts:4224:1` and `utilities.ts:4226:1`; pair retains `utilities.ts:4222:1`. My previous report incorrectly collected every raw owner and used a substitute inventory. This section replaces that reconstruction.

The reproducible extractor is [authoritative-inventory.py](authoritative-inventory.py), and [evidence/authoritative-inventory.json](evidence/authoritative-inventory.json) preserves each exact key, retained finding and unit, historical first stop, and all raw occurrences.

Each row below uses that exact retained historical site and unit. Base/head first stops follow **pair's rule: lowering and deferred phases only**. Ordinary file-scan findings do not define pair unit first stops; their first finding is shown separately. `unchanged (censored)` retains the same other body stop and does not claim acceptance of a hidden site. All `lowers` rows complete all 26 scoped passes and also have no ordinary scan finding. Coordinates are relative to the shared adapted tree.

| ID | Historical site | Exact retained unit | Base first stop | Head classification and first stop | Ordinary scan first stop (same both sides) |
| --- | --- | --- | --- | --- | --- |
| 1 | `src/compiler/emitter.ts:836:10` | `src/compiler/emitter.ts:752:1` (emitFiles) | R0 at `src/compiler/emitter.ts:836:10` | **other stop**: R1 at `src/compiler/factory/nodeFactory.ts:6111:21` | R17 at `src/compiler/emitter.ts:864:28` |
| 2 | `src/compiler/factory/nodeConverters.ts:55:13` | `src/compiler/factory/nodeConverters.ts:41:1` (createNodeConverters) | R0 at `src/compiler/factory/nodeConverters.ts:55:13` | **unchanged**: R0 at `src/compiler/factory/nodeConverters.ts:55:13` | R9 at `src/compiler/factory/nodeConverters.ts:66:40` |
| 3 | `src/compiler/factory/nodeFactory.ts:7287:20` | `src/compiler/factory/nodeFactory.ts:7286:1` (propagateNameFlags) | R0 at `src/compiler/factory/nodeFactory.ts:7287:20` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R7 at `src/compiler/factory/nodeFactory.ts:7287:98` |
| 4 | `src/compiler/factory/utilities.ts:189:9` | `src/compiler/factory/utilities.ts:188:1` (createMemberAccessForPropertyName) | R0 at `src/compiler/factory/utilities.ts:189:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R10 at `src/compiler/factory/utilities.ts:190:29` |
| 5 | `src/compiler/factory/utilities.ts:216:9` | `src/compiler/factory/utilities.ts:215:1` (createJsxFactoryExpressionFromEntityName) | R0 at `src/compiler/factory/utilities.ts:216:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R11 at `src/compiler/factory/utilities.ts:218:23` |
| 6 | `src/compiler/factory/utilities.ts:310:9` | `src/compiler/factory/utilities.ts:309:1` (createForOfBindingStatement) | R0 at `src/compiler/factory/utilities.ts:310:9` | **unchanged**: R0 at `src/compiler/factory/utilities.ts:310:9` | R7 at `src/compiler/factory/utilities.ts:311:40` |
| 7 | `src/compiler/factory/utilities.ts:335:9` | `src/compiler/factory/utilities.ts:334:1` (createExpressionFromEntityName) | R0 at `src/compiler/factory/utilities.ts:335:9` | **unchanged**: R0 at `src/compiler/factory/utilities.ts:335:9` | R7 at `src/compiler/factory/utilities.ts:336:62` |
| 8 | `src/compiler/factory/utilities.ts:349:9` | `src/compiler/factory/utilities.ts:348:1` (createExpressionForPropertyName) | R0 at `src/compiler/factory/utilities.ts:349:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R12 at `src/compiler/factory/utilities.ts:354:39` |
| 9 | `src/compiler/factory/utilities.ts:466:26` | `src/compiler/factory/utilities.ts:465:1` (createExpressionForObjectLiteralElementLike) | R13 at `src/compiler/factory/utilities.ts:362:107` | **unchanged (censored)**: R13 at `src/compiler/factory/utilities.ts:362:107` | R14 at `src/compiler/factory/utilities.ts:472:85` |
| 10 | `src/compiler/factory/utilities.ts:1024:21` | `src/compiler/factory/utilities.ts:1015:1` (tryGetPropertyNameOfBindingOrAssignmentElement) | R0 at `src/compiler/factory/utilities.ts:1024:21` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R7 at `src/compiler/factory/utilities.ts:1056:20` |
| 11 | `src/compiler/factory/utilities.ts:1096:17` | `src/compiler/factory/utilities.ts:1092:1` (getJSDocTypeAliasName) | R0 at `src/compiler/factory/utilities.ts:1096:17` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 12 | `src/compiler/factory/utilities.ts:1573:9` | `src/compiler/factory/utilities.ts:1570:1` (formatIdentifierWorker) | R0 at `src/compiler/factory/utilities.ts:1573:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 13 | `src/compiler/factory/utilities.ts:1746:27` | `src/compiler/factory/utilities.ts:1739:1` (containsObjectRestOrSpread) | no stop; all 26 scoped passes complete | **lowers**: no stop; all 26 scoped passes complete | none |
| 14 | `src/compiler/program.ts:886:9` | `src/compiler/program.ts:885:1` (isExclusivelyTypeOnlyImportOrExport) | R0 at `src/compiler/program.ts:886:9` | **other stop**: R18 at `src/compiler/program.ts:889:9` | none |
| 15 | `src/compiler/transformer.ts:203:20` | `src/compiler/transformer.ts:202:1` (wrapCustomTransformer) | R0 at `src/compiler/transformer.ts:203:20` | **other stop**: R1 at `src/compiler/factory/nodeFactory.ts:6111:21` | none |
| 16 | `src/compiler/transformers/destructuring.ts:202:14` | `src/compiler/transformers/destructuring.ts:197:1` (bindingOrAssignmentElementAssignsToName) | R1 at `src/compiler/binder.ts:1109:17` | **unchanged (censored)**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 17 | `src/compiler/transformers/destructuring.ts:524:20` | `src/compiler/transformers/destructuring.ts:522:1` (isSimpleBindingOrAssignmentElement) | R0 at `src/compiler/transformers/destructuring.ts:524:20` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 18 | `src/compiler/transformers/destructuring.ts:558:9` | `src/compiler/transformers/destructuring.ts:556:1` (createDestructuringPropertyAccess) | R0 at `src/compiler/transformers/destructuring.ts:558:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R16 at `src/compiler/transformers/destructuring.ts:559:214` |
| 19 | `src/compiler/transformers/namedEvaluation.ts:74:9` | `src/compiler/transformers/namedEvaluation.ts:67:1` (getAssignedNameOfPropertyName) | R0 at `src/compiler/transformers/namedEvaluation.ts:74:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R7 at `src/compiler/transformers/namedEvaluation.ts:74:60` |
| 20 | `src/compiler/transformers/utilities.ts:356:9` | `src/compiler/transformers/utilities.ts:355:1` (collectExportedVariableInfo) | R0 at `src/compiler/transformers/utilities.ts:356:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R15 at `src/compiler/transformers/utilities.ts:367:36` |
| 21 | `src/compiler/utilities.ts:2398:17` | `src/compiler/utilities.ts:2388:1` (entityNameToString) | R0 at `src/compiler/utilities.ts:2398:17` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 22 | `src/compiler/utilities.ts:4177:9` | `src/compiler/utilities.ts:4176:1` (getAssignmentDeclarationKindWorker) | R0 at `src/compiler/utilities.ts:4177:9` | **other stop**: R5 at `src/compiler/utilities.ts:4117:1` | none |
| 23 | `src/compiler/utilities.ts:4211:9` | `src/compiler/utilities.ts:4210:1` (getElementOrPropertyAccessArgumentExpressionOrName) | R0 at `src/compiler/utilities.ts:4211:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 24 | `src/compiler/utilities.ts:4229:13` | `src/compiler/utilities.ts:4222:1` (getElementOrPropertyAccessName) | R0 at `src/compiler/utilities.ts:4229:13` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 25 | `src/compiler/utilities.ts:4300:11` | `src/compiler/utilities.ts:4297:1` (isSpecialPropertyDeclaration) | R0 at `src/compiler/utilities.ts:4300:11` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 26 | `src/compiler/utilities.ts:4752:10` | `src/compiler/utilities.ts:4748:1` (getParameterSymbolFromJSDoc) | R1 at `src/compiler/binder.ts:1109:17` | **unchanged (censored)**: R1 at `src/compiler/binder.ts:1109:17` | R6 at `src/compiler/utilities.ts:4761:12` |
| 27 | `src/compiler/utilities.ts:5183:12` | `src/compiler/utilities.ts:5182:1` (getExportAssignmentExpression) | R0 at `src/compiler/utilities.ts:5183:12` | **lowers**: no stop; all 26 scoped passes complete | none |
| 28 | `src/compiler/utilities.ts:5389:18` | `src/compiler/utilities.ts:5385:1` (isDynamicName) | R0 at `src/compiler/utilities.ts:5389:18` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 29 | `src/compiler/utilities.ts:5438:12` | `src/compiler/utilities.ts:5437:1` (getTextOfIdentifierOrLiteral) | R0 at `src/compiler/utilities.ts:5438:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R7 at `src/compiler/utilities.ts:5438:40` |
| 30 | `src/compiler/utilities.ts:5443:12` | `src/compiler/utilities.ts:5442:1` (getEscapedTextOfIdentifierOrLiteral) | R0 at `src/compiler/utilities.ts:5443:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | R7 at `src/compiler/utilities.ts:5443:33` |
| 31 | `src/compiler/utilities.ts:5468:12` | `src/compiler/utilities.ts:5467:1` (isProtoSetter) | R0 at `src/compiler/utilities.ts:5468:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 32 | `src/compiler/utilities.ts:6148:27` | `src/compiler/utilities.ts:6147:1` (hasInvalidEscape) | R0 at `src/compiler/utilities.ts:6148:27` | **lowers**: no stop; all 26 scoped passes complete | none |
| 33 | `src/compiler/utilities.ts:6782:41` | `src/compiler/utilities.ts:6780:1` (getThisParameter) | R0 at `src/compiler/utilities.ts:6782:41` | **other stop**: R1 at `src/compiler/binder.ts:2258:13` | none |
| 34 | `src/compiler/utilities.ts:6910:12` | `src/compiler/utilities.ts:6909:1` (getEffectiveReturnTypeNode) | R0 at `src/compiler/utilities.ts:6910:12` | **other stop**: R1 at `src/compiler/binder.ts:2258:13` | R8 at `src/compiler/utilities.ts:6911:9` |
| 35 | `src/compiler/utilities.ts:7539:9` | `src/compiler/utilities.ts:7538:1` (tryGetPropertyAccessOrIdentifierToString) | R0 at `src/compiler/utilities.ts:7539:9` | **unchanged**: R0 at `src/compiler/utilities.ts:7539:9` | R7 at `src/compiler/utilities.ts:7540:66` |
| 36 | `src/compiler/utilities.ts:11129:12` | `src/compiler/utilities.ts:11128:1` (getEscapedTextOfJsxAttributeName) | R0 at `src/compiler/utilities.ts:11129:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 37 | `src/compiler/utilities.ts:11134:12` | `src/compiler/utilities.ts:11133:1` (getTextOfJsxAttributeName) | R0 at `src/compiler/utilities.ts:11134:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 38 | `src/compiler/utilities.ts:11156:12` | `src/compiler/utilities.ts:11155:1` (intrinsicTagNameToString) | R0 at `src/compiler/utilities.ts:11156:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 39 | `src/compiler/utilities.ts:11207:12` | `src/compiler/utilities.ts:11206:1` (getNameFromImportAttribute) | R0 at `src/compiler/utilities.ts:11207:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 40 | `src/compiler/utilitiesPublic.ts:620:9` | `src/compiler/utilitiesPublic.ts:619:1` (isEmptyBindingPattern) | R2 at `src/compiler/utilitiesPublic.ts:621:22` | **unchanged (censored)**: R2 at `src/compiler/utilitiesPublic.ts:621:22` | none |
| 41 | `src/compiler/utilitiesPublic.ts:630:9` | `src/compiler/utilitiesPublic.ts:629:1` (isEmptyBindingElement) | R0 at `src/compiler/utilitiesPublic.ts:630:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 42 | `src/compiler/utilitiesPublic.ts:638:12` | `src/compiler/utilitiesPublic.ts:636:1` (walkUpBindingElementsAndPatterns) | R0 at `src/compiler/utilitiesPublic.ts:638:12` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 43 | `src/compiler/utilitiesPublic.ts:927:20` | `src/compiler/utilitiesPublic.ts:925:1` (getDeclarationIdentifier) | R1 at `src/compiler/binder.ts:1109:17` | **unchanged (censored)**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 44 | `src/compiler/utilitiesPublic.ts:1038:13` | `src/compiler/utilitiesPublic.ts:1036:1` (getJSDocParameterTagsWorker) | R1 at `src/compiler/binder.ts:1109:17` | **unchanged (censored)**: R1 at `src/compiler/binder.ts:1109:17` | none |
| 45 | `src/compiler/utilitiesPublic.ts:1319:9` | `src/compiler/utilitiesPublic.ts:1318:1` (getEffectiveTypeParameterDeclarations) | R0 at `src/compiler/utilitiesPublic.ts:1319:9` | **other stop**: R1 at `src/compiler/binder.ts:2258:13` | R3 at `src/compiler/utilitiesPublic.ts:1335:47` |
| 46 | `src/compiler/utilitiesPublic.ts:1391:38` | `src/compiler/utilitiesPublic.ts:1390:1` (isOptionalChainRoot) | R0 at `src/compiler/utilitiesPublic.ts:1391:38` | **lowers**: no stop; all 26 scoped passes complete | none |
| 47 | `src/compiler/utilitiesPublic.ts:2639:18` | `src/compiler/utilitiesPublic.ts:2638:1` (isRestParameter) | R0 at `src/compiler/utilitiesPublic.ts:2639:18` | **other stop**: R4 at `src/compiler/utilitiesPublic.ts:2639:70` | none |
| 48 | `src/compiler/visitorPublic.ts:440:9` | `src/compiler/visitorPublic.ts:436:1` (addDefaultValueAssignmentIfNeeded) | R0 at `src/compiler/visitorPublic.ts:440:9` | **other stop**: R1 at `src/compiler/binder.ts:1109:17` | none |

Exact reason words:

- **R0**: checked predicate membership over an open tag domain
- **R1**: writing a checked object field without its source-slot type certificate
- **R2**: a Map or Set interface implemented by a value that isn't a Map or Set
- **R3**: a read narrowed to undefined of a property a call may write
- **R4**: a read of field type through type JSDocTypeExpression, which may hold a checked view's value
- **R5**: checked predicate overload target BindableObjectDefinePropertyCall
- **R6**: a value of type ParameterDeclaration | undefined seen as Symbol | undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace
- **R7**: an object refinement using an open numeric enum as a literal tag
- **R8**: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.JSDocTypeExpression
- **R9**: an unproven predicate argument for parameter predicate (argument "modifier => !isExportModifier(modifier) && !isDefaultModifier(modifier)")
- **R10**: a value of type ElementAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace
- **R11**: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace
- **R12**: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace
- **R13**: a value of type AccessorDeclaration & { readonly name: BigIntLiteral | ComputedPropertyName | Identifier | NoSubstitutionTemplateLiteral | NumericLiteral | StringLiteral; }
- **R14**: a cast the runtime can't check
- **R15**: a value of type ModuleExportName[] | undefined seen as {}[] | undefined, which can write {} where ModuleExportName is read
- **R16**: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace
- **R17**: a method read as a value (hasGlobalName would lose its object, and this with it)
- **R18**: a read of field isTypeOnly through type ImportClause, which may hold a checked view's value

## 2. Units whose first stop changes

For the 48 retained units, using pair's lowering/deferred first-stop definition:

| Head destination | Units |
| --- | ---: |
| Lowers outright; no ordinary, body, or scoped-tail finding | **3** |
| Kind-through-Node read | **0** |
| Other exact reason | **34** |
| Total changed first stops | **37** |

The three newly finding-free units are:

- `src/compiler/utilities.ts:5182:1` — `getExportAssignmentExpression`; all 26 scoped passes complete.
- `src/compiler/utilities.ts:6147:1` — `hasInvalidEscape`; all 26 scoped passes complete.
- `src/compiler/utilitiesPublic.ts:1390:1` — `isOptionalChainRoot`; all 26 scoped passes complete.

Of the 34 other destinations, 31 stop during body lowering at `writing a checked object field without its source-slot type certificate`. The remaining three stop at `checked predicate overload target BindableObjectDefinePropertyCall`, `a read of field type through type JSDocTypeExpression, which may hold a checked view's value`, and `a read of field isTypeOnly through type ImportClause, which may hold a checked view's value`. Every exact location and diagnostic is in the table and [evidence/inventory-changed-first-stops.json](evidence/inventory-changed-first-stops.json).

These metrics answer different scopes and must not be substituted for one another:

| Scope / first-stop priority | Lower | Kind/Node | Elsewhere | Changes |
| --- | ---: | ---: | ---: | ---: |
| **48 retained units; pair lowering/deferred** | **3** | **0** | **34** | **37** |
| Same 48 units; ordinary scans before body | 3 | 0 | 21 | 24 |
| Entire 4,566-unit entry; pair lowering/deferred | 3 | 0 | 36 | 39 |
| Entire entry; ordinary scans before body | 3 | 0 | 22 | 25 |

The whole-entry pair count includes the two overload owners that distinct dedup did not retain. Ordinary-scan priority additionally masks 13 body-stop changes in the canonical group. The earlier report's 25-unit headline was the whole-entry scan-priority control, not pair's first-stop metric for this canonical list.

## 3. Regressions and panics

Across the entire entry: **0** previously finding-free units become stopped, and **0** new panics. Neither side has a unit or pass panic. Ordinary scan findings are identical between base and head. Both pair-priority and scan-priority regression ledgers are retained in [evidence/regressions.json](evidence/regressions.json) and [evidence/ordered-regressions.json](evidence/ordered-regressions.json).

## 4. Kind-through-Node entries for Lower

**Empty group.** No head unit in either first-stop view stops at `a read of field kind through type Node, which may hold a checked view's value`. There are therefore no overload-check-call, CheckedCast, predicate-narrowing, or other entry/read pairs to assign to #zpzm9ta from this run. Earlier body stops, predominantly source-slot certificate checks, prevent the expected transition to that pass. The actual newly exposed widened reads are `JSDocTypeExpression.type` at `utilitiesPublic.ts:2639:70` and `ImportClause.isTypeOnly` at `program.ts:889:9`; their complete entry sets are in the traces.

## Method, validation, and limits

Both sides use one base-adapted, checker-clean `src/tsc/tsc.ts` entry and identical ranges `[0,1500)`, `[1500,3000)`, `[3000,4600)`. The scoped-tail prototype is from `typescript/latent-whole-passes` at `4cadb42b6f48af2e4a99f334da775f308ea88818`. `latentDeferredStop` unconditionally probes `*l.latentSnapshot()`: 347 traced preflight events per side have **zero shared result pointers**. The historical 243 false Node.kind stops are excluded.

First-error body probes use `LWP_SKIP_SCAN=1`; independent ordinary scans use the same inputs and intersecting file ranges with body attempts omitted. All six body/tail children and all six ordinary-scan children have recorded exit 0, complete relation trailers, and no relation overflow. A direct combined scan/body control confirms the same earlier scan refusal before the changed body stop at `factory/utilities.ts:1015:1`. The first body range is byte-identical across revisions.

The prototype runs 26 scoped tail operations in production order through `voidDiscards`. It explicitly omits `findCycles` and dependent Program/plan/ownership proofs, and sibling bodies/call graphs are incomplete. “Lowers outright” means finding-free through this scoped tail, not certified production whole-tsc acceptance. The 503 deferred generic declarations are not counted as lowered.

Raw streams, exact input hashes, child statuses, build/binary hashes, and final Go overlay sources (`.go.txt`) are under [evidence/](evidence/). `analyze.py` and `scan-audit.py` independently reproduce both whole-entry views directly from the committed gzip evidence. `authoritative-inventory.py` rebuilds the exact supplied inventory using its archived merge and AST normalizer. No Go evidence file has a `.go` extension.

Every measurement wrapper records and propagates its own child's exit code, including signals, with a 1,700-second deadline. All authoritative children completed within it. Controls verify exit 7 and child SIGTERM -15 (wrapper shell 143). Optional follow-ups whose supervisors became unavailable after runtime reattachment have unknown exit status and are excluded; see `evidence/controls/unobserved-followups.json`. The earlier standalone inventory recovery is retained only as superseded control evidence.

Reproduce preparation with `prepare.py REPOSITORY NEW_SCRATCH`, then build `./stage3/census/latent/tool` using `-buildvcs=false -overlay NEW_SCRATCH/final/overlay.json`. Run each body range separately with `LWP_SKIP_SCAN=1 LWP_TRACE=TRACE python3 run-chunk.py BINARY ENTRY DEST START END prototype`. For ordinary scans use `scan-overlay.py`, rebuild, and run each range with `LWP_SCAN_ONLY=1`, no `LWP_SKIP_SCAN`, and mode `stock`. Never wrap the complete range matrix in one unbounded process.
