interface WatchedFile {
	readonly fileName: string;
}
interface WatchedFileWithIsClosed extends WatchedFile {
	isClosed?: boolean;
}
interface WatchedFileWithUnchangedPolls extends WatchedFileWithIsClosed {
	unchangedPolls: number;
}
function pollWatchedFileQueue<T extends WatchedFileWithIsClosed>(
	host: { getModifiedTime: (fileName: string) => number },
	queue: (T | undefined)[],
	pollIndex: number,
	chunkSize: number,
	callbackOnWatchFileStat?: (watchedFile: T, pollIndex: number, fileChanged: boolean) => void,
): number {
	let definedValueCopyToIndex = pollIndex;
	for (let canVisit = queue.length; chunkSize > 0 && canVisit > 0; canVisit--) {
		const watchedFile = queue[pollIndex];
		if (watchedFile === undefined) {
			pollIndex++;
			continue;
		}
		if (watchedFile.isClosed === true) {
			queue[pollIndex] = undefined;
			pollIndex++;
			continue;
		}
		chunkSize--;
		const fileChanged = host.getModifiedTime(watchedFile.fileName) > 0;
		if (callbackOnWatchFileStat !== undefined) {
			callbackOnWatchFileStat(watchedFile, pollIndex, fileChanged);
		}
		if (queue[pollIndex] !== undefined) {
			if (definedValueCopyToIndex < pollIndex) {
				queue[definedValueCopyToIndex] = watchedFile;
				queue[pollIndex] = undefined;
			}
			definedValueCopyToIndex++;
		}
		pollIndex++;
	}
	return pollIndex;
}
const host = {
	getModifiedTime(fileName: string): number {
		return fileName.length - 1;
	},
};
function pollQueue(queue: (WatchedFileWithUnchangedPolls | undefined)[], pollIndex: number, chunkSize: number): number {
	return pollWatchedFileQueue(host, queue, pollIndex, chunkSize, onWatchFileStat);

	function onWatchFileStat(watchedFile: WatchedFileWithUnchangedPolls, pollIndex: number, fileChanged: boolean): void {
		if (fileChanged) {
			watchedFile.unchangedPolls = 0;
		} else {
			watchedFile.unchangedPolls++;
			queue[pollIndex] = undefined;
		}
	}
}
const queue: (WatchedFileWithUnchangedPolls | undefined)[] = [
	undefined,
	{ fileName: "ab", unchangedPolls: 3 },
	{ fileName: "c", unchangedPolls: 1 },
	{ fileName: "de", isClosed: true, unchangedPolls: 0 },
	{ fileName: "fg", unchangedPolls: 5 },
];
console.log(String(pollQueue(queue, 0, 3)));
for (const file of queue) {
	console.log(file === undefined ? "hole" : file.fileName + " " + String(file.unchangedPolls));
}
