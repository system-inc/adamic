## Downstream uses of the fifteen a* bound values

**6 of the 15 a* sites have a downstream callback, forwarding store/copy, rebinding or indirect call:** **B10** `program.ts:1957`, **B22** `watch.ts:678`, **B26** `watch.ts:774`, **B27** `watch.ts:775`, **B47** `watchPublic.ts:548`, and **B48** `watchPublic.ts:551`. **None of the identified caller sites passes more function arguments than that origin’s M.** In particular, no selected bound value is used as an array-method callback, invoked through `.apply`, or invoked through `Reflect` in this audit.

Counting non-invoking presence/comparison guards as well gives a **conservative 11/15**: `B03` (`checker.ts:54295`), `B10` (`program.ts:1957`), `B14` (`program.ts:5115`), `B22` (`watch.ts:678`), `B24` (`watch.ts:760`), `B26` (`watch.ts:774`), `B27` (`watch.ts:775`), `B35` (`watch.ts:863`), `B47` (`watchPublic.ts:548`), `B48` (`watchPublic.ts:551`), `B49` (`tsbuildPublic.ts:301`). The six-site headline excludes those guards and excludes each bind’s initial storage; otherwise initial storage alone would count every site. Guards are listed separately below. The guard-inclusive count is a may-flow count: the shared `ProgramHost.now`/solution-host path includes the later overwrite at B49, so it is not proof that the B35 wrapper itself executes at `host.now()`.

This follows the bound result and its receiver-only rebindings, not every unrelated function with the same member name. It follows field reads, local aliases, shorthand/destructured cache records, source-call arguments/returns and the source `copyProperties`/`assign`/`clone`/`extend` helpers. A fresh closure which calls the bound function is inspected at that call; its own identity is not silently equated with the captured bound function. Returned host containers are followed to their local consumers; their initial return is not counted as another function-valued store. The scope is the same 79 source roots, with no harness/server/services code and no invented outside caller bodies.

### All fifteen origins

| Origin | Initial storage of the bound result | M | Downstream non-direct uses | Ordinary typed calls found (argument counts) | More than M at an identified caller? |
|---|---|---:|---|---|---|
| B02 `checker.ts:54284` | `ModuleSpecifierResolutionHost.getSymlinkCache` | 0 | none found | `moduleSpecifiers.ts:736` (0), `moduleSpecifiers.ts:811` (0) | No identified case |
| B03 `checker.ts:54295` | `ModuleSpecifierResolutionHost.getGlobalTypingsCacheLocation` | 0 | none found | `moduleNameResolver.ts:3056` (0), `moduleSpecifiers.ts:1235` (0) | No identified case |
| B10 `program.ts:1957` | `Program.getGlobalTypingsCacheLocation` | 0 | U001, U020 | `moduleNameResolver.ts:3056` (0), `moduleSpecifiers.ts:1235` (0) | No identified case |
| B14 `program.ts:5115` | `ParseConfigFileHost.directoryExists` | 1 | none found | `utilities.ts:7819` (1) | No identified case |
| B15 `program.ts:5116` | `ParseConfigFileHost.getDirectories` | 1 | none found | none found | No identified case |
| B22 `watch.ts:678` | `WatchHost.setTimeout` | 3 | U037, U042 | `tsbuildPublic.ts:2074` (3), `watchPublic.ts:872` (3), `watchPublic.ts:900` (3) | No identified case |
| B24 `watch.ts:760` | `CompilerHost.getDefaultLibLocation` | 0 | none found | `program.ts:1582` (0) | No identified case |
| B26 `watch.ts:774` | `CompilerHost.directoryExists` | 1 | U012, U014, U015, U021, U024, U027, U029, U031, U045, U046 | `moduleNameResolver.ts:824` (1), `program.ts:460` (1), `utilities.ts:7819` (1) | No identified case |
| B27 `watch.ts:775` | `CompilerHost.getDirectories` | 1 | U022, U028, U030, U032 | `moduleNameResolver.ts:825` (1) | No identified case |
| B35 `watch.ts:863` | `ProgramHost<T>.now` | 0 | none found | `tsbuildPublic.ts:205` (0) | No identified case |
| B46 `watchPublic.ts:542` | `compilerHost.getModuleResolutionCache` | 0 | none found | `program.ts:1606` (0), `program.ts:1625` (0) | No identified case |
| B47 `watchPublic.ts:548` | `customHasInvalidatedResolutions` | 1 | U043 | `resolutionCache.ts:749` (1) | No identified case |
| B48 `watchPublic.ts:551` | `customHasInvalidLibResolutions` | 1 | U044 | `resolutionCache.ts:754` (1) | No identified case |
| B49 `tsbuildPublic.ts:301` | `host.now` | 0 | none found | `tsbuildPublic.ts:205` (0) | No identified case |
| B55 `tsbuildPublic.ts:442` | `compilerHost.getModuleResolutionCache` | 0 | none found | `program.ts:1606` (0), `program.ts:1625` (0) | No identified case |

