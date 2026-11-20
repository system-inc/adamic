// rulings.md:72's second slice (#0b9xqhh), containers: an element of a viewed array spread into a copy. The copy
// would move the element's hidden label as the type Target claims, unchecked, so it stops at compile time
// (Lower's condition 2): checkViewCopies stops a whole-object copy of a checked view's type.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const labelled = { kind: 2, label: "two" };
const hidden: Source = labelled;
const sources: Source[] = [hidden];
const viewed: readonly Target[] = sources;
const first = viewed[0];
if (first !== undefined) {
    const copy = { ...first };
    console.log(String(copy.kind));
}
