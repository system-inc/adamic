# Family 5 required-member audit

For the 16 remaining candidate forms and C02 alternatives listed in the
request, six qualify, three fail and seven are n/a under the required-member
test. The stopped C03 is included separately as a positive control.
Qualification describes this one test. It does not establish constructor
provenance, writable compatibility in other views, or landing readiness.

## Inputs and method

The audit used the full unmodified Stage 3 apply of
`150b7b5e8ec45cd88e294dbb728322ec41247737` with TypeScript 6.0.3
`050880ce59e30b356b686bd3144efe24f875ebc8`. It used the scout at
`464502ff430b1c88b2e79ebf6b21badcb09e02af` and read Apple’s three
heads for the proposed edits, including the extensions on family5-3.

- [Scout report](https://github.com/system-inc/adamic/blob/464502ff430b1c88b2e79ebf6b21badcb09e02af/stage3/typescript/family5-scout/REPORT.md)
- Apple family5-1 `ec85574494ce7fcfd3f19f4cb9f3ccb16bea5d26`, family5-2 `2f61ad06deb1ac786680d6526b51555af71c47e2`, family5-3 `3f5526c9d3f837322bbf6d7a1b0e58f4801570d0`
- [C06 extensions](https://github.com/system-inc/adamic/blob/3f5526c9d3f837322bbf6d7a1b0e58f4801570d0/stage3/adapt/99-platforms-apple-resolution-results/README.md)
- [C16 extension](https://github.com/system-inc/adamic/blob/3f5526c9d3f837322bbf6d7a1b0e58f4801570d0/stage3/adapt/99-apple-export-results/README.md)

One stock checker program loaded all 262 TypeScript roots under
`src/compiler`, `src/services` and `src/server`, with strict and
exact-optional checking, and enumerated 1,479 interface and class
symbols. For each named owner the audit identified its symbol by its main
declaration in the owning file. It walked `getBaseTypes()` plus the
checker-resolved heritage types, including class `implements`,
transitive bases and intersection arms. It read each subtype’s own
symbol members, declaration locations, `SymbolFlags.Optional` and
`getTypeOfSymbolAtLocation`. Merged declarations share one owner
symbol. As cross-checks, it also looked for required effective members
via `getPropertyOfType`, and for named structural matches via
`checker.isTypeAssignableTo` on the pre-edit types. The results below
describe declared types in those roots. Runtime object literals and
consumer declarations outside those roots are outside the inventory.

## Named owner declarations

“No redeclarations” means the API walk found no interface or class with
its own declaration of that member in the specified heritage graph.

| Candidate and member proposed | Verdict | Every own redeclaration on a subtype |
| --- | --- | --- |
| C01b `TextRange.source?: SourceMapSource` | **qualifies** | `SourceMapRange` declares optional `source?: SourceMapSource` at `src/compiler/types.ts:8333`. The two companion `updated as T` annotations add no member. |
| C02 `Node.getPositionOfLineAndCharacter?` callable property | **fails** | `SourceFile` declares required `getPositionOfLineAndCharacter(line: number, character: number): number` in its services augmentation at `src/services/types.ts:172`. `SourceFileObject` declares required `getPositionOfLineAndCharacter(line: number, character: number, allowEdits?: true): number` at `src/services/services.ts:1133`. |
| C02 alternative, optional method on `SourceFile` | **fails** | `SourceFileObject` at `src/services/services.ts:1133` declares the same required three-parameter method described above. The required two-parameter method at `src/services/types.ts:172` merges into the proposed owner itself. |
| C02 alternative, optional callable property on `SourceFile` | **fails** | `SourceFileObject` at `src/services/services.ts:1133` declares the same required method. The required method at `src/services/types.ts:172` is also a same-owner collision with the proposed property. |
| C06 `ResolvedModuleWithFailedLookupLocations.isInvalidated?: boolean` | **qualifies** | No redeclarations. |
| C06 `ResolvedModuleWithFailedLookupLocations.files?: Set<Path>` | **qualifies** | No redeclarations. |
| C06 `ResolvedTypeReferenceDirectiveWithFailedLookupLocations.isInvalidated?: boolean` | **qualifies** | No redeclarations. |
| C06 `ResolvedTypeReferenceDirectiveWithFailedLookupLocations.files?: Set<Path>` | **qualifies** | No redeclarations. |
| C06 `ResolvedTypeReferenceDirectiveWithFailedLookupLocations.alternateResult?: never` | **qualifies** | No redeclarations. The separate module-result owner already has optional `alternateResult?: string \| undefined` at `src/compiler/types.ts:8131`. Apple’s extension adds the slot only to the type-reference result. |
| C07 `SourceFile.skipTrivia?: (pos: number) => number` | **qualifies** | No redeclarations. `SourceMapSource` declares an optional callback with that signature at `src/compiler/types.ts:8340`. It is a separate owner outside the `SourceFile` hierarchy. |
| C09b `SolutionBuilderHostBase.getGlobalTypingsCacheLocation?: () => string \| undefined` | **qualifies** | No redeclarations. |
| C09b `SolutionBuilderHostBase.onUnRecoverableConfigFileDiagnostic?: (diagnostic: Diagnostic) => void` | **qualifies** | No redeclarations. |
| C10 `CompilerHost.onUnRecoverableConfigFileDiagnostic?: (diagnostic: Diagnostic) => void` | **qualifies** | No redeclarations. |
| C11 `IncrementalCompilationOptions.createProgram?: CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram>` | **qualifies** | No heritage redeclarations. The structural cross-check found `IncrementalProgramOptions<T>` declares optional `createProgram?: CreateProgram<T>` at `src/compiler/watchPublic.ts:138`. |
| C03 control `Type.constraint?: Type` | **fails** | Optional `constraint?: Type` on `TypeParameter` at `src/compiler/types.ts:6897` and `IndexedAccessType` at `src/compiler/types.ts:6932`. Required `constraint: Type` on `SubstitutionType` at `src/compiler/types.ts:7005`. |

`SourceFileObject` explicitly implements `SourceFile` at
`src/services/services.ts:1067`. `JsonSourceFile` and `TsConfigSourceFile`
inherit the required method from the merged `SourceFile` without adding
a declaration. No qualifying named owner had an inherited-required
match. For the C03 control, the required `SubstitutionType` slot explains
the new mutable variance refusal when seen through optional writable
`Type.constraint`.

## Site annotations

These forms change an annotation at its use or construction point. The
anonymous type literals and local `Pick` below introduce no member on a
named interface or class, so they have no owner heritage rows.

| Candidate | Verdict | Annotation |
| --- | --- | --- |
| C02 alternative, site annotation | **n/a** | Type the relation at its use of the `SourceFileLike` view. |
| C08 | **n/a** | Use existing `ErrorOutputContainer` for the five narrow output annotations in `checker.ts`. |
| C12b | **n/a** | Add `"packageJsonScope"` to the `importingSourceFile` parameter’s `Pick<SourceFile, ...>` at `src/compiler/moduleSpecifiers.ts:186`. |
| C13 | **n/a** | Add `types?: never` to the private `getNamedPragmaArguments` return and local dictionary annotations at `src/compiler/parser.ts:10764` and `:10768`. |
| C14b | **n/a** | Add `types?: { value: string; pos: number; end: number; }` and allow `undefined` in the private dictionary index value at `src/compiler/parser.ts:10706`. |
| C15 | **n/a** | Name the existing module and type-reference result interfaces as the `map` result type arguments at `src/compiler/program.ts:1623` and `:1667`. |
| C16 | **n/a** | Add `packageRootPath?: never` and `blockedByExports?: never` to the two private result annotations at `src/compiler/moduleSpecifiers.ts:1020` and `:1122`. |

## Earlier scout variants

C01 without the companion assertions has the same **qualifies** result
as C01b for the required-member test. The scout independently found
strict diagnostics for that form. C14 before the `undefined` index
correction is **n/a** for the same reason as C14b.

The broader C09 `ProgramHost` placement **fails** for
`onUnRecoverableConfigFileDiagnostic`. `WatchCompilerHostOfConfigFile`
inherits a required `onUnRecoverableConfigFileDiagnostic: DiagnosticReporter`
from its other base, `ConfigFileDiagnosticsReporter`, declared at
`src/compiler/commandLineParser.ts:2219`. No subtype owns a further
declaration and the `getGlobalTypingsCacheLocation` query found no
required slot. This inherited collision is consistent with the scout’s
TS2320 result for C09.

The scout excluded C04 and C05 from the selected candidates after its
build-info witness falsified their global diagnostic-marker premise.
The subtype test by itself finds optional
`reportsUnnecessary?: {} \| undefined` on `Diagnostic`
(`src/compiler/types.ts:7288`) and on `ReusableDiagnostic`
(`src/compiler/builder.ts:96`), descendants of the respective related
information interfaces. For `reportsDeprecated`, `Diagnostic` declares
optional `reportsDeprecated?: {} \| undefined` at
`src/compiler/types.ts:7290` and the reusable owner has no subtype
redeclaration. These marker variants **qualify** on requiredness alone.
Their original counterexample still rejects the proposed declarations.
[Scout’s diagnostic witness](https://github.com/system-inc/adamic/blob/464502ff430b1c88b2e79ebf6b21badcb09e02af/stage3/typescript/family5-scout/evidence/diagnostic-hazard.json)
