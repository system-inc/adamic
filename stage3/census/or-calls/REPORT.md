# or call census

Task 1w5bve2. Report only; no adaptation or compiler change.

PASS for the measured compiler predicate shape: all seven sites and every predicate position were invoked, every observed return was boolean, and value differences are zero. The complete oracle has 398 units, 106366 passing tests, one expected failure and zero skips. Across all src sites, there were 2186858 calls to returned closures; services/mapCode.ts:157 was not invoked. The module-detection rest site has non-boolean returns; the other three rest sites have none observed.

Base: `b7a8ce57d0c6aee7e03b6bd3102dff673865403a` (fetched origin/main tip). TypeScript source: `050880ce59e30b356b686bd3144efe24f875ebc8`, stock checker 6.0.3.

The completed apply contains 711 TypeScript files under src. There are 23 calls of compiler/core.ts's or: ten compiler, thirteen services, zero server and zero harness. The ten compiler locations in the request are correct for the pinned upstream source. Adaptation moves them by one or two lines, as shown below.

Adaptation 49 retains predicate overloads 1 (two predicates) and 2 (three predicates), adds nonempty-rest overload 3, and makes general-rest overload 4 return U | undefined. Upstream overload 3 corresponds to adapted overloads 3 and 4. There are seven predicate sites and three rest sites in compiler; across all src there are nineteen predicate sites and four rest sites.

## Checker census

The program includes every src/**/*.ts file, using upstream tsconfig-base options plus noEmit and node/mocha/chai type declarations. Symbols are resolved through aliases and matched to the core.ts declarations. No textual name-only matching determines this table. Full instantiated signatures, argument types and declaration text are retained in adapted-census.json. An additional symbol-reference audit finds exactly five declarations, eleven imports and twenty-three direct calls, with no local alias uses.

