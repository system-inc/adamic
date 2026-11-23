# Readonly directory enumeration on DirectoryStructureHost

Adamic ruled on this type-only adaptation for family 80 on #bbssdcy on
Nov 22 2026 at about 22:58Z. Her Nov 23 answers carried the ruling forward
member by member, allowing the test runner's `readDirectory` writes and, at
07:21Z, its `getDirectories` writes on the same terms. Compiler accepts
under the widened split after its review and census.
The adapter adds `readonly` to `DirectoryStructureHost.getDirectories` and
`DirectoryStructureHost.readDirectory` at `src/compiler/watchUtilities.ts:74`
and `:75`, preserving both optional indexed-method types and their calls.

The two relations are at the same expression, `cachedDirectoryStructureHost || host`,
at `src/compiler/watchPublic.ts:464:60`, in `createWatchProgram`. The
source types are `CachedDirectoryStructureHost | WatchCompilerHostOfFilesAndCompilerOptionsOrConfigFile<T>`
and `CachedDirectoryStructureHost | undefined`.

The owner carries `@internal`. This applies the writable-location invariance
and readonly covariance ruling in `docs/rulings.md:95`, the public declaration
scope in row 397 and the internal adaptation shape in row 398. The adapter
requires the internal interface and the exact text of both properties
before inserting either modifier.

## getDirectories checker reference trace

Stock TypeScript 6.0.3 searched 713 source roots under `src` in the fully
adapted base. Its reference groups contained the declaration, seven reads or
calls, the separate `CachedDirectoryStructureHost` method declaration and
one construction site. The construction site is the shorthand
`getDirectories,` at `watchUtilities.ts:130:9` in the fresh object returned by
`createCachedDirectoryStructureHost`. The language service marks the
initializer as a write reference.

The three same-name assignments resolve to these other property declarations.

| Assignment site | Receiver type | Property declaration |
| --- | --- | --- |
| `src/compiler/program.ts:4951:27` | `CompilerHost` | `ModuleResolutionHost.getDirectories` at `src/compiler/types.ts:8031:5` |
| `src/compiler/program.ts:4969:27` | `CompilerHost` | `ModuleResolutionHost.getDirectories` at `src/compiler/types.ts:8031:5` |
| `src/compiler/sys.ts:1627:73` | `NodeModuleResolutionHost` | `NodeModuleResolutionHost.getDirectories` at `src/compiler/sys.ts:1521:13` |

```ts
host.compilerHost.getDirectories = path =>
host.compilerHost.getDirectories = originalGetDirectories;
if (nodeSystem.getDirectories) moduleResolutionHost.getDirectories = (path: string) => nodeSystem.getDirectories(path);
```


In `src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts`,
`createLoggerTrackingHostCalls(host: TestServerHost)` at line 24 dispatches
`"getDirectories"` to the single-argument wrapper.

```ts
// Line 29
getDirectories: setCallsTrackingWithSingleArgFn("getDirectories"),
// Line 41
host[prop] = originals[prop];
// Line 47
forEachHostProperty(prop => host[prop] = storage[prop]);
// Line 55
(host as any)[prop] = (f: string) => {
```

Line 55 writes through `as any`. Its key type is `CalledMapsWithSingleArg`,
which contains `"getDirectories"` at line 20. Line 54 binds the original, and
the wrapper records each path and calls that bound original at lines 56 and
57. Lines 41 and 47 write through the `TestServerHost` view, whose
`getDirectories` callable is declared at
`src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1062:5`.
Their key type is `CalledMaps`, and `forEachHostProperty` enumerates
`"getDirectories"` at line 85. The swaps put the original back for
verification and restore the wrapper afterward. The fixture passes the
same host to the session, and directory caching reads the callable through
`Readonly<DirectoryStructureHost>`. Outside these test-host writes, the
trace found zero assignments replacing the internal callable.

Adamic's Nov 23 07:21Z answer on #bbssdcy applies the same exception as
for `readDirectory` because the modifier is type-only, the JavaScript is
byte-identical, and the swaps run as before in test-runner code Adamic
does not compile.


## readDirectory checker reference trace

Stock TypeScript 6.0.3 searched 713 roots. Of 10 references in the
internal property's reference groups, the only write flag is the
`readDirectory,` initializer at `watchUtilities.ts:131:9` in the fresh
cached-host object. The computed-key audit also found two bookkeeping
assignments and three assignments to the callable declared on
`TestServerHost` at `virtualFileSystemWithWatch.ts:1072:5`. The latter
occur in `src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts`.

```ts
// Line 41
host[prop] = originals[prop];
// Line 47
forEachHostProperty(prop => host[prop] = storage[prop]);
// Line 65
(host as any)[prop] = (f: string, arg1?: U, arg2?: V, arg3?: W, arg4?: X) => {
```

Line 65 installs the wrapper through `as any`. Lines 41 and 47 swap
and restore it through the `TestServerHost` view. The fixture passes the
same object to the session, and directory caching reads it through
`Readonly<DirectoryStructureHost>`. Adamic explicitly allowed this
test-runner case. Outside these test-host writes, the trace found zero
assignments replacing the internal callable.
