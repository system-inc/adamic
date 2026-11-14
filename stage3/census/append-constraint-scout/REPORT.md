# Append constraint scout

## Resolved signature and site

The native checker resolves `scanner.ts:2934:34` to:

```ts
(to: { pos: number; end: number; value: number; }[] | undefined,
 value: { pos: number; end: number; value: number; }):
 { pos: number; end: number; value: number; }[]
```

Inferred type argument: `T = { pos: number; end: number; value: number; }`; constraint: `{}`. There are zero unbound type parameters. This is the concrete structural element shape, not `{}`. The selected overload is adapted `core.ts:933`, not the implementation at line 936. Both all-79-root and `sys.ts`-root probes agree. Querying the contextual type before the signature in a fresh process gives the same result. Stock TypeScript 6.0.3, including after semantic diagnostics, agrees.

The argument at column 41 is `decimalEscapes`; the callee starts at column 34 and the second argument at column 57. Its own type is `(TextRange & { value: number; })[] | undefined`, declared at adapted scanner.ts:2642. Its ordinary contextual type is the concrete first parameter above.

Adapted scanner.ts:

```text
2929            Debug.assertEqual(charCodeUnchecked(pos - 1), CharacterCodes.backslash);
2930            const ch = charCodeChecked(pos);
2931            if (ch >= CharacterCodes._1 && ch <= CharacterCodes._9) {
2932                const start = pos;
2933                scanDigits();
2934                decimalEscapes = append(decimalEscapes, { pos: start, end: pos, value: +tokenValue });
2935                return true;
2936            }
2937            return false;
```

Adapted core.ts retains the upstream overloads (now lines 931, 933, 935), followed by the implementation at 936:

```ts
export function append<T extends {}>(to: T[], value: T | undefined): T[];
export function append<T extends {}>(to: T[] | undefined, value: T): T[];
export function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined;
export function append<T extends {}>(to: T[] | undefined, value: T | undefined): T[] | undefined {
    if (value === undefined) return to;
    if (to === undefined) return [value];
    to.push(value);
    return to;
}
```

## Where lowering gets `{}[]`

This is the separate **constraint-binding relation rule**, not a failure to instantiate the resolved call signature. The isolated scanner scan reproduces the exact reported refusal while the ordinary contextual type remains concrete. Reading the unchanged production code gives this path:

1. `internal/lower/constraint_relation.go:54` registers/calls `constraintRelationRefusal` (rule 3350, registration at 59), which invokes `typeArgumentBindings`.
2. `internal/lower/type_argument_bindings.go:50` obtains `GetResolvedSignature(node)`. `signatureTypeArgumentBindings` at 138 reads its generic `Target()` at 145, its instantiation mapper at 149–153, and the declared type parameter constraint at 160. At 164 it records the actual inferred argument, instantiated constraint, selected overload generic signature and mapper separately.
3. `internal/lower/constraint_relation.go:80` sends an inferred binding to `boundArgumentsRefusal`. At 120–128 it starts with the actual bindings, **replaces this parameter with its constraint** (`bound[index] = binding.constraint`, line 125), and constructs `asConstraint`.
4. At 129 and 139 it reads the selected overload generic signature's declared parameter. At 140 it explicitly constructs both `given` (actual instantiation) and `slot` (constraint substitution). Here declared `T[] | undefined` becomes concrete-element `[] | undefined` in `given`, but `{}[] | undefined` in `slot`. It does not select the implementation signature.
5. The assignability guard at 157 accepts the argument into `slot`; line 161 calls `judgeBinding` with that synthesized slot. `judgeBinding` uses `freshOrWidened` at 201 and calls `wideningRefusal` at 206.
6. `internal/lower/invariance.go:1164` constructs the exact `a value of type ... seen as ..., which can write ... where ... is read` reason.

The ordinary argument-view route is different: `invariance.go:700–713` reads `slotType`, and `slotType` at 1391–1397 reads a carried explicit slot or `GetContextualType`. The checker's ordinary contextual call-argument route reads the resolved signature (`cohere/TypeScript/tsc/internal/checker/checker.go:30326–30353`) and parameter type (`relater.go:1939–1942`). It does not explain this refusal: the constraint rule manufactures the wider relation target on purpose, as its comment at `constraint_relation.go:22–25` states.

