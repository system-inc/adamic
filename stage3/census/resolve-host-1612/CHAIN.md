# Resolver closure and member observations

The symbol-bound conservative closure contains 1,081 source functions and 3,881
reference/call edges. Every function (file:line:name), every edge and all 28
member accesses are retained in resolver-audit.json. This deliberately includes
dead feature branches, nested callbacks and parser helpers. It resolves lexical
symbols rather than conflating different nested functions named tryResolve.
Node/ECMAScript library declarations have no source body and end that graph.

The host-carrying chain under the fixed entry options is:

1. sys.ts nodeSystem.require -> resolveJSModule (:1678) ->
   tryResolveJSModuleWorker (:1750) -> nodeModuleNameResolverWorker (:1803),
   with features None, Node10 options, JavaScript extensions, no cache/config
   lookup/redirect and no supplied conditions.
2. isTraceEnabled (:123) returns false before host.trace is read. tryResolve
   (:1913) calls tryLoadModuleUsingOptionalResolutionSettings (:1555); paths,
   baseUrl and rootDirs are all absent, so those optional routes return empty.
3. Non-relative names: loadModuleFromNearestNodeModulesDirectory (:2997) ->
   loadModuleFromNearestNodeModulesDirectoryWorker (:3006) -> lookup (:3028)
   -> forEachAncestorDirectoryStoppingAtGlobalCache (:3051), which optionally
   calls getGlobalTypingsCacheLocation (:3056) ->
   loadModuleFromImmediateNodeModulesDirectory (:3064) ->
   loadModuleFromSpecificNodeModulesDirectory (:3091).
4. Package/file work: getPackageJsonInfo (:2445) -> directoryProbablyExists
   (utilities.ts:7819), fileExists (:2469), readJson (:7802) ->
   readJsonOrUndefined (:7785 readFile). The JSON parser tail sees text rather
   than the host. File search uses loadModuleFromFile (:2072),
   loadModuleFromFileNoImplicitExtensions (:2089), tryAddingExtensions (:2131),
   tryFile (:2189), tryFileLookup (:2199, fileExists :2201).
   Directory search uses loadNodeModuleFromDirectory (:2217) and
   loadNodeModuleFromDirectoryWorker (:2489), package main/typesVersions/index
   selection and the same file loader. All helper edges and exact lines are in
   resolver-audit.json; the executable source is the authority for line shifts.
5. Relative names: normalizePathForCJSResolution (:1966) ->
   nodeLoadModuleByRelativeName (:1986) -> the same file/directory machinery.
6. Successful resolution: createResolvedModuleWithFailedLookupLocationsHandlingSymlink
   -> getOriginalAndResolvedFileName -> realPath (:1974, presence :1975, call
   :1979) and arePathsEqual (:511, boolean type/value dispatch :512).
   Package IDs may read peer dependency package JSON and use the same realPath.
7. A failed package lookup can retry with Bundler diagnostic state (:1889-1897),
   but the extension argument is JavaScript & (TypeScript | Declaration) = 0.
   The nearest-directory worker skips both passes (:3017,:3023). The retry
   cannot reach exports or current-directory reads.

## Exhaustive current-directory exclusions

Successful require resolutions do not read getCurrentDirectory. All direct
resolver-module sites were checked:

- :488,:489 getEffectiveTypeRoots: not in the closure.
- :2331 loadEntrypointsFromTargetExports: not in the closure.
- :2582 loadModuleFromSelfNameReference: SelfName bit is absent.
- :2663 loadModuleFromImports: Imports bit is absent.
- :2870,:2898,:2924,:2974 exports/imports target helpers: Exports/Imports bits
  are absent on normal resolution; the diagnostic retry has no lookup pass.
- utilities.ts:6609 getPathsBasePath: guarded by absent options.paths.

This covers successful package/relative/file/directory/typesVersions routes,
not only the failed fixture. Both successful fixture resolutions record zero
current-directory calls; a separate direct forwarder check supplies zero
arguments and asserts nodeSystem as receiver.

getDirectories' only resolver test/call (:820,:825) are in
getAutomaticTypeDirectiveNames, outside the closure. trace's only presence test
(:124) is short-circuited by fixed options (:1755). These are the only three
path exceptions; no fourth exception was needed.

## Host implementation tails and zero-write condition

The concrete Node System tails were also inspected (upstream sys.ts): fileExists
(:1901)/directoryExists (:1905) -> fileSystemEntryExists (:1886) -> statSync
(:1629); readFile (:1787) -> fs.readFileSync and buffer decoding; realpath
(:1917) -> fsRealpath chosen at :1491; getDirectories (:1909) ->
getAccessibleFileSystemEntries. getCurrentDirectory is the memoized process.cwd
closure (:1501). trace and getGlobalTypingsCacheLocation are absent on the
native nodeSystem; the fixture supplies receiver-checking own methods. None
of these tails reads a forwarded function's identity or .length. Buffer/string
lengths and cached directory-exists booleans are different values.

The binding audit records every nodeSystem reference and zero post-construction
member writes. A production-source assignment sweep also finds zero writes to
these System slots. TestRunner helper writes are to virtual/test systems, not
this getNodeSystem object. Replacing the exported sys variable does not mutate
the captured nodeSystem object. Fixture spies/deletions are setup before any
fresh host is constructed; no source slots mutate during a resolution.

All forwarded member reads in the conservative closure were classified as
direct calls, optional calls or explicit presence checks (observers.json).
The source sweep found no computed host-member reads, Object/Reflect host
inspection, member equality comparison, member .length read, or escaping member
value. Identity reads = 0; .length reads = 0. Presence checks remain explicit
in presence.json. The ordinary lane never enters this require path; its path
counts are vacuous and do not establish this proof.
