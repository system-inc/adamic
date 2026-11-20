// rulings.md:72's second slice (#0b9xqhh), containers: JSON.stringify of an element of a viewed array reads
// every field of it through no view, so it stops at compile time (Lower's condition 2): checkViewArrays stops
// a whole read of what may reach a viewed array.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const sources: Source[] = [hidden];
const viewed: readonly Target[] = sources;
const first = viewed[0];
if (first !== undefined) {
    console.log(JSON.stringify(first) ?? "undefined");
}
