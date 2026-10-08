# Step 17: counted and uncounted unions

## Measured census and scope

Delivery base: `4885cec50290686df487b62aac47c85d871ed40c`, the fetched
`compiler/area-next-fixtures` tip. Census/report pin:
`ec0b16c04f3bbdfbaf3932ab01307f90b08156fd`; its measured compiler is
`69501280a81259fb512edbb8dd0e52c6eb0d88c8`. The census is not a fresh
measurement of the delivery base. Its 82 adapted files retain the original
hashes. The input is checker-rejected and continuing after errors.

Observation: the ledger below retains every exact Refused/NotYet reason whose
text mentions union, null, undefined, optional, presence or differently held
values, including zero-credit reasons. This intentionally broad inventory keeps
neighbouring blockers visible. A named alias without those words can hide a
union; its complete semantic classification is not available from this table.
Generic T/U/V/K arms require specialization before they establish a counted arm.
Intersections with an undefined-valued field are not themselves mixed unions.
No row is claimed retired merely because its text belongs to this inventory.

Roots are inferred implementation obligations, not measured ownership or a
language ruling. Refused overload families belong to overload-results; mutable
widening, readonly-to-writable and unproven optional compatibility remain soundness
refusals. Checked-view rows additionally depend on checked boundary validation.
Optional call rows additionally depend on closure/argument evaluation. Object-only
unions are neighbours, not automatically step 17 mixed unions.

The hidden column copies `bytes_revealed_if_fixed_alone`: outermost-cause
attribution, an estimate of bytes a census might next examine. It is neither
measured revealed bytes nor an implementation counterfactual. Nested spans never
receive a second credit. The boundaries column counts distinct blocked spans.
Examples are up to three distinct actual diagnostic or caller-attempt positions,
including shadowed boundaries; these are not three independently credited roots.
If fewer sites exist, the table says so rather than inventing witnesses.
All exact rows, original contributing examples and provenance are retained in
[the machine ledger](step-17-unions/inventory.json).

## Exact reasons

