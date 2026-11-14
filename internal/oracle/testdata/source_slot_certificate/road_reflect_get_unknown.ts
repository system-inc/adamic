// Reflect.get in the one form lowering takes, of the object held as unknown, reads the slot as unknown, a value no type claims.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
const box: unknown = raw;
const read: unknown = Reflect.get(Object(box), "child", box);
console.log(typeof read);
