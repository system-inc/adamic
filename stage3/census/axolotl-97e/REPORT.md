# axochk: two onError measurements

Main tip and branch base: `80c2d3282e44b527555d59690bb413cc883d2bf2`. Branch: `typescript/axolotl-97e-measure`. Upstream TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. Only this census directory is delivered. No adaptation is written.

Setup used `GOPROXY=https://proxy.golang.org|direct` and `timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1`. A preliminary scratch apply overlapped the submodule fetch; the measured tree was restored afresh with stage3/apply.sh after setup succeeded. Upstream builds and oracle attempts ran serially after setup; all output went to files.

## Measurement (a)

Only WriteFileCallback.onError (types.ts:4584) and CompilerHost.getSourceFile.onError (types.ts:8142) change from `(message: string) => void` to `(message: unknown) => void`. The function-type syntax is retained. getSourceFileByPath is unchanged. `ax97-widen.diff.gz` retains the exact source change.

Upstream locks its build compiler to 5.9.3. The widened scratch package and lockfile pin stock TypeScript 6.0.3 so npm run build and the oracle build use 6.0.3. A matched unmodified control uses the same installed compiler. No tsconfig options are changed. build:tests runs explicitly as well, because npm run build short-circuits if build:compiler fails. Every requested tsc -p runs regardless of prior exits.

| Check | Control exit | Widened exit |
| --- | ---: | ---: |
| npm run build | 0 | 1 |
| npm run build:tests | included in control build | 1 |
| tsc -p src/compiler | 0 | 2 |
| tsc -p src/services | 0 | 0 |
| tsc -p src/server | 0 | 0 |
| tsc -p src/harness | 0 | 0 |
| tsc -p src/testRunner | 0 | 0 |

New unique diagnostics relative to the matched control: **5**. Full continuation messages and project occurrences follow.

```text
src/compiler/program.ts:629:9: error TS2322: Type '((fileName: string, languageVersionOrOptions: ScriptTarget | CreateSourceFileOptions, onError?: ((message: unknown) => void) | undefined, shouldCreateNewSourceFile?: boolean | undefined) => SourceFile | undefined) | undefined' is not assignable to type '((fileName: string, languageVersionOrOptions: ScriptTarget | CreateSourceFileOptions, onError?: ((message: string) => void) | undefined, shouldCreateNewSourceFile?: boolean | undefined) => SourceFile | undefined) | undefined'.
  Type '(fileName: string, languageVersionOrOptions: ScriptTarget | CreateSourceFileOptions, onError?: ((message: unknown) => void) | undefined, shouldCreateNewSourceFile?: boolean | undefined) => SourceFile | undefined' is not assignable to type '(fileName: string, languageVersionOrOptions: ScriptTarget | CreateSourceFileOptions, onError?: ((message: string) => void) | undefined, shouldCreateNewSourceFile?: boolean | undefined) => SourceFile | undefined'.
    Types of parameters 'onError' and 'onError' are incompatible.
      Types of parameters 'message' and 'message' are incompatible.
        Type 'unknown' is not assignable to type 'string'.
```

Seen in: a-build, a-build-tests, a-tsc-compiler.

```text
src/compiler/program.ts:2647:60: error TS2345: Argument of type '((message: string) => void) | undefined' is not assignable to parameter of type '((message: unknown) => void) | undefined'.
  Type '(message: string) => void' is not assignable to type '(message: unknown) => void'.
    Types of parameters 'message' and 'message' are incompatible.
      Type 'unknown' is not assignable to type 'string'.
```

Seen in: a-build, a-build-tests, a-tsc-compiler.

```text
src/compiler/program.ts:3638:157: error TS2322: Type 'unknown' is not assignable to type 'string | number | boolean | readonly string[] | SourceFile | undefined'.
```

Seen in: a-build, a-build-tests, a-tsc-compiler.

```text
src/compiler/utilities.ts:6709:108: error TS2345: Argument of type 'unknown' is not assignable to parameter of type 'string | number | boolean | readonly string[] | SourceFile | undefined'.
```

Seen in: a-build, a-build-tests, a-tsc-compiler.

```text
src/compiler/watchPublic.ts:768:85: error TS2345: Argument of type '((message: string) => void) | undefined' is not assignable to parameter of type '((message: unknown) => void) | undefined'.
  Type '(message: string) => void' is not assignable to type '(message: unknown) => void'.
    Types of parameters 'message' and 'message' are incompatible.
      Type 'unknown' is not assignable to type 'string'.
```

