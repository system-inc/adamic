// A graph program whose overwrite value is a fresh graph literal made in the literal's own field, then closed
// into a cycle through the result (Memory's pin on #k7g6f9h).
interface Item {
    label: string;
    peer: Item | undefined;
}
function attach(source: { readonly label: string }): Item {
    return { ...source, peer: { label: "fresh", peer: undefined } };
}
const made = attach({ label: "b" });
if (made.peer) {
    made.peer.peer = made;
}
console.log(`${made.label} ${made.peer?.label} ${made.peer?.peer?.label}`);