## Count by callee

**191 matching findings at 191 distinct source positions across all 79 files.** All 79 independent processes exited 0, wrote their terminal relation report, and emitted zero lowering units. No timeout or missing file. Four workers completed in 685.471 seconds; the largest file, checker.ts, completed in 216.021 seconds. The 4,399 distinct findings across all reasons are not the requested family total.

| Callee | Matching sites |
| --- | ---: |
| `append` | 169 |
| `convertOrRepopulateDiagnosticMessageChainArray` | 3 |
| `createNodeArray` | 14 |
| `insertStatementAfterPrologue` | 2 |
| `insertStatementsAfterPrologue` | 2 |
| `pollWatchedFileQueue` | 1 |
| `appendIfUnique` | 0 |
| `addRange` | 0 |
| **Total** | **191** |

The 169 `append` findings exhaust every emitted reason containing `seen as {}[]`, including its `| undefined` form. The other 22 targets are arrays of a constrained element type: `Node` (14), `DiagnosticMessageChain | ReusableDiagnosticMessageChain` (3), `Statement` (4), and `WatchedFileWithIsClosed | undefined` (1, from parameter `(T | undefined)[]`). Adapted `appendIfUnique<T>` at core.ts:1028 and `addRange<T>` at 991–995 have no declared constraint; `signatureTypeArgumentBindings` skips a parameter with no constraint at type_argument_bindings.go:160–162. Their zero here is an observed count, not a compilation claim.

Findings are deduplicated by `(kind, where, reason, text)`. Stock AST attribution identifies the *outer call receiving the refused expression*: for example, `append(arrayOf(...), ...)` is attributed to `append`, with argument `arrayOf(...)`, rather than to the nested `arrayOf`. Every requested row follows; line and column are one-based in the adapted tree. Argument whitespace is collapsed only in this table; evidence/selected.json retains exact source text and full reasons.

### append (169)