| Rank | Kind and reason | Hidden attribution bytes | Boundaries | Diagnostic or attempt witnesses | Root obligation |
| ---: | --- | ---: | ---: | --- | --- |
| 19 | NotYet: a function returning T \| undefined | 24,802 | 87 | checker.ts:6626:9<br>checker.ts:25177:5<br>checker.ts:25222:5 | Concrete representation / element and field storage |
| 27 | Refused: overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram \| undefined | 12,027 | 5 | builder.ts:1663:1<br>builder.ts:1668:1<br>builder.ts:1673:1 | Existing soundness relation; not representation permission |
| 40 | NotYet: a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | 6,424 | 21 | builder.ts:1597:9<br>builder.ts:1584:1<br>builder.ts:1608:9 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 41 | NotYet: a value of type string \| null | 6,421 | 2 | sourcemap.ts:36:1<br>sourcemap.ts:102:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 42 | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 6,371 | 69 | checker.ts:5667:9<br>checker.ts:5664:5<br>checker.ts:6583:13 | Existing soundness relation; not representation permission |
| 46 | NotYet: a function returning U \| undefined | 5,575 | 10 | core.ts:33:1<br>core.ts:50:1<br>core.ts:67:1 | Concrete representation / element and field storage |
| 48 | NotYet: a call through ?. (an optional call) | 5,095 | 151 | builder.ts:670:13<br>builder.ts:635:1<br>builder.ts:710:5 | Optional evaluation or physical own-field presence |
| 57 | NotYet: a function returning __String \| undefined | 3,489 | 26 | binder.ts:661:5<br>binder.ts:916:17<br>binder.ts:887:5 | Concrete representation / element and field storage |
| 70 | Refused: overload 1 of getSuperContainer result SuperContainer \| undefined cannot be served by implementation result SuperContainerOrFunctions \| undefined | 2,597 | 5 | checker.ts:31633:9<br>checker.ts:31630:5<br>checker.ts:31642:17 | Existing soundness relation; not representation permission |
| 75 | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined | 2,411 | 41 | binder.ts:1688:13<br>binder.ts:1639:5<br>builder.ts:1042:5 | Existing soundness relation; not representation permission |
| 81 | NotYet: a value of type T \| null \| undefined | 2,252 | 53 | builder.ts:2478:9<br>builder.ts:2452:1<br>builder.ts:2477:5 | Concrete representation / element and field storage |
| 82 | NotYet: an optional chain longer than one step | 2,229 | 12 | checker.ts:4742:9<br>checker.ts:4730:5<br>checker.ts:36469:13 | Optional evaluation or physical own-field presence |
| 83 | Refused: overload 1 of find result U \| undefined cannot be served by implementation result T \| undefined | 2,207 | 39 | checker.ts:9053:13<br>checker.ts:9052:9<br>checker.ts:9420:17 | Existing soundness relation; not representation permission |
| 87 | NotYet: ?. to a number, which would be number \| undefined | 2,099 | 4 | checker.ts:10416:25<br>checker.ts:10395:13<br>checker.ts:10445:25 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 88 | Refused: overload 1 of visitFunctionBody result Block cannot be served by implementation result ConciseBody \| undefined | 2,087 | 8 | transformers/es2017.ts:657:9<br>transformers/es2017.ts:649:5<br>visitorPublic.ts:508:1 | Existing soundness relation; not representation permission |
| 96 | NotYet: a value of type T \| undefined | 1,837 | 22 | checker.ts:33507:21<br>checker.ts:33495:5<br>core.ts:921:1 | Concrete representation / element and field storage |
| 102 | NotYet: a function returning PackageJson[K] \| undefined | 1,663 | 4 | moduleNameResolver.ts:359:1<br>moduleNameResolver.ts:360:1<br>moduleNameResolver.ts:361:1 | Concrete representation / element and field storage |
| 105 | Refused: overload 1 of sameMap result U[] cannot be served by implementation result readonly U[] \| undefined | 1,633 | 30 | builder.ts:526:5<br>builder.ts:521:1<br>builder.ts:564:5 | Existing soundness relation; not representation permission |
| 107 | Refused: overload 1 of findAncestor result T \| undefined cannot be served by implementation result Node \| undefined | 1,603 | 26 | checker.ts:2015:9<br>checker.ts:1963:5<br>checker.ts:1986:9 | Existing soundness relation; not representation permission |
| 119 | NotYet: a value of type string \| null \| undefined | 1,382 | 6 | sourcemap.ts:701:5<br>sourcemap.ts:699:1<br>sys.ts:916:5 | Concrete representation / element and field storage |
| 120 | NotYet: a function returning Path \| undefined | 1,378 | 2 | resolutionCache.ts:258:1<br>resolutionCache.ts:468:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 122 | NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1,351 | 1 | transformers/namedEvaluation.ts:322:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 123 | NotYet: a value of type ResolvedConfigFileName \| undefined | 1,346 | 1 | tsbuildPublic.ts:1393:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 130 | NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1,299 | 1 | transformers/namedEvaluation.ts:354:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 138 | NotYet: checked view field modifiers of type NodeArray<ModifierLike> \| undefined | 1,183 | 76 | binder.ts:359:13<br>binder.ts:350:1<br>binder.ts:2984:17 | Checked-view validation and stored/observed conversion |
| 160 | Refused: overload 1 of combine result T[] \| undefined cannot be served by implementation result T \| T[] \| undefined | 975 | 4 | core.ts:947:1<br>core.ts:949:1<br>core.ts:951:1 | Existing soundness relation; not representation permission |
| 166 | Refused: overload 2 of getParseTreeNode result T \| undefined cannot be served by implementation result Node \| undefined | 941 | 21 | checker.ts:2041:9<br>checker.ts:2040:5<br>checker.ts:50597:9 | Existing soundness relation; not representation permission |
| 170 | Refused: overload 1 of convertOptionsFromJson result WatchOptions \| undefined cannot be served by implementation result CompilerOptions \| TypeAcquisition \| WatchOptions \| undefined | 933 | 3 | commandLineParser.ts:3775:1<br>commandLineParser.ts:3776:1<br>commandLineParser.ts:3777:1 | Existing soundness relation; not representation permission |
| 174 | Refused: overload 1 of singleOrMany result T \| T[] cannot be served by implementation result T \| readonly T[] \| undefined | 909 | 20 | core.ts:1152:1<br>core.ts:1154:1<br>core.ts:1156:1 | Existing soundness relation; not representation permission |
| 178 | NotYet: a function returning AnyValidImportOrReExport \| undefined | 864 | 1 | utilities.ts:4381:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 181 | Refused: overload 1 of findLast result U \| undefined cannot be served by implementation result T \| undefined | 842 | 6 | checker.ts:3665:9<br>checker.ts:3664:5<br>core.ts:178:1 | Existing soundness relation; not representation permission |
| 228 | NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 589 | 1 | factory/utilities.ts:1673:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 234 | NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 533 | 4 | transformers/ts.ts:1201:17<br>transformers/ts.ts:1197:5<br>transformers/typeSerializer.ts:201:9 | Concrete representation / element and field storage |
| 235 | NotYet: a union of differently held members variable a function value captures | 524 | 28 | checker.ts:52365:25<br>checker.ts:52358:5<br>emitter.ts:827:13 | Captured-cell representation |
| 242 | NotYet: a tuple element of type string \| number \| boolean \| readonly string[] \| SourceFile \| undefined | 509 | 2 | watch.ts:291:9<br>watch.ts:276:1<br>watch.ts:294:9 | Concrete representation / element and field storage |
| 285 | NotYet: a narrowed scalar in a boxed union field | 377 | 7 | binder.ts:3194:13<br>binder.ts:3188:5<br>checker.ts:39374:13 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 302 | NotYet: optional chaining to .size on a value | 341 | 20 | builder.ts:705:5<br>builder.ts:700:1<br>builder.ts:756:5 | Optional evaluation or physical own-field presence |
| 307 | NotYet: writing a possibly absent optional own field | 329 | 16 | builder.ts:1962:21<br>builder.ts:1909:5<br>checker.ts:2529:9 | Optional evaluation or physical own-field presence |
| 325 | NotYet: checked view field modifiers of type NodeArray<Modifier> \| undefined | 300 | 27 | binder.ts:894:17<br>binder.ts:887:5<br>binder.ts:926:17 | Checked-view validation and stored/observed conversion |
| 336 | NotYet: checked view field textSourceNode of type BigIntLiteral \| Identifier \| JsxNamespacedName \| NumericLiteral \| PrivateIdentifier \| StringLiteralLike \| undefined | 280 | 7 | checker.ts:27759:13<br>checker.ts:27757:5<br>emitter.ts:5270:9 | Checked-view validation and stored/observed conversion |
| 339 | NotYet: in when a possibly absent spread has no property presence descriptors | 277 | 11 | checker.ts:2575:13<br>checker.ts:2567:5<br>checker.ts:2578:9 | Optional evaluation or physical own-field presence |
| 342 | NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 273 | 1 | moduleSpecifiers.ts:880:5<br>moduleSpecifiers.ts:879:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 349 | NotYet: an object with a nullable field viewed as unknown or object (null and undefined slot tags) | 260 | 2 | builder.ts:295:5<br>builder.ts:291:1<br>builder.ts:296:5 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 353 | NotYet: a number \| undefined argument to substring | 255 | 2 | builder.ts:1617:5<br>builder.ts:1616:1<br>sourcemap.ts:378:5 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 365 | NotYet: a value of type false \| RegExpExecArray \| null | 235 | 2 | parser.ts:10715:5<br>parser.ts:10714:1<br>parser.ts:10753:5 | Concrete representation / element and field storage |
| 370 | NotYet: a function returning HasJSDoc \| undefined | 232 | 2 | utilities.ts:4812:1<br>utilities.ts:12017:21<br>utilities.ts:12013:5 | Concrete representation / element and field storage |
| 371 | NotYet: checked view field initializer of type ForInitializer \| undefined | 231 | 17 | binder.ts:1132:17<br>binder.ts:1096:5<br>checker.ts:49225:17 | Checked-view validation and stored/observed conversion |
| 381 | NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 216 | 1 | core.ts:1975:5<br>core.ts:1971:1<br>core.ts:1973:1 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 386 | NotYet: checked view field multiLine of type boolean \| undefined | 210 | 17 | binder.ts:1833:13<br>binder.ts:1818:5<br>binder.ts:3003:17 | Checked-view validation and stored/observed conversion |
| 401 | NotYet: a function returning TEntry \| undefined | 191 | 1 | transformers/utilities.ts:829:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 417 | NotYet: a value of type __String \| undefined | 174 | 32 | binder.ts:755:9<br>binder.ts:749:5<br>checker.ts:14421:13 | Concrete representation / element and field storage |
| 431 | NotYet: a boolean \| undefined variable a function value captures | 164 | 157 | core.ts:36:13<br>checker.ts:8460:9<br>core.ts:35:48 | Captured-cell representation |
| 437 | Refused: an unproven relation from Identifier to Identifier: optional field id has no proven compatible presence/type | 160 | 1 | utilitiesPublic.ts:932:5<br>utilitiesPublic.ts:931:1 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 449 | Refused: an unproven relation from LiteralLikeNode to TemplateLiteralLikeNode: optional field rawText has no proven compatible presence/type | 152 | 1 | utilities.ts:2010:13<br>utilities.ts:1980:1 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 450 | Refused: an unproven relation from Type to TypeParameter: optional field constraint has no proven compatible presence/type | 152 | 14 | checker.ts:5568:9<br>checker.ts:5567:5<br>checker.ts:7164:21 | Existing soundness relation; not representation permission |
| 456 | NotYet: checked view field _propertyAccessExpressionLikeQualifiedNameBrand of type void \| undefined | 150 | 2 | parser.ts:9589:21<br>parser.ts:9584:13<br>transformers/typeSerializer.ts:615:9 | Checked-view validation and stored/observed conversion |
| 459 | Refused: an unproven relation from SolutionBuilderHostBase<T> to SolutionBuilderHost<T>: optional field reportErrorSummary has no proven compatible presence/type | 145 | 1 | tsbuildPublic.ts:313:5<br>executeCommandLine.ts:812:1<br>tsbuildPublic.ts:306:1 | Existing soundness relation; not representation permission |
| 464 | NotYet: a value of type boolean \| V \| undefined | 142 | 1 | utilities.ts:942:9<br>utilities.ts:930:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 468 | Refused: a type argument makes a value of type (left: MappedPosition, right: MappedPosition) => boolean seen as EqualityComparer<SourceMappedPosition> \| undefined, which can write string \| undefined where string is read | 141 | 1 | sourcemap.ts:763:13<br>sourcemap.ts:699:1<br>sourcemap.ts:754:5 | Existing soundness relation; not representation permission |
| 470 | NotYet: a value of type (AmbientModuleDeclaration & { name: StringLiteral; }) \| undefined | 138 | 1 | moduleSpecifiers.ts:920:5<br>moduleSpecifiers.ts:879:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 499 | NotYet: checked view field _optionalChainBrand of type void | 110 | 5 | checker.ts:34741:9<br>checker.ts:34740:5<br>checker.ts:35529:9 | Checked-view validation and stored/observed conversion |
| 506 | NotYet: a value of type V \| undefined | 106 | 2 | utilities.ts:939:9<br>utilities.ts:930:1<br>utilities.ts:941:9 | Concrete representation / element and field storage |
| 510 | Refused: an unproven relation from ImportEqualsDeclaration & { moduleReference: ExternalModuleReference; } to ImportEqualsDeclaration: optional field moduleReference.id has no proven compatible presence/type | 103 | 1 | utilities.ts:3780:5<br>utilities.ts:3778:1 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 511 | NotYet: checked view field modifiers of type NodeArray<Modifier> \| NodeArray<ModifierLike> \| undefined | 102 | 10 | checker.ts:11910:13<br>checker.ts:11844:5<br>checker.ts:16335:17 | Checked-view validation and stored/observed conversion |
| 515 | NotYet: a function returning TPrivateEntry \| undefined | 100 | 1 | transformers/utilities.ts:855:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 525 | NotYet: a number \| undefined argument to slice | 94 | 2 | program.ts:751:13<br>program.ts:711:1<br>program.ts:2981:13 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 540 | NotYet: a value of type U \| readonly U[] \| undefined | 88 | 2 | core.ts:403:13<br>core.ts:399:1<br>core.ts:422:13 | Concrete representation / element and field storage |
| 555 | NotYet: checked view field constraint of type JSDocTypeExpression \| undefined | 78 | 3 | checker.ts:49168:17<br>checker.ts:49064:5<br>emitter.ts:1895:21 | Checked-view validation and stored/observed conversion |
| 558 | NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 76 | 1 | tsbuildPublic.ts:723:5<br>executeCommandLine.ts:812:1<br>tsbuildPublic.ts:340:1 | Concrete representation / element and field storage |
| 567 | NotYet: a value of type K \| undefined | 74 | 2 | core.ts:560:9<br>core.ts:551:1<br>core.ts:561:9 | Concrete representation / element and field storage |
| 571 | NotYet: a function returning TOut \| undefined | 73 | 1 | core.ts:1778:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 575 | NotYet: a BinaryExpression with a string and a union of differently held members | 72 | 1 | moduleNameResolver.ts:946:9<br>moduleNameResolver.ts:944:1 (fewer than three recorded sites) | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 578 | NotYet: checked view field awaitModifier of type AwaitKeyword \| undefined | 69 | 12 | checker.ts:28367:17<br>checker.ts:28361:5<br>checker.ts:49229:17 | Checked-view validation and stored/observed conversion |
| 603 | NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 60 | 5 | checker.ts:30773:13<br>checker.ts:30763:5<br>transformers/es2015.ts:1151:9 | Concrete representation / element and field storage |
| 606 | NotYet: a template interpolating an object, an array, a map, a function or undefined | 59 | 3 | core.ts:1786:5<br>checker.ts:32344:5<br>debug.ts:226:13 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 616 | NotYet: storing Path \| undefined in a field | 55 | 1 | builder.ts:668:13<br>builder.ts:635:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 646 | NotYet: a value of type U \| undefined | 45 | 2 | core.ts:498:13<br>builderState.ts:193:5<br>core.ts:484:9 | Concrete representation / element and field storage |
| 652 | NotYet: checked view field id of type number \| undefined | 41 | 38 | binder.ts:3001:17<br>binder.ts:2846:5<br>binder.ts:3074:21 | Checked-view validation and stored/observed conversion |
| 657 | NotYet: a value of type HasJSDoc \| undefined | 38 | 7 | binder.ts:2538:13<br>binder.ts:2526:5<br>checker.ts:4652:9 | Concrete representation / element and field storage |
| 675 | NotYet: JSON.stringify a union containing containers without runtime element metadata | 0 | 1 | commandLineParser.ts:2960:9<br>commandLineParser.ts:2956:5 (fewer than three recorded sites) | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 676 | NotYet: a BinaryExpression with a number \| undefined and a number | 0 | 2 | checker.ts:25061:29<br>checker.ts:25022:5<br>checker.ts:25064:29 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 680 | NotYet: a Map of VisitResult<ExportAssignment \| LateVisibilityPaintedStatement \| undefined> | 0 | 3 | transformers/declarations.ts:987:17<br>transformers/declarations.ts:953:5<br>transformers/declarations.ts:977:13 | Concrete representation / element and field storage |
| 692 | NotYet: a function returning ClassStaticBlockDeclaration \| Decorator \| PrivateIdentifierGetAccessorDeclaration \| ... 5 more ... \| undefined | 0 | 2 | checker.ts:47028:5<br>checker.ts:47068:13<br>checker.ts:47056:5 | Concrete representation / element and field storage |
| 697 | NotYet: a function returning InferenceContext \| (T & undefined) | 0 | 1 | checker.ts:26263:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 702 | NotYet: a function returning T \| EmptyStatement \| undefined | 0 | 3 | factory/nodeFactory.ts:7161:5<br>factory/nodeFactory.ts:7162:5<br>factory/nodeFactory.ts:7163:5 | Concrete representation / element and field storage |
| 707 | NotYet: a function returning TypeMapper \| (T & undefined) | 0 | 1 | checker.ts:26378:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 708 | NotYet: a function returning TypeOnlyAliasDeclaration \| undefined | 0 | 3 | checker.ts:4455:5<br>checker.ts:30703:17<br>checker.ts:30676:5 | Concrete representation / element and field storage |
| 714 | NotYet: a template interpolating a union with an object, an array, a map or a function in it | 0 | 1 | core.ts:1786:5<br>commandLineParser.ts:2206:1 (fewer than three recorded sites) | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 715 | NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 0 | 3 | transformers/classFields.ts:925:13<br>transformers/classFields.ts:908:5<br>transformers/classFields.ts:2705:13 | Concrete representation / element and field storage |
| 716 | NotYet: a value of type (ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }) \| undefined | 0 | 11 | checker.ts:3728:13<br>checker.ts:3708:5<br>checker.ts:3737:9 | Concrete representation / element and field storage |
| 718 | NotYet: a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined | 0 | 3 | checker.ts:6636:13<br>checker.ts:6555:9<br>checker.ts:8456:13 | Concrete representation / element and field storage |
| 719 | NotYet: a value of type (VariableDeclaration & { name: Identifier; }) \| undefined | 0 | 1 | transformers/jsx.ts:110:9<br>transformers/jsx.ts:92:1<br>transformers/jsx.ts:109:5 | Concrete representation / element and field storage |
| 729 | NotYet: a value of type Children \| undefined | 0 | 3 | emitter.ts:4663:5<br>emitter.ts:4675:5<br>emitter.ts:4679:5 | Concrete representation / element and field storage |
| 731 | NotYet: a value of type ClassNamedEvaluationHelperBlock \| undefined | 0 | 1 | transformers/classFields.ts:2184:13<br>transformers/classFields.ts:2098:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 732 | NotYet: a value of type ClassStaticBlockDeclaration \| Decorator \| PrivateIdentifierGetAccessorDeclaration \| ... 5 more ... \| undefined | 0 | 1 | checker.ts:44188:21<br>checker.ts:44163:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 733 | NotYet: a value of type ClassThisAssignmentBlock \| undefined | 0 | 1 | transformers/classFields.ts:2183:13<br>transformers/classFields.ts:2098:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 736 | NotYet: a value of type DiagnosticRelatedInformation[][] \| readonly (DiagnosticRelatedInformation \| readonly DiagnosticRelatedInformation[] \| undefined)[] where an array goes | 0 | 1 | core.ts:378:9<br>checker.ts:36540:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 737 | NotYet: a value of type Diagnostic[][] \| readonly (Diagnostic \| readonly Diagnostic[] \| undefined)[] where an array goes | 0 | 1 | core.ts:378:9<br>program.ts:2923:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 743 | NotYet: a value of type Extension[][] \| readonly (readonly Extension[] \| Extension \| undefined)[] where an array goes | 0 | 1 | core.ts:378:9<br>program.ts:3432:5<br>utilities.ts:9972:1 | Concrete representation / element and field storage |
| 754 | NotYet: a value of type Path \| undefined | 0 | 2 | program.ts:2697:9<br>program.ts:2688:5<br>program.ts:3615:9 | Concrete representation / element and field storage |
| 759 | NotYet: a value of type RedirectsCacheKey \| undefined | 0 | 1 | moduleNameResolver.ts:1031:9<br>moduleNameResolver.ts:1030:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 764 | NotYet: a value of type T \| T[] \| readonly T[] \| undefined | 0 | 1 | core.ts:378:9<br>core.ts:375:1 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 771 | NotYet: a value of type TypeOnlyAliasDeclaration \| undefined | 0 | 1 | checker.ts:3353:17<br>checker.ts:3298:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 774 | NotYet: a value of type VariableDeclaration[][] \| readonly (readonly VariableDeclaration[] \| VariableDeclaration \| undefined)[] where an array goes | 0 | 1 | core.ts:378:9<br>transformers/declarations.ts:1779:5 (fewer than three recorded sites) | Concrete representation / element and field storage |
| 775 | NotYet: a value of type false \| TypeOnlyAliasDeclaration \| undefined | 0 | 3 | checker.ts:4352:9<br>checker.ts:4351:5<br>checker.ts:4446:9 | Concrete representation / element and field storage |
| 776 | NotYet: a value of type string[][] \| readonly (string \| readonly string[] \| undefined)[] where an array goes | 0 | 1 | core.ts:378:9<br>moduleNameResolver.ts:813:1<br>moduleSpecifiers.ts:1345:1 | Concrete representation / element and field storage |
| 785 | NotYet: an array of boolean \| undefined | 0 | 6 | binder.ts:1933:17<br>binder.ts:1912:5<br>binder.ts:2005:13 | Concrete representation / element and field storage |
| 788 | NotYet: apply without a dense argument literal (length, presence and argument representations must be proven) | 0 | 1 | sourcemap.ts:315:13<br>sourcemap.ts:313:5 (fewer than three recorded sites) | Optional evaluation or physical own-field presence |
| 797 | NotYet: checked view field asteriskToken of type AsteriskToken \| undefined | 0 | 8 | utilities.ts:2929:17<br>checker.ts:21234:5<br>utilities.ts:2933:17 | Checked-view validation and stored/observed conversion |
| 799 | NotYet: checked view field dotDotDotToken of type Token<SyntaxKind.DotDotDotToken> \| undefined | 0 | 3 | checker.ts:17733:17<br>checker.ts:17726:5<br>checker.ts:49158:17 | Checked-view validation and stored/observed conversion |
| 802 | NotYet: checked view field expression of type Expression \| undefined | 0 | 9 | utilities.ts:2899:17<br>checker.ts:21224:5<br>checker.ts:39416:5 | Checked-view validation and stored/observed conversion |
| 806 | NotYet: checked view field jsDocPropertyTags of type readonly JSDocPropertyLikeTag[] \| undefined | 0 | 1 | emitter.ts:1860:21<br>emitter.ts:1556:5 (fewer than three recorded sites) | Checked-view validation and stored/observed conversion |
| 807 | NotYet: checked view field label of type Identifier \| undefined | 0 | 6 | emitter.ts:1713:21<br>emitter.ts:1556:5<br>emitter.ts:1715:21 | Checked-view validation and stored/observed conversion |
| 811 | NotYet: checked view field modifiers of type undefined | 0 | 2 | checker.ts:7377:25<br>checker.ts:6769:9<br>checker.ts:7364:13 | Checked-view validation and stored/observed conversion |
| 813 | NotYet: checked view field name of type JSDocNameReference \| undefined | 0 | 1 | emitter.ts:1899:21<br>emitter.ts:1556:5 (fewer than three recorded sites) | Checked-view validation and stored/observed conversion |
| 841 | NotYet: checked view field readonlyToken of type MinusToken \| PlusToken \| ReadonlyKeyword \| undefined | 0 | 3 | checker.ts:13083:25<br>checker.ts:13037:5<br>checker.ts:49208:17 | Checked-view validation and stored/observed conversion |
| 845 | NotYet: checked view field typeArguments of type NodeArray<TypeNode> \| undefined | 0 | 1 | checker.ts:36755:17<br>checker.ts:36540:5 (fewer than three recorded sites) | Checked-view validation and stored/observed conversion |
| 846 | NotYet: checked view field typeExpression of type JSDocTypeExpression \| undefined | 0 | 1 | parser.ts:9711:29<br>parser.ts:9705:13 (fewer than three recorded sites) | Checked-view validation and stored/observed conversion |
| 848 | NotYet: checked view field typeParameters of type readonly JSDocTemplateTag[] \| undefined | 0 | 1 | emitter.ts:1862:21<br>emitter.ts:1556:5 (fewer than three recorded sites) | Checked-view validation and stored/observed conversion |
| 851 | NotYet: for...of over a union of differently held members | 0 | 1 | core.ts:2173:5<br>parser.ts:2364:5<br>scanner.ts:1023:1 | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 855 | NotYet: null comparison with a scalar | 0 | 1 | debug.ts:250:9<br>parser.ts:8915:13 (fewer than three recorded sites) | Operation on tagged values (narrowing, conversion, comparison or iteration) |
| 1681 | Refused: an unproven relation from Declaration to NamedDeclaration: optional field name has no proven compatible presence/type | 0 | 1 | checker.ts:26159:17<br>checker.ts:26104:5 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 1684 | Refused: an unproven relation from InstantiableType \| UnionOrIntersectionType to TypeParameter: optional field constraint has no proven compatible presence/type | 0 | 1 | checker.ts:15149:9<br>checker.ts:15148:5 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 1685 | Refused: an unproven relation from ObjectType to AnonymousType: optional field target has no proven compatible presence/type | 0 | 2 | checker.ts:16568:17<br>checker.ts:16561:5<br>checker.ts:20967:9 | Existing soundness relation; not representation permission |
| 1686 | Refused: an unproven relation from Type to SyntheticDefaultModuleType: optional field syntheticType has no proven compatible presence/type | 0 | 2 | checker.ts:37997:13<br>checker.ts:37994:5<br>checker.ts:38009:13 | Existing soundness relation; not representation permission |
| 1687 | Refused: an unproven relation from Type to TypeVariable: optional field constraint has no proven compatible presence/type | 0 | 1 | checker.ts:23898:21<br>checker.ts:23544:9 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 1688 | Refused: an unproven relation from UnionOrIntersectionType to UnionType: optional field resolvedReducedType has no proven compatible presence/type | 0 | 1 | checker.ts:23160:17<br>checker.ts:23135:9 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 1695 | Refused: overload 1 of evaluate result EvaluatorResult<string \| undefined> cannot be served by implementation result EvaluatorResult<string \| number \| undefined> | 0 | 3 | utilities.ts:11320:5<br>utilities.ts:11321:5<br>utilities.ts:11322:5 | Existing soundness relation; not representation permission |
| 1701 | Refused: overload 1 of parseModifiers result NodeArray<Modifier> \| undefined cannot be served by implementation result NodeArray<ModifierLike> \| undefined | 0 | 13 | parser.ts:3957:9<br>parser.ts:3955:5<br>parser.ts:4044:9 | Existing soundness relation; not representation permission |
| 1704 | Refused: overload 1 of parseOptionalToken result Token<TKind> cannot be served by implementation result Node \| undefined | 0 | 21 | parser.ts:4069:9<br>parser.ts:4034:5<br>parser.ts:4035:5 | Existing soundness relation; not representation permission |
| 1705 | Refused: overload 1 of parseOptionalTokenJSDoc result Token<TKind> cannot be served by implementation result Node \| undefined | 0 | 2 | parser.ts:2531:5<br>parser.ts:2532:5 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 1711 | Refused: overload 2 of makeSerializePropertySymbol result (p: Symbol, isStatic: boolean, baseType: Type \| undefined) => T \| T[] cannot be served by implementation result (p: Symbol, isStatic: boolean, baseType: Type \| undefined) => T \| (T \| AccessorDeclaration)[] \| AccessorDeclaration | 0 | 2 | checker.ts:10787:13<br>checker.ts:10798:13 (fewer than three recorded sites) | Existing soundness relation; not representation permission |
| 1712 | Refused: overload 3 of getGlobalType result GenericType cannot be served by implementation result ObjectType \| undefined | 0 | 47 | checker.ts:17481:5<br>checker.ts:17482:5<br>checker.ts:17483:5 | Existing soundness relation; not representation permission |

