// A plain object through { readonly length: number } copies as Node's spread does (Adamic's condition 2 on
// #k7g6f9h, the control beside backstop_tuple.ts).
function copy(source: { readonly length: number }): void {
    const made = { ...source, k: 1 };
    console.log(JSON.stringify(made));
}
copy({ length: 2 });
console.log("end");
