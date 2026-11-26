// A closure through { readonly length: number }: Node copies its own enumerable keys, none, which a SpreadWith
// literal doesn't build yet, so both backends stop there, after Node's output so far (Adamic's run-time NotYet
// on #k7g6f9h).
function copy(source: { readonly length: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ length: 2 });
copy((first: number, second: number): number => first + second);
console.log("end");
