import { multiMapAddValue, multiMapRemoveValue, AssertionLevel, closeFileWatcher, closeFileWatcherOf, combinePaths, Comparison, contains, containsPath, createGetCanonicalFileName, createMultiMap, Debug, directorySeparator, emptyArray, emptyFileSystemEntries, endsWith, enumerateInsertsAndDeletes, getDirectoryPath, getFallbackOptions, getNormalizedAbsolutePath, getRelativePathFromDirectory, getRelativePathToDirectoryOrUrl, getRootLength, getStringComparer, isArray, isNodeLikeSystem, isString, mapDefined, matchesExclude, matchFiles, memoize, noop, normalizePath, normalizeSlashes, orderedRemoveItem, PollingWatchKind, resolveJSModule, some, startsWith, timestamp, unorderedRemoveItem, WatchDirectoryKind, WatchFileKind, writeFileEnsuringDirectories, } from "./_namespaces/ts.js";
import { errorCode, propertyOf } from "./hostErrors.js";
import * as _fs from "node:fs";
import * as _path from "node:path";
import * as _os from "node:os";
import * as _crypto from "node:crypto";
/**
 * djb2 hashing algorithm
 * http://www.cse.yorku.ca/~oz/hash.html
 *
 * @internal
 */
export function generateDjb2Hash(data) {
    let acc = 5381;
    for (let i = 0; i < data.length; i++) {
        acc = ((acc << 5) + acc) + data.charCodeAt(i);
    }
    return acc.toString();
}
/**
 * Set a high stack trace limit to provide more information in case of an error.
 * Called for command-line and server use cases.
 * Not called if TypeScript is used as a library.
 *
 * @internal
 */
