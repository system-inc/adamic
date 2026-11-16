import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
class Holder {
    readonly v: Box;
    constructor(v: Box) {
        this.v = v;
    }
}
interface HolderWeak { readonly v: Weak<Box> }
type Strong = { readonly kind: "v"; readonly h: Holder } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly h: HolderWeak } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", h: new Holder(box()) };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? viewed.h.v?.label ?? "none" : "w");
