import type { Weak } from "adamic";
interface Box { readonly label: string }
function box(): Box { return { label: "label" }; }
type Strong = { readonly kind: "v"; readonly v: ReadonlySet<Box> } | { readonly kind: "w"; readonly w: number };
type Viewed = { readonly kind: "v"; readonly v: ReadonlySet<Weak<Box>> } | { readonly kind: "w"; readonly w: number };
const strong: Strong = { kind: "v", v: new Set<Box>([box()]) };
const viewed: Viewed = strong;
console.log(viewed.kind === "v" ? String(viewed.v.size) : "w");
