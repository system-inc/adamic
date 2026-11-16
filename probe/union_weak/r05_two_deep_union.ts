import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Inner = { readonly kind: "a"; readonly b: Box } | { readonly kind: "c"; readonly c: number };
type InnerWeak = { readonly kind: "a"; readonly b: Weak<Box> } | { readonly kind: "c"; readonly c: number };
type Strong = { readonly kind: "v"; readonly v: Inner } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: InnerWeak } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: { kind: "a", b: box() } };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" && viewed.v.kind === "a" ? viewed.v.b?.label ?? "none" : "w");
