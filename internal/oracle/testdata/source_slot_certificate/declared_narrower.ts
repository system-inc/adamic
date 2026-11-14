// The declared part of W: the view writes a value raw's type holds, so the writes alone would leave
// raw.child unchecked, but Box declares child narrower than raw reads it, so the read is checked.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
const fuller = { ready: false, other: 7 };
view.child = fuller;
console.log(String(raw.child.other));
