# Every source bind site and coverage

Stock TypeScript 6.0.3 parsed 81 source files (79 compiler files plus the two tsc entry files), with 10694 syntactic latent candidates before eligibility/deferral. There are 56 real property-call bind sites and 2 local binder maybeBind calls. This is a syntax inventory, not a reach analysis.

Each site without an exact whole-program seam observation is **uncovered**: actual declared target counts, production reach, fallback cause and check outcome are unknown. The two maybeBind calls resolve to binder’s local binding-analysis helper, not core’s generic function-binding helper. Adaptation 99-correctness-binding-views has inlined core maybeBind’s compiler callers into direct bind expressions; the residual generic core helper is upstream core.ts:1523, adapted core.ts:1520.

- `src/compiler/binder.ts:1968:36` — maybeBind_call; unit `src/compiler/binder.ts:1966:9`; UNCOVERED; counts/reach/outcome unknown.
  >                 const maybeBound = maybeBind(left);

- `src/compiler/binder.ts:1984:36` — maybeBind_call; unit `src/compiler/binder.ts:1982:9`; UNCOVERED; counts/reach/outcome unknown.
  >                 const maybeBound = maybeBind(right);

- `src/compiler/builder.ts:1750:30` — bind; unit `src/compiler/builder.ts:1743:5`; UNCOVERED; counts/reach/outcome unknown.
  >                 writeFile || host.writeFile?.bind(host),

- `src/compiler/builder.ts:1838:34` — bind; unit `src/compiler/builder.ts:1764:5`; UNCOVERED; counts/reach/outcome unknown.
  >                     writeFile || host.writeFile?.bind(host),

- `src/compiler/builder.ts:1937:78` — bind; unit `src/compiler/builder.ts:1932:5`; UNCOVERED; counts/reach/outcome unknown.
  >         if (!getEmitDeclarations(state.compilerOptions)) return writeFile || host.writeFile?.bind(host);

- `src/compiler/checker.ts:54284:26` — bind; unit `src/compiler/checker.ts:54280:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getSymlinkCache: host.getSymlinkCache?.bind(host),

- `src/compiler/checker.ts:54295:40` — bind; unit `src/compiler/checker.ts:54280:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getGlobalTypingsCacheLocation: host.getGlobalTypingsCacheLocation?.bind(host),

- `src/compiler/core.ts:1520:12` — bind; unit `src/compiler/core.ts:1519:1`; UNCOVERED; counts/reach/outcome unknown.
  >     return fn?.bind(obj);

- `src/compiler/program.ts:495:21` — bind; unit `src/compiler/program.ts:449:1`; UNCOVERED; counts/reach/outcome unknown.
  >         createHash: system.createHash?.bind(system),

- `src/compiler/program.ts:1605:42` — bind; unit `src/compiler/program.ts:1516:1`; UNCOVERED; counts/reach/outcome unknown.
  >         actualResolveModuleNamesWorker = host.resolveModuleNameLiterals.bind(host);

- `src/compiler/program.ts:1651:58` — bind; unit `src/compiler/program.ts:1516:1`; UNCOVERED; counts/reach/outcome unknown.
  >         actualResolveTypeReferenceDirectiveNamesWorker = host.resolveTypeReferenceDirectiveReferences.bind(host);

- `src/compiler/program.ts:1687:32` — bind; unit `src/compiler/program.ts:1516:1`; UNCOVERED; counts/reach/outcome unknown.
  >         actualResolveLibrary = host.resolveLibrary.bind(host);

- `src/compiler/program.ts:1735:22` — bind; unit `src/compiler/program.ts:1516:1`; UNCOVERED; counts/reach/outcome unknown.
  >     const readFile = host.readFile.bind(host) as typeof host.readFile;

- `src/compiler/program.ts:1951:19` — bind; unit `src/compiler/program.ts:1516:1`; UNCOVERED; counts/reach/outcome unknown.
  >         realpath: host.realpath?.bind(host),

