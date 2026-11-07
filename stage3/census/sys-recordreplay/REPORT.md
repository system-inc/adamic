# sys.ts recordreplay measurement

Task b5w3ycg. Measurement only. Base origin/main e8f3723bd; pinned TypeScript 6.0.3 source 050880ce59e30b356b686bd3144efe24f875ebc8.

The complete lane inventory is pending. No zero count or false-only conclusion is claimed at this checkpoint. STATUS.md and COUNT.json, when present, describe completed pieces only.

Adaptation 93-builtin-imports pins the whole getNodeSystem function with before and after SHA-256 values in proofs/targets.json. Its adapt.cjs validates the whole owner before modifying builtin imports. The recordreplay property initializer is inside that owner. Adaptation 79 pins its nested exit method; other adapters make narrower edits inside nested functions. No recordreplay adaptation is proposed.

The scratch measurement replaces exactly `!!(process as any).recordreplay` with `!!(globalThis as any).__sysRecordReplay((process as any).recordreplay, process)`. The observer returns the exact input value, preserves the preceding logical-or short circuit, and reads the property once. Each evaluation records its Boolean result and actual primitive value, process identity and own-property descriptor. VM contexts receive the same observer. Observer-ready events distinguish processes with no evaluations from missing observer activation. Independent seeds include truthy values, falsy values, one accessor read, short circuiting and three VM entry paths.

The lane is executed through its unchanged shard.py and shard-plan.json in pieces of twenty units, retaining ordinary identity, API, baseline and count guards. Each completed piece preserves raw per-process observations, child exits, shard reports and logs, and is committed and plainly pushed. The concurrent shard deadline is 90 seconds, as in the record-table measurement; the upstream 40 second test timeout, selected tests and expectations are unchanged.

The enclosing function below is quoted whole from the completed full apply before instrumentation.

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
