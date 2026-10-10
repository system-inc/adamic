class Holder { readonly length: number = 1; }
let value: unknown = new Holder();
function change(): void { value = `bytes${12}`; }
if (value instanceof Holder) {
    change();
    console.log(`${value.length}`);
}