| Upstream site | Adapted site | Overload | Argument expression and checker type | Result consumer |
| --- | --- | --- | --- | --- |
| `src/compiler/utilities.ts:9003` | `src/compiler/utilities.ts:9004` | 4 | `...checks`: `((file: SourceFile, options: CompilerOptions) => Node \| true \| undefined)[]` | combined(file, options), assigned to externalModuleIndicator |
| `src/compiler/utilities.ts:10053` | `src/compiler/utilities.ts:10054` | 3 | `hasJSFileExtension`: `(fileName: string) => boolean`<br>`hasTSFileExtension`: `(fileName: string) => boolean` | hasExtension(text) in usesExtensionsOnImports |
| `src/compiler/checker.ts:4748` | `src/compiler/checker.ts:4749` | 2 | `isImportDeclaration`: `(node: Node) => node is ImportDeclaration`<br>`isJSDocImportTag`: `(node: Node) => node is JSDocImportTag`<br>`isExportDeclaration`: `(node: Node) => node is ExportDeclaration` | findAncestor |
| `src/compiler/checker.ts:4767` | `src/compiler/checker.ts:4768` | 1 | `isImportEqualsDeclaration`: `(node: Node) => node is ImportEqualsDeclaration`<br>`isExportDeclaration`: `(node: Node) => node is ExportDeclaration` | findAncestor |
| `src/compiler/checker.ts:4778` | `src/compiler/checker.ts:4779` | 1 | `isImportEqualsDeclaration`: `(node: Node) => node is ImportEqualsDeclaration`<br>`isExportDeclaration`: `(node: Node) => node is ExportDeclaration` | findAncestor |
| `src/compiler/checker.ts:10839` | `src/compiler/checker.ts:10840` | 3 | `isPropertyDeclaration`: `(node: Node) => node is PropertyDeclaration`<br>`isAccessor`: `(node: Node) => node is AccessorDeclaration`<br>`isVariableDeclaration`: `(node: Node) => node is VariableDeclaration`<br>`isPropertySignature`: `(node: Node) => node is PropertySignature`<br>`isBinaryExpression`: `(node: Node) => node is BinaryExpression`<br>`isPropertyAccessExpression`: `(node: Node) => node is PropertyAccessExpression` | Array.prototype.find |
| `src/compiler/checker.ts:10911` | `src/compiler/checker.ts:10912` | 1 | `isPropertyDeclaration`: `(node: Node) => node is PropertyDeclaration`<br>`isVariableDeclaration`: `(node: Node) => node is VariableDeclaration` | Array.prototype.find |
| `src/compiler/checker.ts:48433` | `src/compiler/checker.ts:48435` | 1 | `isImportDeclaration`: `(node: Node) => node is ImportDeclaration`<br>`isImportEqualsDeclaration`: `(node: Node) => node is ImportEqualsDeclaration` | findAncestor |
| `src/compiler/checker.ts:50073` | `src/compiler/checker.ts:50075` | 2 | `isJSDocLinkLike`: `(node: Node) => node is JSDocLink \| JSDocLinkCode \| JSDocLinkPlain`<br>`isJSDocNameReference`: `(node: Node) => node is import("./types.js").JSDocNameReference`<br>`isJSDocMemberName`: `(node: Node) => node is JSDocMemberName` | findAncestor |
| `src/compiler/checker.ts:50082` | `src/compiler/checker.ts:50084` | 1 | `isClassLike`: `(node: Node) => node is ClassLikeDeclaration`<br>`isInterfaceDeclaration`: `(node: Node) => node is InterfaceDeclaration` | findAncestor |
| `src/services/utilities.ts:491` | `src/services/utilities.ts:491` | 2 | `isJSDocNameReference`: `(node: Node) => node is import("./_namespaces/ts.js").JSDocNameReference`<br>`isJSDocLinkLike`: `(node: Node) => node is JSDocLink \| JSDocLinkCode \| JSDocLinkPlain`<br>`isJSDocMemberName`: `(node: Node) => node is import("./_namespaces/ts.js").JSDocMemberName` | findAncestor |
| `src/services/utilities.ts:3391` | `src/services/utilities.ts:3391` | 3 | `syntaxRequiresTrailingCommaOrSemicolonOrASI`: `(kind: SyntaxKind) => kind is SyntaxKind.PropertySignature \| SyntaxKind.MethodSignature \| SyntaxKind.CallSignature \| SyntaxKind.ConstructSignature \| SyntaxKind.IndexSignature`<br>`syntaxRequiresTrailingFunctionBlockOrSemicolonOrASI`: `(kind: SyntaxKind) => kind is SyntaxKind.MethodDeclaration \| SyntaxKind.Constructor \| SyntaxKind.GetAccessor \| SyntaxKind.SetAccessor \| SyntaxKind.FunctionDeclaration`<br>`syntaxRequiresTrailingModuleBlockOrSemicolonOrASI`: `(kind: SyntaxKind) => kind is SyntaxKind.ModuleDeclaration`<br>`syntaxRequiresTrailingSemicolonOrASI`: `(kind: SyntaxKind) => kind is SyntaxKind.PropertyDeclaration \| SyntaxKind.VariableStatement \| SyntaxKind.ExpressionStatement \| SyntaxKind.DoStatement \| SyntaxKind.ContinueStatement \| SyntaxKind.BreakStatement \| SyntaxKind.ReturnStatement \| SyntaxKind.ThrowStatement \| SyntaxKind.DebuggerStatement \| SyntaxKind.TypeAliasDeclaration \| SyntaxKind.NamespaceExportDeclaration \| SyntaxKind.ImportEqualsDeclaration \| SyntaxKind.ImportDeclaration \| SyntaxKind.ExportAssignment \| SyntaxKind.ExportDeclaration` | syntaxMayBeASICandidate(kind), used by positionIsASICandidate |
| `src/services/refactors/convertImport.ts:115` | `src/services/refactors/convertImport.ts:115` | 1 | `isImportDeclaration`: `(node: import("../_namespaces/ts.js").Node) => node is ImportDeclaration`<br>`isJSDocImportTag`: `(node: import("../_namespaces/ts.js").Node) => node is import("../_namespaces/ts.js").JSDocImportTag` | findAncestor |
| `src/services/codefixes/addMissingResolutionModeImportAttribute.ts:46` | `src/services/codefixes/addMissingResolutionModeImportAttribute.ts:46` | 1 | `isImportDeclaration`: `(node: import("../_namespaces/ts.js").Node) => node is import("../_namespaces/ts.js").ImportDeclaration`<br>`isImportTypeNode`: `(node: import("../_namespaces/ts.js").Node) => node is import("../_namespaces/ts.js").ImportTypeNode` | findAncestor |
| `src/services/codefixes/fixAddMissingMember.ts:724` | `src/services/codefixes/fixAddMissingMember.ts:724` | 2 | `isFunctionTypeNode`: `(node: Node) => node is import("../_namespaces/ts.js").FunctionTypeNode`<br>`isMethodSignature`: `(node: Node) => node is import("../_namespaces/ts.js").MethodSignature`<br>`isMethodDeclaration`: `(node: Node) => node is MethodDeclaration` | core.find |
| `src/services/codefixes/fixUnreferenceableDecoratorMetadata.ts:60` | `src/services/codefixes/fixUnreferenceableDecoratorMetadata.ts:60` | 2 | `isImportClause`: `(node: Node) => node is ImportClause`<br>`isImportSpecifier`: `(node: Node) => node is ImportSpecifier`<br>`isImportEqualsDeclaration`: `(node: Node) => node is ImportEqualsDeclaration` | core.find, after predicate cast |
| `src/services/completions.ts:4716` | `src/services/completions.ts:4716` | 1 | `isSourceFile`: `(node: Node) => node is SourceFile`<br>`isModuleDeclaration`: `(node: Node) => node is ModuleDeclaration` | findAncestor |
| `src/services/completions.ts:5962` | `src/services/completions.ts:5962` | 2 | `isImportDeclaration`: `(node: Node) => node is ImportDeclaration`<br>`isImportEqualsDeclaration`: `(node: Node) => node is ImportEqualsDeclaration`<br>`isJSDocImportTag`: `(node: Node) => node is JSDocImportTag` | findAncestor |
| `src/services/findAllReferences.ts:2298` | `src/services/findAllReferences.ts:2298` | 1 | `isClassLike`: `(node: Node) => node is ClassLikeDeclaration`<br>`isInterfaceDeclaration`: `(node: Node) => node is InterfaceDeclaration` | tryCast |
| `src/services/mapCode.ts:157` | `src/services/mapCode.ts:157` | 1 | `isClassLike`: `(node: Node) => node is ClassLikeDeclaration`<br>`isInterfaceDeclaration`: `(node: Node) => node is import("./_namespaces/ts.js").InterfaceDeclaration` | core.find |
| `src/services/mapCode.ts:163` | `src/services/mapCode.ts:163` | 1 | `isClassLike`: `(node: Node) => node is ClassLikeDeclaration`<br>`isInterfaceDeclaration`: `(node: Node) => node is import("./_namespaces/ts.js").InterfaceDeclaration` | findAncestor |
| `src/services/mapCode.ts:217` | `src/services/mapCode.ts:217` | 1 | `isBlock`: `(node: Node) => node is Block`<br>`isSourceFile`: `(node: Node) => node is SourceFile` | immediate call in findAncestor callback |
| `src/services/smartSelection.ts:206` | `src/services/smartSelection.ts:206` | 1 | `isImportDeclaration`: `(node: Node) => node is import("./_namespaces/ts.js").ImportDeclaration`<br>`isImportEqualsDeclaration`: `(node: Node) => node is import("./_namespaces/ts.js").ImportEqualsDeclaration` | isImport callback to groupChildren |

