# Binary expression census breakdown

Table pin: `dc6b1529ae9d2a2210672e105c8bb6374619a59d`. Compiler base: `44583d3283fdd8674085a7ddcce040cf2a73a94e`. Replay: `9a1f14c5d994aa855625e7cfa295677060348fec`.

Raw CSV SHA-256: `9bbb204bcdaf6d8ea75c9a343a8265fac804041088096c476337cff42f77905d`. All 81 adapted source hashes match the saved manifest. Counts deduplicate the raw table site key; they are observations on a checker-rejected entry project, not counts of programs proven to compile.

| Original row | Operator | Unique sites |
| --- | --- | ---: |
| a BinaryExpression with a value and a value | `&#124;&#124;` | 237 |
| a BinaryExpression with a value and a value | `&&` | 196 |
| a BinaryExpression with a value and a value | `=` | 87 |
| a BinaryExpression with a value and a value | `!==` | 60 |
| a BinaryExpression with a value and a boolean | `&&` | 349 |
| a BinaryExpression with a value and a boolean | `&#124;&#124;` | 15 |
| a BinaryExpression with a number and a number | `=` | 103 |
| a BinaryExpression with a number and a number | `&&` | 62 |
| a BinaryExpression with a number and a number | `&#124;&#124;` | 40 |
| a BinaryExpression with a boolean and a value | `&&` | 145 |
| a BinaryExpression with a boolean and a value | `&#124;&#124;` | 8 |
| a BinaryExpression with a number and a boolean | `&&` | 122 |
| a BinaryExpression with a number and a boolean | `&#124;&#124;` | 28 |

Total: **1,452 unique sites**. `selected-raw.csv` preserves the repeated attempt context; `sites.csv` records each exact failing node, its source expression, operator and checker types.

## Checker types by shape