Seen in: a-build, a-build-tests, a-tsc-compiler.

Oracle command: `bash stage3/oracle/run.sh /workspace/scratch/ax97-widened /workspace/scratch/ax97-oracle-a --workers 4 --limit-seconds 1800`.

**Stopped at the oracle build gate. The full suite did not run. Moved baselines are unmeasured, not zero.** Oracle phases: {"install": {"command": ["npm", "ci", "--no-audit", "--no-fund"], "exit": 0, "seconds": 1.767}, "build": {"command": ["npm", "run", "build"], "exit": 1, "seconds": 1.732}}. The widening is not buildable under the requested checks; no workaround or typecheck bypass was introduced.

## Measurement (b): runtime

The counted copy starts from the original adapted tree, without widening. It keeps upstream's lockfile and uses the full oracle procedure from adaptation 91: all runners, no filter, four workers, 1,800-second suite limit. The two temporary counters run at entry to each catch, before the onError guard. The preload records typeof e, typeof e?.message, process ID, stack and active Mocha runnable. Observation files from every process are retained, including preload activation records.

Oracle counts: {"passing": 106366, "failing": 1, "pending": 0}. Phase exits: {"install": 0, "build": 0, "tests": 1}. Baselines moving besides api/typescript.d.ts: none.

| Site (original adapted line) | typeof e | typeof e?.message | Count |
| --- | --- | --- | ---: |
| read-406 | no caught value | no caught value | 0 |
| write-441 | no caught value | no caught value | 0 |

Totals: {"read-406": 0, "write-441": 0}. Every unlisted pair has count zero.

There were zero caught values with non-string messages in this run. Neither catch was exercised. This does not show that caught values were all Error objects. No producing test exists for a nonzero non-string pair.

The isolated smoke check uses the actual built compiler and the same observer, with a separate output directory excluded from suite counts. It drives each catch with Error, string, number, null, undefined, {}, and {message:17}. Both sites recorded exactly seven smoke catches: (object,string) 1, (string,undefined) 1, (number,undefined) 1, (object,undefined) 2, (undefined,undefined) 1 and (object,number) 1. Raw smoke observations and summary are retained.

## Static host review

Locations refer to the adapted scratch tree. The AST inventory retains every body; the I/O occurrence and factory-call logs cross-check aliases and wrapper paths. The table includes conservative candidates even when they bypass the two wrappers.

The paths enter through program.ts:473/476 (createCompilerHostWorker) and watch.ts:757/763 (createCompilerHostFromProgramHost), used by watch and solution builds. The ordinary compiler baseline runner uses fakes.CompilerHost.getSourceFile at fakesHosts.ts:354, which parses directly instead of calling createGetSourceFile. Its writer also writes directly. Thus passing a baseline case does not imply either catch was exercised.

