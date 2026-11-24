interface Host { read(): string; }
interface Holder { host: Host; }
let reads = 0;
function make(ready: boolean): Holder | undefined {
    reads++;
    if (ready) return { host: { read() { return "present"; } } };
    return undefined;
}
function probe(ready: boolean): string {
    const holder = make(ready);
    if (holder?.host.read) return "present";
    return "absent";
}
console.log(probe(false));
console.log(probe(true));
console.log(String(reads));