| File | Line:column | Argument | Seen as |
| --- | --- | --- | --- |
| `builder.ts` | 1328:25 | `emitSignatures` | `{}[] \| undefined` |
| `builder.ts` | 1372:21 | `affectedFilesPendingEmit` | `{}[] \| undefined` |
| `builder.ts` | 1425:34 | `fileIdsList` | `{}[] \| undefined` |
| `builder.ts` | 1455:33 | `result` | `{}[] \| undefined` |
| `builder.ts` | 1500:70 | `result` | `{}[] \| undefined` |
| `builder.ts` | 1503:33 | `result` | `{}[] \| undefined` |
| `builder.ts` | 1517:29 | `result` | `{}[] \| undefined` |
| `builder.ts` | 1585:40 | `changeFileSet` | `{}[] \| undefined` |
| `checker.ts` | 5355:34 | `results` | `{}[] \| undefined` |
| `checker.ts` | 5372:30 | `results` | `{}[] \| undefined` |
| `checker.ts` | 5416:37 | `bestContainers` | `{}[]` |
| `checker.ts` | 5431:31 | `concatenate(concatenate([container], additionalContainers), reexportContainers)` | `{}[]` |
| `checker.ts` | 5447:26 | `res` | `{}[]` |
| `checker.ts` | 5608:40 | `contained` | `{}[] \| undefined` |
| `checker.ts` | 5616:39 | `nonContained` | `{}[] \| undefined` |
| `checker.ts` | 8287:52 | `newLocals` | `{}[] \| undefined` |
| `checker.ts` | 8290:52 | `oldLocals` | `{}[] \| undefined` |
| `checker.ts` | 13086:39 | `outerTypeParameters` | `{}[] \| undefined` |
| `checker.ts` | 13095:46 | `outerAndOwnTypeParameters` | `{}[] \| undefined` |
| `checker.ts` | 14576:37 | `signatures` | `{}[] \| undefined` |
| `checker.ts` | 14592:23 | `indexInfos` | `{}[] \| undefined` |
| `checker.ts` | 14650:37 | `indexInfos` | `{}[] \| undefined` |
| `checker.ts` | 14656:37 | `indexInfos` | `{}[] \| undefined` |
| `checker.ts` | 15272:42 | `constraints` | `{}[] \| undefined` |
| `checker.ts` | 15274:46 | `constraints` | `{}[]` |
| `checker.ts` | 15290:46 | `constraints` | `{}[]` |
| `checker.ts` | 15648:45 | `indexTypes` | `{}[] \| undefined` |
| `checker.ts` | 15652:45 | `indexTypes` | `{}[] \| undefined` |
| `checker.ts` | 15711:71 | `!writeTypes ? propTypes.slice() : writeTypes` | `{}[]` |
| `checker.ts` | 16799:61 | `inferences` | `{}[] \| undefined` |
| `checker.ts` | 16812:45 | `inferences` | `{}[] \| undefined` |
| `checker.ts` | 16817:45 | `inferences` | `{}[] \| undefined` |
| `checker.ts` | 16822:45 | `inferences` | `{}[] \| undefined` |
| `checker.ts` | 16835:45 | `inferences` | `{}[] \| undefined` |
| `checker.ts` | 17285:42 | `constraints` | `{}[] \| undefined` |
| `checker.ts` | 17297:50 | `constraints` | `{}[] \| undefined` |
| `checker.ts` | 17498:28 | `types` | `{}[] \| undefined` |
| `checker.ts` | 18039:36 | `arrayOf(type.target.fixedLength, i => getStringLiteralType("" + i))` | `{}[]` |
| `checker.ts` | 19342:52 | `types` | `{}[]` |
| `checker.ts` | 19895:49 | `extraTypes` | `{}[]` |
| `checker.ts` | 19996:27 | `getIdentifierChain(node.left)` | `{}[]` |
| `checker.ts` | 22448:53 | `relatedInformation` | `{}[] \| undefined` |
| `checker.ts` | 22978:31 | `propTypes` | `{}[] \| undefined` |
| `checker.ts` | 26909:73 | `inference.contraCandidates` | `{}[] \| undefined` |
| `checker.ts` | 26914:63 | `inference.candidates` | `{}[] \| undefined` |
| `checker.ts` | 32402:58 | `indexInfoCandidates` | `{}[] \| undefined` |
| `checker.ts` | 32436:30 | `types` | `{}[] \| undefined` |
| `checker.ts` | 33510:41 | `components` | `{}[] \| undefined` |
| `checker.ts` | 41397:46 | `missingElements` | `{}[] \| undefined` |
| `checker.ts` | 41635:44 | `oldTypeParameters` | `{}[] \| undefined` |
| `checker.ts` | 41636:44 | `newTypeParameters` | `{}[] \| undefined` |
| `checker.ts` | 43550:37 | `candidates` | `{}[] \| undefined` |
| `checker.ts` | 45889:33 | `yieldTypes` | `{}[] \| undefined` |
| `checker.ts` | 45890:34 | `returnTypes` | `{}[] \| undefined` |
| `checker.ts` | 45891:32 | `nextTypes` | `{}[] \| undefined` |
| `checker.ts` | 45982:40 | `allIterationTypes` | `{}[] \| undefined` |
| `checker.ts` | 46414:47 | `methodParameterTypes` | `{}[] \| undefined` |
| `checker.ts` | 46416:40 | `methodReturnTypes` | `{}[] \| undefined` |
| `checker.ts` | 46432:38 | `returnTypes` | `{}[] \| undefined` |
| `checker.ts` | 46452:34 | `returnTypes` | `{}[] \| undefined` |
| `checker.ts` | 46456:34 | `returnTypes` | `{}[] \| undefined` |
| `checker.ts` | 47577:53 | `notImplementedInfo.get(derivedClassDecl)?.missedProperties` | `{}[] \| undefined` |
| `checker.ts` | 51500:48 | `mods` | `{}[] \| undefined` |
| `commandLineParser.ts` | 3567:21 | `extendedConfigPath` | `{}[]` |
| `commandLineParser.ts` | 3661:50 | `rootCompilerOptions` | `{}[] \| undefined` |
| `factory/emitNode.ts` | 251:31 | `emitNode.helpers` | `{}[] \| undefined` |
| `factory/nodeFactory.ts` | 6929:24 | `target` | `{}[]` |
| `moduleNameResolver.ts` | 2254:26 | `entrypoints` | `{}[] \| undefined` |
| `moduleSpecifiers.ts` | 503:44 | `nodeModulesSpecifiers` | `{}[] \| undefined` |
| `moduleSpecifiers.ts` | 524:46 | `redirectPathsSpecifiers` | `{}[] \| undefined` |
| `moduleSpecifiers.ts` | 532:45 | `relativeSpecifiers` | `{}[] \| undefined` |
| `moduleSpecifiers.ts` | 535:42 | `pathsSpecifiers` | `{}[] \| undefined` |
| `moduleSpecifiers.ts` | 548:41 | `relativeSpecifiers` | `{}[] \| undefined` |
| `parser.ts` | 8027:31 | `list` | `{}[] \| undefined` |
| `parser.ts` | 8034:27 | `list` | `{}[] \| undefined` |
| `parser.ts` | 8041:31 | `list` | `{}[] \| undefined` |
| `parser.ts` | 8050:31 | `list` | `{}[] \| undefined` |
| `parser.ts` | 9453:47 | `children` | `{}[] \| undefined` |
| `parser.ts` | 9642:56 | `jsDocPropertyTags` | `{}[] \| undefined` |
| `parser.ts` | 9700:41 | `parameters` | `{}[] \| undefined` |
| `parser.ts` | 10053:44 | `commentDirectives` | `{}[] \| undefined` |
| `parser.ts` | 10063:44 | `commentDirectives` | `{}[] \| undefined` |
| `program.ts` | 3356:34 | `imports` | `{}[] \| undefined` |
| `program.ts` | 3374:38 | `imports` | `{}[] \| undefined` |
| `programDiagnostics.ts` | 313:38 | `relatedInfo` | `{}[] \| undefined` |
| `scanner.ts` | 2934:41 | `decimalEscapes` | `{}[] \| undefined` |
| `scanner.ts` | 3010:46 | `groupNameReferences` | `{}[] \| undefined` |
| `scanner.ts` | 3635:13 | `commentDirectives` | `{}[] \| undefined` |
| `transformer.ts` | 635:30 | `emitHelpers` | `{}[] \| undefined` |
| `transformers/classFields.ts` | 2185:35 | `membersArray` | `{}[] \| undefined` |
| `transformers/classFields.ts` | 2186:35 | `membersArray` | `{}[] \| undefined` |
| `transformers/classFields.ts` | 2187:35 | `membersArray` | `{}[] \| undefined` |
| `transformers/classFields.ts` | 2188:35 | `membersArray` | `{}[] \| undefined` |
| `transformers/declarations.ts` | 803:32 | `newParams` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 181:30 | `expressions` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 294:17 | `pendingDeclaration.pendingExpressions` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 306:75 | `pendingExpressions` | `{}[]` |
| `transformers/destructuring.ts` | 315:37 | `pendingExpressions` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 321:62 | `pendingExpressions` | `{}[]` |
| `transformers/destructuring.ts` | 409:42 | `bindingElements` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 493:49 | `restContainingElements` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 494:42 | `bindingElements` | `{}[] \| undefined` |
| `transformers/destructuring.ts` | 497:42 | `bindingElements` | `{}[] \| undefined` |
| `transformers/es2015.ts` | 512:13 | `taggedTemplateStringDeclarations` | `{}[]` |
| `transformers/es2015.ts` | 2756:42 | `assignments` | `{}[] \| undefined` |
| `transformers/es2018.ts` | 218:13 | `taggedTemplateStringDeclarations` | `{}[]` |
| `transformers/es2018.ts` | 496:21 | `chunkObject` | `{}[] \| undefined` |
| `transformers/es2018.ts` | 757:24 | `statements` | `{}[]` |
| `transformers/es2018.ts` | 1264:49 | `statements` | `{}[] \| undefined` |
| `transformers/es2018.ts` | 1273:45 | `statements` | `{}[] \| undefined` |
| `transformers/es2018.ts` | 1307:41 | `statements` | `{}[] \| undefined` |
| `transformers/es2018.ts` | 1325:41 | `statements` | `{}[] \| undefined` |
| `transformers/esDecorators.ts` | 753:41 | `leadingBlockStatements` | `{}[] \| undefined` |
| `transformers/esDecorators.ts` | 786:49 | `leadingBlockStatements` | `{}[]` |
| `transformers/esDecorators.ts` | 914:50 | `trailingBlockStatements` | `{}[] \| undefined` |
| `transformers/esDecorators.ts` | 924:46 | `trailingBlockStatements` | `{}[] \| undefined` |
| `transformers/esnext.ts` | 201:21 | `topLevelStatements` | `{}[]` |
| `transformers/esnext.ts` | 397:20 | `statements` | `{}[]` |
| `transformers/esnext.ts` | 590:38 | `expressions` | `{}[] \| undefined` |
| `transformers/legacyDecorators.ts` | 593:34 | `expressions` | `{}[] \| undefined` |
| `transformers/module/esnextAnd2015.ts` | 101:68 | `importsAndRequiresToRewriteOrShim` | `{}[] \| undefined` |
| `transformers/module/esnextAnd2015.ts` | 266:13 | `statements` | `{}[] \| undefined` |
| `transformers/module/esnextAnd2015.ts` | 297:17 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 247:64 | `importsAndRequiresToRewriteOrShim` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 284:20 | `statements` | `{}[]` |
| `transformers/module/module.ts` | 290:21 | `statements` | `{}[]` |
| `transformers/module/module.ts` | 308:16 | `statements` | `{}[]` |
| `transformers/module/module.ts` | 627:20 | `statements` | `{}[]` |
| `transformers/module/module.ts` | 631:17 | `statements` | `{}[]` |
| `transformers/module/module.ts` | 643:16 | `statements` | `{}[]` |
| `transformers/module/module.ts` | 1485:21 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1505:17 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1560:21 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1577:21 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1604:21 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1753:17 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1772:33 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1788:17 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1805:33 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1850:44 | `variables` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1853:44 | `variables` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1875:44 | `variables` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1876:46 | `expressions` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1880:46 | `expressions` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1886:37 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1894:37 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 1898:33 | `statements` | `{}[] \| undefined` |
| `transformers/module/module.ts` | 2138:29 | `statements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 802:17 | `hoistedStatements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 816:40 | `hoistedStatements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 837:13 | `statements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 894:33 | `statements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 901:42 | `expressions` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 909:37 | `statements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 1189:29 | `statements` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 1374:38 | `expressions` | `{}[] \| undefined` |
| `transformers/module/system.ts` | 1992:40 | `exportedNames` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1070:37 | `newMembers` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1128:37 | `decorators` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1132:37 | `decorators` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1136:37 | `decorators` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1147:37 | `properties` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1151:37 | `properties` | `{}[] \| undefined` |
| `transformers/ts.ts` | 1155:37 | `properties` | `{}[] \| undefined` |
| `transformers/utilities.ts` | 239:56 | `exportedNames` | `{}[] \| undefined` |
| `transformers/utilities.ts` | 288:52 | `exportedNames` | `{}[] \| undefined` |
| `transformers/utilities.ts` | 327:40 | `exportedNames` | `{}[] \| undefined` |
| `transformers/utilities.ts` | 367:36 | `exportedNames` | `{}[] \| undefined` |
| `utilities.ts` | 10136:31 | `requires` | `{}[] \| undefined` |

