// A closure reads raw's slot after the view's write, at the type raw declares. A ruled negative witness (#2msw6t8, rulings.md:17): road_closure_later_repaired.ts writes through raw, with no downcast, and agrees with Node.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw: { kind: "box"; child: { ready: boolean; other: number } } = { kind: "box", child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: true };
const read = (): string => String(raw.child.other); console.log(read());
