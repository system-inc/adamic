// An assignment pattern reads the slot at the binding's type. A ruled negative witness (#80cs7rz, rulings.md:19, checked where it's made by #zpzm9ta): road_destructure_assignment_repaired.ts writes through raw, with no downcast, and agrees with Node.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw: { kind: "box"; child: { ready: boolean; other: number } } = { kind: "box", child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
let child = { ready: false, other: 0 };
({ child } = raw);
console.log(String(child.other));
