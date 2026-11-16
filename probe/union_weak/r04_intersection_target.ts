import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Tagged = Box & { readonly tag: number };
type Strong = { readonly kind: "v"; readonly v: Tagged } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: Weak<Tagged> } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: { label: "label", tag: 1 } };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? viewed.v?.label ?? "none" : "w");
