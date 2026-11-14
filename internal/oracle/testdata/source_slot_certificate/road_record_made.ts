// An object made through Record<"child", Child> reads the slot at Child, which line 2 declares. A ruled negative witness (#2msw6t8, rulings.md:17): road_record_made_repaired.ts writes through made, with no downcast, and agrees with Node.
type Child = { ready: boolean; other: number };
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const made: Record<"child", Child> & { kind: "box" } = { kind: "box", child: { ready: true, other: 42 } };
const view = (made as Base) as Box;
view.child = { ready: true };
console.log(String(made.child.other));
