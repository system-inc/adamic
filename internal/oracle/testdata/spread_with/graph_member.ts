// A graph program whose overwrite value is a graph member: the result is adopted, then stored back into what the
// value reaches, which closes a cycle (Memory's pin on #k7g6f9h).
interface Item {
    label: string;
    peer: Item | undefined;
}
function attach(source: { readonly label: string }, peer: Item): Item {
    return { ...source, peer };
}
const first: Item = { label: "a", peer: undefined };
const second = attach({ label: "b" }, first);
first.peer = second;
console.log(`${second.label} ${second.peer?.label} ${first.peer?.label} ${first.peer?.peer?.label}`);
