// An array element of raw reads the slot at the element's type, the type raw declares. A ruled negative witness (#2msw6t8, rulings.md:17): road_array_element_repaired.ts writes through raw, with no downcast, and agrees with Node.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw: { kind: "box"; child: { ready: boolean; other: number } } = { kind: "box", child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: true };
const list = [raw]; const first = list[0]; if (first !== undefined) console.log(String(first.child.other));
