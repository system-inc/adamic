import { type CompilerOptions, type GetEffectiveTypeRootsHost, type ModuleResolutionHost, type ResolutionMode, type ResolvedModuleWithFailedLookupLocations, type ResolvedProjectReference, type ResolvedTypeReferenceDirectiveWithFailedLookupLocations } from "./_namespaces/ts.js";
export declare function getEffectiveTypeRoots(options: CompilerOptions, host: GetEffectiveTypeRootsHost): string[] | undefined;
/**
 * @param {string | undefined} containingFile - file that contains type reference directive, can be undefined if containing file is unknown.
 * This is possible in case if resolution is performed for directives specified via 'types' parameter. In this case initial path for secondary lookups
 * is assumed to be the same as root directory of the project.
 */
export declare function resolveTypeReferenceDirective(typeReferenceDirectiveName: string, containingFile: string | undefined, options: CompilerOptions, host: ModuleResolutionHost, redirectedReference?: ResolvedProjectReference, cache?: TypeReferenceDirectiveResolutionCache, resolutionMode?: ResolutionMode): ResolvedTypeReferenceDirectiveWithFailedLookupLocations;
/**
 * Given a set of options, returns the set of type directive names
 *   that should be included for this program automatically.
 * This list could either come from the config file,
 *   and/or from enumerating the types root + initial secondary types lookup location given "*" compat wildcard.
 * More type directives might appear in the program later as a result of loading actual source files;
 *   this list is only the set of defaults that are implicitly included.
 */
export declare function getAutomaticTypeDirectiveNames(options: CompilerOptions, host: ModuleResolutionHost): string[];
export interface TypeReferenceDirectiveResolutionCache extends PerDirectoryResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>, NonRelativeNameResolutionCache<ResolvedTypeReferenceDirectiveWithFailedLookupLocations>, PackageJsonInfoCache {
}
export interface ModeAwareCache<T> {
    get(key: string, mode: ResolutionMode): T | undefined;
    set(key: string, mode: ResolutionMode, value: T): this;
    delete(key: string, mode: ResolutionMode): this;
    has(key: string, mode: ResolutionMode): boolean;
    forEach(cb: (elem: T, key: string, mode: ResolutionMode) => void): void;
    size(): number;
}
/**
 * Cached resolutions per containing directory.
 * This assumes that any module id will have the same resolution for sibling files located in the same folder.
 */
export interface PerDirectoryResolutionCache<T> {
    getFromDirectoryCache(name: string, mode: ResolutionMode, directoryName: string, redirectedReference: ResolvedProjectReference | undefined): T | undefined;
    getOrCreateCacheForDirectory(directoryName: string, redirectedReference?: ResolvedProjectReference): ModeAwareCache<T>;
    clear(): void;
    /**
     *  Updates with the current compilerOptions the cache will operate with.
     *  This updates the redirects map as well if needed so module resolutions are cached if they can across the projects
     */
    update(options: CompilerOptions): void;
}
export interface NonRelativeNameResolutionCache<T> {
    getFromNonRelativeNameCache(nonRelativeName: string, mode: ResolutionMode, directoryName: string, redirectedReference: ResolvedProjectReference | undefined): T | undefined;
    getOrCreateCacheForNonRelativeName(nonRelativeName: string, mode: ResolutionMode, redirectedReference?: ResolvedProjectReference): PerNonRelativeNameCache<T>;
    clear(): void;
    /**
     *  Updates with the current compilerOptions the cache will operate with.
     *  This updates the redirects map as well if needed so module resolutions are cached if they can across the projects
     */
    update(options: CompilerOptions): void;
}
export interface PerNonRelativeNameCache<T> {
    get(directory: string): T | undefined;
    set(directory: string, result: T): void;
}
export interface ModuleResolutionCache extends PerDirectoryResolutionCache<ResolvedModuleWithFailedLookupLocations>, NonRelativeModuleNameResolutionCache, PackageJsonInfoCache {
    getPackageJsonInfoCache(): PackageJsonInfoCache;
}
/**
 * Stored map from non-relative module name to a table: directory -> result of module lookup in this directory
 * We support only non-relative module names because resolution of relative module names is usually more deterministic and thus less expensive.
 */
export interface NonRelativeModuleNameResolutionCache extends NonRelativeNameResolutionCache<ResolvedModuleWithFailedLookupLocations>, PackageJsonInfoCache {
    /** @deprecated Use getOrCreateCacheForNonRelativeName */
    getOrCreateCacheForModuleName(nonRelativeModuleName: string, mode: ResolutionMode, redirectedReference?: ResolvedProjectReference): PerModuleNameCache;
}
export interface PackageJsonInfoCache {
    clear(): void;
}
export type PerModuleNameCache = PerNonRelativeNameCache<ResolvedModuleWithFailedLookupLocations>;
export declare function createModuleResolutionCache(currentDirectory: string, getCanonicalFileName: (s: string) => string, options?: CompilerOptions, packageJsonInfoCache?: PackageJsonInfoCache): ModuleResolutionCache;
export declare function createTypeReferenceDirectiveResolutionCache(currentDirectory: string, getCanonicalFileName: (s: string) => string, options?: CompilerOptions, packageJsonInfoCache?: PackageJsonInfoCache): TypeReferenceDirectiveResolutionCache;
export declare function resolveModuleNameFromCache(moduleName: string, containingFile: string, cache: ModuleResolutionCache, mode?: ResolutionMode): ResolvedModuleWithFailedLookupLocations | undefined;
export declare function resolveModuleName(moduleName: string, containingFile: string, compilerOptions: CompilerOptions, host: ModuleResolutionHost, cache?: ModuleResolutionCache, redirectedReference?: ResolvedProjectReference, resolutionMode?: ResolutionMode): ResolvedModuleWithFailedLookupLocations;
export declare function bundlerModuleNameResolver(moduleName: string, containingFile: string, compilerOptions: CompilerOptions, host: ModuleResolutionHost, cache?: ModuleResolutionCache, redirectedReference?: ResolvedProjectReference): ResolvedModuleWithFailedLookupLocations;
export declare function nodeModuleNameResolver(moduleName: string, containingFile: string, compilerOptions: CompilerOptions, host: ModuleResolutionHost, cache?: ModuleResolutionCache, redirectedReference?: ResolvedProjectReference): ResolvedModuleWithFailedLookupLocations;
export declare function classicNameResolver(moduleName: string, containingFile: string, compilerOptions: CompilerOptions, host: ModuleResolutionHost, cache?: NonRelativeModuleNameResolutionCache, redirectedReference?: ResolvedProjectReference): ResolvedModuleWithFailedLookupLocations;
