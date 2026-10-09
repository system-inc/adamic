# Discriminant writes in TypeScript 6.0.3

This is a stock-checker census with an explicitly limited construction analysis.
It does **not** establish a complete whole-program escape classification.
Source commit: `050880ce59e30b356b686bd3144efe24f875ebc8`.
The inherited compiler project has **zero stock diagnostics**, and SHA256
checks match all **77** original compiler files to the census.
Generated diagnostics and code outside `src/compiler` are excluded.

## Counts and comparison units

**437 resolved checker-candidate write events**, including **79 `kind` writes**:
76 object-literal initializers and 3 constructor assignments, all inside
construction. There are **zero resolved outside-construction `kind` assignments**.
This is not a claim that dynamic-key copying can never write `kind`.

The remaining 358 events comprise **274 other finite-literal discriminant
writes** and **84 broader/partial checker candidates**. The latter are listed
separately rather than silently claiming every checker candidate is a fixed tag.
For example, `CommandLineOption.type` can be a literal **or a Map**, and
control-flow `antecedent` fields can distinguish presence from absence.
`flags` qualifies through `TypeSystemEntity` and `FlowType`; overlapping enum
values mean it is not a unique nominal identity. See the JSON witnesses.

The local construction partition is **388 inside**, **47
outside local construction**, and **2 not established**.
The distinction between local construction and global escape is essential below.
Assignments/update syntax total 93; creation syntax totals 344
(341 explicit object properties and 3 spread/property events).

| Property | Creation initializer/spread | Other write syntax | Inside | Outside local construction | Not established | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| kind | 76 | 3 | 79 | 0 | 0 | 79 |
| antecedent | 1 | 8 | 2 | 7 | 0 | 9 |
| commonJsModuleIndicator | 0 | 1 | 0 | 1 | 0 | 1 |
| computed | 4 | 0 | 4 | 0 | 0 | 4 |
| directoryExists | 1 | 0 | 1 | 0 | 0 | 1 |
| done | 2 | 0 | 2 | 0 | 0 | 2 |
| error | 2 | 0 | 2 | 0 | 0 | 2 |
| errorMessage | 1 | 0 | 1 | 0 | 0 | 1 |
| externalModuleIndicator | 0 | 5 | 1 | 4 | 0 | 5 |
| extraInitializersName | 1 | 0 | 1 | 0 | 0 | 1 |
| file | 8 | 0 | 8 | 0 | 0 | 8 |
| flags | 1 | 61 | 33 | 27 | 2 | 62 |
| innerExpression | 3 | 0 | 3 | 0 | 0 | 3 |
| modifiers | 0 | 7 | 5 | 2 | 0 | 7 |
| module | 2 | 0 | 2 | 0 | 0 | 2 |
| modulePath | 2 | 0 | 2 | 0 | 0 | 2 |
| operator | 0 | 2 | 2 | 0 | 0 | 2 |
| parameterIndex | 1 | 0 | 1 | 0 | 0 | 1 |
| parameterName | 1 | 0 | 1 | 0 | 0 | 1 |
| reportFallback | 1 | 0 | 1 | 0 | 0 | 1 |
| scoped | 34 | 0 | 34 | 0 | 0 | 34 |
| signature | 5 | 4 | 5 | 4 | 0 | 9 |
| sym | 1 | 0 | 1 | 0 | 0 | 1 |
| type | 196 | 1 | 196 | 1 | 0 | 197 |
| useCaseSensitiveFileNames | 1 | 0 | 1 | 0 | 0 | 1 |
| version | 0 | 1 | 0 | 1 | 0 | 1 |

## Exact discovery and counting rule

`discriminant-writes.cjs` uses npm stock TypeScript 6.0.3, the compiler's inherited
tsconfig, and checker types at AST expressions, declarations, type nodes and
call-signature return types, recursively including union constituents and
constraints. No text search discovers properties, unions or writes.

For each resolved union property, stock `getCheckFlags(property)` must contain
`CheckFlags.Discriminant` (HasNonUniformType | HasLiteralType). The ledger saves
the union and every constituent's property type and declaration. Direct top-level
type-parameter property types are excluded from candidate indexing. The finite
subset additionally requires only literal/enum-literal/null/undefined/never
constituents and at least two distinct member property domains. The broader
subset is a **candidate list**, not a reimplementation of stock's private
recursive generic-type predicate or a proof of a globally immutable tag.
All properties named `kind` are counted independently of these predicates,
including pragma flags, comments, module-specifier results and generator blocks.

Each write resolves its receiver property symbol through the checker and root
symbols to declarations. Candidate witnesses must share a declaration; for
explicit accesses the receiver must be stock-assignable to a witness member.
For contextual object literals, declarations are filtered to members admitting
at least one constituent of the initializer's checked value type. This prevents
a Map-valued command-line option initializer from being attributed to the
boolean option's declaration merely because contextual union symbols combine
all declarations. The literal's own declaration is also retained in JSON.
When more than one declaration remains possible, all are reported, not one
chosen by spelling. A property is counted once per syntactic write, even when
many unions witness it. A spread counts once per selected copied property;
it does not count as a mutation of the spread source.

AST forms cover assignment/compound assignment, ++/--, delete, finite literal
bracket keys, destructuring assignment targets, object property/shorthand/method
initializers, spreads, class field initializers and constructor parameter
properties, plus checker-resolved Object.assign and Object.defineProperty.
For compounds the recorded type is the checked result type, rather than the
right operand alone. Counts are source sites, not execution frequencies.

## Exact local construction rule and its limit

An initializer of a new object literal, including spread into that new object,
is inside construction. A function-style allocator constructor's `this` is a
fresh receiver; its assignments count inside only before `this` is returned,
stored externally, or passed to a retaining call. The allocator's Node, Token,
Identifier, Symbol, Type and Signature constructors were inspected. The three
`kind` assignments follow only writes to `this.pos` and `this.end`.

For an ordinary function, an assignment counts inside only at a reviewed span
where all paths reaching it provide a fresh local allocation/clone, and no
preceding operation publishes that receiver. Internal allocating helpers may
return the fresh object to the enclosing construction function; that return
alone is not publication. Property writes on that local and helpers such as
setTextRange/setParent that only populate the fresh receiver are not publication.
Return to the caller, storing into a cache, map, object field or externally
reachable collection, or passing it to an unproved retaining callback ends
construction. The reviewed span list is in `discriminant-report.py`.

Incoming receiver parameters and cache/global/property lookups are outside
**this function's local construction**, even if a caller sometimes invokes the
function while constructing an object. Mixed fresh/existing paths are also
outside the guaranteed construction category. This does not assert that every
such execution is after a global escape. Parser `finishNode` and `withJSDoc`,
for example, are construction helpers whose callers need an interprocedural
freshness proof. I did not do that whole-program proof. The stock checker has
no escape-analysis API; function names and a `const` local do not supply one.
Two modifier-writing parser spans are explicitly not established rather than
assuming that a parsed modifier cannot have been reused or published.

## Definite outside retag

`src/compiler/tsbuildPublic.ts:1921` obtains `status` from
`state.projectStatus.get(nextProjectPath)`. After checking `status.type ===
UpToDateStatusType.UpToDate`, it writes
`UpToDateStatusType.UpToDateWithUpstreamTypes` if declaration output is unchanged.
The object already resides in the project-status cache. This is a build-status
retag, not a syntax-node retag or a fresh construction. It lets referenced
projects update output timestamps instead of rebuilding unchanged declarations.
Under the fixed-discriminant ruling this write needs an adaptation, such as
replacing the cached record with a newly constructed status. No adaptation is
made here.

Other outside entries update binding flags, aggregate node flags, merge symbols,
mark optional chains, update incremental signatures, recompute module indicators,
or temporarily replace and restore control-flow antecedents. Those purposes
are given with each row. No resolved write changes an existing syntax node's
`kind` to a different SyntaxKind.

## Locations: kind first

