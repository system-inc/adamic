# Step 16: generics at runtime and inside generics

## Evidence and counting

The compiler base is `8cb5e7c1`. The guarded full census was rerun on this base over the adapted pinned TypeScript compiler directory. It is measurement on a checker-rejected program, with 324 diagnostics, and cannot emit IR or invoke either backend.

Roots count distinct attempted `unit` positions for each exact reason. Sites count distinct diagnostic positions. Repeated imported-helper observations and refusal-scan/lowering repetitions do not add roots. Root counts across reasons overlap and must not be summed as distinct programs.

Hidden bytes are the frozen outermost-cause attribution from `codex/stage3-hidden-ranking` at `6c4fc1af`, measured on `ed6e2975`, not this base. They are exact-reason-wide totals, so short cast/predicate reasons can include historical nongeneric sites. They estimate exposure after removing one outer reason; they do not claim successful lowering or measured retirement. `unmeasured` means this exact reason has no row in that historical ledger. Zero means a measured row with no credited bytes.

The AST type-parameter names select exact type-bearing diagnostics, plus all explicit generic/type-parameter/monomorphization diagnostics. Lexical reading-X failures are excluded. Concrete-name collisions enter the scoped table only where TypeScript AST ranges prove a binder with that spelling at the witness. The remaining collisions are retained separately below. Cast and predicate findings with an AST target containing an enclosing binder are included even when the short reason omits the target type. Generic refusals remain soundness obligations, not automatic compiler lessons.

The frozen giant-body refusal table was also reconciled by exact reason. Its site counts are retained in the machine ledger, not substituted for attempted roots. It measures compiler 0d3f2715 over eight selected giant bodies. Table-only reasons stay visible with zero current roots and unmeasured hidden bytes. Witnesses use diagnostic positions on this base where available, then historical table/boundary positions. Fewer than three distinct witnesses exist for several reasons; those rows retain every available witness rather than inventing locations. Historical-only rows have zero current roots. Full provenance and selected observations are in [baseline.json.gz](step-16-generics/baseline.json.gz).

