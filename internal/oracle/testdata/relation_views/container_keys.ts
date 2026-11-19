// rulings.md:72's second slice (#0b9xqhh), containers: Object.keys of an element of a viewed array reads the
// element whole, through no type, so it stops at compile time (Lower's condition 2). Nothing stopped it before
// this slice: checkViewArrays now holds every Object.keys, values and entries operand that may be an element
// of a relation view's array.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const sources: Source[] = [hidden];
const viewed: readonly Target[] = sources;
const first = viewed[0];
if (first !== undefined) {
    console.log(String(Object.keys(first).length));
}
