// An object literal's own getter makes it a class with an accessor, which may reach a SpreadWith literal's
// source, so it stops (#k7g6f9h, #vjvcnnj).
const sized = {
    label: "l",
    get size(): number {
        return 3;
    },
};
function copy(source: { readonly label: string }): void {
    const made = { ...source, k: 1 };
    console.log(`${made.label} ${made.k}`);
}
copy(sized);
console.log(sized.size);
