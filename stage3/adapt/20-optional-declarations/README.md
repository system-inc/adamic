# 20: truthful optional declarations

This adaptation widens an owned optional declaration from `p?: T` to
`p?: (T) | undefined` when a TS2412, TS2375 or TS2379 finding supplies evidence
that the implementation deliberately stores undefined. The question mark,
assignments, object keys and every runtime expression remain. A present undefined
key has different `in`, `hasOwn`, enumeration and spread behavior from an absent
key. Removing a write or conditionally constructing a key is not this adaptation.

The resolver uses stock npm TypeScript 6.0.3's compiler API, not source regexes.
It parses current text, resolves property symbols to their declarations, checks
that every non-undefined source constituent fits the declared value type, and
edits each selected owner once per program. It rechecks the whole dependency
closure after each pass, stopping only when no eligible owner remains. This also
tolerates preceding type-import adaptations: no source line is a rewrite address.
Type insertions use parsed UTF-16 offsets and preserve existing source trivia.
Nested type wrappers compose. An optional interface method becomes an optional
callable property with the union around the entire callable. Live methods,
accessors and overload sets are excluded.

Both absence and present undefined are now admitted. Consumers must therefore
prove more than key presence before using the value. Every new or changed
stock-checker diagnostic is reported with location in `newly_exposed_consumers`;
this unit does not edit those consumers or repair base-interface relationships.
The evidence-to-owner ledger is `resolved_findings`. `initially_declined` retains
seed exclusions even if another legitimate write to a shared owner removes the
finding incidentally. `declined` lists remaining optional-code findings and reasons.

The fixed checker options are strict, exactOptionalPropertyTypes,
noUncheckedIndexedAccess, noImplicitReturns, noFallthroughCasesInSwitch,
erasableSyntaxOnly and verbatimModuleSyntax; ES2024, ESNext modules, Bundler
resolution, force module detection, ES2024 library and no implicit type packages.
The stock resolver is tooling for rejected source. The census's Adamic checker
remains the measurement authority, including its sounder regex declarations.

## Running

Input is the externally cached TypeScript tree pinned by stage 3: repository
`https://github.com/microsoft/TypeScript.git`, tag `v6.0.3`, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. No upstream source is committed here.
Generate upstream's diagnostics file before adapting a prepared compiler tree.

Install the stock tool in scratch, then run the requested command directly:

```sh
npm install --prefix /tmp/adamic-optional/npm typescript@6.0.3 --no-audit --no-fund
export NODE_PATH=/tmp/adamic-optional/npm/node_modules
node stage3/adapt/20-optional-declarations/adapt.cjs /tmp/adamic-optional/typescript > /tmp/adamic-optional/adapt.json 2> /tmp/adamic-optional/adapt.stderr
node stage3/adapt/20-optional-declarations/adapt.cjs /tmp/adamic-optional/typescript > /tmp/adamic-optional/idempotence.json 2> /tmp/adamic-optional/idempotence.stderr
```

`TSC_ADAPT_TYPESCRIPT` may instead name the stock `lib/typescript.js` explicitly.
Otherwise resolution uses the tooling's normal Node module search, then the
checkout's dependencies. A version other than 6.0.3 is rejected, including
upstream's locked build compiler (5.9.3). The JSON report goes to stdout and
progress to stderr; only eligible owner declarations are written in the checkout.

## Declined contracts

The original survey's C and R cases remain separate obligations:

| Case | Original pinned site | Reason |
| --- | --- | --- |
| E022, E-index | `src/compiler/checker.ts:33923` | `childrenTypes[0]` is an unproven indexed read, even after `length === 1`; it cannot seed a declaration edit. |
| E023, E-index | `src/compiler/checker.ts:43517` | `getTypeArguments(type as GenericType)[0]` requires an arity/presence invariant. |
| E070, E-generic | `src/compiler/factory/nodeFactory.ts:5202` | `T["comment"]` may be a narrower arbitrary specialization. |
| E027, E-other | `src/compiler/emitter.ts:1014` | The expanded chain concerns `Bundle \| SourceFile` and missing required fields, not optionality. |
| E087, additional generic review | `src/compiler/resolutionCache.ts:1386` | The reset receiver is a type parameter; its specialization may narrow the field. |

A bare indexed read is never a seed. An existing explicit undefined branch can
be evidence of intentional absence. Generic/unknown/indexed-access value
contracts, required or inferred slots, external owners and live methods remain
unchanged. A structural chain must actually end in an undefined incompatibility;
a diagnostic code alone is insufficient. Shared-owner changes can discharge an
indexed diagnostic incidentally, but do not prove the indexed read invariant.

## Measurement and proof

The sections below record observed runs, rather than extrapolating the survey.

Input preparation generated `diagnosticInformationMap.generated.ts`, giving 78
compiler roots. Its `reportsUnnecessary` literal adds one TS2375 beyond the
historical 77-root survey. Adaptation 10 marks all 3,719 diagnosed imports,
including its generated `DiagnosticMessage` import, before this unit is measured.
The census commit is `429c117`; the branch began at main `ef3d907`.

### Whole-program census

The unchanged Adamic gate reports **2833 → 2163** checker diagnostics. No entry
in this whole-program check reaches lowering. All reason counts follow, including
zeros for reasons that appear on only one side.

| Reason | After 10 | After 10 + 20 |
| --- | ---: | ---: |
| TS1294 | 180 | 180 |
| TS1484 | 0 | 0 |
| TS2304 | 6 | 6 |
| TS2320 | 0 | 6 |
| TS2322 | 120 | 119 |
| TS2339 | 28 | 28 |
| TS2345 | 733 | 738 |
| TS2366 | 1 | 0 |
| TS2375 | 77 | 15 |
| TS2379 | 31 | 15 |
| TS2412 | 617 | 6 |
| TS2420 | 1 | 1 |
| TS2430 | 0 | 10 |
| TS2488 | 11 | 11 |
| TS2532 | 225 | 225 |
| TS2538 | 7 | 7 |
| TS2556 | 2 | 2 |
| TS2591 | 54 | 54 |
| TS2684 | 2 | 2 |
| TS2722 | 2 | 2 |
| TS2740 | 1 | 1 |
| TS2769 | 10 | 10 |
| TS7006 | 1 | 1 |
| TS7029 | 83 | 83 |
| TS7030 | 252 | 252 |
| TS7031 | 1 | 1 |
| TS18046 | 8 | 8 |
| TS18048 | 380 | 380 |