### convertOrRepopulateDiagnosticMessageChainArray (3)

| File | Line:column | Argument | Seen as |
| --- | --- | --- | --- |
| `builder.ts` | 544:66 | `chain.next as T[]` | `(DiagnosticMessageChain \| ReusableDiagnosticMessageChain)[] \| undefined` |
| `builder.ts` | 550:66 | `chain.next as T[]` | `(DiagnosticMessageChain \| ReusableDiagnosticMessageChain)[] \| undefined` |
| `builder.ts` | 553:65 | `chain.next as T[]` | `(DiagnosticMessageChain \| ReusableDiagnosticMessageChain)[] \| undefined` |

### createNodeArray (14)

| File | Line:column | Argument | Seen as |
| --- | --- | --- | --- |
| `parser.ts` | 3113:32 | `list` | `Node[]` |
| `parser.ts` | 3558:32 | `list` | `Node[]` |
| `parser.ts` | 3665:32 | `list` | `Node[]` |
| `parser.ts` | 3699:32 | `list` | `Node[]` |
| `parser.ts` | 4834:62 | `types` | `Node[]` |
| `parser.ts` | 6076:45 | `[...children.slice(0, children.length - 1), newLast]` | `Node[]` |
| `parser.ts` | 6186:32 | `list` | `Node[]` |
| `parser.ts` | 8054:40 | `list` | `Node[]` |
| `parser.ts` | 9020:59 | `tags` | `Node[]` |
| `parser.ts` | 9021:93 | `parts` | `Node[]` |
| `parser.ts` | 9302:44 | `parts` | `Node[]` |
| `parser.ts` | 9508:39 | `concatenate([finishNode(textOnly, commentStart, commentEnd)], comments) as JSDocComment[]` | `Node[]` |
| `parser.ts` | 9702:40 | `parameters \|\| []` | `Node[]` |
| `parser.ts` | 9863:40 | `typeParameters` | `Node[]` |

