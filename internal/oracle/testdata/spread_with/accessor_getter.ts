// A class whose getter is on its prototype reaches a SpreadWith literal's source: Node's spread copies own
// enumerable keys only, so the getter is never copied, natively either (#k7g6f9h, #vjvcnnj's narrowing).
class Sized {
    constructor(readonly label: string) {}
    get weight(): number {
        return 20;
    }
}
function copy(source: { readonly label: string }): void {
    const made = { ...source, k: 1 };
    console.log(`${made.label} ${made.k}`);
}
copy(new Sized("t"));
console.log(String(new Sized("u").weight));
