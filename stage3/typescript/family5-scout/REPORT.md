# Family 5 optional-property scout — #pncjpqy

## Result

The regenerated **first-error tsc-entry** pass exactly reproduces the Nov 18 record: **131 distinct source positions / 131 meter sites / 45 reasons**, with every reason count identical. The two largest groups are **26 TextRange → SourceMapRange / source** and **14 SourceFile → SourceFileLike / getPositionOfLineAndCharacter**. There are 131 raw family events, all body findings; no extra tail events.

| Road | Groups | Current sites |
| --- | --- | --- |
| (i) Truthful declaration / construction-result annotation | 33 | 109 |
| (ii) Narrowed rule, subject to the stated proof premise | 6 | 8 |
| (iii) Real hidden-field hazard; stays refused | 6 | 14 |

These counts classify **the first named fact**, not the number of units a change would admit. Candidates were not installed in apply or measured with modified lowering. A later optional property or another relation rule can still stop. In particular, Type → TypeParameter is not certified as a whole, and serialized compiler options are not thereby certified as the internal CompilerOptions shape.

Concrete counterexamples come from input objects: package.json retains numeric `name` / `version`; accepted .tsbuildinfo retains a numeric `repopulateInfo`; and a cached related-information record can carry **`reportsUnnecessary: null`**, which both original spread constructors preserve. The last case rejects two declaration candidates **despite 0 stock diagnostics**.

## Rule, scope and reproduction

