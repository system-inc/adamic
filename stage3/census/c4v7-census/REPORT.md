# c4v-7 census

sys.ts:304:13: cleared in CommonJSBundle. scanner.ts:2934:41: cleared in CommonJSBundle. EcmaScript pending. **12 new refusal-scan findings.**

Compiler branch `compiler/constraint-body-rigid-3`, c4v-7, task #p0fy047. Head: `1d48ecb5a5321717daf901579e49a23291995589`; main and confirmed merge-base: `5c8338c4232c37afb304ce6865458b67655c590a`. Measurement and reporting only; no adaptation or compiler changes.

## Part 2 — paired refusal scan

All 79 adapted compiler files at each commit, CommonJSBundle. Scan-only sees refusal-scan findings only; all 4,397 main and 4,024 head findings have phase `refusal_scan`. It does not lower bodies. The full first-stop census is needed to name lowering-phase stops. All controls and per-file output validations passed.

The adapted source identity check is in [source-identity.json](source-identity.json). Finding identity is `(kind, file, line, column, reason)`. Paths are normalized to `src/compiler/`; identical lines map by identity. “Changed” means the same key has different normalized diagnostic text; reason/kind changes count as gone plus new.

| Before | After | Gone | New | Changed |
| ---: | ---: | ---: | ---: | ---: |
| 4397 | 4024 | 385 | 12 | 0 |