The storage contracts and declaration locations are recorded in `evidence/a-star-uses.json.gz`. For B46/B55 the destination is `compilerHost.getModuleResolutionCache`; for B49 it is the solution-builder host’s `now`. B47/B48 first enter the two local `customHasInvalidated...` variables rather than a field.

### Every identified downstream callback/store/copy/rebind/indirect-call use

Each row is an API-discovered read of a bound identity or its receiver-only rebound descendant. `.call`’s first argument is the receiver and is excluded from the function-argument count. Rebinding here supplies no fixed function arguments. Copies/stores do not invoke the value; the caller column names the downstream invocation evidence. The source alias/call fields are retained for all origins, including the nine with no such use.

| Use | Origin(s) | Source quote | Form | Can this use introduce more arguments than M? |
|---|---|---|---|---|
| U001 | B10 | `src/compiler/checker.ts:54295:40` `host.getGlobalTypingsCacheLocation?.bind(host)` | bind | No invocation here; 0 fixed function arguments. Descendant calls remain at the original M. |
| U012 | B26 | `src/compiler/program.ts:528:11` `originalDirectoryExists = host.directoryExists` | alias-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U014 | B26 | `src/compiler/program.ts:610:30` `originalDirectoryExists.call(host, directory)` | call | No: exactly 1 function argument(s); no spread. |
| U015 | B26 | `src/compiler/program.ts:627:9` `originalDirectoryExists` | field-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U020 | B10 | `src/compiler/program.ts:2636:13` `getGlobalTypingsCacheLocation: program.getGlobalTypingsCacheLocation` | field-store | No invocation: alias/field movement. The getter remains a zero-argument getter. |
| U021 | B26 | `src/compiler/program.ts:4895:11` `originalDirectoryExists = host.compilerHost.directoryExists` | alias-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U022 | B27 | `src/compiler/program.ts:4896:11` `originalGetDirectories = host.compilerHost.getDirectories` | alias-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U024 | B26 | `src/compiler/program.ts:4909:17` `originalDirectoryExists.call(host.compilerHost, path)` | call | No: exactly 1 function argument(s); no spread. |
| U027 | B26 | `src/compiler/program.ts:4941:81` `originalDirectoryExists.call(host.compilerHost, path)` | call | No: exactly 1 function argument(s); no spread. |
| U028 | B27 | `src/compiler/program.ts:4942:17` `originalGetDirectories.call(host.compilerHost, path)` | call | No: exactly 1 function argument(s); no spread. |
| U029 | B26 | `src/compiler/program.ts:4957:9` `host.compilerHost.directoryExists = originalDirectoryExists` | store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U030 | B27 | `src/compiler/program.ts:4958:9` `host.compilerHost.getDirectories = originalGetDirectories` | store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U031 | B26 | `src/compiler/program.ts:5115:26` `directoryStructureHost.directoryExists?.bind(directoryStructureHost)` | bind | No invocation here; 0 fixed function arguments. Descendant calls remain at the original M. |
| U032 | B27 | `src/compiler/program.ts:5116:25` `directoryStructureHost.getDirectories?.bind(directoryStructureHost)` | bind | No invocation here; 0 fixed function arguments. Descendant calls remain at the original M. |
| U037 | B22 | `src/compiler/tsbuildPublic.ts:326:5` `copyProperties(host, watchHost)` | object-copy | No invocation: own-property copy; the copied timer is called with 3 arguments (M=3). See core.ts:1513 below. |
| U042 | B22 | `src/compiler/watch.ts:873:5` `copyProperties(result, createWatchHost(system, reportWatchStatus))` | object-copy | No invocation: own-property copy; the copied timer is called with 3 arguments (M=3). See core.ts:1513 below. |
| U043 | B47 | `src/compiler/watchPublic.ts:628:77` `resolutionCache.createHasInvalidatedResolutions(customHasInvalidatedResolutions, customHasInvalidLibResolutions)` | callback-or-argument | No: the source callback consumer invokes it with one argument: `resolutionCache.ts:749` `customHasInvalidatedResolutions(path)` (1 argument). |
| U044 | B48 | `src/compiler/watchPublic.ts:628:77` `resolutionCache.createHasInvalidatedResolutions(customHasInvalidatedResolutions, customHasInvalidLibResolutions)` | callback-or-argument | No: the source callback consumer invokes it with one argument: `resolutionCache.ts:754` `customHasInvalidatedLibResolutions(libFileName)` (1 argument). |
| U045 | B26 | `src/compiler/watchPublic.ts:632:13` `originalDirectoryExists` | destructure-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U046 | B26 | `src/compiler/watchPublic.ts:660:9` `compilerHost.directoryExists = originalDirectoryExists` | store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |

