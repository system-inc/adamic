# Adapted helper signatures and bodies

Exact source text from the finished adapted tree; no candidate edits.

## src/compiler/program.ts — getDiagnosticsHelper

```typescript
function getDiagnosticsHelper<T extends Diagnostic>(
        sourceFile: SourceFile | undefined,
        getDiagnostics: (sourceFile: SourceFile, cancellationToken: CancellationToken | undefined) => readonly T[],
        cancellationToken: CancellationToken | undefined,
    ): readonly T[] {
        if (sourceFile) {
            return sortAndDeduplicateDiagnostics(getDiagnostics(sourceFile, cancellationToken));
        }
        return sortAndDeduplicateDiagnostics(flatMap(program.getSourceFiles(), sourceFile => {
            if (cancellationToken) {
                cancellationToken.throwIfCancellationRequested();
            }
            return getDiagnostics(sourceFile, cancellationToken);
        }));
    }
```

## src/compiler/resolutionCache.ts — resolveNamesWithLocalCache

```typescript
function resolveNamesWithLocalCache<Entry, SourceFile, T extends ResolutionWithFailedLookupLocations, R extends ResolutionWithResolvedFileName>({
        entries,
        containingFile,
        containingSourceFile,
        redirectedReference,
        options,
        perFileCache,
        reusedNames,
        loader,
        getResolutionWithResolvedFileName,
        deferWatchingNonRelativeResolution,
        shouldRetryResolution,
        logChanges,
    }: ResolveNamesWithLocalCacheInput<Entry, SourceFile, T, R>): readonly T[] {
        const path = resolutionHost.toPath(containingFile);
        const resolutionsInFile = perFileCache.get(path) || perFileCache.set(path, createModeAwareCache()).get(path)!;
        const resolvedModules: T[] = [];
        const hasInvalidatedNonRelativeUnresolvedImport = logChanges && isFileWithInvalidatedNonRelativeUnresolvedImports(path);

        // All the resolutions in this file are invalidated if this file wasn't resolved using same redirect
        const program = resolutionHost.getCurrentProgram();
        const oldRedirect = program && program.getRedirectFromSourceFile(containingFile)?.resolvedRef;
        const unmatchedRedirects = oldRedirect ?
            !redirectedReference || redirectedReference.sourceFile.path !== oldRedirect.sourceFile.path :
            !!redirectedReference;

        const seenNamesInFile = createModeAwareCache<true>();
        for (const entry of entries) {
            const name = loader.nameAndMode.getName(entry);
            const mode = loader.nameAndMode.getMode(entry, containingSourceFile, redirectedReference?.commandLine.options || options);
            let resolution = resolutionsInFile.get(name, mode);
            // Resolution is valid if it is present and not invalidated
            if (
                !seenNamesInFile.has(name, mode) &&
                (allModuleAndTypeResolutionsAreInvalidated || unmatchedRedirects || !resolution || resolution.isInvalidated ||
                    // If the name is unresolved import that was invalidated, recalculate
                    (hasInvalidatedNonRelativeUnresolvedImport && !isExternalModuleNameRelative(name) && shouldRetryResolution(resolution)))
            ) {
                const existingResolution = resolution;
                resolution = loader.resolve(name, mode);
                if (resolutionHost.onDiscoveredSymlink && resolutionIsSymlink(resolution)) {
                    resolutionHost.onDiscoveredSymlink();
                }
                resolutionsInFile.set(name, mode, resolution);
                if (resolution !== existingResolution && !resolutionHost.skipWatchingFailedLookups?.(path)) {
                    watchFailedLookupLocationsOfExternalModuleResolutions(name, resolution, path, getResolutionWithResolvedFileName, deferWatchingNonRelativeResolution);
                    if (existingResolution) {
                        stopWatchFailedLookupLocationOfResolution(existingResolution, path, getResolutionWithResolvedFileName);
                    }
                }

                if (logChanges && filesWithChangedSetOfUnresolvedImports && !resolutionIsEqualTo(existingResolution, resolution)) {
                    filesWithChangedSetOfUnresolvedImports.push(path);
                    // reset log changes to avoid recording the same file multiple times
                    logChanges = false;
                }
            }
            else {
                const host = getModuleResolutionHost(resolutionHost);
                if (isTraceEnabled(options, host) && !seenNamesInFile.has(name, mode)) {
                    const resolved = getResolutionWithResolvedFileName(resolution!);
                    trace(
                        host,
                        perFileCache === resolvedModuleNames as unknown ?
                            resolved?.resolvedFileName ?
                                resolved.packageId ?
                                    Diagnostics.Reusing_resolution_of_module_0_from_1_of_old_program_it_was_successfully_resolved_to_2_with_Package_ID_3 :
                                    Diagnostics.Reusing_resolution_of_module_0_from_1_of_old_program_it_was_successfully_resolved_to_2 :
                                Diagnostics.Reusing_resolution_of_module_0_from_1_of_old_program_it_was_not_resolved :
                            resolved?.resolvedFileName ?
                            resolved.packageId ?
                                Diagnostics.Reusing_resolution_of_type_reference_directive_0_from_1_of_old_program_it_was_successfully_resolved_to_2_with_Package_ID_3 :
                                Diagnostics.Reusing_resolution_of_type_reference_directive_0_from_1_of_old_program_it_was_successfully_resolved_to_2 :
                            Diagnostics.Reusing_resolution_of_type_reference_directive_0_from_1_of_old_program_it_was_not_resolved,
                        name,
                        containingFile,
                        resolved?.resolvedFileName,
                        resolved?.packageId && packageIdToString(resolved.packageId),
                    );
                }
            }
            Debug.assert(resolution !== undefined && !resolution.isInvalidated);
            seenNamesInFile.set(name, mode, true);
            resolvedModules.push(resolution);
        }
        reusedNames?.forEach(entry =>
            seenNamesInFile.set(
                loader.nameAndMode.getName(entry),
                loader.nameAndMode.getMode(entry, containingSourceFile, redirectedReference?.commandLine.options || options),
                true,
            )
        );
        if (resolutionsInFile.size() !== seenNamesInFile.size()) {
            // Stop watching and remove the unused name
            resolutionsInFile.forEach((resolution, name, mode) => {
                if (!seenNamesInFile.has(name, mode)) {
                    if (!resolutionHost.skipWatchingFailedLookups?.(path)) {
                        stopWatchFailedLookupLocationOfResolution(resolution, path, getResolutionWithResolvedFileName);
                    }
                    resolutionsInFile.delete(name, mode);
                }
            });
        }
        return resolvedModules;

        function resolutionIsEqualTo(oldResolution: T | undefined, newResolution: T | undefined): boolean {
            if (oldResolution === newResolution) {
                return true;
            }
            if (!oldResolution || !newResolution) {
                return false;
            }
            const oldResult = getResolutionWithResolvedFileName(oldResolution);
            const newResult = getResolutionWithResolvedFileName(newResolution);
            if (oldResult === newResult) {
                return true;
            }
            if (!oldResult || !newResult) {
                return false;
            }
            return oldResult.resolvedFileName === newResult.resolvedFileName;
        }
    }
```