TS2412/2375/2379 total **725 → 36**, a net reduction of **689**. Other counts
net +19: TS2345 rises by 5, TS2320 by 6 and TS2430 by 10 because widening
callable/field owners exposes incompatible consumer and base-interface contracts.
TS2322 falls by 1 and TS2366 by 1. Every other count is unchanged. These
increases are reported rather than repaired.

### Resolver counts and source diff

During shared apply, before upstream dependencies are installed, the stock npm
checker reports **2728 → 2057**, with TS2412 **617 → 6**,
TS2375 **77 → 15** and TS2379 **31 → 15**. Its regex library and other
checker differences explain why its nonoptional counts differ from Adamic.
One TS2307 is the not-yet-installed upstream `source-map-support` dependency;
apply does not install upstream build dependencies. The earlier direct run with
those dependencies installed and without adaptation 10 was **6446 → 5775**;
subtracting its 3,719 import findings gives **2727 → 2056**. Owner edits, evidence,
declines and the complete new-consumer ledger are identical in both runs.
The script edits **405 distinct declarations** for a starting population of
**725 optional-code findings**. The evidence ledger has 829 property-to-owner
rows across rechecks, representing 737 distinct diagnostic messages/locations;
that is not 737 independent original findings, since rechecking changes chains
and introduces new structural findings. All sharing resolves to one owner edit.
The second run edits **0** declarations.

Upstream source diff: **26 files, 406 lines replaced** (406 insertions and
406 deletions), all declaration types or optional interface signatures. Owners
sharing a line and multiline wrappers mean changed-line counts need not equal
owner counts.

| File | Replaced lines |
| --- | ---: |
| `src/compiler/builder.ts` | 13 |
| `src/compiler/builderPublic.ts` | 2 |
| `src/compiler/builderState.ts` | 2 |
| `src/compiler/checker.ts` | 4 |
| `src/compiler/commandLineParser.ts` | 6 |
| `src/compiler/emitter.ts` | 4 |
| `src/compiler/moduleNameResolver.ts` | 2 |
| `src/compiler/moduleSpecifiers.ts` | 1 |
| `src/compiler/parser.ts` | 2 |
| `src/compiler/performanceCore.ts` | 2 |
| `src/compiler/program.ts` | 3 |
| `src/compiler/programDiagnostics.ts` | 1 |
| `src/compiler/resolutionCache.ts` | 5 |
| `src/compiler/sourcemap.ts` | 5 |
| `src/compiler/tracing.ts` | 2 |
| `src/compiler/transformers/declarations/diagnostics.ts` | 1 |
| `src/compiler/transformers/destructuring.ts` | 1 |
| `src/compiler/transformers/es2015.ts` | 1 |
| `src/compiler/transformers/esDecorators.ts` | 5 |
| `src/compiler/transformers/jsx.ts` | 1 |
| `src/compiler/tsbuildPublic.ts` | 6 |
| `src/compiler/types.ts` | 311 |
| `src/compiler/utilities.ts` | 1 |
| `src/compiler/watch.ts` | 6 |
| `src/compiler/watchPublic.ts` | 18 |
| `src/compiler/watchUtilities.ts` | 1 |

### Consumers exposed by the census

The following **35 new or changed diagnostic chains** are the conservative
whole-program delta. A changed chain can be a persisting error with a different
structural cause. Locations are in the adapted pinned tree. None was fixed.