export function setStackTraceLimit() {
    if (Error.stackTraceLimit < 100) { // Also tests that we won't set the property if it doesn't exist.
        Error.stackTraceLimit = 100;
    }
}
export var FileWatcherEventKind;
(function (FileWatcherEventKind) {
    FileWatcherEventKind[FileWatcherEventKind["Created"] = 0] = "Created";
    FileWatcherEventKind[FileWatcherEventKind["Changed"] = 1] = "Changed";
    FileWatcherEventKind[FileWatcherEventKind["Deleted"] = 2] = "Deleted";
})(FileWatcherEventKind || (FileWatcherEventKind = {}));
/** @internal */
export var PollingInterval;
(function (PollingInterval) {
    PollingInterval[PollingInterval["High"] = 2000] = "High";
    PollingInterval[PollingInterval["Medium"] = 500] = "Medium";
    PollingInterval[PollingInterval["Low"] = 250] = "Low";
})(PollingInterval || (PollingInterval = {}));
/** @internal */
export const missingFileModifiedTime = new Date(0); // Any subsequent modification will occur after this time
/** @internal */
export function getModifiedTime(host, fileName) {
    return host.getModifiedTime(fileName) || missingFileModifiedTime;
}
function createPollingIntervalBasedLevels(levels) {
    return {
        [PollingInterval.Low]: levels.Low,
        [PollingInterval.Medium]: levels.Medium,
        [PollingInterval.High]: levels.High,
    };
}
const defaultChunkLevels = { Low: 32, Medium: 64, High: 256 };
let pollingChunkSize = createPollingIntervalBasedLevels(defaultChunkLevels);
/** @internal */
export let unchangedPollThresholds = createPollingIntervalBasedLevels(defaultChunkLevels);
function setCustomPollingValues(system) {
    if (!system.getEnvironmentVariable) {
        return;
    }
    const pollingIntervalChanged = !!getCustomLevels("TSC_WATCH_POLLINGINTERVAL");
    pollingChunkSize = getCustomPollingBasedLevels("TSC_WATCH_POLLINGCHUNKSIZE", defaultChunkLevels) || pollingChunkSize;
    unchangedPollThresholds = getCustomPollingBasedLevels("TSC_WATCH_UNCHANGEDPOLLTHRESHOLDS", defaultChunkLevels) || unchangedPollThresholds;
    function getLevel(envVar, level) {
        return system.getEnvironmentVariable(`${envVar}_${level.toUpperCase()}`);
    }
    function getCustomLevels(baseVariable) {
        let customLevels;
        setCustomLevel("Low");
        setCustomLevel("Medium");
        setCustomLevel("High");
        return customLevels;
        function setCustomLevel(level) {
            const customLevel = getLevel(baseVariable, level);
            if (customLevel) {
                (customLevels || (customLevels = {}))[level] = Number(customLevel);
            }
        }
    }
    function getCustomPollingBasedLevels(baseVariable, defaultLevels) {
        const customLevels = getCustomLevels(baseVariable);
        return (pollingIntervalChanged || customLevels) &&
            createPollingIntervalBasedLevels(customLevels ? { ...defaultLevels, ...customLevels } : defaultLevels);
    }
}
function pollWatchedFileQueue(host, queue, pollIndex, chunkSize, callbackOnWatchFileStat) {
    let definedValueCopyToIndex = pollIndex;
    // Max visit would be all elements of the queue
    for (let canVisit = queue.length; chunkSize && canVisit; nextPollIndex(), canVisit--) {
        const watchedFile = queue[pollIndex];
        if (!watchedFile) {
            continue;
        }
        else if (watchedFile.isClosed) {
            queue[pollIndex] = undefined;
            continue;
        }
        // Only files polled count towards chunkSize
        chunkSize--;
        const fileChanged = onWatchedFileStat(watchedFile, getModifiedTime(host, watchedFile.fileName));
        if (watchedFile.isClosed) {
            // Closed watcher as part of callback
            queue[pollIndex] = undefined;
            continue;
        }
        callbackOnWatchFileStat?.(watchedFile, pollIndex, fileChanged);
        // Defragment the queue while we are at it
        if (queue[pollIndex]) {
            // Copy this file to the non hole location
            if (definedValueCopyToIndex < pollIndex) {
                queue[definedValueCopyToIndex] = watchedFile;
                queue[pollIndex] = undefined;
            }
            definedValueCopyToIndex++;
        }
    }
    // Return next poll index
    return pollIndex;
    function nextPollIndex() {
        pollIndex++;
        if (pollIndex === queue.length) {
            if (definedValueCopyToIndex < pollIndex) {
                // There are holes from definedValueCopyToIndex to end of queue, change queue size
                queue.length = definedValueCopyToIndex;
            }
            pollIndex = 0;
            definedValueCopyToIndex = 0;
        }
    }
}
function createDynamicPriorityPollingWatchFile(host) {
    const watchedFiles = [];
    const changedFilesInLastPoll = [];
    const lowPollingIntervalQueue = createPollingIntervalQueue(PollingInterval.Low);
    const mediumPollingIntervalQueue = createPollingIntervalQueue(PollingInterval.Medium);
    const highPollingIntervalQueue = createPollingIntervalQueue(PollingInterval.High);
    return watchFile;
    function watchFile(fileName, callback, defaultPollingInterval) {
        const file = {
            fileName,
            callback,
            unchangedPolls: 0,
            mtime: getModifiedTime(host, fileName),
        };
        watchedFiles.push(file);
        addToPollingIntervalQueue(file, defaultPollingInterval);
        return {
            close: () => {
                file.isClosed = true;
                // Remove from watchedFiles
                unorderedRemoveItem(watchedFiles, file);
                // Do not update polling interval queue since that will happen as part of polling
            },
        };
    }
    function createPollingIntervalQueue(pollingInterval) {
        return { files: [], pollingInterval, pollIndex: 0, pollScheduled: false };
    }
    function pollPollingIntervalQueue(_timeoutType, queue) {
        queue.pollIndex = pollQueue(queue.files, queue.pollingInterval, queue.pollIndex, pollingChunkSize[queue.pollingInterval]);
        // Set the next polling index and timeout
        if (queue.files.length) {
            scheduleNextPoll(queue.pollingInterval);
        }
        else {
            Debug.assert(queue.pollIndex === 0);
            queue.pollScheduled = false;
        }
    }
    function pollLowPollingIntervalQueue(_timeoutType, queue) {
        // Always poll complete list of changedFilesInLastPoll
        pollQueue(changedFilesInLastPoll, PollingInterval.Low, /*pollIndex*/ 0, changedFilesInLastPoll.length);
        // Finally do the actual polling of the queue
        pollPollingIntervalQueue(_timeoutType, queue);
        // Schedule poll if there are files in changedFilesInLastPoll but no files in the actual queue
        // as pollPollingIntervalQueue wont schedule for next poll
        if (!queue.pollScheduled && changedFilesInLastPoll.length) {
            scheduleNextPoll(PollingInterval.Low);
        }
    }
    function pollQueue(queue, pollingInterval, pollIndex, chunkSize) {
        return pollWatchedFileQueue(host, queue, pollIndex, chunkSize, onWatchFileStat);
        function onWatchFileStat(watchedFile, pollIndex, fileChanged) {
            if (fileChanged) {
                watchedFile.unchangedPolls = 0;
                // Changed files go to changedFilesInLastPoll queue
                if (queue !== changedFilesInLastPoll) {
                    queue[pollIndex] = undefined;
                    addChangedFileToLowPollingIntervalQueue(watchedFile);
                }
            }
            else if (watchedFile.unchangedPolls !== unchangedPollThresholds[pollingInterval]) {
                watchedFile.unchangedPolls++;
            }
            else if (queue === changedFilesInLastPoll) {
                // Restart unchangedPollCount for unchanged file and move to low polling interval queue
                watchedFile.unchangedPolls = 1;
                queue[pollIndex] = undefined;
                addToPollingIntervalQueue(watchedFile, PollingInterval.Low);
            }
            else if (pollingInterval !== PollingInterval.High) {
                watchedFile.unchangedPolls++;
                queue[pollIndex] = undefined;
                addToPollingIntervalQueue(watchedFile, pollingInterval === PollingInterval.Low ? PollingInterval.Medium : PollingInterval.High);
            }
        }
    }
    function pollingIntervalQueue(pollingInterval) {
        switch (pollingInterval) {
            case PollingInterval.Low:
                return lowPollingIntervalQueue;
            case PollingInterval.Medium:
                return mediumPollingIntervalQueue;
            case PollingInterval.High:
                return highPollingIntervalQueue;
        }
    }
    function addToPollingIntervalQueue(file, pollingInterval) {
        pollingIntervalQueue(pollingInterval).files.push(file);
        scheduleNextPollIfNotAlreadyScheduled(pollingInterval);
    }
    function addChangedFileToLowPollingIntervalQueue(file) {
        changedFilesInLastPoll.push(file);
        scheduleNextPollIfNotAlreadyScheduled(PollingInterval.Low);
    }
    function scheduleNextPollIfNotAlreadyScheduled(pollingInterval) {
        if (!pollingIntervalQueue(pollingInterval).pollScheduled) {
            scheduleNextPoll(pollingInterval);
        }
    }
    function scheduleNextPoll(pollingInterval) {
        pollingIntervalQueue(pollingInterval).pollScheduled = (() => {
            const timerHost = host;
            const callback = pollingInterval === PollingInterval.Low ? pollLowPollingIntervalQueue : pollPollingIntervalQueue;
            const timeoutDelay = pollingInterval;
            const t1 = pollingInterval === PollingInterval.Low ? "pollLowPollingIntervalQueue" : "pollPollingIntervalQueue";
            const t2 = pollingIntervalQueue(pollingInterval);
            return timerHost.setTimeout(() => callback(t1, t2), timeoutDelay, t1);
        })();
    }
}
function createUseFsEventsOnParentDirectoryWatchFile(fsWatch, useCaseSensitiveFileNames, getModifiedTime, fsWatchWithTimestamp) {
    // One file can have multiple watchers
    const fileWatcherCallbacks = createMultiMap();
    const fileTimestamps = fsWatchWithTimestamp ? new Map() : undefined;
    const dirWatchers = new Map();
    const toCanonicalName = createGetCanonicalFileName(useCaseSensitiveFileNames);
    return nonPollingWatchFile;
    function nonPollingWatchFile(fileName, callback, _pollingInterval, fallbackOptions) {
        const filePath = toCanonicalName(fileName);
        if (multiMapAddValue(fileWatcherCallbacks, filePath, callback).length === 1 && fileTimestamps) {
            fileTimestamps.set(filePath, getModifiedTime(fileName) || missingFileModifiedTime);
        }
        const dirPath = getDirectoryPath(filePath) || ".";
        const watcher = dirWatchers.get(dirPath) ||
            createDirectoryWatcher(getDirectoryPath(fileName) || ".", dirPath, fallbackOptions);
        watcher.referenceCount++;
        return {
            close: () => {
                if (watcher.referenceCount === 1) {
                    watcher.close();
                    dirWatchers.delete(dirPath);
                }
                else {
                    watcher.referenceCount--;
                }
                multiMapRemoveValue(fileWatcherCallbacks, filePath, callback);
            },
        };
    }
    function createDirectoryWatcher(dirName, dirPath, fallbackOptions) {
        const watcher = fsWatch(dirName, FileSystemEntryKind.Directory, (eventName, relativeFileName) => {
            // When files are deleted from disk, the triggered "rename" event would have a relativefileName of "undefined"
            if (!isString(relativeFileName))
                return;
            const fileName = getNormalizedAbsolutePath(relativeFileName, dirName);
            const filePath = toCanonicalName(fileName);
            // Some applications save a working file via rename operations
            const callbacks = fileName && fileWatcherCallbacks.get(filePath);
            if (callbacks) {
                let currentModifiedTime;
                let eventKind = FileWatcherEventKind.Changed;
                if (fileTimestamps) {
                    const existingTime = fileTimestamps.get(filePath);
                    if (eventName === "change") {
                        currentModifiedTime = getModifiedTime(fileName) || missingFileModifiedTime;
                        if (currentModifiedTime.getTime() === existingTime.getTime())
                            return;
                    }
                    currentModifiedTime ||= getModifiedTime(fileName) || missingFileModifiedTime;
                    fileTimestamps.set(filePath, currentModifiedTime);
                    if (existingTime === missingFileModifiedTime)
                        eventKind = FileWatcherEventKind.Created;
                    else if (currentModifiedTime === missingFileModifiedTime)
                        eventKind = FileWatcherEventKind.Deleted;
                }
                for (const fileCallback of callbacks) {
                    fileCallback(fileName, eventKind, currentModifiedTime);
                }
            }
        }, 
        /*recursive*/ false, PollingInterval.Medium, fallbackOptions);
        watcher.referenceCount = 0;
        dirWatchers.set(dirPath, watcher);
        return watcher;
    }
}
function createFixedChunkSizePollingWatchFile(host) {
    const watchedFiles = [];
    let pollIndex = 0;
    let pollScheduled;
    return watchFile;
    function watchFile(fileName, callback) {
        const file = {
            fileName,
            callback,
            mtime: getModifiedTime(host, fileName),
        };
        watchedFiles.push(file);
        scheduleNextPoll();
        return {
            close: () => {
                file.isClosed = true;
                unorderedRemoveItem(watchedFiles, file);
            },
        };
    }
    function pollQueue(_timeoutType) {
        pollScheduled = undefined;
        pollIndex = pollWatchedFileQueue(host, watchedFiles, pollIndex, pollingChunkSize[PollingInterval.Low]);
        scheduleNextPoll();
    }
    function scheduleNextPoll() {
        if (!watchedFiles.length || pollScheduled)
            return;
        pollScheduled = (() => {
            const timerHost = host;
            const callback = pollQueue;
            const timeoutDelay = PollingInterval.High;
            const t1 = "pollQueue";
            return timerHost.setTimeout(() => callback(t1), timeoutDelay, t1);
        })();
    }
}
function createSingleWatcherPerName(cache, useCaseSensitiveFileNames, name, callback, createWatcher, createForwarder) {
    const toCanonicalFileName = createGetCanonicalFileName(useCaseSensitiveFileNames);
    const path = toCanonicalFileName(name);
    const existing = cache.get(path);
    if (existing) {
        existing.callbacks.push(callback);
    }
    else {
        cache.set(path, {
            watcher: createWatcher((
            // Cant infer types correctly so lets satisfy checker
            (param1, param2, param3) => cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3)))),
            callbacks: [callback],
        });
    }
    return {
        close: () => {
            const watcher = cache.get(path);
            // Watcher is not expected to be undefined, but if it is normally its because
            // exception was thrown somewhere else and watch state is not what it should be
            if (!watcher)
                return;
            if (!orderedRemoveItem(watcher.callbacks, callback) || watcher.callbacks.length)
                return;
            cache.delete(path);
            closeFileWatcherOf(watcher);
        },
    };
}
/**
 * Returns true if file status changed
 */
