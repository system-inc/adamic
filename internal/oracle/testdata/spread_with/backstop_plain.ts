// A plain object through { readonly tag: number } copies as Node's spread does (Adamic's condition on #k7g6f9h, the
// control beside backstop_tuple.ts, through a view no library type in the closed world fits).
function copy(source: { readonly tag: number }): void {
    const made = { ...source, k: 1 };
    console.log(String(JSON.stringify(made)));
}
copy({ tag: 2 });
console.log("end");