## Presence dependency and port witnesses

The requested `nullish/PRESENCE.md` is at
`stage3/interface-downcasts/nullish/PRESENCE.md` on
`codex/views-lazy-admission` (`f7246408`), not at the branch root. It observes
native Object.keys listing a deleted field through a literal alias, although in
and hasOwn correctly report absence. Present-but-undefined is retained. Its
measured optional-field branch was `86b3fe99`; the held successor is
`codex/optional-field-write-2` (`7e7464e6`). These branches were read, not merged.

Selector gap 3 is `stage1/cohere/selector/gaps/3_optional_boolean.ts`;
values gap 2 is documented in `stage1/cohere/values/GAPS.md`. Selector gap 2
also witnesses `string | boolean | undefined`. The requested React IR hir-01
fixture has not been located in this base; a substitute is not claimed to be
that port's original witness. Current outcomes will be measured separately.

No compiler change, native fixture, counts change, or retired root is claimed
by this census document. Ranking compiler 69501280 already differs from this
base; no total in this table is advertised as today's remaining bytes.

## Tagged representation design

Observation: `ir.Union` already stores one counted-reference word. The word is
NULL for undefined, the immortal `adamic_null` sentinel for null, an immortal
boolean box for false/true, an owned number box for numbers, or the original
string/object/array/Map/closure pointer. `adamic_heap.kind` discriminates live
pointers; identity is the payload pointer, not a wrapper identity. This is the
proposed canonical mixed-union representation, reusing the existing runtime.
A zero pointer and the null sentinel must never be interchangeable within it.
NaN and signed zero retain numeric equality and conversion semantics.

