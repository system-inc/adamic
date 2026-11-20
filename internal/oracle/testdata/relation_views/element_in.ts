// rulings.md:72's second slice (#0b9xqhh), Lower's condition 2: an element of a relation-viewed array, the second
// hiding a number as its label, asked by the in operator for the label's presence, which reads no value: Node agrees.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

const honest = { kind: 1, label: "one" };
const lying = { kind: 2, label: 5 };
const first: Source = honest;
const second: Source = lying;
const sources: Source[] = [first, second];
const viewed: readonly Target[] = sources;
for (const target of viewed) {
    console.log("label" in target ? "has" : "lacks");
}
