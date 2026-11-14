# Paired latent census

| Measure | Raw findings | Distinct reads |
| --- | ---: | ---: |
| Lowered (no remaining finding) | 0 | 0 |
| Moved by reason | 0 | 0 |
| New findings | 0 | 0 |
| New sites | 0 | 0 |
| Units whose first stop moved | 0 | 0 |
| Regressions | 0 | 0 |
| before NotYet | 52 | 52 |
| before Refused | 143 | 139 |
| after NotYet | 52 | 52 |
| after Refused | 143 | 139 |

| Reason | Raw before | Raw after | Reads before | Reads after | Reads lowered |
| --- | ---: | ---: | ---: | ---: | ---: |
| a CompilerOptions seen as { extendedDiagnostics?: boolean; diagnostics?: boolean; } (fixed objects and records have different storage; copy explicitly) | 1 | 1 | 1 | 1 | 0 |
| a Map or Set interface implemented by a value that isn't a Map or Set | 3 | 3 | 3 | 3 | 0 |
| a branded supertype inside a checked predicate's array target | 2 | 2 | 2 | 2 | 0 |
| a cast that makes a viewed readonly array writable | 1 | 1 | 1 | 1 | 0 |
| a cast the runtime can't check | 31 | 31 | 31 | 31 | 0 |
| a checked predicate target whose check would read past the depth bound (IncrementalBundleEmitBuildInfo) | 1 | 1 | 1 | 1 | 0 |
| a computed field name | 1 | 1 | 1 | 1 | 0 |
| a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 | 1 | 0 |
| a generic function value without a concrete contextual signature | 1 | 1 | 1 | 1 | 0 |
| a generic instantiation whose type argument for TIn, constrained to Node \| undefined, can't be read | 17 | 17 | 17 | 17 | 0 |
| a generic instantiation whose type argument for TIn, constrained to Node, can't be read | 9 | 9 | 9 | 9 | 0 |
| a method in object destructuring | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getCanonicalFileName would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getParsedCommandLine would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSourceFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (getSourceFileByPath would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasInvalidatedLibResolutions would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (hasInvalidatedResolutions would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (liftToBlock would lose its object, and this with it) | 5 | 5 | 5 | 5 | 0 |
| a method read as a value (onReleaseOldSourceFile would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (onReleaseParsedCommandLine would lose its object, and this with it) | 1 | 1 | 1 | 1 | 0 |
| a method read as a value (toKey would lose its object, and this with it) | 2 | 2 | 2 | 2 | 0 |
| a method sharing its name with a getter; rename the getter or use a callable arrow field in the interface and implementation | 2 | 2 | 2 | 2 | 0 |
| a read narrowed to undefined of a binding a call may write | 1 | 1 | 1 | 1 | 0 |
| a read the checker narrowed to never of a binding a call may write | 1 | 1 | 1 | 1 | 0 |
| a template interpolating an object, an array, a map, a function or undefined | 1 | 1 | 1 | 1 | 0 |
| a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 1 | 1 | 1 | 1 | 0 |
| a value of type AssignmentExpression<EqualsToken> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Block seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type CachedDirectoryStructureHost \| WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as DirectoryStructureHost, whose method getDirectories becomes a writable field: a write of ((path: string) => string[]) \| undefined would replace a method of type (path: string) => string[] | 1 | 1 | 1 | 1 | 0 |
| a value of type CallExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type ClassExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost & ResolutionCacheHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost & ResolutionCacheHost seen as ResolutionCacheHost, which can write boolean \| (() => boolean) \| undefined where (() => boolean) & (boolean \| (() => boolean) \| undefined) is read | 1 | 1 | 1 | 1 | 0 |
| a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 1 | 1 | 1 | 1 | 0 |
| a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 | 4 | 4 | 4 | 0 |
| a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 | 4 | 4 | 4 | 0 |
| a value of type NodeArray<ClassElement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<ParameterDeclaration> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type NodeArray<Statement> seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 | 3 | 0 |
| a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type PropertyAccessExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type StringLiteral seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableDeclaration seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type VariableStatement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 | 2 | 0 |
| a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as (CompilerHost \| ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }, which can write DiagnosticReporter \| undefined where DiagnosticReporter is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as DirectoryStructureHost, which can write ((path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]) \| undefined where ((path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]) & ((path: string, extensions?: ... \| undefined, exclude?: ... \| undefined, include?: ... \| undefined, depth?: number \| undefined) => .... is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T> seen as WatchFactoryHost & { trace?(s: string): void; }, which can write boolean \| (() => boolean) where () => boolean is read | 1 | 1 | 1 | 1 | 0 |
| a value of type WildcardDirectoryWatcher<T> seen as { watcher: FileWatcher; }, which can write FileWatcher where T is read | 2 | 2 | 2 | 2 | 0 |
| a value of type YieldExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 | 1 | 0 |
| a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 1 | 1 | 1 | 1 | 0 |
| a value of type any: any has no proven type | 1 | 1 | 1 | 1 | 0 |
| a value of type never[] seen as Path[], which can write Path where never is read | 1 | 1 | 1 | 1 | 0 |
| a value of type string[] seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what string[] can't hold | 1 | 1 | 1 | 1 | 0 |
| a write through a union of string into a slot that holds false in FilePresenceUnknownOnHost | 1 | 1 | 1 | 1 | 0 |
| an object refinement using an open numeric enum as a literal tag | 29 | 29 | 25 | 25 | 0 |
| an unproven relation from Node \| undefined to T \| undefined: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven relation from string to T: the source is not assignable to the target | 1 | 1 | 1 | 1 | 0 |
| an unproven value assigned to a numeric literal or enum member slot SyntaxKind.Identifier | 1 | 1 | 1 | 1 | 0 |
| function identity comparison in a program with specialized generic function values | 1 | 1 | 1 | 1 | 0 |
| new an Identifier | 1 | 1 | 1 | 1 | 0 |
| optional property id in ExternalModuleReference absent from structural source never, which can hide fields | 2 | 2 | 2 | 2 | 0 |
| optional property source in SourceMapRange absent from structural source TextRange, which can hide fields | 4 | 4 | 4 | 4 | 0 |
| optional record property access | 1 | 1 | 1 | 1 | 0 |
| overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined | 1 | 1 | 1 | 1 | 0 |
| overload 1 of getProperties result readonly InitializedPropertyDeclaration[] cannot be served by implementation result readonly PropertyDeclaration[] | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitArray parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitArrayWorker parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitNodes parameter visitor cannot be served by implementation parameter visitor | 1 | 1 | 1 | 1 | 0 |
| overload 1 of visitParameterList parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | 1 | 1 | 0 |
| overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor | 1 | 1 | 1 | 1 | 0 |
| writing a checked object field that may hold undefined or null, whose checked write can't change the representation its slot holds | 3 | 3 | 3 | 3 | 0 |
| writing a checked object field without its source-slot type certificate | 1 | 1 | 1 | 1 | 0 |
| writing a possibly absent optional own field | 2 | 2 | 2 | 2 | 0 |

## Raw Findings

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

## Distinct Reads

Moved by reason:


New findings:

None.

Units whose first stop moved:

None.

Panics: 0 before, 0 after.

The JSON retains every lowered site, move, new finding, unit stop and panic.