All union slots use the reference ownership protocol, including scalar boxes.
Retain/release ignore NULL and immortal boxes/sentinels, count number boxes,
and count references through their existing heap kind. Scalar payloads have no
outgoing counted edges. Container destruction and cycle traversal inspect the
actual heap kind, not a static guess that all union arms are objects. Assignment
retains the incoming arm before releasing an aliased outgoing arm; replacing
object with number/undefined must release exactly the old owner. Number-to-union
boxing may allocate; no claim of allocation-free scalar union storage is made.

Fields store that same pointer in `adamic_value.reference`, and shape metadata
marks the slot as counted. Arrays, Map values, captured cells, parameters and
results need the same arm-preserving conversions. Lowering must load the stored
representation before narrowing to an observed scalar/reference arm. A flow
annotation cannot change the physical slot's representation. Checked views must
validate the observed tag and refinements and convert rather than reinterpret.
The JavaScript backend keeps the actual JS value; native representation details
must not create JS wrappers observable through equality, keys or JSON.

Optional presence is separate from the union's value tag and from initialization
readiness. Each property has physical own presence; absent reads as undefined,
while a present undefined retains its key. Writing undefined to `x?: T` remains
checker-refused under exactOptionalPropertyTypes unless T includes undefined.
An admitted write to `x?: T | undefined` establishes own presence; delete clears
presence and releases the old owner. Readiness distinguishes staged uninitialized
fields from initialized undefined. Object.keys/entries, hasOwn, in, spread and
JSON must observe the same shared physical presence through every alias and view.
JSON omits undefined object values but emits null for undefined array elements;
spread copies present undefined and skips absence. Fixed shapes may reserve slots
for optional fields but may not advertise them as own keys before a write.

