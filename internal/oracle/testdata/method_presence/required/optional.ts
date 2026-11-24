interface Host { read(): string; }
let reads = 0;
function make(ready: boolean): Host | undefined {
    reads++;
    if (ready) return { read() { return "present"; } };
    return undefined;
}
function probe(ready: boolean): string {
    return !make(ready)?.read ? "absent" : "present";
}
console.log(probe(false));
console.log(probe(true));
console.log(String(reads));
