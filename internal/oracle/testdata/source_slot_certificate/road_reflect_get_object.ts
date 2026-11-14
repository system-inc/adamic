// Reflect.get in the one form lowering takes reads the slot as unknown.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
const read: unknown = Reflect.get(Object(raw), "child", raw);
console.log(typeof read);