Read first: [rulings.md at c56b0229, lines 71/72/75](https://github.com/system-inc/adamic/blob/c56b022995e1e960bad71626fd2074bc46aee485/docs/rulings.md#L71); [full retained entries](evidence/rulings-71-72-75.txt). Line 71 refuses every relation position unless the source declaration names the optional field, with the whole-program nominal-class and fresh-literal exemptions; a spread does not prove absence. Line 72 gives the eventual checked-read/write end state, not a current blanket admission. Line 75 makes a false never promise a soundness problem, so `P?: never` below is used only for an audited constructor result that never acquires P, not to wish away an existing cache field.

The words are built by [optional_widening.go:400](https://github.com/system-inc/adamic/blob/c56b022995e1e960bad71626fd2074bc46aee485/internal/lower/optional_widening.go#L400). `optionalWidened` checks assignable pairs, unions/constraints, container elements, properties and callback return/parameter positions. A missing optional member is refused unless `classOptionalAbsent` proves the whole possible class hierarchy omits it. `optionalValue` exempts the literal at its own position and follows conditional branches; `optionalAtSite` separately judges a spread source, skipping only keys explicitly overwritten after it. Assignment handles do not borrow RHS freshness. Consequently, a table entry can be an outer argument/assignment whose offending triple is nested in its signature or element type.

There is no whole group here whose only premise is “every runtime value is a nominal class instance.” Node/Type allocation uses function constructors behind program interfaces; some services SourceFiles are classes, but that does not make every SourceFile nominal. Roads (ii) below are specific origin/key-set certificates, not a proposal to erase checks on arbitrary aliases or spreads.

Main is **c56b022995e1e960bad71626fd2074bc46aee485**. The report branch is based directly on that commit. The existing uninstrumented own-SHA entry tool and the same c56 shared apply product used by census-2 were reused after verifying the binary hash and all 81 source hashes. Its TypeScript input is v6.0.3 **050880ce59e30b356b686bd3144efe24f875ebc8**, with submodules initialized at their pinned gitlinks. [Provenance](evidence/provenance.json), [original preparation logs](evidence/preparation/), [source manifest](evidence/adapted-sources.sha256).

The exact run uses `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_TIMES=/dev/stderr`, **no LATENT_ONLY_FILE**, rooted at `/workspace/depthgate2-adapted/src/tsc/tsc.ts`. It completed normally in **2181.948 s**, peak RSS **3,720,952 KiB**, with 0 checker diagnostics and a normal relation_report trailer. All 81 entry-closure source files are covered. Body counting is exactly the meter’s unique `(kind, where, reason, text)` key; tail events were retained and counted separately. [Commands](evidence/commands.txt), [full stream/log/status](evidence/run/), [all family events](evidence/family-events.jsonl), [summary](evidence/run-summary.json), [record report](evidence/record-report.json.gz).

The construction audit covers the compiled tsc source/factory set and real file inputs; it does not claim to audit arbitrary foreign JavaScript clients. Declaration roads use the [adaptation 72 certificate pattern](https://github.com/system-inc/adamic/blob/c56b022995e1e960bad71626fd2074bc46aee485/stage3/adapt/72-readonly-diagnostic-slots/README.md): state an owned, truthful contract and let the compiled program judge every relation/write at its own site. Existing unsafe hatches are not certified by a zero stock count. The [named-write/declaration audit](evidence/write-audit.json) enumerates 201 declarations and 387 writes/initializers, explicitly excluding a claim that dynamic keys or spreads prove absence; those constructors are traced separately below.

## All 45 groups

Coordinates are on the unchanged adapted tree, relative to `src/compiler/` unless prefixed otherwise. IDs G01–G45 follow the record’s reason order. A = argument, W = assignment (including initialization), R = return, O = object-literal property, S = spread. These are **outer syntax positions**; callback results / array elements may contain the actual S→T pair. The two pragma target names are abbreviated by the compiler’s own reason formatter; their full schema is the reference-argument intersection in `types.ts:10342` onward. Literal reason strings and full messages are preserved in the evidence.

| Group | S | T | P | Sites | Positions | Road | Construction / evidence |
| --- | --- | --- | --- | --- | --- | --- | --- |
| G01 | `TextRange` | `SourceMapRange` | `source` | 26 | A:25, W:1 | (i) C01b | Node/Token/Identifier constructors in utilities.ts:8507,8521,8533; baseNodeFactory.ts:41 and node factories; range helpers utilities.ts:7869,7890 return either the original range or fresh pos/end objects |
| G02 | `SourceFile` | `SourceFileLike` | `getPositionOfLineAndCharacter` | 14 | A:13, W:1 | (i) C02 | Compiler SourceFiles are allocated by baseNodeFactory.ts:41 via the Node constructor (utilities.ts:8507) and initialized by parser.ts:1951 |
| G03 | `Type` | `TypeParameter` | `constraint` | 12 | A:10, W:1, R:1 | (i) C03 | checker.ts:5521/5575 constructs Type/TypeParameter through the allocator Type constructor (utilities.ts:8491); TypeParameter, substitution and this-type constraint caches store Type-valued constraints (checker.ts:16855-16869,17258,17927) |
| G04 | `Identifier` | `SourceMapRange` | `source` | 7 | A:7 | (i) C01b | `factory/nodeFactory.ts:1326:5` → base allocation |
| G05 | `{ errors?: Diagnostic[] \| undefined; }` | `ErrorOutputContainer` | `skipLogging` | 6 | A:6 | (i) C08 | checkTypeAssignableTo (checker.ts:21354) and four resultObj annotations erase the skipLogging field of ErrorOutputContainer; values are forwarded containers or fresh {} |
| G06 | `Expression` | `SourceMapRange` | `source` | 6 | A:6 | (i) C01b | Node/Token/Identifier constructors in utilities.ts:8507,8521,8533; baseNodeFactory.ts:41 and node factories; range helpers utilities.ts:7869,7890 return either the original range or fresh pos/end objects |
| G07 | `PackageJsonPathFields` | `PackageJson` | `version` | 5 | A:5 | (iii) | Concrete stock TypeScript inputs: package.json {name:42,version:"1.0.0",types:"index.d.ts"} remains object with numeric name after readJson; {name:"pkg",version:7,types:"index.d.ts"} remains cached PackageJsonPathFields with numeric version |
| G08 | `DiagnosticRelatedInformation` | `DiagnosticReadView` | `reportsUnnecessary` | 4 | A:4 | (iii) | A valid ReusableDiagnosticRelatedInformation record in an accepted build-info relatedInformation array carries extra reportsUnnecessary:null |
| G09 | `SolutionBuilderHost<T>` | `ModuleResolutionHost` | `getGlobalTypingsCacheLocation` | 3 | A:3 | (i) C09b | createSolutionBuilderHostBase (tsbuildPublic.ts:294) obtains createProgramHost and adds known reporting functions |
| G10 | `{ preserve: true; }` | `FileReference` | `resolutionMode` | 3 | S:3 | (ii) | parser.ts:10624,10627,10630 spreads only conditional branches {preserve} / {} and {resolutionMode:parsed} / {} |
| G11 | `BigIntLiteral` | `SourceMapRange` | `source` | 3 | A:3 | (i) C01b | `factory/nodeFactory.ts:1244:5` → base allocation |
| G12 | `InterfaceType` | `TypeParameter` | `constraint` | 2 | W:2 | (i) C03 | checker.ts:5521/5575 constructs Type/TypeParameter through the allocator Type constructor (utilities.ts:8491); TypeParameter, substitution and this-type constraint caches store Type-valued constraints (checker.ts:16855-16869,17258,17927) |
| G13 | `{ resolvedModule: ResolvedModuleFull; }` | `ResolvedModuleWithFailedLookupLocations` | `failedLookupLocations` | 2 | W:1, R:1 | (i) C15 | program.ts:1615-1628 and1653-1668 creates new outer resolution wrappers in Array.map callbacks |
| G14 | `ResolvedModuleWithFailedLookupLocations` | `CachedResolvedModuleWithFailedLookupLocations` | `isInvalidated` | 2 | W:1, O:1 | (i) C06 | Resolution wrappers come from module resolver/host result constructors and cache maps; resolutionCache.ts:1482 writes isInvalidated=true on reused resolution objects |
| G15 | `object` | `PackageJson` | `name` | 2 | W:2 | (iii) | Concrete stock TypeScript inputs: package.json {name:42,version:"1.0.0",types:"index.d.ts"} remains object with numeric name after readJson; {name:"pkg",version:7,types:"index.d.ts"} remains cached PackageJsonPathFields with numeric version |
| G16 | `Pick<SourceFile, "fileName" \| "impliedNodeFormat">` | `Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">` | `packageJsonScope` | 2 | A:2 | (i) C12b | getModuleSpecifierPreferences importingSourceFile (moduleSpecifiers.ts:186) is a Pick that erases packageJsonScope |
| G17 | `Node` | `SourceMapRange` | `source` | 2 | A:1, R:1 | (i) C01b | Node/Token/Identifier constructors in utilities.ts:8507,8521,8533; baseNodeFactory.ts:41 and node factories; range helpers utilities.ts:7869,7890 return either the original range or fresh pos/end objects |
| G18 | `VariableDeclaration` | `SourceMapRange` | `source` | 2 | A:2 | (i) C01b | `factory/nodeFactory.ts:4251:5` → base allocation |
| G19 | `VariableStatement` | `SourceMapRange` | `source` | 2 | A:2 | (i) C01b | `factory/nodeFactory.ts:3876:5` → base allocation |
| G20 | `object` | `CompilerOptions` | `all` | 1 | S:1 | (ii) | optionMapToObject (commandLineParser.ts:2729) calls Object.fromEntries |
| G21 | `IncrementalCompilationOptions` | `IncrementalProgramOptions<EmitAndSemanticDiagnosticsBuilderProgram>` | `createProgram` | 1 | A:1 | (i) C11 | performIncrementalCompilation (watch.ts:987) forwards its input into createIncrementalProgram |
| G22 | `{ resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective \| undefined; }` | `ResolvedTypeReferenceDirectiveWithFailedLookupLocations` | `failedLookupLocations` | 1 | R:1 | (i) C15 | program.ts:1615-1628 and1653-1668 creates new outer resolution wrappers in Array.map callbacks |
| G23 | `ResolvedTypeReferenceDirectiveWithFailedLookupLocations` | `CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations` | `isInvalidated` | 1 | O:1 | (i) C06 | Resolution wrappers come from module resolver/host result constructors and cache maps; resolutionCache.ts:1482 writes isInvalidated=true on reused resolution objects |
| G24 | `ResolvedModuleWithFailedLookupLocations` | `ResolutionWithFailedLookupLocations` | `isInvalidated` | 1 | R:1 | (i) C06 | Resolution wrappers come from module resolver/host result constructors and cache maps; resolutionCache.ts:1482 writes isInvalidated=true on reused resolution objects |
| G25 | `ResolvedTypeReferenceDirectiveWithFailedLookupLocations` | `ResolutionWithFailedLookupLocations` | `isInvalidated` | 1 | R:1 | (i) C06 | Resolution wrappers come from module resolver/host result constructors and cache maps; resolutionCache.ts:1482 writes isInvalidated=true on reused resolution objects |
| G26 | `JSDoc[]` | `JSDocArray` | `jsDocCache` | 1 | W:1 | (ii) | withJSDoc (parser.ts:1820-1834) calls mapDefined |
| G27 | `CompilerHost` | `CompilerHost & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }` | `onUnRecoverableConfigFileDiagnostic` | 1 | A:1 | (i) C10 | CompilerHost is created by program.ts createCompilerHostWorker; watch/config helpers install DiagnosticReporter-compatible callbacks or omit them |
| G28 | `SolutionBuilderHost<T>` | `ProgramHost<T> & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; }` | `onUnRecoverableConfigFileDiagnostic` | 1 | A:1 | (i) C09b | createSolutionBuilderHostBase (tsbuildPublic.ts:294) obtains createProgramHost and adds known reporting functions |
| G29 | `{ path: string; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; }` | `Resolved` | `originalPath` | 1 | R:1 | (ii) | withPackageId (moduleNameResolver.ts:127-140) returns r && a fresh literal containing exactly path,extension,packageId,resolvedUsingTsExtension |
| G30 | `{ moduleFileToTry: string; }` | `{ moduleFileToTry: string; packageRootPath?: string; blockedByExports?: true; verbatimFromExports?: true; }` | `packageRootPath` | 1 | S:1 | (i) C16 | tryGetModuleNameFromExportsOrImports (moduleSpecifiers.ts:1020) and tryGetModuleNameFromExports (:1122) return fresh objects with exactly moduleFileToTry, or recursive results of the same constructors |
| G31 | `{ resolutionMode: ModuleKind.CommonJS \| ModuleKind.ESNext; }` | `FileReference` | `preserve` | 1 | S:1 | (ii) | parser.ts:10624,10627,10630 spreads only conditional branches {preserve} / {} and {resolutionMode:parsed} / {} |
| G32 | `SerializedDiagnosticMessageChain` | `DiagnosticMessageChain` | `repopulateInfo` | 1 | S:1 | (iii) | A version-matched accepted .tsbuildinfo has a SerializedDiagnosticMessageChain with valid messageText/category/code and an undeclared repopulateInfo:42 |
| G33 | `DiagnosticRelatedInformation` | `Diagnostic` | `reportsUnnecessary` | 1 | W:1 | (iii) | A valid ReusableDiagnosticRelatedInformation record in an accepted build-info relatedInformation array carries extra reportsUnnecessary:null |
| G34 | `ReusableDiagnosticRelatedInformation` | `ReusableDiagnostic` | `reportsUnnecessary` | 1 | W:1 | (iii) | A valid ReusableDiagnosticRelatedInformation record in an accepted build-info relatedInformation array carries extra reportsUnnecessary:null |
| G35 | `SourceFile` | `SourceMapSource` | `skipTrivia` | 1 | A:1 | (i) C07 | SourceFile constructors and services SourceFileObject omit skipTrivia |
| G36 | `ArrayBindingPattern` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:2799:5` → base allocation |
| G37 | `ClassDeclaration` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:4383:5` → base allocation |
| G38 | `EnumDeclaration` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:4510:5` → base allocation |
| G39 | `InitializedVariableDeclaration` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:4251:5` → base allocation |
| G40 | `ModuleDeclaration` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:4544:5` → base allocation |
| G41 | `ParameterDeclaration` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:1645:5` → base allocation |
| G42 | `SuperExpression` | `SourceMapRange` | `source` | 1 | A:1 | (i) C01b | `factory/nodeFactory.ts:1521:5` → base allocation |
| G43 | `{ [index: string]: string \| undefined; }` | `{ types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ...` | `types` | 1 | O:1 | (i) C13 | getNamedPragmaArguments builds {} and writes keys from pragma.args |
| G44 | `{ [index: string]: string \| { value: string; pos: number; end: number; }; }` | `{ types?: { value: string; pos: number; end: number; }; } & { lib?: { value: string; pos: number; end: number; }; } & { path?: { value: string; pos: number; end: number; }; } & { "no-default-lib"?: string; } & { "resolution-mode"?: string; } & ...` | `types` | 1 | O:1 | (i) C14b | extractPragmas builds a fresh argument dictionary |
| G45 | `{ include: readonly string[] \| undefined; exclude: readonly string[] \| undefined; }` | `TSConfig & { watchOptions?: object \| undefined; }` | `watchOptions` | 1 | S:1 | (ii) | convertToTSConfig (commandLineParser.ts:2694-2697) spreads a conditional whose only constructors are {include,exclude} and {} |

Detailed constructor paths, the absence/compatibility premise, and each group’s site IDs are also retained in [group-decisions.json](evidence/group-decisions.json). [Whole constructor bodies](evidence/constructors.md) and [exact node/function contexts](evidence/site-contexts.json.gz) make the shared traces reviewable.

## Declaration roads and stock checks

Each candidate was applied **independently to a scratch src copy**, checked with **stock TypeScript 6.0.3 over the 79 compiler project roots**, then reset. The frozen census tree never changed. Both configurations are retained: (a) inherited compiler tsconfig + noEmit and pinned bundled Node types; (b) loader-strict flags including noUncheckedIndexedAccess, exactOptionalPropertyTypes and strict/useUnknownInCatchVariables, ES2024/ESNext/Bundler, no prelude, pinned Node declarations. Both unchanged baselines have **0 diagnostics**. Exact roots/options and every diagnostic are in [stock evidence](evidence/stock/); [checker script](evidence/stock.cjs). The initial dependency-resolution attempt without the pinned Node type root was corrected before these baselines/candidates.

All **14 selected candidates** have **0 → 0 diagnostics in both configurations**, so no stock diagnostic is new, gone, changed or relocated. Every changed file also has **byte-identical JavaScript** under stock transpileModule (ES2020/ESNext). They are syntactically type-only; this is not a release build/API emit or native-emission claim. Candidate source files and edit-applying scripts are **not committed**. The proposed declarations and results are review metadata, not landed adaptations. [Selected results/hashes](evidence/selected-candidates.json).

| Candidate | Groups | Original declaration/view location | Proposed type-only contract | Published API | Project / strict diagnostics | Changed-file JS |
| --- | --- | --- | --- | --- | --- | --- |
| C01b | G01, G04, G06, G11, G17, G18, G19, G36, G37, G38, G39, G40, G41, G42 | `types.ts:28`<br>`factory/nodeFactory.ts:7182`<br>`factory/nodeFactory.ts:7183` | `source?: SourceMapSource;`<br>`setOriginal(updated as T, original);`<br>`setTextRange(updated as T, original);` | public | 0 → 0 / 0 → 0 | identical |
| C02 | G02 | `types.ts:942` | `/** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) \| undefined;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C03 | G03, G12 | `types.ts:6450` | `/** @internal */ constraint?: Type;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C06 | G14, G23, G24, G25 | `types.ts:8118`<br>`types.ts:8150` | `/** @internal */ isInvalidated?: boolean;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C07 | G35 | `types.ts:4326` | `/** @internal */ skipTrivia?: (pos: number) => number;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C08 | G05 | `checker.ts:21354`<br>`checker.ts:21449`<br>`checker.ts:21497`<br>`checker.ts:21578`<br>`checker.ts:21668` | `{ errors?: Diagnostic[] \| undefined; } → ErrorOutputContainer` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C09b | G09, G28 | `tsbuildPublic.ts:215` | `/** @internal */ getGlobalTypingsCacheLocation?: () => string \| undefined;`<br>`/** @internal */ onUnRecoverableConfigFileDiagnostic?: (diagnostic: Diagnostic) => void;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C10 | G27 | `types.ts:8164` | `/** @internal */ onUnRecoverableConfigFileDiagnostic?: (diagnostic: Diagnostic) => void;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C11 | G21 | `watch.ts:974` | `createProgram?: import("./watchPublic.js").CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram>;` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C12b | G16 | `moduleSpecifiers.ts:186` | `importingSourceFile: Pick<SourceFile, "fileName" \| "impliedNodeFormat"> → importingSourceFile: Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C13 | G43 | `parser.ts:10764`<br>`parser.ts:10768` | `{ [index: string]: string \| undefined; } → { [index: string]: string \| undefined; types?: never; }` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C14b | G44 | `parser.ts:10706` | `{ [index: string]: string \| { value: string; pos: number; end: number; }; } → { [index: string]: string \| { value: string; pos: number; end: number; } \| undefined; types?: { value: string; pos: number; end: number; }; }` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C15 | G13, G22 | `program.ts:1623`<br>`program.ts:1667` | `).map(resolved => → ).map<ResolvedModuleWithFailedLookupLocations>(resolved =>; ).map(resolvedTypeReferenceDirective => → ).map<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>(resolvedTypeReferenceDirective =>` | internal / @internal member | 0 → 0 / 0 → 0 | identical |
| C16 | G30 | `moduleSpecifiers.ts:1031`<br>`moduleSpecifiers.ts:1122` | `{ moduleFileToTry: string; } → { moduleFileToTry: string; packageRootPath?: never; }` | internal / @internal member | 0 → 0 / 0 → 0 | identical |

Only C01b adds a **published** member (`TextRange.source`); a structural client whose pos/end object carries source of another type would newly fail assignment. It would require the public sanction/review work analogous to adaptation 72 before landing. The other selected changes are private annotations, internal interfaces, or explicitly @internal members of public owners; no published addition is proposed for them. Services’ existing public SourceFile method remains the reason to put C02’s compiler-only optional declaration on Node. No sanction or API file is changed here.

Measured rejected refinements matter: **C01 alone** has 0 project diagnostics but **2 strict TS2379** at nodeFactory.ts:7182–7183; C01b adds the two erased `updated as T` views, recovering the already returned `T extends Node`, and gets 0. **C09** on ProgramHost causes **TS2320**; C09b places the members on SolutionBuilderHostBase and gets 0. **C14** omits undefined from the index value and causes **TS2411**; C14b includes it and gets 0. **C04/C05** (diagnostic marker declarations) have **0 diagnostics and identical JS**, but are **rejected as untruthful** by the cache witness, so they are not declaration roads. All original diagnostics/messages remain in the result files.

### C01b — G01, G04, G06, G11, G17, G18, G19, G36, G37, G38, G39, G40, G41, G42

Node/Token/Identifier constructors in utilities.ts:8507,8521,8533; baseNodeFactory.ts:41 and node factories; range helpers utilities.ts:7869,7890 return either the original range or fresh pos/end objects. Named-write/declaration audit finds no TextRange-compatible source field with an incompatible type. Other source:string/Type fields belong to non-TextRange owners. SourceMapRange itself can carry a compatible SourceMapSource (customTransforms.ts:154 is an existing positive API fixture), so never is not the proposed contract.

Owned constructor/write facts plus adaptation 72's relation-site certificate. The declaration alone caused two strict TS2379 errors at nodeFactory.ts:7182-7183; two erased updated as T views in the existing update<T> helper recover the already returned T extends Node and remove them. Foreign client compatibility is a public contract consequence, not silently assumed.

### C02 — G02

Compiler SourceFiles are allocated by baseNodeFactory.ts:41 via the Node constructor (utilities.ts:8507) and initialized by parser.ts:1951. They omit the method; services SourceFileObject (services/services.ts:1067,1133) implements the matching callable. scanner.ts:468 uses it only if present. No incompatible constructor or write found.

Put the internal optional method on Node (types.ts:942), inherited by compiler SourceFile. This avoids optional/required merging on SourceFile, whose public services augmentation already declares a required method at services/types.ts:172.

### C03 — G03, G12

checker.ts:5521/5575 constructs Type/TypeParameter through the allocator Type constructor (utilities.ts:8491); TypeParameter, substitution and this-type constraint caches store Type-valued constraints (checker.ts:16855-16869,17258,17927). AST TypeParameterDeclaration.constraint is a different Node/TypeNode field, not a Type-object counterexample.

Add the compatible internal cache declaration on Type, inherited by InterfaceType; never would be false for real TypeParameter/substitution instances.

### C06 — G14, G23, G24, G25

Resolution wrappers come from module resolver/host result constructors and cache maps; resolutionCache.ts:1482 writes isInvalidated=true on reused resolution objects. The named-write audit finds that Boolean write, no conflicting declaration/value.

Both public result-interface owners gain an @internal Boolean slot; never is false because cached values really acquire true.

### C07 — G35

SourceFile constructors and services SourceFileObject omit skipTrivia. The distinct SourceMapSource constructor (utilities.ts:8546-8550 and services/services.ts:1321) writes only the matching (pos:number)=>number callback. No SourceFile constructor/writer supplies another type.

Declare a compatible optional internal callback on SourceFile; retain support for a legitimate SourceFile/source-map intersection, rather than globally assert never.

### C08 — G05

checkTypeAssignableTo (checker.ts:21354) and four resultObj annotations erase the skipLogging field of ErrorOutputContainer; values are forwarded containers or fresh {}. Real containers at checker.ts:36175 and45952 use skipLogging:true. All named writes/initializers are Boolean-compatible.

Name ErrorOutputContainer in the existing owned annotations; preserve the actual Boolean flag rather than claim absence.

### C09b — G09, G28

createSolutionBuilderHostBase (tsbuildPublic.ts:294) obtains createProgramHost and adds known reporting functions. The default objects omit getGlobalTypingsCacheLocation; installed getters/reporters and parse-config callback writes are callable with the stated signature (program.ts:1966; tsbuildPublic.ts:596,598).

Add internal optional contracts to SolutionBuilderHostBase, not ProgramHost. The broader ProgramHost candidate caused TS2320 through ConfigFileDiagnosticsReporter inheritance and is rejected.

### C10 — G27

CompilerHost is created by program.ts createCompilerHostWorker; watch/config helpers install DiagnosticReporter-compatible callbacks or omit them. program.ts:5131 uses the optional callback with a returnUndefined fallback. Named callback writes are listed in the audit.

Add an @internal optional callable to CompilerHost; constructors are not nominal-class-only, so the nominal exemption is not the premise.

### C11 — G21

performIncrementalCompilation (watch.ts:987) forwards its input into createIncrementalProgram. CLI/controller constructors omit createProgram; supplied overrides are the same CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram> factory contract used by watchPublic.ts:132-151.

Declare the optional factory on the internal IncrementalCompilationOptions owner, not never.

### C12b — G16

getModuleSpecifierPreferences importingSourceFile (moduleSpecifiers.ts:186) is a Pick that erases packageJsonScope. The value paths are SourceFile/FutureSourceFile metadata; types.ts:4305 already gives FutureSourceFile its compatible PackageJsonInfo scope. Resolver construction and implied-format setup supply that same cache object, or omit it.

Add packageJsonScope to only this parameter Pick. Both sites (:200 and :243) read this parameter; no unrelated Pick is needed.

### C13 — G43

getNamedPragmaArguments builds {} and writes keys from pragma.args. addPragmaForMatch checks the selected kind; commentPragmas (types.ts:10273-10316) gives positional SingleLine/MultiLine pragmas no types key. reference has types only in the XML kind.

On the private return and argMap annotations, types?:never truthfully states absence for all current positional constructors. This relies on the immutable current pragma-spec table/kind check, not on a generic string index signature proving absence.

### C14b — G44

extractPragmas builds a fresh argument dictionary. For reference/types the fixed metadata has captureSpan:true, and parser.ts:10714-10720 writes the value/pos/end object. Other XML pragmas do not construct a types key.

Declare types?:{value:string;pos:number;end:number}, and include undefined in the index value type. The first index-signature candidate was rejected with TS2411.

### C15 — G13, G22

program.ts:1615-1628 and1653-1668 creates new outer resolution wrappers in Array.map callbacks. The only fields are resolvedModule/resolvedTypeReferenceDirective; any spread is confined to the nested module object. No failedLookupLocations field is copied into the new outer wrapper.

Explicit map result type arguments name the public result interfaces at their fresh construction, without changing JavaScript or globally changing Array.

### C16 — G30

tryGetModuleNameFromExportsOrImports (moduleSpecifiers.ts:1020) and tryGetModuleNameFromExports (:1122) return fresh objects with exactly moduleFileToTry, or recursive results of the same constructors. tryDirectoryWithPackageJson (:1277) spreads that known result. No arbitrary package-JSON object is returned as that result.

Add packageRootPath?:never to both private fixed-field helper result annotations; this is constructor-specific, not a never claim about every structurally compatible object.

## Narrowed-rule roads — premises only

### G10 — 3 site(s)

parser.ts:10624,10627,10630 spreads only conditional branches {preserve} / {} and {resolutionMode:parsed} / {}. preserve is derived by equality to "true"; parsed is validated by parseResolutionMode. Neither nonempty constructor has the other key.

**Premise:** Recognize certified unaliased fresh literal branches as the spread operand, checking each original branch before copy; do not treat a spread of an arbitrary view, alias, or assignment handle as absence evidence. Needs a ruling; no lowering code proposed.

### G20 — 1 site(s)

optionMapToObject (commandLineParser.ts:2729) calls Object.fromEntries. Every compiled-program caller feeds serializeCompilerOptions/serializeWatchOptions. serializeOptionBaseObject (:2795-2847) skips all because its fixed declaration (:641) has the Command_line_Options category; watch options have no all definition. Runtime witness serializing {all:true,module:ESNext} has no all key.

**Premise:** A call-specific key-set certificate for this serializer result, not Object.fromEntries in general. This proves only all absence: serialized module is a string, so it does not certify the entire object as CompilerOptions or promise the next field will pass.

### G26 — 1 site(s)

withJSDoc (parser.ts:1820-1834) calls mapDefined. core.ts:494 allocates result=[] and pushes elements; callback receives only input element/index and cannot see the result before return. The array is unpublished until node.jsDoc=jsDoc. getJSDocTags later legitimately adds jsDocCache (utilitiesPublic.ts:1271); runtime witness confirms absent-before/present-after.

**Premise:** A closed fresh-array factory / no-escape proof at this publication point; never is not an invariant of all JSDoc[] values after aliasing. Do not exempt arbitrary array-returning calls.

### G29 — 1 site(s)

withPackageId (moduleNameResolver.ts:127-140) returns r && a fresh literal containing exactly path,extension,packageId,resolvedUsingTsExtension. The false branch is undefined; the new object has no originalPath and no spreads.

**Premise:** Carry literal freshness through this short-circuit expression without treating a named object or assignment handle as fresh; equivalent to the existing conditional-branch premise, not a structural-object blanket exception.

### G31 — 1 site(s)

parser.ts:10624,10627,10630 spreads only conditional branches {preserve} / {} and {resolutionMode:parsed} / {}. preserve is derived by equality to "true"; parsed is validated by parseResolutionMode. Neither nonempty constructor has the other key.

**Premise:** Recognize certified unaliased fresh literal branches as the spread operand, checking each original branch before copy; do not treat a spread of an arbitrary view, alias, or assignment handle as absence evidence. Needs a ruling; no lowering code proposed.

### G45 — 1 site(s)

convertToTSConfig (commandLineParser.ts:2694-2697) spreads a conditional whose only constructors are {include,exclude} and {}. Neither branch has watchOptions; the containing object already writes its own watchOptions separately.

**Premise:** Certify the two original unaliased literals before spread. Explicit known keys, not the spread operation or the wide configFile view, establish absence.

These are ruling questions; no lowering implementation is proposed or built. The conditional cases rely on the original literal branches having known keys before publication, never on `{...view}` establishing absence. The JSDoc case is allocation-time only: the same array legitimately acquires jsDocCache after publication. The serializer case proves only the omitted all key; [runtime observations](evidence/rule-evidence.json) also show serialized module is a string. `P?: never` on all arrays or arbitrary object results would be false. No declaration edit was proposed for these six rule-only groups, so a candidate stock-diagnostic delta is not applicable to them.

## Hazards that stay refused

### G07, G15 — 7 sites

Concrete stock TypeScript inputs: package.json {name:42,version:"1.0.0",types:"index.d.ts"} remains object with numeric name after readJson; {name:"pkg",version:7,types:"index.d.ts"} remains cached PackageJsonPathFields with numeric version. moduleNameResolver.ts:2470 and utilities.ts:7785-7810 preserve raw extras. Both resolver/cache witnesses are recorded.

All declared S fields are valid in the relevant witness: object imposes no name contract; PackageJsonPathFields has a valid string name/types and does not declare version. A string/never declaration would be false without input validation.

### G08, G33, G34 — 6 sites

A valid ReusableDiagnosticRelatedInformation record in an accepted build-info relatedInformation array carries extra reportsUnnecessary:null. All required declared source fields are present and valid. Original convertToDiagnosticRelatedInformation (builder.ts:618) and toReusableDiagnosticRelatedInformation (:1543) are executed as source-extracted probes and preserve null by spread. Null is outside the target {}|undefined marker type.

This is a group-wide source-pool counterexample, not a claim that every listed coordinate was executed with the witness. C04/C05 gave zero stock diagnostics but are rejected as untruthful global declarations. Neither spread strips the hidden extra; input validation or stronger provenance would be needed before widening.

### G32 — 1 sites

A version-matched accepted .tsbuildinfo has a SerializedDiagnosticMessageChain with valid messageText/category/code and an undeclared repopulateInfo:42. readBuilderProgram keeps it in reusable diagnostics; emitter.ts:1144, watchPublic.ts:105 and builder.ts:2277 trace the input. builder.ts:534-567 checks message/category/code/next but never validates/removes that extra before returning or spreading it.

The hidden value is a number, not the target callback. A spread and the current isFullChain guard do not establish absence; claiming repopulateInfo?:never/function on the serialized source is false.

Package and cache input witnesses use stock TypeScript APIs. The diagnostic-marker witness additionally executes the **two exact original functions**, extracted from the unchanged builder source and type-erased by stock TypeScript; it is labeled as a source-extracted constructor probe, not an end-to-end native miscompile reproduction. Every declared source field in the witnesses is valid; the bad field is an **extra absent from S’s contract**. Marker null is intentional: `{}` accepts non-nullish numbers/strings too, so 42 would not demonstrate that marker-type mismatch. Source-type-pool witnesses do not claim that every listed coordinate was executed on that input. [Package/repopulate witnesses](evidence/hazards.json), [marker witness and original function text](evidence/diagnostic-hazard.json), [reproduction scripts](evidence/diagnostic-hazard.cjs). A stock diagnostic count cannot prove these input extras absent; spreading preserves them.

## Whole-site readings

Eighteen representatives were read whole, in the order below, starting with the two biggest groups. [The complete enclosing-function text](evidence/whole-readings.md) is retained; a callback counts as its whole function, with its enclosing statement also present in the AST context evidence. Constructor bodies above supplement the local reads.

| Group | Site ID | Site | Whole enclosing function |
| --- | --- | --- | --- |
| G01 | S105 | `transformers/legacyDecorators.ts:434:40` | `transformers/legacyDecorators.ts:429:5` finishClassElement |
| G02 | S126 | `utilities.ts:7960:37` | `utilities.ts:7959:1` positionsAreOnSameLine |
| G03 | S008 | `checker.ts:19948:30` | `checker.ts:19946:33` <arrow/anonymous> |
| G05 | S009 | `checker.ts:21355:119` | `checker.ts:21354:5` checkTypeAssignableTo |
| G07 | S033 | `moduleNameResolver.ts:2409:82` | `moduleNameResolver.ts:2407:1` getVersionPathsOfPackageJsonInfo |
| G08 | S018 | `checker.ts:2819:109` | `checker.ts:2819:85` <arrow/anonymous> |
| G10 | S048 | `parser.ts:10624:158` | `parser.ts:10616:81` <arrow/anonymous> |
| G13 | S055 | `program.ts:1615:42` | `program.ts:1615:42` <arrow/anonymous> |
| G20 | S027 | `commandLineParser.ts:2679:13` | `commandLineParser.ts:2660:1` convertToTSConfig |
| G21 | S131 | `watch.ts:990:53` | `watch.ts:987:1` performIncrementalCompilation |
| G26 | S053 | `parser.ts:1827:40` | `parser.ts:1820:5` withJSDoc |
| G29 | S032 | `moduleNameResolver.ts:140:12` | `moduleNameResolver.ts:127:1` withPackageId |
| G32 | S002 | `builder.ts:559:14` | `builder.ts:534:1` convertOrRepopulateDiagnosticMessageChain |
| G33 | S003 | `builder.ts:598:36` | `builder.ts:597:28` <arrow/anonymous> |
| G34 | S001 | `builder.ts:1528:48` | `builder.ts:1527:32` <arrow/anonymous> |
| G43 | S052 | `parser.ts:10760:45` | `parser.ts:10750:1` addPragmaForMatch |
| G44 | S051 | `parser.ts:10728:53` | `parser.ts:10697:1` extractPragmas |
| G30 | S040 | `moduleSpecifiers.ts:1277:30` | `moduleSpecifiers.ts:1249:5` tryDirectoryWithPackageJson |

## Complete site list

All 131 positions follow. Full literal messages, kinds/phases and owning units are in [sites.json](evidence/sites.json); [family-events.jsonl](evidence/family-events.jsonl) preserves every matching event. G links the site to the full (S,T,P) reason above.

| ID | G | Site | Position | Measured unit | Source expression (first line) |
| --- | --- | --- | --- | --- | --- |
| S066 | G01 | `transformers/classFields.ts:2469:42` | argument | `transformers/classFields.ts:355:1` | moveRangePastModifiers(property) |
| S067 | G01 | `transformers/classFields.ts:2505:43` | argument | `transformers/classFields.ts:355:1` | moveRangePastModifiers(property) |
| S076 | G01 | `transformers/es2015.ts:2635:71` | argument | `transformers/es2015.ts:489:1` | closeBraceLocation |
| S077 | G01 | `transformers/es2015.ts:2804:52` | argument | `transformers/es2015.ts:489:1` | getRangeUnion(declarations) |
| S078 | G01 | `transformers/es2015.ts:3045:52` | argument | `transformers/es2015.ts:489:1` | createRange(declarations[0]!.pos, last(declarations).end) |
| S083 | G01 | `transformers/esDecorators.ts:1078:56` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(node) |
| S084 | G01 | `transformers/esDecorators.ts:1086:56` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(node) |
| S085 | G01 | `transformers/esDecorators.ts:1230:40` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(original) |
| S086 | G01 | `transformers/esDecorators.ts:1357:56` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(member) |
| S087 | G01 | `transformers/esDecorators.ts:1389:56` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(member) |
| S091 | G01 | `transformers/esDecorators.ts:1719:40` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastModifiers(node) |
| S092 | G01 | `transformers/esDecorators.ts:2286:33` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(original) |
| S093 | G01 | `transformers/esDecorators.ts:2294:35` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(original) |
| S094 | G01 | `transformers/esDecorators.ts:609:56` | argument | `transformers/esDecorators.ts:296:1` | node.name ?? moveRangePastDecorators(node) |
| S095 | G01 | `transformers/esDecorators.ts:618:56` | argument | `transformers/esDecorators.ts:296:1` | node.name ?? moveRangePastDecorators(node) |
| S096 | G01 | `transformers/esDecorators.ts:892:52` | argument | `transformers/esDecorators.ts:296:1` | moveRangePastDecorators(node) |
| S097 | G01 | `transformers/esDecorators.ts:922:62` | argument | `transformers/esDecorators.ts:296:1` | node.name ?? moveRangePastDecorators(node) |
| S105 | G01 | `transformers/legacyDecorators.ts:434:40` | argument | `transformers/legacyDecorators.ts:89:1` | moveRangePastModifiers(original) |
| S106 | G01 | `transformers/legacyDecorators.ts:516:40` | argument | `transformers/legacyDecorators.ts:89:1` | moveRangePastModifiers(node) |
| S107 | G01 | `transformers/legacyDecorators.ts:659:35` | argument | `transformers/legacyDecorators.ts:89:1` | moveRangePastModifiers(member) |
| S108 | G01 | `transformers/legacyDecorators.ts:697:39` | argument | `transformers/legacyDecorators.ts:89:1` | moveRangePastModifiers(node) |
| S111 | G01 | `transformers/ts.ts:1625:40` | argument | `transformers/ts.ts:236:1` | moveRangePastModifiers(node) |
| S114 | G01 | `transformers/ts.ts:2528:39` | argument | `transformers/ts.ts:236:1` | createRange(node.name ? node.name.pos : node.pos, node.end) |
| S115 | G01 | `transformers/ts.ts:2531:38` | argument | `transformers/ts.ts:236:1` | createRange(-1, node.end) |
| S117 | G01 | `transformers/ts.ts:994:45` | argument | `transformers/ts.ts:236:1` | moveRangePastDecorators(node) |
| S031 | G01 | `factory/nodeFactory.ts:7489:45` | assignment | `factory/nodeFactory.ts:7432:1` | mergeTokenSourceMapRanges(tokenSourceMapRanges, destEmitNode.tokenSourceMapRanges!) |
| S022 | G02 | `checker.ts:49348:52` | argument | `checker.ts:1490:1` | sourceFile |
| S043 | G02 | `parser.ts:10015:36` | argument | `parser.ts:9929:1` | sourceFile |
| S044 | G02 | `parser.ts:10015:48` | argument | `parser.ts:9929:1` | result |
| S045 | G02 | `parser.ts:10087:37` | argument | `parser.ts:9929:1` | origSourceFile |
| S046 | G02 | `parser.ts:10244:42` | argument | `parser.ts:9929:1` | sourceFile |
| S058 | G02 | `program.ts:738:57` | argument | `program.ts:717:1` | file |
| S059 | G02 | `program.ts:739:76` | argument | `program.ts:717:1` | file |
| S123 | G02 | `utilities.ts:2532:70` | argument | `utilities.ts:2524:1` | sourceFile |
| S124 | G02 | `utilities.ts:7945:37` | argument | `utilities.ts:7943:1` | sourceFile |
| S125 | G02 | `utilities.ts:7950:37` | argument | `utilities.ts:7949:1` | sourceFile |
| S126 | G02 | `utilities.ts:7960:37` | argument | `utilities.ts:7959:1` | sourceFile |
| S127 | G02 | `utilities.ts:7972:37` | argument | `utilities.ts:7969:1` | sourceFile |
| S128 | G02 | `utilities.ts:7978:37` | argument | `utilities.ts:7976:1` | sourceFile |
| S122 | G02 | `utilities.ts:1261:24` | assignment | `utilities.ts:1240:1` | getSourceFileOfNode(node) |
| S004 | G03 | `checker.ts:13244:71` | argument | `checker.ts:1490:1` | baseConstructorType |
| S005 | G03 | `checker.ts:13276:58` | argument | `checker.ts:1490:1` | implementsType |
| S007 | G03 | `checker.ts:16607:33` | argument | `checker.ts:1490:1` | getTypeParametersForMapper(signature)! |
| S015 | G03 | `checker.ts:23690:67` | argument | `checker.ts:1490:1` | source |
| S016 | G03 | `checker.ts:23696:71` | argument | `checker.ts:1490:1` | constraint |
| S021 | G03 | `checker.ts:35806:48` | argument | `checker.ts:1490:1` | getTypeParametersForMapper(signature)! |
| S023 | G03 | `checker.ts:6968:54` | argument | `checker.ts:1490:1` | type |
| S024 | G03 | `checker.ts:7131:128` | argument | `checker.ts:1490:1` | getConstraintTypeFromMappedType(type) |
| S025 | G03 | `checker.ts:7142:90` | argument | `checker.ts:1490:1` | getModifiersTypeFromMappedType(target) |
| S026 | G03 | `checker.ts:9276:95` | argument | `checker.ts:1490:1` | type |
| S008 | G03 | `checker.ts:19948:30` | assignment | `checker.ts:1490:1` | append(result, getDeclaredTypeOfSymbol(symbol)) |
| S006 | G03 | `checker.ts:13612:16` | return | `checker.ts:1490:1` | links.declaredType \|\| (links.declaredType = createTypeParameter(symbol)) |
| S068 | G04 | `transformers/classFields.ts:2611:42` | argument | `transformers/classFields.ts:355:1` | propertyOriginalNode.name |
| S069 | G04 | `transformers/classFields.ts:3306:50` | argument | `transformers/classFields.ts:355:1` | node |
| S072 | G04 | `transformers/classThis.ts:122:59` | argument | `transformers/classThis.ts:104:1` | node.name |
| S103 | G04 | `transformers/generators.ts:2076:50` | argument | `transformers/generators.ts:337:1` | node |
| S109 | G04 | `transformers/legacyDecorators.ts:835:50` | argument | `transformers/legacyDecorators.ts:89:1` | node |
| S110 | G04 | `transformers/namedEvaluation.ts:205:68` | argument | `transformers/namedEvaluation.ts:179:1` | node.name |
| S116 | G04 | `transformers/ts.ts:2564:33` | argument | `transformers/ts.ts:236:1` | node.name |
| S009 | G05 | `checker.ts:21355:119` | argument | `checker.ts:1490:1` | errorOutputObject |
| S010 | G05 | `checker.ts:21498:139` | argument | `checker.ts:1490:1` | resultObj |
| S011 | G05 | `checker.ts:21591:145` | argument | `checker.ts:1490:1` | resultObj |
| S012 | G05 | `checker.ts:21594:134` | argument | `checker.ts:1490:1` | resultObj |
| S013 | G05 | `checker.ts:21681:145` | argument | `checker.ts:1490:1` | resultObj |
| S014 | G05 | `checker.ts:21684:134` | argument | `checker.ts:1490:1` | resultObj |
| S070 | G06 | `transformers/classFields.ts:931:41` | argument | `transformers/classFields.ts:355:1` | name.expression |
| S071 | G06 | `transformers/classFields.ts:934:47` | argument | `transformers/classFields.ts:355:1` | name.expression |
| S081 | G06 | `transformers/es2018.ts:789:51` | argument | `transformers/es2018.ts:155:1` | node.expression |
| S082 | G06 | `transformers/es2018.ts:793:53` | argument | `transformers/es2018.ts:155:1` | node.expression |
| S088 | G06 | `transformers/esDecorators.ts:1596:45` | argument | `transformers/esDecorators.ts:296:1` | name.expression |
| S089 | G06 | `transformers/esDecorators.ts:1599:51` | argument | `transformers/esDecorators.ts:296:1` | name.expression |
| S033 | G07 | `moduleNameResolver.ts:2409:82` | argument | `moduleNameResolver.ts:2407:1` | packageJsonInfo.contents.packageJsonContent |
| S034 | G07 | `moduleNameResolver.ts:2422:51` | argument | `moduleNameResolver.ts:2421:1` | packageJsonInfo.contents.packageJsonContent |
| S036 | G07 | `moduleNameResolver.ts:2494:56` | argument | `moduleNameResolver.ts:2489:1` | packageJson.contents.packageJsonContent |
| S037 | G07 | `moduleNameResolver.ts:2497:93` | argument | `moduleNameResolver.ts:2489:1` | packageJson.contents.packageJsonContent |
| S038 | G07 | `moduleNameResolver.ts:2498:116` | argument | `moduleNameResolver.ts:2489:1` | packageJson.contents.packageJsonContent |
| S018 | G08 | `checker.ts:2819:109` | argument | `checker.ts:1490:1` | r |
| S019 | G08 | `checker.ts:2819:174` | argument | `checker.ts:1490:1` | r |
| S129 | G08 | `utilities.ts:8809:39` | argument | `utilities.ts:8802:1` | d1i |
| S130 | G08 | `utilities.ts:8809:44` | argument | `utilities.ts:8802:1` | d2i |
| S118 | G09 | `tsbuildPublic.ts:454:17` | argument | `tsbuildPublic.ts:428:1` | host |
| S119 | G09 | `tsbuildPublic.ts:476:17` | argument | `tsbuildPublic.ts:428:1` | host |
| S120 | G09 | `tsbuildPublic.ts:490:17` | argument | `tsbuildPublic.ts:428:1` | host |
| S048 | G10 | `parser.ts:10624:158` | spread | `parser.ts:10602:1` | ...(preserve ? { preserve } : {}) |
| S049 | G10 | `parser.ts:10627:104` | spread | `parser.ts:10602:1` | ...(preserve ? { preserve } : {}) |
| S050 | G10 | `parser.ts:10630:100` | spread | `parser.ts:10602:1` | ...(preserve ? { preserve } : {}) |
| S074 | G11 | `transformers/es2015.ts:2359:35` | argument | `transformers/es2015.ts:489:1` | firstAccessor.name |
| S075 | G11 | `transformers/es2015.ts:2368:41` | argument | `transformers/es2015.ts:489:1` | firstAccessor.name |
| S090 | G11 | `transformers/esDecorators.ts:1611:50` | argument | `transformers/esDecorators.ts:296:1` | node.name |
| S017 | G12 | `checker.ts:25216:31` | assignment | `checker.ts:1490:1` | classType && getBaseTypes(classType)[0] |
| S020 | G12 | `checker.ts:31813:31` | assignment | `checker.ts:1490:1` | classType && getBaseTypes(classType)[0] |
| S055 | G13 | `program.ts:1615:42` | assignment | `program.ts:1521:1` | (moduleNames, containingFile, redirectedReference, options, containingSourceFile, reusedNames) => |
| S056 | G13 | `program.ts:1616:13` | return | `program.ts:1521:1` | host.resolveModuleNames!( |
| S062 | G14 | `resolutionCache.ts:1053:26` | assignment | `resolutionCache.ts:581:1` | ts_resolveLibrary(libraryName, resolveFrom, options, host, libraryResolutionCache) |
| S061 | G14 | `resolutionCache.ts:1028:27` | property in object literal | `resolutionCache.ts:581:1` | resolvedModuleNames |
| S035 | G15 | `moduleNameResolver.ts:2470:36` | assignment | `moduleNameResolver.ts:2445:1` | readJson(packageJsonPath, host) as PackageJson |
| S039 | G15 | `moduleNameResolver.ts:831:89` | assignment | `moduleNameResolver.ts:813:1` | readJson(packageJsonPath, host) as PackageJson |
| S041 | G16 | `moduleSpecifiers.ts:200:71` | argument | `moduleSpecifiers.ts:182:1` | importingSourceFile |
| S042 | G16 | `moduleSpecifiers.ts:243:63` | argument | `moduleSpecifiers.ts:182:1` | importingSourceFile |
| S073 | G17 | `transformers/es2015.ts:2172:49` | argument | `transformers/es2015.ts:489:1` | node |
| S030 | G17 | `factory/emitNode.ts:133:12` | return | `factory/emitNode.ts:132:1` | node.emitNode?.sourceMapRange ?? node |
| S080 | G18 | `transformers/es2017.ts:629:13` | argument | `transformers/es2017.ts:120:1` | node |
| S100 | G18 | `transformers/esnext.ts:617:39` | argument | `transformers/esnext.ts:72:1` | node |
| S099 | G19 | `transformers/esnext.ts:596:42` | argument | `transformers/esnext.ts:72:1` | node |
| S104 | G19 | `transformers/generators.ts:751:17` | argument | `transformers/generators.ts:337:1` | node |
| S027 | G20 | `commandLineParser.ts:2679:13` | spread | `commandLineParser.ts:2660:1` | ...optionMapToObject(optionMap) |
| S131 | G21 | `watch.ts:990:53` | argument | `watch.ts:987:1` | input |
| S057 | G22 | `program.ts:1661:13` | return | `program.ts:1521:1` | host.resolveTypeReferenceDirectives!( |
| S064 | G23 | `resolutionCache.ts:999:27` | property in object literal | `resolutionCache.ts:581:1` | resolvedTypeReferenceDirectives |
| S060 | G24 | `resolutionCache.ts:1021:16` | return | `resolutionCache.ts:581:1` | resolveNamesWithLocalCache({ |
| S063 | G25 | `resolutionCache.ts:992:16` | return | `resolutionCache.ts:581:1` | resolveNamesWithLocalCache({ |
| S053 | G26 | `parser.ts:1827:40` | assignment | `parser.ts:1437:1` | jsDoc |
| S054 | G27 | `program.ts:1584:67` | argument | `program.ts:1521:1` | host |
| S121 | G28 | `tsbuildPublic.ts:501:66` | argument | `tsbuildPublic.ts:428:1` | host |
| S032 | G29 | `moduleNameResolver.ts:140:12` | return | `moduleNameResolver.ts:127:1` | r && { path: r.path, extension: r.ext, packageId, resolvedUsingTsExtension: r.resolvedUsingTsExtension } |
| S040 | G30 | `moduleSpecifiers.ts:1277:30` | spread | `moduleSpecifiers.ts:1186:1` | ...fromExports |
| S047 | G31 | `parser.ts:10624:111` | spread | `parser.ts:10602:1` | ...(parsed ? { resolutionMode: parsed } : {}) |
| S002 | G32 | `builder.ts:559:14` | spread | `builder.ts:534:1` | ...chain |
| S003 | G33 | `builder.ts:598:36` | assignment | `builder.ts:590:1` | convertToDiagnosticRelatedInformation(diagnostic, diagnosticFilePath, newProgram, toPathInBuildInfoDirectory) |
| S001 | G34 | `builder.ts:1528:48` | assignment | `builder.ts:1258:1` | toReusableDiagnosticRelatedInformation(diagnostic, diagnosticFilePath) |
| S029 | G35 | `emitter.ts:1409:32` | argument | `emitter.ts:1213:1` | sourceFile |
| S101 | G36 | `transformers/generators.ts:1389:79` | argument | `transformers/generators.ts:337:1` | node.name |
| S098 | G37 | `transformers/esnext.ts:539:43` | argument | `transformers/esnext.ts:72:1` | node |
| S112 | G38 | `transformers/ts.ts:2038:62` | argument | `transformers/ts.ts:236:1` | node |
| S102 | G39 | `transformers/generators.ts:1392:13` | argument | `transformers/generators.ts:337:1` | node |
| S113 | G40 | `transformers/ts.ts:2041:46` | argument | `transformers/ts.ts:236:1` | node |
| S065 | G41 | `transformers/classFields.ts:2465:42` | argument | `transformers/classFields.ts:355:1` | propertyOriginalNode |
| S079 | G42 | `transformers/es2015.ts:4835:39` | argument | `transformers/es2015.ts:489:1` | node |
| S052 | G43 | `parser.ts:10760:45` | property in object literal | `parser.ts:10750:1` | argument |
| S051 | G44 | `parser.ts:10728:53` | property in object literal | `parser.ts:10697:1` | argument |
| S028 | G45 | `commandLineParser.ts:2694:9` | spread | `commandLineParser.ts:2660:1` | ...(configParseResult.options.configFile?.configFileSpecs ? { |

## Audit scope

Only this REPORT.md and its evidence are committed. No adaptation, ledger, `internal/`, API sanction or lowering change is made. Scratch candidate sources are reset and never committed. [Validation](evidence/validation.txt) records the exact census recount, full-site coverage, unchanged source/binary pins, candidate checks and archive integrity; [SHA256SUMS](evidence/SHA256SUMS) covers every other evidence file.
