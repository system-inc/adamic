# Step 09: public any contracts, decisions open

Measured on TypeScript 6.0.3, pin 050880ce59e30b356b686bd3144efe24f875ebc8, with the adapted pipeline from origin/codex/stage3-explicit-any-2 at 1e920a31b601422c75a26ead083e9c1423c34b66. That branch supplies RESIDUE.md; current main does not contain it. This scout changes no adaptation, compiler or public declaration. Main/compiler base is 45487a809f89885a3fc651cd590e7dabf31362dc.

The residue has **19 source contracts**: 16 JSON/config/predicate observations, two host timer storage observations and one staged allocator observation. A twentieth observation is a concrete Node timer annotation that the latent checker reports any; it is retained as an anomaly, not counted as an unedited any. Three option-conversion overload rows share one implementation and are not three independent execution sites.

## What Node actually did

The instrumented scratch tsc CLI matched unchanged stdout, stderr and exit goldens on **301/301** acceptance projects in 121.311 seconds. The existing driver runs --noEmit: it neither watches nor builds projects. All 301 project captures and their exact shape counts are in evidence/project-captures.json.gz; evidence/observations.json is their aggregate. Four public-only functions are tree-shaken out of the CLI (see instrumentation.json). No zero-hit function is declared safe or impossible.

The raw-config worker's json argument was undefined in all 301 runs: this corpus enters through a JsonSourceFile, not the arbitrary JavaScript public json path. It returns ParsedCommandLine objects with raw data and normalized options. The list predicate ran 2,608 times: values were 301 empty arrays, 312 string-element arrays, 765 booleans, 929 strings and 301 object-shaped raw option values. Every result was boolean; these observations do not prove the declared predicate for invalid external arrays.

The CLI called getNodeConstructor 7,280 times across all 301 projects. Bare Node headers completed 18,204,760 allocations with parent undefined; Token/Identifier headers also had parent undefined. Type headers completed 87,197 allocations with only flags and no symbol. This is the allocation stage, not the state after factory/checker initialization. The counts include standard-library parsing repeated by separate CLI processes.

No setTimeout/clearTimeout registration or cancellation was observed in these 301 noEmit projects. Focused API calls, explicitly separate from acceptance coverage, returned a Node Timeout object and cancelled the same handle. The timer fixture also uses a numeric custom-host handle. The acceptance corpus cannot establish all custom-host domains or callback ownership.

Focused actual API calls returned primitive number, boolean and string through readJson/readJsonOrUndefined and convertToObject; convertToObject also returned null/arrays. readJson's existing || {} changes null and false to {}. Malformed input recovers differently through strict/loose readers. The full API uses NodeObject as its allocator provider; its parent is undefined too. evidence/counterexamples.json distinguishes those API calls from bare CLI constructor observations.

## Rules and interpretation

The accepted escape-hatch document says: "As written: any, as unknown as, expando additions, Object.defineProperty and Function are refused" and requires proven predicate bodies. A loud boundary check below is a proposed checked representation/ABI decision, not permission to admit any or cast unknown to an interface. A check that cannot prove initialization, ownership or callable argument relationships must leave the value refused. Costs below are inferred operation/allocation costs, not benchmark results. Nothing is decided in this scout.

## Caller coverage

contracts.json records the exact adapted file:line:column, declaration/signature, every resolved internal source reference and direct call, and linked public owner declarations. It scans all src, including services, server, tools, harness and tests; test callers are recorded alongside production callers. Shorthand captured methods are resolved by their value symbol. Structural timer/allocator member names have a separate complete member-reference ledger with context and test/production labels; name matches are candidate uses, not a claim that structurally distinct hosts share one symbol. External application callers cannot be enumerated by an internal census.

Every reference and all declaration text are retained in contracts.json, including function values passed onward. The Markdown below lists all source-symbol references and links to supplemental member lists. Public-owned raw fields still permit arbitrary JavaScript, including invalid option values. A recursive JSON type is sound for parser-produced JSON, but cannot silently replace that public input contract.

## S01: src/compiler/commandLineParser.ts:2472:17

Public convertToObject returns a recursive value, including primitive JSON. Replacing its public any needs an additional API sanction; it is not necessarily an object.

Source owner: convertToObject, src/compiler/commandLineParser.ts:2472.

```typescript
export function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): any
```

Public/dependent declaration owners:

- JsonConfigValue: src/compiler/commandLineParser.ts:2426.

```typescript
export type JsonConfigValue = string | number | boolean | null | undefined | JsonConfigValue[] | JsonConfigObject;
```

- convertToObject: src/compiler/commandLineParser.ts:2472.

```typescript
export function convertToObject(sourceFile: JsonSourceFile, errors: Diagnostic[]): any
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3425:31: direct-call, `convertToObject`.
- src/testRunner/unittests/config/tsconfigParsing.ts:265:37: reference, `convertToObject`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: Check/represent recursive JSON plus recovery undefined at the boundary, preserving number/string/boolean/null/array roots and existing readJson fallback on falsy values. Cost: O(1) root tag and O(nodes) if a whole result must be validated. Object-only checks reject observed primitive returns. Loose recovery/error diagnostics and getters at public raw-input APIs need separate policies.
- **internal only wrapper**: Give internal parser/result owners JsonConfigValue and propagate it through readJsonOrUndefined/readJson and their consumers. A wrapper calling the proven internal parser can erase to zero JavaScript; calling an any-returning public API still requires a sound check. Public convertToObject/config results require a sanction or a separate checked ABI shim; casts to object are false.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S02: src/compiler/commandLineParser.ts:2611:72

The existing predicate accepts every array without checking its elements. Declaring its input to already be validated CompilerOptionsValue would retain the lie.

Source owner: isCompilerOptionsValue, src/compiler/commandLineParser.ts:2611.

```typescript
function isCompilerOptionsValue(option: CommandLineOption | undefined, value: any): value is CompilerOptionsValue
```

Public/dependent declaration owners:

- CompilerOptionsValue: src/compiler/types.ts:7412.

```typescript
export type CompilerOptionsValue = string | number | boolean | (string | number)[] | string[] | MapLike<string[]> | PluginImport[] | ProjectReference[] | null | undefined;
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

- convertCompilerOptionsFromJson: src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

Internal callers/references:

