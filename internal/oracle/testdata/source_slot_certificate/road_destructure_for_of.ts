// A for-of's destructuring reads the slot at the binding's type, as a declaration's does (#rm22wzv): road_destructure_for_of_repaired.ts writes through raw, with no downcast, and agrees with Node.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
for (const { child } of [raw]) {
    console.log(String(child.other));
}