- `src/compiler/program.ts:1957:40` — bind; unit `src/compiler/program.ts:1516:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getGlobalTypingsCacheLocation: host.getGlobalTypingsCacheLocation?.bind(host),

- `src/compiler/program.ts:2627:23` — bind; unit `src/compiler/program.ts:2599:5`; UNCOVERED; counts/reach/outcome unknown.
  >             realpath: host.realpath?.bind(host),

- `src/compiler/program.ts:2633:25` — bind; unit `src/compiler/program.ts:2599:5`; UNCOVERED; counts/reach/outcome unknown.
  >             createHash: host.createHash?.bind(host),

- `src/compiler/program.ts:2635:20` — bind; unit `src/compiler/program.ts:2599:5`; UNCOVERED; counts/reach/outcome unknown.
  >             trace: host.trace?.bind(host),

- `src/compiler/program.ts:5115:26` — bind; unit `src/compiler/program.ts:5104:1`; UNCOVERED; counts/reach/outcome unknown.
  >         directoryExists: directoryStructureHost.directoryExists?.bind(directoryStructureHost),

- `src/compiler/program.ts:5116:25` — bind; unit `src/compiler/program.ts:5104:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getDirectories: directoryStructureHost.getDirectories?.bind(directoryStructureHost),

- `src/compiler/program.ts:5117:19` — bind; unit `src/compiler/program.ts:5104:1`; UNCOVERED; counts/reach/outcome unknown.
  >         realpath: directoryStructureHost.realpath?.bind(directoryStructureHost),

- `src/compiler/tsbuildPublic.ts:301:16` — bind; unit `src/compiler/tsbuildPublic.ts:294:1`; UNCOVERED; counts/reach/outcome unknown.
  >     host.now = system.now?.bind(system); // For testing

- `src/compiler/tsbuildPublic.ts:437:46` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveModuleNameLiterals = host.resolveModuleNameLiterals?.bind(host);

- `src/compiler/tsbuildPublic.ts:438:60` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveTypeReferenceDirectiveReferences = host.resolveTypeReferenceDirectiveReferences?.bind(host);

- `src/compiler/tsbuildPublic.ts:439:35` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveLibrary = host.resolveLibrary?.bind(host);

- `src/compiler/tsbuildPublic.ts:440:39` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveModuleNames = host.resolveModuleNames?.bind(host);

- `src/compiler/tsbuildPublic.ts:441:51` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveTypeReferenceDirectives = host.resolveTypeReferenceDirectives?.bind(host);

- `src/compiler/tsbuildPublic.ts:442:45` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.getModuleResolutionCache = host.getModuleResolutionCache?.bind(host);

- `src/compiler/tsbuildPublic.ts:499:16` — bind; unit `src/compiler/tsbuildPublic.ts:428:1`; UNCOVERED; counts/reach/outcome unknown.
  >         write: host.trace?.bind(host),

- `src/compiler/watch.ts:676:20` — bind; unit `src/compiler/watch.ts:672:1`; UNCOVERED; counts/reach/outcome unknown.
  >         watchFile: system.watchFile?.bind(system) || returnNoopFileWatcher,

- `src/compiler/watch.ts:677:25` — bind; unit `src/compiler/watch.ts:672:1`; UNCOVERED; counts/reach/outcome unknown.
  >         watchDirectory: system.watchDirectory?.bind(system) || returnNoopFileWatcher,

- `src/compiler/watch.ts:678:21` — bind; unit `src/compiler/watch.ts:672:1`; UNCOVERED; counts/reach/outcome unknown.
  >         setTimeout: system.setTimeout?.bind(system) || noop,

- `src/compiler/watch.ts:679:23` — bind; unit `src/compiler/watch.ts:672:1`; UNCOVERED; counts/reach/outcome unknown.
  >         clearTimeout: system.clearTimeout?.bind(system) || noop,

- `src/compiler/watch.ts:760:32` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getDefaultLibLocation: host.getDefaultLibLocation?.bind(host),