| Write | Property / form | Written value type | Declaration(s) | Classification / purpose |
| --- | --- | --- | --- | --- |
| src/compiler/checker.ts:7877:62 | kind / object initializer | SyntaxKind.MultiLineCommentTrivia | src/compiler/types.ts:3874:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:9919:48 | kind / object initializer | SyntaxKind.MultiLineCommentTrivia | src/compiler/types.ts:3874:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:16122:18 | kind / object initializer | TypePredicateKind | src/compiler/types.ts:5696:5; src/compiler/types.ts:5703:5; src/compiler/types.ts:5710:5; src/compiler/types.ts:5717:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:20610:52 | kind / object initializer | TypeMapKind.Simple | src/compiler/types.ts:7087:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:20614:52 | kind / object initializer | TypeMapKind.Array | src/compiler/types.ts:7088:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:20618:52 | kind / object initializer | TypeMapKind.Function | src/compiler/types.ts:7090:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:20622:52 | kind / object initializer | TypeMapKind.Deferred | src/compiler/types.ts:7089:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:20626:52 | kind / object initializer | TypeMapKind.Composite \| TypeMapKind.Merged | src/compiler/types.ts:7091:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:41754:65 | kind / object initializer | SyntaxKind | src/compiler/checker.ts:41754:65 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:45283:74 | kind / object initializer | SyntaxKind.VariableDeclaration | src/compiler/checker.ts:45283:74 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:49492:67 | kind / object initializer | SyntaxKind | src/compiler/checker.ts:49492:67 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitNode.ts:205:110 | kind / object initializer | SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia | src/compiler/types.ts:3874:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitNode.ts:218:112 | kind / object initializer | SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia | src/compiler/types.ts:3874:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:332:27 | kind / object initializer | "paths" \| "node_modules" \| "redirect" \| "relative" \| "ambient" \| undefined | src/compiler/moduleSpecifiers.ts:332:27 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:400:13 | kind / object initializer | "ambient" | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:414:30 | kind / object initializer | "paths" \| "node_modules" \| "redirect" \| "relative" \| "ambient" \| undefined | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:415:37 | kind / object initializer | undefined | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:484:18 | kind / object initializer | undefined | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:507:26 | kind / object initializer | "node_modules" | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:552:40 | kind / object initializer | "paths" | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:553:45 | kind / object initializer | "redirect" | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:554:43 | kind / object initializer | "node_modules" | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/moduleSpecifiers.ts:555:11 | kind / object initializer | "relative" | src/compiler/moduleSpecifiers.ts:380:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1760:73 | kind / object initializer | FileIncludeKind.SourceFromProjectReference | src/compiler/types.ts:4627:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1766:90 | kind / object initializer | FileIncludeKind.OutputFromProjectReference | src/compiler/types.ts:4627:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1772:191 | kind / object initializer | FileIncludeKind.OutputFromProjectReference | src/compiler/types.ts:4627:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1782:93 | kind / object initializer | FileIncludeKind.RootFile | src/compiler/types.ts:4610:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1802:25 | kind / object initializer | FileIncludeKind.AutomaticTypeDirectiveFile | src/compiler/types.ts:4647:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1817:82 | kind / object initializer | FileIncludeKind.LibFile | src/compiler/types.ts:4616:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:1821:91 | kind / object initializer | FileIncludeKind.LibFile | src/compiler/types.ts:4616:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:2038:13 | kind / object initializer | FilePreprocessingDiagnosticsKind.ResolutionDiagnostics | src/compiler/types.ts:4684:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:3766:19 | kind / object initializer | FileIncludeKind.ReferenceFile | src/compiler/types.ts:4640:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:3786:93 | kind / object initializer | FileIncludeKind.TypeReferenceDirective | src/compiler/types.ts:4640:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:3896:87 | kind / object initializer | FileIncludeKind.LibReferenceDirective | src/compiler/types.ts:4640:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:3900:21 | kind / object initializer | FilePreprocessingDiagnosticsKind.FilePreprocessingLibReferenceDiagnostic | src/compiler/types.ts:4669:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:3901:31 | kind / object initializer | FileIncludeKind.LibReferenceDirective | src/compiler/types.ts:4640:5; src/compiler/types.ts:4670:32 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:3966:27 | kind / object initializer | FileIncludeKind.Import | src/compiler/types.ts:4640:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/program.ts:4579:13 | kind / object initializer | FilePreprocessingDiagnosticsKind.FilePreprocessingFileExplainingDiagnostic | src/compiler/types.ts:4675:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/scanner.ts:953:21 | kind / object initializer | CommentKind | src/compiler/types.ts:3874:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformer.ts:338:146 | kind / object initializer | SyntaxKind.Unknown \| SyntaxKind.EndOfFileToken \| SyntaxKind.SingleLineCommentTrivia \| SyntaxKind.MultiLineCommentTrivia \| SyntaxKind.NewLineTrivia \| SyntaxKind.WhitespaceTrivia \| SyntaxKind.ShebangTrivia \| SyntaxKind.ConflictMarkerTrivia \| SyntaxKind.NonTextFileMarkerTrivia \| SyntaxKind.NumericLiteral \| SyntaxKind.BigIntLiteral \| SyntaxKind.StringLiteral \| SyntaxKind.JsxText \| SyntaxKind.JsxTextAllWhiteSpaces \| SyntaxKind.RegularExpressionLiteral \| SyntaxKind.NoSubstitutionTemplateLiteral \| SyntaxKind.TemplateHead \| SyntaxKind.TemplateMiddle \| SyntaxKind.TemplateTail \| SyntaxKind.OpenBraceToken \| SyntaxKind.CloseBraceToken \| SyntaxKind.OpenParenToken \| SyntaxKind.CloseParenToken \| SyntaxKind.OpenBracketToken \| SyntaxKind.CloseBracketToken \| SyntaxKind.DotToken \| SyntaxKind.DotDotDotToken \| SyntaxKind.SemicolonToken \| SyntaxKind.CommaToken \| SyntaxKind.QuestionDotToken \| SyntaxKind.LessThanToken \| SyntaxKind.LessThanSlashToken \| SyntaxKind.GreaterThanToken \| SyntaxKind.LessThanEqualsToken \| SyntaxKind.GreaterThanEqualsToken \| SyntaxKind.EqualsEqualsToken \| SyntaxKind.ExclamationEqualsToken \| SyntaxKind.EqualsEqualsEqualsToken \| SyntaxKind.ExclamationEqualsEqualsToken \| SyntaxKind.EqualsGreaterThanToken \| SyntaxKind.PlusToken \| SyntaxKind.MinusToken \| SyntaxKind.AsteriskToken \| SyntaxKind.AsteriskAsteriskToken \| SyntaxKind.SlashToken \| SyntaxKind.PercentToken \| SyntaxKind.PlusPlusToken \| SyntaxKind.MinusMinusToken \| SyntaxKind.LessThanLessThanToken \| SyntaxKind.GreaterThanGreaterThanToken \| SyntaxKind.GreaterThanGreaterThanGreaterThanToken \| SyntaxKind.AmpersandToken \| SyntaxKind.BarToken \| SyntaxKind.CaretToken \| SyntaxKind.ExclamationToken \| SyntaxKind.TildeToken \| SyntaxKind.AmpersandAmpersandToken \| SyntaxKind.BarBarToken \| SyntaxKind.QuestionToken \| SyntaxKind.ColonToken \| SyntaxKind.AtToken \| SyntaxKind.QuestionQuestionToken \| SyntaxKind.BacktickToken \| SyntaxKind.HashToken \| SyntaxKind.EqualsToken \| SyntaxKind.PlusEqualsToken \| SyntaxKind.MinusEqualsToken \| SyntaxKind.AsteriskEqualsToken \| SyntaxKind.AsteriskAsteriskEqualsToken \| SyntaxKind.SlashEqualsToken \| SyntaxKind.PercentEqualsToken \| SyntaxKind.LessThanLessThanEqualsToken \| SyntaxKind.GreaterThanGreaterThanEqualsToken \| SyntaxKind.GreaterThanGreaterThanGreaterThanEqualsToken \| SyntaxKind.AmpersandEqualsToken \| SyntaxKind.BarEqualsToken \| SyntaxKind.BarBarEqualsToken \| SyntaxKind.AmpersandAmpersandEqualsToken \| SyntaxKind.QuestionQuestionEqualsToken \| SyntaxKind.CaretEqualsToken \| SyntaxKind.Identifier \| SyntaxKind.PrivateIdentifier \| SyntaxKind.JSDocCommentTextToken \| SyntaxKind.BreakKeyword \| SyntaxKind.CaseKeyword \| SyntaxKind.CatchKeyword \| SyntaxKind.ClassKeyword \| SyntaxKind.ConstKeyword \| SyntaxKind.ContinueKeyword \| SyntaxKind.DebuggerKeyword \| SyntaxKind.DefaultKeyword \| SyntaxKind.DeleteKeyword \| SyntaxKind.DoKeyword \| SyntaxKind.ElseKeyword \| SyntaxKind.EnumKeyword \| SyntaxKind.ExportKeyword \| SyntaxKind.ExtendsKeyword \| SyntaxKind.FalseKeyword \| SyntaxKind.FinallyKeyword \| SyntaxKind.ForKeyword \| SyntaxKind.FunctionKeyword \| SyntaxKind.IfKeyword \| SyntaxKind.ImportKeyword \| SyntaxKind.InKeyword \| SyntaxKind.InstanceOfKeyword \| SyntaxKind.NewKeyword \| SyntaxKind.NullKeyword \| SyntaxKind.ReturnKeyword \| SyntaxKind.SuperKeyword \| SyntaxKind.SwitchKeyword \| SyntaxKind.ThisKeyword \| SyntaxKind.ThrowKeyword \| SyntaxKind.TrueKeyword \| SyntaxKind.TryKeyword \| SyntaxKind.TypeOfKeyword \| SyntaxKind.VarKeyword \| SyntaxKind.VoidKeyword \| SyntaxKind.WhileKeyword \| SyntaxKind.WithKeyword \| SyntaxKind.ImplementsKeyword \| SyntaxKind.InterfaceKeyword \| SyntaxKind.LetKeyword \| SyntaxKind.PackageKeyword \| SyntaxKind.PrivateKeyword \| SyntaxKind.ProtectedKeyword \| SyntaxKind.PublicKeyword \| SyntaxKind.StaticKeyword \| SyntaxKind.YieldKeyword \| SyntaxKind.AbstractKeyword \| SyntaxKind.AccessorKeyword \| SyntaxKind.AsKeyword \| SyntaxKind.AssertsKeyword \| SyntaxKind.AssertKeyword \| SyntaxKind.AnyKeyword \| SyntaxKind.AsyncKeyword \| SyntaxKind.AwaitKeyword \| SyntaxKind.BooleanKeyword \| SyntaxKind.ConstructorKeyword \| SyntaxKind.DeclareKeyword \| SyntaxKind.GetKeyword \| SyntaxKind.InferKeyword \| SyntaxKind.IntrinsicKeyword \| SyntaxKind.IsKeyword \| SyntaxKind.KeyOfKeyword \| SyntaxKind.ModuleKeyword \| SyntaxKind.NamespaceKeyword \| SyntaxKind.NeverKeyword \| SyntaxKind.OutKeyword \| SyntaxKind.ReadonlyKeyword \| SyntaxKind.RequireKeyword \| SyntaxKind.NumberKeyword \| SyntaxKind.ObjectKeyword \| SyntaxKind.SatisfiesKeyword \| SyntaxKind.SetKeyword \| SyntaxKind.StringKeyword \| SyntaxKind.SymbolKeyword \| SyntaxKind.TypeKeyword \| SyntaxKind.UndefinedKeyword \| SyntaxKind.UniqueKeyword \| SyntaxKind.UnknownKeyword \| SyntaxKind.UsingKeyword \| SyntaxKind.FromKeyword \| SyntaxKind.GlobalKeyword \| SyntaxKind.BigIntKeyword \| SyntaxKind.OverrideKeyword \| SyntaxKind.OfKeyword \| SyntaxKind.DeferKeyword \| SyntaxKind.QualifiedName \| SyntaxKind.ComputedPropertyName \| SyntaxKind.TypeParameter \| SyntaxKind.Parameter \| SyntaxKind.Decorator \| SyntaxKind.PropertySignature \| SyntaxKind.PropertyDeclaration \| SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.ClassStaticBlockDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| SyntaxKind.IndexSignature \| SyntaxKind.TypePredicate \| SyntaxKind.TypeReference \| SyntaxKind.FunctionType \| SyntaxKind.ConstructorType \| SyntaxKind.TypeQuery \| SyntaxKind.TypeLiteral \| SyntaxKind.ArrayType \| SyntaxKind.TupleType \| SyntaxKind.OptionalType \| SyntaxKind.RestType \| SyntaxKind.UnionType \| SyntaxKind.IntersectionType \| SyntaxKind.ConditionalType \| SyntaxKind.InferType \| SyntaxKind.ParenthesizedType \| SyntaxKind.ThisType \| SyntaxKind.TypeOperator \| SyntaxKind.IndexedAccessType \| SyntaxKind.MappedType \| SyntaxKind.LiteralType \| SyntaxKind.NamedTupleMember \| SyntaxKind.TemplateLiteralType \| SyntaxKind.TemplateLiteralTypeSpan \| SyntaxKind.ImportType \| SyntaxKind.ObjectBindingPattern \| SyntaxKind.ArrayBindingPattern \| SyntaxKind.BindingElement \| SyntaxKind.ArrayLiteralExpression \| SyntaxKind.ObjectLiteralExpression \| SyntaxKind.PropertyAccessExpression \| SyntaxKind.ElementAccessExpression \| SyntaxKind.CallExpression \| SyntaxKind.NewExpression \| SyntaxKind.TaggedTemplateExpression \| SyntaxKind.TypeAssertionExpression \| SyntaxKind.ParenthesizedExpression \| SyntaxKind.FunctionExpression \| SyntaxKind.ArrowFunction \| SyntaxKind.DeleteExpression \| SyntaxKind.TypeOfExpression \| SyntaxKind.VoidExpression \| SyntaxKind.AwaitExpression \| SyntaxKind.PrefixUnaryExpression \| SyntaxKind.PostfixUnaryExpression \| SyntaxKind.BinaryExpression \| SyntaxKind.ConditionalExpression \| SyntaxKind.TemplateExpression \| SyntaxKind.YieldExpression \| SyntaxKind.SpreadElement \| SyntaxKind.ClassExpression \| SyntaxKind.OmittedExpression \| SyntaxKind.ExpressionWithTypeArguments \| SyntaxKind.AsExpression \| SyntaxKind.NonNullExpression \| SyntaxKind.MetaProperty \| SyntaxKind.SyntheticExpression \| SyntaxKind.SatisfiesExpression \| SyntaxKind.TemplateSpan \| SyntaxKind.SemicolonClassElement \| SyntaxKind.Block \| SyntaxKind.EmptyStatement \| SyntaxKind.VariableStatement \| SyntaxKind.ExpressionStatement \| SyntaxKind.IfStatement \| SyntaxKind.DoStatement \| SyntaxKind.WhileStatement \| SyntaxKind.ForStatement \| SyntaxKind.ForInStatement \| SyntaxKind.ForOfStatement \| SyntaxKind.ContinueStatement \| SyntaxKind.BreakStatement \| SyntaxKind.ReturnStatement \| SyntaxKind.WithStatement \| SyntaxKind.SwitchStatement \| SyntaxKind.LabeledStatement \| SyntaxKind.ThrowStatement \| SyntaxKind.TryStatement \| SyntaxKind.DebuggerStatement \| SyntaxKind.VariableDeclaration \| SyntaxKind.VariableDeclarationList \| SyntaxKind.FunctionDeclaration \| SyntaxKind.ClassDeclaration \| SyntaxKind.InterfaceDeclaration \| SyntaxKind.TypeAliasDeclaration \| SyntaxKind.EnumDeclaration \| SyntaxKind.ModuleDeclaration \| SyntaxKind.ModuleBlock \| SyntaxKind.CaseBlock \| SyntaxKind.NamespaceExportDeclaration \| SyntaxKind.ImportEqualsDeclaration \| SyntaxKind.ImportDeclaration \| SyntaxKind.ImportClause \| SyntaxKind.NamespaceImport \| SyntaxKind.NamedImports \| SyntaxKind.ImportSpecifier \| SyntaxKind.ExportAssignment \| SyntaxKind.ExportDeclaration \| SyntaxKind.NamedExports \| SyntaxKind.NamespaceExport \| SyntaxKind.ExportSpecifier \| SyntaxKind.MissingDeclaration \| SyntaxKind.ExternalModuleReference \| SyntaxKind.JsxElement \| SyntaxKind.JsxSelfClosingElement \| SyntaxKind.JsxOpeningElement \| SyntaxKind.JsxClosingElement \| SyntaxKind.JsxFragment \| SyntaxKind.JsxOpeningFragment \| SyntaxKind.JsxClosingFragment \| SyntaxKind.JsxAttribute \| SyntaxKind.JsxAttributes \| SyntaxKind.JsxSpreadAttribute \| SyntaxKind.JsxExpression \| SyntaxKind.JsxNamespacedName \| SyntaxKind.CaseClause \| SyntaxKind.DefaultClause \| SyntaxKind.HeritageClause \| SyntaxKind.CatchClause \| SyntaxKind.ImportAttributes \| SyntaxKind.ImportAttribute \| SyntaxKind.ImportTypeAssertionContainer \| SyntaxKind.PropertyAssignment \| SyntaxKind.ShorthandPropertyAssignment \| SyntaxKind.SpreadAssignment \| SyntaxKind.EnumMember \| SyntaxKind.Bundle \| SyntaxKind.JSDocTypeExpression \| SyntaxKind.JSDocNameReference \| SyntaxKind.JSDocMemberName \| SyntaxKind.JSDocAllType \| SyntaxKind.JSDocUnknownType \| SyntaxKind.JSDocNullableType \| SyntaxKind.JSDocNonNullableType \| SyntaxKind.JSDocOptionalType \| SyntaxKind.JSDocFunctionType \| SyntaxKind.JSDocVariadicType \| SyntaxKind.JSDocNamepathType \| SyntaxKind.JSDoc \| SyntaxKind.JSDocText \| SyntaxKind.JSDocTypeLiteral \| SyntaxKind.JSDocSignature \| SyntaxKind.JSDocLink \| SyntaxKind.JSDocLinkCode \| SyntaxKind.JSDocLinkPlain \| SyntaxKind.JSDocTag \| SyntaxKind.JSDocAugmentsTag \| SyntaxKind.JSDocImplementsTag \| SyntaxKind.JSDocAuthorTag \| SyntaxKind.JSDocDeprecatedTag \| SyntaxKind.JSDocClassTag \| SyntaxKind.JSDocPublicTag \| SyntaxKind.JSDocPrivateTag \| SyntaxKind.JSDocProtectedTag \| SyntaxKind.JSDocReadonlyTag \| SyntaxKind.JSDocOverrideTag \| SyntaxKind.JSDocCallbackTag \| SyntaxKind.JSDocOverloadTag \| SyntaxKind.JSDocEnumTag \| SyntaxKind.JSDocParameterTag \| SyntaxKind.JSDocReturnTag \| SyntaxKind.JSDocThisTag \| SyntaxKind.JSDocTypeTag \| SyntaxKind.JSDocTemplateTag \| SyntaxKind.JSDocTypedefTag \| SyntaxKind.JSDocSeeTag \| SyntaxKind.JSDocPropertyTag \| SyntaxKind.JSDocThrowsTag \| SyntaxKind.JSDocSatisfiesTag \| SyntaxKind.JSDocImportTag \| SyntaxKind.SyntaxList \| SyntaxKind.NotEmittedStatement \| SyntaxKind.NotEmittedTypeElement \| SyntaxKind.PartiallyEmittedExpression \| SyntaxKind.CommaListExpression \| SyntaxKind.SyntheticReferenceExpression \| SyntaxKind.Count | src/compiler/transformer.ts:338:146 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2110:73 | kind / object initializer | "untransformed" | src/compiler/transformers/classFields.ts:310:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2133:77 | kind / object initializer | "untransformed" | src/compiler/transformers/classFields.ts:310:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2795:17 | kind / object initializer | PrivateIdentifierKind.Field | src/compiler/transformers/classFields.ts:284:5; src/compiler/transformers/classFields.ts:289:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2806:17 | kind / object initializer | PrivateIdentifierKind.Field | src/compiler/transformers/classFields.ts:284:5; src/compiler/transformers/classFields.ts:289:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2838:13 | kind / object initializer | PrivateIdentifierKind.Method | src/compiler/transformers/classFields.ts:276:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2865:17 | kind / object initializer | PrivateIdentifierKind.Accessor | src/compiler/transformers/classFields.ts:264:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2897:17 | kind / object initializer | PrivateIdentifierKind.Accessor | src/compiler/transformers/classFields.ts:264:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/classFields.ts:2923:13 | kind / object initializer | PrivateIdentifierKind.Accessor | src/compiler/transformers/classFields.ts:264:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/es2015.ts:484:14 | kind / object initializer | SpreadSegmentKind | src/compiler/transformers/es2015.ts:479:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:351:17 | kind / object initializer | "class" | src/compiler/transformers/esDecorators.ts:351:17 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:365:17 | kind / object initializer | "class-element" | src/compiler/transformers/esDecorators.ts:365:17 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:382:17 | kind / object initializer | "name" | src/compiler/transformers/esDecorators.ts:382:17 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:398:21 | kind / object initializer | "other" | src/compiler/transformers/esDecorators.ts:398:21 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:888:19 | kind / object initializer | "class" | src/compiler/factory/emitHelpers.ts:57:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:1315:17 | kind / object initializer | "accessor" \| "method" \| "getter" \| "setter" \| "field" | src/compiler/factory/emitHelpers.ts:76:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2189:13 | kind / object initializer | CodeBlockKind.With | src/compiler/transformers/generators.ts:300:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2213:13 | kind / object initializer | CodeBlockKind.Exception | src/compiler/transformers/generators.ts:266:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2310:13 | kind / object initializer | CodeBlockKind.Loop | src/compiler/transformers/generators.ts:292:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2328:13 | kind / object initializer | CodeBlockKind.Loop | src/compiler/transformers/generators.ts:292:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2355:13 | kind / object initializer | CodeBlockKind.Switch | src/compiler/transformers/generators.ts:285:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2369:13 | kind / object initializer | CodeBlockKind.Switch | src/compiler/transformers/generators.ts:285:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2390:13 | kind / object initializer | CodeBlockKind.Labeled | src/compiler/transformers/generators.ts:277:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/generators.ts:2400:13 | kind / object initializer | CodeBlockKind.Labeled | src/compiler/transformers/generators.ts:277:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:905:9 | kind / object initializer | InvalidatedProjectKind.UpdateOutputFileStamps | src/compiler/tsbuildPublic.ts:850:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:946:9 | kind / object initializer | InvalidatedProjectKind.Build | src/compiler/tsbuildPublic.ts:855:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1268:21 | kind / object initializer | InvalidatedProjectKind.UpdateOutputFileStamps | src/compiler/tsbuildPublic.ts:1268:21 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1304:13 | kind / object initializer | InvalidatedProjectKind.Build | src/compiler/tsbuildPublic.ts:1304:13 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10250:9 | kind / object initializer | PragmaKindFlags.TripleSlashXML | src/compiler/types.ts:10360:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10254:9 | kind / object initializer | PragmaKindFlags.TripleSlashXML | src/compiler/types.ts:10369:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10258:9 | kind / object initializer | PragmaKindFlags.TripleSlashXML | src/compiler/types.ts:10375:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10261:9 | kind / object initializer | PragmaKindFlags.SingleLine | src/compiler/types.ts:10378:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10264:9 | kind / object initializer | PragmaKindFlags.SingleLine | src/compiler/types.ts:10381:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10268:9 | kind / object initializer | PragmaKindFlags.MultiLine | src/compiler/types.ts:10387:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10272:9 | kind / object initializer | PragmaKindFlags.MultiLine | src/compiler/types.ts:10393:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10276:9 | kind / object initializer | PragmaKindFlags.MultiLine | src/compiler/types.ts:10399:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/types.ts:10280:9 | kind / object initializer | PragmaKindFlags.MultiLine | src/compiler/types.ts:10405:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/utilities.ts:8509:5 | kind / = | SyntaxKind | src/compiler/types.ts:943:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8523:5 | kind / = | SyntaxKind | src/compiler/types.ts:943:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8535:5 | kind / = | SyntaxKind | src/compiler/types.ts:943:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |

## Other finite-literal discriminants

| Write | Property / form | Written value type | Declaration(s) | Classification / purpose |
| --- | --- | --- | --- | --- |
| src/compiler/binder.ts:636:9 | flags / \|= | number | src/compiler/types.ts:6038:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: addDeclarationToSymbol. |
| src/compiler/binder.ts:1030:13 | flags / &= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindContainer. |
| src/compiler/binder.ts:1032:17 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindContainer. |
| src/compiler/binder.ts:1033:40 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindContainer. |
| src/compiler/binder.ts:1037:17 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindContainer. |
| src/compiler/binder.ts:1040:17 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindContainer. |
| src/compiler/binder.ts:1066:13 | flags / = | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindContainer. |
| src/compiler/binder.ts:1104:13 | flags / &= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindChildren. |
| src/compiler/binder.ts:1112:17 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindChildren. |
| src/compiler/binder.ts:1802:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: bindLabeledStatement. |
| src/compiler/binder.ts:2345:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: setExportContextFlag. |
| src/compiler/binder.ts:2348:13 | flags / &= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: setExportContextFlag. |
| src/compiler/checker.ts:2729:13 | flags / \|= | number | src/compiler/types.ts:6038:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: mergeSymbol. |
| src/compiler/checker.ts:5005:9 | flags / = | number | src/compiler/types.ts:6038:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: getCommonJsExportEquals. |
| src/compiler/checker.ts:13571:17 | flags / \|= | number | src/compiler/types.ts:6440:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: getDeclaredTypeOfEnum. |
| src/compiler/checker.ts:13840:9 | flags / \|= | number | src/compiler/types.ts:6038:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: addDeclarationToLateBoundSymbol. |
| src/compiler/checker.ts:14183:9 | flags / \|= | number | src/compiler/types.ts:7020:22 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:14264:17 | flags / = | number | src/compiler/types.ts:7020:22 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:18453:17 | flags / \|= | number | src/compiler/types.ts:6440:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:25934:9 | flags / = | TypeFlags | src/compiler/types.ts:6440:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:25995:9 | flags / \|= | number | src/compiler/types.ts:6038:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:28634:31 | flags / object initializer | 0 | src/compiler/types.ts:4274:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:33599:21 | flags / \|= | number | src/compiler/types.ts:6038:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:33606:25 | flags / \|= | number | src/compiler/types.ts:6038:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/checker.ts:37724:17 | flags / \|= | number | src/compiler/types.ts:6038:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: mergeJSSymbols. |
| src/compiler/commandLineParser.ts:128:5 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:329:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:339:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:352:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:367:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:377:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:385:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:394:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:402:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:409:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:416:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:423:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:431:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:438:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:445:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:452:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:461:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:470:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:479:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:490:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:500:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:511:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:521:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:530:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:540:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:549:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:559:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:641:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:650:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:658:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:667:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:676:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:685:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:693:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:720:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:730:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:758:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:771:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:783:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:795:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:806:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:818:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:828:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:852:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:861:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:869:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:879:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:888:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:897:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:907:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:920:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:930:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:940:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:950:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:960:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:970:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:980:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:990:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1000:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1010:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1022:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1031:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1040:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1049:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1058:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1068:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1077:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1086:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1117:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1127:9 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1143:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1158:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1171:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1181:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1190:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1201:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1208:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1220:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1229:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1239:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1248:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1256:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1267:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1275:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1286:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1295:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1304:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1315:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1325:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1337:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1344:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1351:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1363:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1371:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1380:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1393:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1402:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1411:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1418:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1440:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1449:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1460:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1471:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1480:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1488:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1496:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1504:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1512:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1521:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1530:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1540:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1549:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1561:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1570:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1580:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1590:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1599:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1608:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1616:9 | type / object initializer | "number" | src/compiler/types.ts:7791:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1624:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1633:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1643:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1653:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1665:13 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1685:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1735:5 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1752:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1760:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1768:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1775:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1782:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1798:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1806:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1814:13 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1819:9 | type / object initializer | "boolean" | src/compiler/types.ts:7797:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2346:9 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2353:5 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2359:5 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2365:5 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2374:13 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2385:25 | type / object initializer | "object" | src/compiler/types.ts:7824:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2394:25 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2403:25 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2413:25 | type / object initializer | "string" | src/compiler/types.ts:7785:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/expressionToTypeNode.ts:167:20 | reportFallback / object initializer | boolean | src/compiler/expressionToTypeNode.ts:162:26; src/compiler/expressionToTypeNode.ts:163:26 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:733:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:747:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:758:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:770:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:805:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:822:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:841:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:849:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:869:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:882:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:898:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:918:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:937:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:961:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:973:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:996:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1012:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1023:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1036:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1117:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1154:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1173:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1187:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1214:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1224:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1284:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1347:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1372:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1383:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1417:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1454:5 | scoped / object initializer | false | src/compiler/types.ts:8420:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1469:5 | scoped / object initializer | true | src/compiler/types.ts:8415:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/emitHelpers.ts:1477:5 | scoped / object initializer | true | src/compiler/types.ts:8415:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/nodeFactory.ts:1335:39 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:2936:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3001:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3071:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3370:9 | operator / = | PrefixUnaryOperator | src/compiler/types.ts:2433:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3396:9 | operator / = | PostfixUnaryOperator | src/compiler/types.ts:2445:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3778:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:4278:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:4552:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6049:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6121:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6136:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6328:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6337:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6352:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6361:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:6395:9 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:7395:5 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: makeSynthetic. |
| src/compiler/parser.ts:1403:5 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: updateSourceFile. |
| src/compiler/parser.ts:1857:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: withJSDoc. |
| src/compiler/parser.ts:2603:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: finishNode. |
| src/compiler/parser.ts:2611:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: finishNode. |
| src/compiler/parser.ts:4447:13 | flags / = | NodeFlags | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/parser.ts:6406:21 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: tryReparseOptionalChain. |
| src/compiler/parser.ts:6514:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: parseTaggedTemplateRest. |
| src/compiler/parser.ts:7482:17 | flags / \|= | number | src/compiler/types.ts:944:5 | not established: Modifier returned by parsing; freshness/reuse and earlier publication need an interprocedural parser analysis. |
| src/compiler/parser.ts:8113:21 | flags / \|= | number | src/compiler/types.ts:944:5 | not established: Modifier returned by parsing; freshness/reuse and earlier publication need an interprocedural parser analysis. |
| src/compiler/parser.ts:9686:21 | flags / \|= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/program.ts:3313:9 | flags / &= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/program.ts:3314:9 | flags / &= | number | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/sourcemap.ts:534:69 | done / object initializer | false | ../api/node_modules/typescript/lib/lib.es2015.iterable.d.ts:28:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sourcemap.ts:559:37 | done / object initializer | true | src/compiler/sourcemap.ts:557:47 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:1296:34 | computed / object initializer | false | src/compiler/factory/emitHelpers.ts:93:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:1299:34 | computed / object initializer | true | src/compiler/factory/emitHelpers.ts:92:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:1304:38 | computed / object initializer | true | src/compiler/factory/emitHelpers.ts:92:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:1309:38 | computed / object initializer | true | src/compiler/factory/emitHelpers.ts:92:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/module/module.ts:2501:5 | scoped / object initializer | true | src/compiler/types.ts:8415:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1148:17 | type / object initializer | UpToDateStatusType.UpToDate | src/compiler/tsbuild.ts:80:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1154:52 | type / object initializer | UpToDateStatusType.Unbuildable | src/compiler/tsbuild.ts:64:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1471:45 | type / object initializer | UpToDateStatusType.ContainerOnly | src/compiler/tsbuild.ts:72:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1477:49 | type / object initializer | UpToDateStatusType.ComputingUpstream | src/compiler/tsbuild.ts:153:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1500:21 | type / object initializer | UpToDateStatusType.UpstreamBlocked | src/compiler/tsbuild.ts:144:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1509:25 | type / object initializer | UpToDateStatusType.ForceBuild | src/compiler/tsbuild.ts:172:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1526:13 | type / object initializer | UpToDateStatusType.OutputMissing | src/compiler/tsbuild.ts:93:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1535:13 | type / object initializer | UpToDateStatusType.ErrorReadingFile | src/compiler/tsbuild.ts:102:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1542:13 | type / object initializer | UpToDateStatusType.TsVersionOutputOfDate | src/compiler/tsbuild.ts:157:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1553:13 | type / object initializer | UpToDateStatusType.OutOfDateBuildInfoWithErrors | src/compiler/tsbuild.ts:119:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1572:17 | type / object initializer | UpToDateStatusType.OutOfDateBuildInfoWithErrors | src/compiler/tsbuild.ts:119:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1587:17 | type / object initializer | UpToDateStatusType.OutOfDateBuildInfoWithPendingEmit | src/compiler/tsbuild.ts:119:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1606:17 | type / object initializer | UpToDateStatusType.OutOfDateOptions | src/compiler/tsbuild.ts:119:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1626:17 | type / object initializer | UpToDateStatusType.Unbuildable | src/compiler/tsbuild.ts:64:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1648:21 | type / object initializer | UpToDateStatusType.OutOfDateWithSelf | src/compiler/tsbuild.ts:110:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1680:13 | type / object initializer | UpToDateStatusType.OutOfDateRoots | src/compiler/tsbuild.ts:127:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1704:21 | type / object initializer | UpToDateStatusType.OutputMissing | src/compiler/tsbuild.ts:93:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1712:21 | type / object initializer | UpToDateStatusType.OutOfDateWithSelf | src/compiler/tsbuild.ts:110:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1740:21 | type / object initializer | UpToDateStatusType.OutOfDateWithUpstream | src/compiler/tsbuild.ts:166:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1757:17 | type / object initializer | UpToDateStatusType.OutOfDateWithUpstream | src/compiler/tsbuild.ts:166:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1782:9 | type / object initializer | UpToDateStatusType.UpToDate \| UpToDateStatusType.UpToDateWithUpstreamTypes \| UpToDateStatusType.UpToDateWithInputFileText | src/compiler/tsbuild.ts:80:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1800:18 | type / object initializer | UpToDateStatusType.Unbuildable | src/compiler/tsbuild.ts:64:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1885:9 | type / object initializer | UpToDateStatusType.UpToDate | src/compiler/tsbuild.ts:80:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/tsbuildPublic.ts:1921:29 | type / = | UpToDateStatusType.UpToDateWithUpstreamTypes | src/compiler/tsbuild.ts:80:9 | outside local construction: Retagging a cached build-status record, UpToDate to UpToDateWithUpstreamTypes, so downstream output timestamps are refreshed without rebuilding unchanged declarations. Not a syntax node. |
| src/compiler/tsbuildPublic.ts:1930:33 | type / object initializer | UpToDateStatusType.OutOfDateWithUpstream | src/compiler/tsbuild.ts:166:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/utilities.ts:969:13 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: aggregateChildData. |
| src/compiler/utilities.ts:975:9 | flags / \|= | number | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: aggregateChildData. |
| src/compiler/utilities.ts:8473:5 | flags / = | SymbolFlags | src/compiler/types.ts:6038:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8491:5 | flags / = | TypeFlags | src/compiler/types.ts:6440:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8499:5 | flags / = | SignatureFlags | src/compiler/types.ts:7020:22 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8511:5 | flags / = | NodeFlags.None | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8525:5 | flags / = | NodeFlags.None | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:8537:5 | flags / = | NodeFlags.None | src/compiler/types.ts:944:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/utilities.ts:10689:9 | flags / = | NodeFlags | src/compiler/types.ts:944:5 | outside local construction: Updating semantic/parser metadata flags on an incoming, cached, reused, or possibly aliased node/type/symbol/signature; not changing SyntaxKind. Function: setNodeFlags. |

