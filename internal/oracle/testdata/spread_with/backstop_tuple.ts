// A tuple held in a union with a plain object type, spread with an overwrite: Node copies its indices, which a
// SpreadWith literal doesn't build yet, so both backends stop there, after Node's output so far (Adamic's run-time
// NotYet on #k7g6f9h). Seen alone as a plain object type, a tuple stops at compile time where it's passed, and a
// { readonly length: number } view would fit console.log's function type and be refused; in a union the tuple
// reaches the spread (J-9pj0a8k). The copy is printed by its keys, since JSON.stringify of the union stops first.
type Either = readonly [number, number] | { readonly a: number };
function copy(source: Either): void {
    const made = { ...source, k: 1 };
    console.log(Object.keys(made).join(",") + " " + String(made.k));
}
copy({ a: 2 });
copy([1, 2]);
console.log("end");