## Runtime measurement

Completed 398 of 398 unchanged lane units, in 20 pieces of at most 20 and at most two units concurrently. The largest piece took 106.44 seconds; the sum of piece times is 1848.33 seconds. Aggregate counts: {'pass': 106366, 'fail': 1, 'skip': 0}, including the unchanged manifest's expected failure. Every accepted unit passes its ordinary lane identity/API/baseline/test-count guards. These counts include each unit's normal startup; this is the full shard inventory used by the lane, rather than a single monolithic process.

Each closure invocation calls each reached predicate exactly once, unbound, with the original argument list, in original order, and stops on the original truthiness test. The observer records that return value, then updates a candidate with JavaScript || using the same recorded values. It returns the original value. Object.is checks exact value/reference identity, including NaN and signed zero. Calls that throw would make calls differ from completed; reconciliation requires equality. Skipped predicates are not called or inferred to return false.

Built JavaScript call sites are identified by source maps and checked against the checker census. Only three bundles contain or: _tsc.js (10 calls), run.js (23), typescript.js (23). The rest parameter name is read from each emitted function. TypeScript sources and declarations are not instrumented.

T/F/O below means boolean true / boolean false / other. Each semicolon-separated triple is one function position, in fs order. The named counters below disambiguate the variable checks list. A zero observation is not proof of a universal return-type claim.