## src/compiler/resolutionCache.ts — watchFailedLookupLocationsOfExternalModuleResolutions

```typescript
function watchFailedLookupLocationsOfExternalModuleResolutions<T extends ResolutionWithFailedLookupLocations, R extends ResolutionWithResolvedFileName>(
        name: string,
        resolution: T,
        filePath: Path,
        getResolutionWithResolvedFileName: GetResolutionWithResolvedFileName<T, R>,
        deferWatchingNonRelativeResolution: boolean,
    ) {
        (resolution.files ??= new Set()).add(filePath);
        if (resolution.files.size !== 1) return;
        if (!deferWatchingNonRelativeResolution || isExternalModuleNameRelative(name)) {
            watchFailedLookupLocationOfResolution(resolution);
        }
        else {
            nonRelativeExternalModuleResolutions.add(resolution);
        }
        const resolved = getResolutionWithResolvedFileName(resolution);
        if (resolved && resolved.resolvedFileName) {
            const key = resolutionHost.toPath(resolved.resolvedFileName);
            let resolutions = resolvedFileToResolution.get(key);
            if (!resolutions) resolvedFileToResolution.set(key, resolutions = new Set());
            resolutions.add(resolution);
        }
    }
```

## src/compiler/resolutionCache.ts — stopWatchFailedLookupLocationOfResolution

```typescript
function stopWatchFailedLookupLocationOfResolution<T extends ResolutionWithFailedLookupLocations, R extends ResolutionWithResolvedFileName>(
        resolution: T,
        filePath: Path,
        getResolutionWithResolvedFileName: GetResolutionWithResolvedFileName<T, R>,
    ) {
        Debug.checkDefined(resolution.files).delete(filePath);
        if (resolution.files!.size) return;
        resolution.files = undefined;
        const resolved = getResolutionWithResolvedFileName(resolution);
        if (resolved && resolved.resolvedFileName) {
            const key = resolutionHost.toPath(resolved.resolvedFileName);
            const resolutions = resolvedFileToResolution.get(key);
            if (resolutions?.delete(resolution) && !resolutions.size) resolvedFileToResolution.delete(key);
        }

        const { failedLookupLocations, affectingLocations, alternateResult } = resolution;
        if (resolutionsWithFailedLookups.delete(resolution)) {
            let removeAtRoot = false;
            if (failedLookupLocations) {
                for (const failedLookupLocation of failedLookupLocations) {
                    removeAtRoot = stopWatchFailedLookupLocation(failedLookupLocation, removeAtRoot);
                }
            }
            if (alternateResult) removeAtRoot = stopWatchFailedLookupLocation(alternateResult, removeAtRoot);
            if (removeAtRoot) removeDirectoryWatcher(rootPath);
        }
        else if (affectingLocations?.length) {
            resolutionsWithOnlyAffectingLocations.delete(resolution);
        }

        if (affectingLocations) {
            for (const affectingLocation of affectingLocations) {
                const watcher = fileWatchesOfAffectingLocations.get(affectingLocation)!;
                watcher.resolutions--;
            }
        }
    }
```

## src/compiler/resolutionCache.ts — removeResolutionsOfFileFromCache

```typescript
function removeResolutionsOfFileFromCache<T extends ResolutionWithFailedLookupLocations, R extends ResolutionWithResolvedFileName>(
        cache: Map<string, ModeAwareCache<T>>,
        filePath: Path,
        getResolutionWithResolvedFileName: GetResolutionWithResolvedFileName<T, R>,
    ) {
        // Deleted file, stop watching failed lookups for all the resolutions in the file
        const resolutions = cache.get(filePath);
        if (resolutions) {
            if (!resolutionHost.skipWatchingFailedLookups?.(filePath)) {
                resolutions.forEach(resolution =>
                    stopWatchFailedLookupLocationOfResolution(
                        resolution,
                        filePath,
                        getResolutionWithResolvedFileName,
                    )
                );
            }
            cache.delete(filePath);
        }
    }
```

## src/compiler/utilitiesPublic.ts — sortAndDeduplicateDiagnostics

```typescript
export function sortAndDeduplicateDiagnostics<T extends Diagnostic>(diagnostics: readonly T[]): SortedReadonlyArray<T> {
    return sortAndDeduplicate<T>(diagnostics, compareDiagnostics, diagnosticsEqualityComparer);
}
```