This design does not take the held optional-field branch as a dependency merge.
Its alias enumeration and bounded hasOwn proofs remain dependencies to reconcile,
not reasons to weaken their admission checks.

## Silent miscompile audit

- Static union-arm metadata substituted for a dynamic tag can retain/release a
  scalar as a pointer, leak replaced references, or dereference the null sentinel.
- Dropping null/undefined discrimination changes strict equality, typeof,
  short-circuiting, optional calls, defaults, ??, String and JSON independently.
- Reading a narrowed field in its observed representation reinterprets a number
  box pointer as a double; stale alias/callback writes invalidate flow evidence.
- Object, array, Map and closure all look object-like to some JS predicates.
  A typeof object check alone does not establish their native layouts or types.
- Mutable widening and optional hidden fields remain type lies regardless of
  how accurately the union is stored; tags do not certify payload contracts.
- Captured cells, closure adapters, rest/default arguments and overload results
  can silently disagree about boxing even when direct calls work.
- Spread/reuse/region copies must preserve value tags, presence and readiness;
  retaining the same payload must preserve object identity and mutation visibility.
- Cycles through a union reference arm must reach the cycle finder; scalar arms
  cannot erase ownership obligations or be mistaken for an outgoing edge.
- Falsy zero, negative zero, NaN, false and empty strings are present values;
  truthiness cannot stand in for optional presence or initialization.