| Reason | Gone | New | Changed |
| --- | ---: | ---: | ---: |
| a function taking SourceFile seen as one taking SourceFile \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| a value of type (JsxElement \| JsxExpression \| JsxFragment \| JsxSelfClosingElement \| JsxText)[] seen as Node[], which can write Node where JsxElement \| JsxExpression \| JsxFragment \| JsxSelfClosingElement \| JsxText is read | 1 | 0 | 0 |
| a value of type (TextRange & { name: string; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { name: string; } is read | 1 | 0 | 0 |
| a value of type (TextRange & { value: number; })[] \| undefined seen as {}[] \| undefined, which can write {} where TextRange & { value: number; } is read | 1 | 0 | 0 |
| a value of type (WatchedFileWithUnchangedPolls \| undefined)[] seen as (WatchedFileWithIsClosed \| undefined)[], which can write WatchedFileWithIsClosed \| undefined where WatchedFileWithUnchangedPolls \| undefined is read | 1 | 0 | 0 |
| a value of type (chain: ReusableDiagnosticMessageChain) =&gt; RepopulateDiagnosticChainInfo seen as (chain: DiagnosticMessageChain \| ReusableDiagnosticMessageChain) =&gt; RepopulateDiagnosticChainInfo \| undefined, which can write ReusableDiagnosticMessageChain[] \| undefined where DiagnosticMessageChain[] \| undefined is read | 1 | 0 | 0 |
| a value of type (readonly IncrementalBuildInfoFileId[])[] \| undefined seen as {}[] \| undefined, which can write {} where readonly IncrementalBuildInfoFileId[] is read | 1 | 0 | 0 |
| a value of type (resolution: ResolvedModuleWithFailedLookupLocations) =&gt; ResolvedModuleFull \| undefined seen as GetResolutionWithResolvedFileName&lt;CachedResolvedModuleWithFailedLookupLocations, ResolutionWithResolvedFileName&gt;, which can write string \| undefined where string is read | 3 | 0 | 0 |
| a value of type (sourceFile: SourceFile) =&gt; readonly DiagnosticWithLocation[] seen as (sourceFile: SourceFile, cancellationToken: CancellationToken \| undefined) =&gt; readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 0 | 0 |
| a value of type (sourceFile: SourceFile, cancellationToken: CancellationToken \| undefined) =&gt; readonly DiagnosticWithLocation[] seen as (sourceFile: SourceFile, cancellationToken: CancellationToken \| undefined) =&gt; readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 0 | 0 |
| a value of type AssertClause seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type AssertEntry seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type BindingElement[] \| undefined seen as {}[] \| undefined, which can write {} where BindingElement is read | 1 | 0 | 0 |
| a value of type BindingOrAssignmentElement[] \| undefined seen as {}[] \| undefined, which can write {} where BindingOrAssignmentElement is read | 3 | 0 | 0 |
| a value of type Block seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type BreakStatement seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type CallExpression[] \| undefined seen as {}[] \| undefined, which can write {} where CallExpression is read | 2 | 0 | 0 |
| a value of type CallSignatureDeclaration seen as Mutable&lt;SignatureDeclarationBase&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type CaseBlock seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ClassElement[] \| undefined seen as {}[] \| undefined, which can write {} where ClassElement is read | 5 | 0 | 0 |
| a value of type CommentDirective[] \| undefined seen as {}[] \| undefined, which can write {} where CommentDirective is read | 3 | 0 | 0 |
| a value of type ConstructSignatureDeclaration seen as Mutable&lt;SignatureDeclarationBase&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ConstructorTypeNode seen as Mutable&lt;SignatureDeclarationBase&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ContinueStatement seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type Decorator[] \| undefined seen as {}[] \| undefined, which can write {} where Decorator is read | 3 | 0 | 0 |
| a value of type DiagnosticRelatedInformation[] \| undefined seen as {}[] \| undefined, which can write {} where DiagnosticRelatedInformation is read | 1 | 0 | 0 |
| a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 0 | 2 | 0 |
| a value of type DiagnosticWithLocation[] \| undefined seen as {}[] \| undefined, which can write {} where DiagnosticWithLocation is read | 1 | 0 | 0 |
| a value of type ElementWithComputedPropertyName[] \| undefined seen as {}[] \| undefined, which can write {} where ElementWithComputedPropertyName is read | 1 | 0 | 0 |
| a value of type EmitHelper[] \| undefined seen as {}[] \| undefined, which can write {} where EmitHelper is read | 2 | 0 | 0 |
| a value of type ExpressionStatement seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type Expression[] seen as {}[], which can write {} where Expression is read | 2 | 0 | 0 |
| a value of type Expression[] \| undefined seen as {}[] \| undefined, which can write {} where Expression is read | 10 | 0 | 0 |
| a value of type Identifier[] seen as {}[], which can write {} where Identifier is read | 1 | 0 | 0 |
| a value of type Identifier[] \| undefined seen as {}[] \| undefined, which can write {} where Identifier is read | 1 | 0 | 0 |
| a value of type ImportAttribute seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ImportAttributes seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ImportClause seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ImportDeclaration seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ImportTypeAssertionContainer seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ImportTypeNode seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type IncrementalBuildInfoDiagnosticOfFile[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoDiagnosticOfFile is read | 1 | 0 | 0 |
| a value of type IncrementalBuildInfoDiagnostic[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoDiagnostic is read | 2 | 0 | 0 |
| a value of type IncrementalBuildInfoEmitSignature[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoEmitSignature is read | 1 | 0 | 0 |
| a value of type IncrementalBuildInfoFileId[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoFileId is read | 1 | 0 | 0 |
| a value of type IncrementalBuildInfoFilePendingEmit[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoFilePendingEmit is read | 1 | 0 | 0 |
| a value of type IncrementalBuildInfoResolvedRoot[] \| undefined seen as {}[] \| undefined, which can write {} where IncrementalBuildInfoResolvedRoot is read | 1 | 0 | 0 |
| a value of type IndexInfo[] \| undefined seen as {}[] \| undefined, which can write {} where IndexInfo is read | 3 | 0 | 0 |
| a value of type IndexSignatureDeclaration seen as Mutable&lt;SignatureDeclarationBase&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type IterationTypes[] \| undefined seen as {}[] \| undefined, which can write {} where IterationTypes is read | 1 | 0 | 0 |
| a value of type JSDocAugmentsTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocCallbackTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocComment[] seen as Node[], which can write Node where JSDocComment is read | 3 | 0 | 0 |
| a value of type JSDocFunctionType seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocImplementsTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocImportTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocLink seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocLinkCode seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocLinkPlain seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocNameReference seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocOverloadTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocParameterTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocParameterTag[] seen as Node[], which can write Node where JSDocParameterTag is read | 1 | 0 | 0 |
| a value of type JSDocParameterTag[] \| undefined seen as {}[] \| undefined, which can write {} where JSDocParameterTag is read | 1 | 0 | 0 |
| a value of type JSDocPropertyLikeTag[] \| undefined seen as {}[] \| undefined, which can write {} where JSDocPropertyLikeTag is read | 1 | 0 | 0 |
| a value of type JSDocPropertyTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocPropertyTag[] \| undefined seen as {}[] \| undefined, which can write {} where JSDocPropertyTag is read | 1 | 0 | 0 |
| a value of type JSDocSeeTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocSignature seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocTag[] seen as Node[], which can write Node where JSDocTag is read | 1 | 0 | 0 |
| a value of type JSDocTemplateTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocText seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocTypeExpression seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocTypeLiteral seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocTypedefTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JSDocUnknownTag seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type JsxChild[] seen as Node[], which can write Node where JsxChild is read | 1 | 0 | 0 |
| a value of type Map&lt;Path, ModeAwareCache&lt;CachedResolvedModuleWithFailedLookupLocations&gt;&gt; seen as Map&lt;string, ModeAwareCache&lt;ResolutionWithFailedLookupLocations&gt;&gt;, which can write string where Path is read | 1 | 0 | 0 |
| a value of type Map&lt;Path, ModeAwareCache&lt;CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations&gt;&gt; seen as Map&lt;string, ModeAwareCache&lt;ResolutionWithFailedLookupLocations&gt;&gt;, which can write string where Path is read | 1 | 0 | 0 |
| a value of type Map&lt;string, SingleFileWatcher&lt;FileWatcherCallback&gt;&gt; seen as Map&lt;string, SingleFileWatcher&lt;FileWatcherCallback \| FsWatchCallback&gt;&gt;, which can write SingleFileWatcher&lt;FileWatcherCallback \| FsWatchCallback&gt; where SingleFileWatcher&lt;FileWatcherCallback&gt; is read | 1 | 0 | 0 |
| a value of type Map&lt;string, SingleFileWatcher&lt;FsWatchCallback&gt;&gt; seen as Map&lt;string, SingleFileWatcher&lt;FileWatcherCallback \| FsWatchCallback&gt;&gt;, which can write SingleFileWatcher&lt;FileWatcherCallback \| FsWatchCallback&gt; where SingleFileWatcher&lt;FsWatchCallback&gt; is read | 1 | 0 | 0 |
| a value of type MappedTypeNode seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type ModifierLike[] seen as Node[], which can write Node where ModifierLike is read | 1 | 0 | 0 |
| a value of type ModifierLike[] \| undefined seen as {}[] \| undefined, which can write {} where ModifierLike is read | 4 | 0 | 0 |
| a value of type Modifier[] \| undefined seen as {}[] \| undefined, which can write {} where Modifier is read | 1 | 0 | 0 |
| a value of type ModuleExportName[] \| undefined seen as {}[] \| undefined, which can write {} where ModuleExportName is read | 4 | 0 | 0 |
| a value of type Mutable&lt;ArrayBindingPattern&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ArrayBindingPattern is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ArrayLiteralExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ArrayLiteralExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ArrayTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ArrayType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ArrowFunction&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.ArrowFunction is read | 1 | 0 | 0 |
| a value of type Mutable&lt;AsExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.AsExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;AwaitExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.AwaitExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;BinaryExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.BinaryExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;BindingElement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.BindingElement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;Bundle&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.Bundle is read | 1 | 0 | 0 |
| a value of type Mutable&lt;CallChain&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.CallExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;CallExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.CallExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;CaseClause&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.CaseClause is read | 1 | 0 | 0 |
| a value of type Mutable&lt;CatchClause&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.CatchClause is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ClassDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ClassDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ClassExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ClassExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ClassStaticBlockDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ClassStaticBlockDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;CommaListExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.CommaListExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ComputedPropertyName&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ComputedPropertyName is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ConditionalExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ConditionalExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ConditionalTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ConditionalType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ConstructorDeclaration&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write Node where ClassLikeDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;Decorator&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.Decorator is read | 1 | 0 | 0 |
| a value of type Mutable&lt;DefaultClause&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.DefaultClause is read | 1 | 0 | 0 |
| a value of type Mutable&lt;DeleteExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.DeleteExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;DoStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.DoStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ElementAccessChain&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ElementAccessExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ElementAccessExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ElementAccessExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;EnumDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.EnumDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;EnumMember&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.EnumMember is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ExportAssignment&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ExportAssignment is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ExportDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ExportDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ExportSpecifier&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ExportSpecifier is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ExpressionWithTypeArguments&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ExpressionWithTypeArguments is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ExternalModuleReference&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ExternalModuleReference is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ForInStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ForInStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ForOfStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ForOfStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ForStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ForStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;FunctionDeclaration&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.FunctionDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;FunctionExpression&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.FunctionExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;FunctionTypeNode&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write SyntaxKind.MethodSignature \| SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| ... 6 more ... \| SyntaxKind.JSDocFunctionType where SyntaxKind.FunctionType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;GetAccessorDeclaration&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write Node where InterfaceDeclaration \| ObjectLiteralExpression \| TypeLiteralNode \| ClassLikeDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;HeritageClause&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.HeritageClause is read | 1 | 0 | 0 |
| a value of type Mutable&lt;IfStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.IfStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ImportEqualsDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ImportEqualsDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ImportSpecifier&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ImportSpecifier is read | 1 | 0 | 0 |
| a value of type Mutable&lt;IndexedAccessTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.IndexedAccessType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;InferTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.InferType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;InterfaceDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.InterfaceDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JSDoc&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JSDoc is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JSDocEnumTag&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JSDocEnumTag is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JSDocMemberName&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JSDocMemberName is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxAttribute&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxAttribute is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxAttributes&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxAttributes is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxClosingElement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxClosingElement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxElement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxElement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxFragment&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxFragment is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxNamespacedName&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxNamespacedName is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxOpeningElement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxOpeningElement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxSelfClosingElement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxSelfClosingElement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxSpreadAttribute&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxSpreadAttribute is read | 1 | 0 | 0 |
| a value of type Mutable&lt;JsxText&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.JsxText is read | 1 | 0 | 0 |
| a value of type Mutable&lt;LabeledStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.LabeledStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;LiteralTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.LiteralType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;MetaProperty&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.MetaProperty is read | 1 | 0 | 0 |
| a value of type Mutable&lt;MethodDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.MethodDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;MethodSignature&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write Node where InterfaceDeclaration \| TypeLiteralNode is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ModuleBlock&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ModuleBlock is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ModuleDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ModuleDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;NamedExports&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.NamedExports is read | 1 | 0 | 0 |
| a value of type Mutable&lt;NamedTupleMember&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.NamedTupleMember is read | 1 | 0 | 0 |
| a value of type Mutable&lt;NamespaceExportDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.NamespaceExportDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;NewExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.NewExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;NonNullChain&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.NonNullExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;NonNullExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.NonNullExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ObjectBindingPattern&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ObjectBindingPattern is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ObjectLiteralExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ObjectLiteralExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;OptionalTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.OptionalType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ParameterDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.Parameter is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ParenthesizedExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ParenthesizedExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ParenthesizedTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ParenthesizedType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PartiallyEmittedExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PartiallyEmittedExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PostfixUnaryExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PostfixUnaryExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PrefixUnaryExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PrefixUnaryExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PropertyAccessChain&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PropertyAccessExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PropertyAccessExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PropertyAccessExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PropertyAssignment&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PropertyAssignment is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PropertyDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PropertyDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;PropertySignature&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.PropertySignature is read | 1 | 0 | 0 |
| a value of type Mutable&lt;QualifiedName&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.QualifiedName is read | 1 | 0 | 0 |
| a value of type Mutable&lt;RestTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.RestType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;SatisfiesExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.SatisfiesExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;SetAccessorDeclaration&gt; seen as Mutable&lt;SignatureDeclarationBase&gt;, which can write Node where InterfaceDeclaration \| ObjectLiteralExpression \| TypeLiteralNode \| ClassLikeDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ShorthandPropertyAssignment&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ShorthandPropertyAssignment is read | 1 | 0 | 0 |
| a value of type Mutable&lt;SourceFile&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.SourceFile is read | 1 | 0 | 0 |
| a value of type Mutable&lt;SpreadAssignment&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.SpreadAssignment is read | 1 | 0 | 0 |
| a value of type Mutable&lt;SpreadElement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.SpreadElement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;SyntheticReferenceExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.SyntheticReferenceExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;T&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where T["kind"] is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TaggedTemplateExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TaggedTemplateExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TemplateExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TemplateExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TemplateLiteralTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TemplateLiteralType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TemplateLiteralTypeSpan&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TemplateLiteralTypeSpan is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TemplateSpan&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TemplateSpan is read | 1 | 0 | 0 |
| a value of type Mutable&lt;ThrowStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.ThrowStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TryStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TryStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TupleTypeNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TupleType is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypeAliasDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypeAliasDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypeAssertion&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypeAssertionExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypeLiteralNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypeLiteral is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypeOfExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypeOfExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypePredicateNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypePredicate is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypeQueryNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypeQuery is read | 1 | 0 | 0 |
| a value of type Mutable&lt;TypeReferenceNode&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.TypeReference is read | 1 | 0 | 0 |
| a value of type Mutable&lt;VariableDeclaration&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.VariableDeclaration is read | 1 | 0 | 0 |
| a value of type Mutable&lt;VariableDeclarationList&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.VariableDeclarationList is read | 1 | 0 | 0 |
| a value of type Mutable&lt;VariableStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.VariableStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;VoidExpression&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.VoidExpression is read | 1 | 0 | 0 |
| a value of type Mutable&lt;WhileStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.WhileStatement is read | 1 | 0 | 0 |
| a value of type Mutable&lt;WithStatement&gt; seen as Mutable&lt;Node&gt;, which can write SyntaxKind where SyntaxKind.WithStatement is read | 1 | 0 | 0 |
| a value of type NamedImports seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type NamespaceExport seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type NamespaceImport seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type NonNullable&lt;T&gt;[] seen as Node[], which can write Node where NonNullable&lt;T&gt; is read | 1 | 0 | 0 |
| a value of type ObjectLiteralElementLike[] \| undefined seen as {}[] \| undefined, which can write {} where ObjectLiteralElementLike is read | 4 | 0 | 0 |
| a value of type ParameterDeclaration[] \| undefined seen as {}[] \| undefined, which can write {} where ParameterDeclaration is read | 1 | 0 | 0 |
| a value of type PropertyName[] \| undefined seen as {}[] \| undefined, which can write {} where PropertyName is read | 1 | 0 | 0 |
| a value of type RequireOrImportCall[] \| undefined seen as {}[] \| undefined, which can write {} where RequireOrImportCall is read | 1 | 0 | 0 |
| a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string \| undefined where string is read | 0 | 4 | 0 |
| a value of type ReturnStatement seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type Signature[] \| undefined seen as {}[] \| undefined, which can write {} where Signature is read | 2 | 0 | 0 |
| a value of type Statement[] seen as {}[], which can write {} where Statement is read | 11 | 0 | 0 |
| a value of type Statement[] \| undefined seen as {}[] \| undefined, which can write {} where Statement is read | 28 | 0 | 0 |
| a value of type StringLiteralLike[] \| undefined seen as {}[] \| undefined, which can write {} where StringLiteralLike is read | 2 | 0 | 0 |
| a value of type StringLiteralType[] seen as {}[], which can write {} where StringLiteralType is read | 1 | 0 | 0 |
| a value of type SwitchStatement seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type Symbol[] seen as {}[], which can write {} where Symbol is read | 3 | 0 | 0 |
| a value of type Symbol[] \| undefined seen as {}[] \| undefined, which can write {} where Symbol is read | 4 | 0 | 0 |
| a value of type T seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 0 | 0 |
| a value of type T[] seen as (DiagnosticMessageChain \| ReusableDiagnosticMessageChain)[] \| undefined, which can write DiagnosticMessageChain \| ReusableDiagnosticMessageChain where T is read | 3 | 0 | 0 |
| a value of type T[] seen as Node[], which can write Node where T is read | 1 | 0 | 0 |
| a value of type T[] seen as Statement[], which can write Statement where T is read | 4 | 0 | 0 |
| a value of type TemplateLiteralTypeSpan[] seen as Node[], which can write Node where TemplateLiteralTypeSpan is read | 1 | 0 | 0 |
| a value of type TemplateSpan[] seen as Node[], which can write Node where TemplateSpan is read | 1 | 0 | 0 |
| a value of type TypeNode[] seen as Node[], which can write Node where TypeNode is read | 1 | 0 | 0 |
| a value of type TypeOperatorNode seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type TypeParameterDeclaration seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type TypeParameterDeclaration[] seen as Node[], which can write Node where TypeParameterDeclaration is read | 1 | 0 | 0 |
| a value of type TypeParameter[] \| undefined seen as {}[] \| undefined, which can write {} where TypeParameter is read | 4 | 0 | 0 |
| a value of type Type[] seen as {}[], which can write {} where Type is read | 5 | 0 | 0 |
| a value of type Type[] \| undefined seen as {}[] \| undefined, which can write {} where Type is read | 24 | 0 | 0 |
| a value of type VariableDeclaration[] seen as {}[], which can write {} where VariableDeclaration is read | 2 | 0 | 0 |
| a value of type VariableDeclaration[] \| undefined seen as {}[] \| undefined, which can write {} where VariableDeclaration is read | 3 | 0 | 0 |
| a value of type YieldExpression seen as Mutable&lt;Node&gt;, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 0 | 0 |
| a value of type [Identifier, BindingOrAssignmentElement][] \| undefined seen as {}[] \| undefined, which can write {} where [Identifier, BindingOrAssignmentElement] is read | 1 | 0 | 0 |
| a value of type __String[] \| undefined seen as {}[] \| undefined, which can write {} where __String is read | 1 | 0 | 0 |
| a value of type string[] seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 1 | 0 | 0 |
| a value of type string[] seen as {}[], which can write {} where string is read | 1 | 0 | 0 |
| a value of type string[] \| undefined seen as {}[] \| undefined, which can write {} where string is read | 7 | 0 | 0 |
| a value of type { name: __String; oldSymbol: Symbol; }[] \| undefined seen as {}[] \| undefined, which can write {} where { name: __String; oldSymbol: Symbol; } is read | 1 | 0 | 0 |
| an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read | 0 | 5 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind | 10 | 0 | 0 |
| optional property constraint in TypeParameter absent from structural source Type, which can hide fields | 2 | 0 | 0 |
| optional property createProgram in IncrementalProgramOptions&lt;BuilderProgram&gt; absent from structural source IncrementalCompilationOptions, which can hide fields | 1 | 0 | 0 |
| optional property modifiers in ArrowFunction absent from structural source never, which can hide fields | 0 | 1 | 0 |

Of the scout’s 191 selected ‘seen as’ sites, 191 are present on main, **191 are gone on head** (191 removed by this head versus main), and 0 remain on head. The older scout’s sys.ts:307:13 maps to main/head sys.ts:304:13, proven by reconstructing the scout’s full sys.ts source hash from its owner pins; the other 190 sites use byte-identical source files. See [scout-site-mapping.json](scout-site-mapping.json). Exact key membership is preserved in [selected-191-pairing.json](selected-191-pairing.json).

Full counts by reason, including zero-delta reasons, are in [pairing.json](pairing.json). Complete raw streams and process statuses are under `part2/`; all gone/new/changed diagnostics are retained as JSON.

## Every new finding

New findings inside a generic body: 0. Full enclosing function chains and bound call slots are recorded below.

### 1. builder.ts:1332:30

Kind: `NotYet`. Rule: `3350 constraintRelationRefusal (no diagnostic label emitted)`. Phase: `refusal_scan`.

Reason: an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/builder.ts:1332:30: stage 0 can't lower an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read yet
```

Adapted source line:

```typescript
                            [fileId, !isString(emitSignature) && emitSignature[0] === actualSignature ? emptyArray as [] : emitSignature],
```

Enclosing function: `ArrowFunction@1315:60` at `builder.ts:1315:60`.

Enclosing function chain: `ArrowFunction@1315:60` (`builder.ts:1315:60`) → `getBuildInfo` (`builder.ts:1258:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `judgeBoundArray` at `internal/lower/constraint_relation.go:356`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "argument", "index": 2, "callee": "append", "argument": "emitSignature === undefined ?\r\n                            fileId : // There is no emit, encode as false\r\n                            // fileId, signature: emptyArray if signature only differs in dtsMap option than our own compilerOptions otherwise EmitSignature\r\n                            [fileId, !isString(emitSignature) && emitSignature[0] === actualSignature ? emptyArray as [] : emitSignature]"}`.

### 2. builder.ts:1376:26

Kind: `NotYet`. Rule: `3350 constraintRelationRefusal (no diagnostic label emitted)`. Phase: `refusal_scan`.

Reason: an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/builder.ts:1376:26: stage 0 can't lower an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read yet
```

Adapted source line:

```typescript
                        [fileId] : // Pending on Dts only
```

Enclosing function: `getBuildInfo` at `builder.ts:1258:1`.

Enclosing function chain: `getBuildInfo` (`builder.ts:1258:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `judgeBoundArray` at `internal/lower/constraint_relation.go:356`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "argument", "index": 2, "callee": "append", "argument": "pendingEmit === fullEmitForOptions ?\r\n                        fileId : // Pending full emit per options\r\n                        pendingEmit === BuilderFileEmit.Dts ?\r\n                        [fileId] : // Pending on Dts only\r\n                        [fileId, pendingEmit]"}`.

### 3. builder.ts:1455:42

Kind: `NotYet`. Rule: `3350 constraintRelationRefusal (no diagnostic label emitted)`. Phase: `refusal_scan`.

Reason: an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/builder.ts:1455:42: stage 0 can't lower an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read yet
```

Adapted source line:

```typescript
                result = append(result, [toFileId(file.resolvedPath), toFileId(path)]);
```

Enclosing function: `ArrowFunction@1452:31` at `builder.ts:1452:31`.

Enclosing function chain: `ArrowFunction@1452:31` (`builder.ts:1452:31`) → `toResolvedRoot` (`builder.ts:1450:5`) → `getBuildInfo` (`builder.ts:1258:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `judgeBoundArray` at `internal/lower/constraint_relation.go:356`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "argument", "index": 2, "callee": "append", "argument": "[toFileId(file.resolvedPath), toFileId(path)]"}`.

### 4. builder.ts:1504:21

Kind: `NotYet`. Rule: `3350 constraintRelationRefusal (no diagnostic label emitted)`. Phase: `refusal_scan`.

Reason: an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/builder.ts:1504:21: stage 0 can't lower an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read yet
```

Adapted source line:

```typescript
                    toFileId(key),
```

Enclosing function: `ArrowFunction@1497:33` at `builder.ts:1497:33`.

Enclosing function chain: `ArrowFunction@1497:33` (`builder.ts:1497:33`) → `toIncrementalBuildInfoDiagnostics` (`builder.ts:1495:5`) → `getBuildInfo` (`builder.ts:1258:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `judgeBoundArray` at `internal/lower/constraint_relation.go:356`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "argument", "index": 2, "callee": "append", "argument": "[\r\n                    toFileId(key),\r\n                    toReusableDiagnostic(value, key),\r\n                ]"}`.

### 5. builder.ts:1518:17

Kind: `NotYet`. Rule: `3350 constraintRelationRefusal (no diagnostic label emitted)`. Phase: `refusal_scan`.

Reason: an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/builder.ts:1518:17: stage 0 can't lower an element of an array literal a type argument is inferred from, whose type in its constraint {} can't be read yet
```

Adapted source line:

```typescript
                toFileId(key),
```

Enclosing function: `toIncrementalBuildInfoEmitDiagnostics` at `builder.ts:1512:5`.

Enclosing function chain: `toIncrementalBuildInfoEmitDiagnostics` (`builder.ts:1512:5`) → `getBuildInfo` (`builder.ts:1258:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `judgeBoundArray` at `internal/lower/constraint_relation.go:356`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "argument", "index": 2, "callee": "append", "argument": "[\r\n                toFileId(key),\r\n                toReusableDiagnostic(value, key),\r\n            ]"}`.

### 6. program.ts:2796:16

Kind: `Refused`. Rule: `adamic/invariant-mutable`. Phase: `refusal_scan`.

Reason: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/program.ts:2796:16: Adamic 0.1 refuses a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

Adapted source line:

```typescript
        return getDiagnosticsHelper(sourceFile, getSyntacticDiagnosticsForFile, cancellationToken);
```

Enclosing function: `getSyntacticDiagnostics` at `program.ts:2795:5`.

Enclosing function chain: `getSyntacticDiagnostics` (`program.ts:2795:5`) → `createProgram` (`program.ts:1516:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `constraintRelation -> judgeBinding -> wideningRefusal` at `internal/lower/constraint_relation.go:90`.

Binding judged at the call: `{"parameter": "T", "argument": "DiagnosticWithLocation", "constraint": "Diagnostic", "callee": "getDiagnosticsHelper", "declaration": "program.ts:2779"}`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "return", "value": "getDiagnosticsHelper(sourceFile, getSyntacticDiagnosticsForFile, cancellationToken)"}`.

### 7. program.ts:2829:16

Kind: `Refused`. Rule: `adamic/invariant-mutable`. Phase: `refusal_scan`.

Reason: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/program.ts:2829:16: Adamic 0.1 refuses a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

Adapted source line:

```typescript
        return getDiagnosticsHelper(sourceFile, getDeclarationDiagnosticsForFile, cancellationToken);
```

Enclosing function: `getDeclarationDiagnostics` at `program.ts:2828:5`.

Enclosing function chain: `getDeclarationDiagnostics` (`program.ts:2828:5`) → `createProgram` (`program.ts:1516:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `constraintRelation -> judgeBinding -> wideningRefusal` at `internal/lower/constraint_relation.go:90`.

Binding judged at the call: `{"parameter": "T", "argument": "DiagnosticWithLocation", "constraint": "Diagnostic", "callee": "getDiagnosticsHelper", "declaration": "program.ts:2779"}`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "return", "value": "getDiagnosticsHelper(sourceFile, getDeclarationDiagnosticsForFile, cancellationToken)"}`.

### 8. resolutionCache.ts:774:17

Kind: `Refused`. Rule: `adamic/invariant-mutable`. Phase: `refusal_scan`.

Reason: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/resolutionCache.ts:774:17: Adamic 0.1 refuses a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

Adapted source line:

```typescript
                stopWatchFailedLookupLocationOfResolution(
```

Enclosing function: `ArrowFunction@772:35` at `resolutionCache.ts:772:35`.

Enclosing function chain: `ArrowFunction@772:35` (`resolutionCache.ts:772:35`) → `cleanupLibResolutionWatching` (`resolutionCache.ts:771:5`) → `createResolutionCache` (`resolutionCache.ts:579:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `constraintRelation -> judgeBinding -> wideningRefusal` at `internal/lower/constraint_relation.go:90`.

Binding judged at the call: `{"parameter": "R", "argument": "ResolvedModuleFull", "constraint": "ResolutionWithResolvedFileName", "callee": "stopWatchFailedLookupLocationOfResolution", "declaration": "resolutionCache.ts:1379"}`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "argument", "index": 1, "callee": "resolvedLibraries.forEach", "argument": "(resolution, libFileName) => {\r\n            if (!newProgram?.resolvedLibReferences?.has(libFileName)) {\r\n                stopWatchFailedLookupLocationOfResolution(\r\n                    resolution,\r\n                    resolutionHost.toPath(getInferredLibraryNameResolveFrom(resolutionHost.getCompilationSettings(), getCurrentDirectory(), libFileName)),\r\n                    getResolvedModuleFromResolution,\r\n                );\r\n                resolvedLibraries.delete(libFileName);\r\n            }\r\n        }"}`.

### 9. resolutionCache.ts:1053:13

Kind: `Refused`. Rule: `adamic/invariant-mutable`. Phase: `refusal_scan`.

Reason: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/resolutionCache.ts:1053:13: Adamic 0.1 refuses a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

Adapted source line:

```typescript
            watchFailedLookupLocationsOfExternalModuleResolutions(libraryName, resolution, path, getResolvedModuleFromResolution, /*deferWatchingNonRelativeResolution*/ false);
```

Enclosing function: `resolveLibrary` at `resolutionCache.ts:1041:5`.

Enclosing function chain: `resolveLibrary` (`resolutionCache.ts:1041:5`) → `createResolutionCache` (`resolutionCache.ts:579:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `constraintRelation -> judgeBinding -> wideningRefusal` at `internal/lower/constraint_relation.go:90`.

Binding judged at the call: `{"parameter": "R", "argument": "ResolvedModuleFull", "constraint": "ResolutionWithResolvedFileName", "callee": "watchFailedLookupLocationsOfExternalModuleResolutions", "declaration": "resolutionCache.ts:1102"}`.

The finding is outside a generic body.

### 10. resolutionCache.ts:1056:17

Kind: `Refused`. Rule: `adamic/invariant-mutable`. Phase: `refusal_scan`.

Reason: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/resolutionCache.ts:1056:17: Adamic 0.1 refuses a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

Adapted source line:

```typescript
                stopWatchFailedLookupLocationOfResolution(existingResolution, path, getResolvedModuleFromResolution);
```

Enclosing function: `resolveLibrary` at `resolutionCache.ts:1041:5`.

Enclosing function chain: `resolveLibrary` (`resolutionCache.ts:1041:5`) → `createResolutionCache` (`resolutionCache.ts:579:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `constraintRelation -> judgeBinding -> wideningRefusal` at `internal/lower/constraint_relation.go:90`.

Binding judged at the call: `{"parameter": "R", "argument": "ResolvedModuleFull", "constraint": "ResolutionWithResolvedFileName", "callee": "stopWatchFailedLookupLocationOfResolution", "declaration": "resolutionCache.ts:1379"}`.

The finding is outside a generic body.

### 11. resolutionCache.ts:1471:9

Kind: `Refused`. Rule: `adamic/invariant-mutable`. Phase: `refusal_scan`.

Reason: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/resolutionCache.ts:1471:9: Adamic 0.1 refuses a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)
```

Adapted source line:

```typescript
        removeResolutionsOfFileFromCache(resolvedModuleNames, filePath, getResolvedModuleFromResolution);
```

Enclosing function: `removeResolutionsOfFile` at `resolutionCache.ts:1470:5`.

Enclosing function chain: `removeResolutionsOfFile` (`resolutionCache.ts:1470:5`) → `createResolutionCache` (`resolutionCache.ts:579:1`).

Registered rule: `3350 constraintRelationRefusal`; emitter `constraintRelation -> judgeBinding -> wideningRefusal` at `internal/lower/constraint_relation.go:90`.

Binding judged at the call: `{"parameter": "R", "argument": "ResolvedModuleFull", "constraint": "ResolutionWithResolvedFileName", "callee": "removeResolutionsOfFileFromCache", "declaration": "resolutionCache.ts:1435"}`.

The finding is outside a generic body.

### 12. transformers/ts.ts:1054:13

Kind: `Refused`. Rule: `adamic/no-optional-widening`. Phase: `refusal_scan`.

Reason: optional property modifiers in ArrowFunction absent from structural source never, which can hide fields

Text:

```text
/workspace/scratch/c4v7-census/adapted-head/src/compiler/transformers/ts.ts:1054:13: Adamic 0.1 refuses optional property modifiers in ArrowFunction absent from structural source never, which can hide fields; declare modifiers on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

Adapted source line:

```typescript
            filter(constructor.parameters, (p): p is ParameterPropertyDeclaration => isParameterPropertyDeclaration(p, constructor));
```

Enclosing function: `transformClassMembers` at `transformers/ts.ts:1048:5`.

Enclosing function chain: `transformClassMembers` (`transformers/ts.ts:1048:5`) → `transformTypeScript` (`transformers/ts.ts:235:1`).

Diagnostic emitter: `optionalWideningRefusal` at `internal/lower/optional_widening.go:380`.

The finding is outside a generic body.

Syntactic slot context: `{"kind": "initializer", "binding": "parametersWithPropertyAssignments", "value": "constructor &&\r\n            filter(constructor.parameters, (p): p is ParameterPropertyDeclaration => isParameterPropertyDeclaration(p, constructor))"}`.

## Setup and bounds

Fresh detached worktrees per measured commit were created in writable scratch. Recursive submodules initialized successfully at both commits; no credential error occurred. Upstream TypeScript HEAD and peeled v6.0.3 both equal `050880ce59e30b356b686bd3144efe24f875ebc8`. Each stage3/api install passed. The initial npm calls used the default npm cache; they were repeated successfully with an explicit scratch cache and logs. All census inputs, overlays, outputs, build products, Go caches, and subsequent npm caches use scratch. The shared checkout was not modified.

The scan wrapper was copied from `typescript/promise-discard-sites` at `2dc0593b937f8fbaf6c32f8107a899f78c223966`. That branch has no promise-discard-sites README; its REPORT.md and the latent README document the procedure. Each scan binary was built once. All 158 actual compiler-file headers have `checker_rejected: false` and zero diagnostics. The fixed legacy measurement label “measured on a checker-rejected program” does not describe the input status. The inherited raw header says `latent_mode: first-error`; this label was not rewritten, while the scratch driver explicitly calls `LatentScanOnly` and zero lowering units are checked. Scratch overlays alone replace the driver call with `LatentScanOnly`; ordinary Lower remains disabled and output guards are enabled. Each compiler-file process has a 1,500-second process-group bound and cleanup; at most three run concurrently. A non-forced checkpoint push follows every batch of at most ten paired files.


## Part 1

CommonJSBundle: all 79 roots ran twice and validated. 73 roots now stop at sys.ts:514:17 (an existing unchecked-cast refusal). The sys.ts root now stops at utilitiesPublic.ts:1280:52 (an existing unproven-predicate refusal). Five roots retain their exact baseline tuples. The lowering-phase roots are debug.ts, diagnosticInformationMap.generated.ts, and types.ts.

Every root’s exact before/after tuple and source excerpts are in [the CommonJSBundle report](../next-stops-c4v7/CommonJSBundle/REPORT.md). EcmaScript is pending.
