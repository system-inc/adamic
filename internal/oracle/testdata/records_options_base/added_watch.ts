// #6c4swvt's record-to-fixed direction, tsc's commandLineParser.ts:3100: a MapLike, or undefined, seen as
// WatchOptions in a conditional's branch. Its enum and array members read as Node reads them.
interface MapLike<T> {
    [index: string]: T;
}
type Value = string | number | boolean | string[] | undefined;
const enum WatchFileKind {
    FixedPollingInterval = 0,
    UseFsEvents = 4,
}
interface WatchOptions {
    watchFile?: WatchFileKind;
    excludeFiles?: string[];
    [option: string]: Value;
}
// extend's own for...in reads a record's values the allocation graph can't place, so it stands in as the second's.
function extend(first: MapLike<Value>, second: MapLike<Value>): MapLike<Value> {
    second["watchFile"] = first["watchFile"];
    return second;
}
function merged(existing: MapLike<Value> | undefined, parsed: MapLike<Value> | undefined): WatchOptions | undefined {
    return existing && parsed ? extend(existing, parsed) : existing || parsed;
}
const watch = merged({ watchFile: 4 }, { excludeFiles: ["x.ts"] });
if (watch) {
    const files = watch.excludeFiles;
    const shown = files ? String(files.length) + ":" + (files[0] ?? "none") : "none";
    console.log(String(watch.watchFile === WatchFileKind.UseFsEvents) + ":" + shown);
}
