// By width a { readonly a: number } may hold more than a: Node's spread copies extra, and k is appended after
// it, so keys, JSON, for-in and a later spread each hold extra (#k7g6f9h).
function overwrite(source: { readonly a: number }): { a: number; k: number } {
    return { ...source, k: 5 };
}
const wide = { a: 1, extra: 2 };
const copy = overwrite(wide);
console.log(Object.keys(copy).join(","));
console.log(String(JSON.stringify(copy)));
const seen: string[] = [];
for (const key in copy) {
    seen.push(key);
}
console.log(seen.join(","));
console.log(`${copy.a} ${copy.k}`);
const again = { ...copy };
console.log(String(JSON.stringify(again)));
