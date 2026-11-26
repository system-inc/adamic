// moduleNameResolver.ts 279:17's shape: a result read from a cache Map, spread with three optional arrays, each
// a call's value, so the source is copied before any of them runs (#k7g6f9h).
interface Diagnostic {
    messageText: string;
}
interface ResolvedModuleWithFailedLookupLocations {
    readonly resolvedModule: { resolvedFileName: string } | undefined;
    failedLookupLocations?: string[];
    affectingLocations?: string[];
    resolutionDiagnostics?: Diagnostic[];
    alternateResult?: string;
}
function initializeResolutionField<T>(value: T[]): T[] | undefined {
    return value.length ? value : undefined;
}
function initializeResolutionFieldForReadonlyCache<T>(fromCache: T[] | undefined, value: T[]): T[] | undefined {
    if (!fromCache?.length) return initializeResolutionField(value);
    if (!value.length) return fromCache.slice();
    return [...fromCache, ...value];
}
const cache = new Map<string, ResolvedModuleWithFailedLookupLocations>();
cache.set("a", { resolvedModule: { resolvedFileName: "/a.ts" }, failedLookupLocations: ["/x"] });
cache.set("b", { resolvedModule: undefined, alternateResult: "/alt" });
function fromCache(key: string, failedLookupLocations: string[], affectingLocations: string[], diagnostics: Diagnostic[]): ResolvedModuleWithFailedLookupLocations | undefined {
    const resultFromCache = cache.get(key);
    if (resultFromCache) {
        return {
            ...resultFromCache,
            failedLookupLocations: initializeResolutionFieldForReadonlyCache(resultFromCache.failedLookupLocations, failedLookupLocations),
            affectingLocations: initializeResolutionFieldForReadonlyCache(resultFromCache.affectingLocations, affectingLocations),
            resolutionDiagnostics: initializeResolutionFieldForReadonlyCache(resultFromCache.resolutionDiagnostics, diagnostics),
        };
    }
    return undefined;
}
for (const key of ["a", "b", "c"]) {
    const result = fromCache(key, ["/y"], [], [{ messageText: "m" }]);
    if (result) {
        console.log(Object.keys(result).join(","));
        console.log(JSON.stringify(result));
    } else {
        console.log("missing");
    }
}