- Exception cleanup, early returns and panic paths must use actual ownership;
  single evaluation and retain-before-release matter for aliasing assignments.
- NaN packing optimizations require an independent collision proof and must not
  silently canonicalize user NaNs into undefined. Performance is unmeasured.

## Questions for @system_adamic, with held programs

No answer is assumed, and these refusals will remain in place.

1. May representation work discharge optional compatibility or writable variance?
   Example: `interface A { value: string } interface B { value: string | number }
   const a: A = {value: 'ok'}; const b: B = a; b.value = 1;`
   The existing invariant-mutable refusal is necessary: a.value is no longer a
   string. Recommendation: keep this refused; canonical tags alone prove nothing.
2. May a differently held object union narrow using typeof alone?
   Example: `function f(x: {a: number} | number[]): number {
   if (typeof x === 'object') return x instanceof Array ? x.length : x.a; }`
   Existing unsupported tag paths must remain NotYet until the predicate/layout
   evidence is implemented. This unit will not treat typeof object as Array proof.
3. Do optional writes expand all fixed-shape aliases, or only an origin-proven
   storage family? Example: `interface A { x?: number | undefined }
   const a: A = {}; a.x = undefined; console.log(Object.keys(a).join(','));`
   It must print x if admitted. The held branch supplies bounded origin proofs;
   this unit will not extend them to opaque/accessor/prototype-bearing objects.