- src/compiler/commandLineParser.ts:2618:38: direct-call, `isCompilerOptionsValue`.
- src/compiler/commandLineParser.ts:3819:9: direct-call, `isCompilerOptionsValue`.

Acceptance observation: 301/301 projects; phases {"arg0": 2608, "arg1": 2608, "return": 2608}. All observed domains are retained under this row in contracts.json; no universal input-domain inference is made.

- **loud boundary check**: Do not strengthen the existing shallow array predicate and claim byte equality: [{}] currently returns true. Validate elements at the conversion/use boundary, preserving invalid-input diagnostics, or explicitly approve a changed failure policy. Cost: O(elements) checks. An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S03: src/compiler/commandLineParser.ts:3018:44

Public raw-config input accepts arbitrary JavaScript; preserve its caller shape through a generic boundary and distinguish raw from validated options.

Source owner: parseJsonConfigFileContent, src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Public/dependent declaration owners:

- ParsedCommandLine: src/compiler/types.ts:7710.

```typescript
export interface ParsedCommandLine {
    raw?: any;
// Full declaration retained in contracts.json.
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Internal callers/references:

- src/harness/compilerImpl.ts:40:27: reference, `parseJsonConfigFileContent`.
- src/testRunner/unittests/config/helpers.ts:15:15: reference, `parseJsonConfigFileContent`.
- src/testRunner/unittests/tscWatch/watchApi.ts:50:40: reference, `parseJsonConfigFileContent`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S04: src/compiler/commandLineParser.ts:3065:5

Internal raw-config worker inherits the public erased input; a JSON-only annotation would exclude supported JavaScript input.

Source owner: parseJsonConfigFileContentWorker, src/compiler/commandLineParser.ts:3064.

```typescript
function parseJsonConfigFileContentWorker(
    json: any,
    sourceFile: TsConfigSourceFile | undefined,
    host: ParseConfigHost,
    basePath: string,
    existingOptions: CompilerOptions = {},
    existingWatchOptions: WatchOptions | undefined,
    configFileName?: string,
    resolutionStack: Path[] = [],
    extraFileExtensions: readonly FileExtensionInfo[] = [],
    extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>,
): ParsedCommandLine
```

Public/dependent declaration owners:

- ParsedCommandLine: src/compiler/types.ts:7710.

```typescript
export interface ParsedCommandLine {
    raw?: any;
// Full declaration retained in contracts.json.
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3019:12: direct-call, `parseJsonConfigFileContentWorker`.
- src/compiler/commandLineParser.ts:3031:20: direct-call, `parseJsonConfigFileContentWorker`.

Acceptance observation: 301/301 projects; phases {"arg0": 301, "arg1": 301, "arg2": 301, "arg3": 301, "arg4": 301, "arg5": 301, "arg6": 301, "arg7": 301, "arg8": 301, "arg9": 301, "return": 301}. All observed domains are retained under this row in contracts.json; no universal input-domain inference is made.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S05: src/compiler/commandLineParser.ts:3359:43

Raw config comes from the public arbitrary-JavaScript input and extends pipeline; requires the correlated raw-owner shape.

Source owner: canJsonReportNoInputFiles, src/compiler/commandLineParser.ts:3359.

```typescript
export function canJsonReportNoInputFiles(raw: any): boolean
```

Public/dependent declaration owners:

- ParsedCommandLine: src/compiler/types.ts:7710.

```typescript
export interface ParsedCommandLine {
    raw?: any;
// Full declaration retained in contracts.json.
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3197:49: direct-call, `canJsonReportNoInputFiles`.
- src/compiler/watchPublic.ts:4:5: reference, `canJsonReportNoInputFiles`.
- src/compiler/watchPublic.ts:979:47: direct-call, `canJsonReportNoInputFiles`.
- src/compiler/tsbuildPublic.ts:9:5: reference, `canJsonReportNoInputFiles`.
- src/compiler/tsbuildPublic.ts:1245:17: direct-call, `canJsonReportNoInputFiles`.
- src/server/project.ts:11:5: reference, `canJsonReportNoInputFiles`.
- src/server/project.ts:3157:13: direct-call, `canJsonReportNoInputFiles`.

Acceptance observation: 301/301 projects; phases {"arg0": 301, "return": 301}. All observed domains are retained under this row in contracts.json; no universal input-domain inference is made.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S06: src/compiler/commandLineParser.ts:3502:5

Raw config conversion inherits the public boundary and mutates/carries normalized properties; requires the raw versus validated owner contract.

Source owner: parseOwnConfigOfJson, src/compiler/commandLineParser.ts:3501.

```typescript
function parseOwnConfigOfJson(
    json: any,
    host: ParseConfigHost,
    basePath: string,
    configFileName: string | undefined,
    errors: Diagnostic[],
): ParsedTsconfig
```

Public/dependent declaration owners:

- ParsedCommandLine: src/compiler/types.ts:7710.

```typescript
export interface ParsedCommandLine {
    raw?: any;
// Full declaration retained in contracts.json.
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3429:9: direct-call, `parseOwnConfigOfJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S07: src/compiler/commandLineParser.ts:3730:45

Raw compileOnSave input is classified by existing option conversion; a validated boolean input would change accepted invalid-input/error behavior.

Source owner: convertCompileOnSaveOptionFromJson, src/compiler/commandLineParser.ts:3730.

```typescript
function convertCompileOnSaveOptionFromJson(jsonOption: any, basePath: string, errors: Diagnostic[]): boolean
```

Public/dependent declaration owners:

- ParsedCommandLine: src/compiler/types.ts:7710.

```typescript
export interface ParsedCommandLine {
    raw?: any;
// Full declaration retained in contracts.json.
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3515:26: direct-call, `convertCompileOnSaveOptionFromJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S08: src/compiler/commandLineParser.ts:3738:48

Public compiler-option conversion accepts arbitrary raw property values; generic caller shape and validation result need public API ownership.

Source owner: convertCompilerOptionsFromJson, src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

Public/dependent declaration owners:

- convertCompilerOptionsFromJson: src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

Internal callers/references:

- src/harness/fourslashImpl.ts:349:41: reference, `convertCompilerOptionsFromJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S09: src/compiler/commandLineParser.ts:3744:48

Public acquisition-option conversion has the same arbitrary-input boundary.

Source owner: convertTypeAcquisitionFromJson, src/compiler/commandLineParser.ts:3744.

```typescript
export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; }
```

Public/dependent declaration owners:

- convertTypeAcquisitionFromJson: src/compiler/commandLineParser.ts:3744.

```typescript
export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; }
```

Internal callers/references:

No resolved source-symbol reference. For allocator callback/member invocation use the supplemental structural member ledger below; an uncalled public-only API is still an external contract.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S10: src/compiler/commandLineParser.ts:3757:47

Internal compiler-options worker inherits that boundary; a validated options type would claim validation before it happens.

Source owner: convertCompilerOptionsFromJsonWorker, src/compiler/commandLineParser.ts:3757.

```typescript
function convertCompilerOptionsFromJsonWorker(jsonOptions: any, basePath: string, errors: Diagnostic[], configFileName?: string): CompilerOptions
```

Public/dependent declaration owners:

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

- convertCompilerOptionsFromJson: src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3512:21: direct-call, `convertCompilerOptionsFromJsonWorker`.
- src/compiler/commandLineParser.ts:3740:21: direct-call, `convertCompilerOptionsFromJsonWorker`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S11: src/compiler/commandLineParser.ts:3770:47

Internal acquisition worker inherits that boundary.

Source owner: convertTypeAcquisitionFromJsonWorker, src/compiler/commandLineParser.ts:3770.

```typescript
function convertTypeAcquisitionFromJsonWorker(jsonOptions: any, basePath: string, errors: Diagnostic[], configFileName?: string): TypeAcquisition
```

Public/dependent declaration owners:

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

- convertTypeAcquisitionFromJson: src/compiler/commandLineParser.ts:3744.

```typescript
export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; }
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3513:29: direct-call, `convertTypeAcquisitionFromJsonWorker`.
- src/compiler/commandLineParser.ts:3746:21: direct-call, `convertTypeAcquisitionFromJsonWorker`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S12: src/compiler/commandLineParser.ts:3776:44

Watch-option conversion inherits raw JavaScript values; preserve invalid-input diagnostics rather than asserting the final options shape.

Source owner: convertWatchOptionsFromJsonWorker, src/compiler/commandLineParser.ts:3776.

```typescript
function convertWatchOptionsFromJsonWorker(jsonOptions: any, basePath: string, errors: Diagnostic[]): WatchOptions | undefined
```

Public/dependent declaration owners:

- ParsedCommandLine: src/compiler/types.ts:7710.

```typescript
export interface ParsedCommandLine {
    raw?: any;
// Full declaration retained in contracts.json.
```

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3514:26: direct-call, `convertWatchOptionsFromJsonWorker`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S13: src/compiler/commandLineParser.ts:3780:81

Raw option-conversion overload; must correlate raw fields, option kinds and normalized results.

Source owner: convertOptionsFromJson, src/compiler/commandLineParser.ts:3780.

```typescript
function convertOptionsFromJson(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: any, basePath: string, defaultOptions: undefined, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]): WatchOptions | undefined;
```

Public/dependent declaration owners:

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

- convertCompilerOptionsFromJson: src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

- convertTypeAcquisitionFromJson: src/compiler/commandLineParser.ts:3744.

```typescript
export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; }
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3759:5: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3772:5: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3777:12: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3781:10: reference, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3782:10: reference, `convertOptionsFromJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S14: src/compiler/commandLineParser.ts:3781:81

Raw option-conversion overload with defaults; same owner correlation.

Source owner: convertOptionsFromJson, src/compiler/commandLineParser.ts:3781.

```typescript
function convertOptionsFromJson(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: any, basePath: string, defaultOptions: CompilerOptions | TypeAcquisition, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[]): CompilerOptions | TypeAcquisition;
```

Public/dependent declaration owners:

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

- convertCompilerOptionsFromJson: src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

- convertTypeAcquisitionFromJson: src/compiler/commandLineParser.ts:3744.

```typescript
export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; }
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3759:5: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3772:5: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3777:12: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3780:10: reference, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3782:10: reference, `convertOptionsFromJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S15: src/compiler/commandLineParser.ts:3782:81

Raw option-conversion implementation; same owner correlation.

Source owner: convertOptionsFromJson, src/compiler/commandLineParser.ts:3782.

```typescript
function convertOptionsFromJson(optionsNameMap: Map<string, CommandLineOption>, jsonOptions: any, basePath: string, defaultOptions: CompilerOptions | TypeAcquisition | WatchOptions | undefined, diagnostics: DidYouMeanOptionsDiagnostics, errors: Diagnostic[])
```

Public/dependent declaration owners:

- parseJsonConfigFileContent: src/compiler/commandLineParser.ts:3018.

```typescript
export function parseJsonConfigFileContent(json: any, host: ParseConfigHost, basePath: string, existingOptions?: CompilerOptions, configFileName?: string, resolutionStack?: Path[], extraFileExtensions?: readonly FileExtensionInfo[], extendedConfigCache?: Map<string, ExtendedConfigCacheEntry>, existingWatchOptions?: WatchOptions): ParsedCommandLine
```

- convertCompilerOptionsFromJson: src/compiler/commandLineParser.ts:3738.

```typescript
export function convertCompilerOptionsFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: CompilerOptions; errors: Diagnostic[]; }
```

- convertTypeAcquisitionFromJson: src/compiler/commandLineParser.ts:3744.

```typescript
export function convertTypeAcquisitionFromJson(jsonOptions: any, basePath: string, configFileName?: string): { options: TypeAcquisition; errors: Diagnostic[]; }
```

Internal callers/references:

- src/compiler/commandLineParser.ts:3759:5: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3772:5: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3777:12: direct-call, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3780:10: reference, `convertOptionsFromJson`.
- src/compiler/commandLineParser.ts:3781:10: reference, `convertOptionsFromJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: An explicit checked raw-value boundary must accept invalid option values long enough to preserve tsc diagnostics, and validate only the requested field when consumed. A whole JSON assertion excludes supported JavaScript values, accessors and cycles. Cost: O(1) per scalar use, O(elements/fields) for a recursive structural check; repeated checks/copies can add traversal and allocation. A panic on unsupported input changes the API failure contract and needs a ruling.
- **internal only wrapper**: Carry a caller-specific Raw<C> separately from normalized CompilerOptions/TypeAcquisition/WatchOptions, with internal field conversions returning explicit validated unions. Use the existing recursive JsonConfigValue only for parser-produced values. Generic preservation can erase to zero added JavaScript; field verification and dynamic dictionary representation remain required. No unknown-to-final-type assertion or alias for any is allowed. Public any entrypoints remain refused unless a checked entry shim or a public sanction is approved.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S16: src/compiler/sys.ts:52:31

