// A string over a number slot and a number over a string slot: each slot takes the kind of its value, and the
// slot the overwrite replaces is never copied (#k7g6f9h). for-in isn't read: lowering stops it on a value without a
// proven fixed plain-object origin, which a SpreadWith result isn't.
function toText(source: { readonly k: number; readonly label: string }): { k: string; label: string } {
    return { ...source, k: "ten" };
}
function toNumber(source: { readonly k: string; readonly n: number }): { k: number; n: number } {
    return { ...source, k: 5 };
}
const text = toText({ k: 10, label: "one" });
console.log(Object.keys(text).join(","));
console.log(String(JSON.stringify(text)));
console.log(`${typeof text.k} ${text.k} ${text.label}`);
const made = toNumber({ k: "x" + String(1), n: 2 });
console.log(String(JSON.stringify(made)));
console.log(`${typeof made.k} ${made.k + made.n}`);