| Adapted site | Closure calls | Per-position T/F/O | Value differences | Non-boolean in fs observed |
| --- | ---: | --- | ---: | --- |
| `src/compiler/utilities.ts:9004` | 164404 | 1: 0/0/164404; 2: 22/0/118916; 3: 0/0/431 | 0 | yes (283751) |
| `src/compiler/utilities.ts:10054` | 1791 | 1: 101/1690/0; 2: 3/1687/0 | 0 | no observed non-boolean |
| `src/compiler/checker.ts:4749` | 188322 | 1: 62035/126287/0; 2: 199/126088/0; 3: 7053/119035/0 | 0 | predicate overload |
| `src/compiler/checker.ts:4768` | 93 | 1: 0/93/0; 2: 12/81/0 | 0 | predicate overload |
| `src/compiler/checker.ts:4779` | 94 | 1: 0/94/0; 2: 3/91/0 | 0 | predicate overload |
| `src/compiler/checker.ts:10840` | 2487 | 1: 472/2015/0; 2: 276/1739/0; 3: 0/1739/0; 4: 464/1275/0; 5: 564/711/0; 6: 48/663/0 | 0 | no observed non-boolean |
| `src/compiler/checker.ts:10912` | 1546 | 1: 472/1074/0; 2: 0/1074/0 | 0 | predicate overload |
| `src/compiler/checker.ts:48435` | 44 | 1: 12/32/0; 2: 0/32/0 | 0 | predicate overload |
| `src/compiler/checker.ts:50075` | 1658269 | 1: 157/1658112/0; 2: 46/1658066/0; 3: 67/1657999/0 | 0 | predicate overload |
| `src/compiler/checker.ts:50084` | 162 | 1: 8/154/0; 2: 4/150/0 | 0 | predicate overload |
| `src/services/utilities.ts:491` | 168785 | 1: 35/168750/0; 2: 60/168690/0; 3: 66/168624/0 | 0 | predicate overload |
| `src/services/utilities.ts:3391` | 176 | 1: 2/174/0; 2: 2/172/0; 3: 2/170/0; 4: 77/93/0 | 0 | no observed non-boolean |
| `src/services/refactors/convertImport.ts:115` | 291 | 1: 26/265/0; 2: 3/262/0 | 0 | predicate overload |
| `src/services/codefixes/addMissingResolutionModeImportAttribute.ts:46` | 9 | 1: 3/6/0; 2: 1/5/0 | 0 | predicate overload |
| `src/services/codefixes/fixAddMissingMember.ts:724` | 55 | 1: 17/38/0; 2: 35/3/0; 3: 3/0/0 | 0 | predicate overload |
| `src/services/codefixes/fixUnreferenceableDecoratorMetadata.ts:60` | 5 | 1: 1/4/0; 2: 4/0/0; 3: 0/0/0 | 0 | predicate overload |
| `src/services/completions.ts:4716` | 45 | 1: 9/36/0; 2: 4/32/0 | 0 | predicate overload |
| `src/services/completions.ts:5962` | 60 | 1: 11/49/0; 2: 37/12/0; 3: 0/12/0 | 0 | predicate overload |
| `src/services/findAllReferences.ts:2298` | 17 | 1: 15/2/0; 2: 2/0/0 | 0 | predicate overload |
| `src/services/mapCode.ts:157` | 0 | 1: 0/0/0; 2: 0/0/0 | 0 | predicate overload |
| `src/services/mapCode.ts:163` | 28 | 1: 3/25/0; 2: 0/25/0 | 0 | predicate overload |
| `src/services/mapCode.ts:217` | 100 | 1: 29/71/0; 2: 15/56/0 | 0 | predicate overload |
| `src/services/smartSelection.ts:206` | 75 | 1: 2/73/0; 2: 0/73/0 | 0 | predicate overload |

Seven compiler predicate sites: all invoked = True; observed non-boolean returns = 0; value differences = 0. This conclusion is bounded to the recorded oracle. A direct || chain returns the first truthy operand or the last operand, so non-boolean values alone do not imply a value difference. The boolean-return observation is recorded separately from identity equality.

## Named function counters

The function keys in this table use zero-based positions; the runtime table above uses one-based positions.