Already annotated NodeJS.Timeout | undefined. Stock TypeScript resolves Timeout and undefined, neither any. Latent still reports any; external/type-resolution cause is not diagnosed. This is not an unedited explicit any.

Source owner: clearTimeout, src/compiler/sys.ts:52.

```typescript
declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;
```

Public/dependent declaration owners:

- System: src/compiler/sys.ts:1396.

```typescript
export interface System {
    setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout?(timeoutId: any): void;
// Full declaration retained in contracts.json.
```

Internal callers/references:

- src/compiler/sys.ts:1507:13: reference, `clearTimeout`.
- src/compiler/sys.ts:1605:13: reference, `clearTimeout`.

Member ledger `setTimeout` (44 mentions, including declarations and tests):

- src/compiler/sys.ts:51:18, production.
- src/compiler/sys.ts:247:5, production.
- src/compiler/sys.ts:378:68, production.
- src/compiler/sys.ts:460:5, production.
- src/compiler/sys.ts:491:30, production.
- src/compiler/sys.ts:584:5, production.
- src/compiler/sys.ts:600:5, production.
- src/compiler/sys.ts:771:37, production.
- src/compiler/sys.ts:971:5, production.
- src/compiler/sys.ts:994:5, production.
- src/compiler/sys.ts:1055:101, production.
- src/compiler/sys.ts:1059:107, production.
- src/compiler/sys.ts:1123:17, production.
- src/compiler/sys.ts:1441:5, production.
- src/compiler/sys.ts:1506:13, production.
- src/compiler/sys.ts:1604:13, production.
- src/compiler/watch.ts:679:9, production.
- src/compiler/watch.ts:679:46, production.
- src/compiler/watchPublic.ts:169:5, production.
- src/compiler/watchPublic.ts:863:19, production.
- src/compiler/watchPublic.ts:868:57, production.
- src/compiler/watchPublic.ts:882:19, production.
- src/compiler/watchPublic.ts:890:37, production.
- src/compiler/tsbuildPublic.ts:2062:24, production.
- src/compiler/tsbuildPublic.ts:2068:58, production.
- src/server/types.ts:33:5, production.
- src/server/utilities.ts:30:57, production.
- src/server/utilities.ts:63:34, production.
- src/server/session.ts:404:48, production.
- src/server/typingInstallerAdapter.ts:242:19, production.
- src/harness/harnessLanguageService.ts:593:5, test.
- src/lib/dom.generated.d.ts:41731:5, production.
- src/lib/dom.generated.d.ts:44107:18, production.
- src/lib/webworker.generated.d.ts:13837:5, production.
- src/lib/webworker.generated.d.ts:14957:18, production.
- src/testRunner/parallel/host.ts:326:44, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:13, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:30, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103:5, test.
- src/testRunner/unittests/sys/symlinkWatching.ts:20:16, test.
- src/testRunner/unittests/tscWatch/watchApi.ts:190:14, test.
- src/testRunner/unittests/tsserver/session.ts:42:5, test.
- src/tsserver/nodeServer.ts:268:9, production.
- src/tsserver/nodeServer.ts:268:22, production.