```text
src/compiler/checker.ts:1683:9: error TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & ...) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBuilderF...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'ArrowFunction' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'ArrowFunction' is not assignable to type 'ArrowFunction & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'ArrowFunction' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
src/compiler/checker.ts:21434:98: error TS2379: Argument of type '{ errors?: Diagnostic[] | undefined; }' is not assignable to parameter of type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'errors' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/checker.ts:46159:176: error TS2345: Argument of type 'ErrorOutputContainer | undefined' is not assignable to parameter of type '{ errors?: Diagnostic[]; } | undefined'.
  Type 'ErrorOutputContainer' is not assignable to type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'errors' are incompatible.
      Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
        Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/commandLineParser.ts:2677:9: error TS2322: Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }[] | undefined' is not assignable to type 'readonly ProjectReference[] | undefined'.
  Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.
    Type '{ prepend?: (boolean) | undefined; circular?: (boolean) | undefined; path: string; originalPath: undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
      Types of property 'originalPath' are incompatible.
        Type 'undefined' is not assignable to type 'string'.
src/compiler/emitter.ts:1011:33: error TS2345: Argument of type 'Bundle | SourceFile' is not assignable to parameter of type 'Bundle'.
  Property 'sourceFiles' is missing in type 'SourceFile' but required in type 'Bundle'.
src/compiler/emitter.ts:1014:31: error TS2345: Argument of type 'Bundle | SourceFile' is not assignable to parameter of type 'SourceFile'.
  Type 'Bundle' is missing the following properties from type 'SourceFile': statements, endOfFileToken, fileName, path, and 27 more.
src/compiler/moduleNameResolver.ts:1271:11: error TS2320: Interface 'ModuleOrTypeReferenceResolutionCache<T>' cannot simultaneously extend types 'PerDirectoryResolutionCache<T>' and 'NonRelativeNameResolutionCache<T>'.
  Named property 'isReadonly' of types 'PerDirectoryResolutionCache<T>' and 'NonRelativeNameResolutionCache<T>' are not identical.
src/compiler/moduleNameResolver.ts:1450:97: error TS2345: Argument of type 'ModuleResolutionCache | undefined' is not assignable to parameter of type 'NonRelativeModuleNameResolutionCache | undefined'.
  Type 'ModuleResolutionCache' is not assignable to type 'NonRelativeModuleNameResolutionCache' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'isReadonly' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/compiler/moduleNameResolver.ts:3035:85: error TS2345: Argument of type 'ModuleResolutionCache | undefined' is not assignable to parameter of type 'NonRelativeModuleNameResolutionCache | undefined'.
  Type 'ModuleResolutionCache' is not assignable to type 'NonRelativeModuleNameResolutionCache' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'isReadonly' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/compiler/moduleNameResolver.ts:852:18: error TS2320: Interface 'TypeReferenceDirectiveResolutionCache' cannot simultaneously extend types 'PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' and 'NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>'.
  Named property 'isReadonly' of types 'PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' and 'NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' are not identical.
src/compiler/moduleNameResolver.ts:900:18: error TS2320: Interface 'ModuleResolutionCache' cannot simultaneously extend types 'PerDirectoryResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'NonRelativeModuleNameResolutionCache'.
  Named property 'isReadonly' of types 'PerDirectoryResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'NonRelativeModuleNameResolutionCache' are not identical.
src/compiler/moduleNameResolver.ts:910:18: error TS2320: Interface 'NonRelativeModuleNameResolutionCache' cannot simultaneously extend types 'NonRelativeNameResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'PackageJsonInfoCache'.
  Named property 'isReadonly' of types 'NonRelativeNameResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'PackageJsonInfoCache' are not identical.
src/compiler/program.ts:5105:5: error TS2375: Type '(CompilerHost | ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Type 'CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'directoryExists' are incompatible.
      Type '((directoryName: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
        Type 'undefined' is not assignable to type '(path: string) => boolean'.
src/compiler/resolutionCache.ts:180:18: error TS2320: Interface 'CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations' cannot simultaneously extend types 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations'.
  Named property 'affectingLocations' of types 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations' are not identical.
src/compiler/resolutionCache.ts:180:18: error TS2320: Interface 'CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations' cannot simultaneously extend types 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations'.
  Named property 'failedLookupLocations' of types 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations' are not identical.
src/compiler/tracing.ts:175:17: error TS2339: Property 'phase' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:24: error TS2339: Property 'name' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:30: error TS2339: Property 'args' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:36: error TS2339: Property 'time' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:42: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tsbuildPublic.ts:787:5: error TS2322: Type 'WriteFileCallback | undefined' is not assignable to type '((path: string, data: string, writeByteOrderMark?: boolean | undefined) => void) | undefined'.
  Type 'WriteFileCallback' is not assignable to type '(path: string, data: string, writeByteOrderMark?: boolean | undefined) => void'.
    Types of parameters 'writeByteOrderMark' and 'writeByteOrderMark' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/compiler/types.ts:1840:18: error TS2430: Interface 'SignatureDeclarationBase' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'PropertyName | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:2085:18: error TS2430: Interface 'FunctionDeclaration' incorrectly extends interface 'DeclarationStatement'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'Identifier | NumericLiteral | StringLiteral'.
      Type 'undefined' is not assignable to type 'Identifier | NumericLiteral | StringLiteral'.
src/compiler/types.ts:3543:18: error TS2430: Interface 'ClassLikeDeclarationBase' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:3551:18: error TS2430: Interface 'ClassDeclaration' incorrectly extends interface 'DeclarationStatement'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'Identifier | NumericLiteral | StringLiteral'.
      Type 'undefined' is not assignable to type 'Identifier | NumericLiteral | StringLiteral'.
src/compiler/types.ts:3567:18: error TS2430: Interface 'ClassElement' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'PropertyName | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:3572:18: error TS2430: Interface 'TypeElement' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'PropertyName | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:3719:18: error TS2430: Interface 'ImportClause' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:4071:18: error TS2430: Interface 'JSDocTypedefTag' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:4079:18: error TS2430: Interface 'JSDocCallbackTag' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:6659:18: error TS2430: Interface 'DeferredTypeReference' incorrectly extends interface 'TypeReference'.
  Types of property 'mapper' are incompatible.
    Type 'TypeMapper | undefined' is not assignable to type 'TypeMapper'.
      Type 'undefined' is not assignable to type 'TypeMapper'.
src/compiler/watch.ts:754:118: error TS2375: Type 'ProgramHost<any>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'directoryExists' are incompatible.
    Type '((path: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
      Type 'undefined' is not assignable to type '(path: string) => boolean'.
src/compiler/watchPublic.ts:383:52: error TS2379: Argument of type '{ configFileName: string; optionsToExtend: CompilerOptions | undefined; watchOptionsToExtend: WatchOptions; extraFileExtensions: readonly FileExtensionInfo[]; system: System; createProgram: CreateProgram<...> | undefined; reportDiagnostic: DiagnosticReporter | undefined; reportWatchStatus: WatchStatusReporter | unde...' is not assignable to parameter of type 'CreateWatchCompilerHostOfConfigFileInput<T>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'createProgram' are incompatible.
    Type 'CreateProgram<T> | undefined' is not assignable to type 'CreateProgram<T>'.
      Type 'undefined' is not assignable to type 'CreateProgram<T>'.
src/compiler/watchPublic.ts:463:120: error TS2379: Argument of type 'WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>' is not assignable to parameter of type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Type 'WatchCompilerHostOfConfigFile<T> & Partial<WatchCompilerHostOfFilesAndCompilerOptions<T>>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'directoryExists' are incompatible.
      Type '((path: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
        Type 'undefined' is not assignable to type '(path: string) => boolean'.
src/compiler/watchPublic.ts:464:11: error TS2375: Type 'CachedDirectoryStructureHost | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Type 'WatchCompilerHostOfConfigFile<T> & Partial<WatchCompilerHostOfFilesAndCompilerOptions<T>>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'directoryExists' are incompatible.
      Type '((path: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
        Type 'undefined' is not assignable to type '(path: string) => boolean'.
```

### Consumers exposed by the stock resolver

Stock 6.0.3 additionally records every new or changed chain in its JSON report;
its 34 entries follow. They overlap with the census list but are retained because
the checkers do not have identical standard libraries or diagnostics.