4. Do union conversions authorize arbitrary source coercion protocols?
   Example: `const x: number | {toString: () => string} =
   {toString: () => 'x'}; console.log(`${x}`);`
   Existing object interpolation/conversion restrictions remain. Built-in scalar
   union storage does not decide acceptance of observable user conversion code.
5. May a union representation alone admit broader overload boundaries?
   Example: `function f(x: number): string; function f(x: number): string | number
   { return x; } console.log(f(1));` The implementation violates the overload's
   result promise. Per-overload specialization/checks belong to overload-results;
   visitNode remains on that unit's list.

Unresolved generic T | undefined stops need concrete specialization or constraint
proof before arm classification. The leading concrete storage shape in this
inventory is string | null (6,421 historically attributed bytes), followed by
string | null | undefined (1,382). Preserving these ordinary JS values requires
no new coercion, mutation, overload or widening permission. They are the proposed
focused implementation target, after recording current source outcomes.

## Probe outcomes on the delivery base

[Baseline log](step-17-unions/evidence/baseline.log) records original source Node,
then actual lowering outcomes. All six source programs exit 0 with empty stderr.
The tsc probes are reductions retaining sourcemap.ts's content arm types and
null tests, not claims of reproducing a full caller's census boundary. No code
was copied from cohere. Port reductions already in Adamic were preserved.

| Probe | Source Node stdout | Current compiler outcome |
| --- | --- | --- |
| selector-optional-boolean.a | true | NotYet: a field of type boolean \| undefined |
| values-optional-boolean.a | true | NotYet: a field of type boolean \| undefined |
| hir-optional-boolean.a | true | Both backends, ASan/UBSan, release and leaks agree |
| tsc-source-content.a | source; missing | NotYet: a value of type string \| null |
| tsc-nullish-content.a | string; source; object; null; undefined; undefined | NotYet: a value of type string \| null \| undefined |
| presence.a | false; []; false; true; [first]; true | Refused: delete; fixed shape |

The HIR original was found on `stage1-hir/wip` at `3714b319`,
`stage1/cohere/high_level_intermediate_representation/testdata/optional-boolean-gap.a`.
Its GAPS.md confirms the requested first HIR gap and records 74/1,465 graph impact.
That is the port's historical census, not newly retired graph bytes. HIR's
interface read works already; the two class-field reductions still stop. No
port workaround is removed by measuring these reductions.

The focused runner now pins each remaining stop rather than accepting any error.
It originally supplied a relative path to the oracle cache; that harness error
was corrected and the passing measurement uses absolute paths. These standalone
probes have not been registered as positive count fixtures; no allocation row is
claimed for a refused program. No inherited-field/delete policy was changed.

## Implemented concrete shape

`string | null` and `string | null | undefined` now use canonical `ir.Union`
storage. No runtime ABI was added. Null uses the existing immortal sentinel;
undefined uses NULL; a string remains its counted string pointer. Parameters,
locals, plain-object fields and array elements preserve those tags. `typeof`
uses the live tag instead of assigning one meaning to every missing pointer.
Native coalescing converts the retained non-null string after testing both tags.

A narrowed field is loaded once in its stored union representation and checked
before conversion to string. A stale alias write to null therefore takes the
existing union-member panic, in both backends. Explicit null/undefined comparisons,
typeof and ?? observe the stored tag even after a stale checker narrowing. They
must not panic merely to ask what value is there. Empty strings stay present.

The fixture exercises heap strings, repeated alias writes, replacement with null
and undefined, coalescing, identity, spread, array reads/writes/push, missing entries
and stale-narrowing observations. All successful runs agree byte for byte with
source Node, including stderr and exit, under ASan/UBSan, release and leak checks.
A separate checked fixture demonstrates the stale field-use panic; source Node
continues and prints null. Existing refusal tests for nullable number/boolean,
unknown reflection and unsound relations pass. Class fields with mixed storage,
boolean class fields and the original delete/presence boundary remain stopped.

