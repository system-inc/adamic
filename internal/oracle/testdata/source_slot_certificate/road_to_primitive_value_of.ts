// ToPrimitive reads an object's own valueOf by name once every read is judged, so a view write to it keeps its stop.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; valueOf: { ready: boolean } }
const raw = { kind: "box" as const, valueOf: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.valueOf = { ready: false };
console.log(String(view.valueOf.ready));
console.log(String(raw.valueOf.other));
