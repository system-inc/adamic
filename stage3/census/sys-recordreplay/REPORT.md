# sys.ts recordreplay: complete lane measurement

Task b5w3ycg. Base origin/main `e8f3723bd41abcebedabfcc7546fdf104a9bbf08`; pinned TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`. Measurement only; no adaptation or host-axiom row is proposed.

All **398 lane units** ran exactly once in 20 pieces. Every ordinary identity, API, baseline and count guard passed. Totals are **106366 passes, one expected failure, zero skips**. The underlying Node exit 1 for this expected failure is retained; its ordinary shard guard exits 0 after approving that expected result. The sole sanctioned failure is `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.

| Expression result | Evaluations |
| --- | ---: |
| Truthy | 0 |
| Falsy | 398 |
| Total | 398 |

| Actual process.recordreplay value | Evaluations |
| --- | ---: |
| undefined | 398 |

The actual value above was captured at the original property read. All 398 receivers were the actual Node process, and none had an own recordreplay property. Raw rows also record whether the receiver is Node's process and whether it has an own property. Reconciliation records these distributions without assuming a value for an unobserved branch. This establishes the measured lane's observations; Compiler judges whether an adaptation is established shape or whether a host-axiom row with Language is appropriate.

## Method and validation

The sole scratch source change replaces exactly `!!(process as any).recordreplay` with `!!(globalThis as any).__sysRecordReplay((process as any).recordreplay, process)`. It preserves the preceding logical-or short circuit and performs one property read at the same point. The observer records the actual value and returns that exact value before Boolean conversion. It is installed in the Node global and the three VM context entry paths. Observer-ready events distinguish an active observer with zero evaluations from missing activation. Missing instrumentation throws instead of returning a quiet fallback.

Independent seeds exercise truthy and falsy values, a getter read exactly once, a preceding true short circuit, and all three VM entry paths. Seed observations are stored separately and excluded from the oracle count. The baseline and instrumented source manifests compare all 746 source files; only this expression in sys.ts differs. The complete prepared source/build manifest stays unchanged across every piece and after the final unit.

The unchanged stage3/lane shard.py and shard-plan.json execute the whole run.sh oracle inventory. Four concurrent units use the same 90 second shard deadline as the record-table measurement; the upstream 40 second test timeout, selected tests and expected results are unchanged. Each piece retains raw process observations, child exits, full ordinary shard reports and logs, and is committed and plainly pushed. The largest piece took 102.917 seconds. No command reached the 30 minute no-output stop, and no unit was cut short.

| Proof | Result |
| --- | --- |
| Required environment setup and pinned source | All accepted children exit 0 |
| Full unchanged apply --build-product | Child exit 0 |
| Instrumented compiler and test harness build | Children exit 0 |
| Observer seeds | Child exit 0 |
| generate_shards --check | Child exit 0 |
| Every ordinary lane shard | Child exit 0; all expected identity/API/baseline/count guards pass |
| Public declarations and typescript.internal.d.ts | Byte-identical to the unchanged baseline |
| Independent reconciliation | Reopens every raw stream and shard report, verifies archive hashes, exact unit inventory and complete counts |
| Source/build input freeze | Exact hashes unchanged through all 398 units |

The fourth piece's first packaging check exited 2 because the exact raw owner artifact used CRLF. Retaining those exact bytes as getNodeSystem.ts.txt.gz resolved the whitespace check; its LF quote below is complete. The already completed piece was pushed, then execution resumed at unit 81. No unit was rerun and no oracle result changed. That check failure and its successful correction are retained alongside the actual child exits.

COUNT.json is the cumulative measured ledger; reconciliation.json independently rebuilds its counts from raw archives and reopens all 398 ordinary reports. Every piece archive has its adjacent SHA-256. preparation.tar.gz retains setup, full apply, build, seed and command/child-exit logs; both source manifests; the prepared source/build hash manifest; exact sys.ts input bytes; the unchanged lane plan; and seed traces. Its internal SHA256SUMS and preparation.sha256 verify those bytes. Scripts, provenance.json and declarations.json identify the exact measured work. No result is submitted to Loom.

## Who pins the enclosing owner

