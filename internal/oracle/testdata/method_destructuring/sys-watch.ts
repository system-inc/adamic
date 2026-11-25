interface WatchHost {
    getAccessibleSortedChildDirectories(path: string): readonly string[];
    realpath(path: string): string;
}
function watch({ getAccessibleSortedChildDirectories, realpath }: WatchHost): string {
    return realpath(getAccessibleSortedChildDirectories("root")[0]!);
}
function nodeSystem(): string {
    function realpath(path: string): string {
        return "/" + path;
    }
    return watch({ getAccessibleSortedChildDirectories: path => [path], realpath });
}
console.log(nodeSystem());
