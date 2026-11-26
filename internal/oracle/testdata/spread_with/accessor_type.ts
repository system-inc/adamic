// The spread's own type is a class whose getter is on its prototype, which Node's spread never copies, so the
// SpreadWith literal copies as Node does (#k7g6f9h, #vjvcnnj's narrowing).
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