**93-builtin-imports and 99-wasm-crypto-import both pin the whole getNodeSystem function**, including the recordreplay initializer, with before and after SHA-256 values in their proofs/targets.json files. Their adapt.cjs files validate the whole owner. owner-pins.json retains both manifests. The final full-apply function below has raw CRLF SHA-256 `9d0f66d64f8d6c9a713b73b514429f485254c918e26791dae9540331a4aa357d`, exactly the crypto adaptation's after-owner pin. getNodeSystem.ts.txt.gz preserves those raw bytes; this full quote normalizes only line endings to LF.

```ts
function getNodeSystem(): System {
        const nativePattern = /^native |^\([^)]+\)$|^(?:internal[\\/]|[\w\s]+(?:\.js)?$)/;
        let activeSession: import("inspector").Session | "stopping" | undefined;
        let profilePath = "./profile.cpuprofile";

        const isMacOs = process.platform === "darwin";
        const isLinuxOrMacOs = process.platform === "linux" || isMacOs;

        const statSyncOptions = { throwIfNoEntry: false } as const;

        const platform: string = _os.platform();
        const useCaseSensitiveFileNames = isFileSystemCaseSensitive();
        const fsRealpath = !!_fs.realpathSync.native ? process.platform === "win32" ? fsRealPathHandlingLongPath : _fs.realpathSync.native : _fs.realpathSync;

        // If our filename is "sys.js", then we are executing unbundled on the raw tsc output.
        // In that case, simulate a faked path in the directory where a bundle would normally
        // appear (e.g. the directory containing lib.*.d.ts files).
        //
        // Note that if we ever emit as files like cjs/mjs, this check will be wrong.
        const executingFilePath = __filename.endsWith("sys.js") ? _path.join(_path.dirname(__dirname), "__fake__.js") : __filename;

        const fsSupportsRecursiveFsWatch = process.platform === "win32" || isMacOs;
        const getCurrentDirectory = memoize(() => process.cwd());
        const { watchFile, watchDirectory } = createSystemWatchFunctions({
            pollingWatchFileWorker: fsWatchFileWorker,
            getModifiedTime,
            setTimeout,
            clearTimeout,
            fsWatchWorker,
            useCaseSensitiveFileNames,
            getCurrentDirectory,
            fileSystemEntryExists,
            // Node 4.0 `fs.watch` function supports the "recursive" option on both OSX and Windows
            // (ref: https://github.com/nodejs/node/pull/2649 and https://github.com/Microsoft/TypeScript/issues/4643)
            fsSupportsRecursiveFsWatch,
            getAccessibleSortedChildDirectories: path => getAccessibleFileSystemEntries(path).directories,
            realpath,
            tscWatchFile: process.env.TSC_WATCHFILE,
            useNonPollingWatchers: !!process.env.TSC_NONPOLLING_WATCHER,
            tscWatchDirectory: process.env.TSC_WATCHDIRECTORY,
            inodeWatching: isLinuxOrMacOs,
            fsWatchWithTimestamp: isMacOs,
            sysLog,
        });
        const nodeSystem: System = {
            args: process.argv.slice(2),
            newLine: _os.EOL,
            useCaseSensitiveFileNames,
            write(s: string): void {
                process.stdout.write(s);
            },
            getWidthOfTerminal() {
                return process.stdout.columns ?? 0;
            },
            writeOutputIsTTY() {
                return !!process.stdout.isTTY;
            },
            readFile,
            writeFile,
            watchFile,
            watchDirectory,
            preferNonRecursiveWatch: !fsSupportsRecursiveFsWatch,
            resolvePath: path => _path.resolve(path),
            fileExists,
            directoryExists,
            getAccessibleFileSystemEntries,
            createDirectory(directoryName: string) {
                if (!nodeSystem.directoryExists(directoryName)) {
                    // Wrapped in a try-catch to prevent crashing if we are in a race
                    // with another copy of ourselves to create the same directory
                    try {
                        _fs.mkdirSync(directoryName);
                    }
                    catch (e) {
                        if (errorCode(e) !== "EEXIST") {
                            // Failed for some other reason (access denied?); still throw
                            throw e;
                        }
                    }
                }
            },
            getExecutingFilePath() {
                return executingFilePath;
            },
            getCurrentDirectory,
            getDirectories,
            getEnvironmentVariable(name: string) {
                return process.env[name] || "";
            },
            readDirectory,
            getModifiedTime,
            setModifiedTime,
            deleteFile,
            createHash: createSHA256Hash,
            createSHA256Hash,
            getMemoryUsage() {
                if (global.gc) {
                    global.gc();
                }
                return process.memoryUsage().heapUsed;
            },
            getFileSize(path) {
                const stat = statSync(path);
                if (stat?.isFile()) {
                    return stat.size;
                }
                return 0;
            },
            exit(exitCode?: number): undefined {
                disableCPUProfiler(() => process.exit(exitCode));
            },
            enableCPUProfiler,
            disableCPUProfiler,
            cpuProfilingEnabled: () => !!activeSession || contains(process.execArgv, "--cpu-prof") || contains(process.execArgv, "--prof"),
            realpath,
            debugMode: !!process.env.NODE_INSPECTOR_IPC || !!process.env.VSCODE_INSPECTOR_OPTIONS || some(process.execArgv, arg => /^--(?:inspect|debug)(?:-brk)?(?:=\d+)?$/i.test(arg)) || !!(process as any).recordreplay,
            tryEnableSourceMapsForHost() {
                try {
                    require("source-map-support").install();
                }
                catch {
                    // Could not enable source maps.
                }
            },
            setTimeout,
            clearTimeout,
            clearScreen: () => {
                process.stdout.write("\x1B[2J\x1B[3J\x1B[H");
            },
            setBlocking: () => {
                const handle = (process.stdout as any)?._handle as { setBlocking?: (value: boolean) => void; };
                if (handle && handle.setBlocking) {
                    handle.setBlocking(true);
                }
            },
            base64decode: input => Buffer.from(input, "base64").toString("utf8"),
            base64encode: input => Buffer.from(input).toString("base64"),
            require: (baseDir, moduleName) => {
                try {
                    const modulePath = resolveJSModule(moduleName, baseDir, nodeSystem);
                    return { module: require(modulePath), modulePath, error: undefined };
                }
                catch (error) {
                    return { module: undefined, modulePath: undefined, error };
                }
            },
        };
        return nodeSystem;

        /** Calls fs.statSync, returning undefined if any errors are thrown */
        function statSync(path: string): import("fs").Stats | undefined {
            // throwIfNoEntry is available in Node 14.17 and above, which matches our supported range.
            try {
                return _fs.statSync(path, statSyncOptions);
            }
            catch {
                // This should never happen as we are passing throwIfNoEntry: false,
                // but guard against this just in case (e.g. a polyfill doesn't check this flag).
                return undefined;
            }
        }

        /**
         * Uses the builtin inspector APIs to capture a CPU profile
         * See https://nodejs.org/api/inspector.html#inspector_example_usage for details
         */
        function enableCPUProfiler(path: string, cb: () => void) {
            if (activeSession) {
                cb();
                return false;
            }
            const inspector: typeof import("inspector") = require("inspector");
            if (!inspector || !inspector.Session) {
                cb();
                return false;
            }
            const session = new inspector.Session();
            session.connect();

            session.post("Profiler.enable", () => {
                session.post("Profiler.start", () => {
                    activeSession = session;
                    profilePath = path;
                    cb();
                });
            });
            return true;
        }

        /**
         * Strips non-TS paths from the profile, so users with private projects shouldn't
         * need to worry about leaking paths by submitting a cpu profile to us
         */
        function cleanupPaths(profile: import("inspector").Profiler.Profile) {
            let externalFileCounter = 0;
            const remappedPaths = new Map<string, string>();
            const normalizedDir = normalizeSlashes(_path.dirname(executingFilePath));
            // Windows rooted dir names need an extra `/` prepended to be valid file:/// urls
            const fileUrlRoot = `file://${getRootLength(normalizedDir) === 1 ? "" : "/"}${normalizedDir}`;
            for (const node of profile.nodes) {
                if (node.callFrame.url) {
                    const url = normalizeSlashes(node.callFrame.url);
                    if (containsPath(fileUrlRoot, url, useCaseSensitiveFileNames)) {
                        node.callFrame.url = getRelativePathToDirectoryOrUrl(fileUrlRoot, url, fileUrlRoot, createGetCanonicalFileName(useCaseSensitiveFileNames), /*isAbsolutePathAnUrl*/ true);
                    }
                    else if (!nativePattern.test(url)) {
                        node.callFrame.url = (remappedPaths.has(url) ? remappedPaths : remappedPaths.set(url, `external${externalFileCounter}.js`)).get(url)!;
                        externalFileCounter++;
                    }
                }
            }
            return profile;
        }

        function disableCPUProfiler(cb: () => void) {
            if (activeSession && activeSession !== "stopping") {
                const s = activeSession;
                activeSession.post("Profiler.stop", (err, { profile }) => {
                    if (!err) {
                        if (statSync(profilePath)?.isDirectory()) {
                            profilePath = _path.join(profilePath, `${(new Date()).toISOString().replace(/:/g, "-")}+P${process.pid}.cpuprofile`);
                        }
                        try {
                            _fs.mkdirSync(_path.dirname(profilePath), { recursive: true });
                        }
                        catch {
                            // do nothing and ignore fallible fs operation
                        }
                        _fs.writeFileSync(profilePath, JSON.stringify(cleanupPaths(profile))!);
                    }
                    activeSession = undefined;
                    s.disconnect();
                    cb();
                });
                activeSession = "stopping";
                return true;
            }
            else {
                cb();
                return false;
            }
        }

        function isFileSystemCaseSensitive(): boolean {
            // win32\win64 are case insensitive platforms
            if (platform === "win32" || platform === "win64") {
                return false;
            }
            // If this file exists under a different case, we must be case-insensitve.
            return !fileExists(swapCase(__filename));
        }

        /** Convert all lowercase chars to uppercase, and vice-versa */
        function swapCase(s: string): string {
            return s.replace(/\w/g, ch => {
                const up = ch.toUpperCase();
                return ch === up ? ch.toLowerCase() : up;
            });
        }

        function fsWatchFileWorker(fileName: string, callback: FileWatcherCallback, pollingInterval: number): FileWatcher {
            _fs.watchFile(fileName, { persistent: true, interval: pollingInterval }, fileChanged);
            let eventKind: FileWatcherEventKind;
            return {
                close: () => _fs.unwatchFile(fileName, fileChanged),
            };

            function fileChanged(curr: import("fs").Stats, prev: import("fs").Stats) {
                // previous event kind check is to ensure we recongnize the file as previously also missing when it is restored or renamed twice (that is it disappears and reappears)
                // In such case, prevTime returned is same as prev time of event when file was deleted as per node documentation
                const isPreviouslyDeleted = +prev.mtime === 0 || eventKind === FileWatcherEventKind.Deleted;
                if (+curr.mtime === 0) {
                    if (isPreviouslyDeleted) {
                        // Already deleted file, no need to callback again
                        return;
                    }
                    eventKind = FileWatcherEventKind.Deleted;
                }
                else if (isPreviouslyDeleted) {
                    eventKind = FileWatcherEventKind.Created;
                }
                // If there is no change in modified time, ignore the event
                else if (+curr.mtime === +prev.mtime) {
                    return;
                }
                else {
                    // File changed
                    eventKind = FileWatcherEventKind.Changed;
                }
                callback(fileName, eventKind, curr.mtime);
            }
        }

        function fsWatchWorker(
            fileOrDirectory: string,
            recursive: boolean,
            callback: FsWatchCallback,
        ) {
            // Node 4.0 `fs.watch` function supports the "recursive" option on both OSX and Windows
            // (ref: https://github.com/nodejs/node/pull/2649 and https://github.com/Microsoft/TypeScript/issues/4643)
            return _fs.watch(
                fileOrDirectory,
                fsSupportsRecursiveFsWatch ?
                    { persistent: true, recursive: !!recursive } : { persistent: true },
                callback,
            );
        }

        function readFile(fileName: string, _encoding?: string): string | undefined {
            let buffer: Buffer;
            try {
                buffer = _fs.readFileSync(fileName);
            }
            catch {
                return undefined;
            }
            let len = buffer.length;
            if (len >= 2 && buffer[0]! === 0xFE && buffer[1]! === 0xFF) {
                // Big endian UTF-16 byte order mark detected. Since big endian is not supported by node.js,
                // flip all byte pairs and treat as little endian.
                len &= ~1; // Round down to a multiple of 2
                for (let i = 0; i < len; i += 2) {
                    const temp = buffer[i]!;
                    buffer[i] = buffer[i + 1]!;
                    buffer[i + 1] = temp;
                }
                return buffer.toString("utf16le", 2);
            }
            if (len >= 2 && buffer[0]! === 0xFF && buffer[1]! === 0xFE) {
                // Little endian UTF-16 byte order mark detected
                return buffer.toString("utf16le", 2);
            }
            if (len >= 3 && buffer[0]! === 0xEF && buffer[1]! === 0xBB && buffer[2]! === 0xBF) {
                // UTF-8 byte order mark detected
                return buffer.toString("utf8", 3);
            }
            // Default is UTF-8 with no byte order mark
            return buffer.toString("utf8");
        }

        function writeFile(fileName: string, data: string, writeByteOrderMark?: boolean): void {
            // If a BOM is required, emit one
            if (writeByteOrderMark) {
                data = byteOrderMarkIndicator + data;
            }

            let fd: number | undefined;

            try {
                fd = _fs.openSync(fileName, "w");
                _fs.writeSync(fd, data, /*position*/ undefined, "utf8");
            }
            finally {
                if (fd !== undefined) {
                    _fs.closeSync(fd);
                }
            }
        }

        function getAccessibleFileSystemEntries(path: string): FileSystemEntries {
            try {
                const entries = _fs.readdirSync(path || ".", { withFileTypes: true });
                const files: string[] = [];
                const directories: string[] = [];
                for (const dirent of entries) {
                    // withFileTypes is not supported before Node 10.10.
                    const entry = typeof dirent === "string" ? dirent : dirent.name;

                    // This is necessary because on some file system node fails to exclude
                    // "." and "..". See https://github.com/nodejs/node/issues/4002
                    if (entry === "." || entry === "..") {
                        continue;
                    }

                    let stat: import("fs").Stats | import("fs").Dirent | undefined;
                    if (typeof dirent === "string" || dirent.isSymbolicLink()) {
                        const name = combinePaths(path, entry);

                        stat = statSync(name);
                        if (!stat) {
                            continue;
                        }
                    }
                    else {
                        stat = dirent;
                    }

                    if (stat.isFile()) {
                        files.push(entry);
                    }
                    else if (stat.isDirectory()) {
                        directories.push(entry);
                    }
                }
                files.sort();
                directories.sort();
                return { files, directories };
            }
            catch {
                return emptyFileSystemEntries;
            }
        }

        function readDirectory(path: string, extensions?: readonly string[], excludes?: readonly string[], includes?: readonly string[], depth?: number): string[] {
            return matchFiles(path, extensions, excludes, includes, useCaseSensitiveFileNames, process.cwd(), depth, getAccessibleFileSystemEntries, realpath);
        }

        function fileSystemEntryExists(path: string, entryKind: FileSystemEntryKind): boolean {
            const stat = statSync(path);
            if (!stat) {
                return false;
            }
            switch (entryKind) {
                case FileSystemEntryKind.File:
                    return stat.isFile();
                case FileSystemEntryKind.Directory:
                    return stat.isDirectory();
                default:
                    return false;
            }
        }

        function fileExists(path: string): boolean {
            return fileSystemEntryExists(path, FileSystemEntryKind.File);
        }

        function directoryExists(path: string): boolean {
            return fileSystemEntryExists(path, FileSystemEntryKind.Directory);
        }

        function getDirectories(path: string): string[] {
            return getAccessibleFileSystemEntries(path).directories.slice();
        }

        function fsRealPathHandlingLongPath(path: string): string {
            return path.length < 260 ? _fs.realpathSync.native(path) : _fs.realpathSync(path);
        }

        function realpath(path: string): string {
            try {
                return fsRealpath(path);
            }
            catch {
                return path;
            }
        }

        function getModifiedTime(path: string) {
            return statSync(path)?.mtime;
        }

        function setModifiedTime(path: string, time: Date) {
            try {
                _fs.utimesSync(path, time, time);
            }
            catch {
                return;
            }
        }

        function deleteFile(path: string) {
            try {
                return _fs.unlinkSync(path);
            }
            catch {
                return;
            }
        }

        function createSHA256Hash(data: string): string {
            const hash = _crypto.createHash("sha256");
            hash.update(data);
            return hash.digest("hex");
        }
    }
```
