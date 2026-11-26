// A class with only a setter, on its prototype, reaches a SpreadWith literal's source and copies as Node's spread
// does (#k7g6f9h, #vjvcnnj's narrowing).
class Gate {
    label = "g";
    set open(value: boolean) {
        console.log(String(value));
    }
}
function copy(source: { readonly label: string }): void {
    const made = { ...source, k: 1 };
    console.log(`${made.label} ${made.k}`);
}
const gate = new Gate();
gate.open = true;
copy(gate);
