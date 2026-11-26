// Node reads the spread's own keys before the field's value runs, so a value that writes the source doesn't show
// in the result (#k7g6f9h).
const box: { a: number } = { a: 1 };
function bump(): number {
    box.a = 5;
    return 2;
}
function make(source: { readonly a: number }): { a: number; k: number } {
    return { ...source, k: bump() };
}
const made = make(box);
console.log(`${made.a} ${made.k} ${box.a}`);
