// By width a { readonly a: number } may hold more than a: Node's spread copies extra, and k is appended after
// it, so keys, JSON and a later spread each hold extra (#k7g6f9h). for-in isn't read here: lowering stops it on
// a value without a proven fixed plain-object origin, which a SpreadWith result isn't.
function overwrite(source: { readonly a: number }): { a: number; k: number } {
    return { ...source, k: 5 };
}
const wide = { a: 1, extra: 2 };
const copy = overwrite(wide);
console.log(Object.keys(copy).join(","));
console.log(String(JSON.stringify(copy)));
console.log(`${copy.a} ${copy.k}`);
const again = { ...copy };
console.log(String(JSON.stringify(again)));