| File:line | Member | Throws or delegation |
| --- | --- | --- |
| src/harness/fakesHosts.ts:56 | readFile | Catches every VFS read exception and returns undefined. |
| src/harness/fakesHosts.ts:66 | writeFile | VFS mkdirpSync/writeFileSync: createIOError returns Error with string message. |
| src/harness/fakesHosts.ts:80 | directoryExists | _getStats catches stat failures; no deliberate throw. |
| src/harness/fakesHosts.ts:85 | createDirectory | VFS mkdirpSync: Error with string message. |
| src/harness/fakesHosts.ts:216 | directoryExists | Delegates to sys.directoryExists. |
| src/harness/fakesHosts.ts:220 | readFile | Delegates to sys.readFile. |
| src/harness/fakesHosts.ts:301 | directoryExists | Delegates to sys.directoryExists. |
| src/harness/fakesHosts.ts:321 | readFile | Delegates to sys.readFile. |
| src/harness/fakesHosts.ts:325 | writeFile | Delegates to fake sys.writeFile, then records output; VFS failures are Error. |
| src/harness/harnessIO.ts:104 | createDirectory | Rethrows native mkdirSync Error unless directory exists; recursive ENOENT handling. |
| src/harness/harnessIO.ts:125 | readFile | Delegates to ts.sys.readFile, which swallows read failures. |
| src/harness/harnessIO.ts:126 | writeFile | Delegates to ts.sys.writeFile; native fs Error. |
| src/harness/harnessIO.ts:131 | directoryExists | Delegates to ts.sys.directoryExists. |
| src/harness/harnessIO.ts:1361 | readFile | In-memory lookup; no deliberate throw. |
| src/harness/harnessLanguageService.ts:288 | readFile | Catches every VFS read exception and returns undefined. |
| src/harness/harnessLanguageService.ts:297 | directoryExists | VFS statSync can throw createIOError, an Error. |
| src/harness/harnessLanguageService.ts:416 | directoryExists | Delegates to fake sys.directoryExists. |
| src/harness/harnessLanguageService.ts:428 | readFile | Delegates to fake sys.readFile. |
| src/harness/harnessLanguageService.ts:532 | readFile | Reads snapshot or delegates to host.readFile; missing snapshots use the VFS reader. |
| src/harness/harnessLanguageService.ts:544 | writeFile | ts.noop; no throw. |
| src/harness/harnessLanguageService.ts:554 | directoryExists | Returns true; no throw. |
| src/harness/harnessLanguageService.ts:565 | createDirectory | ts.notImplemented throws Error("Not implemented"). |
| src/harness/incrementalUtils.ts:476 | readFile | Captures existing reader, not a new implementation. |
| src/harness/incrementalUtils.ts:477 | compilerHost.readFile | Snapshot/disk reader wrapper; delegates to original host; no new deliberate throw. |
| src/testRunner/unittests/config/showConfig.ts:25 | readFile | In-memory config lookup; no deliberate throw. |
| src/testRunner/unittests/customTransforms.ts:19 | readFile | Map lookup; no deliberate throw. |
| src/testRunner/unittests/customTransforms.ts:20 | writeFile | Map output storage; no deliberate throw. |
| src/testRunner/unittests/helpers/baseline.ts:445 | sys.readFile | Delegates to original reader and rewrites build-info version; Debug.assert throws Error on an invariant failure. |
| src/testRunner/unittests/helpers/baseline.ts:461 | sys.writeFile | Rewrites build info, delegates to original writer; no new deliberate throw. |
| src/testRunner/unittests/helpers/solutionBuilder.ts:76 | sys.readFile | Restores original reader. |
| src/testRunner/unittests/helpers/solutionBuilder.ts:78 | sys.writeFile | Restores original writer. |
| src/testRunner/unittests/helpers/solutionBuilder.ts:103 | sys.readFile | Restores original reader. |
| src/testRunner/unittests/helpers/solutionBuilder.ts:104 | sys.writeFile | Restores original writer. |
| src/testRunner/unittests/helpers/tsc.ts:74 | sys.readFile | Read-counting wrapper delegates to original reader; no new throw. |
| src/testRunner/unittests/helpers/tsc.ts:88 | sys.readFile | Restores original reader. |
| src/testRunner/unittests/helpers/tscWatch.ts:230 | compilerHost.writeFile | Replaces compiler host writer with ts.notImplemented: Error("Not implemented"). This replacement bypasses the measuring wrapper. |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1043 | readFile | Virtual file lookup or undefined; no deliberate throw. |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1057 | directoryExists | Virtual folder lookup; no deliberate throw. |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1143 | createDirectory | Debug.assert on base-folder/existence invariants throws Error. |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1155 | writeFile | Throws Error(`Directory not found: ${base}`); otherwise virtual write/modify. |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1385 | host.writeFile | Write-counting wrapper delegates to original virtual writer; no new throw. |
| src/testRunner/unittests/helpers.ts:130 | writeFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/helpers.ts:137 | readFile | Map lookup; no deliberate throw. |
| src/testRunner/unittests/moduleResolution.ts:40 | directoryExists | Set membership; no deliberate throw. |
| src/testRunner/unittests/moduleResolution.ts:51 | readFile | Map lookup; no deliberate throw. |
| src/testRunner/unittests/moduleResolution.ts:397 | writeFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:407 | readFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:510 | writeFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:520 | readFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:928 | readFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:930 | directoryExists | Returns false; no throw. |
| src/testRunner/unittests/moduleResolution.ts:1042 | writeFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:1048 | readFile | Source-file map lookup; no deliberate throw. |
| src/testRunner/unittests/moduleResolution.ts:1081 | writeFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:1087 | readFile | Source text lookup; no deliberate throw. |
| src/testRunner/unittests/moduleResolution.ts:1110 | writeFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/moduleResolution.ts:1116 | readFile | Source text lookup; no deliberate throw. |
| src/testRunner/unittests/programApi.ts:121 | writeFile | Throws new Error("unsupported"). Custom CompilerHost supplies getSourceFile directly, bypassing the read wrapper. |
| src/testRunner/unittests/programApi.ts:129 | readFile | Source text lookup; no deliberate throw. |
| src/testRunner/unittests/publicApi.ts:188 | readFile | AssertionError or Error("cannot read file"); used by validateLocaleAndSetLanguage, outside these catches. |
| src/testRunner/unittests/publicApi.ts:299 | readFile | Bound TestServerHost reader. |
| src/testRunner/unittests/publicApi.ts:301 | directoryExists | Bound TestServerHost directory lookup. |
| src/testRunner/unittests/services/extract/helpers.ts:107 | readFile | ts.notImplemented: Error("Not implemented"). |
| src/testRunner/unittests/services/hostNewLineSupport.ts:20 | readFile | Snapshot text lookup; no deliberate throw. |
| src/testRunner/unittests/services/languageService.ts:50 | readFile | Object lookup; no deliberate throw. |
| src/testRunner/unittests/services/languageService.ts:104 | readFile | Map lookup; no deliberate throw. |
| src/testRunner/unittests/services/languageService.ts:190 | readFile | Delegates to TestServerHost reader. |
| src/testRunner/unittests/services/languageService.ts:200 | readFile | Delegates to TestServerHost reader. |
| src/testRunner/unittests/sys/symlinkWatching.ts:73 | writeFile | Async test helper, not a host implementation; delegates to native sys.writeFile through delayedOp. |
| src/testRunner/unittests/tscWatch/watchApi.ts:561 | readFile | Delegates to TestServerHost reader. |
| src/testRunner/unittests/tscWatch/watchApi.ts:767 | readFile | Delegates to TestServerHost reader. |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts:28 | directoryExists | Tracking-map initializer, not the callback; dynamic wrapper at line 55 delegates to original host. |
| src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts:30 | readFile | Tracking-map initializer, not the callback; dynamic wrapper at line 55 delegates to original host. |
| src/testRunner/unittests/tsserver/configuredProjects.ts:1300 | host.readFile | Returns undefined for config file, otherwise delegates to original reader. |
| src/testRunner/unittests/tsserver/getEditsForFileRename.ts:46 | readFile | Delegates to TestServerHost reader. |
| src/testRunner/unittests/tsserver/plugins.ts:300 | readFile | Bound project reader delegates to its server host. |
| src/testRunner/unittests/tsserver/session.ts:20 | readFile | ts.returnUndefined; no throw. |
| src/testRunner/unittests/tsserver/session.ts:21 | writeFile | ts.noop; no throw. |
| src/testRunner/unittests/tsserver/session.ts:26 | directoryExists | Returns false; no throw. |
| src/testRunner/unittests/tsserver/session.ts:28 | createDirectory | ts.noop; no throw. |

