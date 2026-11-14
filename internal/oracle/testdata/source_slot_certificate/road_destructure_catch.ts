// A destructuring in a catch reads the slot at the binding's type.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
try {
    throw raw;
} catch (error) {
    const { child } = error as typeof raw;
    console.log(String(child.other));
}