```text
src/compiler/checker.ts:1683:9: TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & { ...; }) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBui...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'FunctionExpression' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'FunctionExpression' is not assignable to type 'FunctionExpression & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'FunctionExpression' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
src/compiler/checker.ts:21434:98: TS2379: Argument of type '{ errors?: Diagnostic[] | undefined; }' is not assignable to parameter of type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'errors' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/checker.ts:46159:176: TS2345: Argument of type 'ErrorOutputContainer | undefined' is not assignable to parameter of type '{ errors?: Diagnostic[]; } | undefined'.
  Type 'ErrorOutputContainer' is not assignable to type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'errors' are incompatible.
      Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
        Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/commandLineParser.ts:2677:9: TS2322: Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }[] | undefined' is not assignable to type 'readonly ProjectReference[] | undefined'.
  Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.
    Type '{ path: string; originalPath: undefined; prepend?: (boolean) | undefined; circular?: (boolean) | undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
      Types of property 'originalPath' are incompatible.
        Type 'undefined' is not assignable to type 'string'.
src/compiler/emitter.ts:1011:33: TS2345: Argument of type 'SourceFile | Bundle' is not assignable to parameter of type 'Bundle'.
  Property 'sourceFiles' is missing in type 'SourceFile' but required in type 'Bundle'.
src/compiler/emitter.ts:1014:31: TS2345: Argument of type 'SourceFile | Bundle' is not assignable to parameter of type 'SourceFile'.
  Type 'Bundle' is missing the following properties from type 'SourceFile': statements, endOfFileToken, fileName, path, and 27 more.
src/compiler/moduleNameResolver.ts:852:18: TS2320: Interface 'TypeReferenceDirectiveResolutionCache' cannot simultaneously extend types 'PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' and 'NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>'.
  Named property 'isReadonly' of types 'PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' and 'NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' are not identical.
src/compiler/moduleNameResolver.ts:900:18: TS2320: Interface 'ModuleResolutionCache' cannot simultaneously extend types 'PerDirectoryResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'NonRelativeModuleNameResolutionCache'.
  Named property 'isReadonly' of types 'PerDirectoryResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'NonRelativeModuleNameResolutionCache' are not identical.
src/compiler/moduleNameResolver.ts:910:18: TS2320: Interface 'NonRelativeModuleNameResolutionCache' cannot simultaneously extend types 'NonRelativeNameResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'PackageJsonInfoCache'.
  Named property 'isReadonly' of types 'NonRelativeNameResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'PackageJsonInfoCache' are not identical.
src/compiler/moduleNameResolver.ts:1271:11: TS2320: Interface 'ModuleOrTypeReferenceResolutionCache<T>' cannot simultaneously extend types 'PerDirectoryResolutionCache<T>' and 'NonRelativeNameResolutionCache<T>'.
  Named property 'isReadonly' of types 'PerDirectoryResolutionCache<T>' and 'NonRelativeNameResolutionCache<T>' are not identical.
src/compiler/moduleNameResolver.ts:1450:97: TS2345: Argument of type 'ModuleResolutionCache | undefined' is not assignable to parameter of type 'NonRelativeModuleNameResolutionCache | undefined'.
  Type 'ModuleResolutionCache' is not assignable to type 'NonRelativeModuleNameResolutionCache' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'isReadonly' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/compiler/moduleNameResolver.ts:3035:85: TS2345: Argument of type 'ModuleResolutionCache | undefined' is not assignable to parameter of type 'NonRelativeModuleNameResolutionCache | undefined'.
  Type 'ModuleResolutionCache' is not assignable to type 'NonRelativeModuleNameResolutionCache' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'isReadonly' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/compiler/program.ts:5105:5: TS2375: Type '(CompilerHost | ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Type 'CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'directoryExists' are incompatible.
      Type '((directoryName: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
        Type 'undefined' is not assignable to type '(path: string) => boolean'.
src/compiler/resolutionCache.ts:180:18: TS2320: Interface 'CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations' cannot simultaneously extend types 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations'.
  Named property 'affectingLocations' of types 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations' are not identical.
src/compiler/tracing.ts:175:17: TS2339: Property 'phase' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:24: TS2339: Property 'name' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:30: TS2339: Property 'args' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:36: TS2339: Property 'time' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:42: TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tsbuildPublic.ts:787:5: TS2322: Type 'WriteFileCallback | undefined' is not assignable to type '((path: string, data: string, writeByteOrderMark?: boolean | undefined) => void) | undefined'.
  Type 'WriteFileCallback' is not assignable to type '(path: string, data: string, writeByteOrderMark?: boolean | undefined) => void'.
    Types of parameters 'writeByteOrderMark' and 'writeByteOrderMark' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/compiler/types.ts:1840:18: TS2430: Interface 'SignatureDeclarationBase' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'PropertyName | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:2085:18: TS2430: Interface 'FunctionDeclaration' incorrectly extends interface 'DeclarationStatement'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'Identifier | StringLiteral | NumericLiteral'.
      Type 'undefined' is not assignable to type 'Identifier | StringLiteral | NumericLiteral'.
src/compiler/types.ts:3543:18: TS2430: Interface 'ClassLikeDeclarationBase' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:3551:18: TS2430: Interface 'ClassDeclaration' incorrectly extends interface 'DeclarationStatement'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'Identifier | StringLiteral | NumericLiteral'.
      Type 'undefined' is not assignable to type 'Identifier | StringLiteral | NumericLiteral'.
src/compiler/types.ts:3567:18: TS2430: Interface 'ClassElement' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'PropertyName | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:3572:18: TS2430: Interface 'TypeElement' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'PropertyName | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:3719:18: TS2430: Interface 'ImportClause' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:4071:18: TS2430: Interface 'JSDocTypedefTag' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:4079:18: TS2430: Interface 'JSDocCallbackTag' incorrectly extends interface 'NamedDeclaration'.
  Types of property 'name' are incompatible.
    Type 'Identifier | undefined' is not assignable to type 'DeclarationName'.
      Type 'undefined' is not assignable to type 'DeclarationName'.
src/compiler/types.ts:6659:18: TS2430: Interface 'DeferredTypeReference' incorrectly extends interface 'TypeReference'.
  Types of property 'mapper' are incompatible.
    Type 'TypeMapper | undefined' is not assignable to type 'TypeMapper'.
      Type 'undefined' is not assignable to type 'TypeMapper'.
src/compiler/watch.ts:754:118: TS2375: Type 'ProgramHost<any>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'directoryExists' are incompatible.
    Type '((path: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
      Type 'undefined' is not assignable to type '(path: string) => boolean'.
src/compiler/watchPublic.ts:383:52: TS2379: Argument of type '{ configFileName: string; optionsToExtend: CompilerOptions | undefined; watchOptionsToExtend: WatchOptions; extraFileExtensions: readonly FileExtensionInfo[]; system: System; createProgram: CreateProgram<...> | undefined; reportDiagnostic: DiagnosticReporter | undefined; reportWatchStatus: WatchStatusReporter | unde...' is not assignable to parameter of type 'CreateWatchCompilerHostOfConfigFileInput<T>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'createProgram' are incompatible.
    Type 'CreateProgram<T> | undefined' is not assignable to type 'CreateProgram<T>'.
      Type 'undefined' is not assignable to type 'CreateProgram<T>'.
src/compiler/watchPublic.ts:463:120: TS2379: Argument of type 'WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>' is not assignable to parameter of type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Type 'WatchCompilerHostOfFilesAndCompilerOptions<T> & Partial<WatchCompilerHostOfConfigFile<T>>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'directoryExists' are incompatible.
      Type '((path: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
        Type 'undefined' is not assignable to type '(path: string) => boolean'.
src/compiler/watchPublic.ts:464:11: TS2375: Type 'CachedDirectoryStructureHost | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Type 'WatchCompilerHostOfFilesAndCompilerOptions<T> & Partial<WatchCompilerHostOfConfigFile<T>>' is not assignable to type 'DirectoryStructureHost' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'directoryExists' are incompatible.
      Type '((path: string) => boolean) | undefined' is not assignable to type '(path: string) => boolean'.
        Type 'undefined' is not assignable to type '(path: string) => boolean'.
```

### Initially declined seeds

