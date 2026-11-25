interface WatchHost {
    getAccessibleSortedChildDirectories(path: string): readonly string[];
    realpath: (path: string) => string;
}
function recursive({ getAccessibleSortedChildDirectories, realpath }: WatchHost): string {
    return realpath(getAccessibleSortedChildDirectories("root")[0]!);
}
function watch({ getAccessibleSortedChildDirectories, realpath }: WatchHost): string {
    return recursive({ getAccessibleSortedChildDirectories, realpath });
}
function nodeSystem(): string {
    function realpath(path: string): string {
        return "/" + path;
    }
    return watch({ getAccessibleSortedChildDirectories: path => [path], realpath: path => realpath(path) });
}
console.log(nodeSystem());