`copyProperties` performs the actual dynamic read and store at **`src/compiler/core.ts:1513`**, quoted exactly:

```typescript
first[id] = second[id];
```

For B22, the two call sites above use `second` equal to the watch host, and `id = "setTimeout"` copies that bound value. The receiver/label timer calls are `tsbuildPublic.ts:2074`, `watchPublic.ts:872`, and `watchPublic.ts:900`; each passes exactly `(callback, timeoutDelay, t1)`, three function arguments, matching M=3. Being the *first/target* argument of `copyProperties` does not read an existing `now` or directory method; the key set comes from `second`. This prevents counting unrelated existing target members as spread uses.

For B26, the cache forwarding is included, not stopped at the first alias: `changeCompilerHostLikeToUseCache` captures `originalDirectoryExists`, returns it as a shorthand field, the watch caller destructures that record, and later restores `compilerHost.directoryExists`. Its `originalDirectoryExists.call(host, directory)` passes one argument. The originals record is modeled **per helper call**: the solution builder’s separate call on its `ProgramHost` does not inherit the watch compiler host’s bound method merely because both calls use the same helper body. Any B26 attribution to that separate solution-host restoration is excluded.

B47/B48 are passed as callbacks to `resolutionCache.createHasInvalidatedResolutions` at `watchPublic.ts:628`. The actual implementation is `resolutionCache.ts:739`; it captures them in new one-parameter predicates and invokes them at `:749` and `:754` as `customHasInvalidatedResolutions(path)` and `customHasInvalidatedLibResolutions(libFileName)`. It supplies no array index, array object, rest tail or spread. The later predicate-field aliases in `program.ts:1603` and `:1684` refer to those fresh wrapper predicates, not directly to the original two bound-function identities; their calls into the original identities are the two explicit one-argument calls just listed.

### Presence/comparison reads (no function invocation)