### insertStatementAfterPrologue (2)

| File | Line:column | Argument | Seen as |
| --- | --- | --- | --- |
| `utilities.ts` | 1168:41 | `to` | `Statement[]` |
| `utilities.ts` | 1173:41 | `to` | `Statement[]` |

### insertStatementsAfterPrologue (2)

| File | Line:column | Argument | Seen as |
| --- | --- | --- | --- |
| `utilities.ts` | 1153:42 | `to` | `Statement[]` |
| `utilities.ts` | 1158:42 | `to` | `Statement[]` |

### pollWatchedFileQueue (1)

| File | Line:column | Argument | Seen as |
| --- | --- | --- | --- |
| `sys.ts` | 307:13 | `queue` | `(WatchedFileWithIsClosed \| undefined)[]` |

### Related array-valued constraints: not (2 additional)

These two findings are listed separately from the 191 array-of-element-constraint sites above. Here `T extends unknown[]` is itself an array constraint: inferred tuple `[node: Node]` is judged as `unknown[]`. Their locations name the `not(...)` binding call, not an array-valued runtime argument. Including this related form gives **193 enumerated findings**.

| File | Line:column | Callee | Callback argument | Inferred T | Seen as |
| --- | --- | --- | --- | --- | --- |
| `checker.ts` | 45143:68 | `not` | `isOmittedExpression` | `[node: Node]` | `unknown[]` |
| `checker.ts` | 49029:113 | `not` | `isInterfaceDeclaration` | `[node: Node]` | `unknown[]` |