Additional reviewed dynamic wrapper: src/testRunner/unittests/tsserver/cachingFileSystemInformation.ts:55 records calls then invokes originals[prop](f); lines 41 and 47 restore callbacks. It does not introduce a throw value.

VFS throw producers: src/harness/vfsUtil.ts:1379 creates new Error with a string message. readFileSync:658 throws ENOENT/EISDIR/EBADF at 660-662; writeFileSync:674 throws EROFS/EPERM/EISDIR/EBADF at 675, 678, 687-688. Directory creation and path walking use the same createIOError, including ELOOP at 1056. The resolver at 1246 delegates to the harness reader. Native Buffer/path failures are Error/TypeError with string messages. src/compiler/core.ts:1886 notImplemented throws new Error. Debug.fail/assert throw Error.

No reviewed I/O implementation deliberately throws a primitive, nullish value, a plain object without message, or an object with non-string message. fakesHosts.ts:133 throws processExitSentinel from exit, outside these I/O paths. TestServerHost.exit at virtualFileSystemWithWatch.ts:1285 throws Error("System Exit"). OperationCanceledException sites in builder/cancellation tests are cancellation paths, not readFile/writeFile implementations. Primitive throws in evaluation inputs and code-fix source strings are compiled/evaluated test programs, not compiler I/O callbacks. This is a test-suite observation, not a guarantee for arbitrary external hosts.

## Raw evidence and reproduction

All raw outputs are gzip files beside this report. ax97-commands.json.gz records each preparation, build, project check and oracle attempt with cwd, exit and elapsed seconds. ax97-measure.py.gz reproduces their serial order. ax97-inventory.cjs.gz, ax97-host-inventory.json.gz, ax97-throws.json.gz and occurrence logs preserve the static review inputs. ax97-observe.cjs.gz and ax97-instrument.diff.gz describe the temporary counters. setup.log.gz is the complete required setup output. Oracle report, build, tests and baseline outputs retain their full original bytes.