## Broader and partial checker candidates

| Write | Property / form | Written value type | Declaration(s) | Classification / purpose |
| --- | --- | --- | --- | --- |
| src/compiler/binder.ts:496:64 | antecedent / object initializer | FlowNode \| FlowNode[] \| undefined | src/compiler/types.ts:4192:5; src/compiler/types.ts:4201:5; src/compiler/types.ts:4208:5; src/compiler/types.ts:4216:5; src/compiler/types.ts:4230:5; src/compiler/types.ts:4237:5; src/compiler/types.ts:4252:5; src/compiler/types.ts:4222:5; src/compiler/types.ts:4258:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/binder.ts:1375:35 | antecedent / = | never[] | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/binder.ts:1688:13 | antecedent / = | FlowNode[] \| undefined | src/compiler/types.ts:4208:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/binder.ts:3193:13 | commonJsModuleIndicator / = | Node | src/compiler/types.ts:4411:22 | outside local construction: Setting/recomputing module classification on a source-file argument, rather than allocating a source file in this function. |
| src/compiler/builder.ts:531:15 | file / object spread | SourceFile \| undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builder.ts:606:9 | file / object spread | string \| false \| undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builder.ts:607:9 | file / object initializer | SourceFile \| undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builder.ts:1327:39 | signature / object initializer | false | src/compiler/builder.ts:1327:39 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builder.ts:1939:37 | signature / = | string | src/compiler/builderState.ts:106:9 | outside local construction: Updating an existing incremental-builder file-info record with a computed declaration signature. |
| src/compiler/builder.ts:1943:37 | signature / = | string | src/compiler/builderState.ts:106:9 | outside local construction: Updating an existing incremental-builder file-info record with a computed declaration signature. |
| src/compiler/builder.ts:2231:30 | signature / object initializer | string | src/compiler/builderState.ts:106:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builder.ts:2234:38 | signature / object initializer | string \| undefined | src/compiler/builderState.ts:106:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builder.ts:2271:75 | signature / object initializer | undefined | src/compiler/builderState.ts:106:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builderState.ts:332:17 | signature / object initializer | string \| undefined | src/compiler/builderState.ts:106:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/builderState.ts:395:9 | signature / = | string \| undefined | src/compiler/builderState.ts:106:9 | outside local construction: Updating an existing incremental-builder file-info record with a computed declaration signature. |
| src/compiler/builderState.ts:458:9 | signature / = | string | src/compiler/builderState.ts:106:9 | outside local construction: Updating an existing incremental-builder file-info record with a computed declaration signature. |
| src/compiler/checker.ts:2564:41 | file / object spread | SourceFile \| undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:9227:53 | sym / object initializer | Symbol \| undefined | src/compiler/checker.ts:9227:53 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:16122:24 | parameterName / object initializer | string \| undefined | src/compiler/types.ts:5697:5; src/compiler/types.ts:5704:5; src/compiler/types.ts:5711:5; src/compiler/types.ts:5718:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:16122:39 | parameterIndex / object initializer | number \| undefined | src/compiler/types.ts:5698:5; src/compiler/types.ts:5705:5; src/compiler/types.ts:5712:5; src/compiler/types.ts:5719:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:21705:44 | innerExpression / object initializer | Expression \| undefined | src/compiler/checker.ts:21705:44 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:21711:44 | innerExpression / object initializer | undefined | src/compiler/checker.ts:21711:44 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:21711:82 | errorMessage / object initializer | DiagnosticMessage | src/compiler/checker.ts:21711:82 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:21716:44 | innerExpression / object initializer | JsxElement \| JsxSelfClosingElement \| JsxFragment | src/compiler/checker.ts:21716:44 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/checker.ts:28922:17 | antecedent / = | FlowNode[] | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/checker.ts:28924:17 | antecedent / = | FlowNode[] \| undefined | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/checker.ts:28966:17 | antecedent / = | FlowNode[] | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/checker.ts:28968:17 | antecedent / = | FlowNode[] \| undefined | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/checker.ts:29101:21 | antecedent / = | FlowNode[] | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/checker.ts:29103:21 | antecedent / = | FlowNode[] \| undefined | src/compiler/types.ts:4208:5 | outside local construction: Reusing a control-flow label: add incoming edges, or temporarily replace and restore the antecedents during finally/reduce-label analysis. |
| src/compiler/checker.ts:36742:34 | file / object initializer | SourceFile \| undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:291:9 | type / object initializer | Map<string, WatchFileKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:305:9 | type / object initializer | Map<string, WatchDirectoryKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:317:9 | type / object initializer | Map<string, PollingWatchKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:336:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:349:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:575:5 | type / object initializer | Map<string, ScriptTarget> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:608:5 | type / object initializer | Map<string, ModuleKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:706:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:709:13 | type / object initializer | Map<string, string> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:741:9 | type / object initializer | Map<string, JsxEmit> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:838:9 | type / object initializer | Map<string, ImportsNotUsedAsValues> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1098:9 | type / object initializer | Map<string, ModuleResolutionKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1139:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1155:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1168:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1217:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1264:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1427:9 | type / object initializer | Map<string, NewLineKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1661:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1672:9 | type / object initializer | Map<string, ModuleDetectionKind> | src/compiler/types.ts:7803:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1803:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:1811:9 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2343:5 | type / object initializer | "listOrElement" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2382:21 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2391:21 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2400:21 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/commandLineParser.ts:2410:21 | type / object initializer | "list" | src/compiler/types.ts:7831:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/expressionToTypeNode.ts:167:14 | type / object initializer | TypeNode \| undefined | src/compiler/expressionToTypeNode.ts:161:9; src/compiler/expressionToTypeNode.ts:162:9; src/compiler/expressionToTypeNode.ts:163:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/factory/nodeFactory.ts:2332:9 | modifiers / = | undefined | src/compiler/types.ts:2232:22 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:2357:13 | modifiers / = | undefined | src/compiler/types.ts:2232:22 | outside local construction: Updating a receiver supplied to this function; it does not allocate that receiver. Function: finishUpdateFunctionTypeNode. |
| src/compiler/factory/nodeFactory.ts:2376:9 | modifiers / = | NodeArray<Modifier> \| undefined | src/compiler/types.ts:2237:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3187:9 | modifiers / = | NodeArray<Modifier> \| undefined | src/compiler/types.ts:2753:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:3255:9 | modifiers / = | NodeArray<Modifier> \| undefined | src/compiler/types.ts:2760:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:4310:9 | modifiers / = | NodeArray<Modifier> \| undefined | src/compiler/types.ts:2087:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/factory/nodeFactory.ts:4376:17 | modifiers / = | NodeArray<ModifierLike> \| undefined | src/compiler/types.ts:2087:5 | outside local construction: Updating a receiver supplied to this function; it does not allocate that receiver. Function: finishUpdateFunctionDeclaration. |
| src/compiler/factory/nodeFactory.ts:6075:9 | externalModuleIndicator / = | undefined | src/compiler/types.ts:4401:5 | inside: Reviewed fresh allocation/clone in this function, before return or publication; allocator this is a fresh constructor receiver. |
| src/compiler/parser.ts:1341:5 | externalModuleIndicator / = | Node \| undefined | src/compiler/types.ts:4401:5 | outside local construction: Setting/recomputing module classification on a source-file argument, rather than allocating a source file in this function. |
| src/compiler/program.ts:4952:51 | directoryExists / object initializer | ((path: string) => boolean) \| undefined | src/compiler/program.ts:4952:51 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sys.ts:1619:30 | module / object initializer | any | src/compiler/types.ts:7739:9; src/compiler/types.ts:7738:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sys.ts:1619:59 | modulePath / object initializer | string | src/compiler/types.ts:7738:20 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sys.ts:1619:71 | error / object initializer | undefined | src/compiler/types.ts:7738:41 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sys.ts:1622:30 | module / object initializer | undefined | src/compiler/types.ts:7739:9 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sys.ts:1622:49 | modulePath / object initializer | undefined | src/compiler/types.ts:7739:28; src/compiler/types.ts:7738:20 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/sys.ts:1622:72 | error / object initializer | any | src/compiler/types.ts:7739:52; src/compiler/types.ts:7738:41 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/transformers/esDecorators.ts:1407:69 | extraInitializersName / object initializer | Identifier \| undefined | src/compiler/transformers/esDecorators.ts:1407:69 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/utilities.ts:8622:9 | file / object initializer | undefined | src/compiler/types.ts:7307:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/utilities.ts:8724:9 | file / object initializer | undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/utilities.ts:8739:9 | file / object initializer | undefined | src/compiler/types.ts:7293:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/utilities.ts:8987:17 | externalModuleIndicator / = | true \| Node \| undefined | src/compiler/types.ts:4401:5 | outside local construction: Setting/recomputing module classification on a source-file argument, rather than allocating a source file in this function. |
| src/compiler/utilities.ts:8992:17 | externalModuleIndicator / = | Node \| undefined | src/compiler/types.ts:4401:5 | outside local construction: Setting/recomputing module classification on a source-file argument, rather than allocating a source file in this function. |
| src/compiler/utilities.ts:9004:58 | externalModuleIndicator / = | true \| Node \| undefined | src/compiler/types.ts:4401:5 | outside local construction: Setting/recomputing module classification on a source-file argument, rather than allocating a source file in this function. |
| src/compiler/watch.ts:846:9 | useCaseSensitiveFileNames / object initializer | () => boolean | src/compiler/watchPublic.ts:181:5 | inside: Direct object creation; this property is initialized before publication of the literal. |
| src/compiler/watchPublic.ts:808:17 | version / = | false | src/compiler/watchPublic.ts:428:9 | outside local construction: Reusing a watch-host cache entry and changing its version state to false (file presence unknown). |