Member ledger `clearTimeout` (39 mentions, including declarations and tests):

- src/compiler/sys.ts:52:18, production.
- src/compiler/sys.ts:585:5, production.
- src/compiler/sys.ts:601:5, production.
- src/compiler/sys.ts:768:13, production.
- src/compiler/sys.ts:972:5, production.
- src/compiler/sys.ts:995:5, production.
- src/compiler/sys.ts:1124:17, production.
- src/compiler/sys.ts:1442:5, production.
- src/compiler/sys.ts:1507:13, production.
- src/compiler/sys.ts:1605:13, production.
- src/compiler/watch.ts:680:9, production.
- src/compiler/watch.ts:680:48, production.
- src/compiler/watchPublic.ts:171:5, production.
- src/compiler/watchPublic.ts:857:14, production.
- src/compiler/watchPublic.ts:863:39, production.
- src/compiler/watchPublic.ts:882:39, production.
- src/compiler/watchPublic.ts:887:18, production.
- src/compiler/tsbuildPublic.ts:2062:53, production.
- src/compiler/tsbuildPublic.ts:2066:23, production.
- src/server/types.ts:34:5, production.
- src/server/utilities.ts:27:23, production.
- src/server/utilities.ts:39:19, production.
- src/server/session.ts:450:48, production.
- src/harness/harnessLanguageService.ts:597:5, test.
- src/lib/dom.generated.d.ts:41718:5, production.
- src/lib/dom.generated.d.ts:44094:18, production.
- src/lib/webworker.generated.d.ts:13824:5, production.
- src/lib/webworker.generated.d.ts:14944:18, production.
- src/testRunner/parallel/host.ts:320:29, test.
- src/testRunner/parallel/worker.ts:61:22, test.
- src/testRunner/parallel/worker.ts:67:22, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:13, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:32, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111:5, test.
- src/testRunner/unittests/tscWatch/watchApi.ts:191:14, test.
- src/testRunner/unittests/tsserver/reloadProjects.ts:86:14, test.
- src/testRunner/unittests/tsserver/session.ts:45:5, test.
- src/tsserver/nodeServer.ts:269:9, production.
- src/tsserver/nodeServer.ts:269:24, production.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: Do not add a check to paper over this observation: source already says NodeJS.Timeout | undefined. Stock checker agrees. First diagnose the latent checker external declaration resolution. Cost: investigation only; no unsupported source any is claimed.
- **internal only wrapper**: An internal concrete Node timer type is already present here. Wrapping it does not fix the unexplained latent any; preserve the anomaly and use System/WatchHost generic handles only at their actual public declarations.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S17: src/compiler/sys.ts:464:9

Host timer handle in polling storage. System permits numeric test handles, Node handles and custom opaque handles. Carry H through registration, storage and cancellation; public System/WatchHost changes need a lane sanction.

