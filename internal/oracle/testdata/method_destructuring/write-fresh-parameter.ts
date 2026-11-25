interface WatchHost { directory(path: string): string }
function recursive({ directory }: WatchHost): string {
    return directory("root");
}
function watch({ directory }: WatchHost): string {
    return recursive({ directory });
}
const counts: number[] = [0];
const i: number = 0;
counts[i] = 1;
console.log(watch({ directory: path => "/" + path }) + String(counts[0]!));
