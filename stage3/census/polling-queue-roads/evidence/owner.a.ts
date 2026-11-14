function createDynamicPriorityPollingWatchFile(host: {
    getModifiedTime: NonNullable<System["getModifiedTime"]>;
    setTimeout: NonNullable<System["setTimeout"]>;
}): HostWatchFile {
    interface PollingIntervalQueue {
        files: WatchedFileWithUnchangedPolls[];
        pollingInterval: PollingInterval;
        pollIndex: number;
        pollScheduled: unknown;
    }

    const watchedFiles: WatchedFileWithUnchangedPolls[] = [];
    const changedFilesInLastPoll: WatchedFileWithUnchangedPolls[] = [];
    const lowPollingIntervalQueue = createPollingIntervalQueue(PollingInterval.Low);
    const mediumPollingIntervalQueue = createPollingIntervalQueue(PollingInterval.Medium);
    const highPollingIntervalQueue = createPollingIntervalQueue(PollingInterval.High);
    return watchFile;

    function watchFile(fileName: string, callback: FileWatcherCallback, defaultPollingInterval: PollingInterval): FileWatcher {
        const file: WatchedFileWithUnchangedPolls = {
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

    function createPollingIntervalQueue(pollingInterval: PollingInterval): PollingIntervalQueue {
        return { files: [] as WatchedFileWithUnchangedPolls[], pollingInterval, pollIndex: 0, pollScheduled: false };
    }

    function pollPollingIntervalQueue(_timeoutType: string, queue: PollingIntervalQueue) {
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

    function pollLowPollingIntervalQueue(_timeoutType: string, queue: PollingIntervalQueue) {
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

    function pollQueue(queue: (WatchedFileWithUnchangedPolls | undefined)[], pollingInterval: PollingInterval, pollIndex: number, chunkSize: number) {
        return pollWatchedFileQueue(
            host,
            queue,
            pollIndex,
            chunkSize,
            onWatchFileStat,
        );

        function onWatchFileStat(watchedFile: WatchedFileWithUnchangedPolls, pollIndex: number, fileChanged: boolean) {
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

    function pollingIntervalQueue(pollingInterval: PollingInterval) {
        switch (pollingInterval) {
            case PollingInterval.Low:
                return lowPollingIntervalQueue;
            case PollingInterval.Medium:
                return mediumPollingIntervalQueue;
            case PollingInterval.High:
                return highPollingIntervalQueue;
        }
    }

    function addToPollingIntervalQueue(file: WatchedFileWithUnchangedPolls, pollingInterval: PollingInterval) {
        pollingIntervalQueue(pollingInterval).files.push(file);
        scheduleNextPollIfNotAlreadyScheduled(pollingInterval);
    }

    function addChangedFileToLowPollingIntervalQueue(file: WatchedFileWithUnchangedPolls) {
        changedFilesInLastPoll.push(file);
        scheduleNextPollIfNotAlreadyScheduled(PollingInterval.Low);
    }

    function scheduleNextPollIfNotAlreadyScheduled(pollingInterval: PollingInterval) {
        if (!pollingIntervalQueue(pollingInterval).pollScheduled) {
            scheduleNextPoll(pollingInterval);
        }
    }

    function scheduleNextPoll(pollingInterval: PollingInterval) {
        pollingIntervalQueue(pollingInterval).pollScheduled = (() => {
            const timerHost = host;
            const { setTimeout } = timerHost;
            const callback = pollingInterval === PollingInterval.Low ? pollLowPollingIntervalQueue : pollPollingIntervalQueue;
            const timeoutDelay = pollingInterval;
            const t1 = pollingInterval === PollingInterval.Low ? "pollLowPollingIntervalQueue" : "pollPollingIntervalQueue";
            const t2 = pollingIntervalQueue(pollingInterval);
            return setTimeout.call(timerHost, () => callback(t1, t2), timeoutDelay, t1);
        })();
    }
}