| Read | Origin(s) in the may-flow audit | Quoted value read | Can add function arguments? |
|---|---|---|---|
| U002 `src/compiler/moduleNameResolver.ts:820:9` | B26 | `host.directoryExists` | No: presence/comparison only |
| U003 `src/compiler/moduleNameResolver.ts:820:33` | B27 | `host.getDirectories` | No: presence/comparison only |
| U009 `src/compiler/moduleSpecifiers.ts:1235:40` | B03, B10 | `host.getGlobalTypingsCacheLocation` | No: presence/comparison only |
| U013 `src/compiler/program.ts:605:9` | B26 | `originalDirectoryExists` | No: presence/comparison only |
| U016 `src/compiler/program.ts:1582:32` | B24 | `host.getDefaultLibLocation` | No: presence/comparison only |
| U023 `src/compiler/program.ts:4904:9` | B26 | `originalDirectoryExists` | No: presence/comparison only |
| U025 `src/compiler/program.ts:4937:9` | B27 | `originalGetDirectories` | No: presence/comparison only |
| U026 `src/compiler/program.ts:4941:54` | B26 | `originalDirectoryExists` | No: presence/comparison only |
| U035 `src/compiler/tsbuildPublic.ts:205:12` | B35, B49 | `host.now` | No: presence/comparison only |
| U038 `src/compiler/tsbuildPublic.ts:2061:10` | B22 | `hostWithWatch.setTimeout` | No: presence/comparison only |
| U040 `src/compiler/utilities.ts:7819:13` | B14, B26 | `host.directoryExists` | No: presence/comparison only |
| U047 `src/compiler/watchPublic.ts:862:14` | B22 | `host.setTimeout` | No: presence/comparison only |
| U049 `src/compiler/watchPublic.ts:887:14` | B22 | `host.setTimeout` | No: presence/comparison only |

### Cross-checks and coverage

The API pass scanned **257367** member/alias read expressions, **52426** source calls and **51** object-spread syntax nodes. It retained 50 attributed read/use records, 38 non-direct callable member-name reads for the independent surface cross-check, and 546 conservative host-carrier argument records. The latter are an over-approximation used to inspect copy/index/Reflect paths, not additional function invocation sites. Callable identities are constrained by checker types; generic callback contracts stay open. Object carriers stay conservative. The originals cache record is modeled separately at each call, including its shorthand return fields and destructured aliases; the audit reports any other function-valued literal/array carrier it cannot model. There are zero such unresolved carriers in this result.

The independent non-direct member-read cross-check leaves only three non-guard/non-bind same-name cases outside the attributed identity set: native `Date.now` at `performanceCore.ts:103` (a different symbol and number-valued contract), and the two fresh invalidation-wrapper aliases at `program.ts:1603`/`:1684` explained above. Source index/Reflect carrier reads were inspected; the relevant dynamic function-valued read is `second[id]` in the explicit copy helper. No attributed `.apply`, `Reflect` invocation, array-method callback or untyped direct call remains in the records. All recorded direct calls have fixed argument lists of size ≤ M, and both callback consumers have source bodies with one-argument invocation lists.

**Limit:** this is a compiler-API static may-use audit, not execution tracing or a theorem of exact path feasibility. It is flow/context insensitive except for the cache originals record, retains source field/call contracts and explicit unknown host boundaries, and cannot constrain code outside the 79 roots. The six positive sites and their quoted uses are reviewable; “none found” is scoped to this audit. In particular, the result does not assert that outside JavaScript callers can never supply extra arguments. No native tsc run or production lowering change was used.

Reproduce the additional audit with `node stage3/census/bind-sites-static/uses-run.cjs ADAPTED_TREE NEW_OUTPUT_DIRECTORY`, then render with `python3 stage3/census/bind-sites-static/write-uses.py NEW_OUTPUT_DIRECTORY`. The renderer checks the 15 origins, six affected sites, fixed direct-call vectors, callback body vectors and absence of an attributed apply/Reflect/array-callback use before writing this section. `write-report.py` incorporates the saved section without changing the original measurements.
