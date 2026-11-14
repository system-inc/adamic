// Object.entries yields the slot's value at its element type.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
for (const [key, value] of Object.entries(raw)) {
    console.log(key + " " + typeof value);
}
