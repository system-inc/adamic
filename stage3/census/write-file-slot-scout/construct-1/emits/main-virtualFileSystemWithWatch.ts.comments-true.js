import { harnessSessionCurrentDirectory, harnessTypingInstallerCacheLocation, } from "../../../harness/harnessLanguageService.js";
import { replaceAll } from "../../../harness/tsserverLogger.js";
import { createWatchUtils, ensureWatchablePath, } from "../../../harness/watchUtils.js";
import { append, arrayFrom, clear, clone, combinePaths, compareStringsCaseSensitive, createGetCanonicalFileName, createSystemWatchFunctions, Debug, directorySeparator, emptyArray, FileSystemEntryKind, FileWatcherEventKind, filterMutate, generateDjb2Hash, getBaseFileName, getDirectoryPath, getNormalizedAbsolutePath, getRelativePathToDirectoryOrUrl, hasProperty, insertSorted, isArray, isString, libMap, mapDefined, matchFiles, noop, patchWriteFileEnsuringDirectory, sys, targetToLibMap, toPath, } from "../../_namespaces/ts.js";
import { timeIncrements } from "../../_namespaces/vfs.js";
import { jsonToReadableText } from "../helpers.js";
import { getPathForTypeScriptTestLocation, tscTypeScriptTestLocation, } from "./contents.js";
import { createTypesRegistryFileContent, getTypesRegistryFileLocation, } from "./typingsInstaller.js";
export const libFile = {
    path: getPathForTypeScriptTestLocation("lib.d.ts"),
    content: `interface Boolean {}
interface Function {}
interface CallableFunction {}
interface NewableFunction {}
interface IArguments {}
interface Number { toExponential: any; }
interface Object {}
interface RegExp {}
interface String { charAt: any; }
interface Array<T> { length: number; [n: number]: T; }
interface ReadonlyArray<T> {}
declare const console: { log(msg: any): void; };`,
};
export var TestServerHostOsFlavor;
(function (TestServerHostOsFlavor) {
    TestServerHostOsFlavor[TestServerHostOsFlavor["Windows"] = 0] = "Windows";
    TestServerHostOsFlavor[TestServerHostOsFlavor["MacOs"] = 1] = "MacOs";
    TestServerHostOsFlavor[TestServerHostOsFlavor["Linux"] = 2] = "Linux";
})(TestServerHostOsFlavor || (TestServerHostOsFlavor = {}));
export function isFile(fileOrFolderOrSymLink) {
    return isString(fileOrFolderOrSymLink.content) || fileOrFolderOrSymLink.libFile;
}
export function isSymLink(fileOrFolderOrSymLink) {
    return isString(fileOrFolderOrSymLink.symLink);
}
function isFsFolder(s) {
    return !!s && isArray(s.entries);
}
function isFsFileOrFsLibFile(s) {
    return isFsFile(s) || isFsLibFile(s);
}
function isFsLibFile(s) {
    return !!s && s.libFile;
}
export function isFsFile(s) {
    return !!s && isString(s.content);
}
function getFsFileOrLibFileContent(entry) {
    return entry.content ?? (entry.content = libFile.content);
}
function isFsSymLink(s) {
    return !!s && isString(s.symLink);
}
const exitMessage = "System Exit";
class Callbacks {
    host;
    callbackType;
    swallowExitException;
    map = new Map();
    nextId = 1;
    invoke = invokeKey => this.invokeWorker(invokeKey);
    hasChanges = false;
    serializedKeys = new Map();
    constructor(host, callbackType, swallowExitException) {
        this.host = host;
        this.callbackType = callbackType;
        this.swallowExitException = swallowExitException;
    }
    getNextId() {
        return this.nextId;
    }
    register(cb, args, ms) {
        const timeoutId = this.nextId;
        this.nextId++;
        this.map.set(timeoutId, { cb, args, ms, time: this.host.getTime() });
        this.hasChanges = true;
        return timeoutId;
    }
    unregister(id) {
        if (typeof id === "number") {
            this.hasChanges = this.map.delete(id) || this.hasChanges;
        }
    }
    log(logChanges) {
        const details = [];
        this.map.forEach(({ args }, timeoutId) => {
            details.push(`${timeoutId}: ${args[0]}${!logChanges || this.serializedKeys.has(timeoutId) ? "" : " *new*"}`);
            if (logChanges)
                this.serializedKeys.set(timeoutId, args[0]);
        });
        const deleted = [];
        if (logChanges && this.serializedKeys.size !== this.map.size) {
            this.serializedKeys.forEach((value, key) => {
                if (this.map.has(key))
                    return;
                deleted.push(`${key}: ${value} *deleted*`);
                this.serializedKeys.delete(key);
            });
        }
        return `${this.callbackType} callback:: count: ${this.map.size}` +
            (deleted.length ? "\r\n" + deleted.join("\r\n") : "") +
            (details.length ? "\r\n" + details.join("\r\n") : "");
    }
    invokeCallback(timeoutId) {
        const data = this.map.get(timeoutId);
        if (!data)
            return;
        const { cb, args, ms, time } = data;
        this.map.delete(timeoutId);
        this.serializedKeys.delete(timeoutId);
        if (ms !== undefined) {
            const newTime = ms + time;
            if (this.host.getTime() < newTime) {
                this.host.setTime(newTime);
            }
        }
        cb(...args);
    }
    invokeWorker(invokeKey) {
        try {
            if (invokeKey)
                return this.invokeCallback(invokeKey);
            // Note: invoking a callback may result in new callbacks been queued,
            // so do not clear the entire callback list regardless. Only remove the
            // ones we have invoked.
            const keys = arrayFrom(this.map.keys());
            for (const key of keys) {
                this.invokeCallback(key);
            }
        }
        catch (e) {
            if (this.swallowExitException && e.message === exitMessage) {
                return;
            }
            throw e;
        }
    }
    switchToBaseliningInvoke(logger, serializeOutputOrder) {
        this.invoke = invokeKey => {
            logger.log(`Before running ${this.log()}`);
            this.host.serializeState(logger.logs, serializeOutputOrder);
            if (invokeKey !== undefined)
                logger.log(`Invoking ${this.callbackType} callback:: timeoutId:: ${invokeKey}:: ${this.map.get(invokeKey).args[0]}`);
            this.invokeWorker(invokeKey);
            logger.log(`After running ${this.callbackType} callback:: count: ${this.map.size}`);
            this.host.serializeState(logger.logs, serializeOutputOrder);
        };
    }
    serialize(baseline) {
        if (this.hasChanges) {
            baseline.push(this.log(/*logChanges*/ true), "");
            this.hasChanges = false;
        }
    }
}
export var Tsc_WatchFile;
(function (Tsc_WatchFile) {
    Tsc_WatchFile["DynamicPolling"] = "DynamicPriorityPolling";
})(Tsc_WatchFile || (Tsc_WatchFile = {}));
export var Tsc_WatchDirectory;
(function (Tsc_WatchDirectory) {
    Tsc_WatchDirectory["WatchFile"] = "RecursiveDirectoryUsingFsWatchFile";
    Tsc_WatchDirectory["NonRecursiveWatchDirectory"] = "RecursiveDirectoryUsingNonRecursiveWatchDirectory";
    Tsc_WatchDirectory["DynamicPolling"] = "RecursiveDirectoryUsingDynamicPriorityPolling";
})(Tsc_WatchDirectory || (Tsc_WatchDirectory = {}));
export var SerializeOutputOrder;
(function (SerializeOutputOrder) {
    SerializeOutputOrder[SerializeOutputOrder["None"] = 0] = "None";
    SerializeOutputOrder[SerializeOutputOrder["BeforeDiff"] = 1] = "BeforeDiff";
    SerializeOutputOrder[SerializeOutputOrder["AfterDiff"] = 2] = "AfterDiff";
})(SerializeOutputOrder || (SerializeOutputOrder = {}));
export function cloneFsMap(fs) {
    const snap = new Map();
    const clonedMap = new Map();
    const getCloned = (entry) => {
        const existing = clonedMap.get(entry);
        if (existing)
            return existing;
        const cloned = clone(entry);
        clonedMap.set(entry, cloned);
        if (isFsFolder(cloned)) {
            cloned.entries = cloned.entries.map(clone);
        }
        return cloned;
    };
    fs.forEach((value, key) => {
        snap.set(key, getCloned(value));
    });
    return snap;
}
export class TestServerHost {
    args = [];
    output = [];
    fs = new Map();
    time = timeIncrements;
    getCanonicalFileName;
    toPath;
    timeoutCallbacks = new Callbacks(this, "Timeout", /*swallowExitException*/ true);
    immediateCallbacks = new Callbacks(this, "Immedidate");
    pendingInstalls = new Callbacks(this, "PendingInstalls");
    screenClears = [];
    watchUtils;
    runWithFallbackPolling;
    useCaseSensitiveFileNames;
    newLine;
    windowsStyleRoot;
    environmentVariables;
    executingFilePath;
    currentDirectory;
    require;
    importPlugin;
    storeSignatureInfo = true;
    watchFile;
    inodeWatching;
    inodes;
    watchDirectory;
    service;
    osFlavor;
    preferNonRecursiveWatch;
    globalTypingsCacheLocation;
    typesRegistry;
    constructor(fileOrFolderorSymLinkList, { useCaseSensitiveFileNames, executingFilePath, currentDirectory, newLine, windowsStyleRoot, environmentVariables, runWithFallbackPolling, osFlavor, typingsInstallerGlobalCacheLocation, typingsInstallerTypesRegistry, } = {}) {
        this.useCaseSensitiveFileNames = !!useCaseSensitiveFileNames;
        this.newLine = newLine || "\n";
        this.osFlavor = osFlavor || TestServerHostOsFlavor.Windows;
        this.windowsStyleRoot = windowsStyleRoot;
        this.environmentVariables = environmentVariables;
        this.currentDirectory = this.getHostSpecificPath(currentDirectory ?? harnessSessionCurrentDirectory);
        this.getCanonicalFileName = createGetCanonicalFileName(!!useCaseSensitiveFileNames);
        this.watchUtils = createWatchUtils("PolledWatches", "FsWatches", s => this.getCanonicalFileName(s), this);
        this.toPath = s => toPath(s, this.currentDirectory, this.getCanonicalFileName);
        this.executingFilePath = this.getHostSpecificPath(executingFilePath ?? tscTypeScriptTestLocation);
        this.runWithFallbackPolling = !!runWithFallbackPolling;
        const tscWatchFile = this.environmentVariables && this.environmentVariables.get("TSC_WATCHFILE");
        const tscWatchDirectory = this.environmentVariables && this.environmentVariables.get("TSC_WATCHDIRECTORY");
        this.preferNonRecursiveWatch = this.osFlavor === TestServerHostOsFlavor.Linux;
        this.inodeWatching = this.osFlavor !== TestServerHostOsFlavor.Windows;
        if (this.inodeWatching)
            this.inodes = new Map();
        const { watchFile, watchDirectory } = createSystemWatchFunctions({
            // We dont have polling watch file
            // it is essentially fsWatch but lets get that separate from fsWatch and
            // into watchedFiles for easier testing
            pollingWatchFileWorker: this.watchFileWorker.bind(this),
            getModifiedTime: this.getModifiedTime.bind(this),
            setTimeout: this.setTimeout.bind(this),
            clearTimeout: this.clearTimeout.bind(this),
            fsWatchWorker: this.fsWatchWorker.bind(this),
            fileSystemEntryExists: this.fileSystemEntryExists.bind(this),
            useCaseSensitiveFileNames: this.useCaseSensitiveFileNames,
            getCurrentDirectory: this.getCurrentDirectory.bind(this),
            fsSupportsRecursiveFsWatch: !this.preferNonRecursiveWatch,
            getAccessibleSortedChildDirectories: path => this.getDirectories(path),
            realpath: this.realpath.bind(this),
            tscWatchFile,
            tscWatchDirectory,
            inodeWatching: this.inodeWatching,
            fsWatchWithTimestamp: this.osFlavor === TestServerHostOsFlavor.MacOs,
            sysLog: s => this.write(s + this.newLine),
        });
        this.watchFile = watchFile;
        this.watchDirectory = watchDirectory;
        this.reloadFS(fileOrFolderorSymLinkList);
        // Ensure current directory exists
        this.ensureFileOrFolder({ path: this.currentDirectory });
        // Ensure libs
        this.ensureFileOrFolder({
            path: combinePaths(getDirectoryPath(this.executingFilePath), "lib.d.ts"),
            content: libFile.content,
        });
        targetToLibMap.forEach(libFile => this.ensureLib(libFile));
        libMap.forEach(libFile => this.ensureLib(libFile));
        // Ensure global cache location has required data
        this.globalTypingsCacheLocation = this.getHostSpecificPath(typingsInstallerGlobalCacheLocation ??
            harnessTypingInstallerCacheLocation);
        this.typesRegistry = typingsInstallerTypesRegistry;
        if (this.directoryExists(this.globalTypingsCacheLocation)) {
            // Package.json
            const npmConfigPath = combinePaths(this.globalTypingsCacheLocation, "package.json");
            if (!this.fileExists(npmConfigPath)) {
                this.writeFile(npmConfigPath, '{ "private": true }');
            }
            // Typings registry
            this.ensureTypingRegistryFile();
        }
    }
    ensureLib(path) {
        this.ensureFileOrFolder({
            path: combinePaths(getDirectoryPath(this.executingFilePath), path),
            libFile: true,
        });
    }
    ensureTypingRegistryFile() {
        this.ensureFileOrFolder({
            path: getTypesRegistryFileLocation(this.globalTypingsCacheLocation),
            content: jsonToReadableText(createTypesRegistryFileContent(this.typesRegistry ?
                isString(this.typesRegistry) ?
                    [this.typesRegistry] :
                    this.typesRegistry :
                emptyArray)),
        });
    }
    static createWatchedSystem(fileOrFolderList, params) {
        const useDefaultCurrentDirectory = !params?.currentDirectory;
        if (useDefaultCurrentDirectory)
            (params ??= {}).currentDirectory = "/home/src/workspaces/project";
        else
            ensureWatchablePath(params.currentDirectory, `currentDirectory: ${params.currentDirectory}`);
        const host = new TestServerHost(fileOrFolderList, params);
        if (useDefaultCurrentDirectory && fileOrFolderList !== emptyArray) {
            const folder = host.getRealFolder(host.toPath(host.currentDirectory));
            Debug.assert(folder.entries.length, `currentDirectory: Not specified, using default as "/home/src/workspaces/project". The files specified do not belong to it.`);
        }
        return host;
    }
    static createServerHost(fileOrFolderList, params) {
        if (params?.currentDirectory)
            params.currentDirectory = undefined;
        if (params)
            params.currentDirectory = params?.overrideCurrentDirectory;
        const host = new TestServerHost(fileOrFolderList, params);
        // Just like sys, patch the host to use writeFile
        patchWriteFileEnsuringDirectory(host);
        return host;
    }
    static getCreateWatchedSystem(forTsserver) {
        return (!forTsserver ? TestServerHost.createWatchedSystem : TestServerHost.createServerHost);
    }
    nextInode = 0;
    setInode(path) {
        if (this.inodes)
            this.inodes.set(path, this.nextInode++);
    }
    // Output is pretty
    writeOutputIsTTY() {
        return true;
    }
    getNewLine() {
        return this.newLine;
    }
    toNormalizedAbsolutePath(s) {
        return getNormalizedAbsolutePath(s, this.currentDirectory);
    }
    toFullPath(s) {
        return this.toPath(this.toNormalizedAbsolutePath(s));
    }
    getHostSpecificPath(s) {
        if (this.windowsStyleRoot) {
            let result = s;
            if (s.startsWith(directorySeparator)) {
                result = this.windowsStyleRoot + s.substring(1);
            }
            if (!result.startsWith(directorySeparator))
                result = replaceAll(result, directorySeparator, "\\");
            return result;
        }
        return s;
    }
    now() {
        this.time += timeIncrements;
        return new Date(this.time);
    }
    getTime() {
        return this.time;
    }
    setTime(time) {
        this.time = time;
    }
    switchToBaseliningInvoke(logger, serializeOutputOrder) {
        const originalSetTime = this.setTime;
        this.setTime = time => {
            logger.log(`Host is moving to new time`);
            return originalSetTime.call(this, time);
        };
        this.timeoutCallbacks.switchToBaseliningInvoke(logger, serializeOutputOrder);
        this.immediateCallbacks.switchToBaseliningInvoke(logger, serializeOutputOrder);
        this.pendingInstalls.switchToBaseliningInvoke(logger, serializeOutputOrder);
    }
    ensureInitialFileOrFolder(f) {
        if (!f.watchableExempt)
            ensureWatchablePath(getDirectoryPath(f.path), `Directory path of FileOrFolderOrSymLink: ${f.path}`);
        this.ensureFileOrFolder(f);
    }
    reloadFS(fileOrFolderOrSymLinkList) {
        Debug.assert(this.fs.size === 0);
        if (isArray(fileOrFolderOrSymLinkList)) {
            fileOrFolderOrSymLinkList.forEach(f => this.ensureInitialFileOrFolder(!this.windowsStyleRoot ?
                f :
                { ...f, path: this.getHostSpecificPath(f.path) }));
        }
        else {
            for (const key in fileOrFolderOrSymLinkList) {
                if (hasProperty(fileOrFolderOrSymLinkList, key)) {
                    const path = this.getHostSpecificPath(key);
                    const value = fileOrFolderOrSymLinkList[key];
                    if (isString(value)) {
                        this.ensureInitialFileOrFolder({ path, content: value });
                    }
                    else {
                        this.ensureInitialFileOrFolder({ path, ...value });
                    }
                }
            }
        }
        // Ensure root folder exists
        this.ensureFileOrFolder({ path: !this.windowsStyleRoot ? "/" : this.getHostSpecificPath("/") });
    }
    modifyFile(filePath, content, options) {
        const path = this.toFullPath(filePath);
        const currentEntry = this.fs.get(path);
        if (!currentEntry || !isFsFileOrFsLibFile(currentEntry)) {
            throw new Error(`file not present: ${filePath}`);
        }
        if (isFsLibFile(currentEntry) || options?.invokeFileDeleteCreateAsPartInsteadOfChange) {
            this.removeFileOrFolder(currentEntry, /*isRenaming*/ false, options);
            this.ensureFileOrFolder({ path: filePath, content }, /*ignoreWatchInvokedWithTriggerAsFileCreate*/ undefined, /*ignoreParentWatch*/ undefined, options);
        }
        else {
            currentEntry.content = content;
            currentEntry.modifiedTime = this.now();
            if (options?.ignoreWatches)
                return;
            if (options && options.invokeDirectoryWatcherInsteadOfFileChanged) {
                const directoryFullPath = getDirectoryPath(currentEntry.fullPath);
                this.fs.get(getDirectoryPath(currentEntry.path)).modifiedTime = this.now();
                this.invokeFileWatcher(directoryFullPath, FileWatcherEventKind.Changed, /*modifiedTime*/ undefined);
                this.invokeFsWatchesCallbacks(directoryFullPath, "rename", currentEntry.fullPath, options.useTildeAsSuffixInRenameEventFileName);
                this.invokeRecursiveFsWatches(directoryFullPath, "rename", currentEntry.fullPath, options.useTildeAsSuffixInRenameEventFileName);
            }
            else {
                this.invokeFileAndFsWatches(currentEntry.fullPath, FileWatcherEventKind.Changed, currentEntry.fullPath, currentEntry.modifiedTime, options?.useTildeAsSuffixInRenameEventFileName);
            }
        }
    }
    renameFile(fileName, newFileName) {
        const fullPath = getNormalizedAbsolutePath(fileName, this.currentDirectory);
        const path = this.toPath(fullPath);
        const file = this.fs.get(path);
        Debug.assert(!!file);
        // Only remove the file
        this.removeFileOrFolder(file, /*isRenaming*/ true);
        // Add updated folder with new folder name
        const newFullPath = getNormalizedAbsolutePath(newFileName, this.currentDirectory);
        const newFile = this.toFsFileOrLibFile({ path: newFullPath, content: file.content });
        const newPath = newFile.path;
        const basePath = getDirectoryPath(newPath);
        Debug.assert(basePath !== path);
        this.ensureFileOrFolder({ path: getDirectoryPath(newFullPath) });
        const baseFolder = this.fs.get(basePath);
        this.addFileOrFolderInFolder(baseFolder, newFile);
    }
    renameFolder(folderName, newFolderName, skipFolderEntryWatches) {
        const fullPath = getNormalizedAbsolutePath(folderName, this.currentDirectory);
        const path = this.toPath(fullPath);
        const folder = this.fs.get(path);
        Debug.assert(!!folder);
        const newFullPath = getNormalizedAbsolutePath(newFolderName, this.currentDirectory);
        const newFolder = this.toFsFolder(newFullPath);
        // Invoke watches for files in the folder as deleted (from old path)
        this.renameFolderEntries(folder, newFolder, skipFolderEntryWatches);
        // Only remove the folder
        this.removeFileOrFolder(folder, /*isRenaming*/ true);
        // Add updated folder with new folder name
        const newPath = newFolder.path;
        const basePath = getDirectoryPath(newPath);
        this.ensureFileOrFolder({ path: getDirectoryPath(newFullPath) });
        const baseFolder = this.fs.get(basePath);
        this.addFileOrFolderInFolder(baseFolder, newFolder);
    }
    renameFolderEntries(oldFolder, newFolder, skipWatches) {
        for (const entry of oldFolder.entries) {
            this.fs.delete(entry.path);
            if (!skipWatches)
                this.invokeFileAndFsWatches(entry.fullPath, FileWatcherEventKind.Deleted, entry.fullPath);
            entry.fullPath = combinePaths(newFolder.fullPath, getBaseFileName(entry.fullPath));
            entry.path = this.toPath(entry.fullPath);
            if (newFolder !== oldFolder) {
                newFolder.entries.push(entry);
            }
            this.fs.set(entry.path, entry);
            this.setInode(entry.path);
            if (!skipWatches)
                this.invokeFileAndFsWatches(entry.fullPath, FileWatcherEventKind.Created, entry.fullPath);
            if (isFsFolder(entry)) {
                this.renameFolderEntries(entry, entry, skipWatches);
            }
        }
    }
    ensureFileOrFolder(fileOrDirectoryOrSymLink, ignoreWatchInvokedWithTriggerAsFileCreate, ignoreParentWatch, options) {
        if (isFile(fileOrDirectoryOrSymLink)) {
            const file = this.toFsFileOrLibFile(fileOrDirectoryOrSymLink);
            // file may already exist when updating existing type declaration file
            if (!this.fs.get(file.path)) {
                const baseFolder = this.ensureFolder(getDirectoryPath(file.fullPath), ignoreParentWatch, options);
                this.addFileOrFolderInFolder(baseFolder, file, ignoreWatchInvokedWithTriggerAsFileCreate, options);
            }
        }
        else if (isSymLink(fileOrDirectoryOrSymLink)) {
            const symLink = this.toFsSymLink(fileOrDirectoryOrSymLink);
            Debug.assert(!this.fs.get(symLink.path));
            const baseFolder = this.ensureFolder(getDirectoryPath(symLink.fullPath), ignoreParentWatch, options);
            this.addFileOrFolderInFolder(baseFolder, symLink, ignoreWatchInvokedWithTriggerAsFileCreate, options);
        }
        else {
            const fullPath = getNormalizedAbsolutePath(fileOrDirectoryOrSymLink.path, this.currentDirectory);
            this.ensureFolder(getDirectoryPath(fullPath), ignoreParentWatch, options);
            this.ensureFolder(fullPath, ignoreWatchInvokedWithTriggerAsFileCreate, options);
        }
    }
    ensureFolder(fullPath, ignoreWatch, options) {
        const path = this.toPath(fullPath);
        let folder = this.fs.get(path);
        if (!folder) {
            folder = this.toFsFolder(fullPath);
            const baseFullPath = getDirectoryPath(fullPath);
            if (fullPath !== baseFullPath) {
                // Add folder in the base folder
                const baseFolder = this.ensureFolder(baseFullPath, ignoreWatch, options);
                this.addFileOrFolderInFolder(baseFolder, folder, ignoreWatch, options);
            }
            else {
                // root folder
                Debug.assert(this.fs.size === 0 || !!this.windowsStyleRoot);
                this.fs.set(path, folder);
                this.setInode(path);
            }
        }
        Debug.assert(isFsFolder(folder));
        return folder;
    }
    addFileOrFolderInFolder(folder, fileOrDirectory, ignoreWatch, options) {
        if (!this.fs.has(fileOrDirectory.path)) {
            insertSorted(folder.entries, fileOrDirectory, (a, b) => compareStringsCaseSensitive(getBaseFileName(a.path), getBaseFileName(b.path)));
        }
        folder.modifiedTime = this.now();
        this.fs.set(fileOrDirectory.path, fileOrDirectory);
        this.setInode(fileOrDirectory.path);
        if (ignoreWatch || options?.ignoreWatches) {
            return;
        }
        const inodeWatching = this.inodeWatching;
        if (options?.skipInodeCheckOnCreate)
            this.inodeWatching = false;
        this.invokeFileAndFsWatches(fileOrDirectory.fullPath, FileWatcherEventKind.Created, fileOrDirectory.fullPath, fileOrDirectory.modifiedTime, options?.useTildeAsSuffixInRenameEventFileName);
        if (this.osFlavor !== TestServerHostOsFlavor.MacOs) {
            this.invokeFileAndFsWatches(folder.fullPath, FileWatcherEventKind.Changed, fileOrDirectory.fullPath, folder.modifiedTime, options?.useTildeAsSuffixInRenameEventFileName);
            const folderOfFolder = getDirectoryPath(folder.fullPath);
            if (folderOfFolder !== folder.fullPath) {
                this.invokeFsWatches(folderOfFolder, "change", folder.fullPath, options?.useTildeAsSuffixInRenameEventFileName);
            }
        }
        this.inodeWatching = inodeWatching;
    }
    removeFileOrFolder(fileOrDirectory, isRenaming, options) {
        const basePath = getDirectoryPath(fileOrDirectory.path);
        const baseFolder = this.fs.get(basePath);
        if (basePath !== fileOrDirectory.path) {
            Debug.assert(!!baseFolder);
            baseFolder.modifiedTime = this.now();
            filterMutate(baseFolder.entries, entry => entry !== fileOrDirectory);
        }
        this.fs.delete(fileOrDirectory.path);
        if (isFsFolder(fileOrDirectory)) {
            Debug.assert(fileOrDirectory.entries.length === 0 || isRenaming);
        }
        if (!options?.ignoreDelete && !options?.ignoreWatches)
            this.invokeFileAndFsWatches(fileOrDirectory.fullPath, FileWatcherEventKind.Deleted, fileOrDirectory.fullPath, /*modifiedTime*/ undefined, options?.useTildeAsSuffixInRenameEventFileName);
        this.inodes?.delete(fileOrDirectory.path);
    }
    rimrafSync(fileOrFolderPath) {
        const fileOrFolder = this.getRealFileOrFolder(fileOrFolderPath);
        if (isFsFileOrFsLibFile(fileOrFolder))
            this.removeFileOrFolder(fileOrFolder);
        else if (isFsFolder(fileOrFolder))
            this.deleteFolder(fileOrFolder.fullPath, /*recursive*/ true);
    }
    deleteFile(filePath) {
        const file = this.getRealFileOrFolder(filePath);
        Debug.assert(isFsFileOrFsLibFile(file));
        this.removeFileOrFolder(file);
    }
    deleteFolder(folderPath, recursive) {
        const path = this.toFullPath(folderPath);
        const currentEntry = this.fs.get(path);
        Debug.assert(isFsFolder(currentEntry));
        if (recursive && currentEntry.entries.length) {
            const subEntries = currentEntry.entries.slice();
            subEntries.forEach(fsEntry => {
                if (isFsFolder(fsEntry)) {
                    this.deleteFolder(fsEntry.fullPath, recursive);
                }
                else {
                    this.removeFileOrFolder(fsEntry);
                }
            });
        }
        this.removeFileOrFolder(currentEntry);
    }
    watchFileWorker(fileName, cb, pollingInterval) {
        return this.watchUtils.pollingWatch(this.toNormalizedAbsolutePath(fileName), { cb, pollingInterval });
    }
    fsWatchWorker(fileOrDirectory, recursive, cb) {
        if (this.runWithFallbackPolling)
            throw new Error("Need to use fallback polling instead of file system native watching");
        let inode;
        if (this.inodeWatching) {
            const entry = this.getRealFileOrFolder(fileOrDirectory);
            // Error if the path does not exist
            if (!entry)
                throw new Error("Cannot watch missing file or folder");
            inode = this.inodes?.get(entry.path);
            if (inode === undefined)
                throw new Error("inode not found");
        }
        const result = this.watchUtils.fsWatch(this.toNormalizedAbsolutePath(fileOrDirectory), recursive, {
            cb,
            inode,
        });
        result.on = noop;
        return result;
    }
    invokeFileWatcher(fileFullPath, eventKind, modifiedTime) {
        this.watchUtils.pollingWatches.forEach(fileFullPath, ({ cb }, fullPath) => cb(fullPath, eventKind, modifiedTime));
    }
    fsWatchCallback(watches, fullPath, eventName, eventFullPath, useTildeSuffix) {
        const path = this.toPath(fullPath);
        const currentInode = this.inodes?.get(path);
        watches.forEach(path, ({ cb, inode }) => {
            if (this.inodeWatching && inode !== undefined && inode !== currentInode)
                return;
            let relativeFileName = eventFullPath ? this.getRelativePathToDirectory(fullPath, eventFullPath) : "";
            if (useTildeSuffix)
                relativeFileName = (relativeFileName ? relativeFileName : getBaseFileName(fullPath)) + "~";
            cb(eventName, relativeFileName);
        });
    }
    invokeFsWatchesCallbacks(fullPath, eventName, eventFullPath, useTildeSuffix) {
        this.fsWatchCallback(this.watchUtils.fsWatches, fullPath, eventName, eventFullPath, useTildeSuffix);
    }
    invokeFsWatchesRecursiveCallbacks(fullPath, eventName, eventFullPath, useTildeSuffix) {
        this.fsWatchCallback(this.watchUtils.fsWatchesRecursive, fullPath, eventName, eventFullPath, useTildeSuffix);
    }
    getRelativePathToDirectory(directoryFullPath, fileFullPath) {
        return getRelativePathToDirectoryOrUrl(directoryFullPath, fileFullPath, this.currentDirectory, this.getCanonicalFileName, /*isAbsolutePathAnUrl*/ false);
    }
    invokeRecursiveFsWatches(fullPath, eventName, eventFullPath, useTildeSuffix) {
        this.invokeFsWatchesRecursiveCallbacks(fullPath, eventName, eventFullPath, useTildeSuffix);
        const basePath = getDirectoryPath(fullPath);
        if (this.getCanonicalFileName(fullPath) !== this.getCanonicalFileName(basePath)) {
            this.invokeRecursiveFsWatches(basePath, eventName, eventFullPath, useTildeSuffix);
        }
    }
    invokeFsWatches(fullPath, eventName, eventFullPath, useTildeSuffix) {
        this.invokeFsWatchesCallbacks(fullPath, eventName, eventFullPath, useTildeSuffix);
        this.invokeFsWatchesCallbacks(getDirectoryPath(fullPath), eventName, eventFullPath, useTildeSuffix);
        this.invokeRecursiveFsWatches(fullPath, eventName, eventFullPath, useTildeSuffix);
    }
    invokeFileAndFsWatches(fileOrFolderFullPath, eventKind, eventFullPath, modifiedTime, useTildeSuffix) {
        this.invokeFileWatcher(fileOrFolderFullPath, eventKind, modifiedTime);
        this.invokeFsWatches(fileOrFolderFullPath, eventKind === FileWatcherEventKind.Changed ? "change" : "rename", eventFullPath, useTildeSuffix);
    }
    toFsEntry(path) {
        const fullPath = getNormalizedAbsolutePath(path, this.currentDirectory);
        return {
            path: this.toPath(fullPath),
            fullPath,
            modifiedTime: this.now(),
        };
    }
    toFsFileOrLibFile(file) {
        const fsFile = this.toFsEntry(file.path);
        fsFile.content = file.content;
        fsFile.fileSize = file.fileSize;
        fsFile.libFile = file.libFile;
        return fsFile;
    }
    toFsSymLink(symLink) {
        const fsSymLink = this.toFsEntry(symLink.path);
        fsSymLink.symLink = getNormalizedAbsolutePath(symLink.symLink, getDirectoryPath(fsSymLink.fullPath));
        return fsSymLink;
    }
    toFsFolder(path) {
        const fsFolder = this.toFsEntry(path);
        fsFolder.entries = []; // https://github.com/Microsoft/TypeScript/issues/19873
        return fsFolder;
    }
    getRealFsEntry(isFsEntry, path, fsEntry = this.fs.get(path)) {
        if (isFsEntry(fsEntry)) {
            return fsEntry;
        }
        if (isFsSymLink(fsEntry)) {
            return this.getRealFsEntry(isFsEntry, this.toPath(fsEntry.symLink));
        }
        if (fsEntry) {
            // This fs entry is something else
            return undefined;
        }
        const realpath = this.toPath(this.realpath(path));
        if (path !== realpath) {
            return this.getRealFsEntry(isFsEntry, realpath);
        }
        return undefined;
    }
    isFsFile(fsEntry) {
        return !!this.getRealFile(fsEntry.path, fsEntry);
    }
    getRealFile(path, fsEntry) {
        return this.getRealFsEntry(isFsFileOrFsLibFile, path, fsEntry);
    }
    isFsFolder(fsEntry) {
        return !!this.getRealFolder(fsEntry.path, fsEntry);
    }
    getRealFolder(path, fsEntry = this.fs.get(path)) {
        return this.getRealFsEntry(isFsFolder, path, fsEntry);
    }
    getRealFileOrFolder(s) {
        return this.getRealFsEntry((entry) => !!entry && !isFsSymLink(entry), this.toFullPath(s));
    }
    fileSystemEntryExists(s, entryKind) {
        return entryKind === FileSystemEntryKind.File ? this.fileExists(s) : this.directoryExists(s);
    }
    fileExists(s) {
        const path = this.toFullPath(s);
        return !!this.getRealFile(path);
    }
    getModifiedTime(s) {
        return this.getRealFileOrFolder(s)?.modifiedTime;
    }
    setModifiedTime(s, date) {
        const fsEntry = this.getRealFileOrFolder(s);
        if (fsEntry) {
            fsEntry.modifiedTime = date;
            this.invokeFileAndFsWatches(fsEntry.fullPath, FileWatcherEventKind.Changed, fsEntry.fullPath, fsEntry.modifiedTime);
        }
    }
    readFile(s) {
        const fsEntry = this.getRealFile(this.toFullPath(s));
        return fsEntry ? getFsFileOrLibFileContent(fsEntry) : undefined;
    }
    getFileSize(s) {
        const path = this.toFullPath(s);
        const entry = this.fs.get(path);
        if (isFsFileOrFsLibFile(entry)) {
            return entry.fileSize ? entry.fileSize : getFsFileOrLibFileContent(entry).length;
        }
        return undefined; // TODO: GH#18217
    }
    directoryExists(s) {
        const path = this.toFullPath(s);
        return !!this.getRealFolder(path);
    }
    getDirectories(s) {
        const path = this.toFullPath(s);
        const folder = this.getRealFolder(path);
        if (folder) {
            return mapDefined(folder.entries, entry => this.isFsFolder(entry) ? getBaseFileName(entry.fullPath) : undefined);
        }
        Debug.fail(folder ? "getDirectories called on file" : "getDirectories called on missing folder");
        return [];
    }
    readDirectory(path, extensions, exclude, include, depth) {
        return matchFiles(path, extensions, exclude, include, this.useCaseSensitiveFileNames, this.getCurrentDirectory(), depth, dir => {
            const directories = [];
            const files = [];
            const folder = this.getRealFolder(this.toPath(dir));
            if (folder) {
                folder.entries.forEach(entry => {
                    if (this.isFsFolder(entry)) {
                        directories.push(getBaseFileName(entry.fullPath));
                    }
                    else if (this.isFsFile(entry)) {
                        files.push(getBaseFileName(entry.fullPath));
                    }
                    else {
                        Debug.fail("Unknown entry");
                    }
                });
            }
            return { directories, files };
        }, path => this.realpath(path));
    }
    createHash(s) {
        return `${generateDjb2Hash(s)}-${s}`;
    }
    createSHA256Hash(s) {
        return sys.createSHA256Hash(s);
    }
    // TOOD: record and invoke callbacks to simulate timer events
    setTimeout(callback, ms, label) {
        return this.timeoutCallbacks.register(callback, [label], ms);
    }
    getNextTimeoutId() {
        return this.timeoutCallbacks.getNextId();
    }
    clearTimeout(timeoutId) {
        this.timeoutCallbacks.unregister(timeoutId);
    }
    clearScreen() {
        this.screenClears.push(this.output.length);
    }
    runQueuedTimeoutCallbacks(timeoutId) {
        this.timeoutCallbacks.invoke(timeoutId);
    }
    runQueuedImmediateCallbacks() {
        this.immediateCallbacks.invoke();
    }
    setImmediate(callback, ...args) {
        return this.immediateCallbacks.register(callback, args);
    }
    clearImmediate(timeoutId) {
        this.immediateCallbacks.unregister(timeoutId);
    }
    scheduleInstall(cb, ...args) {
        this.pendingInstalls.register(cb, args);
    }
    runPendingInstalls() {
        this.pendingInstalls.invoke();
    }
    createDirectory(directoryName) {
        const folder = this.toFsFolder(directoryName);
        // base folder has to be present
        const base = getDirectoryPath(folder.path);
        const baseFolder = this.fs.get(base);
        Debug.assert(isFsFolder(baseFolder));
        Debug.assert(!this.fs.get(folder.path));
        this.addFileOrFolderInFolder(baseFolder, folder);
    }
    writeFile(path, content) {
        const file = this.toFsFileOrLibFile({ path, content });
        // base folder has to be present
        const base = getDirectoryPath(file.path);
        const folder = this.getRealFolder(base);
        if (!folder) {
            throw new Error(`Directory not found: ${base}`);
        }
        if (folder.path === base) {
            if (!this.fs.has(file.path)) {
                this.addFileOrFolderInFolder(folder, file);
            }
            else {
                this.modifyFile(path, content);
            }
        }
        else {
            this.writeFile(this.realpath(path), content);
        }
    }
    prependFile(path, content, options) {
        this.modifyFile(path, content + this.readFile(path), options);
    }
    appendFile(path, content, options) {
        this.modifyFile(path, this.readFile(path) + content, options);
    }
    replaceFileText(file, searchValue, replaceValue) {
        const content = Debug.checkDefined(this.readFile(file));
        this.writeFile(file, content.replace(searchValue, replaceValue));
    }
    write(message) {
        if (Debug.isDebugging)
            console.log(message);
        this.output.push(message);
    }
    getOutput() {
        return this.output;
    }
    clearOutput() {
        clear(this.output);
        this.screenClears.length = 0;
    }
    serializeOutput(baseline) {
        const output = this.getOutput();
        if (!this.output.length && !this.screenClears.length)
            return;
        let start = 0;
        baseline.push("Output::");
        for (const screenClear of this.screenClears) {
            baselineOutputs(baseline, output, start, screenClear);
            start = screenClear;
            baseline.push(">> Screen clear");
        }
        baselineOutputs(baseline, output, start);
        baseline.push("");
        this.clearOutput();
    }
    getSnap() {
        return cloneFsMap(this.fs);
    }
    snap() {
        this.serializedDiff = cloneFsMap(this.fs);
    }
    serializeState(baseline, serializeOutput) {
        if (serializeOutput === SerializeOutputOrder.BeforeDiff)
            this.serializeOutput(baseline);
        this.diff(baseline);
        if (serializeOutput === SerializeOutputOrder.AfterDiff)
            this.serializeOutput(baseline);
        this.serializeWatches(baseline);
        this.timeoutCallbacks.serialize(baseline);
        this.immediateCallbacks.serialize(baseline);
        this.pendingInstalls.serialize(baseline);
        this.service?.baseline();
    }
    writtenFiles;
    serializedDiff = new Map();
    diff(baseline) {
        this.fs.forEach((newFsEntry, path) => {
            diffFsEntry(baseline, this.serializedDiff.get(path), newFsEntry, this.inodes?.get(path), this.writtenFiles);
        });
        this.serializedDiff.forEach((oldFsEntry, path) => {
            const newFsEntry = this.fs.get(path);
            if (!newFsEntry) {
                diffFsEntry(baseline, oldFsEntry, newFsEntry, this.inodes?.get(path), this.writtenFiles);
            }
        });
        baseline.push("");
        this.snap();
        this.writtenFiles?.clear();
    }
    serializeWatches(baseline) {
        return this.watchUtils.serializeWatches(baseline);
    }
    realpath(s) {
        const fullPath = this.toNormalizedAbsolutePath(s);
        const path = this.toPath(fullPath);
        if (getDirectoryPath(path) === path) {
            // Root
            return s;
        }
        const dirFullPath = this.realpath(getDirectoryPath(fullPath));
        const realFullPath = combinePaths(dirFullPath, getBaseFileName(fullPath));
        const fsEntry = this.fs.get(this.toPath(realFullPath));
        if (isFsSymLink(fsEntry)) {
            return this.realpath(fsEntry.symLink);
        }
        // realpath supports non-existent files, so there may not be an fsEntry
        return fsEntry?.fullPath || realFullPath;
    }
    exitCode;
    resolvePath = (s) => s;
    getExecutingFilePath = () => this.executingFilePath;
    getCurrentDirectory = () => this.currentDirectory;
    exit(exitCode) {
        this.exitCode = exitCode;
        throw new Error(exitMessage);
    }
    getEnvironmentVariable(name) {
        return this.environmentVariables && this.environmentVariables.get(name) || "";
    }
}
function diffFsFile(baseline, fsEntry, newInode) {
    if (!isFsLibFile(fsEntry)) {
        baseline.push(`//// [${fsEntry.fullPath}]${inodeString(newInode)}\r\n${fsEntry.content}`, "");
    }
    else if (fsEntry.content) {
        baseline.push(`//// [${fsEntry.fullPath}] *Lib*${inodeString(newInode)}`, "");
    }
}
function diffFsSymLink(baseline, fsEntry, newInode) {
    baseline.push(`//// [${fsEntry.fullPath}] symlink(${fsEntry.symLink})${inodeString(newInode)}`, "");
}
function inodeString(inode) {
    return inode !== undefined ? ` Inode:: ${inode}` : "";
}
function diffFsEntry(baseline, oldFsEntry, newFsEntry, newInode, writtenFiles) {
    const file = newFsEntry && newFsEntry.fullPath;
    if (isFsFileOrFsLibFile(oldFsEntry)) {
        if (isFsFileOrFsLibFile(newFsEntry)) {
            if (oldFsEntry.content !== newFsEntry.content) {
                diffFsFile(baseline, newFsEntry, newInode);
            }
            else if (oldFsEntry.modifiedTime !== newFsEntry.modifiedTime) {
                if (oldFsEntry.fullPath !== newFsEntry.fullPath) {
                    baseline.push(`//// [${file}] file was renamed from file ${oldFsEntry.fullPath}${inodeString(newInode)}`);
                }
                else if (writtenFiles && !writtenFiles.has(newFsEntry.path)) {
                    baseline.push(`//// [${file}] file changed its modified time${inodeString(newInode)}`);
                }
                else {
                    baseline.push(`//// [${file}] file written with same contents${inodeString(newInode)}`);
                }
            }
        }
        else {
            baseline.push(`//// [${oldFsEntry.fullPath}] deleted`);
            if (isFsSymLink(newFsEntry)) {
                diffFsSymLink(baseline, newFsEntry, newInode);
            }
        }
    }
    else if (isFsSymLink(oldFsEntry)) {
        if (isFsSymLink(newFsEntry)) {
            if (oldFsEntry.symLink !== newFsEntry.symLink) {
                diffFsSymLink(baseline, newFsEntry, newInode);
            }
            else if (oldFsEntry.modifiedTime !== newFsEntry.modifiedTime) {
                if (oldFsEntry.fullPath !== newFsEntry.fullPath) {
                    baseline.push(`//// [${file}] symlink was renamed from symlink ${oldFsEntry.fullPath}${inodeString(newInode)}`);
                }
                else if (writtenFiles && !writtenFiles.has(newFsEntry.path)) {
                    baseline.push(`//// [${file}] symlink changed its modified time${inodeString(newInode)}`);
                }
                else {
                    baseline.push(`//// [${file}] symlink written with same link${inodeString(newInode)}`);
                }
            }
        }
        else {
            baseline.push(`//// [${oldFsEntry.fullPath}] deleted symlink`);
            if (isFsFileOrFsLibFile(newFsEntry)) {
                diffFsFile(baseline, newFsEntry, newInode);
            }
        }
    }
    else if (isFsFileOrFsLibFile(newFsEntry)) {
        diffFsFile(baseline, newFsEntry, newInode);
    }
    else if (isFsSymLink(newFsEntry)) {
        diffFsSymLink(baseline, newFsEntry, newInode);
    }
}
function baselineOutputs(baseline, output, start, end = output.length) {
    let baselinedOutput;
    for (let i = start; i < end; i++) {
        baselinedOutput = append(baselinedOutput, sanitizeSysOutput(output[i]));
    }
    if (baselinedOutput)
        baseline.push(baselinedOutput.join(""));
}
function sanitizeSysOutput(output) {
    return output
        .replace(/Elapsed::\s\d+(?:\.\d+)?ms/g, "Elapsed:: *ms")
        .replace(/\d\d:\d\d:\d\d\s(?:A|P)M/g, "HH:MM:SS AM");
}
export function changeToHostTrackingWrittenFiles(inputHost) {
    const host = inputHost;
    if (host.writtenFiles)
        return host;
    const originalWriteFile = host.writeFile;
    host.writtenFiles = new Map();
    host.writeFile = (fileName, content) => {
        originalWriteFile.call(host, fileName, content);
        const path = host.toFullPath(fileName);
        host.writtenFiles.set(path, (host.writtenFiles.get(path) || 0) + 1);
    };
    return host;
}
export function osFlavorToString(osFlavor) {
    switch (osFlavor) {
        case TestServerHostOsFlavor.Windows:
            return "Windows";
        case TestServerHostOsFlavor.MacOs:
            return "MacOs";
        case TestServerHostOsFlavor.Linux:
            return "Linux";
        default:
            Debug.assertNever(osFlavor);
    }
}
