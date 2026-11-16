import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Strong = { readonly kind: "v"; readonly v: readonly [Box] } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: readonly [Weak<Box>] } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: [box()] };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? viewed.v[0]?.label ?? "none" : "w");