| Adapted site | Position:function | True | False | Other |
| --- | --- | ---: | ---: | ---: |
| `src/compiler/utilities.ts:9004` | `0:isFileProbablyExternalModule` | 0 | 0 | 164404 |
| `src/compiler/utilities.ts:9004` | `1:isFileForcedToBeModuleByFormat` | 22 | 0 | 118457 |
| `src/compiler/utilities.ts:9004` | `1:isFileModuleFromUsingJSXTag` | 0 | 0 | 459 |
| `src/compiler/utilities.ts:9004` | `2:isFileForcedToBeModuleByFormat` | 0 | 0 | 431 |
| `src/compiler/utilities.ts:10054` | `0:hasJSFileExtension` | 101 | 1690 | 0 |
| `src/compiler/utilities.ts:10054` | `1:hasTSFileExtension` | 3 | 1687 | 0 |
| `src/compiler/checker.ts:4749` | `0:isImportDeclaration` | 62035 | 126287 | 0 |
| `src/compiler/checker.ts:4749` | `1:isJSDocImportTag` | 199 | 126088 | 0 |
| `src/compiler/checker.ts:4749` | `2:isExportDeclaration` | 7053 | 119035 | 0 |
| `src/compiler/checker.ts:4768` | `0:isImportEqualsDeclaration` | 0 | 93 | 0 |
| `src/compiler/checker.ts:4768` | `1:isExportDeclaration` | 12 | 81 | 0 |
| `src/compiler/checker.ts:4779` | `0:isImportEqualsDeclaration` | 0 | 94 | 0 |
| `src/compiler/checker.ts:4779` | `1:isExportDeclaration` | 3 | 91 | 0 |
| `src/compiler/checker.ts:10840` | `0:isPropertyDeclaration` | 472 | 2015 | 0 |
| `src/compiler/checker.ts:10840` | `1:isAccessor` | 276 | 1739 | 0 |
| `src/compiler/checker.ts:10840` | `2:isVariableDeclaration` | 0 | 1739 | 0 |
| `src/compiler/checker.ts:10840` | `3:isPropertySignature` | 464 | 1275 | 0 |
| `src/compiler/checker.ts:10840` | `4:isBinaryExpression` | 564 | 711 | 0 |
| `src/compiler/checker.ts:10840` | `5:isPropertyAccessExpression` | 48 | 663 | 0 |
| `src/compiler/checker.ts:10912` | `0:isPropertyDeclaration` | 472 | 1074 | 0 |
| `src/compiler/checker.ts:10912` | `1:isVariableDeclaration` | 0 | 1074 | 0 |
| `src/compiler/checker.ts:48435` | `0:isImportDeclaration` | 12 | 32 | 0 |
| `src/compiler/checker.ts:48435` | `1:isImportEqualsDeclaration` | 0 | 32 | 0 |
| `src/compiler/checker.ts:50075` | `0:isJSDocLinkLike` | 157 | 1658112 | 0 |
| `src/compiler/checker.ts:50075` | `1:isJSDocNameReference` | 46 | 1658066 | 0 |
| `src/compiler/checker.ts:50075` | `2:isJSDocMemberName` | 67 | 1657999 | 0 |
| `src/compiler/checker.ts:50084` | `0:isClassLike` | 8 | 154 | 0 |
| `src/compiler/checker.ts:50084` | `1:isInterfaceDeclaration` | 4 | 150 | 0 |
| `src/services/utilities.ts:491` | `0:isJSDocNameReference` | 35 | 168750 | 0 |
| `src/services/utilities.ts:491` | `1:isJSDocLinkLike` | 60 | 168690 | 0 |
| `src/services/utilities.ts:491` | `2:isJSDocMemberName` | 66 | 168624 | 0 |
| `src/services/utilities.ts:3391` | `0:syntaxRequiresTrailingCommaOrSemicolonOrASI` | 2 | 174 | 0 |
| `src/services/utilities.ts:3391` | `1:syntaxRequiresTrailingFunctionBlockOrSemicolonOrASI` | 2 | 172 | 0 |
| `src/services/utilities.ts:3391` | `2:syntaxRequiresTrailingModuleBlockOrSemicolonOrASI` | 2 | 170 | 0 |
| `src/services/utilities.ts:3391` | `3:syntaxRequiresTrailingSemicolonOrASI` | 77 | 93 | 0 |
| `src/services/refactors/convertImport.ts:115` | `0:isImportDeclaration` | 26 | 265 | 0 |
| `src/services/refactors/convertImport.ts:115` | `1:isJSDocImportTag` | 3 | 262 | 0 |
| `src/services/codefixes/addMissingResolutionModeImportAttribute.ts:46` | `0:isImportDeclaration` | 3 | 6 | 0 |
| `src/services/codefixes/addMissingResolutionModeImportAttribute.ts:46` | `1:isImportTypeNode` | 1 | 5 | 0 |
| `src/services/codefixes/fixAddMissingMember.ts:724` | `0:isFunctionTypeNode` | 17 | 38 | 0 |
| `src/services/codefixes/fixAddMissingMember.ts:724` | `1:isMethodSignature` | 35 | 3 | 0 |
| `src/services/codefixes/fixAddMissingMember.ts:724` | `2:isMethodDeclaration` | 3 | 0 | 0 |
| `src/services/codefixes/fixUnreferenceableDecoratorMetadata.ts:60` | `0:isImportClause` | 1 | 4 | 0 |
| `src/services/codefixes/fixUnreferenceableDecoratorMetadata.ts:60` | `1:isImportSpecifier` | 4 | 0 | 0 |
| `src/services/completions.ts:4716` | `0:isSourceFile` | 9 | 36 | 0 |
| `src/services/completions.ts:4716` | `1:isModuleDeclaration` | 4 | 32 | 0 |
| `src/services/completions.ts:5962` | `0:isImportDeclaration` | 11 | 49 | 0 |
| `src/services/completions.ts:5962` | `1:isImportEqualsDeclaration` | 37 | 12 | 0 |
| `src/services/completions.ts:5962` | `2:isJSDocImportTag` | 0 | 12 | 0 |
| `src/services/findAllReferences.ts:2298` | `0:isClassLike` | 15 | 2 | 0 |
| `src/services/findAllReferences.ts:2298` | `1:isInterfaceDeclaration` | 2 | 0 | 0 |
| `src/services/mapCode.ts:163` | `0:isClassLike` | 3 | 25 | 0 |
| `src/services/mapCode.ts:163` | `1:isInterfaceDeclaration` | 0 | 25 | 0 |
| `src/services/mapCode.ts:217` | `0:isBlock` | 29 | 71 | 0 |
| `src/services/mapCode.ts:217` | `1:isSourceFile` | 15 | 56 | 0 |
| `src/services/smartSelection.ts:206` | `0:isImportDeclaration` | 2 | 73 | 0 |
| `src/services/smartSelection.ts:206` | `1:isImportEqualsDeclaration` | 0 | 73 | 0 |