All 43 initial exclusions follow. An eligible independent write may widen the
same owner and remove an excluded diagnostic incidentally. E022 and E023 are
not seeds even though their shared owners are widened elsewhere.

| Site | Code | Reason |
| --- | --- | --- |
| `src/compiler/builder.ts:1597:36` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:5964:13` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:6035:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:7837:133` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:8082:62` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:10941:33` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:17890:21` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:33923:17` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:43517:20` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:43740:86` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:50153:21` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/emitter.ts:862:55` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/emitter.ts:938:70` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/emitter.ts:1011:33` | TS2379 | expanded chain is not a present-undefined optional value mismatch |
| `src/compiler/emitter.ts:1014:31` | TS2379 | expanded chain is not a present-undefined optional value mismatch |
| `src/compiler/factory/nodeFactory.ts:1216:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:1223:13` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:1315:41` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:1403:41` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:5202:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:5209:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:5496:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:7181:25` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:7422:9` | TS2412 | generic or unresolved receiver contract |
| `src/compiler/moduleNameResolver.ts:132:13` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/moduleNameResolver.ts:627:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/moduleNameResolver.ts:635:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:5094:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:990:43` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1019:43` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1332:58` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1332:58` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1386:9` | TS2412 | generic or unresolved receiver contract |
| `src/compiler/transformers/classFields.ts:2745:16` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/declarations.ts:1668:62` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/declarations/diagnostics.ts:207:50` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/declarations/diagnostics.ts:310:50` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/declarations/diagnostics.ts:424:50` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:689:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:749:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:773:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:568:17` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchUtilities.ts:125:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |

### Remaining optional-code findings

All 36 final exclusions follow; complete expanded chains are emitted in the
script report. Generic specialization, indexed evidence and unrelated chains
remain separate from representation-only corrections.

| Site | Code | Reason |
| --- | --- | --- |
| `src/compiler/checker.ts:5964:13` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:6035:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:7837:133` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:8082:62` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:10941:33` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:17890:21` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:21434:98` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:43517:20` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:43740:86` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/commandLineParser.ts:2663:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/emitter.ts:862:55` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/emitter.ts:938:70` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:1216:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:1315:41` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:1403:41` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:5496:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/moduleNameResolver.ts:132:13` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/moduleNameResolver.ts:627:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/moduleNameResolver.ts:635:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:5105:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:990:43` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1019:43` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1332:58` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1332:58` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1386:9` | TS2412 | generic or unresolved receiver contract |
| `src/compiler/transformers/classFields.ts:2745:16` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:689:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:749:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:773:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watch.ts:754:118` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:371:65` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:383:52` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:463:120` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:464:11` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:568:17` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchUtilities.ts:125:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |

### Upstream build consumer errors

Pristine `npm run build` succeeds. Adapted `npm run build` exits 1 after its
stock build compiler exits 2 with the nine diagnostics below. The checked test
build reports the same nine. The strict source-census options are unchanged.
The error at `tsbuildPublic.ts:787` exposes callable parameter variance previously
hidden by method bivariance. `services.ts:1769` similarly exposes a readonly
array parameter mismatch. The remaining seven service errors follow the
`getCompilerHost` key-presence guard, which now correctly leaves its callable
possibly undefined. Consumer fixes are outside this unit's scope.

```text
src/compiler/tsbuildPublic.ts(787,5): error TS2322: Type 'WriteFileCallback | undefined' is not assignable to type '((path: string, data: string, writeByteOrderMark?: boolean | undefined) => void) | undefined'.
  Type 'WriteFileCallback' is not assignable to type '(path: string, data: string, writeByteOrderMark?: boolean | undefined) => void'.
    Types of parameters 'writeByteOrderMark' and 'writeByteOrderMark' are incompatible.
      Type 'boolean | undefined' is not assignable to type 'boolean'.
        Type 'undefined' is not assignable to type 'boolean'.
src/services/services.ts(1769,13): error TS2322: Type '((typeDirectiveNames: FileReference[] | string[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => (ResolvedTypeReferenceDirective | undefined)[]) | undefined' is not assignable to type '((typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => (ResolvedTypeReferenceDirective | undefined)[]) | undefined'.
  Type '(typeDirectiveNames: FileReference[] | string[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => (ResolvedTypeReferenceDirective | undefined)[]' is not assignable to type '(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => (ResolvedTypeReferenceDirective | undefined)[]'.
    Types of parameters 'typeDirectiveNames' and 'typeReferenceDirectiveNames' are incompatible.
      Type 'string[] | readonly FileReference[]' is not assignable to type 'FileReference[] | string[]'.
        Type 'readonly FileReference[]' is not assignable to type 'FileReference[] | string[]'.
          The type 'readonly FileReference[]' is 'readonly' and cannot be assigned to the mutable type 'string[]'.
src/services/services.ts(1779,39): error TS18048: 'compilerHost' is possibly 'undefined'.
src/services/services.ts(1782,13): error TS2345: Argument of type 'CompilerHost | undefined' is not assignable to parameter of type 'CompilerHostLikeForCache'.
  Type 'undefined' is not assignable to type 'CompilerHostLikeForCache'.
src/services/services.ts(1786,9): error TS18048: 'compilerHost' is possibly 'undefined'.
src/services/services.ts(1788,32): error TS2345: Argument of type 'CompilerHost | undefined' is not assignable to parameter of type 'CompilerHost'.
  Type 'undefined' is not assignable to type 'CompilerHost'.