| Original row | Operator | Left checker type | Right checker type | Whole checker type | Sites |
| --- | --- | --- | --- | --- | ---: |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | IntrinsicType | Type | 38 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | Type | Type | 19 |
| a BinaryExpression with a value and a value | !== | NodeArray<ModifierLike> &#124; undefined | readonly ModifierLike[] &#124; undefined | boolean | 16 |
| a BinaryExpression with a value and a value | = | Type | Type | Type | 16 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type | IntrinsicType | Type | 12 |
| a BinaryExpression with a value and a value | !== | NodeArray<TypeNode> &#124; undefined | readonly TypeNode[] &#124; undefined | boolean | 9 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | Type &#124; undefined | Type &#124; undefined | 9 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol &#124; undefined | TransientSymbol | Symbol | 9 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | Type &#124; undefined | Type &#124; undefined | 8 |
| a BinaryExpression with a value and a value | &#124;&#124; | DeclarationName &#124; undefined | Declaration &#124; undefined | ArrayBindingPattern &#124; BigIntLiteral &#124; ComputedPropertyName &#124; Declaration &#124; JsxNamespacedName &#124; ObjectBindingPattern &#124; PrivateIdentifier &#124; undefined | 7 |
| a BinaryExpression with a value and a value | &#124;&#124; | ModuleExportName &#124; undefined | ModuleExportName | ModuleExportName | 7 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 7 |
| a BinaryExpression with a value and a value | !== | NodeArray<Modifier> &#124; undefined | readonly Modifier[] &#124; undefined | boolean | 6 |
| a BinaryExpression with a value and a value | && | JSDocTypeTag &#124; undefined | JSDocTypeExpression | JSDocTypeExpression &#124; undefined | 6 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 6 |
| a BinaryExpression with a value and a value | &#124;&#124; | SymbolTable &#124; undefined | SymbolTable | SymbolTable | 6 |
| a BinaryExpression with a value and a value | !== | NodeArray<Statement> | readonly Statement[] | boolean | 5 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | Node | Node &#124; undefined | 5 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | SourceFile | SourceFile &#124; undefined | 5 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | Type | Type &#124; undefined | 5 |
| a BinaryExpression with a value and a value | = | Type | IntrinsicType | IntrinsicType | 5 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol &#124; undefined | Symbol | Symbol | 5 |
| a BinaryExpression with a value and a value | && | Expression &#124; undefined | Expression | Expression &#124; undefined | 4 |
| a BinaryExpression with a value and a value | && | Node | Node | Node | 4 |
| a BinaryExpression with a value and a value | && | SymbolTable &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 4 |
| a BinaryExpression with a value and a value | = | BaseType[] | never[] | never[] | 4 |
| a BinaryExpression with a value and a value | = | Node[] | never[] | never[] | 4 |
| a BinaryExpression with a value and a value | = | Type &#124; undefined | Type | Type | 4 |
| a BinaryExpression with a value and a value | &#124;&#124; | Diagnostic[] &#124; undefined | never[] | Diagnostic[] | 4 |
| a BinaryExpression with a value and a value | &#124;&#124; | IterationTypes &#124; undefined | IterationTypes &#124; undefined | IterationTypes &#124; undefined | 4 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | DeclarationName &#124; undefined | DeclarationName &#124; undefined | 3 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | TypeNode | TypeNode &#124; undefined | 3 |
| a BinaryExpression with a value and a value | && | TypeNode &#124; undefined | Type | Type &#124; undefined | 3 |
| a BinaryExpression with a value and a value | = | Map<string, CommandLineOption> | Map<string, CommandLineOption> | Map<string, CommandLineOption> | 3 |
| a BinaryExpression with a value and a value | = | Signature[] &#124; undefined | never[] | never[] | 3 |
| a BinaryExpression with a value and a value | &#124;&#124; | Date &#124; undefined | Date | Date | 3 |
| a BinaryExpression with a value and a value | &#124;&#124; | ModuleExportName &#124; undefined | Identifier | ModuleExportName | 3 |
| a BinaryExpression with a value and a value | &#124;&#124; | PropertyName &#124; undefined | Identifier | PropertyName | 3 |
| a BinaryExpression with a value and a value | &#124;&#124; | Signature[] &#124; undefined | never[] | Signature[] | 3 |
| a BinaryExpression with a value and a value | &#124;&#124; | SourceMapSource &#124; undefined | SourceMapSource | SourceMapSource | 3 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type[] &#124; undefined | Type[] &#124; undefined | Type[] &#124; undefined | 3 |
| a BinaryExpression with a value and a value | !== | NodeArray<Expression> | readonly Expression[] | boolean | 2 |
| a BinaryExpression with a value and a value | !== | NodeArray<JsxChild> | readonly JsxChild[] | boolean | 2 |
| a BinaryExpression with a value and a value | !== | NodeArray<ModifierLike> &#124; undefined | readonly Modifier[] &#124; undefined | boolean | 2 |
| a BinaryExpression with a value and a value | !== | NodeArray<ParameterDeclaration> | readonly ParameterDeclaration[] | boolean | 2 |
| a BinaryExpression with a value and a value | && | ClassLikeDeclaration &#124; undefined | ExpressionWithTypeArguments &#124; undefined | ExpressionWithTypeArguments &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Declaration[] &#124; undefined | SourceFile &#124; undefined | SourceFile &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | ExportCollisionTrackerTable &#124; undefined | ExportDeclaration &#124; undefined | ExportDeclaration &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | ExpressionWithTypeArguments &#124; undefined | BaseType[] | BaseType[] &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | FlowLabel &#124; undefined | FlowNode[] &#124; undefined | FlowNode[] &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Identifier &#124; undefined | NamedImportBindings &#124; undefined | NamedImportBindings &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | IterationTypes &#124; undefined | Type | Type &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | JSDocReturnTag &#124; undefined | JSDocTypeExpression &#124; undefined | JSDocTypeExpression &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | JSDocSignature &#124; SignatureDeclaration &#124; undefined | ClassDeclaration &#124; ClassExpression &#124; InterfaceDeclaration &#124; Node &#124; ObjectLiteralExpression &#124; TypeLiteralNode | ClassDeclaration &#124; ClassExpression &#124; InterfaceDeclaration &#124; Node &#124; ObjectLiteralExpression &#124; TypeLiteralNode &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | JSDocSignature &#124; SignatureDeclaration &#124; undefined | TypeNode &#124; undefined | TypeNode &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | JSDocThisTag &#124; undefined | JSDocTypeExpression | JSDocTypeExpression &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | JSDocTypeExpression &#124; undefined | TypeNode | TypeNode &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | ParameterDeclaration &#124; undefined | TypeNode &#124; undefined | TypeNode &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Signature &#124; undefined | TypePredicate &#124; undefined | TypePredicate &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Symbol | Symbol &#124; undefined | Symbol &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Symbol | SymbolTable | SymbolTable | 2 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | Symbol | Symbol &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | SymbolTable &#124; undefined | SymbolTable &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | Type | Type &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | Type &#124; undefined | Type &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Type | Type &#124; undefined | Type &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | TypeNode &#124; undefined | EntityNameOrEntityNameExpression &#124; undefined | EntityNameOrEntityNameExpression &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | Type[] &#124; undefined | Type | Type &#124; undefined | 2 |
| a BinaryExpression with a value and a value | && | readonly TypeParameter[] &#124; undefined | readonly TypeParameter[] &#124; undefined | readonly TypeParameter[] &#124; undefined | 2 |
| a BinaryExpression with a value and a value | = | AffectedFileResult<EmitResult> | AffectedFileResult<EmitResult> | AffectedFileResult<EmitResult> | 2 |
| a BinaryExpression with a value and a value | = | OptionsNameMap | OptionsNameMap | OptionsNameMap | 2 |
| a BinaryExpression with a value and a value | = | Signature | Signature | Signature | 2 |
| a BinaryExpression with a value and a value | = | TransientSymbol &#124; undefined | TransientSymbol | TransientSymbol | 2 |
| a BinaryExpression with a value and a value | = | VariableDeclaration[] &#124; undefined | never[] | never[] | 2 |
| a BinaryExpression with a value and a value | = | string[] &#124; undefined | never[] | never[] | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | CatchClause &#124; undefined | Block | Block &#124; CatchClause | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | DeclarationName &#124; undefined | Declaration | ArrayBindingPattern &#124; BigIntLiteral &#124; ComputedPropertyName &#124; Declaration &#124; JsxNamespacedName &#124; ObjectBindingPattern &#124; PrivateIdentifier | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | Expression &#124; undefined | Identifier | Expression | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | Identifier &#124; undefined | ClassLikeDeclaration | Identifier &#124; ClassLikeDeclaration | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | PathAndExtension &#124; undefined | PathAndExtension &#124; undefined | PathAndExtension &#124; undefined | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | PropertyName &#124; undefined | BindingName | ArrayBindingPattern &#124; BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; ObjectBindingPattern &#124; PrivateIdentifier &#124; StringLiteral | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | Signature &#124; undefined | Signature &#124; undefined | Signature &#124; undefined | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | StringLiteralType | Type | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | TypeNode &#124; undefined | TypeNode &#124; undefined | TypeNode &#124; undefined | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | WatchOptions &#124; undefined | WatchOptions &#124; undefined | WatchOptions &#124; undefined | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | readonly Diagnostic[] &#124; undefined | never[] | readonly Diagnostic[] | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | readonly TypeParameter[] &#124; undefined | readonly TypeParameter[] &#124; undefined | readonly TypeParameter[] &#124; undefined | 2 |
| a BinaryExpression with a value and a value | &#124;&#124; | string[] &#124; undefined | never[] | string[] | 2 |
| a BinaryExpression with a value and a value | !== | NodeArray<ArrayBindingElement> | readonly ArrayBindingElement[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<BindingElement> | readonly BindingElement[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<CaseOrDefaultClause> | readonly CaseOrDefaultClause[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<ExportSpecifier> | readonly ExportSpecifier[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<ExpressionWithTypeArguments> | readonly ExpressionWithTypeArguments[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<ImportAttribute> | readonly AssertEntry[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<ImportAttribute> | readonly ImportAttribute[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<ImportSpecifier> | readonly ImportSpecifier[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<JsxAttributeLike> | readonly JsxAttributeLike[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<NamedTupleMember &#124; TypeNode> | readonly (NamedTupleMember &#124; TypeNode)[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<ObjectLiteralElementLike> | readonly ObjectLiteralElementLike[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<TemplateLiteralTypeSpan> | readonly TemplateLiteralTypeSpan[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<TemplateSpan> | readonly TemplateSpan[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<TypeElement> &#124; undefined | readonly TypeElement[] &#124; undefined | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<TypeParameterDeclaration> | readonly TypeParameterDeclaration[] | boolean | 1 |
| a BinaryExpression with a value and a value | !== | NodeArray<VariableDeclaration> | readonly VariableDeclaration[] | boolean | 1 |
| a BinaryExpression with a value and a value | && | ((directoryName: string) => boolean) &#124; undefined | ((path: string) => string[]) &#124; undefined | ((path: string) => string[]) &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ArrayBindingPattern &#124; Expression &#124; ObjectBindingPattern &#124; undefined | Expression &#124; undefined | Expression &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Block &#124; undefined | ClassInfo &#124; undefined | ClassInfo &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | CallExpression &#124; undefined | NodeArray<Expression> | NodeArray<Expression> &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ClassImplementingOrExtendingExpressionWithTypeArguments &#124; undefined | InterfaceType | InterfaceType &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ConvertedLoopState &#124; undefined | (node: LabeledStatement) => void | ((node: LabeledStatement) => void) &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration | Symbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | Declaration &#124; undefined | Declaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | Declaration[] &#124; undefined | Declaration[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | Node | Node &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | ParameterDeclaration &#124; undefined | ParameterDeclaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration &#124; undefined | Type &#124; undefined | Type &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration[] &#124; undefined | Declaration &#124; undefined | Declaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Declaration[] &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | DiagnosticRelatedInformation[] &#124; undefined | DiagnosticRelatedInformation[] &#124; undefined | DiagnosticRelatedInformation[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | DirectoryWatcherCallback &#124; undefined | { dirName: string; callback: DirectoryWatcherCallback; } | { dirName: string; callback: DirectoryWatcherCallback; } &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | DotDotDotToken &#124; undefined | Expression &#124; undefined | Expression &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | EmitNode &#124; undefined | EmitHelper[] &#124; undefined | EmitHelper[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | EntityNameExpression &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | EntityNameOrEntityNameExpression | Identifier | Identifier | 1 |
| a BinaryExpression with a value and a value | && | Expression &#124; undefined | Expression &#124; undefined | Expression &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Expression &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | GetAccessorDeclaration &#124; undefined | TypeNode | TypeNode &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Identifier &#124; undefined | Map<string, boolean> &#124; undefined | Map<string, boolean> &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Identifier[] | Identifier &#124; undefined | Identifier &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ImportClause | Identifier &#124; undefined | Identifier &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ImportClause &#124; undefined | NamedImportBindings &#124; undefined | NamedImportBindings &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | IndexInfo &#124; undefined | Type | Type &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | InferenceContext &#124; undefined | IntraExpressionInferenceSite[] &#124; undefined | IntraExpressionInferenceSite[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | InferenceContext &#124; undefined | readonly TypeParameter[] &#124; undefined | readonly TypeParameter[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | InterfaceType | BaseType &#124; undefined | BaseType &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | InterfaceType &#124; undefined | BaseType &#124; undefined | BaseType &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | JSDocParameterTag &#124; JSDocTypeTag &#124; undefined | JSDocTypeExpression &#124; undefined | JSDocTypeExpression &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | JSDocSatisfiesTag &#124; undefined | JSDocTypeExpression | JSDocTypeExpression &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | JSDocSignature &#124; SignatureDeclaration &#124; undefined | Symbol | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | JSDocTypeExpression &#124; undefined | Signature &#124; undefined | Signature &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | CommentRange[] &#124; undefined | CommentRange[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | ConditionalTypeNode &#124; IntroducesNewScopeNode &#124; undefined | ConditionalTypeNode &#124; IntroducesNewScopeNode &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | DiagnosticMessage &#124; undefined | DiagnosticMessage &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | EmitNode &#124; undefined | EmitNode &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | NodeLinks | NodeLinks &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | ResolvedModuleFull &#124; undefined | ResolvedModuleFull &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Node &#124; undefined | Symbol[] | Symbol[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ObjectLiteralExpression &#124; undefined | StringLiteral &#124; undefined | StringLiteral &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | PackageJsonInfo &#124; undefined | VersionPaths &#124; undefined | VersionPaths &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ParameterDeclaration &#124; undefined | DotDotDotToken &#124; undefined | DotDotDotToken &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ParameterDeclaration &#124; undefined | Symbol | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | ReferencedFile &#124; undefined | ReferenceFileLocation &#124; SyntheticReferenceFileLocation | ReferenceFileLocation &#124; SyntheticReferenceFileLocation &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Resolved &#124; undefined | { resolved: Resolved; isExternalLibraryImport: true; } | { resolved: Resolved; isExternalLibraryImport: true; } &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Resolved &#124; undefined | { value: { resolved: Resolved; isExternalLibraryImport: false; }; } | { value: { resolved: Resolved; isExternalLibraryImport: false; }; } &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SearchResult<Resolved> | Resolved &#124; undefined | Resolved &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Set<FileIncludeReason> &#124; undefined | FileReasonToChainCache &#124; undefined | FileReasonToChainCache &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SetAccessorDeclaration &#124; undefined | TypeNode &#124; undefined | TypeNode &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Signature &#124; undefined | Type | Type &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SourceFile | readonly StringLiteralLike[] | readonly StringLiteralLike[] | 1 |
| a BinaryExpression with a value and a value | && | SourceFile &#124; undefined | DetachedCommentInfo &#124; undefined | DetachedCommentInfo &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SourceFile &#124; undefined | Identifier &#124; undefined | Identifier &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SourceFile &#124; undefined | Symbol | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SourceFile &#124; undefined | readonly AmdDependency[] | readonly AmdDependency[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Statement[] &#124; undefined | CaseClause[] &#124; undefined | CaseClause[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol | ClassLikeDeclaration &#124; undefined | ClassLikeDeclaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol | Declaration[] &#124; undefined | Declaration[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol | Symbol | Symbol | 1 |
| a BinaryExpression with a value and a value | && | Symbol | SymbolLinks | SymbolLinks | 1 |
| a BinaryExpression with a value and a value | && | Symbol | Type &#124; undefined | Type &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | Declaration &#124; undefined | Declaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | GenericType | GenericType &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | Node &#124; undefined | Node &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | SourceFile &#124; undefined | SourceFile &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Symbol &#124; undefined | SymbolTable | SymbolTable &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | SymbolTable | Symbol[] &#124; undefined | Symbol[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | System | (name: string) => string | (name: string) => string | 1 |
| a BinaryExpression with a value and a value | && | Token<SyntaxKind.DotDotDotToken> &#124; undefined | Token<SyntaxKind.QuestionToken> &#124; undefined | Token<SyntaxKind.QuestionToken> &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | TsConfigSourceFile &#124; undefined | TsConfigSourceFile &#124; undefined | TsConfigSourceFile &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | DestructuringPattern &#124; undefined | DestructuringPattern &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | IterationTypes &#124; undefined | IterationTypes &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | ObjectType | ObjectType &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Type &#124; undefined | TypeMapper &#124; undefined | TypeMapper &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | TypeNode &#124; undefined | TypeNode &#124; undefined | TypeNode &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | TypePredicate &#124; undefined | TypePredicate &#124; undefined | TypePredicate &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | Type[] &#124; undefined | Type[] | Type[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | WideningContext &#124; undefined | WideningContext | WideningContext &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | readonly Type[] &#124; undefined | Type[] | Type[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | readonly string[] &#124; undefined | RegExp[] | RegExp[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | readonly string[] &#124; undefined | { kind: "ambient" &#124; "node_modules" &#124; "paths" &#124; "redirect" &#124; "relative" &#124; undefined; moduleSpecifiers: readonly string[]; computedWithoutCache: false; } | { kind: "ambient" &#124; "node_modules" &#124; "paths" &#124; "redirect" &#124; "relative" &#124; undefined; moduleSpecifiers: readonly string[]; computedWithoutCache: false; } &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | { base64decode?(input: string): string; } &#124; undefined | ((input: string) => string) &#124; undefined | ((input: string) => string) &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | { base64encode?(input: string): string; } &#124; undefined | ((input: string) => string) &#124; undefined | ((input: string) => string) &#124; undefined | 1 |
| a BinaryExpression with a value and a value | && | { configFilePath: string; useCaseSensitiveFileNames: boolean; } &#124; undefined | GetCanonicalFileName | GetCanonicalFileName &#124; undefined | 1 |
| a BinaryExpression with a value and a value | = | (Identifier &#124; StringLiteral)[] &#124; undefined | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | BaseType[] | BaseType[] | BaseType[] | 1 |
| a BinaryExpression with a value and a value | = | BindingElement[] | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | ChildDirectoryWatcher[] &#124; undefined | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | EmitTextWriter | EmitTextWriter | EmitTextWriter | 1 |
| a BinaryExpression with a value and a value | = | EvolvingArrayType | EvolvingArrayType | EvolvingArrayType | 1 |
| a BinaryExpression with a value and a value | = | ExtendedConfigCacheEntry &#124; undefined | ExtendedConfigCacheEntry &#124; undefined | ExtendedConfigCacheEntry &#124; undefined | 1 |
| a BinaryExpression with a value and a value | = | ExternalModuleInfo | ExternalModuleInfo | ExternalModuleInfo | 1 |
| a BinaryExpression with a value and a value | = | Identifier[] | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | IndexInfo[] &#124; undefined | IndexInfo[] | IndexInfo[] | 1 |
| a BinaryExpression with a value and a value | = | IndexType | IndexType | IndexType | 1 |
| a BinaryExpression with a value and a value | = | Map<string, Signature> | Map<string, Signature> | Map<string, Signature> | 1 |
| a BinaryExpression with a value and a value | = | Map<string, Type> | Map<string, Type> | Map<string, Type> | 1 |
| a BinaryExpression with a value and a value | = | Partial<Levels> &#124; undefined | {} | {} | 1 |
| a BinaryExpression with a value and a value | = | RegExpExecArray &#124; null | RegExpExecArray &#124; null | RegExpExecArray &#124; null | 1 |
| a BinaryExpression with a value and a value | = | ResolvedProjectReference[] &#124; undefined | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | Set<Path> &#124; undefined | Set<Path> | Set<Path> | 1 |
| a BinaryExpression with a value and a value | = | Set<number> &#124; undefined | Set<number> | Set<number> | 1 |
| a BinaryExpression with a value and a value | = | SourceMappedPosition[] &#124; undefined | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | StringMappingType &#124; undefined | StringMappingType | StringMappingType | 1 |
| a BinaryExpression with a value and a value | = | Symbol | Symbol | Symbol | 1 |
| a BinaryExpression with a value and a value | = | Symbol &#124; undefined | Symbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | = | Symbol &#124; undefined | TransientSymbol | TransientSymbol | 1 |
| a BinaryExpression with a value and a value | = | SymbolTable[] &#124; undefined | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | Symbol[] &#124; undefined | Symbol[] &#124; undefined | Symbol[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | = | Type | InterfaceType | InterfaceType | 1 |
| a BinaryExpression with a value and a value | = | Type | TypeParameter | TypeParameter | 1 |
| a BinaryExpression with a value and a value | = | Type | UniqueESSymbolType | UniqueESSymbolType | 1 |
| a BinaryExpression with a value and a value | = | TypeChecker | TypeChecker | TypeChecker | 1 |
| a BinaryExpression with a value and a value | = | TypeParameter | TypeParameter | TypeParameter | 1 |
| a BinaryExpression with a value and a value | = | TypeReference | TypeReference | TypeReference | 1 |
| a BinaryExpression with a value and a value | = | Type[] &#124; undefined | never[] | never[] | 1 |
| a BinaryExpression with a value and a value | = | UnionType | UnionType | UnionType | 1 |
| a BinaryExpression with a value and a value | = | VariableDeclaration &#124; DestructuringAssignment | DestructuringAssignment | DestructuringAssignment | 1 |
| a BinaryExpression with a value and a value | = | WatchOptions &#124; undefined | {} | {} | 1 |
| a BinaryExpression with a value and a value | = | readonly number[] | number[] | number[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | (Identifier &#124; StringLiteral)[] &#124; undefined | never[] | (Identifier &#124; StringLiteral)[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | (JSDoc &#124; JSDocTag)[] &#124; undefined | never[] | (JSDoc &#124; JSDocTag)[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | BaseType[] | never[] | BaseType[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined | BindingName | BigIntLiteral &#124; ComputedPropertyName &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; BindingName | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | CommentRange[] &#124; undefined | never[] | CommentRange[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | CompilerOptions &#124; undefined | CompilerOptions | CompilerOptions | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | CompilerOptions &#124; undefined | {} | CompilerOptions | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | ConstructorDeclaration &#124; undefined | ClassStaticBlockDeclaration &#124; undefined | ClassStaticBlockDeclaration &#124; ConstructorDeclaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Declaration &#124; undefined | Declaration &#124; undefined | Declaration &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | DeclarationName &#124; undefined | CallSignatureDeclaration &#124; ConstructSignatureDeclaration &#124; FunctionDeclaration &#124; IndexSignatureDeclaration &#124; MethodDeclaration &#124; MethodSignature | CallSignatureDeclaration &#124; ConstructSignatureDeclaration &#124; FunctionDeclaration &#124; IndexSignatureDeclaration &#124; MethodDeclaration &#124; MethodSignature &#124; DeclarationName | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | DeclarationName &#124; undefined | DeclarationName &#124; undefined | DeclarationName &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Declaration[] &#124; undefined | never[] | Declaration[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | DiagnosticMessage &#124; undefined | DiagnosticMessage | DiagnosticMessage | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | DiagnosticMessage &#124; undefined | DiagnosticMessage &#124; undefined | DiagnosticMessage &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | EntityName &#124; undefined | EntityName &#124; undefined | EntityName &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Expression &#124; undefined | ArrowFunction &#124; undefined | Expression &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Expression &#124; undefined | Declaration | Declaration &#124; Expression | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Expression &#124; undefined | Expression | Expression | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Expression &#124; undefined | YieldExpression | Expression | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | FlowLabel &#124; undefined | FlowLabel | FlowLabel | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Identifier &#124; undefined | Identifier &#124; PrivateIdentifier &#124; undefined | Identifier &#124; PrivateIdentifier &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | IndexInfo &#124; undefined | IndexInfo &#124; undefined | IndexInfo &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | IndexInfo[] &#124; undefined | never[] | IndexInfo[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | JSDocCallbackTag &#124; JSDocEnumTag &#124; JSDocTypedefTag &#124; SignatureDeclaration &#124; undefined | JSDoc &#124; JSDocTypeLiteral | JSDoc &#124; JSDocCallbackTag &#124; JSDocEnumTag &#124; JSDocTypeLiteral &#124; JSDocTypedefTag &#124; SignatureDeclaration | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | LiteralType &#124; undefined | LiteralType | LiteralType | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Map<string, string> &#124; undefined | Map<string, string> &#124; undefined | Map<string, string> &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Node | Expression | Node | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Node | Node | Node | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Node &#124; undefined | BinaryOperatorToken | Node | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Node &#124; undefined | Identifier | Node | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | NodeArray<Expression> &#124; undefined | never[] | never[] &#124; NodeArray<Expression> | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | NodeArray<ExpressionWithTypeArguments> &#124; undefined | never[] | never[] &#124; NodeArray<ExpressionWithTypeArguments> | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | NodeArray<TypeParameterDeclaration> &#124; undefined | TypeNode &#124; undefined | NodeArray<TypeParameterDeclaration> &#124; TypeNode &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | ObjectType &#124; undefined | Type | Type | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | PotentiallyUnusedIdentifier[] &#124; undefined | never[] | PotentiallyUnusedIdentifier[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | PropertyAccessExpression &#124; undefined | BindingElement &#124; ImportSpecifier | BindingElement &#124; ImportSpecifier &#124; PropertyAccessExpression | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | PropertyName &#124; undefined | SignatureDeclaration | PropertyName &#124; SignatureDeclaration | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | RequireOrImportCall[] &#124; undefined | never[] | RequireOrImportCall[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | SearchResult<Resolved> | SearchResult<Resolved> | SearchResult<Resolved> | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Set<Path> &#124; undefined | Set<Path> &#124; undefined | Set<Path> &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Signature &#124; undefined | Signature | Signature | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | SourceFileLike &#124; undefined | SourceFile | SourceFile &#124; SourceFileLike | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | SourceFile[] &#124; undefined | never[] | SourceFile[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | StringLiteralLike[] &#124; undefined | never[] | StringLiteralLike[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol &#124; undefined | IndexInfo &#124; undefined | IndexInfo &#124; Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol &#124; undefined | TransientSymbol &#124; undefined | Symbol &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol &#124; undefined | Type &#124; undefined | Symbol &#124; Type &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol[] &#124; undefined | Symbol[] | Symbol[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Symbol[] &#124; undefined | never[] | Symbol[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | System &#124; undefined | System | System | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type | Type | Type | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | FreshableIntrinsicType | Type | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | ResolvedType | Type | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type &#124; undefined | undefined | Type &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | TypeNode &#124; undefined | FunctionLikeDeclaration | TypeNode &#124; FunctionLikeDeclaration | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | TypeNode &#124; undefined | TypeNode | TypeNode | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | TypeParameter[] &#124; undefined | TypeParameter[] &#124; undefined | TypeParameter[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | TypePredicate &#124; undefined | TypePredicate | TypePredicate | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | TypePredicate &#124; undefined | TypePredicate &#124; undefined | TypePredicate &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | Type[] | readonly Type[] | Type[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | WideningContext &#124; undefined | WideningContext | WideningContext | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | [string, string] &#124; undefined | never[] | never[] &#124; [string, string] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | readonly Declaration[] &#124; undefined | never[] | readonly Declaration[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | readonly FileReference[] &#124; undefined | never[] | readonly FileReference[] | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | string[] &#124; undefined | string[] &#124; undefined | string[] &#124; undefined | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | { 250: number; 500: number; 2000: number; } &#124; undefined | { 2000: number; 250: number; 500: number; } | { 250: number; 500: number; 2000: number; } | 1 |
| a BinaryExpression with a value and a value | &#124;&#124; | { 250: number; 500: number; 2000: number; } &#124; undefined | { 250: number; 500: number; 2000: number; } | { 250: number; 500: number; 2000: number; } | 1 |
| a BinaryExpression with a value and a boolean | && | Node | boolean | boolean | 29 |
| a BinaryExpression with a value and a boolean | && | Type &#124; undefined | boolean | boolean &#124; undefined | 29 |
| a BinaryExpression with a value and a boolean | && | Node &#124; undefined | boolean | boolean &#124; undefined | 26 |
| a BinaryExpression with a value and a boolean | && | Declaration &#124; undefined | boolean | boolean &#124; undefined | 25 |
| a BinaryExpression with a value and a boolean | && | Expression &#124; undefined | boolean | boolean &#124; undefined | 19 |
| a BinaryExpression with a value and a boolean | && | Symbol &#124; undefined | boolean | boolean &#124; undefined | 17 |
| a BinaryExpression with a value and a boolean | && | PropertyName &#124; undefined | boolean | boolean &#124; undefined | 9 |
| a BinaryExpression with a value and a boolean | && | JSDocSignature &#124; SignatureDeclaration &#124; undefined | boolean | boolean &#124; undefined | 8 |
| a BinaryExpression with a value and a boolean | && | Symbol | boolean | boolean | 8 |
| a BinaryExpression with a value and a boolean | && | TypeNode &#124; undefined | boolean | boolean &#124; undefined | 8 |
| a BinaryExpression with a value and a boolean | && | SourceFile &#124; undefined | boolean | boolean &#124; undefined | 7 |
| a BinaryExpression with a value and a boolean | && | Declaration[] &#124; undefined | boolean | boolean &#124; undefined | 6 |
| a BinaryExpression with a value and a boolean | && | Signature &#124; undefined | boolean | boolean &#124; undefined | 6 |
| a BinaryExpression with a value and a boolean | && | DeclarationName &#124; undefined | boolean | boolean &#124; undefined | 5 |
| a BinaryExpression with a value and a boolean | && | ArrayBindingPattern &#124; BigIntLiteral &#124; ComputedPropertyName &#124; ElementAccessExpression &#124; ... 8 more ... &#124; undefined | boolean | boolean &#124; undefined | 4 |
| a BinaryExpression with a value and a boolean | && | ExpressionWithTypeArguments &#124; undefined | boolean | boolean &#124; undefined | 4 |
| a BinaryExpression with a value and a boolean | && | Identifier &#124; undefined | boolean | boolean &#124; undefined | 4 |
| a BinaryExpression with a value and a boolean | && | IndexSignatureDeclaration &#124; undefined | boolean | boolean &#124; undefined | 4 |
| a BinaryExpression with a value and a boolean | && | Set<__String> | boolean | boolean | 4 |
| a BinaryExpression with a value and a boolean | && | ClassStaticBlockDeclaration &#124; SignatureDeclaration &#124; undefined | boolean | boolean &#124; undefined | 3 |
| a BinaryExpression with a value and a boolean | && | ForInitializer &#124; undefined | boolean | boolean &#124; undefined | 3 |
| a BinaryExpression with a value and a boolean | && | NodeArray<TypeNode> &#124; undefined | boolean | boolean &#124; undefined | 3 |
| a BinaryExpression with a value and a boolean | && | SignatureDeclaration &#124; undefined | boolean | boolean &#124; undefined | 3 |
| a BinaryExpression with a value and a boolean | && | TypePredicate &#124; undefined | boolean | boolean &#124; undefined | 3 |
| a BinaryExpression with a value and a boolean | && | readonly TypeParameter[] &#124; undefined | boolean | boolean &#124; undefined | 3 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | QuestionDotToken &#124; undefined | boolean | boolean &#124; QuestionDotToken | 3 |
| a BinaryExpression with a value and a boolean | && | AccessorDeclaration &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; StringLiteral &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | BindingName | boolean | boolean | 2 |
| a BinaryExpression with a value and a boolean | && | BindingOrAssignmentElementTarget &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | CanonicalDiagnostic &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | ConciseBody &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | ConvertedLoopState &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | ElementAccessExpression &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | FunctionLikeDeclaration &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | JSDocParameterTag &#124; ParameterDeclaration &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | JSDocTypeExpression &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | NamedExportBindings &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | NodeArray<Node> &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | PackageJsonInfo &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | ResolvedModuleFull &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | SourceFile | boolean | boolean | 2 |
| a BinaryExpression with a value and a boolean | && | TypePredicate | boolean | boolean | 2 |
| a BinaryExpression with a value and a boolean | && | Type[] &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | && | WatchOptions &#124; undefined | boolean | boolean &#124; undefined | 2 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | Expression &#124; undefined | boolean | boolean &#124; Expression | 2 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | Symbol &#124; undefined | boolean | boolean &#124; Symbol | 2 |
| a BinaryExpression with a value and a boolean | && | (() => void) &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ((name: string) => boolean) &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | AbstractKeyword &#124; AccessorKeyword &#124; AsyncKeyword &#124; ConstKeyword &#124; DeclareKeyword &#124; Decorator &#124; ... 10 more ... &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ActiveLabel &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | AnyImportOrJsDocImport &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ArrayBindingPattern &#124; ElementAccessExpression &#124; IndexedAccessTypeNode &#124; ObjectBindingPattern &#124; SyntheticExpression &#124; PropertyName &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | AssignmentExpression<EqualsToken> &#124; CallExpression &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | BinaryExpression &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | BindingElement &#124; ParameterDeclaration &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | BindingPattern &#124; PropertyName &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | BuilderState &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | Bundle &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ClassImplementingOrExtendingExpressionWithTypeArguments &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ClassLikeDeclaration | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | ClassLikeDeclaration &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | CompilerOptions &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ConciseBody | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | Declaration | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | DiagnosticCollection | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | DiagnosticMessageChain &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | DotDotDotToken &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | EntityName &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | FileSystemEntries &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | FlowNode | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | FlowNode &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | FunctionDeclaration &#124; FunctionExpression &#124; SuperContainer &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | GenericType | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | HasLocals &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | HeritageClause &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | HostDirectoryWatcher &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ImportAttributes &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ImportCall &#124; ImportDeclaration &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ImportClause &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ImportEqualsDeclaration &#124; NamespaceExport &#124; NamespaceImport &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | IndexInfo &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | InferenceInfo &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | JSDocMemberName &#124; EntityName &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | JSDocTypeExpression &#124; JSDocTypeLiteral &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | JSDocTypeTag &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | LoggingHost &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | MapLike<string[]> &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ModifierLike &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ModuleBody &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ModuleBody &#124; undefined | false | false &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ModuleDeclaration | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | NodeArray<ExpressionWithTypeArguments> | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | NodeArray<JsxAttributeLike> | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | NodeArray<TypeParameterDeclaration> &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | NodeBuilderContext &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ParsedTsconfig &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | PerformanceTime &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | ReferenceFileLocation &#124; SyntheticReferenceFileLocation &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | SetAccessorDeclaration | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | SignatureDeclaration | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | SuperContainerOrFunctions &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | SymbolTable | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | TextSpan &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | Token<SyntaxKind.DotDotDotToken> &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | Type | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | TypeMapper &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | TypeParameter &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | VariableDeclaration &#124; BindingName &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | VariableDeclaration &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | readonly CommentRange[] &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | readonly Expression[] | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | readonly Expression[] &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | readonly FileReference[] | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | readonly StringLiteralLike[] | boolean | boolean | 1 |
| a BinaryExpression with a value and a boolean | && | readonly Type[] &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | readonly string[] &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | && | string[] &#124; undefined | boolean | boolean &#124; undefined | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | ((moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference &#124; undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... &#124; undefined) => ...) &#124; undefined | boolean | boolean &#124; ((moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference &#124; undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... &#124; undefined) => ...) | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | Identifier &#124; undefined | boolean | boolean &#124; Identifier | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | IndexInfo &#124; undefined | boolean | boolean &#124; IndexInfo | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | IterationTypes &#124; undefined | boolean | boolean &#124; IterationTypes | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | NodeArray<TypeNode> &#124; undefined | boolean | boolean &#124; NodeArray<TypeNode> | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | ParameterDeclaration &#124; undefined | boolean | boolean &#124; ParameterDeclaration | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | ResolvedRefAndOutputDts &#124; undefined | boolean | boolean &#124; ResolvedRefAndOutputDts | 1 |
| a BinaryExpression with a value and a boolean | &#124;&#124; | SourceFile &#124; undefined | false | false &#124; SourceFile | 1 |
| a BinaryExpression with a number and a number | && | number | number | number | 61 |
| a BinaryExpression with a number and a number | &#124;&#124; | number | number | number | 33 |
| a BinaryExpression with a number and a number | = | number | number | number | 11 |
| a BinaryExpression with a number and a number | = | Ternary | Ternary | Ternary | 9 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind | SyntaxKind | 7 |
| a BinaryExpression with a number and a number | &#124;&#124; | Comparison | Comparison | Comparison | 6 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.ConflictMarkerTrivia | SyntaxKind.ConflictMarkerTrivia | 5 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.EndOfFileToken | SyntaxKind.EndOfFileToken | 4 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.LessThanToken | SyntaxKind.LessThanToken | 4 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.EqualsToken | SyntaxKind.EqualsToken | 3 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.OpenBraceToken | SyntaxKind.OpenBraceToken | 3 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.AsteriskToken | SyntaxKind.AsteriskToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.AtToken | SyntaxKind.AtToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.CloseBraceToken | SyntaxKind.CloseBraceToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.CloseBracketToken | SyntaxKind.CloseBracketToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.CloseParenToken | SyntaxKind.CloseParenToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.CommaToken | SyntaxKind.CommaToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.DotToken | SyntaxKind.DotToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.GreaterThanToken | SyntaxKind.GreaterThanToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.HashToken | SyntaxKind.HashToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.Identifier &#124; KeywordSyntaxKind | SyntaxKind.Identifier &#124; KeywordSyntaxKind | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.NewLineTrivia | SyntaxKind.NewLineTrivia | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.OpenBracketToken | SyntaxKind.OpenBracketToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.OpenParenToken | SyntaxKind.OpenParenToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.QuestionToken | SyntaxKind.QuestionToken | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.StringLiteral | SyntaxKind.StringLiteral | 2 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.WhitespaceTrivia | SyntaxKind.WhitespaceTrivia | 2 |
| a BinaryExpression with a number and a number | && | ElementFlags | ElementFlags.Optional | ElementFlags.Optional | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | JsxTokenSyntaxKind | JsxTokenSyntaxKind | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | KeywordSyntaxKind | KeywordSyntaxKind | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.AmpersandToken | SyntaxKind.AmpersandToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.BacktickToken | SyntaxKind.BacktickToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.BarToken | SyntaxKind.BarToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.CaretToken | SyntaxKind.CaretToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.ColonToken | SyntaxKind.ColonToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.ExclamationToken | SyntaxKind.ExclamationToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.GreaterThanEqualsToken | SyntaxKind.GreaterThanEqualsToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.GreaterThanGreaterThanToken | SyntaxKind.GreaterThanGreaterThanToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.Identifier | SyntaxKind.Identifier | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.JSDocCommentTextToken | SyntaxKind.JSDocCommentTextToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.LessThanSlashToken | SyntaxKind.LessThanSlashToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.MinusToken | SyntaxKind.MinusToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.MultiLineCommentTrivia | SyntaxKind.MultiLineCommentTrivia | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.NumericLiteral | SyntaxKind.NumericLiteral | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.PercentToken | SyntaxKind.PercentToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.PlusToken | SyntaxKind.PlusToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.SemicolonToken | SyntaxKind.SemicolonToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.ShebangTrivia | SyntaxKind.ShebangTrivia | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.SingleLineCommentTrivia | SyntaxKind.SingleLineCommentTrivia | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.SlashToken | SyntaxKind.SlashToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.TildeToken | SyntaxKind.TildeToken | 1 |
| a BinaryExpression with a number and a number | = | SyntaxKind | SyntaxKind.Unknown | SyntaxKind.Unknown | 1 |
| a BinaryExpression with a number and a number | = | number | 64 | 64 | 1 |
| a BinaryExpression with a number and a number | &#124;&#124; | CheckMode | CheckMode.Normal | CheckMode | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Expression &#124; undefined | false &#124; Expression &#124; undefined | 19 |
| a BinaryExpression with a boolean and a value | && | boolean | Identifier &#124; undefined | false &#124; Identifier &#124; undefined | 10 |
| a BinaryExpression with a boolean and a value | && | boolean | Symbol &#124; undefined | false &#124; Symbol &#124; undefined | 8 |
| a BinaryExpression with a boolean and a value | && | boolean | NodeArray<TypeNode> &#124; undefined | false &#124; NodeArray<TypeNode> &#124; undefined | 7 |
| a BinaryExpression with a boolean and a value | && | boolean | Type &#124; undefined | false &#124; Type &#124; undefined | 7 |
| a BinaryExpression with a boolean and a value | && | boolean | TypeNode &#124; undefined | false &#124; TypeNode &#124; undefined | 7 |
| a BinaryExpression with a boolean and a value | && | boolean | ClassElement &#124; undefined | false &#124; ClassElement &#124; undefined | 6 |
| a BinaryExpression with a boolean and a value | && | boolean | SymbolTable &#124; undefined | false &#124; SymbolTable &#124; undefined | 4 |
| a BinaryExpression with a boolean and a value | && | boolean | ConciseBody &#124; undefined | false &#124; ConciseBody &#124; undefined | 3 |
| a BinaryExpression with a boolean and a value | && | boolean | Node | false &#124; Node | 3 |
| a BinaryExpression with a boolean and a value | &#124;&#124; | boolean | Expression &#124; undefined | true &#124; Expression &#124; undefined | 3 |
| a BinaryExpression with a boolean and a value | && | boolean | AsteriskToken &#124; undefined | false &#124; AsteriskToken &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | AwaitKeyword &#124; undefined | false &#124; AwaitKeyword &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | BigIntLiteral &#124; Identifier &#124; JsxNamespacedName &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike &#124; undefined | false &#124; BigIntLiteral &#124; Identifier &#124; JsxNamespacedName &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteralLike &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | Declaration[] &#124; undefined | false &#124; Declaration[] &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | DotDotDotToken &#124; undefined | false &#124; DotDotDotToken &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | FlowNode &#124; undefined | false &#124; FlowNode &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | JSDocArray &#124; undefined | false &#124; JSDocArray &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | Node &#124; undefined | false &#124; Node &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | NodeArray<ModifierLike> &#124; undefined | false &#124; NodeArray<ModifierLike> &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | PackageJsonInfo &#124; undefined | false &#124; PackageJsonInfo &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | PrivateIdentifierAccessorInfo &#124; PrivateIdentifierInstanceFieldInfo &#124; PrivateIdentifierMethodInfo &#124; PrivateIdentifierStaticFieldInfo &#124; undefined | false &#124; PrivateIdentifierAccessorInfo &#124; PrivateIdentifierInstanceFieldInfo &#124; PrivateIdentifierMethodInfo &#124; PrivateIdentifierStaticFieldInfo &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | QuestionToken &#124; undefined | false &#124; QuestionToken &#124; undefined | 2 |
| a BinaryExpression with a boolean and a value | && | boolean | (() => MapLike<string> &#124; undefined) &#124; undefined | false &#124; (() => MapLike<string> &#124; undefined) &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ((left: Type, right: Type) => boolean) &#124; undefined | false &#124; ((left: Type, right: Type) => boolean) &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ((path: string) => string) &#124; undefined | false &#124; ((path: string) => string) &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ((source: Type, target: Type) => void) &#124; undefined | false &#124; ((source: Type, target: Type) => void) &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | AccessorDeclaration &#124; undefined | false &#124; AccessorDeclaration &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined | false &#124; BigIntLiteral &#124; ComputedPropertyName &#124; Identifier &#124; NoSubstitutionTemplateLiteral &#124; NumericLiteral &#124; PrivateIdentifier &#124; StringLiteral &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | BindingName | false &#124; BindingName | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Block &#124; undefined | false &#124; Block &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ClassInfo &#124; undefined | false &#124; ClassInfo &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ClassLexicalEnvironment &#124; undefined | false &#124; ClassLexicalEnvironment &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ClassLikeDeclaration &#124; undefined | false &#124; ClassLikeDeclaration &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Declaration &#124; undefined | false &#124; Declaration &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | DiagnosticMessage &#124; undefined | false &#124; DiagnosticMessage &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | DiagnosticMessageChain &#124; undefined | false &#124; DiagnosticMessageChain &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | EntityName &#124; undefined | false &#124; EntityName &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ExclamationToken &#124; undefined | false &#124; ExclamationToken &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ForInitializer &#124; undefined | false &#124; ForInitializer &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | FormatDiagnosticsHost &#124; undefined | false &#124; FormatDiagnosticsHost &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Identifier &#124; JSDocNamespaceDeclaration &#124; undefined | false &#124; Identifier &#124; JSDocNamespaceDeclaration &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ImportAttributes &#124; undefined | false &#124; ImportAttributes &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ImportClause &#124; undefined | false &#124; ImportClause &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | JsonSourceFile | false &#124; JsonSourceFile | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Map<string, boolean> | false &#124; Map<string, boolean> | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | ModuleBody &#124; undefined | false &#124; ModuleBody &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Mutable<LiteralExpression>[] &#124; undefined | false &#124; Mutable<LiteralExpression>[] &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | NamedExportBindings &#124; undefined | false &#124; NamedExportBindings &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | NamedTupleMember &#124; ParameterDeclaration &#124; undefined | false &#124; NamedTupleMember &#124; ParameterDeclaration &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | NodeArray<ExpressionWithTypeArguments> &#124; undefined | false &#124; NodeArray<ExpressionWithTypeArguments> &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | NodeArray<HeritageClause> &#124; undefined | false &#124; NodeArray<HeritageClause> &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | NodeArray<JSDocTag> &#124; undefined | false &#124; NodeArray<JSDocTag> &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | NodeArray<TypeParameterDeclaration> &#124; undefined | false &#124; NodeArray<TypeParameterDeclaration> &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | PackageId &#124; undefined | false &#124; PackageId &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Program &#124; undefined | false &#124; Program &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | PropertyName &#124; undefined | false &#124; PropertyName &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | QualifiedName | false &#124; QualifiedName | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Resolved &#124; undefined | false &#124; Resolved &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Signature &#124; undefined | false &#124; Signature &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | SourceFile &#124; undefined | false &#124; SourceFile &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Symbol | false &#124; Symbol | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Symbol[] &#124; undefined | false &#124; Symbol[] &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Token<SyntaxKind.DotDotDotToken> &#124; undefined | false &#124; Token<SyntaxKind.DotDotDotToken> &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Type | false &#124; Type | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | TypeMapper &#124; undefined | false &#124; TypeMapper &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | Type[] &#124; undefined | false &#124; Type[] &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | VariableDeclaration &#124; undefined | false &#124; VariableDeclaration &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | readonly Type[] &#124; undefined | false &#124; readonly Type[] &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | && | boolean | undefined | false &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | &#124;&#124; | boolean | Node &#124; undefined | true &#124; Node &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | &#124;&#124; | boolean | Partial<Levels> &#124; undefined | true &#124; Partial<Levels> &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | &#124;&#124; | boolean | SourceFile &#124; undefined | true &#124; SourceFile &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | &#124;&#124; | boolean | Symbol &#124; undefined | true &#124; Symbol &#124; undefined | 1 |
| a BinaryExpression with a boolean and a value | &#124;&#124; | boolean | undefined | true &#124; undefined | 1 |
| a BinaryExpression with a number and a boolean | && | number | boolean | 0 &#124; boolean | 118 |
| a BinaryExpression with a number and a boolean | &#124;&#124; | number | boolean | number &#124; boolean | 28 |
| a BinaryExpression with a number and a boolean | && | Ternary | boolean | boolean &#124; Ternary.False | 2 |
| a BinaryExpression with a number and a boolean | && | ScriptTarget | boolean | boolean &#124; ScriptTarget.ES3 | 1 |
| a BinaryExpression with a number and a boolean | && | TempFlags | boolean | boolean &#124; TempFlags.Auto | 1 |
