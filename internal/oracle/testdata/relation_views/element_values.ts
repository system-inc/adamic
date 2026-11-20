// rulings.md:72's second slice (#0b9xqhh), Lower's condition 2: an element of a relation-viewed array, whose
// hidden label is a number in the second, read whole by Object.values. It must stop, never read the element raw.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
const sources: Source[] = [first, second];
const viewed: readonly Target[] = sources;
for (const target of viewed) {
    console.log(String(Object.values(target).length));
}
