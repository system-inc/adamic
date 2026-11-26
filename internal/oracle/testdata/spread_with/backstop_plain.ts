// A plain object through { readonly 0: number } copies as Node's spread does (Adamic's condition on #k7g6f9h, the
// control beside backstop_tuple.ts, through the same view).
function copy(source: { readonly 0: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ 0: 2 });
console.log("end");