- `src/compiler/watch.ts:773:16` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         trace: host.trace?.bind(host),

- `src/compiler/watch.ts:774:26` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         directoryExists: directoryStructureHost.directoryExists?.bind(directoryStructureHost),

- `src/compiler/watch.ts:775:25` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getDirectories: directoryStructureHost.getDirectories?.bind(directoryStructureHost),

- `src/compiler/watch.ts:776:19` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         realpath: host.realpath?.bind(host),

- `src/compiler/watch.ts:777:33` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getEnvironmentVariable: host.getEnvironmentVariable?.bind(host) || (() => ""),

- `src/compiler/watch.ts:778:21` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         createHash: host.createHash?.bind(host),

- `src/compiler/watch.ts:779:24` — bind; unit `src/compiler/watch.ts:753:1`; UNCOVERED; counts/reach/outcome unknown.
  >         readDirectory: host.readDirectory?.bind(host),

- `src/compiler/watch.ts:855:19` — bind; unit `src/compiler/watch.ts:842:1`; UNCOVERED; counts/reach/outcome unknown.
  >         realpath: system.realpath?.bind(system),

- `src/compiler/watch.ts:856:33` — bind; unit `src/compiler/watch.ts:842:1`; UNCOVERED; counts/reach/outcome unknown.
  >         getEnvironmentVariable: system.getEnvironmentVariable?.bind(system),

- `src/compiler/watch.ts:860:21` — bind; unit `src/compiler/watch.ts:842:1`; UNCOVERED; counts/reach/outcome unknown.
  >         createHash: system.createHash?.bind(system),

- `src/compiler/watch.ts:863:14` — bind; unit `src/compiler/watch.ts:842:1`; UNCOVERED; counts/reach/outcome unknown.
  >         now: system.now?.bind(system),

- `src/compiler/watchPublic.ts:124:23` — bind; unit `src/compiler/watchPublic.ts:122:1`; UNCOVERED; counts/reach/outcome unknown.
  >     host.createHash = system.createHash?.bind(system);

- `src/compiler/watchPublic.ts:505:56` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.useSourceOfProjectReferenceRedirect = host.useSourceOfProjectReferenceRedirect?.bind(host);

- `src/compiler/watchPublic.ts:528:46` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveModuleNameLiterals = host.resolveModuleNameLiterals?.bind(host);

- `src/compiler/watchPublic.ts:529:39` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveModuleNames = host.resolveModuleNames?.bind(host);

- `src/compiler/watchPublic.ts:531:50` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         compilerHost.resolveModuleNameLiterals = resolutionCache.resolveModuleNameLiterals.bind(resolutionCache);

- `src/compiler/watchPublic.ts:533:60` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveTypeReferenceDirectiveReferences = host.resolveTypeReferenceDirectiveReferences?.bind(host);

- `src/compiler/watchPublic.ts:534:51` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >     compilerHost.resolveTypeReferenceDirectives = host.resolveTypeReferenceDirectives?.bind(host);

- `src/compiler/watchPublic.ts:536:64` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         compilerHost.resolveTypeReferenceDirectiveReferences = resolutionCache.resolveTypeReferenceDirectiveReferences.bind(resolutionCache);

- `src/compiler/watchPublic.ts:539:9` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         resolutionCache.resolveLibrary.bind(resolutionCache) :

- `src/compiler/watchPublic.ts:540:9` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         host.resolveLibrary.bind(host);

- `src/compiler/watchPublic.ts:542:9` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         host.getModuleResolutionCache?.bind(host) :

- `src/compiler/watchPublic.ts:548:9` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         host.hasInvalidatedResolutions?.bind(host) || returnTrue :

- `src/compiler/watchPublic.ts:551:9` — bind; unit `src/compiler/watchPublic.ts:419:1`; UNCOVERED; counts/reach/outcome unknown.
  >         host.hasInvalidatedLibResolutions?.bind(host) || returnTrue :

