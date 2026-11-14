// A class owning get child() puts every .child read through finishAccessors' dispatcher, whose fallback is a plain load.
interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
view.child = { ready: false };
console.log(String(view.child.ready));
class Getter implements Base {
    readonly kind = "other" as const;
    get child(): { ready: boolean; other: number } {
        return { ready: false, other: 1 };
    }
}
const items: { readonly child: { ready: boolean; other: number } }[] = [raw, new Getter()];
for (const item of items) {
    console.log(String(item.child.other));
}