Non-boolean types (position is zero-based in raw keys):

- `src/compiler/utilities.ts:9004`: `{"0:object": 45466, "0:undefined": 118938, "1:object": 28, "1:undefined": 118888, "2:undefined": 431}`

## Setup, validation and retained evidence

Commands source /workspace/adamic-tools/env.sh. cloud/setup.sh, npm ci --prefix stage3/api, the pinned source clone, full apply --build-product, npm ci in the measurement tree, npm run build and npm exec hereby -- tests all exit 0. ADAMIC_TYPESCRIPT_SOURCE is the pinned checkout for apply. Every wrapper reports the child exit code. The post-apply checkpoint push points to the base commit, allowing the final branch to add exactly one report commit.

Two startup attempts are excluded: the first had a duplicate node_modules link in the private view; the second referred to fs instead of the bundle-renamed fs3 and failed at module initialization. Neither ran tests or produced accepted runtime observations. Both attempts and the corrections are retained. The corrected observer passes 1,728 independent return/binding/order/argument/short-circuit seeds and an instrumented-vs-original compiler output/diagnostic smoke comparison.

The reconciler checks the exact planned unit set, unique unit names, successful lane verdicts, raw test counts, unit input hashes, complete closure returns, exact count-file hashes and zero Object.is differences. summary.json contains the complete raw per-site aggregation. evidence.tar.gz contains per-unit reports, process count files, logs, scripts, source hashes, frozen lane plan and provenance, with an internal hash manifest. Build products are excluded. Observers are removed from the scratch bundles afterward and source hashes are rechecked. No started measurement process is left running.