function onWatchedFileStat(watchedFile, modifiedTime) {
    const oldTime = watchedFile.mtime.getTime();
    const newTime = modifiedTime.getTime();
    if (oldTime !== newTime) {
        watchedFile.mtime = modifiedTime;
        // Pass modified times so tsc --build can use it
        watchedFile.callback(watchedFile.fileName, getFileWatcherEventKind(oldTime, newTime), modifiedTime);
        return true;
    }
    return false;
}
/** @internal */
export function getFileWatcherEventKind(oldTime, newTime) {
    return oldTime === 0
        ? FileWatcherEventKind.Created
        : newTime === 0
            ? FileWatcherEventKind.Deleted
            : FileWatcherEventKind.Changed;
}
/** @internal */
export const ignoredPaths = ["/node_modules/.", "/.git", "/.#"];
let curSysLog = noop;
/** @internal */
export function sysLog(s) {
    return curSysLog(s);
}
/** @internal */
export function setSysLog(logger) {
    curSysLog = logger;
}
/**
 * Watch the directory recursively using host provided method to watch child directories
 * that means if this is recursive watcher, watch the children directories as well
 * (eg on OS that dont support recursive watch using fs.watch use fs.watchFile)
 */
function createDirectoryWatcherSupportingRecursive({ watchDirectory, useCaseSensitiveFileNames, getCurrentDirectory, getAccessibleSortedChildDirectories, fileSystemEntryExists, realpath, setTimeout, clearTimeout, }) {
    const cache = new Map();
    const callbackCache = createMultiMap();
    const cacheToUpdateChildWatches = new Map();
    let timerToUpdateChildWatches;
    const filePathComparer = getStringComparer(!useCaseSensitiveFileNames);
    const toCanonicalFilePath = createGetCanonicalFileName(useCaseSensitiveFileNames);
    return (dirName, callback, recursive, options) => recursive ?
        createDirectoryWatcher(dirName, options, callback) :
        watchDirectory(dirName, callback, recursive, options);
    /**
     * Create the directory watcher for the dirPath.
     */
    function createDirectoryWatcher(dirName, options, callback, link) {
        const dirPath = toCanonicalFilePath(dirName);
        let directoryWatcher = cache.get(dirPath);
        if (directoryWatcher) {
            directoryWatcher.refCount++;
        }
        else {
            directoryWatcher = {
                watcher: watchDirectory(dirName, fileName => {
                    if (isIgnoredPath(fileName, options))
                        return;
                    if (options?.synchronousWatchDirectory) {
                        // Call the actual callback
                        if (!cache.get(dirPath)?.targetWatcher)
                            invokeCallbacks(dirName, dirPath, fileName);
                        // Iterate through existing children and update the watches if needed
                        updateChildWatches(dirName, dirPath, options);
                    }
                    else {
                        nonSyncUpdateChildWatches(dirName, dirPath, fileName, options);
                    }
                }, 
                /*recursive*/ false, options),
                refCount: 1,
                childWatches: emptyArray,
                targetWatcher: undefined,
                links: undefined,
            };
            cache.set(dirPath, directoryWatcher);
            updateChildWatches(dirName, dirPath, options);
        }
        if (link)
            (directoryWatcher.links ??= new Set()).add(link);
        const callbackToAdd = callback && { dirName, callback };
        if (callbackToAdd) {
            multiMapAddValue(callbackCache, dirPath, callbackToAdd);
        }
        return {
            dirName,
            close: () => {
                const directoryWatcher = Debug.checkDefined(cache.get(dirPath));
                if (callbackToAdd)
                    multiMapRemoveValue(callbackCache, dirPath, callbackToAdd);
                if (link)
                    directoryWatcher.links?.delete(link);
                directoryWatcher.refCount--;
                if (directoryWatcher.refCount)
                    return;
                cache.delete(dirPath);
                directoryWatcher.links = undefined;
                closeFileWatcherOf(directoryWatcher);
                closeTargetWatcher(directoryWatcher);
                directoryWatcher.childWatches.forEach(closeFileWatcher);
            },
        };
    }
    function invokeCallbacks(dirName, dirPath, fileNameOrInvokeMap, fileNames) {
        let fileName;
        let invokeMap;
        if (isString(fileNameOrInvokeMap)) {
            fileName = fileNameOrInvokeMap;
        }
        else {
            invokeMap = fileNameOrInvokeMap;
        }
        // Call the actual callback
        callbackCache.forEach((callbacks, rootDirName) => {
            if (invokeMap && invokeMap.get(rootDirName) === true)
                return;
            if (rootDirName === dirPath || (startsWith(dirPath, rootDirName) && dirPath[rootDirName.length] === directorySeparator)) {
                if (invokeMap) {
                    if (fileNames) {
                        const existing = invokeMap.get(rootDirName);
                        if (existing) {
                            if (Array.isArray(existing))
                                existing.push(...fileNames);
                        }
                        else {
                            invokeMap.set(rootDirName, fileNames.slice());
                        }
                    }
                    else {
                        invokeMap.set(rootDirName, true);
                    }
                }
                else {
                    callbacks.forEach(({ callback }) => callback(fileName));
                }
            }
        });
        cache.get(dirPath)?.links?.forEach(link => {
            const toPathInLink = (fileName) => combinePaths(link, getRelativePathFromDirectory(dirName, fileName, toCanonicalFilePath));
            if (invokeMap) {
                invokeCallbacks(link, toCanonicalFilePath(link), invokeMap, fileNames?.map(toPathInLink));
            }
            else {
                invokeCallbacks(link, toCanonicalFilePath(link), toPathInLink(fileName));
            }
        });
    }
    function nonSyncUpdateChildWatches(dirName, dirPath, fileName, options) {
        // Iterate through existing children and update the watches if needed
        const parentWatcher = cache.get(dirPath);
        if (parentWatcher && fileSystemEntryExists(dirName, FileSystemEntryKind.Directory)) {
            // Schedule the update and postpone invoke for callbacks
            scheduleUpdateChildWatches(dirName, dirPath, fileName, options);
            return;
        }
        // Call the actual callbacks and remove child watches
        invokeCallbacks(dirName, dirPath, fileName);
        closeTargetWatcher(parentWatcher);
        removeChildWatches(parentWatcher);
    }
    function scheduleUpdateChildWatches(dirName, dirPath, fileName, options) {
        const existing = cacheToUpdateChildWatches.get(dirPath);
        if (existing) {
            existing.fileNames.push(fileName);
        }
        else {
            cacheToUpdateChildWatches.set(dirPath, { dirName, options, fileNames: [fileName] });
        }
        if (timerToUpdateChildWatches) {
            clearTimeout(timerToUpdateChildWatches);
            timerToUpdateChildWatches = undefined;
        }
        timerToUpdateChildWatches = (() => {
            const callback = onTimerToUpdateChildWatches;
            const timeoutDelay = 1000;
            const t1 = "timerToUpdateChildWatches";
            return setTimeout(() => callback(t1), timeoutDelay, t1);
        })();
    }
    function onTimerToUpdateChildWatches(_timeoutType) {
        timerToUpdateChildWatches = undefined;
        sysLog(`sysLog:: onTimerToUpdateChildWatches:: ${cacheToUpdateChildWatches.size}`);
        const start = timestamp();
        const invokeMap = new Map();
        while (!timerToUpdateChildWatches && cacheToUpdateChildWatches.size) {
            const result = cacheToUpdateChildWatches.entries().next();
            Debug.assert(!result.done);
            const { value: [dirPath, { dirName, options, fileNames }] } = result;
            cacheToUpdateChildWatches.delete(dirPath);
            // Because the child refresh is fresh, we would need to invalidate whole root directory being watched
            // to ensure that all the changes are reflected at this time
            const hasChanges = updateChildWatches(dirName, dirPath, options);
            if (!cache.get(dirPath)?.targetWatcher)
                invokeCallbacks(dirName, dirPath, invokeMap, hasChanges ? undefined : fileNames);
        }
        sysLog(`sysLog:: invokingWatchers:: Elapsed:: ${timestamp() - start}ms:: ${cacheToUpdateChildWatches.size}`);
        callbackCache.forEach((callbacks, rootDirName) => {
            const existing = invokeMap.get(rootDirName);
            if (existing) {
                callbacks.forEach(({ callback, dirName }) => {
                    if (isArray(existing)) {
                        existing.forEach(callback);
                    }
                    else {
                        callback(dirName);
                    }
                });
            }
        });
        const elapsed = timestamp() - start;
        sysLog(`sysLog:: Elapsed:: ${elapsed}ms:: onTimerToUpdateChildWatches:: ${cacheToUpdateChildWatches.size} ${timerToUpdateChildWatches}`);
    }
    function removeChildWatches(parentWatcher) {
        if (!parentWatcher)
            return;
        const existingChildWatches = parentWatcher.childWatches;
        parentWatcher.childWatches = emptyArray;
        for (const childWatcher of existingChildWatches) {
            childWatcher.close();
            removeChildWatches(cache.get(toCanonicalFilePath(childWatcher.dirName)));
        }
    }
    function closeTargetWatcher(watcher) {
        if (watcher?.targetWatcher) {
            watcher.targetWatcher.close();
            watcher.targetWatcher = undefined;
        }
    }
    function updateChildWatches(parentDir, parentDirPath, options) {
        // Iterate through existing children and update the watches if needed
        const parentWatcher = cache.get(parentDirPath);
        if (!parentWatcher)
            return false;
        const target = normalizePath(realpath(parentDir));
        let hasChanges;
        let newChildWatches;
        if (filePathComparer(target, parentDir) === Comparison.EqualTo) {
            // if (parentWatcher.target) closeFileWatcher
            hasChanges = enumerateInsertsAndDeletes(fileSystemEntryExists(parentDir, FileSystemEntryKind.Directory) ? mapDefined(getAccessibleSortedChildDirectories(parentDir), child => {
                const childFullName = getNormalizedAbsolutePath(child, parentDir);
                // Filter our the symbolic link directories since those arent included in recursive watch
                // which is same behaviour when recursive: true is passed to fs.watch
                return !isIgnoredPath(childFullName, options) && filePathComparer(childFullName, normalizePath(realpath(childFullName))) === Comparison.EqualTo ? childFullName : undefined;
            }) : emptyArray, parentWatcher.childWatches, (child, childWatcher) => filePathComparer(child, childWatcher.dirName), createAndAddChildDirectoryWatcher, closeFileWatcher, addChildDirectoryWatcher);
        }
        else if (parentWatcher.targetWatcher && filePathComparer(target, parentWatcher.targetWatcher.dirName) === Comparison.EqualTo) {
            hasChanges = false;
            Debug.assert(parentWatcher.childWatches === emptyArray);
        }
        else {
            closeTargetWatcher(parentWatcher);
            parentWatcher.targetWatcher = createDirectoryWatcher(target, options, /*callback*/ undefined, parentDir);
            parentWatcher.childWatches.forEach(closeFileWatcher);
            hasChanges = true;
        }
        parentWatcher.childWatches = newChildWatches || emptyArray;
        return hasChanges;
        /**
         * Create new childDirectoryWatcher and add it to the new ChildDirectoryWatcher list
         */
        function createAndAddChildDirectoryWatcher(childName) {
            const result = createDirectoryWatcher(childName, options);
            addChildDirectoryWatcher(result);
        }
        /**
         * Add child directory watcher to the new ChildDirectoryWatcher list
         */
        function addChildDirectoryWatcher(childWatcher) {
            (newChildWatches || (newChildWatches = [])).push(childWatcher);
        }
    }
    function isIgnoredPath(path, options) {
        return some(ignoredPaths, searchPath => isInPath(path, searchPath)) ||
            isIgnoredByWatchOptions(path, options, useCaseSensitiveFileNames, getCurrentDirectory);
    }
    function isInPath(path, searchPath) {
        if (path.includes(searchPath))
            return true;
        if (useCaseSensitiveFileNames)
            return false;
        return toCanonicalFilePath(path).includes(searchPath);
    }
}
/** @internal */
export var FileSystemEntryKind;
(function (FileSystemEntryKind) {
    FileSystemEntryKind[FileSystemEntryKind["File"] = 0] = "File";
    FileSystemEntryKind[FileSystemEntryKind["Directory"] = 1] = "Directory";
})(FileSystemEntryKind || (FileSystemEntryKind = {}));
function createFileWatcherCallback(callback) {
    return (_fileName, eventKind, modifiedTime) => callback(eventKind === FileWatcherEventKind.Changed ? "change" : "rename", "", modifiedTime);
}
function createFsWatchCallbackForFileWatcherCallback(fileName, callback, getModifiedTime) {
    return (eventName, _relativeFileName, modifiedTime) => {
        if (eventName === "rename") {
            // Check time stamps rather than file system entry checks
            modifiedTime ||= getModifiedTime(fileName) || missingFileModifiedTime;
            callback(fileName, modifiedTime !== missingFileModifiedTime ? FileWatcherEventKind.Created : FileWatcherEventKind.Deleted, modifiedTime);
        }
        else {
            // Change
            callback(fileName, FileWatcherEventKind.Changed, modifiedTime);
        }
    };
}
function isIgnoredByWatchOptions(pathToCheck, options, useCaseSensitiveFileNames, getCurrentDirectory) {
    return (options?.excludeDirectories || options?.excludeFiles) && (matchesExclude(pathToCheck, options?.excludeFiles, useCaseSensitiveFileNames, getCurrentDirectory()) ||
        matchesExclude(pathToCheck, options?.excludeDirectories, useCaseSensitiveFileNames, getCurrentDirectory()));
}
function createFsWatchCallbackForDirectoryWatcherCallback(directoryName, callback, options, useCaseSensitiveFileNames, getCurrentDirectory) {
    return (eventName, relativeFileName) => {
        // In watchDirectory we only care about adding and removing files (when event name is
        // "rename"); changes made within files are handled by corresponding fileWatchers (when
        // event name is "change")
        if (eventName === "rename") {
            // When deleting a file, the passed baseFileName is null
            const fileName = !relativeFileName ? directoryName : normalizePath(combinePaths(directoryName, relativeFileName));
            if (!relativeFileName || !isIgnoredByWatchOptions(fileName, options, useCaseSensitiveFileNames, getCurrentDirectory)) {
                callback(fileName);
            }
        }
    };
}
/** @internal */
export function createSystemWatchFunctions({ pollingWatchFileWorker, getModifiedTime, setTimeout, clearTimeout, fsWatchWorker, fileSystemEntryExists, useCaseSensitiveFileNames, getCurrentDirectory, fsSupportsRecursiveFsWatch, getAccessibleSortedChildDirectories, realpath, tscWatchFile, useNonPollingWatchers, tscWatchDirectory, inodeWatching, fsWatchWithTimestamp, sysLog, }) {
    const pollingWatches = new Map();
    const fsWatches = new Map();
    const fsWatchesRecursive = new Map();
    let dynamicPollingWatchFile;
    let fixedChunkSizePollingWatchFile;
    let nonPollingWatchFile;
    let hostRecursiveDirectoryWatcher;
    let hitSystemWatcherLimit = false;
    return {
        watchFile,
        watchDirectory,
    };
    function watchFile(fileName, callback, pollingInterval, options) {
        options = updateOptionsForWatchFile(options, useNonPollingWatchers);
        const watchFileKind = Debug.checkDefined(options.watchFile);
        switch (watchFileKind) {
            case WatchFileKind.FixedPollingInterval:
                return pollingWatchFile(fileName, callback, PollingInterval.Low, /*options*/ undefined);
            case WatchFileKind.PriorityPollingInterval:
                return pollingWatchFile(fileName, callback, pollingInterval, /*options*/ undefined);
            case WatchFileKind.DynamicPriorityPolling:
                return ensureDynamicPollingWatchFile()(fileName, callback, pollingInterval, /*options*/ undefined);
            case WatchFileKind.FixedChunkSizePolling:
                return ensureFixedChunkSizePollingWatchFile()(fileName, callback, /* pollingInterval */ undefined, /*options*/ undefined);
            case WatchFileKind.UseFsEvents:
                return fsWatch(fileName, FileSystemEntryKind.File, createFsWatchCallbackForFileWatcherCallback(fileName, callback, getModifiedTime), 
                /*recursive*/ false, pollingInterval, getFallbackOptions(options));
            case WatchFileKind.UseFsEventsOnParentDirectory:
                if (!nonPollingWatchFile) {
                    nonPollingWatchFile = createUseFsEventsOnParentDirectoryWatchFile(fsWatch, useCaseSensitiveFileNames, getModifiedTime, fsWatchWithTimestamp);
                }
                return nonPollingWatchFile(fileName, callback, pollingInterval, getFallbackOptions(options));
            default:
                Debug.assertNever(watchFileKind);
        }
    }
    function ensureDynamicPollingWatchFile() {
        return dynamicPollingWatchFile ||= createDynamicPriorityPollingWatchFile({ getModifiedTime, setTimeout });
    }
    function ensureFixedChunkSizePollingWatchFile() {
        return fixedChunkSizePollingWatchFile ||= createFixedChunkSizePollingWatchFile({ getModifiedTime, setTimeout });
    }
    function updateOptionsForWatchFile(options, useNonPollingWatchers) {
        if (options && options.watchFile !== undefined)
            return options;
        switch (tscWatchFile) {
            case "PriorityPollingInterval":
                // Use polling interval based on priority when create watch using host.watchFile
                return { watchFile: WatchFileKind.PriorityPollingInterval };
            case "DynamicPriorityPolling":
                // Use polling interval but change the interval depending on file changes and their default polling interval
                return { watchFile: WatchFileKind.DynamicPriorityPolling };
            case "UseFsEvents":
                // Use notifications from FS to watch with falling back to fs.watchFile
                return generateWatchFileOptions(WatchFileKind.UseFsEvents, PollingWatchKind.PriorityInterval, options);
            case "UseFsEventsWithFallbackDynamicPolling":
                // Use notifications from FS to watch with falling back to dynamic watch file
                return generateWatchFileOptions(WatchFileKind.UseFsEvents, PollingWatchKind.DynamicPriority, options);
            case "UseFsEventsOnParentDirectory":
                useNonPollingWatchers = true;
            // fall through
            default:
                return useNonPollingWatchers ?
                    // Use notifications from FS to watch with falling back to fs.watchFile
                    generateWatchFileOptions(WatchFileKind.UseFsEventsOnParentDirectory, PollingWatchKind.PriorityInterval, options) :
                    // Default to using fs events
                    { watchFile: WatchFileKind.UseFsEvents };
        }
    }
    function generateWatchFileOptions(watchFile, fallbackPolling, options) {
        const defaultFallbackPolling = options?.fallbackPolling;
        return {
            watchFile,
            fallbackPolling: defaultFallbackPolling === undefined ?
                fallbackPolling :
                defaultFallbackPolling,
        };
    }
    function watchDirectory(directoryName, callback, recursive, options) {
        if (fsSupportsRecursiveFsWatch) {
            return fsWatch(directoryName, FileSystemEntryKind.Directory, createFsWatchCallbackForDirectoryWatcherCallback(directoryName, callback, options, useCaseSensitiveFileNames, getCurrentDirectory), recursive, PollingInterval.Medium, getFallbackOptions(options));
        }
        if (!hostRecursiveDirectoryWatcher) {
            hostRecursiveDirectoryWatcher = createDirectoryWatcherSupportingRecursive({
                useCaseSensitiveFileNames,
                getCurrentDirectory,
                fileSystemEntryExists,
                getAccessibleSortedChildDirectories,
                watchDirectory: nonRecursiveWatchDirectory,
                realpath,
                setTimeout,
                clearTimeout,
            });
        }
        return hostRecursiveDirectoryWatcher(directoryName, callback, recursive, options);
    }
    function nonRecursiveWatchDirectory(directoryName, callback, recursive, options) {
        Debug.assert(!recursive);
        const watchDirectoryOptions = updateOptionsForWatchDirectory(options);
        const watchDirectoryKind = Debug.checkDefined(watchDirectoryOptions.watchDirectory);
        switch (watchDirectoryKind) {
            case WatchDirectoryKind.FixedPollingInterval:
                return pollingWatchFile(directoryName, () => callback(directoryName), PollingInterval.Medium, 
                /*options*/ undefined);
            case WatchDirectoryKind.DynamicPriorityPolling:
                return ensureDynamicPollingWatchFile()(directoryName, () => callback(directoryName), PollingInterval.Medium, 
                /*options*/ undefined);
            case WatchDirectoryKind.FixedChunkSizePolling:
                return ensureFixedChunkSizePollingWatchFile()(directoryName, () => callback(directoryName), 
                /* pollingInterval */ undefined, 
                /*options*/ undefined);
            case WatchDirectoryKind.UseFsEvents:
                return fsWatch(directoryName, FileSystemEntryKind.Directory, createFsWatchCallbackForDirectoryWatcherCallback(directoryName, callback, options, useCaseSensitiveFileNames, getCurrentDirectory), recursive, PollingInterval.Medium, getFallbackOptions(watchDirectoryOptions));
            default:
                Debug.assertNever(watchDirectoryKind);
        }
    }
    function updateOptionsForWatchDirectory(options) {
        if (options && options.watchDirectory !== undefined)
            return options;
        switch (tscWatchDirectory) {
            case "RecursiveDirectoryUsingFsWatchFile":
                // Use polling interval based on priority when create watch using host.watchFile
                return { watchDirectory: WatchDirectoryKind.FixedPollingInterval };
            case "RecursiveDirectoryUsingDynamicPriorityPolling":
                // Use polling interval but change the interval depending on file changes and their default polling interval
                return { watchDirectory: WatchDirectoryKind.DynamicPriorityPolling };
            default:
                const defaultFallbackPolling = options?.fallbackPolling;
                return {
                    watchDirectory: WatchDirectoryKind.UseFsEvents,
                    fallbackPolling: defaultFallbackPolling !== undefined ?
                        defaultFallbackPolling :
                        undefined,
                };
        }
    }
    function pollingWatchFile(fileName, callback, pollingInterval, options) {
        return createSingleWatcherPerName(pollingWatches, useCaseSensitiveFileNames, fileName, callback, cb => pollingWatchFileWorker(fileName, cb, pollingInterval, options), (getCallbacks) => (param1, param2, param3) => { const callbacks = getCallbacks(); if (callbacks)
            callbacks.forEach(cb => cb(param1, param2, param3)); });
    }
    function fsWatch(fileOrDirectory, entryKind, callback, recursive, fallbackPollingInterval, fallbackOptions) {
        return createSingleWatcherPerName(recursive ? fsWatchesRecursive : fsWatches, useCaseSensitiveFileNames, fileOrDirectory, callback, cb => fsWatchHandlingExistenceOnHost(fileOrDirectory, entryKind, cb, recursive, fallbackPollingInterval, fallbackOptions), (getCallbacks) => (param1, param2, param3) => { const callbacks = getCallbacks(); if (callbacks)
            callbacks.forEach(cb => cb(param1, param2, param3)); });
    }
    function fsWatchHandlingExistenceOnHost(fileOrDirectory, entryKind, callback, recursive, fallbackPollingInterval, fallbackOptions) {
        let lastDirectoryPartWithDirectorySeparator;
        let lastDirectoryPart;
        if (inodeWatching) {
            lastDirectoryPartWithDirectorySeparator = fileOrDirectory.substring(fileOrDirectory.lastIndexOf(directorySeparator));
            lastDirectoryPart = lastDirectoryPartWithDirectorySeparator.slice(directorySeparator.length);
        }
        /** Watcher for the file system entry depending on whether it is missing or present */
        let watcher = !fileSystemEntryExists(fileOrDirectory, entryKind) ?
            watchMissingFileSystemEntry() :
            watchPresentFileSystemEntry();
        return {
            close: () => {
                // Close the watcher (either existing file system entry watcher or missing file system entry watcher)
                if (watcher) {
                    watcher.close();
                    watcher = undefined;
                }
            },
        };
        function updateWatcher(createWatcher) {
            // If watcher is not closed, update it
            if (watcher) {
                sysLog(`sysLog:: ${fileOrDirectory}:: Changing watcher to ${createWatcher === watchPresentFileSystemEntry ? "Present" : "Missing"}FileSystemEntryWatcher`);
                watcher.close();
                watcher = createWatcher();
            }
        }
        /**
         * Watch the file or directory that is currently present
         * and when the watched file or directory is deleted, switch to missing file system entry watcher
         */
        function watchPresentFileSystemEntry() {
            if (hitSystemWatcherLimit) {
                sysLog(`sysLog:: ${fileOrDirectory}:: Defaulting to watchFile`);
                return watchPresentFileSystemEntryWithFsWatchFile();
            }
            try {
                const presentWatcher = (entryKind === FileSystemEntryKind.Directory || !fsWatchWithTimestamp ? fsWatchWorker : fsWatchWorkerHandlingTimestamp)(fileOrDirectory, recursive, inodeWatching ?
                    callbackChangingToMissingFileSystemEntry :
                    callback);
                // Watch the missing file or directory or error
                presentWatcher.on("error", () => {
                    callback("rename", "");
                    updateWatcher(watchMissingFileSystemEntry);
                });
                return presentWatcher;
            }
            catch (e) {
                // Catch the exception and use polling instead
                // Eg. on linux the number of watches are limited and one could easily exhaust watches and the exception ENOSPC is thrown when creating watcher at that point
                // so instead of throwing error, use fs.watchFile
                hitSystemWatcherLimit ||= propertyOf(e, "code") === "ENOSPC";
                sysLog(`sysLog:: ${fileOrDirectory}:: Changing to watchFile`);
                return watchPresentFileSystemEntryWithFsWatchFile();
            }
        }
        function callbackChangingToMissingFileSystemEntry(event, relativeName) {
            // In some scenarios, file save operation fires event with fileName.ext~ instead of fileName.ext
            // To ensure we see the file going missing and coming back up (file delete and then recreated)
            // and watches being updated correctly we are calling back with fileName.ext as well as fileName.ext~
            // The worst is we have fired event that was not needed but we wont miss any changes
            // especially in cases where file goes missing and watches wrong inode
            let originalRelativeName;
            if (relativeName && endsWith(relativeName, "~")) {
                originalRelativeName = relativeName;
                relativeName = relativeName.slice(0, relativeName.length - 1);
            }
            // because relativeName is not guaranteed to be correct we need to check on each rename with few combinations
            // Eg on ubuntu while watching app/node_modules the relativeName is "node_modules" which is neither relative nor full path
            if (event === "rename" &&
                (!relativeName ||
                    relativeName === lastDirectoryPart ||
                    endsWith(relativeName, lastDirectoryPartWithDirectorySeparator))) {
                const modifiedTime = getModifiedTime(fileOrDirectory) || missingFileModifiedTime;
                if (originalRelativeName)
                    callback(event, originalRelativeName, modifiedTime);
                callback(event, relativeName, modifiedTime);
                if (inodeWatching) {
                    // If this was rename event, inode has changed means we need to update watcher
                    updateWatcher(modifiedTime === missingFileModifiedTime ? watchMissingFileSystemEntry : watchPresentFileSystemEntry);
                }
                else if (modifiedTime === missingFileModifiedTime) {
                    updateWatcher(watchMissingFileSystemEntry);
                }
            }
            else {
                if (originalRelativeName)
                    callback(event, originalRelativeName);
                callback(event, relativeName);
            }
        }
        /**
         * Watch the file or directory using fs.watchFile since fs.watch threw exception
         * Eg. on linux the number of watches are limited and one could easily exhaust watches and the exception ENOSPC is thrown when creating watcher at that point
         */
        function watchPresentFileSystemEntryWithFsWatchFile() {
            return watchFile(fileOrDirectory, createFileWatcherCallback(callback), fallbackPollingInterval, fallbackOptions);
        }
        /**
         * Watch the file or directory that is missing
         * and switch to existing file or directory when the missing filesystem entry is created
         */
        function watchMissingFileSystemEntry() {
            return watchFile(fileOrDirectory, (_fileName, eventKind, modifiedTime) => {
                if (eventKind === FileWatcherEventKind.Created) {
                    modifiedTime ||= getModifiedTime(fileOrDirectory) || missingFileModifiedTime;
                    if (modifiedTime !== missingFileModifiedTime) {
                        callback("rename", "", modifiedTime);
                        // Call the callback for current file or directory
                        // For now it could be callback for the inner directory creation,
                        // but just return current directory, better than current no-op
                        updateWatcher(watchPresentFileSystemEntry);
                    }
                }
            }, fallbackPollingInterval, fallbackOptions);
        }
    }
    function fsWatchWorkerHandlingTimestamp(fileOrDirectory, recursive, callback) {
        let modifiedTime = getModifiedTime(fileOrDirectory) || missingFileModifiedTime;
        return fsWatchWorker(fileOrDirectory, recursive, (eventName, relativeFileName, currentModifiedTime) => {
            if (eventName === "change") {
                currentModifiedTime ||= getModifiedTime(fileOrDirectory) || missingFileModifiedTime;
                if (currentModifiedTime.getTime() === modifiedTime.getTime())
                    return;
            }
            modifiedTime = currentModifiedTime || getModifiedTime(fileOrDirectory) || missingFileModifiedTime;
            callback(eventName, relativeFileName, modifiedTime);
        });
    }
}
/**
 * patch writefile to create folder before writing the file
 *
 * @internal
 */