src/services/services.ts(1796,23): error TS18048: 'compilerHost' is possibly 'undefined'.
src/services/services.ts(1798,20): error TS18048: 'compilerHost' is possibly 'undefined'.
src/services/services.ts(1799,34): error TS18048: 'compilerHost' is possibly 'undefined'.
```

The upstream runtime bundler still produces its compiler/test JavaScript without
these declaration checks. `hereby local --no-typecheck` emits the bundles but
still exits 1 because upstream's service declaration task remains checked.
The baseline suite uses `hereby runtests-parallel --light=false --workers=4
--lint=false --no-typecheck`; this builds the current source bundles and avoids
those checked declaration tasks. This is a runtime oracle build, **not a clean
upstream typecheck**, and is not used for the census. No consumer is hidden from
the diagnostic ledgers and no Adamic checker option is weakened.

### JavaScript identity, idempotence and mutants

All **10 emitted JavaScript files, 28,599,562 bytes**, from upstream's own build
are byte-identical in the direct **20-only** comparison, before adaptation 10
landed, between pristine and adapted trees, including `run.js`, which
bundles the actual compiler into the test harness. Only `.js` output is compared;
declaration output intentionally changes, and source maps/build metadata are
not claimed identical. The source byte comparison uses full bytes rather than
only hashes. SHA-256 values for review follow.

| Emitted file | SHA-256 (both trees) |
| --- | --- |
| `_tsc.js` | `1c59e77a54b186ec43fa7f3e0d3c4bb15ca5eb5ba43e96b1d3a267139eddd3e3` |
| `_tsserver.js` | `1992fc3518a4aa5443638a25794281995ec3ce84820692713296262997d5443b` |
| `_typingsInstaller.js` | `71515c1c1aa29c8616309f7dea4d3a32eed662df3a4dbcf4345e7522ece56ac1` |
| `run.js` | `515554e08cfa2887a10c2d357e3515068788d427af4053e7db016de98350bb94` |
| `tsc.js` | `2cffde0b8c6760dfb0b5b0382bbb7e00ba6a8b2d981b9205b256a700a481d983` |
| `tsserver.js` | `e3ccfeec65ec5c470b8ffc5611878a31182650c7e8a062c38a719f83b523edcb` |
| `tsserverlibrary.js` | `6a0b343fa3a3f53cf198c83a18035c104b6990b2a260272654f9a6b2f2380fdf` |
| `typescript.js` | `569177652966bd528c319171c7dd22860dbf72bde116cbc4f644f1d02bb12e39` |
| `typingsInstaller.js` | `2d48a038523ea7c626fa2796a30fa5fdc07612ea5fab83544423009d7683a213` |
| `watchGuard.js` | `008f2ceb504ee9b8c5d47054d43033e1ec3f2fef03f18d8aa1b2dd70cba6fc88` |

Two runs on one pinned tree both exit 0: 405 declarations changed, then 0.
The scratch fixture independently asserts unchanged source bytes on the second
run, identical emitted JavaScript and Node key fingerprints, three owners for
four write/literal findings, correct callable precedence, an `in`-guard consumer
at its location, and unchanged indexed/generic/wrong-value/live-method contracts.

| Mutant | Named check and observed failure |
| --- | --- |
| Drop `?` only on edited `CompilerOptions.traceResolution` | Census missing-property check: 31 missing-property diagnostics, including TS2741 0 → 24; all 18 directly checked omitted-key literals are caught. Total 2163 → 2195. |
| Omit only that owner edit, restoring `traceResolution?: boolean` | Census optional-count check: TS2375 15 → 16; exact-optional total 36 → 37; total 2163 → 2164. |
| Append a newline to one source file in a copy of the second shared-apply tree | The source-idempotence SHA-256 map comparison rejects the changed copy. Log: `idempotence-mutant.log`. |
| Append a byte-changing statement to a copy of one emitted JS artifact | The same byte-comparison assertion fails. |
| Append `const stage3BaselineMutant: number = "wrong";` to `assignmentCompat1.ts` in an isolated pristine copy | Upstream's filtered baseline oracle exits 3 with three failing comparisons (JS, diagnostics, types/symbols). Restoring only that test input produces 6 passing, exit 0. The deliberately invalid test input is compiled by the test harness; it fails baseline comparison, not the build. |

Mutant trees use identical dependencies. Their initial scratch copies lacked
`source-map-support`, producing an unrelated TS2307; those exploratory counts
were discarded and both census mutants rerun after restoring the dependency.
Only the requested single declaration change distinguishes each final source
mutant. The directly checked literal coverage excludes type assertions (which
bypass construction checks) and keys provided through spreads.

All missing-property mutant locations follow. The one required-property edit
also causes a TS7053 consequence; it is included in the mutant's total above.

| Missing-property site | Code |
| --- | --- |
| `src/compiler/builder.ts:1447:18` | TS2741 |
| `src/compiler/builder.ts:2250:40` | TS2379 |
| `src/compiler/builder.ts:2275:13` | TS2375 |
| `src/compiler/builder.ts:2309:92` | TS2379 |
| `src/compiler/builder.ts:2312:13` | TS2375 |
| `src/compiler/builder.ts:2313:73` | TS2379 |
| `src/compiler/commandLineParser.ts:1858:14` | TS2741 |
| `src/compiler/commandLineParser.ts:2128:5` | TS2322 |
| `src/compiler/commandLineParser.ts:2663:9` | TS2741 |
| `src/compiler/commandLineParser.ts:2691:71` | TS2741 |
| `src/compiler/commandLineParser.ts:2697:79` | TS2741 |
| `src/compiler/commandLineParser.ts:2972:11` | TS2741 |
| `src/compiler/commandLineParser.ts:3064:5` | TS2741 |
| `src/compiler/commandLineParser.ts:3437:41` | TS2741 |
| `src/compiler/commandLineParser.ts:3746:11` | TS2375 |
| `src/compiler/commandLineParser.ts:4264:11` | TS2741 |
| `src/compiler/moduleNameResolver.ts:1755:9` | TS2741 |
| `src/compiler/moduleNameResolver.ts:1800:127` | TS2741 |
| `src/compiler/moduleSpecifiers.ts:1348:55` | TS2741 |
| `src/compiler/moduleSpecifiers.ts:817:96` | TS2741 |
| `src/compiler/transformer.ts:671:31` | TS2741 |
| `src/compiler/utilities.ts:9231:41` | TS2741 |
| `src/compiler/utilities.ts:9237:41` | TS2741 |
| `src/compiler/utilities.ts:9243:41` | TS2741 |
| `src/compiler/utilities.ts:9249:41` | TS2741 |
| `src/compiler/utilities.ts:9255:41` | TS2741 |
| `src/compiler/utilities.ts:9261:41` | TS2741 |
| `src/compiler/utilities.ts:9267:41` | TS2741 |
| `src/compiler/utilities.ts:9280:41` | TS2741 |
| `src/compiler/watchPublic.ts:847:82` | TS2741 |
| `src/compiler/watchPublic.ts:963:17` | TS2741 |

### Reproduction and logs

Toolchain setup was `bash cloud/setup.sh`, then
`source /workspace/adamic-tools/env.sh`. Timing lines: Go ready 1s, clang ready
1s, Node ready 1s, submodules ready 1s, build cache warm 237s, total 237s.
`nproc` is 5; cgroup quota is 4 CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0.
All test output was written directly to logs and read afterward.

The primary unmodified census commands ran every input entry and the whole
program, sequentially, with adaptation 10 already applied:

```sh
go build -o /tmp/adamic-optional/census ./stage3/census/tool > /tmp/adamic-optional/census-build.log 2>&1
/tmp/adamic-optional/census /tmp/adamic-optional/before-20/src/compiler /tmp/adamic-optional/census-before20.jsonl > /tmp/adamic-optional/census-before20.log 2>&1
/tmp/adamic-optional/census /tmp/adamic-optional/pipeline-final/src/compiler /tmp/adamic-optional/census-after20.jsonl > /tmp/adamic-optional/census-after20.log 2>&1
```

For the earlier supplemental after check and the mutant checks, a scratch Go
overlay adds only
`if os.Getenv("CENSUS_WHOLE_ONLY") == "1" { entries = nil }` before the
single-entry loop in `stage3/census/tool/main.go`. Root discovery, `inspect`,
checker options and the final whole-program call are unchanged. The completed
earlier unmodified census's whole-program diagnostics equal the overlay result
exactly. Both primary with-10 passes use the unmodified driver.
The overlay is not committed and is solely a scheduling optimization.

```sh
go build -overlay=/tmp/adamic-optional/census-whole-overlay.json -o /tmp/adamic-optional/census-whole ./stage3/census/tool > /tmp/adamic-optional/census-whole-build.log 2>&1
CENSUS_WHOLE_ONLY=1 /tmp/adamic-optional/census-whole /tmp/adamic-optional/typescript/src/compiler /tmp/adamic-optional/whole-after.jsonl > /tmp/adamic-optional/whole-after.log 2>&1
CENSUS_WHOLE_ONLY=1 /tmp/adamic-optional/census-whole /tmp/adamic-optional/mutant-required/src/compiler /tmp/adamic-optional/whole-required.jsonl > /tmp/adamic-optional/whole-required.log 2>&1
CENSUS_WHOLE_ONLY=1 /tmp/adamic-optional/census-whole /tmp/adamic-optional/mutant-omitted/src/compiler /tmp/adamic-optional/whole-omitted.jsonl > /tmp/adamic-optional/whole-omitted.log 2>&1
```

Supplemental upstream commands were run by hand from each checkout before the
shared pipeline landed. The shared default oracle result is recorded below:

```sh
npm ci --ignore-scripts --no-audit --no-fund > /tmp/adamic-optional/upstream-npm.log 2>&1
npm run build > /tmp/adamic-optional/pristine-build.log 2>&1
npm test -- --workers=4 > /tmp/adamic-optional/pristine-tests.log 2>&1
npm run build > /tmp/adamic-optional/adapted-build-final.log 2>&1
node node_modules/hereby/bin/hereby.js local --no-typecheck > /tmp/adamic-optional/adapted-runtime-build.log 2>&1
node node_modules/hereby/bin/hereby.js tests > /tmp/adamic-optional/adapted-tests-build-checked.log 2>&1
node node_modules/hereby/bin/hereby.js runtests-parallel --light=false --workers=4 --lint=false --no-typecheck > /tmp/adamic-optional/adapted-tests-runtime.log 2>&1
```

Additional checks:

```sh
node --check stage3/adapt/20-optional-declarations/adapt.cjs
node /tmp/adamic-optional/verify-fixture.cjs > /tmp/adamic-optional/verify-fixture.log 2>&1
python3 /tmp/adamic-optional/check-emit.py > /tmp/adamic-optional/emit-check.log 2>&1
python3 /tmp/adamic-optional/check-mutants.py > /tmp/adamic-optional/mutants-check.log 2>&1
node /tmp/adamic-optional/check-required-literals.cjs > /tmp/adamic-optional/required-literals.log 2>&1
go test ./stage3/census/tool ./cmd/adamic-meter ./internal/load -count=1 > /tmp/adamic-optional/package-tests.log 2>&1
```

These checks pass. The census tool has no package tests; meter 9.711s and loader
3.646s pass. The complete Adamic native gate was not run for this source-tooling
unit. No compiler implementation, native fixture, checker option, adaptation 10,
`stage3/apply.sh`, base oracle or source pin was changed.

### Consumers in generated declaration tests

The supplemental APILibCheck test newly reports these **22** generated-declaration
diagnostics. They echo the source-level inheritance incompatibilities already
listed above, at emitted declaration locations. They remain unfixed. The exact
chains are retained in `tests/baselines/local/APILibCheck.errors.txt` in the scratch
adapted tree; none of its reference baselines was changed.

```text
typescript.d.ts(4647,15): error TS2430: Interface 'FunctionDeclaration' incorrectly extends interface 'DeclarationStatement'.
typescript.d.ts(4647,15): error TS2430: Interface 'FunctionDeclaration' incorrectly extends interface 'FunctionLikeDeclarationBase'.
typescript.d.ts(4992,15): error TS2430: Interface 'FunctionExpression' incorrectly extends interface 'FunctionLikeDeclarationBase'.
typescript.d.ts(5415,15): error TS2430: Interface 'ClassLikeDeclarationBase' incorrectly extends interface 'NamedDeclaration'.
typescript.d.ts(5422,15): error TS2430: Interface 'ClassDeclaration' incorrectly extends interface 'DeclarationStatement'.
typescript.d.ts(5530,15): error TS2430: Interface 'ImportClause' incorrectly extends interface 'NamedDeclaration'.
typescript.d.ts(5840,15): error TS2430: Interface 'JSDocTypedefTag' incorrectly extends interface 'NamedDeclaration'.
typescript.d.ts(5847,15): error TS2430: Interface 'JSDocCallbackTag' incorrectly extends interface 'NamedDeclaration'.
typescript.internal.d.ts(9436,15): error TS2430: Interface 'FunctionDeclaration' incorrectly extends interface 'DeclarationStatement'.
typescript.internal.d.ts(9436,15): error TS2430: Interface 'FunctionDeclaration' incorrectly extends interface 'FunctionLikeDeclarationBase'.
typescript.internal.d.ts(9823,15): error TS2430: Interface 'FunctionExpression' incorrectly extends interface 'FunctionLikeDeclarationBase'.
typescript.internal.d.ts(10390,15): error TS2430: Interface 'ClassLikeDeclarationBase' incorrectly extends interface 'NamedDeclaration'.
typescript.internal.d.ts(10397,15): error TS2430: Interface 'ClassDeclaration' incorrectly extends interface 'DeclarationStatement'.
typescript.internal.d.ts(10509,15): error TS2430: Interface 'ImportClause' incorrectly extends interface 'NamedDeclaration'.
typescript.internal.d.ts(10820,15): error TS2430: Interface 'JSDocTypedefTag' incorrectly extends interface 'NamedDeclaration'.
typescript.internal.d.ts(10827,15): error TS2430: Interface 'JSDocCallbackTag' incorrectly extends interface 'NamedDeclaration'.
typescript.internal.d.ts(12999,15): error TS2430: Interface 'DeferredTypeReference' incorrectly extends interface 'TypeReference'.
typescript.internal.d.ts(23491,15): error TS2320: Interface 'TypeReferenceDirectiveResolutionCache' cannot simultaneously extend types 'PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>' and 'NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>'.
typescript.internal.d.ts(23534,15): error TS2320: Interface 'ModuleResolutionCache' cannot simultaneously extend types 'PerDirectoryResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'NonRelativeModuleNameResolutionCache'.
typescript.internal.d.ts(23543,15): error TS2320: Interface 'NonRelativeModuleNameResolutionCache' cannot simultaneously extend types 'NonRelativeNameResolutionCache<ResolvedModuleWithFailedLookupLocations>' and 'PackageJsonInfoCache'.
typescript.internal.d.ts(25095,15): error TS2430: Interface 'BuilderProgramState' incorrectly extends interface 'ReusableBuilderProgramState'.
typescript.internal.d.ts(25518,15): error TS2320: Interface 'CachedResolvedModuleWithFailedLookupLocations' cannot simultaneously extend types 'ResolvedModuleWithFailedLookupLocations' and 'ResolutionWithFailedLookupLocations'.
```

### Exact baseline equality is not established

The pristine hand-run full suite passed **106,367 tests**. The adapted supplemental
full runtime suite exited 1: **106,365 passing, 2 failing**. Its mismatches are
`tests/baselines/local/APILibCheck.errors.txt` and
`tests/baselines/local/api/typescript.d.ts`. The latter deliberately snapshots
public declaration text; truthful optional unions therefore change its expected
output even though every emitted JavaScript byte is identical. APILibCheck tests
the generated public declarations too. References were not updated. These are
observed counterexamples to the claim that type-only edits trivially preserve
every upstream baseline.

The ordinary checked build also fails on the nine consumer diagnostics above.
Consumer repairs and acceptance of changed declaration baselines are outside this
unit, so a successful default upstream oracle cannot be claimed.

### Shared pipeline integration

Merged base pipeline `8728405` and type imports `396f87e` after census `429c117`.
The adaptation script was used unchanged by `stage3/apply.sh` in numeric order.
The input for the primary census table is setup plus adaptation 10; the output
adds adaptation 20. Heavy apply, oracle and census runs were sequential.

The default shared oracle command was:

```sh
stage3/apply.sh /tmp/adamic-optional/pipeline > /tmp/adamic-optional/pipeline-apply.log 2>&1
stage3/oracle/run.sh /tmp/adamic-optional/pipeline /tmp/adamic-optional/pipeline-oracle > /tmp/adamic-optional/pipeline-oracle.log 2>&1
```

Apply passed. The unchanged oracle used `runners=all`, no regex, four workers and
its default limit. Install passed (3.602s), build failed (20.482s), oracle exited
1 after 24.100s. Its `report.json` has `status=fail` and no test phase: the checked
build prevents entry into the default suites. This is the primary oracle result;
the supplemental runtime suite above does not turn it into a pass.

Upstream build regenerates the diagnostics source, replacing its adapted
optional owner and type import. A fresh second apply output, `pipeline-final`,
therefore supplies the census and idempotence input. A second adapter run on
that output changed zero declarations and preserved all **745 source files**
byte-for-byte. The two apply runs selected exactly the same 405 owners and
identical evidence/consumer/decline ledgers as the original direct run.

The generated patch scoreboard reports setup 0; adaptation 10 **72 files,
3,719 lines replaced**; adaptation 20 **26 files, 406 lines replaced**; combined
**73 files, 4,125 lines replaced**. The generated scoreboard is preserved in
`/tmp/adamic-optional/pipeline-patch-set.md`; this unit commits only its adapter
and README.

Both primary censuses ran the unchanged full driver: 81 individual input
attempts followed by all 78 TypeScript roots together. Both exited 0 with a
checker-rejected whole program. The three JSON inputs fail extension loading,
as expected, and are outside the TypeScript root population. Every non-import
diagnostic chain is byte-identical, after removing the tree prefix, to the
earlier 20-only before/after reports; the 3,719 TS1484 findings are removed by
adaptation 10. Thus the complete consumer list above is also verified on 10.

```sh
stage3/apply.sh /tmp/adamic-optional/pipeline-final > /tmp/adamic-optional/pipeline-final-apply.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/adapt/20-optional-declarations/adapt.cjs /tmp/adamic-optional/pipeline-final > /tmp/adamic-optional/pipeline-idempotence.log 2>&1
# before-20 is pristine plus setup and adaptation 10, with the same locked dependencies.
/tmp/adamic-optional/census /tmp/adamic-optional/before-20/src/compiler /tmp/adamic-optional/census-before20.jsonl > /tmp/adamic-optional/census-before20.log 2>&1
/tmp/adamic-optional/census /tmp/adamic-optional/pipeline-final/src/compiler /tmp/adamic-optional/census-after20.jsonl > /tmp/adamic-optional/census-after20.log 2>&1
```

The full-driver census summed durations were **212.840s before**, **200.622s
after**, including all per-entry attempts and the final whole-program call.
These are measured tool durations, not estimates of native compilation time.

Both declaration mutants were rerun on the final with-10 source, using the same
unchanged whole-program gate through the scheduling overlay. The required-key
mutant totals **2,195** diagnostics, versus **2,163**: TS2322 +1, TS2345 +1,
TS2375 +2, TS2379 +3, TS2741 +24, TS7053 +1; every other reason is unchanged.
There are **31** missing-property chains, covering all **18** directly checked
omitted-key literals. The omitted-owner mutant totals **2,164**, with only
TS2375 +1 (15 → 16). Logs: `mutants10-check.log`, `required-literals10.log`;
raw records: `whole10-required.jsonl`, `whole10-omitted.jsonl`.

The previous-adaptation control (`before-20`) then passed upstream's ordinary
`npm run build`. Its **eight emitted compiler JavaScript artifacts, 15,424,668
bytes**, are byte-identical to every `.js` artifact emitted by the default
shared oracle's 10+20 compiler build, despite that build's checked failure.
Log: `pipeline-emit-check.log`. The earlier direct 20-only proof additionally
compares all ten artifacts, including the test harness. Overall unadapted-to-
10+20 bundle bytes differ in esbuild's allocated local names because of
adaptation 10; this unit's before/after comparison holds 10 fixed.

All supplemental fixture, byte-mutant, source-idempotence-mutant, baseline-mutant
and package results above remain applicable to the identical adapter. A successful
default oracle and exact reference-baseline equality remain **failed obligations**,
not waived checks. No consumers or reference baselines were repaired.
