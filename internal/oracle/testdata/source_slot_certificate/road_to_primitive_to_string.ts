// ToPrimitive reads an object's own toString by name once every read is judged, so a view write to it keeps its stop.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; toString: { ready: boolean } }
const raw = { kind: "box" as const, toString: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.toString = { ready: false };
console.log(String(view.toString.ready));
console.log(String(raw.toString.other));
