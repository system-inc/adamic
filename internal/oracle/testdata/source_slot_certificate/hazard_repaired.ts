// The hazard: the view writes { ready } into the slot raw reads as { ready, other }, so raw.child.other is a checked read that stops where Node prints undefined. A ruled negative witness (#2msw6t8, rulings.md:17): hazard_repaired.ts writes through raw, with no downcast, and agrees with Node.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw: { kind: "box"; child: { ready: boolean; other: number } } = { kind: "box", child: { ready: true, other: 42 } };
const held: Base = raw;
const view = raw;
view.child = { ready: false, other: 7 };
console.log(String(view.child.ready));
console.log(String(raw.child.other));
