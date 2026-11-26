// A string over a number slot and a number over a string slot: each slot takes the kind of its value, and the
// slot the overwrite replaces is never copied (#k7g6f9h).
function toText(source: { readonly k: number; readonly label: string }): { k: string; label: string } {
    return { ...source, k: "ten" };
}
function toNumber(source: { readonly k: string; readonly n: number }): { k: number; n: number } {
    return { ...source, k: 5 };
}
const text = toText({ k: 10, label: "one" });
console.log(Object.keys(text).join(","));
console.log(JSON.stringify(text));
console.log(`${typeof text.k} ${text.k} ${text.label}`);
const made = toNumber({ k: "x" + String(1), n: 2 });
const names: string[] = [];
for (const key in made) {
    names.push(key);
}
console.log(names.join(","));
console.log(JSON.stringify(made));
console.log(`${typeof made.k} ${made.k + made.n}`);