| Kind and exact reason | Base roots | Base sites | Historical hidden bytes | Historical boundaries | Up to three witnesses |
| --- | ---: | ---: | ---: | ---: | --- |
| NotYet: a function returning T \| undefined | 171 | 74 | 32292 | 159 | checker.ts:25177<br>checker.ts:25222<br>checker.ts:28510 |
| Refused: a cast the runtime can't check | 60 | 72 | 25281 | 936 | builder.ts:545<br>builder.ts:551<br>builder.ts:554 |
| NotYet: a function returning T | 108 | 100 | 18147 | 106 | checker.ts:10284<br>checker.ts:1985<br>checker.ts:2014 |
| NotYet: a value of type readonly T[] \| undefined | 307 | 8 | 13990 | 329 | core.ts:145<br>core.ts:198<br>core.ts:209 |
| NotYet: a value of type T | 135 | 50 | 9830 | 126 | binder.ts:1477<br>builder.ts:536<br>checker.ts:20544 |
| NotYet: a value of type SolutionBuilderState<T> | 23 | 22 | 7974 | 44 | tsbuildPublic.ts:1204<br>tsbuildPublic.ts:1359<br>tsbuildPublic.ts:1375 |
| NotYet: a function returning U \| undefined | 34 | 10 | 7076 | 30 | core.ts:33<br>core.ts:50<br>core.ts:67 |
| NotYet: a value of type X | 4 | 3 | 5575 | 4 | watchUtilities.ts:750<br>watchUtilities.ts:764<br>watchUtilities.ts:782 |
| NotYet: a value of type T \| Program | 9 | 4 | 5495 | 9 | program.ts:5060<br>watch.ts:338<br>watch.ts:565 |
| NotYet: overload 1 of createBinaryExpressionTrampoline with additional implementation type parameters | 6 | 1 | 4549 | 6 | factory/utilities.ts:1436<br>binder.ts:1920<br>checker.ts:40483 |
| NotYet: a generic function as a value | 26 | 14 | 4031 | 17 | core.ts:1370<br>core.ts:220<br>core.ts:2378 |
| NotYet: an array of T | 61 | 39 | 3740 | 39 | checker.ts:46986<br>core.ts:1030<br>core.ts:1052 |
| NotYet: overload 1 of mutateMapSkippingNewValues with additional implementation type parameters | 4 | 1 | 2626 | 16 | utilities.ts:8199<br>tsbuildPublic.ts:674<br>tsbuildPublic.ts:675 |
| NotYet: a value of type T \| undefined | 151 | 13 | 2148 | 214 | core.ts:2139<br>core.ts:2487<br>core.ts:850 |
| NotYet: overload 1 of mutateMap with additional implementation type parameters | 5 | 1 | 1768 | 5 | utilities.ts:8248<br>utilities.ts:8250<br>utilities.ts:8251 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 | 2 | 1663 | 4 | moduleNameResolver.ts:361<br>moduleNameResolver.ts:379<br>moduleNameResolver.ts:359 |
| NotYet: a value of type IncludeTypeSpaceImports | 3 | 1 | 1425 | 3 | utilities.ts:12174<br>program.ts:3353<br>transformers/module/module.ts:245 |
| NotYet: overload 2 of arrayFrom with additional implementation type parameters | 14 | 1 | 1352 | 11 | core.ts:1339<br>checker.ts:16693<br>checker.ts:24402 |
| NotYet: overload 1 of arrayToMap with additional implementation type parameters | 7 | 1 | 1309 | 7 | core.ts:1401<br>builder.ts:2390<br>commandLineParser.ts:2307 |
| NotYet: overload 1 of getOriginalNode with additional implementation type parameters | 38 | 1 | 1179 | 36 | utilitiesPublic.ts:762<br>checker.ts:6518<br>checker.ts:6534 |
| NotYet: overload 1 of forEachAncestorDirectory with additional implementation type parameters | 5 | 1 | 1162 | 5 | path.ts:1092<br>moduleNameResolver.ts:503<br>path.ts:1094 |
| NotYet: a function returning T \| T[] \| undefined | 4 | 1 | 975 | 4 | core.ts:953<br>core.ts:947<br>core.ts:949 |
| Refused: overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 4 | 1 | 941 | 4 | utilitiesPublic.ts:826<br>checker.ts:2041<br>utilitiesPublic.ts:817 |
| NotYet: a value of type NodeArray<T> \| undefined | 1 | 1 | 896 | 6 | utilities.ts:1111<br>utilities.ts:1101<br>utilities.ts:1102 |
| Refused: overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined | 6 | 1 | 861 | 11 | utilitiesPublic.ts:786<br>checker.ts:2015<br>checker.ts:1986 |
| NotYet: a value of type SourceFile | 2 | 2 | 751 | 2 | program.ts:1109<br>resolutionCache.ts:861<br>program.ts:1104 |
| NotYet: a value of type U | 9 | 3 | 749 | 3 | core.ts:1202<br>core.ts:354<br>core.ts:579 |
| NotYet: a value of type MapLike<T> | 3 | 1 | 725 | 3 | core.ts:1287<br>commandLineParser.ts:3320<br>moduleSpecifiers.ts:1109 |
| NotYet: overload 3 of group with additional implementation type parameters | 7 | 1 | 652 | 6 | core.ts:1452<br>checker.ts:43104<br>core.ts:1448 |
| NotYet: overload 1 of sortAndDeduplicate with additional implementation type parameters | 6 | 1 | 627 | 5 | core.ts:805<br>core.ts:807<br>core.ts:809 |
| NotYet: overload 1 of setSerializerContextAnd with additional implementation type parameters | 4 | 1 | 616 | 4 | transformers/typeSerializer.ts:153<br>transformers/typeSerializer.ts:146<br>transformers/typeSerializer.ts:154 |
| NotYet: a value of type { forEach: (callbackfn: (value: T, key: K, map: Map<K, T>) => void, thisArg?: any) => void; clear: () => void; } | 4 | 1 | 598 | 9 | utilities.ts:8172<br>resolutionCache.ts:681<br>resolutionCache.ts:682 |
| NotYet: a function returning T \| readonly T[] \| undefined | 5 | 1 | 593 | 5 | core.ts:1160<br>core.ts:1152<br>core.ts:1154 |
| Refused: overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 4 | 1 | 519 | 3 | core.ts:342<br>builder.ts:526<br>builder.ts:564 |
| NotYet: a value of type readonly T[] | 21 | 8 | 515 | 18 | checker.ts:46982<br>core.ts:1028<br>core.ts:1050 |
| NotYet: overload 1 of arrayToMultiMap with additional implementation type parameters | 3 | 1 | 494 | 3 | core.ts:1434<br>core.ts:1436<br>core.ts:1438 |
| NotYet: a value of type K | 17 | 8 | 485 | 11 | builderState.ts:171<br>builderState.ts:180<br>checker.ts:44581 |
| NotYet: overload 1 of forEachTrailingCommentRange with additional implementation type parameters | 5 | 1 | 476 | 5 | scanner.ts:938<br>emitter.ts:3945<br>emitter.ts:6120 |
| NotYet: overload 1 of forEachLeadingCommentRange with additional implementation type parameters | 6 | 1 | 475 | 6 | scanner.ts:932<br>emitter.ts:3951<br>emitter.ts:6112 |
| NotYet: overload 1 of arrayToNumericMap with additional implementation type parameters | 3 | 1 | 460 | 3 | core.ts:1420<br>core.ts:1422<br>core.ts:1424 |
| Refused: overload 1 of skipOuterExpressions result T cannot be served by implementation result Node | 4 | 1 | 443 | 4 | factory/utilities.ts:654<br>factory/utilities.ts:656<br>factory/utilities.ts:658 |
| NotYet: a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> | 2 | 1 | 422 | 2 | watchPublic.ts:420<br>watchPublic.ts:415<br>watchPublic.ts:419 |
| NotYet: a value of type NonNullable<T> | 30 | 12 | 416 | 12 | checker.ts:20543<br>core.ts:1468<br>core.ts:246 |
| NotYet: a function returning V | 2 | 2 | 382 | 2 | core.ts:519<br>moduleNameResolver.ts:1068 |
| NotYet: a function returning T1 & T2 | 1 | 1 | 374 | 1 | core.ts:1495 |
| NotYet: a value of type T[] | 15 | 5 | 364 | 13 | core.ts:1003<br>core.ts:2325<br>core.ts:2340 |
| NotYet: a value of type TEntry | 1 | 1 | 327 | 1 | transformers/utilities.ts:842<br>transformers/utilities.ts:839 |
| NotYet: a function returning NonNullable<T> | 2 | 2 | 318 | 2 | core.ts:1909<br>core.ts:706 |
| NotYet: a value of type T1 | 3 | 1 | 260 | 3 | core.ts:1513<br>tsbuildPublic.ts:327<br>watch.ts:874 |
| NotYet: a value of type MapLike<T> \| undefined | 1 | 1 | 234 | 1 | core.ts:1370<br>utilities.ts:10452 |
| NotYet: a function returning U | 4 | 4 | 219 | 4 | core.ts:93<br>transformers/classFields.ts:869<br>transformers/es2017.ts:196 |
| NotYet: a value of type T \| T[] | 3 | 1 | 207 | 3 | core.ts:1760<br>core.ts:1756<br>core.ts:1758 |
| Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined | 6 | 1 | 204 | 6 | core.ts:656<br>binder.ts:1688<br>builder.ts:1042 |
| NotYet: a function returning TEntry \| undefined | 1 | 1 | 191 | 1 | transformers/utilities.ts:829 |
| NotYet: a function returning TOut | 1 | 1 | 191 | 1 | core.ts:1783 |
| NotYet: a value of type Set<K> | 2 | 1 | 176 | 1 | utilities.ts:8316<br>commandLineParser.ts:2706 |
| NotYet: a value of type V | 1 | 1 | 165 | 1 | transformers/utilities.ts:377 |
| NotYet: a call returning T | 2 | 1 | 154 | 1 | watchUtilities.ts:540<br>watchUtilities.ts:539 |
| Refused: an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 1 | 1 | 145 | 1 | tsbuildPublic.ts:313 |
| NotYet: a value of type CustomTransformerFactory \| TransformerFactory<T> | 2 | 1 | 142 | 2 | transformer.ts:209<br>transformer.ts:219<br>transformer.ts:223 |
| NotYet: a value of type boolean \| V \| undefined | 1 | 1 | 142 | 1 | utilities.ts:942 |
| NotYet: a value of type V \| undefined | 1 | 2 | 106 | 2 | utilities.ts:939<br>utilities.ts:941 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 | 1 | 100 | 1 | transformers/utilities.ts:855 |
| NotYet: a value of type T["kind"] | 189 | 9 | 89 | 187 | factory/nodeFactory.ts:1209<br>factory/nodeFactory.ts:1213<br>factory/nodeFactory.ts:1436 |
| NotYet: a value of type U \| readonly U[] \| undefined | 2 | 2 | 88 | 2 | core.ts:403<br>core.ts:422 |
| NotYet: a value of type K \| undefined | 3 | 2 | 74 | 2 | core.ts:560<br>core.ts:561 |
| NotYet: a function returning TOut \| undefined | 3 | 1 | 73 | 4 | core.ts:1778<br>binder.ts:2337<br>checker.ts:3123 |
| NotYet: an array of U | 5 | 3 | 71 | 3 | core.ts:2524<br>core.ts:325<br>core.ts:488 |
| NotYet: a value of type readonly (readonly T[])[] | 1 | 1 | 69 | 1 | core.ts:2537<br>core.ts:2533 |
| NotYet: an array of V | 1 | 1 | 59 | 1 | core.ts:110 |
| NotYet: a value of type T \| null \| undefined | 3 | 1 | 51 | 2 | debug.ts:255<br>builder.ts:2478<br>checker.ts:8465 |
| NotYet: a value of type NonNullable<U> | 1 | 1 | 46 | 1 | core.ts:2501 |
| NotYet: a value of type U \| undefined | 2 | 2 | 45 | 2 | core.ts:484<br>core.ts:498 |
| NotYet: a value of type T \| readonly T[] | 3 | 1 | 44 | 1 | core.ts:463 |
| NotYet: a Map of T | 1 | 1 | 39 | 1 | core.ts:1908 |
| NotYet: an array of NonNullable<T> | 4 | 2 | 39 | 2 | core.ts:739<br>parser.ts:3507 |
| NotYet: a value of type NonNullable<K> | 1 | 1 | 34 | 1 | utilities.ts:940 |
| NotYet: a value of type T \| T[] \| readonly T[] \| undefined | 1 | 1 | 29 | 1 | core.ts:378 |
| NotYet: a value of type TData | 1 | 1 | 27 | 1 | transformers/utilities.ts:824 |
| NotYet: a function returning A | 1 | 1 | 0 | 1 | debug.ts:268 |
| NotYet: a function returning InferenceContext \| (T & undefined) | 1 | 1 | 0 | 1 | checker.ts:26263 |
| NotYet: a function returning R | 1 | 1 | 0 | 1 | expressionToTypeNode.ts:880 |
| NotYet: a function returning T \| EmptyStatement \| undefined | 3 | 1 | 0 | 3 | factory/nodeFactory.ts:7163<br>factory/nodeFactory.ts:7161<br>factory/nodeFactory.ts:7162 |
| NotYet: a function returning T \| Identifier | 1 | 1 | 0 | 1 | factory/nodeFactory.ts:7141 |
| NotYet: a function returning T \| NumericLiteral \| StringLiteral \| BooleanLiteral | 5 | 1 | 0 | 7 | factory/nodeFactory.ts:7146<br>factory/nodeFactory.ts:6518<br>factory/nodeFactory.ts:6526 |
| NotYet: a function returning T \| StringLiteral | 1 | 1 | 0 | 1 | transformers/declarations.ts:835 |
| NotYet: a function returning TResult | 1 | 1 | 0 | 1 | factory/utilities.ts:1475 |
| NotYet: a function returning TypeMapper \| (T & undefined) | 1 | 1 | 0 | 1 | checker.ts:26378 |
| NotYet: a function returning VisitResult<T> | 1 | 1 | 0 | 1 | checker.ts:2501 |
| NotYet: a function returning WatchFactory<X, Y>[T] | 2 | 2 | 0 | 2 | watchUtilities.ts:727<br>watchUtilities.ts:803 |
| NotYet: a value of type Child | 1 | 1 | 0 | 1 | emitter.ts:4754 |
| NotYet: a value of type Children \| undefined | 44 | 3 | 0 | 47 | emitter.ts:4663<br>emitter.ts:4675<br>emitter.ts:4679 |
| NotYet: a value of type Map<Path, ModeAwareCache<T>> \| undefined | 2 | 1 | 0 | 2 | program.ts:2007<br>program.ts:1996<br>program.ts:2003 |
| NotYet: a value of type Map<string, SingleFileWatcher<T>> | 2 | 1 | 0 | 2 | sys.ts:500<br>sys.ts:1191<br>sys.ts:1207 |
| NotYet: a value of type Map<string, WildcardDirectoryWatcher<T>> | 1 | 1 | 0 | 1 | watchUtilities.ts:515<br>watchPublic.ts:1105 |
| NotYet: a value of type Map<string, [K, V[]]> | 1 | 1 | 0 | 1 | checker.ts:44581<br>checker.ts:44689 |
| NotYet: a value of type NodeArray<T> | 1 | 1 | 0 | 1 | emitter.ts:1348<br>emitter.ts:1321 |
| NotYet: a value of type PrivateEnvironment<TData, TEntry> | 3 | 1 | 0 | 4 | transformers/utilities.ts:840<br>transformers/classFields.ts:2794<br>transformers/classFields.ts:2805 |
| NotYet: a value of type SortedArray<T> | 2 | 1 | 0 | 3 | core.ts:769<br>utilities.ts:6101<br>utilities.ts:6114 |
| NotYet: a value of type Source | 1 | 1 | 0 | 1 | checker.ts:27040 |
| NotYet: a value of type SourceFileOrString | 1 | 1 | 0 | 1 | program.ts:2233<br>program.ts:2231 |
| NotYet: a value of type TKind | 1 | 1 | 0 | 1 | factory/nodeFactory.ts:2280 |
| NotYet: a value of type TKind \| Token<TKind> | 1 | 1 | 0 | 1 | factory/nodeFactory.ts:7157 |
| NotYet: a value of type TNode | 1 | 1 | 0 | 1 | transformers/esDecorators.ts:1239<br>transformers/esDecorators.ts:1236 |
| NotYet: a value of type TOuterState | 6 | 6 | 0 | 6 | factory/utilities.ts:1280<br>factory/utilities.ts:1294<br>factory/utilities.ts:1312 |
| NotYet: a value of type readonly Child[] | 1 | 1 | 0 | 1 | emitter.ts:4729<br>emitter.ts:4490 |
| NotYet: a value of type readonly K[] | 1 | 1 | 0 | 2 | utilities.ts:931<br>program.ts:2505<br>program.ts:2516 |
| NotYet: an array of Child | 1 | 2 | 0 | 2 | emitter.ts:4734<br>emitter.ts:4850 |
| NotYet: an array of TState | 1 | 1 | 0 | 1 | factory/utilities.ts:1396 |
| NotYet: overload 1 of createToken with additional implementation type parameters | 18 | 1 | 0 | 18 | factory/nodeFactory.ts:1441<br>factory/nodeFactory.ts:1442<br>factory/nodeFactory.ts:1443 |
| NotYet: overload 1 of resolveTypeReferenceDirectiveNamesReusingOldState with additional implementation type parameters | 4 | 1 | 0 | 4 | program.ts:2191<br>program.ts:2192<br>program.ts:2193 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 3 | unmeasured | unmeasured | scanner.ts:1110<br>scanner.ts:1111<br>scanner.ts:1112 |
| Refused: a function taking (resolvedProjectReference: ResolvedProjectReference) => T \| undefined seen as one taking (resolvedProjectReference: ResolvedProjectReference) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 0 | unmeasured | unmeasured | program.ts:1732<br>program.ts:1937 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 | 6 | unmeasured | unmeasured | factory/utilities.ts:1282<br>factory/utilities.ts:1295<br>factory/utilities.ts:1313 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | unmeasured | unmeasured | emitter.ts:1287 |
| Refused: a function taking NodeArray<T> seen as one taking NodeArray<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | unmeasured | unmeasured | emitter.ts:1293 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 2 | unmeasured | unmeasured | factory/parenthesizerRules.ts:676<br>factory/parenthesizerRules.ts:677 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | unmeasured | unmeasured | core.ts:2378 |
| Refused: a function taking Visitor seen as one taking Visitor<TIn, Node \| undefined> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | unmeasured | unmeasured | checker.ts:7345<br>expressionToTypeNode.ts:582 |
| Refused: a function taking readonly T[] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | unmeasured | unmeasured | resolutionCache.ts:663 |
| Refused: a function taking readonly T[] \| undefined seen as one taking readonly T[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:522 |
| Refused: a type predicate whose return is not proven (normal return has not narrowed value to NonNullable<T>) | 1 | 1 | unmeasured | unmeasured | debug.ts:248 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 26 | 31 | unmeasured | unmeasured | core.ts:1778<br>core.ts:1783<br>core.ts:2459 |
| Refused: a value of type AccessorDeclaration[] seen as T \| (T \| AccessorDeclaration)[] \| AccessorDeclaration, which can write T \| AccessorDeclaration where AccessorDeclaration is read | 0 | 0 | unmeasured | unmeasured | checker.ts:10893 |
| Refused: a value of type Children seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | unmeasured | unmeasured | emitter.ts:4711 |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 1 | unmeasured | unmeasured | tsbuildPublic.ts:676 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; readonly postfix: boolean; } can be written, so it can write what Mutable<T> can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:5084 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint JSDocType & { readonly type: TypeNode \| undefined; } can be written, so it can write what Mutable<T> can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:5094 |
| Refused: a value of type Mutable<T> seen as T, a type parameter whose constraint Node can be written, so it can write what Mutable<T> can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:7184 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 2 | 6 | unmeasured | unmeasured | core.ts:381<br>core.ts:388<br>core.ts:726 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 10 | 1 | 0 | 1 | utilities.ts:10705<br>factory/nodeFactory.ts:5079<br>factory/nodeFactory.ts:5102 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of any would replace | 8 | 10 | unmeasured | unmeasured | factory/nodeFactory.ts:5079<br>factory/nodeFactory.ts:5102<br>factory/nodeFactory.ts:5112 |
| Refused: a value of type T seen as T, a type parameter whose constraint ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:1160 |
| Refused: a value of type T seen as T, a type parameter whose constraint EntityNameOrEntityNameExpression can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | checker.ts:6411 |
| Refused: a value of type T seen as T, a type parameter whose constraint HasModifiers can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:1159 |
| Refused: a value of type T seen as T, a type parameter whose constraint MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| AccessorDeclaration can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:1161 |
| Refused: a value of type T seen as T, a type parameter whose constraint ModifierSyntaxKind can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:543 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 2 | 2 | unmeasured | unmeasured | checker.ts:6408<br>utilities.ts:12346 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:1019 |
| Refused: a value of type T seen as T, a type parameter whose constraint ResolutionWithFailedLookupLocations can be written, so it can write what T can't hold | 1 | 1 | unmeasured | unmeasured | resolutionCache.ts:654 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 2 | 0 | 2 | utilities.ts:10645<br>utilities.ts:10655 |
| Refused: a value of type TKind seen as TKind, a type parameter whose constraint KeywordTypeSyntaxKind can be written, so it can write what TKind can't hold | 1 | 1 | unmeasured | unmeasured | factory/nodeFactory.ts:579 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 | 1 | unmeasured | unmeasured | core.ts:1570 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 | 1 | unmeasured | unmeasured | core.ts:2533 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 2 | 2 | unmeasured | unmeasured | emitter.ts:1483<br>visitorPublic.ts:604 |
| Refused: a value of type never seen as T, a type parameter whose constraint Node can be written, so it can write what never can't hold | 2 | 2 | unmeasured | unmeasured | factory/nodeFactory.ts:6374<br>transformers/declarations.ts:836 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 | 1 | unmeasured | unmeasured | core.ts:1039 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 1 | 1 | unmeasured | unmeasured | core.ts:414 |
| Refused: a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 | 1 | unmeasured | unmeasured | watchPublic.ts:1135 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 | 1 | unmeasured | unmeasured | tsbuildPublic.ts:1352 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 | 1 | unmeasured | unmeasured | utilities.ts:635 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 4 | 6 | unmeasured | unmeasured | expressionToTypeNode.ts:196<br>utilities.ts:10723<br>utilitiesPublic.ts:777 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Type can be written, so it can write what undefined can't hold | 1 | 2 | unmeasured | unmeasured | checker.ts:25223<br>checker.ts:25225 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 1 | 3 | unmeasured | unmeasured | sys.ts:201<br>sys.ts:210<br>sys.ts:220 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 7 | 12 | unmeasured | unmeasured | factory/utilities.ts:1284<br>factory/utilities.ts:1297<br>factory/utilities.ts:1315 |
| Refused: an unproven relation from Declaration to T: the source is not assignable to the target | 1 | 1 | 0 | 1 | utilities.ts:630 |
| Refused: optional property return in ArrayIterator<Node> absent from structural source ArrayIterator<T>, which can hide fields | 1 | 4 | unmeasured | unmeasured | factory/nodeFactory.ts:1179<br>factory/nodeFactory.ts:1191<br>factory/nodeFactory.ts:1204 |
| Refused: optional property return in ArrayIterator<any> absent from structural source ArrayIterator<Child>, which can hide fields | 1 | 4 | unmeasured | unmeasured | emitter.ts:4687<br>emitter.ts:4688<br>emitter.ts:4699 |
| Refused: overload 1 of arrayFrom result U[] cannot be served by implementation result (T \| U)[] | 3 | 1 | 0 | 3 | core.ts:1337<br>builderState.ts:493<br>builderState.ts:624 |
| Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 5 | 1 | 0 | 5 | core.ts:261<br>checker.ts:5667<br>checker.ts:6583 |
| Refused: overload 1 of find result U \| undefined cannot be served by implementation result T \| undefined | 2 | 1 | 0 | 2 | core.ts:162<br>checker.ts:9053<br>checker.ts:9420 |
| Refused: overload 1 of findLast result U \| undefined cannot be served by implementation result T \| undefined | 1 | 1 | 0 | 1 | core.ts:178<br>checker.ts:3665 |
| Refused: overload 1 of replaceDecoratorsAndModifiers result T cannot be served by implementation result ClassDeclaration \| ClassExpression \| GetAccessorDeclaration \| MethodDeclaration \| ParameterDeclaration \| PropertyDeclaration \| SetAccessorDeclaration | 2 | 1 | 0 | 2 | factory/nodeFactory.ts:7103<br>factory/nodeFactory.ts:7104 |
| Refused: overload 1 of replaceModifiers result T cannot be served by implementation result ArrowFunction \| ClassDeclaration \| ClassExpression \| ConstructorDeclaration \| ConstructorTypeNode \| ... 19 more ... \| VariableStatement | 2 | 1 | 0 | 2 | factory/nodeFactory.ts:7066<br>factory/nodeFactory.ts:7067 |
| Refused: overload 1 of replacePropertyName result T cannot be served by implementation result GetAccessorDeclaration \| MethodDeclaration \| MethodSignature \| PropertyAssignment \| PropertyDeclaration \| PropertySignature \| SetAccessorDeclaration | 2 | 1 | 0 | 2 | factory/nodeFactory.ts:7115<br>factory/nodeFactory.ts:7116 |