## Exact tsc boundary before and after

All 82 input hashes and byte lengths match the stock catalogue pinned at ec0b16c0;
[input-hashes.json](step-17-unions/evidence/input-hashes.json) preserves that check.
A detached checkout of delivery base 4885cec5 uses the shared pinned cohere
submodule through a symlink. No worker branch was merged into the delivery.

The first selector used the declaration position 102:5, which is not the failure
position. Its log retains the actual diagnostic at 102:52. Replaying that exact
observed diagnostic on the base exits 0 and reproduces the NotYet reason:

```sh
source /workspace/adamic-tools/env.sh
go run ./stage3/census/latent/replay \
  -project /tmp/hidden-adapted/src/compiler \
  -where /tmp/hidden-adapted/src/compiler/sourcemap.ts:102:52 -kind NotYet \
  -reason 'a value of type string | null' > replay.log 2>&1
```

[Base exact replay](step-17-unions/evidence/tsc-before-exact.log) selects
setSourceContent at 102:5 and reports that same reason at 102:52.
[Fixing exact replay](step-17-unions/evidence/tsc-after-exact.log) selects the same
function and instead reports NotYet `an array of never` at 105:51, for the empty
array initializer. Its exit 1 means the requested old signature no longer matches;
it is not a successful whole-function/native build. The selected concrete
nullable-string root has advanced. The nullish-string reduction also now lowers,
but its other tsc callers have not all been replayed.

The old 6,421 and 1,382 credits remain historical outermost estimates. No whole
hidden census was rerun and no hidden-byte reduction is claimed for this unit.
Generic T/U/V/K, branded strings, checked-view contracts, closure capture/ABI,
other nullable arms and overload specialization remain separate blockers.

## Mutants and independent catchers

| Mutant actually run | Catcher and observed result |
| --- | --- |
| Add one to a copied census byte credit | Exact source-ledger comparison rejects the altered credit |
| Box null as NULL | The field/array fixture completes natively with exit 0, no stderr, but prints undefined in null positions; source Node prints null |
| Treat a tagged union as an untagged nullable pointer in lowering | Nullish-content probe becomes NotYet at typeof; its required successful lowering fails |
| Classify a union's missing pointer as null in native typeof | Native and release stdout disagree with source Node on the undefined arm |
| Refuse nullable-string storage again | TestNullableStringAdmission fails at the previously admitted parameter |
| Drop the narrowed field's tag check in IR | The checked baseline exits 70; mutant JavaScript completes and prints undefined, exposing the unchecked field load |
| Check an explicit null comparison before observing its tag | The stale-observation fixture panics where source Node completes |

Every source mutant was restored before final checks. The first null-box mutant
was tried against only direct-call arguments and survived: that boundary has its
own null conversion and bypasses Box. It was rerun independently against field
and array boxing and caught by a completed wrong-output run, not a clang error or
sanitizer crash. The dropped-check IR mutant is executed within its focused test
and the overall test passes only when the mutant disagrees.

## Counts and validation

`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m
-args -update-counts` passed. Five registered rows were added:

| Fixture | Allocations | Frees | Retains | Releases | Max live | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| hir-optional-boolean.a | 1 | 1 | 0 | 1 | 1 | 0 |
| tsc-source-content.a | 0 | 0 | 2 | 2 | 0 | 0 |
| tsc-nullish-content.a | 0 | 0 | 6 | 6 | 0 | 0 |
| scout_nullable_strings.a | 34 | 34 | 130 | 156 | 5 | 0 |
| scout_nullable_string_stale_field.a | 1 | 0 | 4 | 3 | 1 | 0 |

The HIR row allocates its argument object. The two reduced tsc probes use immortal
strings/null/undefined, so their ownership transfers allocate nothing. The full
fixture's dynamic strings, arrays and objects all free; the final checked fixture
is counted at its intentional panic, so its live object is not a finished-program
leak. Null and undefined need no allocation.

Two existing rows also change structurally: logical_and_reference_maybe.a moves
to its registration position with all six numbers unchanged; taste/17_binder_flow.a
is removed because the delivery base's taste_stage3_test.go already registers it
with lowers=false. Neither is a new compiler acceptance/refusal change. Every
other old row keeps its numbers. The standalone stopped probes are not registered
as positive count fixtures.

Final focused commands, with output in evidence log files:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m -v \
  -run '^TestNullableStringFieldCheckMutant$|^TestScoutUnionSourceOutcomes$|^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^scout_nullable_'
go test ./internal/lower -count=1 -v \
  -run '^TestNullableStringAdmission$|^TestTasteRepresentationLimitsStayExplicit$|^TestUnknownReflectionRefusals$|^TestProvenRelationsRefuse$'
go vet ./internal/lower ./internal/native ./internal/oracle
gofmt -l internal/lower/expression.go internal/lower/object.go \
  internal/lower/nullable_strings.go internal/lower/nullable_strings_test.go \
  internal/native/union.go internal/oracle/scout_unions_test.go
```

The backend selector covers two registered runtime fixtures plus all six measured
probes; the source reductions that still stop pin their exact reason. No whole
package test or full gate was run. Formatting and vet logs are empty. Setup passed
with Go/Node at 0.023s, submodules 0.062s, clang 0.152s, build cache 41.424s and
completion 41.452s; nproc=5. Detached-checkout compilation exhausted workspace
space; removing only regenerable Go cache entries restored it, and no source or
evidence was removed. The delivery carries own commits only on the area tip.

Scalar-only `number | undefined` and `boolean | undefined` do not have a counted
arm. Their existing Maybe pair/packed scalar slots remain valid specializations
of present-versus-undefined; this unit does not force them into allocated boxes.
The class-field gate still rejects optional booleans on this base even though
interface reads have a tagged byte. Completing that class path is remaining work.