All other constrained array-target candidates were reviewed. The one remaining row, checker.ts:18039:36, also views the nested `arrayOf(...)` result as `Type[]` through the actual `append` instantiation; `Type` is not the `{}` constraint. That reason is excluded, while the separate `{}[]` reason at the same position is included. Two `getDiagnosticsHelper` targets are function types returning readonly arrays, not array targets, so are excluded. evidence/all-seen-as.json.gz preserves attribution metadata for every “seen as” row, including exclusions, declaration constraints and a hash identifying its full reason in the raw census.

## Setup and reproducibility

- Fetched origin/main: **894c8a2b2e7a5bd1fd6adc3738a6d92ae16de558**. Fresh `git worktree add --detach` at `/workspace/scratch/append-constraint-scout`; the report branch was created afterward. No shared source checkout edits.
- Recursive submodules succeeded: cohere `5360753e19e4417675a90e92a1823bea66b13918`; its native TypeScript checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. No credential error.
- Separate Microsoft TypeScript clone: HEAD and peeled `v6.0.3` both **050880ce59e30b356b686bd3144efe24f875ebc8**, package version 6.0.3. Exported `ADAMIC_TYPESCRIPT_SOURCE` to it. `npm ci` in stage3/api succeeded (3 packages). Main's unchanged apply pipeline succeeded into scratch and produced exactly 79 compiler .ts files.
- Wrapper source: branch typescript/promise-discard-sites, tip `2dc0593b937f8fbaf6c32f8107a899f78c223966`, scan-only.go.txt blob `cfa38ff895a9e1979d4405f2b308f8433e228b19`. That tip has REPORT.md, **not README.md**; its report documents the method. The wrapper is copied unchanged here.
- Go 1.27.1, Node v24.19.0, npm 11.9.0, Git 2.52.0. All writes used scratch. No compiler or adapter source was changed. build.py generates only Go overlays; ordinary output APIs are disabled and `LATENT_ASSERT_NO_OUTPUT=1` checks the guards on every run. The two-site control reported both mutable-array refusals with zero units.

