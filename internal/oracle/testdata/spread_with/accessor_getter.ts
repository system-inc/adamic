// A class with a getter may reach a SpreadWith literal's source, so it stops (#k7g6f9h, #vjvcnnj).
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
console.log(new Sized("u").weight);