Source owner: pollScheduled, src/compiler/sys.ts:464.

```typescript
pollScheduled: any
```

Public/dependent declaration owners:

- System: src/compiler/sys.ts:1396.

```typescript
export interface System {
    setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout?(timeoutId: any): void;
// Full declaration retained in contracts.json.
```

- WatchHost: src/compiler/watchPublic.ts:160.

```typescript
export interface WatchHost {
    setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout?(timeoutId: any): void;
// Full declaration retained in contracts.json.
```

Internal callers/references:

- src/compiler/sys.ts:484:9: reference, `pollScheduled`.
- src/compiler/sys.ts:490:37: reference, `pollScheduled`.
- src/compiler/sys.ts:491:9: reference, `pollScheduled`.

Member ledger `setTimeout` (44 mentions, including declarations and tests):

- src/compiler/sys.ts:51:18, production.
- src/compiler/sys.ts:247:5, production.
- src/compiler/sys.ts:378:68, production.
- src/compiler/sys.ts:460:5, production.
- src/compiler/sys.ts:491:30, production.
- src/compiler/sys.ts:584:5, production.
- src/compiler/sys.ts:600:5, production.
- src/compiler/sys.ts:771:37, production.
- src/compiler/sys.ts:971:5, production.
- src/compiler/sys.ts:994:5, production.
- src/compiler/sys.ts:1055:101, production.
- src/compiler/sys.ts:1059:107, production.
- src/compiler/sys.ts:1123:17, production.
- src/compiler/sys.ts:1441:5, production.
- src/compiler/sys.ts:1506:13, production.
- src/compiler/sys.ts:1604:13, production.
- src/compiler/watch.ts:679:9, production.
- src/compiler/watch.ts:679:46, production.
- src/compiler/watchPublic.ts:169:5, production.
- src/compiler/watchPublic.ts:863:19, production.
- src/compiler/watchPublic.ts:868:57, production.
- src/compiler/watchPublic.ts:882:19, production.
- src/compiler/watchPublic.ts:890:37, production.
- src/compiler/tsbuildPublic.ts:2062:24, production.
- src/compiler/tsbuildPublic.ts:2068:58, production.
- src/server/types.ts:33:5, production.
- src/server/utilities.ts:30:57, production.
- src/server/utilities.ts:63:34, production.
- src/server/session.ts:404:48, production.
- src/server/typingInstallerAdapter.ts:242:19, production.
- src/harness/harnessLanguageService.ts:593:5, test.
- src/lib/dom.generated.d.ts:41731:5, production.
- src/lib/dom.generated.d.ts:44107:18, production.
- src/lib/webworker.generated.d.ts:13837:5, production.
- src/lib/webworker.generated.d.ts:14957:18, production.
- src/testRunner/parallel/host.ts:326:44, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:13, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:30, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103:5, test.
- src/testRunner/unittests/sys/symlinkWatching.ts:20:16, test.
- src/testRunner/unittests/tscWatch/watchApi.ts:190:14, test.
- src/testRunner/unittests/tsserver/session.ts:42:5, test.
- src/tsserver/nodeServer.ts:268:9, production.
- src/tsserver/nodeServer.ts:268:22, production.

Member ledger `clearTimeout` (39 mentions, including declarations and tests):

- src/compiler/sys.ts:52:18, production.
- src/compiler/sys.ts:585:5, production.
- src/compiler/sys.ts:601:5, production.
- src/compiler/sys.ts:768:13, production.
- src/compiler/sys.ts:972:5, production.
- src/compiler/sys.ts:995:5, production.
- src/compiler/sys.ts:1124:17, production.
- src/compiler/sys.ts:1442:5, production.
- src/compiler/sys.ts:1507:13, production.
- src/compiler/sys.ts:1605:13, production.
- src/compiler/watch.ts:680:9, production.
- src/compiler/watch.ts:680:48, production.
- src/compiler/watchPublic.ts:171:5, production.
- src/compiler/watchPublic.ts:857:14, production.
- src/compiler/watchPublic.ts:863:39, production.
- src/compiler/watchPublic.ts:882:39, production.
- src/compiler/watchPublic.ts:887:18, production.
- src/compiler/tsbuildPublic.ts:2062:53, production.
- src/compiler/tsbuildPublic.ts:2066:23, production.
- src/server/types.ts:34:5, production.
- src/server/utilities.ts:27:23, production.
- src/server/utilities.ts:39:19, production.
- src/server/session.ts:450:48, production.
- src/harness/harnessLanguageService.ts:597:5, test.
- src/lib/dom.generated.d.ts:41718:5, production.
- src/lib/dom.generated.d.ts:44094:18, production.
- src/lib/webworker.generated.d.ts:13824:5, production.
- src/lib/webworker.generated.d.ts:14944:18, production.
- src/testRunner/parallel/host.ts:320:29, test.
- src/testRunner/parallel/worker.ts:61:22, test.
- src/testRunner/parallel/worker.ts:67:22, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:13, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:32, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111:5, test.
- src/testRunner/unittests/tscWatch/watchApi.ts:191:14, test.
- src/testRunner/unittests/tsserver/reloadProjects.ts:86:14, test.
- src/testRunner/unittests/tsserver/session.ts:45:5, test.
- src/tsserver/nodeServer.ts:269:9, production.
- src/tsserver/nodeServer.ts:269:24, production.

Member ledger `pollScheduled` (10 mentions, including declarations and tests):

- src/compiler/sys.ts:252:9, production.
- src/compiler/sys.ts:286:15, production.
- src/compiler/sys.ts:298:19, production.
- src/compiler/sys.ts:310:20, production.
- src/compiler/sys.ts:372:52, production.
- src/compiler/sys.ts:378:47, production.
- src/compiler/sys.ts:464:9, production.
- src/compiler/sys.ts:484:9, production.
- src/compiler/sys.ts:490:37, production.
- src/compiler/sys.ts:491:9, production.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: Keep an opaque, checked host-handle representation paired with the exact registration/cancellation provider; verify callback arity/domain at dispatch. Node Timeout and numeric/custom hosts must coexist. Cost: tag/owner comparisons O(1) per start/stop plus a retained provider/handle pair; Node-specific inspection excludes other hosts. Public check/failure semantics need approval.
- **internal only wrapper**: Thread TimerHost<H,A> through System/WatchHost, queue state, builder state and cancellation; internal storage is H | undefined (dynamic polling queue additionally stores false). Zero runtime cost if specialization and lifetime ownership are proven. Cross-host cancellation is rejected statically. Public erased methods still require a checked ABI shim or sanction; timers also require callback lifetime/ownership support.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S18: src/compiler/tsbuildPublic.ts:2080:5