Commands after the fresh worktree/submodule/upstream/npm/apply setup (use new scratch paths):

```sh
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/scout-tmp
mkdir -p "$TMPDIR"
python3 stage3/census/append-constraint-scout/build.py "$PWD" OVERLAY BINARY
python3 stage3/census/append-constraint-scout/run-files.py BINARY ADAPTED RESULTS
python3 stage3/census/append-constraint-scout/collect.py RESULTS ADAPTED COLLECTED
export NODE_PATH="$PWD/stage3/api/node_modules"
node stage3/census/append-constraint-scout/attribute.cjs ADAPTED/src/compiler COLLECTED/scan-all.jsonl > ATTRIBUTED.json
node stage3/census/append-constraint-scout/constraints.cjs ADAPTED/src/compiler ATTRIBUTED.json > CONSTRAINTS.json
python3 stage3/census/append-constraint-scout/select.py CONSTRAINTS.json SELECTED.json
SCOUT_SIGNATURE=1 SCOUT_SYS_ROOT=1 LATENT_ASSERT_NO_OUTPUT=1 BINARY ADAPTED/src/compiler UNUSED.jsonl
SCOUT_SIGNATURE=1 SCOUT_CONTEXT_FIRST=1 LATENT_ASSERT_NO_OUTPUT=1 BINARY ADAPTED/src/compiler UNUSED-2.jsonl
node stage3/census/append-constraint-scout/stock.cjs ADAPTED/src/compiler
```

The initial combined run's parent was interrupted by an environment refresh. Its continuing child was then deliberately stopped in favor of independent per-file processes; it emitted no file records and contributes nothing to these counts. Only the completed 79-file rerun is counted. evidence/coverage.json verifies every source hash, per-file exit/trailer/zero-unit status, and exact manifest; scan-all.jsonl.gz retains raw headers, findings and relation trailers. provenance.json records pins and binary/overlay hashes; apply.log.gz, patch-set.md, controls and native/stock signature JSON retain the other observations.

## What this cannot see

This is refusal scanning only. It does not lower bodies, instantiate native bodies, emit a native program, or prove successful compilation or behavior. An earlier refusal rule can suppress another rule at the same node, and syntax branches may be pruned by the normal walk. Thus these are emitted findings, not a census of all syntactic append calls or every hypothetical isolated relation judgment.

The original wrapper's metadata uses the fixed legacy string “measured on a checker-rejected program” and the driver's default `latent_mode: first-error`. They do not describe this run: the actual headers have `checker_rejected: false`, no diagnostics, and the invoked wrapper is scan-only, with zero units in every file. This discrepancy is preserved and disclosed rather than changing compiler/tooling sources.

Stock type-only controls use strict options and `types: []`, rather than the upstream build configuration. They report 58 adapted and 2,230 pristine diagnostics; they establish the observed site signature, not a successful upstream build. Native sys-root, all-root and contextual-first measurements agree on the concrete T.

**Optional Node lane-oracle execution control: skipped.** No lane-oracle run or call count was measured; no runtime conclusion is claimed. The requested site is upstream 6.0.3 scanner.ts:2929, and the result established here is the lowering rule's deliberate constraint substitution.
