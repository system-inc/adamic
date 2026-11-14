# Complete full-file finding deltas

Scratch roots and inserted lines are normalized; locations below refer to baseline lines. All phases are included.

## resolution-readonly

Raw 221 → 217; unique 133 → 129; new 0, gone 4, changed 0.

### new

None.

### gone

- `src/compiler/resolutionCache.ts:1053:13` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1056:17` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1471:9` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:774:17` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

### changed

None.

## resolution-exact

Raw 221 → 184; unique 133 → 121; new 2, gone 14, changed 0.

### new

- `src/compiler/resolutionCache.ts:1470:5` — lowering / SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped)

- `src/compiler/resolutionCache.ts:579:1` — lowering / Boundary: function signature or prologue failed; continued at next declaration

### gone

- `src/compiler/resolutionCache.ts:1002:17` — refusal_scan / Refused: a value of type CompilerHost | ResolutionCacheHost seen as ModuleResolutionHost, whose method useCaseSensitiveFileNames becomes a writable field: a write of boolean | (() => boolean) | undefined would replace a method of type () => boolean

- `src/compiler/resolutionCache.ts:1053:13` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1056:17` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1471:42` — refusal_scan / Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedModuleWithFailedLookupLocations>>, which can write string where Path is read

- `src/compiler/resolutionCache.ts:1471:9` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1472:42` — refusal_scan / Refused: a value of type Map<Path, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>> seen as Map<string, ModeAwareCache<CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations>>, which can write string where Path is read

- `src/compiler/resolutionCache.ts:1688:13` — lowering / Boundary: failed statement; state rolled back; continued at next statement or declaration

- `src/compiler/resolutionCache.ts:1688:13` — lowering / Refused: a generic instantiated at void whose body reads the value: mutateMap's U = void (read from the signature), read at src/compiler/utilities.ts:8253:32 as a destructured value

- `src/compiler/resolutionCache.ts:633:5` — lowering / Boundary: failed statement; state rolled back; continued at next statement or declaration

- `src/compiler/resolutionCache.ts:642:5` — lowering / Boundary: failed statement; state rolled back; continued at next statement or declaration

- `src/compiler/resolutionCache.ts:654:9` — lowering / NotYet: a generic function as a value

- `src/compiler/resolutionCache.ts:774:17` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:990:16` — refusal_scan / Refused: optional property isInvalidated in ResolutionWithFailedLookupLocations absent from structural source ResolvedTypeReferenceDirectiveWithFailedLookupLocations, which can hide fields

- `src/compiler/resolutionCache.ts:997:27` — refusal_scan / Refused: optional property isInvalidated in CachedResolvedTypeReferenceDirectiveWithFailedLookupLocations absent from structural source ResolvedTypeReferenceDirectiveWithFailedLookupLocations, which can hide fields

### changed

None.

## resolution-result-view

Raw 221 → 217; unique 133 → 129; new 0, gone 4, changed 0.

### new

None.

### gone

- `src/compiler/resolutionCache.ts:1053:13` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1056:17` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:1471:9` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

- `src/compiler/resolutionCache.ts:774:17` — refusal_scan / Refused: a value of type ResolvedModuleFull seen as ResolutionWithResolvedFileName, which can write string | undefined where string is read

### changed

None.
