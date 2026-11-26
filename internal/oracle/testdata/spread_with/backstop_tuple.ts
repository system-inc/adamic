// A tuple through { readonly 0: number }: Node copies its indices, which a SpreadWith literal doesn't build yet, so
// both backends stop there, after Node's output so far (Adamic's run-time NotYet on #k7g6f9h). The view names 0,
// which no library type in the closed world has, so checkLibrarySpreads passes it (a length view would fit
// console.log's function type).
function copy(source: { readonly 0: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ 0: 5 });
const pair: [number, number] = [1, 2];
copy(pair);
console.log("end");