export function patchWriteFileEnsuringDirectory(sys) {
    // patch writefile to create folder before writing the file
    const originalWriteFile = sys.writeFile;
    sys.writeFile = (path, data, writeBom) => writeFileEnsuringDirectories(path, data, !!writeBom, (path, data, writeByteOrderMark) => originalWriteFile.call(sys, path, data, writeByteOrderMark), path => sys.createDirectory(path), path => sys.directoryExists(path));
}
// TODO: GH#18217 this is used as if it's certainly defined in many places.
export let sys = (() => {
    // NodeJS detects "\uFEFF" at the start of the string and *replaces* it with the actual
    // byte order mark from the specified encoding. Using any other byte order mark does
    // not actually work.
    const byteOrderMarkIndicator = "\uFEFF";
    function getNodeSystem() {
        const nativePattern = /^native |^\([^)]+\)$|^(?:internal[\\/]|[\w\s]+(?:\.js)?$)/;
        let activeSession;
        let profilePath = "./profile.cpuprofile";
        const isMacOs = process.platform === "darwin";
        const isLinuxOrMacOs = process.platform === "linux" || isMacOs;
        const statSyncOptions = { throwIfNoEntry: false };
        const platform = _os.platform();
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
            setTimeout: (callback, ms, label) => setTimeout(callback, ms, label),
            clearTimeout: (timeoutId) => { clearTimeout(timeoutId); },
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
        const nodeSystem = {
            args: process.argv.slice(2),
            newLine: _os.EOL,
            useCaseSensitiveFileNames,
            write(s) {
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
            createDirectory(directoryName) {
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
            getEnvironmentVariable(name) {
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
            exit(exitCode) {
                disableCPUProfiler(() => process.exit(exitCode));
            },
            enableCPUProfiler,
            disableCPUProfiler,
            cpuProfilingEnabled: () => !!activeSession || contains(process.execArgv, "--cpu-prof") || contains(process.execArgv, "--prof"),
            realpath,
            debugMode: !!process.env.NODE_INSPECTOR_IPC || !!process.env.VSCODE_INSPECTOR_OPTIONS || some(process.execArgv, arg => /^--(?:inspect|debug)(?:-brk)?(?:=\d+)?$/i.test(arg)) || !!process.recordreplay,
            tryEnableSourceMapsForHost() {
                try {
                    require("source-map-support").install();
                }
                catch {
                    // Could not enable source maps.
                }
            },
            setTimeout: (callback, ms, label) => setTimeout(callback, ms, label),
            clearTimeout: (timeoutId) => { clearTimeout(timeoutId); },
            clearScreen: () => {
                process.stdout.write("\x1B[2J\x1B[3J\x1B[H");
            },
            setBlocking: () => {
                const handle = process.stdout?._handle;
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
        function statSync(path) {
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
        function enableCPUProfiler(path, cb) {
            if (activeSession) {
                cb();
                return false;
            }
            const inspector = require("inspector");
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
        function cleanupPaths(profile) {
            let externalFileCounter = 0;
            const remappedPaths = new Map();
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
                        node.callFrame.url = (remappedPaths.has(url) ? remappedPaths : remappedPaths.set(url, `external${externalFileCounter}.js`)).get(url);
                        externalFileCounter++;
                    }
                }
            }
            return profile;
        }
        function disableCPUProfiler(cb) {
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
                        _fs.writeFileSync(profilePath, JSON.stringify(cleanupPaths(profile)));
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
        function isFileSystemCaseSensitive() {
            // win32\win64 are case insensitive platforms
            if (platform === "win32" || platform === "win64") {
                return false;
            }
            // If this file exists under a different case, we must be case-insensitve.
            return !fileExists(swapCase(__filename));
        }
        /** Convert all lowercase chars to uppercase, and vice-versa */
        function swapCase(s) {
            return s.replace(/\w/g, ch => {
                const up = ch.toUpperCase();
                return ch === up ? ch.toLowerCase() : up;
            });
        }
        function fsWatchFileWorker(fileName, callback, pollingInterval) {
            _fs.watchFile(fileName, { persistent: true, interval: pollingInterval }, fileChanged);
            let eventKind;
            return {
                close: () => _fs.unwatchFile(fileName, fileChanged),
            };
            function fileChanged(curr, prev) {
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
        function fsWatchWorker(fileOrDirectory, recursive, callback) {
            // Node 4.0 `fs.watch` function supports the "recursive" option on both OSX and Windows
            // (ref: https://github.com/nodejs/node/pull/2649 and https://github.com/Microsoft/TypeScript/issues/4643)
            return _fs.watch(fileOrDirectory, fsSupportsRecursiveFsWatch ?
                { persistent: true, recursive: !!recursive } : { persistent: true }, callback);
        }
        function readFile(fileName, _encoding) {
            let buffer;
            try {
                buffer = _fs.readFileSync(fileName);
            }
            catch {
                return undefined;
            }
            let len = buffer.length;
            if (len >= 2 && buffer[0] === 0xFE && buffer[1] === 0xFF) {
                // Big endian UTF-16 byte order mark detected. Since big endian is not supported by node.js,
                // flip all byte pairs and treat as little endian.
                len &= ~1; // Round down to a multiple of 2
                for (let i = 0; i < len; i += 2) {
                    const temp = buffer[i];
                    buffer[i] = buffer[i + 1];
                    buffer[i + 1] = temp;
                }
                return buffer.toString("utf16le", 2);
            }
            if (len >= 2 && buffer[0] === 0xFF && buffer[1] === 0xFE) {
                // Little endian UTF-16 byte order mark detected
                return buffer.toString("utf16le", 2);
            }
            if (len >= 3 && buffer[0] === 0xEF && buffer[1] === 0xBB && buffer[2] === 0xBF) {
                // UTF-8 byte order mark detected
                return buffer.toString("utf8", 3);
            }
            // Default is UTF-8 with no byte order mark
            return buffer.toString("utf8");
        }
        function writeFile(fileName, data, writeByteOrderMark) {
            // If a BOM is required, emit one
            if (writeByteOrderMark) {
                data = byteOrderMarkIndicator + data;
            }
            let fd;
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
        function getAccessibleFileSystemEntries(path) {
            try {
                const entries = _fs.readdirSync(path || ".", { withFileTypes: true });
                const files = [];
                const directories = [];
                for (const dirent of entries) {
                    // withFileTypes is not supported before Node 10.10.
                    const entry = typeof dirent === "string" ? dirent : dirent.name;
                    // This is necessary because on some file system node fails to exclude
                    // "." and "..". See https://github.com/nodejs/node/issues/4002
                    if (entry === "." || entry === "..") {
                        continue;
                    }
                    let stat;
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
        function readDirectory(path, extensions, excludes, includes, depth) {
            return matchFiles(path, extensions, excludes, includes, useCaseSensitiveFileNames, process.cwd(), depth, getAccessibleFileSystemEntries, realpath);
        }
        function fileSystemEntryExists(path, entryKind) {
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
        function fileExists(path) {
            return fileSystemEntryExists(path, FileSystemEntryKind.File);
        }
        function directoryExists(path) {
            return fileSystemEntryExists(path, FileSystemEntryKind.Directory);
        }
        function getDirectories(path) {
            return getAccessibleFileSystemEntries(path).directories.slice();
        }
        function fsRealPathHandlingLongPath(path) {
            return path.length < 260 ? _fs.realpathSync.native(path) : _fs.realpathSync(path);
        }
        function realpath(path) {
            try {
                return fsRealpath(path);
            }
            catch {
                return path;
            }
        }
        function getModifiedTime(path) {
            return statSync(path)?.mtime;
        }
        function setModifiedTime(path, time) {
            try {
                _fs.utimesSync(path, time, time);
            }
            catch {
                return;
            }
        }
        function deleteFile(path) {
            try {
                return _fs.unlinkSync(path);
            }
            catch {
                return;
            }
        }
        function createSHA256Hash(data) {
            const hash = _crypto.createHash("sha256");
            hash.update(data);
            return hash.digest("hex");
        }
    }
    let sys;
    if (isNodeLikeSystem()) {
        sys = getNodeSystem();
    }
    if (sys) {
        // patch writefile to create folder before writing the file
        patchWriteFileEnsuringDirectory(sys);
    }
    return sys;
})();
/** @internal @knipignore */
export function setSys(s) {
    sys = s;
}
if (sys && sys.getEnvironmentVariable) {
    setCustomPollingValues(sys);
    Debug.setAssertionLevel(/^development$/i.test(sys.getEnvironmentVariable("NODE_ENV"))
        ? AssertionLevel.Normal
        : AssertionLevel.None);
}
if (sys && sys.debugMode) {
    Debug.isDebugging = true;
}
