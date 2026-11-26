// An object literal's own getter is an own enumerable key, which Node's spread runs and copies and
// adamic_object_spread_with passes over, so a SpreadWith literal it may reach stops (#k7g6f9h, #vjvcnnj).
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
