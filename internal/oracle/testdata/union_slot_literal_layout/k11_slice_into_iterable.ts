const kinds: number[] = [0, 1];
const it: Iterable<number | string> = kinds.slice();
for (const x of it) { console.log(String(x)); }
