// The spread's own type declares an accessor, so a SpreadWith literal stops before the registry is asked
// (#k7g6f9h, #vjvcnnj).
class Sized {
    constructor(readonly label: string) {}
    get weight(): number {
        return 20;
    }
}
function copy(source: Sized): void {
    const made = { ...source, k: 1 };
    console.log(`${made.label} ${made.k}`);
}
copy(new Sized("t"));