Stores the same host handle in builder state; needs the shared H contract rather than a Node-only type or ReturnType of an any-returning host.

Source owner: buildNextInvalidatedProjectWorker, src/compiler/tsbuildPublic.ts:2079.

```typescript
function buildNextInvalidatedProjectWorker<T extends BuilderProgram>(state: SolutionBuilderState<T>, changeDetected: boolean)
```

Public/dependent declaration owners:

- System: src/compiler/sys.ts:1396.

```typescript
export interface System {
    setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout?(timeoutId: any): void;
// Full declaration retained in contracts.json.
```

- WatchHost: src/compiler/watchPublic.ts:160.

```typescript
export interface WatchHost {
    setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;
    clearTimeout?(timeoutId: any): void;
// Full declaration retained in contracts.json.
```

Internal callers/references:

- src/compiler/tsbuildPublic.ts:2073:24: direct-call, `buildNextInvalidatedProjectWorker`.

Member ledger `setTimeout` (44 mentions, including declarations and tests):

- src/compiler/sys.ts:51:18, production.
- src/compiler/sys.ts:247:5, production.
- src/compiler/sys.ts:378:68, production.
- src/compiler/sys.ts:460:5, production.
- src/compiler/sys.ts:491:30, production.
- src/compiler/sys.ts:584:5, production.
- src/compiler/sys.ts:600:5, production.
- src/compiler/sys.ts:771:37, production.
- src/compiler/sys.ts:971:5, production.
- src/compiler/sys.ts:994:5, production.
- src/compiler/sys.ts:1055:101, production.
- src/compiler/sys.ts:1059:107, production.
- src/compiler/sys.ts:1123:17, production.
- src/compiler/sys.ts:1441:5, production.
- src/compiler/sys.ts:1506:13, production.
- src/compiler/sys.ts:1604:13, production.
- src/compiler/watch.ts:679:9, production.
- src/compiler/watch.ts:679:46, production.
- src/compiler/watchPublic.ts:169:5, production.
- src/compiler/watchPublic.ts:863:19, production.
- src/compiler/watchPublic.ts:868:57, production.
- src/compiler/watchPublic.ts:882:19, production.
- src/compiler/watchPublic.ts:890:37, production.
- src/compiler/tsbuildPublic.ts:2062:24, production.
- src/compiler/tsbuildPublic.ts:2068:58, production.
- src/server/types.ts:33:5, production.
- src/server/utilities.ts:30:57, production.
- src/server/utilities.ts:63:34, production.
- src/server/session.ts:404:48, production.
- src/server/typingInstallerAdapter.ts:242:19, production.
- src/harness/harnessLanguageService.ts:593:5, test.
- src/lib/dom.generated.d.ts:41731:5, production.
- src/lib/dom.generated.d.ts:44107:18, production.
- src/lib/webworker.generated.d.ts:13837:5, production.
- src/lib/webworker.generated.d.ts:14957:18, production.
- src/testRunner/parallel/host.ts:326:44, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:13, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:468:30, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1103:5, test.
- src/testRunner/unittests/sys/symlinkWatching.ts:20:16, test.
- src/testRunner/unittests/tscWatch/watchApi.ts:190:14, test.
- src/testRunner/unittests/tsserver/session.ts:42:5, test.
- src/tsserver/nodeServer.ts:268:9, production.
- src/tsserver/nodeServer.ts:268:22, production.

Member ledger `clearTimeout` (39 mentions, including declarations and tests):

- src/compiler/sys.ts:52:18, production.
- src/compiler/sys.ts:585:5, production.
- src/compiler/sys.ts:601:5, production.
- src/compiler/sys.ts:768:13, production.
- src/compiler/sys.ts:972:5, production.
- src/compiler/sys.ts:995:5, production.
- src/compiler/sys.ts:1124:17, production.
- src/compiler/sys.ts:1442:5, production.
- src/compiler/sys.ts:1507:13, production.
- src/compiler/sys.ts:1605:13, production.
- src/compiler/watch.ts:680:9, production.
- src/compiler/watch.ts:680:48, production.
- src/compiler/watchPublic.ts:171:5, production.
- src/compiler/watchPublic.ts:857:14, production.
- src/compiler/watchPublic.ts:863:39, production.
- src/compiler/watchPublic.ts:882:39, production.
- src/compiler/watchPublic.ts:887:18, production.
- src/compiler/tsbuildPublic.ts:2062:53, production.
- src/compiler/tsbuildPublic.ts:2066:23, production.
- src/server/types.ts:34:5, production.
- src/server/utilities.ts:27:23, production.
- src/server/utilities.ts:39:19, production.
- src/server/session.ts:450:48, production.
- src/harness/harnessLanguageService.ts:597:5, test.
- src/lib/dom.generated.d.ts:41718:5, production.
- src/lib/dom.generated.d.ts:44094:18, production.
- src/lib/webworker.generated.d.ts:13824:5, production.
- src/lib/webworker.generated.d.ts:14944:18, production.
- src/testRunner/parallel/host.ts:320:29, test.
- src/testRunner/parallel/worker.ts:61:22, test.
- src/testRunner/parallel/worker.ts:67:22, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:13, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:469:32, test.
- src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1111:5, test.
- src/testRunner/unittests/tscWatch/watchApi.ts:191:14, test.
- src/testRunner/unittests/tsserver/reloadProjects.ts:86:14, test.
- src/testRunner/unittests/tsserver/session.ts:45:5, test.
- src/tsserver/nodeServer.ts:269:9, production.
- src/tsserver/nodeServer.ts:269:24, production.

Member ledger `timerToBuildInvalidatedProject` (7 mentions, including declarations and tests):

