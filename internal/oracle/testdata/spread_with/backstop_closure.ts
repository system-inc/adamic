// A function with an own tag through { readonly tag: number }: Node copies its own enumerable keys, tag, which a
// SpreadWith literal doesn't build yet, so both backends stop there, after Node's output so far (Adamic's
// run-time NotYet on #k7g6f9h). The view names tag, which no library type in the closed world has, so
// checkLibrarySpreads passes it (a length view would fit console.log's function type).
function sum(first: number, second: number): number {
    return first + second;
}
sum.tag = 2;
function copy(source: { readonly tag: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ tag: 1 });
copy(sum);
console.log(String(sum(1, 2)));
