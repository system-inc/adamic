import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
interface Other { readonly size: number }
type First = { readonly kind: "v"; readonly v: Weak<Box> } | { readonly kind: "w"; readonly w: Weak<Other> };
type Second = { readonly kind: "w"; readonly w: Weak<Other> } | { readonly kind: "v"; readonly v: Weak<Box> };
const held = box();
const first: First = { kind: "v", v: held };
const second: Second = first;
console.log(second.kind === "v" ? second.v?.label ?? "none" : "w");
