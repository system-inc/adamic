// A declaration's destructuring reads the slot at the binding's type. A ruled negative witness (#2msw6t8, rulings.md:17): road_destructure_declaration_repaired.ts writes through raw, with no downcast, and agrees with Node.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw: { kind: "box"; child: { ready: boolean; other: number } } = { kind: "box", child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
const { child } = raw;
console.log(String(child.other));