- src/compiler/tsbuildPublic.ts:424:5, production.
- src/compiler/tsbuildPublic.ts:545:9, production.
- src/compiler/tsbuildPublic.ts:2065:15, production.
- src/compiler/tsbuildPublic.ts:2066:42, production.
- src/compiler/tsbuildPublic.ts:2068:11, production.
- src/compiler/tsbuildPublic.ts:2080:11, production.
- src/compiler/tsbuildPublic.ts:2094:23, production.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: Keep an opaque, checked host-handle representation paired with the exact registration/cancellation provider; verify callback arity/domain at dispatch. Node Timeout and numeric/custom hosts must coexist. Cost: tag/owner comparisons O(1) per start/stop plus a retained provider/handle pair; Node-specific inspection excludes other hosts. Public check/failure semantics need approval.
- **internal only wrapper**: Thread TimerHost<H,A> through System/WatchHost, queue state, builder state and cancellation; internal storage is H | undefined (dynamic polling queue additionally stores false). Zero runtime cost if specialization and lifetime ownership are proven. Cross-host cancellation is rejected statically. Public erased methods still require a checked ABI shim or sanction; timers also require callback lifetime/ownership support.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S19: src/compiler/utilities.ts:7809:17

tryParseJson has a recursive JSON union, but readJsonOrUndefined/readJson promise object and actually return primitive JSON. Loose recovery also comes through the public config:any result. Propagate the truthful union across those owners and review public sanctions.

Source owner: tryParseJson, src/compiler/utilities.ts:7809.

```typescript
export function tryParseJson(text: string): any
```

Public/dependent declaration owners:

- readJsonOrUndefined: src/compiler/utilities.ts:7786.

```typescript
export function readJsonOrUndefined(path: string, hostOrText: { readFile(fileName: string): string | undefined; } | string): object | undefined
```

- readJson: src/compiler/utilities.ts:7804.

```typescript
export function readJson(path: string, host: { readFile(fileName: string): string | undefined; }): object
```

- readConfigFile: src/compiler/commandLineParser.ts:2267.

```typescript
export function readConfigFile(fileName: string, readFile: (path: string) => string | undefined): { config?: any; error?: Diagnostic | undefined; }
```

- parseConfigFileTextToJson: src/compiler/commandLineParser.ts:2277.

```typescript
export function parseConfigFileTextToJson(fileName: string, jsonText: string): { config?: any; error?: Diagnostic | undefined; }
```

- JsonConfigValue: src/compiler/commandLineParser.ts:2426.

```typescript
export type JsonConfigValue = string | number | boolean | null | undefined | JsonConfigValue[] | JsonConfigObject;
```

Internal callers/references:

- src/compiler/utilities.ts:7793:18: direct-call, `tryParseJson`.
- src/compiler/moduleSpecifiers.ts:127:5: reference, `tryParseJson`.
- src/compiler/moduleSpecifiers.ts:1154:82: direct-call, `tryParseJson`.
- src/compiler/moduleSpecifiers.ts:1256:123: direct-call, `tryParseJson`.
- src/services/utilities.ts:371:5: reference, `tryParseJson`.
- src/services/utilities.ts:3600:21: direct-call, `tryParseJson`.

Acceptance observation: 0/301 projects; phases {}. No execution observed; see the focused fixtures/counterexamples for any additional facts.

- **loud boundary check**: Check/represent recursive JSON plus recovery undefined at the boundary, preserving number/string/boolean/null/array roots and existing readJson fallback on falsy values. Cost: O(1) root tag and O(nodes) if a whole result must be validated. Object-only checks reject observed primitive returns. Loose recovery/error diagnostics and getters at public raw-input APIs need separate policies.
- **internal only wrapper**: Give internal parser/result owners JsonConfigValue and propagate it through readJsonOrUndefined/readJson and their consumers. A wrapper calling the proven internal parser can erase to zero JavaScript; calling an any-returning public API still requires a sound check. Public convertToObject/config results require a sanction or a separate checked ABI shim; casts to object are false.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## S20: src/compiler/utilities.ts:8556:25

First allocator callback still returns an incomplete Node header. Node.parent is required publicly but is initialized undefined; other headers also omit required fields. Decide the staged allocator and completed-owner contracts.

Source owner: getNodeConstructor, src/compiler/utilities.ts:8556.

```typescript
() =>
```

Public/dependent declaration owners:

- Node: src/compiler/types.ts:942.

```typescript
export interface Node extends ReadonlyTextRange {
    readonly parent: Node; // Parent node (initialized by binding)
// Full declaration retained in contracts.json.
```

- Type: src/compiler/types.ts:6444.

```typescript
export interface Type {
    symbol: Symbol;                  // Symbol associated with type (if any)
// Full declaration retained in contracts.json.
```

- ObjectAllocator: src/compiler/utilities.ts:8462.

```typescript
export interface ObjectAllocator {
    getNodeConstructor(): new (kind: SyntaxKind, pos: number, end: number) => Node;
    getTokenConstructor(): new <TKind extends SyntaxKind>(kind: TKind, pos: number, end: number) => Token<TKind>;
    getIdentifierConstructor(): new (kind: SyntaxKind.Identifier, pos: number, end: number) => Identifier;
    getPrivateIdentifierConstructor(): new (kind: SyntaxKind.PrivateIdentifier, pos: number, end: number) => PrivateIdentifier;
    getSourceFileConstructor(): new (kind: SyntaxKind.SourceFile, pos: number, end: number) => SourceFile;
    getSymbolConstructor(): new (flags: SymbolFlags, name: __String) => Symbol;
    getTypeConstructor(): new (checker: TypeChecker, flags: TypeFlags) => Type;
    getSignatureConstructor(): new (checker: TypeChecker, flags: SignatureFlags) => AllocatedSignature;
    getSourceMapSourceConstructor(): new (fileName: string, text: string, skipTrivia?: (pos: number) => number) => SourceMapSource;
}
```

- Type: src/compiler/utilities.ts:8492.

```typescript
function Type(this: Type, checker: TypeChecker, flags: TypeFlags): void
```

- Node: src/compiler/utilities.ts:8508.

```typescript
function Node(this: Mutable<Node>, kind: SyntaxKind, pos: number, end: number): void
```

- Node: src/services/types.ts:48.

```typescript
export interface Node {
// Full declaration retained in contracts.json.
```

- Type: src/services/types.ts:109.

```typescript
export interface Type {
// Full declaration retained in contracts.json.
```

Internal callers/references:

No resolved source-symbol reference. For allocator callback/member invocation use the supplemental structural member ledger below; an uncalled public-only API is still an external contract.

Member ledger `getNodeConstructor` (7 mentions, including declarations and tests):

- src/compiler/debug.ts:706:29, production.
- src/compiler/utilities.ts:8463:5, production.
- src/compiler/utilities.ts:8556:5, production.
- src/compiler/factory/baseNodeFactory.ts:58:75, production.
- src/compiler/parser.ts:437:88, production.
- src/compiler/parser.ts:1736:43, production.
- src/services/services.ts:1328:9, production.

Member ledger `getTokenConstructor` (7 mentions, including declarations and tests):

- src/compiler/debug.ts:708:29, production.
- src/compiler/utilities.ts:8464:5, production.
- src/compiler/utilities.ts:8557:5, production.
- src/compiler/factory/baseNodeFactory.ts:54:77, production.
- src/compiler/parser.ts:436:95, production.
- src/compiler/parser.ts:1737:44, production.
- src/services/services.ts:1329:9, production.

Member ledger `getIdentifierConstructor` (7 mentions, including declarations and tests):

- src/compiler/debug.ts:707:29, production.
- src/compiler/utilities.ts:8465:5, production.
- src/compiler/utilities.ts:8558:5, production.
- src/compiler/factory/baseNodeFactory.ts:46:87, production.
- src/compiler/parser.ts:434:110, production.
- src/compiler/parser.ts:1738:49, production.
- src/services/services.ts:1331:9, production.

Member ledger `getPrivateIdentifierConstructor` (6 mentions, including declarations and tests):

- src/compiler/utilities.ts:8466:5, production.
- src/compiler/utilities.ts:8559:5, production.
- src/compiler/factory/baseNodeFactory.ts:50:101, production.
- src/compiler/parser.ts:435:131, production.
- src/compiler/parser.ts:1739:56, production.
- src/services/services.ts:1332:9, production.

Member ledger `getSourceFileConstructor` (7 mentions, including declarations and tests):

- src/compiler/debug.ts:709:29, production.
- src/compiler/utilities.ts:8467:5, production.
- src/compiler/utilities.ts:8560:5, production.
- src/compiler/factory/baseNodeFactory.ts:42:87, production.
- src/compiler/parser.ts:433:110, production.
- src/compiler/parser.ts:1740:49, production.
- src/services/services.ts:1333:9, production.

Member ledger `getTypeConstructor` (5 mentions, including declarations and tests):

- src/compiler/debug.ts:638:49, production.
- src/compiler/utilities.ts:8469:5, production.
- src/compiler/utilities.ts:8562:5, production.
- src/compiler/checker.ts:1510:32, production.
- src/services/services.ts:1335:9, production.

Acceptance observation: 301/301 projects; phases {"return": 7280}. All observed domains are retained under this row in contracts.json; no universal input-domain inference is made.

- **loud boundary check**: A completed-Node check immediately after allocation would reject actual headers: parent starts undefined and text/symbol/id can be absent. A loud read-before-initialization check at the first required-field read could preserve successful paths. Cost: O(1) initialized/tag tests per unproven read plus state tracking; escaping aliases and parent lifetime must be modeled. Arbitrary descriptor/prototype checks are not an ownership proof.
- **internal only wrapper**: Separate AllocatedNode/Header from completed Node, and TypeHeader from Type; factories prove field completion before exposing the completed interface. Public constructor-return any bridges remain refused until that staged owner relationship is approved. Zero runtime overhead only where initialization is proven; late or optional completion needs flags/checks. Do not annotate an incomplete constructor as already complete.
- **leave refused**: Leave this contract refused with its source location and owner chain. Added execution cost: zero, because no native binary is produced; functional cost: this path prevents the corresponding public entrypoint or compiler body from landing.

Decision: **undecided**.

## Fixtures, mutants and limits

Three extracted .a functions are run with source Node, with independent exact expected stdout, stderr and exit. convertToObject preserves primitive/array/object roots; isCompilerOptionsValue demonstrates the shallow invalid-array acceptance; scheduleBuildInvalidatedProject forwards callback arguments and correlates numeric/Node handles. Each has a successfully executing source mutant rejected only by stdout: returnValue false, array predicate false, and cancellation with undefined. fixtures/status.json records current-main compile observations and exact diagnostics. No native success is claimed. These are scoped scout probes, not additions to internal/oracle or shared fixtures_test.go; local counts.md records the three fixtures and three mutants.

The fixture bodies are cut from original owners without statement rewriting. Minimal dependency interfaces and drivers are scaffolding, not approved contracts. The public-any-ts loader resolves only the Node fixture dependency to the freshly built 6.0.3 API. Native checker failure at that module is recorded honestly. Current Adamic ambient timer declarations are missing too. No native memory/ownership proof is made.

Runtime shape inspection uses own property descriptors (without invoking getters), arrays and depth-two summaries, with counts aggregated per process. It returns every value unchanged, and the 301 goldens verify the tested outputs. Proxy traps, getter behavior, object identity visible through stack inspection, unsupported hosts, watch/build mode, dynamic external callers and deeper alias/ownership contracts remain unmeasured. The instrumented timing is not a boundary-check performance estimate. No full gate or upstream full test suite was run.

## Questions for @system_adamic

1. May public convertToObject/config results acquire a sanctioned recursive JSON-plus-recovery type, or must a separate checked public/native entry shim preserve the exact upstream declarations?
2. Should arbitrary JavaScript raw-config input remain supported natively (including invalid values, accessors and cycles), and must checks feed existing diagnostics rather than panic? Where is the observable failure boundary?
3. Should internal raw-to-normalized conversion carry caller-specific generics, while the shallow predicate remains exactly as written? Who owns the predicate's current false promise for list elements?
4. Is TimerHost<H,A> allowed as a sanctioned public specialization, or should native use a checked opaque handle paired with its original provider? What are the cancellation and callback lifetime guarantees for custom hosts?
5. Should allocated AST/type headers have a separate internal type until completion, or use checked read-before-initialization fields? When may a partially initialized header escape to an API/client?
6. Who diagnoses the concrete NodeJS.Timeout annotation reported any by the latent checker? This scout does not treat it as another public any adaptation.

All choices remain undecided. This report sends no message to @system_adamic.
