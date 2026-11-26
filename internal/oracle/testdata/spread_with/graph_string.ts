// A graph program whose overwrite value is a fresh string, no graph member: adamic_graph_hold would retain it, so a
// hold after the adoption on top of the primitive's count leaves it one count too many (Memory's W7 row on
// #k7g6f9h). The result is adopted, then closed into a cycle.
interface Item {
    label: string;
    peer: Item | undefined;
}
function relabel(holder: { readonly peer: Item | undefined }, times: number): Item {
    return { ...holder, label: "x".repeat(times) };
}
const first: Item = { label: "a", peer: undefined };
const second = relabel({ peer: first }, 2);
first.peer = second;
console.log(`${second.label} ${second.peer?.label} ${first.peer?.label}`);
