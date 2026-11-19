// rulings.md:72's second slice (#0b9xqhh), containers: an initializer views a Source[] as a readonly Target[],
// and an element read by index reads its absent label through Target, checked: undefined, as on Node. The
// array's length reads no element.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const bare: Source = { kind: 1 };
const sources: Source[] = [bare];
const viewed: readonly Target[] = sources;
const first = viewed[0];
if (first !== undefined) {
    console.log(String(first.kind) + " " + (first.label === undefined ? "absent" : first.label));
}
console.log(String(viewed.length));
