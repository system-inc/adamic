# Build a parse config host from System

Adamic ruled this adaptation on Nov 24 2026 (#bbssdcy, the scout pass's proposal 20, casts cause 2); she accepts the head after its review and census.

The cast at `watch.ts:220:39` in `parseConfigFileWithSystem`
(`watch.ts:219`) is covered by `docs/rulings.md` rows 39 and 290. Row 83
governs the handler expando. The adapter creates a local host with the
same diagnostic handler and returns the parser result directly. Its
`maybeBind` import uses the existing compiler namespace import.

Round 04 audited the fully adapted source at `26d185e5`.

- (a) `executeCommandLine.ts:654` passes `sys`. The host reaches
  `getParsedCommandLineOfConfigFile` (`commandLineParser.ts:2233`),
  `parseJsonSourceFileConfigFileContent` (:3044), the config worker
  (:3079), and the package resolver (`moduleNameResolver.ts:1802`).
  The attached grep records cover the downstream calls.
- (b) The parse path has no host identity observer.
- (c) The harness can write `sys.exitCode` in its exit callback
  (`testRunner/unittests/tsc/cancellationToken.ts:57` and
  `testRunner/unittests/helpers/tsc.ts:55`). The host never reads it.
- (d) The parse path has no observer of the retained System key.
  `watch.ts:935`, `watchPublic.ts:1033` through :1042,
  `tsbuildPublic.ts:598` and :600, and `program.ts:5135` use their own
  hosts.
- (e) The host forwards `readFile`, `fileExists`,
  `getCurrentDirectory` and `readDirectory` with System as receiver.
  It reads `useCaseSensitiveFileNames` once. `maybeBind` preserves
  the callable or undefined answer for `directoryExists` and
  `realpath`. Their readers test the value before calling.
  `trace` is gated by traceResolution, which the config resolver
  does not set. `getDirectories` belongs to type directive
  enumeration. The optional `getGlobalTypingsCacheLocation` call
  sees undefined because System does not supply that member.

Run `NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-seats-scout24-parse-config-host/test.cjs TREE`
on a full applied tree. The test compiles both snapshots with locked
stock TypeScript and compares parser observations, System calls and
receivers, config results, and diagnostics. It asserts the adapted
construction reads and the original handler key on returning and
throwing paths.
The review cases cover a case insensitive System, a second namespace import, and a package resolved through another realpath spelling.
