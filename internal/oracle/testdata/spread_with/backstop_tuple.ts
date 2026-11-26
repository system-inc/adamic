// A tuple through { readonly length: number }: Node copies its indices, which a SpreadWith literal doesn't build
// yet, so both backends stop there, after Node's output so far (Adamic's run-time NotYet on #k7g6f9h).
function copy(source: { readonly length: number }): void {
    const made = { ...source, k: 1 };
    console.log(JSON.stringify(made));
}
copy({ length: 2 });
const pair: [number, number] = [1, 2];
copy(pair);
console.log("end");
