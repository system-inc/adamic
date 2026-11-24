interface Host { read(): string; }
function make(ready: boolean): Host | null {
    if (ready) return { read() { return "present"; } };
    return null;
}
function probe(ready: boolean): string {
    return !make(ready)?.read ? "absent" : "present";
}
console.log(probe(false));
console.log(probe(true));
