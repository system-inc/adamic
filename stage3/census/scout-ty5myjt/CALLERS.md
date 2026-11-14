# All src callers

Locations are in the finished baseline adapted tree. `call-inventory.json` records
all matched call expressions, arguments and stock resolved signatures;
`harness-call.json` separately checks the harness signature (the 79 compiler roots
do not import the harness). The all-src search includes harness and tests.
For resolution helpers, **T is the cache wrapper, R is the resolved result**;
the widened `resolvedFileName` belongs to R.

| Helper | Caller site | T | R |
| --- | --- | --- | --- |
| getDiagnosticsHelper | program.ts:2796 getSyntacticDiagnostics | DiagnosticWithLocation | — |
| getDiagnosticsHelper | program.ts:2800 getSemanticDiagnostics | Diagnostic | — |
| getDiagnosticsHelper | program.ts:2829 getDeclarationDiagnostics | DiagnosticWithLocation | — |
| stopWatchFailedLookupLocationOfResolution | resolutionCache.ts:774 cleanupLibResolutionWatching callback | CachedResolvedModuleWithFailedLookupLocations | ResolvedModuleFull |
| watchFailedLookupLocationsOfExternalModuleResolutions | resolutionCache.ts:903 resolveNamesWithLocalCache | enclosing T | enclosing R |
| stopWatchFailedLookupLocationOfResolution | resolutionCache.ts:905 resolveNamesWithLocalCache | enclosing T | enclosing R |
| stopWatchFailedLookupLocationOfResolution | resolutionCache.ts:955 resolveNamesWithLocalCache callback | enclosing T | enclosing R |
| watchFailedLookupLocationsOfExternalModuleResolutions | resolutionCache.ts:1053 resolveLibrary | CachedResolvedModuleWithFailedLookupLocations | ResolvedModuleFull |
| stopWatchFailedLookupLocationOfResolution | resolutionCache.ts:1056 resolveLibrary | CachedResolvedModuleWithFailedLookupLocations | ResolvedModuleFull |
| stopWatchFailedLookupLocationOfResolution | resolutionCache.ts:1445 removeResolutionsOfFileFromCache callback | enclosing T | enclosing R |
| removeResolutionsOfFileFromCache | resolutionCache.ts:1471 removeResolutionsOfFile | CachedResolvedModuleWithFailedLookupLocations | ResolvedModuleFull |
| removeResolutionsOfFileFromCache | resolutionCache.ts:1472 removeResolutionsOfFile | CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations | ResolvedTypeReferenceDirective |
| watchFailedLookupLocationsOfExternalModuleResolutions | harness/incrementalUtils.ts:339 createExpectedResolution | CachedResolvedModuleWithFailedLookupLocations & CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations (ExpectedResolution) | { resolvedFileName: string \| undefined } |

The two instantiations of `resolveNamesWithLocalCache` propagate to its nested
watch/stop calls: resolutionCache.ts:990 in
`resolveTypeReferenceDirectiveReferences` uses
`T = ResolvedTypeReferenceDirectiveWithFailedLookupLocations`,
`R = ResolvedTypeReferenceDirective`; resolutionCache.ts:1019 in
`resolveModuleNameLiterals` uses `T = ResolvedModuleWithFailedLookupLocations`,
`R = ResolvedModuleFull`. Those paths are additional module/type-reference users,
not additional syntactic call sites of the inner helpers.

No single exact caller type covers diagnostics, nor does ResolvedModuleFull cover
all result types of the resolution helpers. The exact-type candidate files are
negative controls, not proposals to drop semantic/type-reference/harness callers.