## Unresolved dynamic-key writes and coverage limits

There are **201** bracket-write sites whose key type is not a finite union of
literal names. They are not included in the 437 resolved property events.
Many are array or dictionary slots. Some involve `any` or generic clones:
`cloneSourceFileWorker` and `cloneNode` use property-copy loops whose runtime
keys may include tag names. These loops skip properties already initialized on
the clone, but proving that skip for every reachable allocator and input needs
runtime-key/provenance reasoning. The stock checker cannot resolve a single
written property declaration there. This census does not call them zero writes.
Each site and its checked receiver/key type follows. Arbitrary user-defined
setters, aliases of reflective builtins, defineProperties descriptor maps,
prototype inheritance and mutations performed by called functions are not
expanded into implicit per-property writes. Consequently this is an exhaustive
ledger for the described AST/resolved-property rule, not a proof that every
possible runtime discriminant mutation has been counted.

| Write | Receiver type | Key type |
| --- | --- | --- |
| src/compiler/binder.ts:1933:17 | (boolean \| undefined)[] | number |
| src/compiler/binder.ts:1934:17 | (Node \| undefined)[] | number |
| src/compiler/builder.ts:1423:9 | IncrementalBuildInfoRoot[] | number |
| src/compiler/builder.ts:1447:17 | CompilerOptions | string |
| src/compiler/checker.ts:2679:9 | Symbol[] | number |
| src/compiler/checker.ts:2935:16 | SymbolLinks[] | number |
| src/compiler/checker.ts:2940:38 | NodeLinks[] | number |
| src/compiler/checker.ts:7443:37 | TypeNode[] | number |
| src/compiler/checker.ts:7452:37 | TypeNode[] | number |
| src/compiler/checker.ts:7960:33 | TypeNode[] | number |
| src/compiler/checker.ts:9521:37 | Statement[] | number |
| src/compiler/checker.ts:9534:25 | Statement[] | number |
| src/compiler/checker.ts:11493:17 | boolean[] | number |
| src/compiler/checker.ts:14236:21 | __String[] | number |
| src/compiler/checker.ts:14434:13 | Symbol[] | number |
| src/compiler/checker.ts:14442:13 | Symbol[] | number |
| src/compiler/checker.ts:14519:13 | boolean[] | number |
| src/compiler/checker.ts:14584:21 | IndexInfo[] | number |
| src/compiler/checker.ts:16161:17 | Type[] | number |
| src/compiler/checker.ts:16169:17 | Type[] | number |
| src/compiler/checker.ts:17885:39 | TypeParameter[] | number |
| src/compiler/checker.ts:17998:59 | ElementFlags[] | number |
| src/compiler/checker.ts:18002:13 | Type[] | number |
| src/compiler/checker.ts:18583:13 | Type[] | number |
| src/compiler/checker.ts:18636:9 | Type[] | number |
| src/compiler/checker.ts:18707:13 | Type[] | number |
| src/compiler/checker.ts:18822:21 | Type[] | number |
| src/compiler/checker.ts:23381:13 | string[] | number |
| src/compiler/checker.ts:23386:17 | Type[] | number |
| src/compiler/checker.ts:23391:17 | Type[] | number |
| src/compiler/checker.ts:24223:17 | Type[][] | number |
| src/compiler/checker.ts:24962:29 | Ternary[] | number |
| src/compiler/checker.ts:24970:21 | Ternary[] | number |
| src/compiler/checker.ts:27159:65 | boolean[] | number |
| src/compiler/checker.ts:28648:55 | EvolvingArrayType[] | number |
| src/compiler/checker.ts:28873:67 | (boolean \| undefined)[] | number |
| src/compiler/checker.ts:28942:67 | (boolean \| undefined)[] | number |
| src/compiler/checker.ts:29127:21 | FlowNode[] | number |
| src/compiler/checker.ts:29128:21 | FlowType[] | number |
| src/compiler/checker.ts:29376:50 | Map<string, Type>[] | number |
| src/compiler/checker.ts:29414:21 | FlowNode[] | number |
| src/compiler/checker.ts:29415:21 | string[] | number |
| src/compiler/checker.ts:29416:21 | Type[][] | number |
| src/compiler/checker.ts:32872:9 | Node[] | number |
| src/compiler/checker.ts:32873:9 | (Type \| undefined)[] | number |
| src/compiler/checker.ts:32874:9 | boolean[] | number |
| src/compiler/checker.ts:32881:9 | Node[] | number |
| src/compiler/checker.ts:32882:9 | (Type \| undefined)[] | number |
| src/compiler/checker.ts:32883:9 | boolean[] | number |
| src/compiler/checker.ts:32896:9 | Node[] | number |
| src/compiler/checker.ts:32897:9 | (InferenceContext \| undefined)[] | number |
| src/compiler/checker.ts:32903:9 | Node[] | number |
| src/compiler/checker.ts:32904:9 | (InferenceContext \| undefined)[] | number |
| src/compiler/checker.ts:32916:9 | TypeMapper[] | number |
| src/compiler/checker.ts:32917:9 | Map<string, Type>[] | number |
| src/compiler/checker.ts:32924:9 | TypeMapper[] | number |
| src/compiler/checker.ts:33142:13 | Symbol[] | number |
| src/compiler/checker.ts:33150:13 | Symbol[] | number |
| src/compiler/checker.ts:36868:17 | Signature[] | number |
| src/compiler/checker.ts:36964:9 | Signature[] | number |
| src/compiler/checker.ts:40598:13 | (Type \| undefined)[] | number |
| src/compiler/checker.ts:40610:13 | (Type \| undefined)[] | number |
| src/compiler/checker.ts:41616:17 | InferenceInfo[] | number |
| src/compiler/checker.ts:41701:13 | Type[] | number |
| src/compiler/commandLineParser.ts:2050:13 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2055:17 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2077:21 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2083:21 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2090:21 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2095:21 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2105:21 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2111:13 | OptionsBase | string |
| src/compiler/commandLineParser.ts:2515:21 | any | string |
| src/compiler/commandLineParser.ts:2693:17 | Record<string, CompilerOptionsValue> | string |
| src/compiler/commandLineParser.ts:2977:13 | CompilerOptions | string |
| src/compiler/commandLineParser.ts:3295:9 | OptionsBase | string |
| src/compiler/commandLineParser.ts:3313:9 | string[] | number |
| src/compiler/commandLineParser.ts:3325:9 | MapLike<string[]> | string |
| src/compiler/commandLineParser.ts:3613:17 | CompilerOptions \| WatchOptions \| TypeAcquisition | string |
| src/compiler/commandLineParser.ts:3785:13 | CompilerOptions \| WatchOptions \| TypeAcquisition | string |
| src/compiler/commandLineParser.ts:4144:21 | MapLike<WatchDirectoryFlags> | string |
| src/compiler/commandLineParser.ts:4159:25 | MapLike<WatchDirectoryFlags> | string |
| src/compiler/commandLineParser.ts:4269:17 | CompilerOptions | string |
| src/compiler/core.ts:303:13 | T[] | number |
| src/compiler/core.ts:1169:5 | T[] | number |
| src/compiler/core.ts:1327:9 | any[] | number |
| src/compiler/core.ts:1355:17 | T | Extract<keyof T, string> |
| src/compiler/core.ts:1428:9 | (T \| U)[] | number |
| src/compiler/core.ts:1470:27 | Record<string, T[]> | string |
| src/compiler/core.ts:1482:13 | any | Extract<keyof T, string> |
| src/compiler/core.ts:1499:13 | any | Extract<keyof T2, string> |
| src/compiler/core.ts:1505:13 | any | Extract<keyof T1, string> |
| src/compiler/core.ts:1516:13 | any | Extract<keyof T2, string> |
| src/compiler/core.ts:1587:9 | (T \| undefined)[] | number |
| src/compiler/core.ts:2205:9 | any[] | number |
| src/compiler/core.ts:2216:13 | any[] | number |
| src/compiler/core.ts:2226:13 | any[] | number |
| src/compiler/core.ts:2230:13 | any[] | number |
| src/compiler/core.ts:2343:9 | T[] | number |
| src/compiler/core.ts:2350:5 | T[] | number |
| src/compiler/debug.ts:189:13 | Partial<Record<AssertionKeys, { level: AssertionLevel; assertion: AnyFunction; }>> | K |
| src/compiler/debug.ts:190:13 | any | K |
| src/compiler/debug.ts:1014:17 | Record<number, FlowGraphNode> | number |
| src/compiler/debug.ts:1059:17 | number[] | number |
| src/compiler/debug.ts:1137:17 | (FlowGraphNode \| undefined)[] | number |
| src/compiler/debug.ts:1145:21 | Connection[] | number |
| src/compiler/debug.ts:1148:21 | Connection[] | number |
| src/compiler/debug.ts:1156:21 | Connection[] | number |
| src/compiler/debug.ts:1169:25 | Connection[] | number |
| src/compiler/debug.ts:1199:17 | string[] | number |
| src/compiler/debug.ts:1237:21 | T[] | number |
| src/compiler/emitter.ts:2855:17 | (boolean \| undefined)[] | number |
| src/compiler/emitter.ts:2856:17 | number[] | number |
| src/compiler/emitter.ts:2857:17 | number[] | number |
| src/compiler/emitter.ts:2858:17 | number[] | number |
| src/compiler/emitter.ts:2859:38 | boolean[] | number |
| src/compiler/emitter.ts:2860:40 | boolean[] | number |
| src/compiler/emitter.ts:5467:71 | string[] | number |
| src/compiler/emitter.ts:5474:34 | string[] | number |
| src/compiler/executeCommandLine.ts:408:30 | { [value: string]: string[]; } | string \| number |
| src/compiler/factory/emitNode.ts:303:13 | EmitHelper[] | number |
| src/compiler/factory/nodeFactory.ts:6145:13 | any | string |
| src/compiler/factory/nodeFactory.ts:6404:13 | T | Extract<keyof T, string> |
| src/compiler/factory/nodeFactory.ts:7539:9 | (TextRange \| undefined)[] | string |
| src/compiler/factory/utilities.ts:1283:9 | TState[] | number |
| src/compiler/factory/utilities.ts:1284:9 | BinaryExpressionState[] | number |
| src/compiler/factory/utilities.ts:1297:9 | BinaryExpressionState[] | number |
| src/compiler/factory/utilities.ts:1315:9 | BinaryExpressionState[] | number |
| src/compiler/factory/utilities.ts:1329:9 | BinaryExpressionState[] | number |
| src/compiler/factory/utilities.ts:1346:9 | BinaryExpressionState[] | number |
| src/compiler/factory/utilities.ts:1352:17 | TState[] | number |
| src/compiler/factory/utilities.ts:1394:9 | BinaryExpressionState[] | number |
| src/compiler/factory/utilities.ts:1395:9 | BinaryExpression[] | number |
| src/compiler/factory/utilities.ts:1396:9 | TState[] | number |
| src/compiler/parser.ts:9037:25 | string[] | number |
| src/compiler/parser.ts:10734:25 | { [index: string]: string \| { value: string; pos: number; end: number; }; } | string |
| src/compiler/parser.ts:10741:25 | { [index: string]: string \| { value: string; pos: number; end: number; }; } | string |
| src/compiler/parser.ts:10794:9 | { [index: string]: string; } | string |
| src/compiler/program.ts:2284:21 | any[] | number |
| src/compiler/program.ts:2302:21 | any[] | number |
| src/compiler/program.ts:2315:52 | Resolution[] | number |
| src/compiler/scanner.ts:404:9 | T[] | number |
| src/compiler/sourcemap.ts:109:13 | (string \| null)[] | number |
| src/compiler/sourcemap.ts:214:21 | number[] | number |
| src/compiler/sourcemap.ts:226:25 | number[] | number |
| src/compiler/sourcemap.ts:760:28 | SourceMappedPosition[][] | number |
| src/compiler/symbolWalker.ts:79:13 | Type[] | number |
| src/compiler/symbolWalker.ts:193:13 | Symbol[] | number |
| src/compiler/sys.ts:200:13 | (T \| undefined)[] | number |
| src/compiler/sys.ts:209:13 | (T \| undefined)[] | number |
| src/compiler/sys.ts:218:17 | (T \| undefined)[] | number |
| src/compiler/sys.ts:219:17 | (T \| undefined)[] | number |
| src/compiler/sys.ts:328:21 | (WatchedFileWithUnchangedPolls \| undefined)[] | number |
| src/compiler/sys.ts:338:17 | (WatchedFileWithUnchangedPolls \| undefined)[] | number |
| src/compiler/sys.ts:343:17 | (WatchedFileWithUnchangedPolls \| undefined)[] | number |
| src/compiler/sys.ts:1802:21 | Buffer<ArrayBufferLike> | number |
| src/compiler/sys.ts:1803:21 | Buffer<ArrayBufferLike> | number |
| src/compiler/transformer.ts:489:9 | VariableDeclaration[][] | number |
| src/compiler/transformer.ts:490:9 | FunctionDeclaration[][] | number |
| src/compiler/transformer.ts:491:9 | Statement[][] | number |
| src/compiler/transformer.ts:492:9 | LexicalEnvironmentFlags[] | number |
| src/compiler/transformer.ts:592:9 | Identifier[][] | number |
| src/compiler/transformers/classFields.ts:1932:13 | Identifier[] | number |
| src/compiler/transformers/classFields.ts:2052:21 | Identifier[] | number |
| src/compiler/transformers/es2017.ts:670:17 | boolean[] | number |
| src/compiler/transformers/es2017.ts:826:21 | boolean[] | number |
| src/compiler/transformers/es2018.ts:1198:13 | boolean[] | number |
| src/compiler/transformers/generators.ts:2116:9 | number[] | number |
| src/compiler/transformers/generators.ts:2125:9 | number[] | number |
| src/compiler/transformers/generators.ts:2142:9 | BlockAction[] | number |
| src/compiler/transformers/generators.ts:2143:9 | number[] | number |
| src/compiler/transformers/generators.ts:2144:9 | CodeBlock[] | number |
| src/compiler/transformers/generators.ts:2157:9 | BlockAction[] | number |
| src/compiler/transformers/generators.ts:2158:9 | number[] | number |
| src/compiler/transformers/generators.ts:2159:9 | CodeBlock[] | number |
| src/compiler/transformers/generators.ts:2246:13 | Identifier[] | number |
| src/compiler/transformers/generators.ts:2529:17 | Mutable<LiteralExpression>[][] | number |
| src/compiler/transformers/generators.ts:2734:9 | OpCode[] | number |
| src/compiler/transformers/generators.ts:2735:9 | (OperationArguments \| undefined)[] | number |
| src/compiler/transformers/generators.ts:2736:9 | (TextRange \| undefined)[] | number |
| src/compiler/transformers/generators.ts:2948:21 | number[][] | number |
| src/compiler/transformers/legacyDecorators.ts:779:13 | Identifier[] | number |
| src/compiler/transformers/module/module.ts:243:9 | ExternalModuleInfo[] | number |
| src/compiler/transformers/module/module.ts:1176:21 | boolean[] | number |
| src/compiler/transformers/module/module.ts:1182:21 | boolean[] | number |
| src/compiler/transformers/module/module.ts:2344:13 | boolean[] | number |
| src/compiler/transformers/module/module.ts:2358:13 | boolean[] | number |
| src/compiler/transformers/module/module.ts:2444:21 | boolean[] | number |
| src/compiler/transformers/module/system.ts:206:22 | ExternalModuleInfo[] | number |
| src/compiler/transformers/module/system.ts:211:9 | Identifier[] | number |
| src/compiler/transformers/module/system.ts:212:25 | Identifier[] | number |
| src/compiler/transformers/module/system.ts:261:13 | boolean[][] | number |
| src/compiler/transformers/module/system.ts:1779:17 | boolean[][] | number |
| src/compiler/transformers/module/system.ts:2038:9 | boolean[] | number |
| src/compiler/transformers/utilities.ts:383:9 | V[][] | number |
| src/compiler/transformers/utilities.ts:666:17 | (readonly Decorator[] \| undefined)[] | number |
| src/compiler/tsbuildPublic.ts:334:53 | CompilerOptions | string |
| src/compiler/utilities.ts:10502:9 | Uint16Array<ArrayBuffer> | number |
| src/compiler/utilities.ts:10504:23 | Uint16Array<ArrayBuffer> | number |
| src/compiler/utilities.ts:10516:13 | Uint16Array<ArrayBuffer> | number |
| src/compiler/visitorPublic.ts:427:13 | ParameterDeclaration[] | number |

