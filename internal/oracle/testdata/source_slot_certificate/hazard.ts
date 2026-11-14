// The hazard: the view writes { ready } into the slot raw reads as { ready, other }, so raw.child.other is a checked read that stops where Node prints undefined.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
console.log(String(raw.child.other));