## Concrete-name collision audit

These exact reasons contain a spelling also used as a type parameter. Their spelling alone does not establish a generic root. They are excluded from the step totals because the witness has no enclosing type-parameter binder with that spelling. This is a lexical classification, not checker-identity proof.

| Exact reason | Roots | Witnesses |
| --- | ---: | --- |
| NotYet: a tuple element of type string \| number \| boolean \| readonly string[] \| SourceFile \| undefined | 1 | watch.ts:291<br>watch.ts:294 |
| NotYet: checked view field parent of type JSDoc \| ModuleBlock \| SourceFile | 1 | checker.ts:4050 |
| NotYet: checked view field parent of type ModuleBlock \| SourceFile | 34 | binder.ts:3058<br>binder.ts:372<br>checker.ts:10596 |
| NotYet: checked view field parent of type SourceFile | 29 | binder.ts:3060<br>binder.ts:663<br>binder.ts:841 |
| NotYet: checked view field parent of type SourceFile \| ModuleBody | 18 | binder.ts:3040<br>binder.ts:411<br>builderState.ts:535 |
| NotYet: checked view field sourceFiles of type readonly SourceFile[] | 2 | emitter.ts:1314<br>emitter.ts:2050 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | utilities.ts:9003 |
| Refused: a function taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | builder.ts:1727<br>builder.ts:1815<br>builder.ts:1914 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | program.ts:582 |
| Refused: a function taking string seen as one taking [fileName: string, text: string, writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | builder.ts:1727<br>builder.ts:1815<br>builder.ts:1914 |
| Refused: a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | builder.ts:2467 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 3 | checker.ts:2534<br>commandLineParser.ts:3795<br>programDiagnostics.ts:296 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 5 | checker.ts:33274<br>parser.ts:2510<br>parser.ts:4573 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | checker.ts:35176 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 56 | checker.ts:13357<br>checker.ts:19332<br>checker.ts:19391 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | checker.ts:37084 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 51 | binder.ts:853<br>binder.ts:861<br>binder.ts:864 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | checker.ts:35191<br>commandLineParser.ts:2281 |
| Refused: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | program.ts:2916 |
| Refused: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | program.ts:2821 |
| Refused: a value of type DiagnosticWithLocation[] \| undefined seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | program.ts:2918 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | moduleSpecifiers.ts:1257<br>moduleSpecifiers.ts:314 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | moduleSpecifiers.ts:1197<br>moduleSpecifiers.ts:284<br>moduleSpecifiers.ts:463 |
| Refused: a value of type ModuleDeclaration seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | binder.ts:2353 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | factory/utilities.ts:685<br>factory/utilities.ts:692 |
| Refused: a value of type SourceFile seen as Mutable<ModuleDeclaration \| SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | binder.ts:3109 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | factory/nodeFactory.ts:6120<br>parser.ts:1403 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 2 | utilities.ts:1247<br>utilities.ts:1267<br>utilities.ts:1282 |
| Refused: a value of type SourceFile seen as SourceFileLike \| undefined, which can write readonly number[] \| undefined where readonly number[] is read | 1 | checker.ts:49331 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 24 | checker.ts:52410<br>checker.ts:52411<br>emitter.ts:1444 |
| Refused: a value of type SourceFile \| undefined seen as FileReasonToChainCache \| undefined, which can write DiagnosticMessageChain[] \| undefined where RedirectInfo \| undefined is read | 1 | programDiagnostics.ts:216 |
| Refused: a value of type SourceFile \| undefined seen as NodeLinks \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of NodeCheckFlags would replace | 1 | checker.ts:34056 |
| Refused: a value of type SourceFile \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | checker.ts:50788 |
| Refused: a value of type Symbol \| undefined seen as SourceFile \| undefined, which can write number \| undefined where number is read | 1 | checker.ts:3784 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where never is read | 2 | programDiagnostics.ts:168<br>programDiagnostics.ts:240<br>programDiagnostics.ts:269 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 | builderState.ts:562<br>builderState.ts:567 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 | transformer.ts:128 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | program.ts:645<br>watch.ts:601 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 3 | builder.ts:2116<br>builder.ts:2467<br>program.ts:5084 |
| Refused: a value of type string[] seen as DiagnosticArguments, which can write string \| number \| boolean \| readonly string[] \| SourceFile \| undefined where string is read | 1 | program.ts:3490 |
| Refused: a value of type { affectedFile: SourceFile; emitKind: BuilderFileEmit.Js \| BuilderFileEmit.JsMap \| BuilderFileEmit.JsInlineMap \| BuilderFileEmit.DtsErrors \| ... 6 more ... \| BuilderFileEmit.All; } seen as { affectedFile: Program \| SourceFile \| undefined; emitKind: BuilderFileEmit; }, which can write Program \| SourceFile \| undefined where SourceFile is read | 1 | builder.ts:1765 |
| Refused: a value of type { noInferenceFallback?: boolean \| undefined; enclosingDeclaration: ModuleDeclaration; enclosingFile: SourceFile \| undefined; flags: NodeBuilderFlags; ... 28 more ...; out: WriterContextOut; } seen as NodeBuilderContext, which can write Node \| undefined where ModuleDeclaration is read | 1 | checker.ts:10220 |
| Refused: an unproven value assigned to a numeric literal or enum member slot SyntaxKind.SourceFile | 7 | checker.ts:11182<br>checker.ts:34055<br>checker.ts:6196 |
| Refused: optional property packageJsonScope in Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope"> absent from structural source Pick<SourceFile, "fileName" \| "impliedNodeFormat">, which can hide fields | 2 | moduleSpecifiers.ts:200<br>moduleSpecifiers.ts:243 |
| Refused: optional property skipTrivia in SourceMapSource absent from structural source SourceFile, which can hide fields | 1 | emitter.ts:1408 |

## Reproduction

Run the setup, adaptation, overlay build and census commands in a checkout of compiler base `8cb5e7c1`. Run the checked-in report scripts from this delivery branch, where the historical ranking object must also be available. Rebuilding the census on a newer compiler is a different measurement.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/scout-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/scout-adapted > /tmp/scout-adapt.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/scout-overlay > /tmp/scout-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/scout-overlay/overlay.json -o /tmp/scout-census ./stage3/census/latent/tool > /tmp/scout-census-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/scout-census /tmp/scout-adapted/src/compiler /tmp/scout-base.jsonl > /tmp/scout-base-census.log 2>&1
node docs/step-16-generics/parameters.cjs "$HOME/.cache/adamic-stage3/api/node_modules/typescript" /tmp/scout-adapted/src/compiler > /tmp/scout-parameters.json
python3 docs/step-16-generics/audit.py /tmp/scout-base.jsonl /tmp/scout-parameters.json
python3 docs/step-16-generics/audit.py --test > /tmp/scout-audit.log 2>&1
python3 docs/step-16-generics/parameters_test.py "$HOME/.cache/adamic-stage3/api/node_modules/typescript" > /tmp/scout-parameters-test.log 2>&1
python3 docs/step-16-generics/parameters_test.py "$HOME/.cache/adamic-stage3/api/node_modules/typescript" --mutant > /tmp/scout-parameters-mutant.log 2>&1
python3 docs/step-16-generics/audit.py --mutant > /tmp/scout-audit-mutant.log 2>&1
python3 docs/step-16-generics/audit.py --mutant-selection > /tmp/scout-selection-mutant.log 2>&1
```

The parameters file is the sorted unique names of every TypeParameterDeclaration in the pinned compiler AST, parsed with TypeScript 6.0.3. The complete name list is retained in baseline.json.gz. The census SHA-256 pins the complete scratch measurement; selected findings remain reviewable without that scratch file.

## Design and disposition

The contract is [0.1.md](0.1.md): concrete monomorphization, nominal invariant classes, Node meaning, and unbounded polymorphic recursion refused. [escape-hatches.md](escape-hatches.md) requires proof at each cast, predicate and mutable view. A successful generic instantiation cannot convert a refused program into an accepted one by replacing a parameter with its constraint. The census contains attempts at generic declarations without a concrete caller; their free binders are evidence about representation demand, not proof that concrete calls fail.

| Shape | Representation and lowering | Disposition |
| --- | --- | --- |
| Direct `T` result, argument or local | Substitute the checker's concrete type throughout the body, then use its existing IR representation and ownership rules. Cache by declaration and checker type identities, including the nested frame owner. | Already lowers for concrete number and string calls. Unspecialized `T` remains NotYet; do not invent a type. |
| `T \| undefined`, callback result `U \| undefined`, optional containers | Read each declaration binder through the resolved call signature's actual mapper, then compose its target with the enclosing mapper. Choose the existing concrete optional representation after substitution. Absence and falsy presence must remain distinct. | Compiler lesson within the existing generic contract. The first build targets exactly a direct `T | undefined` result in an unconstrained generic declaration for represented concrete instantiations; optional containers and dependent members remain outside the mapper path. Unsupported concrete unions remain NotYet. |
| `ArenaIndex<Tag>` inside `Arena<Tag>.push` | Substitute the enclosing class argument before finding the inner class layout. Instantiate fields, constructor, methods and results together. Keep layouts and method targets keyed by checker identity, not only C storage. | The reduced nesting already lowers on this base. Add an ownership and identity regression. No class invariance change. |
| Generic interface, alias, array, tuple, map or conditional/indexed type inside another generic | Ask the checker to instantiate the whole type, including dependent arguments, before representation selection. Reuse the existing concrete container ABI; no runtime type parameter slot. | No universal erased `T` representation. A concrete unsupported container/member continues to say NotYet. Dependent readonly/mutable and nominal checks still apply. |
| Generic function assigned to a concrete callable slot | A specialization adapter needs a concrete callable ABI. A separate source identity must survive adapters, imports, aliases and repeated reads. Captured declarations additionally need the proper environment identity and lifetime. | Proposal below; generic function values stay NotYet in this build. |
| A genuinely polymorphic function value, callable at several types | Either whole-program flow discovers every concrete call and gives each one an adapter, or a proven tagged value ABI and explicit dictionary carries all required operations. Both must preserve source identity and ownership. | Proposal only. No unchecked erased callable or constraint fallback. |
| Recursion at the same concrete type | Register the specialization before lowering its body and reuse it on the recursive edge. | Existing finite monomorphization. |
| Ever-growing instantiations | Retain a compilation bound and diagnose exhaustion; never emit a partially instantiated body. | Existing refusal retained. Distinguishing finite deep nesting from truly unbounded recursion needs the proposal below. |
| Generic soundness refusals | Recheck instantiated mutation, optional presence, readonly-to-writable views, method variance, nominal bounds, casts and predicates using concrete checker types. | Remain refused unless their separate proof obligations are met. Hidden bytes do not authorize relaxing them. |

Several Refused rows print the same T on both sides, or relate never/undefined to a constrained binder. Spelling does not establish checker identity or inhabitation, especially on checker-rejected measurement input. Keep these refusals pending checker-clean reductions and binder-aware relation proofs. Generic overloads with additional implementation binders remain NotYet; result narrowing without an implementation proof remains Refused. This build does not reclassify either family.

The implementation should obtain concrete binders from the resolved signature, using the existing checker shim bridge rather than duplicating checker inference. Reading parameters/results structurally is incomplete: `GetNonNullableType(T)` may be `NonNullable<T>`, which is an intersection, and a type parameter may occur only inside a conditional, optional or object type. The mapper must describe this call, not a previous instantiation of the same declaration. If unavailable, retain conservative inference and NotYet at the first unsupported representation. Never silently substitute the upper bound.

The change belongs in `internal/lower/generic.go`; the existing shim bridge is in `internal/lower/instantiate.go`. It needs no new IR, C runtime allocation, backend entry point, or language policy. Existing concrete representations decide retains/releases, undefined tags and return ownership. Acceptance of an already-approved generic shape is a compiler capability change; changes to the normative refusal rules below are proposals for @system_adamic and are not implemented here.

## Proposals for @system_adamic, not implementation

**Generic function values and identity.** The current NotYet must remain until a value has both a concrete call adapter and stable source identity. Decide whether finite whole-program specialization covers the first release and how a polymorphic value escaping discovery is diagnosed. These currently unsupported programs are part of the proposal:

```a
function identity<T>(value: T): T { return value; }
const text: (value: string) => string = identity;
const number: (value: number) => number = identity;
const left: unknown = text;
const right: unknown = number;
console.log(`${left === right}`); // Node: true, despite different call ABIs.
```

```a
function identity<T>(value: T): T { return value; }
function pass<F>(value: F): F { return value; }
const polymorphic = pass(identity);
console.log(`${polymorphic(4)} ${polymorphic('x')}`);
```

A pointer per specialization would silently print false for the first program. A single untyped pointer would risk calling the wrong ABI in the second. An adapter must carry the stable declaration/environment identity; a closure allocation must not become a different function on each generic reference evaluation. The proposal includes aliasing through unknown, Map/Set keys, multiple imports, repeated reads, generic factories, captured environments and recursive callable fields.

**Monomorphization bounds.** The implementation has a depth limit of 32. The present diagnostic calls exhaustion "instantiated without end", but a finite chain can exhaust the same budget. Propose a separate resource-limit diagnostic and a specified configurable compilation budget, without accepting arbitrary polymorphic recursion or changing the limit in this build. The refused family is:

```a
function grow<T>(value: T, depth: number): number {
    if (depth === 0) return 0;
    return grow([value], depth - 1);
}
console.log(`${grow(1, 1)}`);
```

The runtime branch does not bound static monomorphization. A finite chain of more than 32 distinct generic calls is a separate counterexample to inferring infinity from this implementation guard. Proving or specializing such programs requires an explicit policy decision. Unbounded native code generation is not an escape hatch.

**Instantiated dependent result proofs.** Whole-signature mapper use would newly admit this checker-clean type lie. It is currently NotYet and remains so in this build:

```a
function two<T extends { readonly value: number }>(): T['value'] { return 2; }
const value: 1 = two<{ readonly value: 1 }>();
console.log(`${value}`); // Node prints 2 even though the declared type is 1.
```

Proposal for @system_adamic: recheck each concrete return and assignment relation under the resolved mapper before accepting dependent/indexed results. The concrete return above must be refused, not trusted because its generic constraint is number. Reclassifying this from NotYet to Refused is not implemented here. The first build therefore reads resolved maps only when the declared result is exactly a two-member union of a type parameter and undefined, and every declaration binder is unconstrained. Removing that guard makes the negative outcome fixture fail because the type lie is accepted; matching Node output alone cannot detect it. A mixed `T['value'] | U | undefined` result with `U = never` has the same problem. Requiring exactly two members prevents that dependent member from bypassing the guard; fixture 12 pins this boundary:

```a
function two<T extends { readonly value: number }, U>(): T['value'] | U | undefined { return 2; }
const value: 1 | undefined = two<{ readonly value: 1 }, never>();
console.log(`${value}`); // Node prints 2; the declared slot permits only 1 or undefined.
```

**Instantiated body relations.** Constrained bodies remain on their old inference path because a correct result ABI does not prove local types. Fixture 13 previously stayed NotYet because U could not be read back. Mapping its unrelated optional result would expose this checker-clean local type lie, so that new mapper path is withheld:

```a
function test<T extends { readonly value: number }, U>(): U | undefined {
    const value: T['value'] = 2;
    console.log(`${value}`);
    return undefined;
}
test<{ readonly value: 1 }, string>();
```

There is also a pre-existing accepted type lie, observed on both the exact base and delivery compiler. It is separate from this new capability and must not be mistaken for a proven program:

```a
function test<T extends { readonly value: number }>(): T | undefined {
    const value: T['value'] = 2;
    console.log(`${value}`);
    return undefined;
}
test<{ readonly value: 1 }>();
```

Both print 2 on Node while the instantiated local slot is declared 1. A nested variant that captures the constrained outer T also remains NotYet on both compilers: the base stops at U-or-undefined, and the delivery compiler reaches the unsupported T-indexed local without accepting it:

```a
function outer<T extends { readonly value: number }>(item: T): void {
    function inner<U>(): U | undefined {
        const value: T['value'] = 2;
        console.log(`${value}`);
        return undefined;
    }
    inner<string>();
}
outer<{ readonly value: 1 }>({value: 1});
```

Instantiated body relations must cover enclosing binders as well as declaration binders. Proposal for @system_adamic: prove every instantiated local initializer, assignment and return against its actual concrete type, and refuse this local initializer. That acceptance/refusal change is not implemented here. Removing the new unconstrained-declaration guard makes fixture 13 newly accepted; its negative assertion catches that mutation. A proposed unconstrained intersection-alias variant was independently rejected by TypeScript with TS2322, rather than treated as a valid counterexample.

**Custom mutation methods.** A sound custom `add(key, value)` inside a generic function is currently mistaken for the Set protocol `add(value)`. Preserve the existing refusal in this build. Proposal for @system_adamic: identify library mutation protocols by the resolved declaration, and prove custom method writes from their actual signature/body without weakening instantiated invariance. This refused checker-clean counterexample logs `key` on Node:

```a
class Bag<T> {
    add(key: string, value: T): void { console.log(key); }
}
function put<T>(bag: Bag<T>, value: T): void { bag.add('key', value); }
put(new Bag<number>(), 4);
```

The current diagnostic says a value of type "key" is written where number is read. The method writes no such slot. The probe is supplemental evidence, not a new tsc census count. The TypeScript multimap class reduction keeps construction inside a generic factory and invokes its custom mutation method from the concrete driver, preserving the original class body while isolating this separate classifier gap.

**Generic casts, predicates and mutable views.** No relaxation is proposed. These must continue to fail even after a concrete specialization is available:

```a
function manufacture<T>(value: unknown): T { return value as T; }
```

```a
interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: () => string; }
function poison<Pack extends Animal[]>(pack: Pack, animal: Animal): void {
    pack.push(animal);
}
const dogs: Dog[] = [];
poison(dogs, {name: 'cat'});
```

The first has no runtime proof of an arbitrary `T`. The second writes through a constraint that is wider than the instantiated mutable element. A readonly contract, a checked schema or a proven narrowing can satisfy a separate existing rule; specializing cannot supply the missing proof.

## Silent-miscompile audit

The following are distinct obligations; merely reaching C emission verifies none of them.

- Binder identity: alpha-renamed parameters, shadowed binders, defaults, explicit arguments, overload implementation signatures, dependent constraints and `T` inside intersections/conditionals must map to the actual call. Name equality is not type identity.
- Enclosing substitution: a nested generic function, generic method or class construction must retain both its own binder substitutions and the enclosing class/function environment. Restoring the previous mapper, locals and captures must also happen on failure.
- Cache identity: different structural members, literal unions, array mutability, nominal classes and callbacks can share a native representation. Keys must retain checker identity and nested owner; recursion must not reuse another specialization's body.
- Function identity and ABI: adapter identity, environment identity, parameter count/defaults, omitted arguments, optional results, primitive boxing and unboxing must agree between calls and stored callable slots.
- Optional results: distinguish undefined from zero, false, empty text and a present object. Do not treat `T`'s constraint or `NonNullable<T>` as a concrete binder, or erase the undefined tag.
- Ownership: dynamically allocated strings, arrays, class instances and callback environments returned as `T` must retain exactly the owner the caller receives. Optional absence owns nothing. Shared object identity must survive calls; captures and recursive graphs still need the existing cycle proof.
- Soundness: instantiate readonly/writable and optional-field relations, mutation through constraints, nominal bounds, predicate contracts, cast schemas and function variance before erasure. Both explicit and inferred type arguments need checks.
- Evaluation: side effects in arguments, class constructors, callback calls and optional branches execute once in source order. A specialization may not fold runtime generic operations just because a test uses literal inputs.
- Failure and limits: an unread binder, exhausted budget or unsupported member must produce Refused/NotYet, never reuse an unread-key body with an incompatible signature or leave a partially lowered function reachable.
- Census claims: a generic declaration attempted without arguments cannot be made concrete by the meter. Replayed roots, checker eligibility, first exposed reasons and hidden bytes require separate before/after evidence. No full tsc compilation follows from retiring one blocker.

## Reduced acceptance fixtures and baseline outcomes

The TypeScript reductions come from pristine upstream `050880ce59e30b356b686bd3144efe24f875ebc8`, verified against its own source, rather than from cohere. Every new Adamic source is `.a`. [fixture-baseline.json](step-16-generics/fixture-baseline.json) pins each source hash, exact reduction, independent Node stdout and the outcome on `8cb5e7c1`. The initial eight sources exit 0 with empty stderr on Node; the later recursive stress fixture does too. An innocent invocation does not make an arbitrary cast, a writable view or unbounded static monomorphization safe.

| Fixture in stage3/fixtures/generics | Pristine origin | Base outcome |
| --- | --- | --- |
| 01_identity.a | core.ts:1828 identity; direct body plus enclosing relay | Lowered; both backends match Node, sanitizer and leak checks pass |
| 02_optional_return.a | core.ts:67 firstDefined; computed callback result becomes input | NotYet: a function returning U \| undefined |
| 03_callback_return.a | core.ts:67 firstDefined; one element replaces iteration | NotYet: a function returning U \| undefined |
| 04_function_value.a | core.ts:220 contains; core.ts:1937 equateValues | NotYet: a generic function as a value |
| 05_nested_class.a | Supplemental bxpmash Arena/ArenaIndex regression, rather than a tsc class reduction | Lowered; both backends match Node, sanitizer and leak checks pass |
| 06_polymorphic_recursion.a | Supplemental 0.1 growing-type recursion boundary | Refused: polymorphic recursion |
| 07_generic_cast.a | core.ts:1778 tryCast; arbitrary target cast replaces its unproved predicate contract | Refused: a cast the runtime can't check |
| 08_readonly_view.a | utilities.ts:10644 setTextRangePos | Refused: readonly field pos becomes writable |
| 09_recursive_optional.a | Supplemental recursion added to the core.ts:67 result reduction | Base NotYet: a function returning U \| undefined; final lowering regression |
| 10_identifier_multimap.a | transformers/utilities.ts:389 IdentifierNameMap and :441 IdentifierNameMultiMap; generic construction driver added | Already lowered on base; both backends match Node with sanitizer/leak checks |
| 11_indexed_result.a | Supplemental indexed-result type-lie boundary | NotYet on base and final compiler; Node prints 2 for a declared literal 1 |
| 12_mixed_indexed_result.a | Supplemental mixed indexed/optional result type-lie boundary | NotYet on base and final compiler; Node prints 2 for a declared literal 1 or undefined |
| 13_constrained_local.a | Supplemental constrained indexed-local type-lie boundary | NotYet on base and final compiler; unrelated optional U must not expose an unproved local T["value"] |
| 14_optional_literal_union.a | Supplemental unconstrained three-member union capability boundary | NotYet on base and final compiler; broader optional unions are outside this mapper build |
| 15_array_callback.a | core.ts:67 firstDefined; optional readonly array and loop retained; non-null assertion replaced with explicit checked failure | Base NotYet returning U or undefined; final Lowered with both backends, sanitizer and leak checks |

The oracle registry and `TestStep16GenericOutcomes` record the distinction between capability gaps and policy refusals. The numeric identity mutant changes the valid IR return from its argument to 17. It finishes normally with valid C, no sanitizer findings and no leaks, then fails Node stdout comparison in both backends. This prevents an emission-only regression test from passing a wrong generic result. Optional fixtures preserve zero, false and empty text as present values, exercise explicit undefined at two type arguments, dynamically allocate returned text, and route an enclosing type parameter through a callback result.

## Built optional generic returns

`internal/lower/generic.go` now seeds declaration binders only in unconstrained generic declarations with direct optional type-parameter results from the resolved call signature's mapper and composes its target with the active caller mapper before deciding whether it is concrete. An identity mapping in a recursive call must also compose: otherwise different concrete callers can share an unread cache key and incompatible ABIs. Structural read-back remains available when the shim does not expose a mapping. The existing class-call wrapper's already-composed concrete binder takes precedence: raw resolved targets can still name an outer binder after that wrapper has replaced the active mapper. This distinction is exercised by the existing forwarded `makePair`/`forwardPair` generic factory fixture.

Fixtures 02 and 03 now lower and match independent source Node through JavaScript and native C. Both retain the baseline source bytes and provenance; their baseline NotYet observations remain in fixture-baseline.json. The fixtures preserve zero, false and empty strings, explicit missing values, dynamically allocated text, and generic callback results inside an enclosing specialization. Counted allocations/frees are 5/5 and 8/8 respectively. No new IR representation or runtime allocation scheme was introduced.

The numeric identity mutant and the new optional-result mutant emit valid C, finish with empty stderr and no sanitizer/leak findings, and disagree with Node stdout in both backends. Disabling resolved mapping returns fixture 02 to the original U-or-undefined NotYet. Removing the result-shape guard accepts fixture 11's indexed type lie; omitting the two-member requirement accepts fixture 14's broader literal-union shape. Fixture 12's mixed type lie also stays blocked by the constrained-declaration guard. The negative assertions catch their respective mutation; the literal-union case proves the two-member capability boundary independently of the constraint guard. Removing the unconstrained-declaration requirement also accepts fixture 13's unproved indexed local, caught by its negative assertion. These mutations are restored before the final verification and census. During the broader mapper experiment, raw signature mapping also displaced the class wrapper's concrete binder, and the existing generic factory regression caught the resulting unread T. The final bounded path preserves the original class behavior.

Generic function values and unproved indexed/dependent results remain NotYet. Arbitrary target casts, readonly-to-writable views and growing polymorphic recursion retain their recorded refusals. The supplemental recursive optional fixture 09 exercises number and dynamically allocated string calls through the same recursive declaration. Its unchanged-base NotYet was measured with a Go overlay restoring the only changed production file, generic.go, from 8cb5e7c1. Skipping identity mappings reproduces an incompatible native ABI and is caught by the fixture's clang build. This is separate from the valid result mutants that reach Node output comparison.

The array/callback fixture 15 retains the optional readonly array and loop from core.ts:67. It checks zero, false, empty text, a dynamically allocated callback result after an absent result, an empty typed array and a missing array. The original non-null assertion becomes an explicit missing-element throw; the driver excludes undefined elements, so no broader source-domain equivalence is claimed. Both backends agree with its independent Node source under fresh sanitizers and leak checks. The three newly failing larger census contexts below remain tracked; this fixture does not prove their captured or union-shaped arrays supported.

Unspecialized generic declarations, richer unsupported concrete containers and unproved generic overload relations remain outside this build. The source identity and limit-policy proposals above remain unimplemented.

## Measured root retirement

The full guarded census was repeated over the identical adapted compiler corpus. Checker diagnostics, every root identity/eligibility/body range and every source hash were unchanged. Compare exact `(unit, kind, reason)` sets; a moved diagnostic does not retire a root. This is measurement-only, not whole-program compilation or runtime acceptance.

431 scoped blocker/root pairs disappeared across 384 distinct roots. 76 of those roots have no remaining Refused/NotYet finding in the continuing measurement. The rest expose or retain other blockers. Neither number credits the historical hidden-byte estimates as measured native progress.

The complete affected root names, retired blockers, remaining blockers and newly exposed reasons are in [retirement.json.gz](step-16-generics/retirement.json.gz). The ledger also retains every newly exposed blocker across the corpus, including roots outside the selected baseline subset.

3 roots gained a failure where the baseline measurement had none. These are recorded below, rather than hidden in a net total. 0 roots that lost a Refused blocker became free of all findings; a disappearing refusal in a root that still fails is not acceptance of that program. The baseline and continuing corpus are checker-rejected, so an empty measurement finding set is not a valid-program or backend proof.

| Retired exact blocker | Roots |
| --- | ---: |
| NotYet: a function returning T \| T[] \| undefined | 4 |
| NotYet: a function returning T \| readonly T[] \| undefined | 5 |
| NotYet: a function returning T \| undefined | 79 |
| NotYet: a function returning U \| undefined | 24 |
| NotYet: a generic function as a value | 1 |
| NotYet: a value of type Children \| undefined | 18 |
| NotYet: a value of type CustomTransformerFactory \| TransformerFactory<T> | 2 |
| NotYet: a value of type IncludeTypeSpaceImports | 1 |
| NotYet: a value of type Map<string, SingleFileWatcher<T>> | 2 |
| NotYet: a value of type MapLike<T> | 2 |
| NotYet: a value of type NodeArray<T> | 1 |
| NotYet: a value of type NodeArray<T> \| undefined | 1 |
| NotYet: a value of type NonNullable<T> | 20 |
| NotYet: a value of type Set<K> | 2 |
| NotYet: a value of type SortedArray<T> | 1 |
| NotYet: a value of type T | 46 |
| NotYet: a value of type T \| T[] | 3 |
| NotYet: a value of type T \| readonly T[] | 3 |
| NotYet: a value of type T \| undefined | 78 |
| NotYet: a value of type T[] | 7 |
| NotYet: a value of type U | 5 |
| NotYet: a value of type readonly (readonly T[])[] | 1 |
| NotYet: a value of type readonly T[] | 10 |
| NotYet: a value of type readonly T[] \| undefined | 64 |
| NotYet: an array of T | 21 |
| NotYet: overload 1 of getOriginalNode with additional implementation type parameters | 2 |
| NotYet: overload 2 of arrayFrom with additional implementation type parameters | 10 |
| NotYet: overload 3 of group with additional implementation type parameters | 5 |
| Refused: a cast the runtime can't check | 6 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of any would replace | 2 |
| Refused: optional property return in ArrayIterator<any> absent from structural source ArrayIterator<Child>, which can hide fields | 1 |
| Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 |
| Refused: overload 1 of find result U \| undefined cannot be served by implementation result T \| undefined | 2 |
| Refused: overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined | 1 |

| Root with no remaining measurement finding | Name |
| --- | --- |
| binder.ts:1084:5 | bindEach |
| checker.ts:2800:5 | addDuplicateDeclarationErrorsForSymbols |
| commandLineParser.ts:2718:1 | filterSameAsDefaultInclude |
| commandLineParser.ts:4018:1 | isExcludedFile |
| core.ts:1966:1 | equateStringsCaseSensitive |
| emitter.ts:1395:5 | print |
| emitter.ts:1454:5 | emitIdentifierName |
| emitter.ts:1455:5 | emitIdentifierName |
| emitter.ts:1456:5 | emitIdentifierName |
| emitter.ts:1468:5 | emitJsxAttributeValue |
| emitter.ts:2032:5 | emitMappedTypeParameter |
| emitter.ts:2179:5 | emitQualifiedName |
| emitter.ts:2204:5 | emitTypeParameter |
| emitter.ts:2245:5 | emitPropertySignature |
| emitter.ts:2253:5 | emitPropertyDeclaration |
| emitter.ts:2263:5 | emitMethodSignature |
| emitter.ts:2300:5 | emitCallSignature |
| emitter.ts:2304:5 | emitConstructSignature |
| emitter.ts:2317:5 | emitTemplateTypeSpan |
| emitter.ts:2330:5 | emitTypePredicate |
| emitter.ts:2344:5 | emitTypeReference |
| emitter.ts:2349:5 | emitFunctionType |
| emitter.ts:2360:5 | emitFunctionTypeBody |
| emitter.ts:2365:5 | emitJSDocFunctionType |
| emitter.ts:2372:5 | emitJSDocNullableType |
| emitter.ts:2377:5 | emitJSDocNonNullableType |
| emitter.ts:2382:5 | emitJSDocOptionalType |
| emitter.ts:2387:5 | emitConstructorType |
| emitter.ts:2394:5 | emitTypeQuery |
| emitter.ts:2419:5 | emitRestOrJSDocVariadicType |
| emitter.ts:2431:5 | emitNamedTupleMember |
| emitter.ts:2469:5 | emitInferType |
| emitter.ts:2475:5 | emitParenthesizedType |
| emitter.ts:2551:5 | emitLiteralType |
| emitter.ts:2560:5 | emitImportTypeNode |
| emitter.ts:2746:5 | emitParenthesizedExpression |
| emitter.ts:2760:5 | emitArrowFunction |
| emitter.ts:2765:5 | emitArrowFunctionHead |
| emitter.ts:2971:5 | emitYieldExpression |
| emitter.ts:2992:5 | emitAsExpression |
| emitter.ts:3007:5 | emitSatisfiesExpression |
| emitter.ts:3017:5 | emitMetaProperty |
| emitter.ts:3027:5 | emitTemplateSpan |
| emitter.ts:3047:5 | emitVariableStatement |
| emitter.ts:3073:5 | emitIfStatement |
| emitter.ts:3093:5 | emitWhileClause |
| emitter.ts:3133:5 | emitForInStatement |
| emitter.ts:3146:5 | emitForOfStatement |
| emitter.ts:3345:5 | emitWithStatement |
| emitter.ts:3354:5 | emitSwitchStatement |
| emitter.ts:3364:5 | emitLabeledStatement |
| emitter.ts:3377:5 | emitTryStatement |
| factory/emitNode.ts:271:1 | removeEmitHelper |
| path.ts:538:1 | reducePathComponents |
| path.ts:606:1 | resolvePath |
| sys.ts:262:5 | watchFile |
| sys.ts:467:5 | watchFile |
| sys.ts:879:5 | isIgnoredPath |
| transformer.ts:218:1 | wrapScriptTransformerFactory |
| transformer.ts:222:1 | wrapDeclarationTransformerFactory |
| transformers/classThis.ts:90:1 | classHasClassThisAssignment |
| transformers/es2015.ts:1212:5 | isUninitializedVariableStatement |
| transformers/namedEvaluation.ts:156:1 | classHasExplicitlyAssignedName |
| transformers/ts.ts:873:5 | isClassLikeDeclarationWithTypeScriptSyntax |
| transformers/utilities.ts:124:1 | containsDefaultReference |
| transformers/utilities.ts:729:1 | getAllDecoratorsOfAccessors |
| transformers/utilities.ts:762:1 | getAllDecoratorsOfMethod |
| transformers/utilities.ts:781:1 | getAllDecoratorsOfProperty |
| transformers/utilities.ts:871:1 | isSimpleParameterList |
| utilities.ts:1351:1 | indexOfNode |
| utilities.ts:2751:1 | isHoistedVariableStatement |
| utilities.ts:7617:1 | isExportDefaultSymbol |
| utilities.ts:8609:1 | createDetachedDiagnostic |
| utilities.ts:8682:1 | createFileDiagnostic |
| utilities.ts:8705:1 | formatMessage |
| utilities.ts:8716:1 | createCompilerDiagnostic |

| Newly blocked root | Name | Exact findings |
| --- | --- | --- |
| checker.ts:9080:9 | serializeInferredTypeForDeclaration | NotYet: a value of type readonly T[] \| undefined |
| checker.ts:9285:9 | canReuseTypeNode | NotYet: a value of type (element: Node) => "quit" \| boolean<br>NotYet: a value of type readonly T[] \| undefined |
| checker.ts:9550:13 | mergeRedundantStatements | NotYet: a value of type readonly T[] \| undefined |

## Rebuild on current main

The source scout and its census above are historical, pinned to their named inputs.
The current rebuild is documented in [landing-report.md](step-16-generics/landing-report.md)
and [landing-records.json](step-16-generics/landing-records.json). Main 7a10c877
already admits every supported scout shape, including fixture 14. The redundant
resolved-mapper addition is omitted. No compiler or runtime implementation changes
are included in this rebuild. The historical retirement totals are not remeasured
credits on main. Generic body relations remain compiler/generic-body-relations.