## Reproduction and validation

```
NODE_PATH=<stock-api-node_modules> node stage3/fixtures/assertions/discriminant-writes.cjs <prepared-TypeScript-root> <census-directory> <scratch-output> > <scan-log> 2>&1
python3 stage3/fixtures/assertions/discriminant-report.py <scratch-output>/discriminant-writes-raw.json stage3/fixtures/assertions > <report-log> 2>&1
python3 stage3/fixtures/assertions/discriminant-report.py <scratch-output>/discriminant-writes-raw.json <scratch-output>/mutant --mutant-omit-status-write > <mutant-log> 2>&1
```

Use the prepared upstream tree and stock npm environment from the Reproduction section of this bucket's
`README.md`. The scan finished with zero diagnostics and 437 resolved events,
201 unresolved dynamic-key sites. Report validation checks declaration presence,
unique write keys, the three exact constructor `kind` assignments, and the cached
status retag. The mutant removes only the status write from the real census input;
validation exits nonzero with `lost cached-status retag`. Three additional
mutants omit the Node constructor kind write, erase a resolved declaration,
and duplicate a write; each exits 1 at its corresponding invariant. This proves that an
omitted retag is caught by the report check. It does not prove escape analysis or
Adamic's new immutability rule, which this unit has not implemented or tested.
No fixtures, status.json, compiler code or another bucket were changed.
